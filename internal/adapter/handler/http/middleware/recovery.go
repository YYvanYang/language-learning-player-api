// internal/adapter/handler/http/middleware/recovery.go
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
)

// Recoverer is a middleware that recovers from panics, logs the panic,
// and returns a 500 Internal Server Error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Log the panic
				logger := slog.Default() // Get default logger (set in main)
				// Get request ID if available (assuming RequestID middleware runs before)
				reqID := GetReqID(r.Context())
				logger.ErrorContext(r.Context(), "Panic recovered",
					"error", rvr,
					"request_id", reqID,
					"stack", string(debug.Stack()),
				)

				// Use httputil.RespondError for consistent JSON error response
				err := errors.New("internal server error: recovered from panic")
				httputil.RespondError(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
