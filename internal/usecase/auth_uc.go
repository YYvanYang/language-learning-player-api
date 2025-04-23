// ============================================
// FILE: internal/usecase/auth_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

type AuthUseCase struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository // Dependency for refresh token storage
	secHelper        port.SecurityHelper
	extAuthService   port.ExternalAuthService
	cfg              config.JWTConfig // Store the whole JWT config for expiries
	logger           *slog.Logger
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(
	cfg config.JWTConfig, // Pass whole JWTConfig
	ur port.UserRepository,
	rtr port.RefreshTokenRepository, // Inject RefreshTokenRepository
	sh port.SecurityHelper,
	eas port.ExternalAuthService,
	log *slog.Logger,
) *AuthUseCase {
	if eas == nil {
		log.Warn("AuthUseCase created without ExternalAuthService implementation.")
	}
	// Add check for RefreshTokenRepository
	if rtr == nil {
		// Log as error because refresh token functionality will be broken
		log.Error("AuthUseCase created without RefreshTokenRepository implementation. Refresh tokens cannot be stored or validated.")
	}
	return &AuthUseCase{
		userRepo:         ur,
		refreshTokenRepo: rtr, // Assign injected repo
		secHelper:        sh,
		extAuthService:   eas,
		cfg:              cfg, // Store config
		logger:           log.With("usecase", "AuthUseCase"),
	}
}

// generateAndStoreTokens is a helper to create access/refresh tokens and store the refresh token hash.
func (uc *AuthUseCase) generateAndStoreTokens(ctx context.Context, userID domain.UserID) (accessToken, refreshTokenValue string, err error) {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot generate/store tokens", "userID", userID)
		return "", "", fmt.Errorf("internal server error: authentication system misconfigured")
	}

	accessToken, err = uc.secHelper.GenerateJWT(ctx, userID, uc.cfg.AccessTokenExpiry)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenValue, err = uc.secHelper.GenerateRefreshTokenValue()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token value: %w", err)
	}

	refreshTokenHash := uc.secHelper.HashRefreshTokenValue(refreshTokenValue)
	expiresAt := time.Now().Add(uc.cfg.RefreshTokenExpiry)

	tokenData := &port.RefreshTokenData{
		TokenHash: refreshTokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(), // Set creation time here
	}

	// Save the refresh token data to the repository
	if err = uc.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshTokenValue, nil
}

// RegisterWithPassword handles user registration with email and password.
func (uc *AuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, port.AuthResult, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid email provided during registration", "email", emailStr, "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	exists, err := uc.userRepo.EmailExists(ctx, emailVO)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error checking email existence", "error", err, "email", emailStr)
		return nil, port.AuthResult{}, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, port.AuthResult{}, fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}

	if len(password) < 8 {
		return nil, port.AuthResult{}, fmt.Errorf("%w: password must be at least 8 characters long", domain.ErrInvalidArgument)
	}

	hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to hash password during registration", "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("failed to process password: %w", err)
	}

	user, err := domain.NewLocalUser(emailStr, name, hashedPassword)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create domain user object", "error", err)
		return nil, port.AuthResult{}, fmt.Errorf("failed to create user data: %w", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new user to repository", "error", err, "userID", user.ID)
		// Create should map unique constraints to ErrConflict
		return nil, port.AuthResult{}, fmt.Errorf("failed to register user: %w", err)
	}

	// Generate and store tokens
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, user.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens after registration", "error", tokenErr, "userID", user.ID)
		return nil, port.AuthResult{}, fmt.Errorf("failed to finalize registration session: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "User registered successfully via password", "userID", user.ID, "email", emailStr)
	// **MODIFIED RETURN**: Include user in AuthResult as well for consistency? Or just return user separately?
	// Returning separately is clear, but let's add it to AuthResult too for potential future use cases
	// and consistency with other methods.
	authRes := port.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}
	return user, authRes, nil
}

// LoginWithPassword handles user login with email and password.
func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (port.AuthResult, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}
	user, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Login attempt for non-existent email", "email", emailStr)
			return port.AuthResult{}, domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding user by email during login", "error", err, "email", emailStr)
		return port.AuthResult{}, fmt.Errorf("failed during login process: %w", err)
	}
	if user.AuthProvider != domain.AuthProviderLocal || user.HashedPassword == nil {
		uc.logger.WarnContext(ctx, "Login attempt for user with non-local provider or no password", "email", emailStr, "userID", user.ID, "provider", user.AuthProvider)
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}
	if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
		uc.logger.WarnContext(ctx, "Incorrect password provided for user", "email", emailStr, "userID", user.ID)
		return port.AuthResult{}, domain.ErrAuthenticationFailed
	}

	// Generate and store tokens
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, user.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens during login", "error", tokenErr, "userID", user.ID)
		return port.AuthResult{}, fmt.Errorf("failed to finalize login session: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "User logged in successfully via password", "userID", user.ID)
	return port.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// AuthenticateWithGoogle handles login or registration via Google ID Token.
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (port.AuthResult, error) {
	if uc.extAuthService == nil {
		uc.logger.ErrorContext(ctx, "ExternalAuthService not configured for Google authentication")
		return port.AuthResult{}, fmt.Errorf("google authentication is not enabled")
	}

	extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
	if err != nil {
		return port.AuthResult{}, err // Propagate verification error
	}

	var targetUser *domain.User
	var isNewUser bool

	// Check if user exists by Provider ID first
	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// Case 1: User found by Google ID -> Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		targetUser = user
		isNewUser = false
	} else if errors.Is(err, domain.ErrNotFound) {
		// User not found by Google ID, proceed to check by email (if available)
		if extInfo.Email != "" {
			emailVO, emailErr := domain.NewEmail(extInfo.Email)
			if emailErr != nil {
				uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
				return port.AuthResult{}, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
			}

			userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
			if errEmail == nil {
				// Case 2: User found by email -> Conflict (Strategy C)
				uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but Google ID did not match or was not linked", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
				return port.AuthResult{}, fmt.Errorf("%w: email is already associated with a different account", domain.ErrConflict)
			} else if !errors.Is(errEmail, domain.ErrNotFound) {
				uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
				return port.AuthResult{}, fmt.Errorf("database error during authentication: %w", errEmail)
			}
			uc.logger.DebugContext(ctx, "Email not found, proceeding to create new Google user", "email", extInfo.Email)
		} else {
			uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
		}

		// Case 3: Create new user
		uc.logger.InfoContext(ctx, "Creating new user via Google authentication", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)
		newUser, errCreate := domain.NewGoogleUser(extInfo.Email, extInfo.Name, extInfo.ProviderUserID, extInfo.PictureURL)
		if errCreate != nil {
			uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", errCreate, "extInfo", extInfo)
			return port.AuthResult{}, fmt.Errorf("failed to process user data from Google: %w", errCreate)
		}
		if errDb := uc.userRepo.Create(ctx, newUser); errDb != nil {
			uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", errDb, "googleID", newUser.GoogleID, "email", newUser.Email.String())
			// Create maps unique constraints to ErrConflict
			return port.AuthResult{}, fmt.Errorf("failed to create new user account: %w", errDb)
		}
		uc.logger.InfoContext(ctx, "New user created successfully via Google", "userID", newUser.ID, "email", newUser.Email.String())
		targetUser = newUser
		isNewUser = true

	} else {
		// Handle unexpected database errors during provider ID lookup
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return port.AuthResult{}, fmt.Errorf("database error during authentication: %w", err)
	}

	// Generate and store tokens for the targetUser (either found or newly created)
	accessToken, refreshToken, tokenErr := uc.generateAndStoreTokens(ctx, targetUser.ID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store tokens for Google auth", "error", tokenErr, "userID", targetUser.ID)
		return port.AuthResult{}, fmt.Errorf("failed to finalize authentication session: %w", tokenErr)
	}

	return port.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
		User:         targetUser,
	}, nil
}

// RefreshAccessToken validates a refresh token, revokes it, and issues new access/refresh tokens.
func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshTokenValue string) (port.AuthResult, error) {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot refresh token")
		return port.AuthResult{}, fmt.Errorf("internal server error: authentication system misconfigured")
	}

	if refreshTokenValue == "" {
		return port.AuthResult{}, fmt.Errorf("%w: refresh token is required", domain.ErrAuthenticationFailed)
	}

	tokenHash := uc.secHelper.HashRefreshTokenValue(refreshTokenValue)

	// Find the token data by hash
	tokenData, err := uc.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Refresh token hash not found in storage during refresh attempt")
			// It's crucial to return an authentication error here to prevent probing attacks
			return port.AuthResult{}, domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding refresh token by hash", "error", err)
		return port.AuthResult{}, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check expiry
	if time.Now().After(tokenData.ExpiresAt) {
		uc.logger.WarnContext(ctx, "Expired refresh token presented", "userID", tokenData.UserID, "expiresAt", tokenData.ExpiresAt)
		// Delete the expired token (best effort)
		_ = uc.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
		return port.AuthResult{}, fmt.Errorf("%w: refresh token expired", domain.ErrAuthenticationFailed)
	}

	// --- Rotation: Invalidate the old token ---
	delErr := uc.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
	if delErr != nil && !errors.Is(delErr, domain.ErrNotFound) {
		// Log the failure but proceed to issue new tokens.
		// The user presented a valid token, failure to delete shouldn't block refresh.
		// A background job could clean up orphaned tokens later if needed.
		uc.logger.ErrorContext(ctx, "Failed to delete used refresh token during rotation, proceeding anyway", "error", delErr, "userID", tokenData.UserID)
	}

	// --- Issue new tokens ---
	newAccessToken, newRefreshTokenValue, tokenErr := uc.generateAndStoreTokens(ctx, tokenData.UserID)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate/store new tokens during refresh", "error", tokenErr, "userID", tokenData.UserID)
		// This is a more critical failure. User might be left logged out.
		return port.AuthResult{}, fmt.Errorf("failed to issue new tokens after validating refresh token: %w", tokenErr)
	}

	uc.logger.InfoContext(ctx, "Access token refreshed successfully", "userID", tokenData.UserID)
	return port.AuthResult{AccessToken: newAccessToken, RefreshToken: newRefreshTokenValue}, nil
}

// Logout invalidates all refresh tokens for the given user ID.
// MODIFIED: Now operates based on UserID derived from access token.
func (uc *AuthUseCase) Logout(ctx context.Context, userID domain.UserID) error {
	// Ensure repo dependency is available before proceeding
	if uc.refreshTokenRepo == nil {
		uc.logger.ErrorContext(ctx, "RefreshTokenRepository is nil, cannot logout user", "userID", userID)
		// Still proceed to attempt clearing client-side cookies, but log server error
		return fmt.Errorf("internal server error: authentication system misconfigured")
	}

	// Delete all tokens associated with the user ID
	deletedCount, err := uc.refreshTokenRepo.DeleteByUser(ctx, userID)
	if err != nil {
		// Log actual errors during deletion but don't necessarily fail the logout flow for the client
		uc.logger.ErrorContext(ctx, "Failed to delete refresh tokens during logout", "error", err, "userID", userID)
		// Return nil so the client-side logout can proceed? Or return the error?
		// Let's return nil, as the main goal is logging the user out on the frontend.
		// The backend cleanup failure is logged.
		return nil
		// return fmt.Errorf("failed to process logout request fully: %w", err)
	}

	uc.logger.InfoContext(ctx, "User sessions logged out (all refresh tokens invalidated)", "userID", userID, "count", deletedCount)
	return nil
}

// Compile-time check to ensure AuthUseCase satisfies the port.AuthUseCase interface
var _ port.AuthUseCase = (*AuthUseCase)(nil)
