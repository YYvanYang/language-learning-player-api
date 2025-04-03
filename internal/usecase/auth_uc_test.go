// internal/usecase/auth_uc_test.go
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"your_project/internal/config"
	"your_project/internal/domain"
	"your_project/internal/port/mocks" // Import the generated mocks
	"your_project/internal/usecase"
	"log/slog"
    "io" // For discarding logs

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper to create a dummy logger that discards output
func newTestLogger() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create a dummy JWT config
func newTestJWTConfig() config.JWTConfig {
    return config.JWTConfig{AccessTokenExpiry: 1 * time.Hour}
}


func TestAuthUseCase_RegisterWithPassword_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t) // testify v1.9+ style
	mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t) // Needed for constructor, but not used here
    logger := newTestLogger()
    cfg := newTestJWTConfig()

	uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

	emailStr := "test@example.com"
	password := "password123"
	name := "Test User"
	emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "hashed_password"
    expectedToken := "jwt_token"

	// Define mock expectations
    // 1. FindByEmail should return NotFound
	mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()
    // 2. HashPassword should succeed
    mockSecHelper.On("HashPassword", mock.Anything, password).Return(hashedPassword, nil).Once()
    // 3. Create should succeed (match any user with correct email and hash)
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        return user.Email == emailVO && user.HashedPassword != nil && *user.HashedPassword == hashedPassword
    })).Return(nil).Once()
    // 4. GenerateJWT should succeed
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).Return(expectedToken, nil).Once()


	// Execute the use case method
	user, token, err := uc.RegisterWithPassword(context.Background(), emailStr, password, name)

	// Assert results
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, emailStr, user.Email.String())

	// Verify that mock expectations were met
	mockUserRepo.AssertExpectations(t)
	mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_RegisterWithPassword_EmailExists(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "exists@example.com"
    emailVO, _ := domain.NewEmail(emailStr)
    existingUser := &domain.User{Email: emailVO} // Dummy existing user

    // Expect FindByEmail to find the user
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(existingUser, nil).Once()

    // Execute
    user, token, err := uc.RegisterWithPassword(context.Background(), emailStr, "password", "name")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict)
    assert.Nil(t, user)
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertNotCalled(t, "HashPassword", mock.Anything, mock.Anything) // Should not be called
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)         // Should not be called
}


func TestAuthUseCase_LoginWithPassword_Success(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "test@example.com"
    password := "password123"
    emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "hashed_password" // Assume this is the correct hash
    userID := domain.NewUserID()
    foundUser := &domain.User{
        ID:             userID,
        Email:          emailVO,
        HashedPassword: &hashedPassword,
        AuthProvider:   domain.AuthProviderLocal,
    }
    expectedToken := "jwt_token"

    // Expectations
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundUser, nil).Once()
    mockSecHelper.On("CheckPasswordHash", mock.Anything, password, hashedPassword).Return(true).Once()
    mockSecHelper.On("GenerateJWT", mock.Anything, userID, cfg.AccessTokenExpiry).Return(expectedToken, nil).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, password)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedToken, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_LoginWithPassword_NotFound(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()

    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)
    emailStr := "notfound@example.com"
    emailVO, _ := domain.NewEmail(emailStr)

    // Expect FindByEmail to return NotFound
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, "password")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // Should map NotFound to AuthFailed
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertNotCalled(t, "CheckPasswordHash", mock.Anything, mock.Anything, mock.Anything)
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

 func TestAuthUseCase_LoginWithPassword_WrongPassword(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    // ... setup uc ...
    emailStr := "test@example.com"
    password := "wrongpassword"
    emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "correct_hashed_password"
    foundUser := &domain.User{ /* ... setup user ... */ HashedPassword: &hashedPassword, AuthProvider: domain.AuthProviderLocal}

    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundUser, nil).Once()
    // Expect CheckPasswordHash to return false
    mockSecHelper.On("CheckPasswordHash", mock.Anything, password, hashedPassword).Return(false).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, password)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t) // CheckPasswordHash was called
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthUseCase_AuthenticateWithGoogle_NewUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google123"
    googleEmail := "new.google.user@example.com"
    googleName := "Google User"
    expectedJWT := "new_user_jwt"

    // Mock ExternalAuthService verification
    extInfo := &port.ExternalUserInfo{
        Provider:        domain.AuthProviderGoogle,
        ProviderUserID:  googleUserID,
        Email:           googleEmail,
        IsEmailVerified: true,
        Name:            googleName,
    }
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID should return NotFound
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: FindByEmail should return NotFound
    emailVO, _ := domain.NewEmail(googleEmail)
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: Create should succeed
    // We capture the user being created to use its ID for JWT generation mock
    var createdUser *domain.User
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        createdUser = user // Capture the created user
        return user.GoogleID != nil && *user.GoogleID == googleUserID &&
               user.Email.String() == googleEmail &&
               user.AuthProvider == domain.AuthProviderGoogle
    })).Return(nil).Once()

    // Mock Sec Helper: GenerateJWT should succeed (using captured user ID)
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).
        Run(func(args mock.Arguments) {
            // Ensure the correct user ID is passed before returning
            passedUserID := args.Get(1).(domain.UserID)
            assert.Equal(t, createdUser.ID, passedUserID, "GenerateJWT called with wrong user ID")
        }).
        Return(expectedJWT, nil).
        Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.True(t, isNew, "Expected isNewUser to be true")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
}


func TestAuthUseCase_AuthenticateWithGoogle_ExistingGoogleUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google123"
    googleEmail := "exists.google.user@example.com"
    expectedJWT := "existing_user_jwt"
    existingUserID := domain.NewUserID()

    extInfo := &port.ExternalUserInfo{ /* ... setup extInfo ... */ ProviderUserID: googleUserID, Email: googleEmail}
    foundUser := &domain.User{ID: existingUserID, GoogleID: &googleUserID, AuthProvider: domain.AuthProviderGoogle}

    // Mock ExternalAuthService
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID finds the user
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(foundUser, nil).Once()

    // Mock Sec Helper: GenerateJWT for the existing user
    mockSecHelper.On("GenerateJWT", mock.Anything, existingUserID, cfg.AccessTokenExpiry).Return(expectedJWT, nil).Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.False(t, isNew, "Expected isNewUser to be false")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
    // Ensure FindByEmail and Create were not called
    mockUserRepo.AssertNotCalled(t, "FindByEmail", mock.Anything, mock.Anything)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAuthUseCase_AuthenticateWithGoogle_LinkToLocalUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google456"
    googleEmail := "local.user@example.com" // Email matches existing local user
    expectedJWT := "linked_user_jwt"
    existingUserID := domain.NewUserID()
    emailVO, _ := domain.NewEmail(googleEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: googleUserID, Email: googleEmail, IsEmailVerified: true} // Email verified
    foundLocalUser := &domain.User{ID: existingUserID, Email: emailVO, AuthProvider: domain.AuthProviderLocal, GoogleID: nil} // Local user, no Google ID yet

    // Mock ExternalAuthService
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()

    // Mock User Repo: FindByProviderID returns NotFound
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(nil, domain.ErrNotFound).Once()

    // Mock User Repo: FindByEmail finds the local user
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundLocalUser, nil).Once()

    // Mock User Repo: Update should be called to link the Google ID
    mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        return user.ID == existingUserID && user.GoogleID != nil && *user.GoogleID == googleUserID
    })).Return(nil).Once()

    // Mock Sec Helper: GenerateJWT for the existing (now linked) user
    mockSecHelper.On("GenerateJWT", mock.Anything, existingUserID, cfg.AccessTokenExpiry).Return(expectedJWT, nil).Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.NoError(t, err)
    assert.False(t, isNew, "Expected isNewUser to be false for linking")
    assert.Equal(t, expectedJWT, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertExpectations(t)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything) // Ensure Create is not called
}


func TestAuthUseCase_AuthenticateWithGoogle_EmailConflictDifferentGoogleID(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    attemptedGoogleUserID := "google-attempt-123"
    existingGoogleUserID := "google-exists-456"
    conflictEmail := "conflict.email@example.com"
    emailVO, _ := domain.NewEmail(conflictEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: attemptedGoogleUserID, Email: conflictEmail }
    // User exists with the same email but is already linked to a *different* Google ID
    existingUserWithDifferentGoogleID := &domain.User{
        ID: domain.NewUserID(),
        Email: emailVO,
        GoogleID: &existingGoogleUserID, // Linked to a different Google ID
        AuthProvider: domain.AuthProviderGoogle,
    }

    // Mocks
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, attemptedGoogleUserID).Return(nil, domain.ErrNotFound).Once()
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(existingUserWithDifferentGoogleID, nil).Once()

     // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict) // Expect conflict
    assert.False(t, isNew)
    assert.Empty(t, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    // Ensure Update and Create were not called
    mockUserRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
    mockUserRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
    mockSecHelper.AssertNotCalled(t, "GenerateJWT", mock.Anything, mock.Anything, mock.Anything)
}

// TODO: Add more AuthenticateWithGoogle tests:
// - Verification fails from ExternalAuthService
// - Email exists, but IsEmailVerified is false (if policy enforces verification)
// - Email exists, but provider is different and GoogleID is nil -> Link (covered by LinkToLocalUser)
// - Email exists, provider is different and GoogleID is NOT nil -> Conflict
// - Repository errors (other than NotFound) during find/create/update