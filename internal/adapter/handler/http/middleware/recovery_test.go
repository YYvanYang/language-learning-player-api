// ============================================================
// FILE: internal/adapter/handler/http/middleware/recovery_test.go
// ============================================================
package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil" // For ErrorResponseDTO
)

func TestRecoverer_Panic(t *testing.T) {
	var logBuffer bytes.Buffer
	testHandler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(testHandler)
	slog.SetDefault(testLogger) // Set as default for the middleware to pick up

	// Handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went very wrong")
	})

	// Middleware to test
	recoverMiddleware := middleware.Recoverer

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	// Add request ID for logging check
	req = req.WithContext(SetReqID(req.Context(), "req-panic-456"))
	rr := httptest.NewRecorder()

	// Serve through middleware
	recoverMiddleware(panicHandler).ServeHTTP(rr, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected 500 status code")

	var errResp httputil.ErrorResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err, "Failed to unmarshal error response")
	assert.Equal(t, "INTERNAL_ERROR", errResp.Code)
	assert.Equal(t, "An unexpected internal error occurred.", errResp.Message)
	assert.Equal(t, "req-panic-456", errResp.RequestID) // Check Request ID propagation

	// Assert logs
	logOutput := logBuffer.String()
	t.Logf("Log Output:\n%s", logOutput)
	assert.Contains(t, logOutput, "Panic recovered", "Should log panic recovery")
	assert.Contains(t, logOutput, `error="something went very wrong"`)
	assert.Contains(t, logOutput, `request_id=req-panic-456`)
	assert.Contains(t, logOutput, "stack=", "Should log stack trace")
}

func TestRecoverer_NoPanic(t *testing.T) {
	var logBuffer bytes.Buffer
	testHandler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(testHandler)
	slog.SetDefault(testLogger)

	// Handler that does not panic
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	recoverMiddleware := middleware.Recoverer
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rr := httptest.NewRecorder()

	// Serve
	recoverMiddleware(okHandler).ServeHTTP(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())

	// Assert logs
	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "Panic recovered", "Should not log panic recovery")
}
