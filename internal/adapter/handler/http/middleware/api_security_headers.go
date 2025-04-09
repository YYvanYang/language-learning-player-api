// internal/adapter/handler/http/middleware/api_security_headers.go
package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// ApiSecurityHeaders adds common security-related HTTP headers tailored for API responses.
// It uses a strict Content-Security-Policy.
func ApiSecurityHeaders(next http.Handler) http.Handler {
	// Use a shorter HSTS max-age during testing/initial deployment (e.g., 5 minutes = 300)
	// Remove 'preload' until ready to submit to HSTS preload list.
	hstsMaxAgeSeconds := int(5 * time.Minute / time.Second) // 300 seconds
	hstsHeader := "max-age=" + strconv.Itoa(hstsMaxAgeSeconds) + "; includeSubDomains"

	// Strict Content-Security-Policy for APIs (prevents loading external resources, inline scripts/styles)
	// Adjust 'connect-src' if your API needs to make requests to other origins from client-side JS (unlikely for pure API)
	cspHeader := "default-src 'self'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()

		// HTTP Strict Transport Security (HSTS)
		headers.Set("Strict-Transport-Security", hstsHeader)

		// X-Content-Type-Options
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options (Good practice even for APIs, belt-and-suspenders)
		headers.Set("X-Frame-Options", "DENY")

		// Content-Security-Policy (CSP) - Strict version
		headers.Set("Content-Security-Policy", cspHeader)

		// Referrer-Policy
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (Example: disable features not needed by API)
		headers.Set("Permissions-Policy", "microphone=(), geolocation=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
