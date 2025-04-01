// internal/adapter/handler/http/middleware/logger.go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// ResponseWriterWrapper wraps http.ResponseWriter to capture status code.
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	// Default to 200 OK if WriteHeader is not called
	return &ResponseWriterWrapper{w, http.StatusOK}
}

func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs incoming requests and their processing time.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := slog.Default().With(
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"request_id", GetReqID(r.Context()), // Get request ID from context
		)

		logger.InfoContext(r.Context(), "Request started")

		// Wrap response writer to capture status code
		wrappedWriter := NewResponseWriterWrapper(w)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		logger.InfoContext(r.Context(), "Request finished",
			"status_code", wrappedWriter.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}