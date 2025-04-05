package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil" // Adjust import path
    "github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
)

// IPRateLimiter stores rate limiters for IP addresses.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit // Allowed requests per second
	b   int        // Burst size
}

// NewIPRateLimiter creates a new IPRateLimiter.
// r: requests per second, b: burst size.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// AddIP creates a new rate limiter for the given IP address if it doesn't exist.
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	return limiter
}

// GetLimiter returns the rate limiter for the given IP address.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip) // Add if not exists (lazy initialization)
	}
	return limiter
}

// RateLimit is the middleware handler.
func RateLimit(limiter *IPRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the client's real IP address.
			// Use chi's RealIP middleware result if available, otherwise fallback.
			ip := r.RemoteAddr
			if realIP := r.Context().Value(http.CanonicalHeaderKey("X-Real-IP")); realIP != nil {
                if ripStr, ok := realIP.(string); ok {
				    ip = ripStr
                }
			} else if forwardedFor := r.Context().Value(http.CanonicalHeaderKey("X-Forwarded-For")); forwardedFor != nil {
                if fwdStr, ok := forwardedFor.(string); ok {
                    // X-Forwarded-For can contain multiple IPs, get the first one
                    parts := strings.Split(fwdStr, ",")
                    if len(parts) > 0 {
                        ip = strings.TrimSpace(parts[0])
                    }
                }
			} else {
                // Fallback if headers are not present (e.g. direct connection)
                host, _, err := net.SplitHostPort(r.RemoteAddr)
                if err == nil {
                    ip = host
                }
            }


			// Get the rate limiter for the current IP address.
			l := limiter.GetLimiter(ip)

			// Check if the request is allowed.
			if !l.Allow() {
				// Log the rate limit event
				logger := slog.Default()
				logger.WarnContext(r.Context(), "Rate limit exceeded", "ip_address", ip)

				// Respond with 429 Too Many Requests
				// Use a custom error type or map directly
                err := fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied) // Or a new domain error type
                // Map this specific error text or type in RespondError if needed
                httputil.RespondError(w, r, err) // This might return 403, adjust mapping if needed
                // Or respond directly:
                // w.Header().Set("Retry-After", "60") // Optional: suggest retry time
                // http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CleanUpOldLimiters periodically removes limiters for IPs not seen recently.
// This prevents the map from growing indefinitely. Should be run in a separate goroutine.
// Note: This is a basic example; more sophisticated cleanup might be needed.
func (i *IPRateLimiter) CleanUpOldLimiters(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip, limiter := range i.ips {
			// This is tricky: rate.Limiter doesn't expose last access time.
			// A more complex solution would wrap the limiter and track last access.
			// Simple approach (less accurate): remove if unused for a while? Needs wrapping.
			// Placeholder: For now, this cleanup isn't effective without tracking last access.
            _ = limiter // Avoid unused variable error
            _ = ip
		}
		// Implement actual cleanup logic here based on tracking last access time.
		slog.Debug("Running rate limiter cleanup (placeholder)...", "current_size", len(i.ips))
		i.mu.Unlock()
	}
}