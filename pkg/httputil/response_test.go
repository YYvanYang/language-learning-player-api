package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yvanyang/language-learning-player-api/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-api/pkg/apierrors"   // Adjust import path
)

func TestGetReqID(t *testing.T) {
	reqID := "test-request-123"
	ctxWithID := context.WithValue(context.Background(), RequestIDKey, reqID)
	ctxWithoutID := context.Background()

	// Case 1: Context has ID
	retrievedID := GetReqID(ctxWithID)
	assert.Equal(t, reqID, retrievedID)

	// Case 2: Context does not have ID
	retrievedID = GetReqID(ctxWithoutID)
	assert.Empty(t, retrievedID)

	// Case 3: Context has wrong type
	ctxWithWrongType := context.WithValue(context.Background(), RequestIDKey, 123)
	retrievedID = GetReqID(ctxWithWrongType)
	assert.Empty(t, retrievedID)
}

func TestRespondJSON(t *testing.T) {
	type samplePayload struct {
		Message string `json:"message"`
		Value   int    `json:"value"`
	}
	payload := samplePayload{Message: "Success", Value: 100}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	RespondJSON(rr, req, http.StatusOK, payload)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var responsePayload samplePayload
	err := json.Unmarshal(rr.Body.Bytes(), &responsePayload)
	assert.NoError(t, err)
	assert.Equal(t, payload, responsePayload)
}

func TestRespondJSON_NilPayload(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	RespondJSON(rr, req, http.StatusNoContent, nil)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type")) // Header is still set
	assert.Empty(t, rr.Body.String())
}

// TestRespondJSON_EncodeError requires a type that cannot be marshaled
type unmarshalable struct {
	FuncField func()
}

func TestRespondJSON_EncodeError(t *testing.T) {
	// This test might be flaky depending on internal logging, but tests the principle
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	payload := unmarshalable{FuncField: func() {}}

	RespondJSON(rr, req, http.StatusOK, payload)

	// Expecting 500 because encoding fails after header is written
	assert.Equal(t, http.StatusOK, rr.Code) // Header was already written
	// Body might contain the plain text error or be empty depending on http internals
	assert.Contains(t, rr.Body.String(), "Internal Server Error")
}

func TestMapDomainErrorToHTTP(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string // Base message before potential wrapping
	}{
		{"Not Found", domain.ErrNotFound, http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."},
		{"Conflict", domain.ErrConflict, http.StatusConflict, apierrors.CodeConflict, "A conflict occurred with the current state of the resource."},
		{"Invalid Argument", fmt.Errorf("%w: email is required", domain.ErrInvalidArgument), http.StatusBadRequest, apierrors.CodeInvalidInput, "email is required"}, // Specific message used
		{"Permission Denied", domain.ErrPermissionDenied, http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action."},
		{"Rate Limit", fmt.Errorf("%w: rate limit exceeded", domain.ErrPermissionDenied), http.StatusTooManyRequests, apierrors.CodeRateLimitExceeded, "Too many requests. Please try again later."},
		{"Authentication Failed", domain.ErrAuthenticationFailed, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication failed. Please check your credentials."},
		{"Unauthenticated", domain.ErrUnauthenticated, http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication required. Please log in."},
		{"Wrapped Not Found", fmt.Errorf("specific item not found: %w", domain.ErrNotFound), http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."},
		{"Generic Error", errors.New("some generic error"), http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred."},
		{"Nil Error", nil, http.StatusOK, "", ""}, // Should maybe not be called with nil, but handle gracefully
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// REMOVED the unused 'message' variable declaration and assignment here

			gotStatus, gotCode, gotMessage := MapDomainErrorToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.Equal(t, tt.wantCode, gotCode)
			// Special check for InvalidArgument message
			if errors.Is(tt.err, domain.ErrInvalidArgument) {
				// The MapDomainErrorToHTTP function should return the full error string for ErrInvalidArgument
				assert.Equal(t, tt.err.Error(), gotMessage)
			} else {
				assert.Equal(t, tt.wantMessage, gotMessage)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedCode    string
		expectedMessage string // Expected message IN THE RESPONSE BODY
		includeReqId    bool
	}{
		{"Not Found Error", domain.ErrNotFound, http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found.", true},
		{"Invalid Argument Error", fmt.Errorf("%w: missing field 'name'", domain.ErrInvalidArgument), http.StatusBadRequest, apierrors.CodeInvalidInput, "invalid argument: missing field 'name'", true},
		{"Internal Server Error", errors.New("database connection failed"), http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred.", true}, // Message overwritten for 500
		{"Permission Denied Error", domain.ErrPermissionDenied, http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action.", false},              // Test without reqId in context
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/error", nil)
			reqId := ""
			if tt.includeReqId {
				reqId = "req-" + tt.name
				ctx := context.WithValue(req.Context(), RequestIDKey, reqId)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			RespondError(rr, req, tt.err)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var errResp ErrorResponseDTO
			err := json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCode, errResp.Code)
			assert.Equal(t, tt.expectedMessage, errResp.Message)
			if tt.includeReqId {
				assert.Equal(t, reqId, errResp.RequestID)
			} else {
				assert.Empty(t, errResp.RequestID)
			}
		})
	}
}
