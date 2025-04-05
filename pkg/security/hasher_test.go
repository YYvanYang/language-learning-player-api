// pkg/security/hasher_test.go
package security_test

import (
	"testing"
    "log/slog"
    "io"
	"github.com/yvanyang/language-learning-player-backend/pkg/security" // Adjust

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a dummy logger
func newTestLogger() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBcryptHasher(t *testing.T) {
	hasher := security.NewBcryptHasher(newTestLogger())
	password := "mysecretpassword"

	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, password, hash) // Ensure it's hashed

	// Check correct password
	match := hasher.CheckPasswordHash(password, hash)
	assert.True(t, match, "Correct password should match")

	// Check incorrect password
	match = hasher.CheckPasswordHash("wrongpassword", hash)
	assert.False(t, match, "Incorrect password should not match")

    // Check empty password
	match = hasher.CheckPasswordHash("", hash)
	assert.False(t, match, "Empty password should not match")

    // Check with invalid hash (should return false, log warning)
    match = hasher.CheckPasswordHash(password, "invalidhashformat")
    assert.False(t, match, "Invalid hash format should result in mismatch")
}