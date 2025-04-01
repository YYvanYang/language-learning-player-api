// pkg/security/security.go
package security

import (
	"context"
	"log/slog"
	"time"

	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
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
	// Context is not used by bcrypt, but included for interface consistency
	return s.hasher.HashPassword(password)
}

// CheckPasswordHash compares a plain password with a stored hash.
func (s *Security) CheckPasswordHash(ctx context.Context, password, hash string) bool {
	// Context is not used by bcrypt, but included for interface consistency
	return s.hasher.CheckPasswordHash(password, hash)
}

// GenerateJWT creates a signed JWT (Access Token) for the given user ID.
func (s *Security) GenerateJWT(ctx context.Context, userID domain.UserID, duration time.Duration) (string, error) {
	// Context could potentially be used for tracing in the future
	return s.jwt.GenerateJWT(userID, duration)
}

// VerifyJWT validates a JWT string and returns the UserID contained within.
func (s *Security) VerifyJWT(ctx context.Context, tokenString string) (domain.UserID, error) {
	// Context could potentially be used for tracing in the future
	return s.jwt.VerifyJWT(tokenString)
}

// Compile-time check to ensure Security satisfies the port.SecurityHelper interface
var _ port.SecurityHelper = (*Security)(nil)