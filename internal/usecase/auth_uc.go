// internal/usecase/auth_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
)

type AuthUseCase struct {
	userRepo       port.UserRepository
	secHelper      port.SecurityHelper
	extAuthService port.ExternalAuthService
	jwtExpiry      time.Duration
	logger         *slog.Logger
}

func NewAuthUseCase(
	cfg config.JWTConfig,
	ur port.UserRepository,
	sh port.SecurityHelper,
	eas port.ExternalAuthService,
	log *slog.Logger,
) *AuthUseCase {
	if eas == nil {
		log.Warn("AuthUseCase created without ExternalAuthService implementation.")
	}
	return &AuthUseCase{
		userRepo:       ur,
		secHelper:      sh,
		extAuthService: eas,
		jwtExpiry:      cfg.AccessTokenExpiry,
		logger:         log.With("usecase", "AuthUseCase"),
	}
}

// RegisterWithPassword handles user registration with email and password.
func (uc *AuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid email provided during registration", "email", emailStr, "error", err)
		return nil, "", fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
	}

	// Use EmailExists for a more efficient check
	exists, err := uc.userRepo.EmailExists(ctx, emailVO)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error checking email existence", "error", err, "email", emailStr)
		return nil, "", fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}

	if len(password) < 8 {
		return nil, "", fmt.Errorf("%w: password must be at least 8 characters long", domain.ErrInvalidArgument)
	}

	hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to hash password during registration", "error", err)
		return nil, "", fmt.Errorf("failed to process password: %w", err)
	}

	user, err := domain.NewLocalUser(emailStr, name, hashedPassword)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create domain user object", "error", err)
		return nil, "", fmt.Errorf("failed to create user data: %w", err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new user to repository", "error", err, "userID", user.ID)
		// The Create method already maps unique constraint errors to ErrConflict
		return nil, "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT after registration", "error", err, "userID", user.ID)
		return nil, "", fmt.Errorf("failed to finalize registration: %w", err)
	}

	uc.logger.InfoContext(ctx, "User registered successfully via password", "userID", user.ID, "email", emailStr)
	return user, token, nil
}

// LoginWithPassword handles user login with email and password.
func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (string, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		// Log invalid email format attempt? Or just fail silently. Let's fail silently.
		return "", domain.ErrAuthenticationFailed
	}
	user, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Login attempt for non-existent email", "email", emailStr)
			return "", domain.ErrAuthenticationFailed
		}
		uc.logger.ErrorContext(ctx, "Error finding user by email during login", "error", err, "email", emailStr)
		return "", fmt.Errorf("failed during login process: %w", err)
	}
	if user.AuthProvider != domain.AuthProviderLocal || user.HashedPassword == nil {
		uc.logger.WarnContext(ctx, "Login attempt for user with non-local provider or no password", "email", emailStr, "userID", user.ID, "provider", user.AuthProvider)
		return "", domain.ErrAuthenticationFailed
	}
	if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
		uc.logger.WarnContext(ctx, "Incorrect password provided for user", "email", emailStr, "userID", user.ID)
		return "", domain.ErrAuthenticationFailed
	}
	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT during login", "error", err, "userID", user.ID)
		return "", fmt.Errorf("failed to finalize login: %w", err)
	}
	uc.logger.InfoContext(ctx, "User logged in successfully via password", "userID", user.ID)
	return token, nil
}

// AuthenticateWithGoogle handles login or registration via Google ID Token.
// Implements Strategy C: Conflict if email exists with a different provider/link.
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error) {
	if uc.extAuthService == nil {
		uc.logger.ErrorContext(ctx, "ExternalAuthService not configured for Google authentication")
		return "", false, fmt.Errorf("google authentication is not enabled")
	}

	extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
	if err != nil {
		return "", false, err
	} // Propagate verification error

	// Check if user exists by Provider ID first
	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// Case 1: User found by Google ID -> Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		token, tokenErr := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
		if tokenErr != nil {
			uc.logger.ErrorContext(ctx, "Failed to generate JWT for existing Google user", "error", tokenErr, "userID", user.ID)
			return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr)
		}
		return token, false, nil // Not a new user
	}
	if !errors.Is(err, domain.ErrNotFound) {
		// Handle unexpected database errors during provider ID lookup
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return "", false, fmt.Errorf("database error during authentication: %w", err)
	}

	// User not found by Google ID, proceed to check by email (if available)
	if extInfo.Email != "" {
		emailVO, emailErr := domain.NewEmail(extInfo.Email)
		if emailErr != nil {
			// If Google gives an invalid email format, treat as auth failure
			uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
			return "", false, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
		}

		userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
		if errEmail == nil {
			// Case 2: User found by email -> Conflict (Strategy C)
			// Email exists, but Google ID wasn't linked or didn't match
			uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but Google ID did not match or was not linked", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
			return "", false, fmt.Errorf("%w: email is already associated with a different account", domain.ErrConflict)
		} else if !errors.Is(errEmail, domain.ErrNotFound) {
			// Handle unexpected database errors during email lookup
			uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
			return "", false, fmt.Errorf("database error during authentication: %w", errEmail)
		}
		// If errEmail is ErrNotFound, continue to user creation
		uc.logger.DebugContext(ctx, "Email not found, proceeding to create new Google user", "email", extInfo.Email)
	} else {
		// No email provided by Google, can only create user based on Google ID
		uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
	}

	// Case 3: User not found by Google ID AND (Email not found OR Email not provided) -> Create new user
	uc.logger.InfoContext(ctx, "Creating new user via Google authentication", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)

	// Create the new user domain object
	newUser, err := domain.NewGoogleUser(
		extInfo.Email, extInfo.Name, extInfo.ProviderUserID, extInfo.PictureURL,
	)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", err, "extInfo", extInfo)
		return "", false, fmt.Errorf("failed to process user data from Google: %w", err)
	}

	// Save the new user to the database
	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", err, "googleID", newUser.GoogleID, "email", newUser.Email.String())
		// Create should handle mapping unique constraints to ErrConflict
		return "", false, fmt.Errorf("failed to create new user account: %w", err)
	}

	// Generate JWT for the newly created user
	uc.logger.InfoContext(ctx, "New user created successfully via Google", "userID", newUser.ID, "email", newUser.Email.String())
	token, tokenErr := uc.secHelper.GenerateJWT(ctx, newUser.ID, uc.jwtExpiry)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT for newly created Google user", "error", tokenErr, "userID", newUser.ID)
		// Consider if user creation should be rolled back here - depends on transaction strategy
		return "", true, fmt.Errorf("failed to finalize registration: %w", tokenErr)
	}

	return token, true, nil // Return token and indicate new user
}

// Compile-time check to ensure AuthUseCase satisfies the port.AuthUseCase interface
var _ port.AuthUseCase = (*AuthUseCase)(nil)
