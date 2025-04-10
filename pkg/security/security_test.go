package security

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// Use the same test secret as jwt_test.go for consistency
const testSecurityJwtSecret = "test-super-secret-key-for-unit-testing"

func TestSecurity_Integration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sec, err := NewSecurity(testSecurityJwtSecret, logger)
	assert.NoError(t, err)
	assert.NotNil(t, sec)

	ctx := context.Background()
	password := "verySecureP@ssw0rd"
	userID := domain.NewUserID()
	duration := 5 * time.Minute

	// 1. Hash Password
	hashedPassword, err := sec.HashPassword(ctx, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// 2. Check Correct Password
	match := sec.CheckPasswordHash(ctx, password, hashedPassword)
	assert.True(t, match)

	// 3. Check Incorrect Password
	match = sec.CheckPasswordHash(ctx, "wrongPassword", hashedPassword)
	assert.False(t, match)

	// 4. Generate JWT
	tokenString, err := sec.GenerateJWT(ctx, userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 5. Verify JWT
	parsedUserID, err := sec.VerifyJWT(ctx, tokenString)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)

	// 6. Generate Refresh Token Value
	refreshTokenVal, err := sec.GenerateRefreshTokenValue()
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshTokenVal)
	assert.Len(t, refreshTokenVal, 44) // 32 bytes -> 44 base64 chars

	// 7. Hash Refresh Token Value
	refreshTokenHash := sec.HashRefreshTokenValue(refreshTokenVal)
	assert.NotEmpty(t, refreshTokenHash)
	assert.Len(t, refreshTokenHash, 64) // SHA-256 hash is 64 hex chars

	// Ensure hashing is deterministic
	refreshTokenHash2 := sec.HashRefreshTokenValue(refreshTokenVal)
	assert.Equal(t, refreshTokenHash, refreshTokenHash2)

	// Ensure different values produce different hashes
	refreshTokenVal2, _ := sec.GenerateRefreshTokenValue()
	refreshTokenHash3 := sec.HashRefreshTokenValue(refreshTokenVal2)
	assert.NotEqual(t, refreshTokenHash, refreshTokenHash3)
}

func TestNewSecurity_EmptySecret(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewSecurity("", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}
