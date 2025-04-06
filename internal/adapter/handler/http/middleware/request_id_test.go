// =============================================================
// FILE: internal/adapter/handler/http/middleware/request_id_test.go
// =============================================================
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to set request ID in context for testing GetReqID
func SetReqID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, middleware.RequestIDKey, id)
}


func TestRequestID_GeneratesID(t *testing.T) {
	// Dummy next handler
	var capturedReqID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = middleware.GetReqID(r.Context()) // Capture ID from context
		w.WriteHeader(http.StatusOK)
	})

	// Middleware to test
	reqIDMiddleware := middleware.RequestID

	// Prepare request without X-Request-ID header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Serve
	reqIDMiddleware(nextHandler).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check header
	headerID := rr.Header().Get("X-Request-ID")
	require.NotEmpty(t, headerID, "X-Request-ID header should be set")
	_, err := uuid.Parse(headerID) // Check if it looks like a UUID
	assert.NoError(t, err, "Generated ID should be a valid UUID")

	// Check context
	assert.Equal(t, headerID, capturedReqID, "ID in context should match header")
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	var capturedReqID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = middleware.GetReqID(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	reqIDMiddleware := middleware.RequestID

	existingID := "existing-id-123"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existingID) // Set existing ID
	rr := httptest.NewRecorder()

	// Serve
	reqIDMiddleware(nextHandler).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, existingID, rr.Header().Get("X-Request-ID"), "Header should retain existing ID")
	assert.Equal(t, existingID, capturedReqID, "Context should contain existing ID")
}

func TestGetReqID(t *testing.T) {
    // Test getting ID when present
    expectedID := "test-id-from-ctx"
    ctxWithID := SetReqID(context.Background(), expectedID)
    actualID := middleware.GetReqID(ctxWithID)
    assert.Equal(t, expectedID, actualID)

    // Test getting ID when not present
    ctxWithoutID := context.Background()
    actualID = middleware.GetReqID(ctxWithoutID)
    assert.Empty(t, actualID)

    // Test getting ID with wrong type (should return empty)
    ctxWithWrongType := context.WithValue(context.Background(), middleware.RequestIDKey, 123)
    actualID = middleware.GetReqID(ctxWithWrongType)
    assert.Empty(t, actualID)
}