// internal/adapter/handler/http/middleware/recovery.go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
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

				// Respond with 500
				// TODO: Use httputil.RespondError once available for consistent JSON error response
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}