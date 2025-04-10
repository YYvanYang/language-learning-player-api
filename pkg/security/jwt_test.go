package security

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

const testJwtSecret = "test-super-secret-key-for-unit-testing"

func TestJWTHelper_GenerateAndVerifyJWT(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	helper, err := NewJWTHelper(testJwtSecret, logger)
	assert.NoError(t, err)

	userID := domain.NewUserID()
	duration := 15 * time.Minute

	tokenString, err := helper.GenerateJWT(userID, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the generated token
	parsedUserID, err := helper.VerifyJWT(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)

	// Try verifying with a tampered token (invalid signature)
	tamperedToken := tokenString + "tamper"
	_, err = helper.VerifyJWT(tamperedToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // VerifyJWT maps errors

	// Try verifying an expired token
	shortDuration := -5 * time.Minute // Expired 5 minutes ago
	expiredTokenString, err := helper.GenerateJWT(userID, shortDuration)
	assert.NoError(t, err)

	// Wait a tiny bit to ensure expiry check works reliably
	time.Sleep(50 * time.Millisecond)

	_, err = helper.VerifyJWT(expiredTokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed) // VerifyJWT maps errors
	assert.Contains(t, err.Error(), "token expired")       // Check underlying reason if needed

	// Try verifying with a different secret key
	wrongHelper, _ := NewJWTHelper("wrong-secret", logger)
	_, err = wrongHelper.VerifyJWT(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)

	// Try verifying a malformed token
	malformedToken := "this.is.not.a.jwt"
	_, err = helper.VerifyJWT(malformedToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "malformed token")
}

func TestJWTHelper_VerifyJWT_InvalidUserIDFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	helper, err := NewJWTHelper(testJwtSecret, logger)
	assert.NoError(t, err)

	// Generate a token with a non-UUID string in the UserID claim
	claims := &Claims{
		UserID: "not-a-valid-uuid",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "language-learning-player",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(helper.secretKey)
	assert.NoError(t, err)

	// Try to verify it
	_, err = helper.VerifyJWT(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	assert.Contains(t, err.Error(), "invalid user ID format in token")
}

func TestNewJWTHelper_EmptySecret(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewJWTHelper("", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT secret key cannot be empty")
}
