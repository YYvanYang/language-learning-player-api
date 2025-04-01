// internal/adapter/handler/http/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path (for SecurityHelper interface)
	"your_project/pkg/httputil"    // Adjust import path
)

const UserIDKey ContextKey = "userID" // Use the same type as RequestIDKey

// Authenticator creates a middleware that verifies the JWT token.
func Authenticator(secHelper port.SecurityHelper) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				err := fmt.Errorf("%w: Authorization header missing", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				err := fmt.Errorf("%w: Authorization header format must be Bearer {token}", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			tokenString := headerParts[1]
			if tokenString == "" {
				err := fmt.Errorf("%w: Authorization token missing", domain.ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			// Verify the token using the SecurityHelper
			userID, err := secHelper.VerifyJWT(r.Context(), tokenString)
			if err != nil {
				// VerifyJWT should return domain errors (ErrAuthenticationFailed, ErrUnauthenticated)
				httputil.RespondError(w, r, err)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)

			// Token is valid, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext retrieves the UserID from the context.
// Returns domain.UserID zero value and false if not found or type is wrong.
func GetUserIDFromContext(ctx context.Context) (domain.UserID, bool) {
	userID, ok := ctx.Value(UserIDKey).(domain.UserID)
	return userID, ok
}