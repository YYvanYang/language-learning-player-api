// ==================================================
// FILE: internal/adapter/handler/http/user_handler_test.go
// ==================================================
package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"      // Import port for interfaces
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use setupHandlerTest if defined elsewhere in package http_test
// Ensure it injects UserID into context if needed
func setupUserHandlerTest(method, path string, body interface{}, userID *domain.UserID) (*http.Request, *httptest.ResponseRecorder) {
    var reqBody *bytes.Buffer = nil
    if body != nil {
        b, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(b)
    }
    req := httptest.NewRequest(method, path, reqBody)
    if reqBody != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    ctx := req.Context()
    // Add UserID to context if provided
    if userID != nil {
        ctx = context.WithValue(ctx, middleware.UserIDKey, *userID)
    }
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    return req, rr
}


func TestUserHandler_GetMyProfile_Success(t *testing.T) {
	mockUserUC := mocks.NewMockUserUseCase(t) // Use mock for port.UserUseCase
	handler := adapter.NewUserHandler(mockUserUC)

	userID := domain.NewUserID()
	emailVO, _ := domain.NewEmail("profile@example.com")
	expectedUser := &domain.User{
		ID:           userID,
		Email:        emailVO,
		Name:         "Profile User",
		AuthProvider: domain.AuthProviderLocal,
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		UpdatedAt:    time.Now().Add(-1 * time.Hour),
	}
	expectedResp := dto.MapDomainUserToResponseDTO(expectedUser)

	// Expect use case to be called
	mockUserUC.On("GetUserProfile", mock.Anything, userID).Return(expectedUser, nil).Once()

	req, rr := setupUserHandlerTest(http.MethodGet, "/api/v1/users/me", nil, &userID)
	handler.GetMyProfile(rr, req)

	// Assert
	require.Equal(t, http.StatusOK, rr.Code)
	var actualResp dto.UserResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &actualResp)
	require.NoError(t, err)
	// Compare timestamps carefully
	assert.Equal(t, expectedResp.ID, actualResp.ID)
	assert.Equal(t, expectedResp.Email, actualResp.Email)
	assert.Equal(t, expectedResp.Name, actualResp.Name)
	assert.Equal(t, expectedResp.AuthProvider, actualResp.AuthProvider)
	assert.Equal(t, expectedResp.ProfileImageURL, actualResp.ProfileImageURL)
	// Check timestamp formatting
	_, err = time.Parse(time.RFC3339, actualResp.CreatedAt)
	assert.NoError(t, err, "CreatedAt should be RFC3339 format")
	_, err = time.Parse(time.RFC3339, actualResp.UpdatedAt)
	assert.NoError(t, err, "UpdatedAt should be RFC3339 format")


	mockUserUC.AssertExpectations(t)
}


func TestUserHandler_GetMyProfile_Unauthorized(t *testing.T) {
	mockUserUC := mocks.NewMockUserUseCase(t)
	handler := adapter.NewUserHandler(mockUserUC)

	// Setup request *without* UserID in context
	req, rr := setupUserHandlerTest(http.MethodGet, "/api/v1/users/me", nil, nil) // Pass nil userID

	handler.GetMyProfile(rr, req)

	// Assert
	require.Equal(t, http.StatusUnauthorized, rr.Code) // Expect 401
	var errResp dto.ErrorResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHENTICATED", errResp.Code)

	mockUserUC.AssertNotCalled(t, "GetUserProfile") // Use case should not be called
}


func TestUserHandler_GetMyProfile_NotFound(t *testing.T) {
	mockUserUC := mocks.NewMockUserUseCase(t)
	handler := adapter.NewUserHandler(mockUserUC)

	userID := domain.NewUserID()

	// Expect use case to return NotFound error
	mockUserUC.On("GetUserProfile", mock.Anything, userID).Return(nil, domain.ErrNotFound).Once()

	req, rr := setupUserHandlerTest(http.MethodGet, "/api/v1/users/me", nil, &userID)
	handler.GetMyProfile(rr, req)

	// Assert
	require.Equal(t, http.StatusNotFound, rr.Code) // Mapped from ErrNotFound
	var errResp dto.ErrorResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", errResp.Code)

	mockUserUC.AssertExpectations(t)
}