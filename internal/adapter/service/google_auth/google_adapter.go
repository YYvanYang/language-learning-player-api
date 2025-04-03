// internal/adapter/service/google_auth/google_adapter.go
package googleauthadapter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	// Use the official Google idtoken verifier
	"google.golang.org/api/idtoken"

	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
)

// GoogleAuthService implements the port.ExternalAuthService interface for Google.
type GoogleAuthService struct {
	googleClientID string
	logger         *slog.Logger
	// http.Client is implicitly used by idtoken.Validate, but could be injected for testing/customization
}

// NewGoogleAuthService creates a new GoogleAuthService.
func NewGoogleAuthService(clientID string, logger *slog.Logger) (*GoogleAuthService, error) {
	if clientID == "" {
		return nil, fmt.Errorf("google client ID cannot be empty")
	}
	return &GoogleAuthService{
		googleClientID: clientID,
		logger:         logger.With("service", "GoogleAuthService"),
	}, nil
}

// VerifyGoogleToken verifies a Google ID token and returns standardized user info.
func (s *GoogleAuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*port.ExternalUserInfo, error) {
	if idToken == "" {
		return nil, fmt.Errorf("%w: google ID token cannot be empty", domain.ErrAuthenticationFailed)
	}

	// Validate the token using Google's library.
	// It checks signature, expiration, issuer ('accounts.google.com' or 'https://accounts.google.com'),
	// and audience (must match our client ID).
	payload, err := idtoken.Validate(ctx, idToken, s.googleClientID)
	if err != nil {
		s.logger.WarnContext(ctx, "Google ID token validation failed", "error", err)
		// Map validation errors to our domain error
		// The library might return specific error types, but a general failure is often sufficient here.
		return nil, fmt.Errorf("%w: invalid google token: %v", domain.ErrAuthenticationFailed, err)
	}

	// Token is valid, extract claims.
	// Standard claims reference: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	userID, ok := payload.Claims["sub"].(string) // Subject - Google's unique user ID
	if !ok || userID == "" {
		s.logger.ErrorContext(ctx, "Google ID token missing 'sub' (subject) claim", "claims", payload.Claims)
		return nil, fmt.Errorf("%w: missing required user identifier in token", domain.ErrAuthenticationFailed)
	}

	email, _ := payload.Claims["email"].(string) // Email claim
	emailVerified, _ := payload.Claims["email_verified"].(bool) // Email verified claim
	name, _ := payload.Claims["name"].(string) // Name claim
	picture, _ := payload.Claims["picture"].(string) // Picture URL claim

	// Basic check: ensure we have at least an email if email_verified is true
	if emailVerified && email == "" {
		s.logger.WarnContext(ctx, "Google ID token claims email_verified=true but email is missing", "subject", userID)
		// Decide policy: reject or proceed without email? Rejecting is safer.
		// return nil, fmt.Errorf("%w: verified email missing in token", domain.ErrAuthenticationFailed)
	}
	// If email is not verified by Google, maybe we shouldn't trust it for login/registration?
	// Current policy: Use the email regardless, but store the verification status.
	// Consider enforcing emailVerified == true if needed.


	s.logger.InfoContext(ctx, "Google ID token verified successfully", "subject", userID, "email", email)

	// Map to our standardized ExternalUserInfo struct
	userInfo := &port.ExternalUserInfo{
		Provider:        domain.AuthProviderGoogle,
		ProviderUserID:  userID,
		Email:           email,
		IsEmailVerified: emailVerified,
		Name:            name,
		// PictureURL:      &picture, // Assign if picture is not empty
	}
	if picture != "" {
		userInfo.PictureURL = &picture
	}

	return userInfo, nil
}

// Compile-time check
var _ port.ExternalAuthService = (*GoogleAuthService)(nil)