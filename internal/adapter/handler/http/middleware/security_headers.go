// =============================================================
// FILE: internal/adapter/handler/http/middleware/security_headers.go (MODIFIED - Relaxed for Swagger)
// =============================================================
package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// SecurityHeaders adds common security-related HTTP headers to the response.
func SecurityHeaders(next http.Handler) http.Handler {
	// Calculate max-age for HSTS in seconds (e.g., 1 year)
	// Use a smaller value initially during testing (e.g., 5 minutes = 300)
	hstsMaxAge := int(time.Hour * 24 * 365 / time.Second) // 1 year
	hstsHeader := "max-age=" + strconv.Itoa(hstsMaxAge) + "; includeSubDomains; preload"
	// WARNING: Only enable 'preload' if you understand the implications and intend to submit your domain to the HSTS preload list.

	// Define a Content-Security-Policy (CSP)
	// WARNING: Added 'unsafe-inline' for styles and scripts to allow default Swagger UI to work.
	// This reduces protection against XSS. For better security, consider applying this policy
	// only to the /swagger/ path or configuring Swagger UI to avoid inline code.
	cspHeader := "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests;"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()

		// HTTP Strict Transport Security (HSTS)
		// Tells browsers to always connect via HTTPS. Only set this if your site *always* uses HTTPS.
		headers.Set("Strict-Transport-Security", hstsHeader)

		// X-Content-Type-Options
		// Prevents browsers from MIME-sniffing a response away from the declared content-type.
		headers.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options
		// Protects against clickjacking attacks by controlling whether the page can be displayed in a frame/iframe.
		headers.Set("X-Frame-Options", "DENY") // Or "SAMEORIGIN"

		// Content-Security-Policy (CSP)
		headers.Set("Content-Security-Policy", cspHeader)

		// Referrer-Policy (Optional but recommended)
		// Controls how much referrer information is sent with requests.
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (Optional, newer header)
		// Allows fine-grained control over browser features available to the page.
		// Example: Disable microphone and geolocation
		// headers.Set("Permissions-Policy", "microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
}
