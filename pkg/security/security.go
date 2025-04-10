// ============================================
// FILE: pkg/security/security.go (MODIFIED)
// ============================================
package security

import (
	"context"
	"crypto/rand"     // ADDED
	"encoding/base64" // ADDED
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// Security implements the port.SecurityHelper interface.
type Security struct {
	hasher *BcryptHasher
	jwt    *JWTHelper
	logger *slog.Logger
}

// NewSecurity creates a new Security instance.
func NewSecurity(jwtSecretKey string, logger *slog.Logger) (*Security, error) {
	hasher := NewBcryptHasher(logger)
	jwtHelper, err := NewJWTHelper(jwtSecretKey, logger)
	if err != nil {
		return nil, err
	}
	return &Security{
		hasher: hasher,
		jwt:    jwtHelper,
		logger: logger.With("service", "SecurityHelper"),
	}, nil
}

// HashPassword generates a secure hash of the password.
func (s *Security) HashPassword(ctx context.Context, password string) (string, error) {
	return s.hasher.HashPassword(password)
}

// CheckPasswordHash compares a plain password with a stored hash.
func (s *Security) CheckPasswordHash(ctx context.Context, password, hash string) bool {
	return s.hasher.CheckPasswordHash(password, hash)
}

// GenerateJWT creates a signed JWT (Access Token) for the given user ID.
func (s *Security) GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error) {
	return s.jwt.GenerateJWT(userID, duration)
}

// VerifyJWT validates a JWT string and returns the UserID contained within.
func (s *Security) VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error) {
	return s.jwt.VerifyJWT(tokenString)
}

// GenerateRefreshTokenValue creates a cryptographically secure random string for the refresh token.
// ADDED METHOD
func (s *Security) GenerateRefreshTokenValue() (string, error) {
	numBytes := 32 // Generate a 32-byte random token -> 44 Base64 chars
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		s.logger.Error("Failed to generate random bytes for refresh token", "error", err)
		return "", fmt.Errorf("failed to generate refresh token value: %w", err)
	}
	// Use URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashRefreshTokenValue generates a SHA-256 hash of the refresh token value for storage.
// ADDED METHOD
func (s *Security) HashRefreshTokenValue(tokenValue string) string {
	return Sha256Hash(tokenValue) // Use the helper from hasher.go
}

// Compile-time check to ensure Security satisfies the port.SecurityHelper interface
// ADDED: New methods to SecurityHelper interface in port/service.go
var _ port.SecurityHelper = (*Security)(nil)
