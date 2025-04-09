// pkg/security/jwt.go
package security

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
)

// JWTHelper provides JWT generation and verification functionality.
type JWTHelper struct {
	secretKey []byte // Store as byte slice for JWT library
	logger    *slog.Logger
}

// Claims defines the structure of the JWT claims used in this application.
type Claims struct {
	UserID string `json:"uid"` // Store UserID as string in JWT
	jwt.RegisteredClaims
}

// NewJWTHelper creates a new JWTHelper.
func NewJWTHelper(secretKey string, logger *slog.Logger) (*JWTHelper, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}
	return &JWTHelper{
		secretKey: []byte(secretKey),
		logger:    logger.With("component", "JWTHelper"),
	}, nil
}

// GenerateJWT creates a new JWT token for the given user ID and duration.
func (h *JWTHelper) GenerateJWT(userID domain.UserID, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		UserID: userID.String(), // Convert UserID (UUID) to string
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "language-learning-player", // Optional: Identify issuer
			// Subject: userID.String(), // Can also use Subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		h.logger.Error("Error signing JWT token", "error", err, "userID", userID.String())
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// VerifyJWT validates the token string and returns the UserID.
func (h *JWTHelper) VerifyJWT(tokenString string) (domain.UserID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			h.logger.Warn("JWT token has expired", "error", err)
			return domain.UserID{}, fmt.Errorf("%w: token expired", domain.ErrAuthenticationFailed)
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			h.logger.Warn("Malformed JWT token received", "error", err)
			return domain.UserID{}, fmt.Errorf("%w: malformed token", domain.ErrAuthenticationFailed)
		}
		// Handle other errors like ErrSignatureInvalid, ErrTokenNotValidYet etc.
		h.logger.Warn("JWT token validation failed", "error", err)
		return domain.UserID{}, fmt.Errorf("%w: %v", domain.ErrAuthenticationFailed, err)
	}

	if !token.Valid {
		h.logger.Warn("Invalid JWT token received")
		return domain.UserID{}, domain.ErrAuthenticationFailed // General invalid token
	}

	// Convert UserID string from claim back to domain.UserID (UUID)
	userID, parseErr := domain.UserIDFromString(claims.UserID)
	if parseErr != nil {
		h.logger.Error("Error parsing UserID from valid JWT claims", "error", parseErr, "claimUserID", claims.UserID)
		return domain.UserID{}, fmt.Errorf("%w: invalid user ID format in token", domain.ErrAuthenticationFailed)
	}

	return userID, nil
}
