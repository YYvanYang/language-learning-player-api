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

	existingUser, err := uc.userRepo.FindByEmail(ctx, emailVO)
	if err == nil && existingUser != nil {
		uc.logger.WarnContext(ctx, "Registration attempt with existing email", "email", emailStr)
		return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking for existing email", "error", err, "email", emailStr)
		return nil, "", fmt.Errorf("failed to check email existence: %w", err)
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
		if errors.Is(err, domain.ErrConflict) {
			return nil, "", fmt.Errorf("%w: email already registered", domain.ErrConflict)
		}
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
	if err != nil { return "", false, err }

	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// Case 1: User found by Google ID -> Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		token, tokenErr := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
		if tokenErr != nil { return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr) }
		return token, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return "", false, fmt.Errorf("database error during authentication: %w", err)
	}

	// User not found by Google ID, check email if available
	if extInfo.Email != "" {
		emailVO, emailErr := domain.NewEmail(extInfo.Email)
		if emailErr != nil {
			uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
			return "", false, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
		}

		userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
		if errEmail == nil {
			// Case 2: User found by email -> Conflict (Strategy C)
			uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but Google ID did not match or was not linked", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
			return "", false, fmt.Errorf("%w: email is already associated with a different account", domain.ErrConflict)
		} else if !errors.Is(errEmail, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
			return "", false, fmt.Errorf("database error during authentication: %w", errEmail)
		}
		// If errEmail is ErrNotFound, fall through to user creation.
	} else {
		uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
	}

	// Case 3: User not found by Google ID or conflicting Email -> Create new user
	uc.logger.InfoContext(ctx, "No existing user found by Google ID or conflicting email. Creating new user.", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)

	newUser, err := domain.NewGoogleUser(
		extInfo.Email, extInfo.Name, extInfo.ProviderUserID, extInfo.PictureURL,
	)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", err, "extInfo", extInfo)
		return "", false, fmt.Errorf("failed to process user data from Google: %w", err)
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", err, "googleID", newUser.GoogleID, "email", newUser.Email.String())
		if errors.Is(err, domain.ErrConflict) {
             return "", false, fmt.Errorf("%w: account conflict during creation", domain.ErrConflict)
        }
		return "", false, fmt.Errorf("failed to create new user account: %w", err)
	}

	uc.logger.InfoContext(ctx, "New user created via Google authentication", "userID", newUser.ID, "email", newUser.Email.String())
	token, tokenErr := uc.secHelper.GenerateJWT(ctx, newUser.ID, uc.jwtExpiry)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT for newly created Google user", "error", tokenErr, "userID", newUser.ID)
		return "", true, fmt.Errorf("failed to finalize registration: %w", tokenErr)
	}
	return token, true, nil
}

// Compile-time check to ensure AuthUseCase satisfies the port.AuthUseCase interface
var _ port.AuthUseCase = (*AuthUseCase)(nil)
