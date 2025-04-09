// ============================================
// FILE: pkg/httputil/response_test.go (MODIFIED)
// ============================================
package httputil_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust
	"github.com/yvanyang/language-learning-player-backend/pkg/apierrors"   // Import constants
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"    // Adjust

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON_Success(t *testing.T) {
	type SamplePayload struct {
		Message string `json:"message"`
		Value   int    `json:"value"`
	}
	payload := SamplePayload{Message: "Success", Value: 123}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	httputil.RespondJSON(rr, req, http.StatusOK, payload)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var actualPayload SamplePayload
	err := json.Unmarshal(rr.Body.Bytes(), &actualPayload)
	require.NoError(t, err)
	assert.Equal(t, payload, actualPayload)
}

func TestRespondJSON_NilPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	httputil.RespondJSON(rr, req, http.StatusNoContent, nil) // Use 204 for no body

	assert.Equal(t, http.StatusNoContent, rr.Code)
	// Content-Type might or might not be set by framework/stdlib when body is nil for 204
	// assert.Empty(t, rr.Header().Get("Content-Type")) // Be flexible here
	assert.Empty(t, rr.Body.String())
}

func TestRespondError_MapsDomainErrors(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		expectedStatus    int
		expectedCode      string // Use constant value here
		expectedMsgSubstr string // Check substring for specific messages
	}{
		{"Not Found", domain.ErrNotFound, http.StatusNotFound, apierrors.CodeNotFound, "resource was not found"},
		{"Conflict", domain.ErrConflict, http.StatusConflict, apierrors.CodeConflict, "conflict occurred"},
		{"Invalid Argument", fmt.Errorf("%w: email format invalid", domain.ErrInvalidArgument), http.StatusBadRequest, apierrors.CodeInvalidInput, "email format invalid"},
		{"Permission Denied", domain.ErrPermissionDenied, http.StatusForbidden, apierrors.CodeForbidden, "permission to perform"},
		{"Rate Limit Denied", fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied), http.StatusTooManyRequests, apierrors.CodeRateLimitExceeded, "Too many requests"}, // Test rate limit mapping
		{"Auth Failed", domain.ErrAuthenticationFailed, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication failed"},
		{"Unauthenticated", domain.ErrUnauthenticated, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication required"},
		{"Internal Error", errors.New("database connection lost"), http.StatusInternalServerError, apierrors.CodeInternalError, "unexpected internal error"},
		{"Wrapped Internal Error", fmt.Errorf("repo failed: %w", errors.New("db timeout")), http.StatusInternalServerError, apierrors.CodeInternalError, "unexpected internal error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqID := "req-" + tc.name
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			// Add request ID to context for testing propagation
			ctx := context.WithValue(req.Context(), httputil.RequestIDKey, reqID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			httputil.RespondError(rr, req, tc.err)

			assert.Equal(t, tc.expectedStatus, rr.Code, "HTTP status code mismatch")
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var errResp httputil.ErrorResponseDTO
			err := json.Unmarshal(rr.Body.Bytes(), &errResp)
			require.NoError(t, err, "Failed to unmarshal error response")

			assert.Equal(t, tc.expectedCode, errResp.Code, "Error code mismatch")
			assert.Contains(t, errResp.Message, tc.expectedMsgSubstr, "Error message mismatch")
			assert.Equal(t, reqID, errResp.RequestID, "Request ID mismatch in response")

			// Ensure internal errors are not leaked for 500 status
			if tc.expectedStatus >= 500 {
				assert.NotContains(t, errResp.Message, tc.err.Error(), "Internal error details should not be exposed")
			}
		})
	}
}

func TestGetReqIDFromContext(t *testing.T) {
	expectedID := "my-request-id"
	ctx := context.WithValue(context.Background(), httputil.RequestIDKey, expectedID)
	reqID := httputil.GetReqID(ctx)
	assert.Equal(t, expectedID, reqID)

	reqID = httputil.GetReqID(context.Background()) // No ID in context
	assert.Empty(t, reqID)
}
