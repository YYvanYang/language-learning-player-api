// ============================================
// FILE: pkg/security/jwt_test.go
// ============================================
package security_test

import (
	"testing"
	"time"
    "log/slog"
    "io"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust
	"github.com/yvanyang/language-learning-player-backend/pkg/security"    // Adjust

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a dummy logger
func newTestLoggerForJWT() *slog.Logger {
     return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const testJWTSecret = "test-jwt-secret-key-minimum-32-bytes-long" // Use a sufficiently long key

func TestJWTHelper_GenerateAndVerify_Success(t *testing.T) {
	helper, err := security.NewJWTHelper(testJWTSecret, newTestLoggerForJWT())
	require.NoError(t, err)

	userID := domain.NewUserID()
	duration := 15 * time.Minute

	// Generate token
	tokenString, err := helper.GenerateJWT(userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Verify token immediately
	verifiedUserID, err := helper.VerifyJWT(tokenString)
	require.NoError(t, err)
	assert.Equal(t, userID, verifiedUserID)
}

func TestJWTHelper_Verify_ExpiredToken(t *testing.T) {
	helper, err := security.NewJWTHelper(testJWTSecret, newTestLoggerForJWT())
	require.NoError(t, err)

	userID := domain.NewUserID()
	// Generate a token that expired 1 second ago
	expiredDuration := -1 * time.Second

	expiredTokenString, err := helper.GenerateJWT(userID, expiredDuration)
	require.NoError(t, err)

	// Verify expired token
	_, err = helper.VerifyJWT(expiredTokenString)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed, "Should be authentication failed")
	assert.Contains(t, err.Error(), "token expired", "Error message should mention expiry")
}

func TestJWTHelper_Verify_InvalidSignature(t *testing.T) {
	helper1, err := security.NewJWTHelper(testJWTSecret, newTestLoggerForJWT())
	require.NoError(t, err)
	helper2, err := security.NewJWTHelper("different-secret-key-that-is-long-enough", newTestLoggerForJWT()) // Different secret
	require.NoError(t, err)

	userID := domain.NewUserID()
	duration := 15 * time.Minute

	// Generate token with helper1
	tokenString, err := helper1.GenerateJWT(userID, duration)
	require.NoError(t, err)

	// Verify token with helper2 (different secret)
	_, err = helper2.VerifyJWT(tokenString)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed, "Should be authentication failed")
	assert.Contains(t, err.Error(), "signature is invalid", "Error message should mention invalid signature")
}

func TestJWTHelper_Verify_MalformedToken(t *testing.T) {
	helper, err := security.NewJWTHelper(testJWTSecret, newTestLoggerForJWT())
	require.NoError(t, err)

	malformedToken := "this.is.not.a.jwt"

	_, err = helper.VerifyJWT(malformedToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "malformed", "Error message should mention malformed token")
}

func TestJWTHelper_New_EmptySecret(t *testing.T) {
    _, err := security.NewJWTHelper("", newTestLoggerForJWT())
    require.Error(t, err)
    assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}