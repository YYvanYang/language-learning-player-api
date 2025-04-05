// internal/usecase/auth_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path (needed for JWT expiry)
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
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
	if eas == nil {
		// Log a warning or panic if external auth is mandatory for the use case setup
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

// AuthenticateWithGoogle handles login or registration via Google ID Token.
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error) {
	if uc.extAuthService == nil {
		uc.logger.ErrorContext(ctx, "ExternalAuthService not configured for Google authentication")
		return "", false, fmt.Errorf("google authentication is not enabled") // Service unavailable error
	}

	// 1. Verify Google ID token and get standardized user info
	extInfo, err := uc.extAuthService.VerifyGoogleToken(ctx, googleIdToken)
	if err != nil {
		// VerifyGoogleToken already logs details and returns domain.ErrAuthenticationFailed
		return "", false, err // Propagate the specific error (likely auth failed)
	}

	// --- User Lookup/Creation Logic ---

	// 2. Try finding user by Google ID first (most reliable link)
	user, err := uc.userRepo.FindByProviderID(ctx, extInfo.Provider, extInfo.ProviderUserID)
	if err == nil {
		// User found via Google ID - Login success
		uc.logger.InfoContext(ctx, "User authenticated via existing Google ID", "userID", user.ID, "googleID", extInfo.ProviderUserID)
		token, tokenErr := uc.secHelper.GenerateJWT(ctx, user.ID, uc.jwtExpiry)
		if tokenErr != nil {
			uc.logger.ErrorContext(ctx, "Failed to generate JWT for existing Google user", "error", tokenErr, "userID", user.ID)
			return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr)
		}
		return token, false, nil // Existing user, logged in
	}

	// Handle repository errors other than "Not Found"
	if !errors.Is(err, domain.ErrNotFound) {
		uc.logger.ErrorContext(ctx, "Error finding user by provider ID", "error", err, "provider", extInfo.Provider, "providerUserID", extInfo.ProviderUserID)
		return "", false, fmt.Errorf("database error during authentication: %w", err) // Internal error
	}

	// User not found by Google ID, proceed to check by email

	// 3. Try finding user by email (only if email is present in token)
	if extInfo.Email == "" {
		// Policy decision: If no email from Google, we cannot link or create based on email.
		// Since not found by Google ID either, treat as a new user scenario *without* email linking.
		// Fall through to user creation (step 4).
		uc.logger.InfoContext(ctx, "Google token verified, but no email provided. Proceeding to create new user based on Google ID only.", "googleID", extInfo.ProviderUserID)
	} else {
		// Email is present, try finding by email.
		emailVO, emailErr := domain.NewEmail(extInfo.Email)
		if emailErr != nil {
			// Should ideally not happen if Google provides valid email format, but handle defensively.
			uc.logger.WarnContext(ctx, "Invalid email format received from Google token", "error", emailErr, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
			// Treat as if email wasn't provided? Or fail? Let's fail for now.
			return "", false, fmt.Errorf("%w: invalid email format from provider", domain.ErrAuthenticationFailed)
		}

		userByEmail, errEmail := uc.userRepo.FindByEmail(ctx, emailVO)
		if errEmail == nil {
			// User found by email! Now decide how to handle this existing account.

			// Check if this existing account is already linked to *this* Google ID
			if userByEmail.GoogleID != nil && *userByEmail.GoogleID == extInfo.ProviderUserID && userByEmail.AuthProvider == domain.AuthProviderGoogle {
				// This case should have been caught by FindByProviderID, but check defensively.
				uc.logger.InfoContext(ctx, "User found by email is already correctly linked to this Google ID", "userID", userByEmail.ID, "email", extInfo.Email)
				token, tokenErr := uc.secHelper.GenerateJWT(ctx, userByEmail.ID, uc.jwtExpiry)
				if tokenErr != nil { /* handle error */ return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr) }
				return token, false, nil
			}

			// Check if this existing account is already linked to a *different* Google ID
			if userByEmail.GoogleID != nil && *userByEmail.GoogleID != extInfo.ProviderUserID {
				uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but is linked to a different Google account", "email", extInfo.Email, "existingGoogleID", *userByEmail.GoogleID, "attemptedGoogleID", extInfo.ProviderUserID)
				return "", false, fmt.Errorf("%w: email is linked to a different Google account", domain.ErrConflict)
			}

			// Check if the existing account uses a different provider (e.g., local password) and has no Google ID yet.
			if userByEmail.AuthProvider != domain.AuthProviderGoogle && userByEmail.GoogleID == nil {
				// Strategy: Link Google ID to existing local account.
				// Ensure Google email is verified for safety? (Optional policy)
				if !extInfo.IsEmailVerified {
					uc.logger.WarnContext(ctx, "Attempt to link Google account to existing local user, but Google email is not verified", "email", extInfo.Email, "userID", userByEmail.ID)
					// return "", false, fmt.Errorf("%w: cannot link unverified Google email to existing account", domain.ErrConflict) // Or allow, but with caution.
				}

				uc.logger.InfoContext(ctx, "Linking Google ID to existing user found by email", "userID", userByEmail.ID, "email", extInfo.Email, "googleID", extInfo.ProviderUserID)
				// Update the user in the repository
				userByEmail.GoogleID = &extInfo.ProviderUserID
				// Decide if AuthProvider should change. If they can log in either way, maybe keep 'local' or add 'google' flag?
				// Let's keep AuthProvider as is for now, just add GoogleID link.
				userByEmail.UpdatedAt = time.Now() // Update timestamp
				if updateErr := uc.userRepo.Update(ctx, userByEmail); updateErr != nil {
					uc.logger.ErrorContext(ctx, "Failed to update user repository while linking Google ID", "error", updateErr, "userID", userByEmail.ID)
					return "", false, fmt.Errorf("failed to link Google account: %w", updateErr)
				}

				// Generate token for the now-linked user
				token, tokenErr := uc.secHelper.GenerateJWT(ctx, userByEmail.ID, uc.jwtExpiry)
				if tokenErr != nil { /* handle error */ return "", false, fmt.Errorf("failed to finalize login: %w", tokenErr)}
				return token, false, nil // Existing user, now linked, logged in
			}

			// If none of the above: e.g., user exists with email AND has the same ('google') or different ('facebook') provider
			// This usually implies a conflict state not covered above or should have been caught.
			uc.logger.WarnContext(ctx, "Google auth conflict: Email exists but cannot automatically link", "email", extInfo.Email, "existingUserID", userByEmail.ID, "existingProvider", userByEmail.AuthProvider)
			return "", false, fmt.Errorf("%w: email is already associated with an account", domain.ErrConflict)

		} else if !errors.Is(errEmail, domain.ErrNotFound) {
			// Handle unexpected repository error during email lookup
			uc.logger.ErrorContext(ctx, "Error finding user by email", "error", errEmail, "email", extInfo.Email)
			return "", false, fmt.Errorf("database error during authentication: %w", errEmail) // Internal error
		}
		// If errEmail is ErrNotFound, fall through to user creation (step 4).
	}


	// 4. User not found by Google ID or linkable Email - Create new user
	uc.logger.InfoContext(ctx, "No existing user found by Google ID or linkable email. Creating new user.", "googleID", extInfo.ProviderUserID, "email", extInfo.Email)

	// Use domain constructor for Google user
	newUser, err := domain.NewGoogleUser(
		extInfo.Email,         // Can be empty if not provided by Google
		extInfo.Name,
		extInfo.ProviderUserID,
		extInfo.PictureURL,
	)
	if err != nil {
		// Errors from NewGoogleUser (e.g., empty Google ID, invalid email format if provided)
		uc.logger.ErrorContext(ctx, "Failed to create new Google user domain object", "error", err, "extInfo", extInfo)
		// This likely indicates a problem with the data from Google or our validation.
		return "", false, fmt.Errorf("failed to process user data from Google: %w", err)
	}

	// Persist the new user
	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new Google user to repository", "error", err, "googleID", newUser.GoogleID, "email", newUser.Email.String())
		// Check for potential race condition conflict (e.g., email was created between check and insert)
		// This requires the repo Create method to potentially return ErrConflict.
		// var pgErr *pgconn.PgError
		// if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		// 	 return "", false, fmt.Errorf("%w: email or google id already exists", domain.ErrConflict)
		// }
		return "", false, fmt.Errorf("failed to create new user account: %w", err) // Internal error
	}

	uc.logger.InfoContext(ctx, "New user created via Google authentication", "userID", newUser.ID, "email", newUser.Email.String())

	// Generate JWT for the new user
	token, tokenErr := uc.secHelper.GenerateJWT(ctx, newUser.ID, uc.jwtExpiry)
	if tokenErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate JWT for newly created Google user", "error", tokenErr, "userID", newUser.ID)
		// Critical failure: User created but cannot log in.
		// Options: Delete user? Mark inactive? For now, return error.
		return "", true, fmt.Errorf("failed to finalize registration: %w", tokenErr)
	}

	return token, true, nil // New user created and logged in
}