// internal/adapter/handler/http/auth_handler_test.go
package http_test // Use _test package

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5" // Need chi for routing context if using URL params
	adapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias for handler package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	ucmocks "github.com/yvanyang/language-learning-player-backend/internal/usecase/mocks" // Mocks for usecase interfaces
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper to create a request and response recorder for handler tests
func setupHandlerTest(method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
    var reqBody *bytes.Buffer = nil
    if body != nil {
        b, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(b)
    }

    req := httptest.NewRequest(method, path, reqBody)
    if reqBody != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    // Add context if needed (e.g., for middleware values like UserID)
    // ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID)
    // req = req.WithContext(ctx)

    rr := httptest.NewRecorder()
    return req, rr
}


func TestAuthHandler_Register_Success(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t) // Using mock for the usecase INPUT PORT interface
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator) // Instantiate the actual handler

    reqBody := dto.RegisterRequestDTO{
        Email:    "new@example.com",
        Password: "password123",
        Name:     "New User",
    }
    expectedToken := "new_jwt_token"
    dummyUser := &domain.User{ID: domain.NewUserID(), Email: domain.Email{}} // Dummy user for return

    // Expect the usecase method to be called
    mockAuthUC.On("RegisterWithPassword", mock.Anything, reqBody.Email, reqBody.Password, reqBody.Name).
        Return(dummyUser, expectedToken, nil). // Return success
        Once()

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)

    // Serve the request
    handler.Register(rr, req)

    // Assert the response
    require.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")

    var respBody dto.AuthResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &respBody)
    require.NoError(t, err, "Failed to unmarshal response body")
    assert.Equal(t, expectedToken, respBody.Token)

    mockAuthUC.AssertExpectations(t) // Verify usecase was called
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t)
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator)

    reqBody := dto.RegisterRequestDTO{ // Invalid email
        Email:    "invalid-email",
        Password: "password123",
        Name:     "New User",
    }

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)
    handler.Register(rr, req)

    // Assert validation error response
    require.Equal(t, http.StatusBadRequest, rr.Code)
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "INVALID_INPUT", errResp.Code)
    assert.Contains(t, errResp.Message, "'email' failed validation")

    // Ensure usecase was NOT called
    mockAuthUC.AssertNotCalled(t, "RegisterWithPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Register_UseCaseConflictError(t *testing.T) {
    mockAuthUC := ucmocks.NewMockAuthUseCase(t)
    validator := validation.New()
    handler := adapter.NewAuthHandler(mockAuthUC, validator)

    reqBody := dto.RegisterRequestDTO{
        Email:    "exists@example.com",
        Password: "password123",
        Name:     "Existing User",
    }
    conflictError := fmt.Errorf("%w: email already exists", domain.ErrConflict)

    // Expect usecase to be called and return conflict error
    mockAuthUC.On("RegisterWithPassword", mock.Anything, reqBody.Email, reqBody.Password, reqBody.Name).
        Return(nil, "", conflictError).
        Once()

    req, rr := setupHandlerTest(http.MethodPost, "/api/v1/auth/register", reqBody)
    handler.Register(rr, req)

    // Assert conflict response
    require.Equal(t, http.StatusConflict, rr.Code) // 409 Conflict
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "RESOURCE_CONFLICT", errResp.Code)

    mockAuthUC.AssertExpectations(t)
}

// TODO: Add tests for Login handler (success, auth failed, validation error)
// TODO: Add tests for GoogleCallback handler (success new user, success existing, failure)