// internal/adapter/handler/http/middleware/request_id.go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid" // Using google/uuid for request IDs
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
)

// RequestID is a middleware that injects a unique request ID into the context
// and sets it in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get existing ID from header or generate a new one
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Set ID in response header
		w.Header().Set("X-Request-ID", reqID)

		// Add ID to request context
		ctx := context.WithValue(r.Context(), httputil.RequestIDKey, reqID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetReqID retrieves the request ID from the context.
// Returns an empty string if not found.
func GetReqID(ctx context.Context) string {
	return httputil.GetReqID(ctx)
}
