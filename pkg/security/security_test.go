// ============================================
// FILE: pkg/security/security_test.go
// ============================================
package security_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/security" // Adjust
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLoggerForSecurity() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const testSecurityJWTSecret = "another-secure-key-at-least-32-bytes-long-ok"

func TestSecurity_Hashing(t *testing.T) {
	sec, err := security.NewSecurity(testSecurityJWTSecret, newTestLoggerForSecurity())
	require.NoError(t, err)

	password := "complexP@ss123"

	hash, err := sec.HashPassword(context.Background(), password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	match := sec.CheckPasswordHash(context.Background(), password, hash)
	assert.True(t, match, "Correct password should verify")

	match = sec.CheckPasswordHash(context.Background(), "wrongPass", hash)
	assert.False(t, match, "Incorrect password should not verify")
}

func TestSecurity_JWT(t *testing.T) {
	sec, err := security.NewSecurity(testSecurityJWTSecret, newTestLoggerForSecurity())
	require.NoError(t, err)

	userID := domain.NewUserID()
	duration := 5 * time.Minute

	token, err := sec.GenerateJWT(context.Background(), userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	verifiedID, err := sec.VerifyJWT(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, userID, verifiedID)
}

func TestSecurity_JWT_Expiry(t *testing.T) {
	sec, err := security.NewSecurity(testSecurityJWTSecret, newTestLoggerForSecurity())
	require.NoError(t, err)

	userID := domain.NewUserID()
	duration := -5 * time.Second // Expired

	token, err := sec.GenerateJWT(context.Background(), userID, duration)
	require.NoError(t, err)

	_, err = sec.VerifyJWT(context.Background(), token)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "token expired")
}

func TestSecurity_New_EmptySecret(t *testing.T) {
	_, err := security.NewSecurity("", newTestLoggerForSecurity())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}