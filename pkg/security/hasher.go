// pkg/security/hasher.go
package security

import (
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

const defaultPasswordCost = 12 // Adjust cost based on your security needs and performance

// BcryptHasher provides password hashing using bcrypt.
type BcryptHasher struct {
	cost   int
	logger *slog.Logger
}

// NewBcryptHasher creates a new BcryptHasher.
func NewBcryptHasher(logger *slog.Logger) *BcryptHasher {
	return &BcryptHasher{
		cost:   defaultPasswordCost,
		logger: logger.With("component", "BcryptHasher"),
	}
}

// HashPassword generates a bcrypt hash for the given password.
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		h.logger.Error("Error generating password hash", "error", err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
func (h *BcryptHasher) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Log mismatch errors only at debug level if desired, other errors as warnings/errors
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.logger.Debug("Password hash mismatch", "error", err) // Optional: Log mismatch at debug level
		} else {
			h.logger.Warn("Error comparing password hash", "error", err)
		}
		return false
	}
	return true
}