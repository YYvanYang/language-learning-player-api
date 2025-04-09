// internal/adapter/handler/http/middleware/ratelimit.go
package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"golang.org/x/time/rate"
)

// limiterEntry stores the limiter and the last seen time.
// Point 7: Added struct to track lastSeen
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter stores rate limiters for IP addresses.
type IPRateLimiter struct {
	ips map[string]*limiterEntry // Point 7: Changed map value type
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IPRateLimiter.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips: make(map[string]*limiterEntry),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	// Point 7: Start cleanup goroutine
	go limiter.CleanUpOldLimiters(10*time.Minute, 30*time.Minute) // Example: cleanup every 10min, remove if idle > 30min

	return limiter
}

// Point 7: Modified AddIP to use limiterEntry
func (i *IPRateLimiter) addIP(ip string) *limiterEntry {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Double check existence inside lock
	entry, exists := i.ips[ip]
	if !exists {
		entry = &limiterEntry{
			limiter:  rate.NewLimiter(i.r, i.b),
			lastSeen: time.Now(),
		}
		i.ips[ip] = entry
		slog.Debug("Added new rate limiter entry", "ip_address", ip) // Added debug log
	}
	return entry
}

// Point 7: Modified GetLimiter to return limiter and update lastSeen
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		entry = i.addIP(ip) // addIP handles locking
	} else {
		// Update lastSeen without full lock if possible, but requires locking for write.
		// Simple approach: Lock for update.
		i.mu.Lock()
		entry.lastSeen = time.Now()
		i.mu.Unlock()
	}

	return entry.limiter
}

// RateLimit is the middleware handler.
func RateLimit(limiter *IPRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP logic remains the same
			ip := r.RemoteAddr
			if realIP := r.Context().Value(http.CanonicalHeaderKey("X-Real-IP")); realIP != nil {
				if ripStr, ok := realIP.(string); ok {
					ip = ripStr
				}
			} else if forwardedFor := r.Context().Value(http.CanonicalHeaderKey("X-Forwarded-For")); forwardedFor != nil {
				if fwdStr, ok := forwardedFor.(string); ok {
					parts := strings.Split(fwdStr, ",")
					if len(parts) > 0 {
						ip = strings.TrimSpace(parts[0])
					}
				}
			} else {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					ip = host
				}
			}

			// Get the rate limiter for the current IP address.
			l := limiter.GetLimiter(ip) // This now also updates lastSeen

			if !l.Allow() {
				logger := slog.Default()
				logger.WarnContext(r.Context(), "Rate limit exceeded", "ip_address", ip, "request_id", httputil.GetReqID(r.Context()))

				// Use specific rate limit error message
				err := fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied)
				httputil.RespondError(w, r, err) // httputil maps this to 429
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Point 7: Implemented CleanUpOldLimiters logic
func (i *IPRateLimiter) CleanUpOldLimiters(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		count := 0
		now := time.Now()
		for ip, entry := range i.ips {
			if now.Sub(entry.lastSeen) > maxAge {
				delete(i.ips, ip)
				count++
			}
		}
		if count > 0 {
			slog.Debug("Rate limiter cleanup removed entries", "removed_count", count, "current_size", len(i.ips))
		}
		i.mu.Unlock()
	}
}
