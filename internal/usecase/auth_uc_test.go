// =============================================
// FILE: internal/usecase/auth_uc_test.go
// =============================================
// This file was already provided in all_code.md and looks relatively complete for the tested functions.
// Some TODOs were present. Let's ensure the mocks package name is correct and add the missing ones.
// NOTE: Depends on mocks being generated in `internal/port/mocks` correctly.
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"io" // For discarding logs
	"log/slog"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port" // Import port for interfaces
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Adjusted import path for mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"


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
    var createdUser *domain.User // Capture created user
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        createdUser = user // Capture
        return user.Email == emailVO && user.HashedPassword != nil && *user.HashedPassword == hashedPassword
    })).Return(nil).Once()
    // 4. GenerateJWT should succeed (use captured ID)
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).
        Run(func(args mock.Arguments) {
            require.NotNil(t, createdUser, "Create must be called before GenerateJWT")
            assert.Equal(t, createdUser.ID, args.Get(1).(domain.UserID))
        }).
        Return(expectedToken, nil).
        Once()


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

func TestAuthUseCase_RegisterWithPassword_ShortPassword(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockSecHelper := mocks.NewMockSecurityHelper(t)
	mockExtAuth := mocks.NewMockExternalAuthService(t)
	logger := newTestLogger()
	cfg := newTestJWTConfig()
	uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

	emailStr := "shortpass@example.com"
	emailVO, _ := domain.NewEmail(emailStr)
	shortPassword := "1234567" // Less than 8 chars

	// Expect FindByEmail to return NotFound
	mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(nil, domain.ErrNotFound).Once()

	// Execute
	user, token, err := uc.RegisterWithPassword(context.Background(), emailStr, shortPassword, "Shorty")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
	assert.Contains(t, err.Error(), "at least 8 characters")
	assert.Nil(t, user)
	assert.Empty(t, token)
	mockUserRepo.AssertExpectations(t)
	// Ensure subsequent steps were not called
	mockSecHelper.AssertNotCalled(t, "HashPassword")
	mockUserRepo.AssertNotCalled(t, "Create")
	mockSecHelper.AssertNotCalled(t, "GenerateJWT")
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
    mockSecHelper.AssertNotCalled(t, "CheckPasswordHash")
    mockSecHelper.AssertNotCalled(t, "GenerateJWT")
}

 func TestAuthUseCase_LoginWithPassword_WrongPassword(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "test@example.com"
    password := "wrongpassword"
    emailVO, _ := domain.NewEmail(emailStr)
    hashedPassword := "correct_hashed_password"
    foundUser := &domain.User{ID: domain.NewUserID(), Email: emailVO, HashedPassword: &hashedPassword, AuthProvider: domain.AuthProviderLocal}

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
    mockSecHelper.AssertNotCalled(t, "GenerateJWT")
}

func TestAuthUseCase_LoginWithPassword_NonLocalUser(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    emailStr := "googleuser@example.com"
    emailVO, _ := domain.NewEmail(emailStr)
    foundUser := &domain.User{ID: domain.NewUserID(), Email: emailVO, AuthProvider: domain.AuthProviderGoogle, HashedPassword: nil} // Google user

    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundUser, nil).Once()

    // Execute
    token, err := uc.LoginWithPassword(context.Background(), emailStr, "password")

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
    assert.Empty(t, token)
    mockUserRepo.AssertExpectations(t)
    mockSecHelper.AssertNotCalled(t, "CheckPasswordHash")
    mockSecHelper.AssertNotCalled(t, "GenerateJWT")
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
    var createdUser *domain.User
    mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
        createdUser = user // Capture
        return user.GoogleID != nil && *user.GoogleID == googleUserID &&
               user.Email.String() == googleEmail &&
               user.AuthProvider == domain.AuthProviderGoogle
    })).Return(nil).Once()

    // Mock Sec Helper: GenerateJWT should succeed (using captured user ID)
    mockSecHelper.On("GenerateJWT", mock.Anything, mock.AnythingOfType("domain.UserID"), cfg.AccessTokenExpiry).
        Run(func(args mock.Arguments) {
            require.NotNil(t, createdUser, "Create must be called before GenerateJWT")
            assert.Equal(t, createdUser.ID, args.Get(1).(domain.UserID))
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
    emailVO, _ := domain.NewEmail(googleEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: googleUserID, Email: googleEmail}
    foundUser := &domain.User{ID: existingUserID, Email: emailVO, GoogleID: &googleUserID, AuthProvider: domain.AuthProviderGoogle}

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

func TestAuthUseCase_AuthenticateWithGoogle_EmailConflict(t *testing.T) {
    // Test Scenario: User tries Google Sign-In, email exists locally, Strategy C applied (Conflict)
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "valid_google_token"
    googleUserID := "google456"
    conflictEmail := "local.user@example.com" // Email matches existing local user
    emailVO, _ := domain.NewEmail(conflictEmail)

    extInfo := &port.ExternalUserInfo{ ProviderUserID: googleUserID, Email: conflictEmail }
    foundLocalUser := &domain.User{ID: domain.NewUserID(), Email: emailVO, AuthProvider: domain.AuthProviderLocal, GoogleID: nil} // Local user

    // Mocks
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(extInfo, nil).Once()
    mockUserRepo.On("FindByProviderID", mock.Anything, domain.AuthProviderGoogle, googleUserID).Return(nil, domain.ErrNotFound).Once()
    mockUserRepo.On("FindByEmail", mock.Anything, emailVO).Return(foundLocalUser, nil).Once() // Find by email succeeds

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict) // Expect conflict based on Strategy C
    assert.False(t, isNew)
    assert.Empty(t, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockUserRepo.AssertNotCalled(t, "Create") // Create should not be called
    mockSecHelper.AssertNotCalled(t, "GenerateJWT") // JWT should not be generated
}

func TestAuthUseCase_AuthenticateWithGoogle_VerifyTokenFails(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    mockExtAuth := mocks.NewMockExternalAuthService(t)
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, mockExtAuth, logger)

    googleToken := "invalid_or_expired_token"
    expectedError := fmt.Errorf("%w: google validation failed", domain.ErrAuthenticationFailed)

    // Mock ExternalAuthService verification failure
    mockExtAuth.On("VerifyGoogleToken", mock.Anything, googleToken).Return(nil, expectedError).Once()

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
    assert.False(t, isNew)
    assert.Empty(t, token)

    // Verify mocks
    mockExtAuth.AssertExpectations(t)
    // Ensure no repo or sec helper methods were called
    mockUserRepo.AssertNotCalled(t, "FindByProviderID")
    mockUserRepo.AssertNotCalled(t, "FindByEmail")
    mockUserRepo.AssertNotCalled(t, "Create")
    mockSecHelper.AssertNotCalled(t, "GenerateJWT")
}

func TestAuthUseCase_AuthenticateWithGoogle_NoExternalService(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    mockSecHelper := mocks.NewMockSecurityHelper(t)
    // mockExtAuth is NIL
    logger := newTestLogger()
    cfg := newTestJWTConfig()
    uc := usecase.NewAuthUseCase(cfg, mockUserRepo, mockSecHelper, nil, logger) // Pass nil for extAuth

    googleToken := "any_token"

    // Execute
    token, isNew, err := uc.AuthenticateWithGoogle(context.Background(), googleToken)

    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "google authentication is not enabled")
    assert.False(t, isNew)
    assert.Empty(t, token)
}