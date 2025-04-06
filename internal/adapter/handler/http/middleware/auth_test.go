// ========================================================
// FILE: internal/adapter/handler/http/middleware/auth_test.go
// ========================================================
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticator_Success(t *testing.T) {
	mockSecHelper := mocks.NewMockSecurityHelper(t)
	authMiddleware := middleware.Authenticator(mockSecHelper)

	userID := domain.NewUserID()
	validToken := "valid.jwt.token"

	// Mock VerifyJWT to succeed
	mockSecHelper.On("VerifyJWT", mock.Anything, validToken).Return(userID, nil).Once()

	// Dummy next handler
	nextHandlerCalled := false
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if UserID is in context
		ctxUserID, ok := middleware.GetUserIDFromContext(r.Context())
		assert.True(t, ok, "UserID should be in context")
		assert.Equal(t, userID, ctxUserID, "UserID in context should match")
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	// Serve through middleware
	authMiddleware(dummyHandler).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code, "Next handler should be called")
	assert.True(t, nextHandlerCalled, "Next handler should be called")
	mockSecHelper.AssertExpectations(t)
}

func TestAuthenticator_NoHeader(t *testing.T) {
	mockSecHelper := mocks.NewMockSecurityHelper(t)
	authMiddleware := middleware.Authenticator(mockSecHelper)

	// Dummy next handler (should not be called)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Next handler should not be called")
	})

	// Prepare request without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	// Serve
	authMiddleware(dummyHandler).ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockSecHelper.AssertNotCalled(t, "VerifyJWT")
}

func TestAuthenticator_InvalidHeaderFormat(t *testing.T) {
	mockSecHelper := mocks.NewMockSecurityHelper(t)
	authMiddleware := middleware.Authenticator(mockSecHelper)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("Next handler should not be called") })

	testCases := []string{
		"Bearer",          // Missing token
		"Bearer ",         // Missing token
		"Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==", // Wrong scheme
		"InvalidToken",    // No scheme
	}

	for _, header := range testCases {
		t.Run(header, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", header)
			rr := httptest.NewRecorder()
			authMiddleware(dummyHandler).ServeHTTP(rr, req)
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			mockSecHelper.AssertNotCalled(t, "VerifyJWT")
		})
	}
}

func TestAuthenticator_VerifyJWTErrors(t *testing.T) {
	mockSecHelper := mocks.NewMockSecurityHelper(t)
	authMiddleware := middleware.Authenticator(mockSecHelper)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("Next handler should not be called") })

	testCases := []struct {
		name          string
		token         string
		mockError     error
		expectedStatus int
	}{
		{"Expired Token", "expired.token", fmt.Errorf("%w: token expired", domain.ErrAuthenticationFailed), http.StatusUnauthorized},
		{"Invalid Signature", "invalid.sig.token", fmt.Errorf("%w: invalid signature", domain.ErrAuthenticationFailed), http.StatusUnauthorized},
		{"Malformed Token", "malformed", fmt.Errorf("%w: malformed", domain.ErrAuthenticationFailed), http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock for each test case
			mockSecHelper = mocks.NewMockSecurityHelper(t) // Recreate mock
            authMiddleware = middleware.Authenticator(mockSecHelper) // Recreate middleware with new mock

			mockSecHelper.On("VerifyJWT", mock.Anything, tc.token).Return(domain.UserID{}, tc.mockError).Once()

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rr := httptest.NewRecorder()

			authMiddleware(dummyHandler).ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			mockSecHelper.AssertExpectations(t)
		})
	}
}