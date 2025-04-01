// internal/usecase/auth_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"your_project/internal/config" // Adjust import path (needed for JWT expiry)
	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
)

type AuthUseCase struct {
	userRepo       port.UserRepository
	secHelper      port.SecurityHelper
	extAuthService port.ExternalAuthService // Will be used in Phase 4
	jwtExpiry      time.Duration          // Store JWT expiry duration
	logger         *slog.Logger
}

func NewAuthUseCase(
	cfg config.JWTConfig, // Pass JWT config for expiry
	ur port.UserRepository,
	sh port.SecurityHelper,
	eas port.ExternalAuthService, // Accept ExternalAuthService even if not used yet
	log *slog.Logger,
) *AuthUseCase {
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
		return nil, "", fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err) // Return specific invalid arg error
	}

	// 1. Check if email already exists
	existingUser, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err == nil && existingUser != nil {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking for existing email", "error", err, "email", emailStr)
		return nil, "", fmt.Errorf("failed to check email existence: %w", err) // Internal error
	}
	// If err is ErrNotFound, we can proceed.

	// 2. Validate password strength (basic example - could be more complex)
	if len(password) < 8 {
		return nil, "", fmt.Errorf("%w: password must be at least 8 characters long", domain.ErrInvalidArgument)
	}

	// 3. Hash password
	hashedPassword, err := uc.secHelper.HashPassword(ctx, password)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to hash password during registration", "error", err)
		return nil, "", fmt.Errorf("failed to process password: %w", err) // Internal error
	}

	// 4. Create domain user
	user, err := domain.NewLocalUser(emailStr, name, hashedPassword)
	if err != nil {
		// Should not happen if email was validated, but handle defensively
		uc.logger.ErrorContext(ctx, "Failed to create domain user object", "error", err)
		return nil, "", fmt.Errorf("failed to create user data: %w", err)
	}

	// 5. Save user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new user to repository", "error", err, "userID", user.ID)
		// Check if it was a conflict race condition
		if errors.Is(err, domain.ErrConflict) { // Assuming repo maps unique violation to ErrConflict
			return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
		}
		return nil, "", fmt.Errorf("failed to register user: %w", err) // Internal error
	}

	// 6. Generate JWT
	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT after registration", "error", err, "userID", user.ID)
		// Still return the user, but without a token? Or fail the whole process?
		// Let's fail, as login is expected after registration.
		// Consider cleanup (deleting user?) or marking user as inactive if token fails? For now, just fail.
		return nil, "", fmt.Errorf("failed to finalize registration: %w", err)
	}

	uc.logger.InfoContext(ctx, "User registered successfully via password", "userID", user.ID, "email", emailStr)
	return user, token, nil
}

// LoginWithPassword handles user login with email and password.
func (uc *AuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (string, error) {
	emailVO, err := domain.NewEmail(emailStr)
	if err != nil {
		return "", domain.ErrAuthenticationFailed // Treat invalid email format as auth failure
	}

	// 1. Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Login attempt for non-existent email", "email", emailStr)
			return "", domain.ErrAuthenticationFailed // Use generic auth failed error
		}
		uc.logger.ErrorContext(ctx, "Error finding user by email during login", "error", err, "email", emailStr)
		return "", fmt.Errorf("failed during login process: %w", err) // Internal error
	}

	// 2. Check if user uses local auth and has a password
	if user.AuthProvider != domain.AuthProviderLocal || user.HashedPassword == nil {
		uc.logger.WarnContext(ctx, "Login attempt for user with non-local provider or no password", "email", emailStr, "userID", user.ID, "provider", user.AuthProvider)
		return "", domain.ErrAuthenticationFailed // Account exists but cannot login via password
	}

	// 3. Validate password
	// Note: The domain user's ValidatePassword method uses bcrypt directly.
	// We use the secHelper here as it's the abstraction layer usecase depends on.
	// Alternatively, inject hasher directly or call user.ValidatePassword IF the user object included the hasher dependency (less common).
	if !uc.secHelper.CheckPasswordHash(ctx, password, *user.HashedPassword) {
		uc.logger.WarnContext(ctx, "Incorrect password provided for user", "email", emailStr, "userID", user.ID)
		return "", domain.ErrAuthenticationFailed
	}

	// 4. Generate JWT
	token, err := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT during login", "error", err, "userID", user.ID)
		return "", fmt.Errorf("failed to finalize login: %w", err) // Internal error
	}

	uc.logger.InfoContext(ctx, "User logged in successfully via password", "userID", user.ID)
	return token, nil
}

// AuthenticateWithGoogle - Placeholder for Phase 4
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error) {
	uc.logger.WarnContext(ctx, "AuthenticateWithGoogle called but not implemented yet")
	return "", false, fmt.Errorf("google authentication not implemented")
}