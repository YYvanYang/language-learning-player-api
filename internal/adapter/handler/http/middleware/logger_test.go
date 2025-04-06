// ===========================================================
// FILE: internal/adapter/handler/http/middleware/logger_test.go
// ===========================================================
package middleware_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger(t *testing.T) {
	var logBuffer bytes.Buffer
	testHandler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(testHandler)
	slog.SetDefault(testLogger) // Set as default for the middleware to pick up

	// Dummy next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted) // Set a specific status code
		w.Write([]byte("OK"))
	})

	// Middleware to test
	loggingMiddleware := middleware.RequestLogger

	// Prepare request
	req := httptest.NewRequest(http.MethodPost, "/test/path?query=1", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	// Assuming RequestID middleware runs first and adds ID
	req = req.WithContext(middleware.SetReqID(req.Context(), "req-test-123"))

	rr := httptest.NewRecorder()

	// Serve through middleware
	loggingMiddleware(nextHandler).ServeHTTP(rr, req)

	// Assert response is passed through correctly
	assert.Equal(t, http.StatusAccepted, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	// Assert logs
	logOutput := logBuffer.String()
	t.Logf("Log Output:\n%s", logOutput) // Print logs for debugging if needed

	require.Contains(t, logOutput, "Request started", "Should log request start")
	require.Contains(t, logOutput, "Request finished", "Should log request finish")
	assert.Contains(t, logOutput, `method=POST`)
	assert.Contains(t, logOutput, `path=/test/path`) // Path only, no query
	assert.Contains(t, logOutput, `remote_addr="192.0.2.1:12345"`)
	assert.Contains(t, logOutput, `user_agent="TestAgent/1.0"`)
	assert.Contains(t, logOutput, `request_id=req-test-123`)
	assert.Contains(t, logOutput, `status_code=202`) // Check captured status code
	assert.Contains(t, logOutput, `duration_ms=`)   // Check duration is logged
}