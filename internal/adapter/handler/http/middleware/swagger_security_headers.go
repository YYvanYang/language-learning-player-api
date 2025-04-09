// internal/adapter/handler/http/middleware/swagger_security_headers.go
package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// SwaggerSecurityHeaders adds security headers tailored for Swagger UI compatibility.
// It uses a relaxed Content-Security-Policy allowing 'unsafe-inline'.
func SwaggerSecurityHeaders(next http.Handler) http.Handler {
	// HSTS: Optional for Swagger docs, especially in dev. If used, keep short and no 'preload'.
	hstsMaxAgeSeconds := int(5 * time.Minute / time.Second) // 300 seconds
	hstsHeader := "max-age=" + strconv.Itoa(hstsMaxAgeSeconds) + "; includeSubDomains"

	// Relaxed Content-Security-Policy for Swagger UI
	// Allows inline styles and scripts, and data URIs for images (for icons).
	cspHeader := "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()

		// HTTP Strict Transport Security (HSTS) - Optional for Swagger
		headers.Set("Strict-Transport-Security", hstsHeader)

		// X-Content-Type-Options
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options (Swagger UI might need SAMEORIGIN if embedded, DENY is safer)
		headers.Set("X-Frame-Options", "DENY")

		// Content-Security-Policy (CSP) - Relaxed version
		headers.Set("Content-Security-Policy", cspHeader)

		// Referrer-Policy
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (Generally not critical for Swagger docs)
		// headers.Set("Permissions-Policy", "...")

		next.ServeHTTP(w, r)
	})
}
