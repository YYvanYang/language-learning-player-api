// =============================================================
// FILE: internal/adapter/handler/http/middleware/ratelimit_test.go
// =============================================================
package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"golang.org/x/time/rate"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit_AllowsRequestsWithinLimit(t *testing.T) {
	// Allow 2 requests per second, burst 1
	limiter := middleware.NewIPRateLimiter(rate.Limit(2), 1)
	rateLimitMiddleware := middleware.RateLimit(limiter)
	// Discard logs during this test
    slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))


	// Dummy next handler
	nextHandlerCalled := 0
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled++
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:12345" // Simulate client IP
	rr := httptest.NewRecorder()

	// Serve first request (burst token)
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "First request should be allowed")
	assert.Equal(t, 1, nextHandlerCalled)

	// Wait for limiter to potentially replenish (shouldn't fully yet)
	// time.Sleep(100 * time.Millisecond) // Not strictly needed if burst > 0

	// Serve second request (should also be allowed if rate is high enough or burst > 1)
	// Since burst is 1 and rate is 2rps (1 per 500ms), this might be blocked shortly after the first.
	// Let's wait a bit less than the replenish time to test the burst.
	rr = httptest.NewRecorder() // Reset recorder
	// If this fails, it might be due to timing, increase burst or wait longer.
	// Let's try immediately after the first one, relying on burst = 1
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
    // With burst 1 and rate 2, the first uses the burst, the second needs to wait 500ms
    // So, immediate second request SHOULD be blocked.
	assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Immediate second request should be blocked with burst 1")
    assert.Equal(t, 1, nextHandlerCalled) // Handler not called again


    // Wait longer than replenish time (500ms)
    time.Sleep(600 * time.Millisecond)
    rr = httptest.NewRecorder()
    rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Third request after waiting should be allowed")
    assert.Equal(t, 2, nextHandlerCalled) // Handler called again

}

func TestRateLimit_BlocksExcessiveRequests(t *testing.T) {
	// Allow 1 request per second, burst 1
	limiter := middleware.NewIPRateLimiter(rate.Limit(1), 1)
	rateLimitMiddleware := middleware.RateLimit(limiter)
    slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	nextHandlerCalled := 0
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled++
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.2:12345"

	// First request (allowed by burst)
	rr := httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, nextHandlerCalled)

	// Second request (immediately after, should be blocked)
	rr = httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code) // Expect 429
	assert.Equal(t, 1, nextHandlerCalled) // Should not have increased

	// Third request (still too soon)
	rr = httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Equal(t, 1, nextHandlerCalled)

    // Wait for token replenish (1 second)
    time.Sleep(1100 * time.Millisecond)

    // Fourth request (should be allowed now)
    rr = httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 2, nextHandlerCalled)
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	// Allow 1 request per second, burst 1
	limiter := middleware.NewIPRateLimiter(rate.Limit(1), 1)
	rateLimitMiddleware := middleware.RateLimit(limiter)
    slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	nextHandlerCalled := 0
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled++
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "192.0.2.10:1111"
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "192.0.2.20:2222"

	// Request from IP 1 (allowed)
	rr1 := httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Equal(t, 1, nextHandlerCalled)

	// Request from IP 2 (allowed, different limiter)
	rr2 := httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, 2, nextHandlerCalled)

	// Second request from IP 1 (blocked)
	rr1 = httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusTooManyRequests, rr1.Code)
	assert.Equal(t, 2, nextHandlerCalled)

	// Second request from IP 2 (blocked)
	rr2 = httptest.NewRecorder()
	rateLimitMiddleware(dummyHandler).ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
	assert.Equal(t, 2, nextHandlerCalled)
}

// TODO: Add tests for X-Real-IP and X-Forwarded-For header handling if needed