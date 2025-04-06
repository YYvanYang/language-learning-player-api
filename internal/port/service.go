// internal/port/service.go
package port

import (
	"context"
	"time"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
)

// --- External Service Interfaces ---

// FileStorageService defines the contract for interacting with object storage.
type FileStorageService interface {
	// GetPresignedGetURL returns a temporary, signed URL for reading a private object.
	// expiry defines how long the URL should be valid.
	GetPresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)

	// GetPresignedPutURL returns a temporary, signed URL for uploading/overwriting an object.
	// The client MUST use the HTTP PUT method with this URL.
	GetPresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)

	// DeleteObject removes an object from storage.
	DeleteObject(ctx context.Context, bucket, objectKey string) error

	// TODO: Consider adding methods for checking object existence, getting metadata, etc. if needed.
}

// ExternalUserInfo contains standardized user info retrieved from an external identity provider.
type ExternalUserInfo struct {
	Provider        domain.AuthProvider // e.g., "google"
	ProviderUserID  string              // Unique ID from the provider (e.g., Google subject ID)
	Email           string              // Email address provided by the provider
	IsEmailVerified bool                // Whether the provider claims the email is verified
	Name            string
	PictureURL      *string             // Optional profile picture URL
}

// ExternalAuthService defines the contract for verifying external authentication credentials.
type ExternalAuthService interface {
	// VerifyGoogleToken verifies a Google ID token and returns standardized user info.
	// Returns domain.ErrAuthenticationFailed if the token is invalid or verification fails.
	VerifyGoogleToken(ctx context.Context, idToken string) (*ExternalUserInfo, error)

	// Add methods for other providers if needed (e.g., VerifyFacebookToken, VerifyAppleToken)
}


// --- Internal Helper Service Interfaces ---

// SecurityHelper defines cryptographic operations needed by use cases.
type SecurityHelper interface {
	 // HashPassword generates a secure hash (e.g., bcrypt) of the password.
	 HashPassword(ctx context.Context, password string) (string, error)
	 // CheckPasswordHash compares a plain password with a stored hash.
	 CheckPasswordHash(ctx context.Context, password, hash string) bool
	 // GenerateJWT creates a signed JWT (Access Token) for the given user ID.
	 GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error)
	 // VerifyJWT validates a JWT string and returns the UserID contained within.
	 // Returns domain.ErrUnauthenticated or domain.ErrAuthenticationFailed on failure.
	 VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error)

	 // TODO: Add methods for Refresh Token generation/validation if implementing that flow.
}

// REMOVED UserUseCase interface from here