// ===================================================
// FILE: internal/adapter/handler/http/upload_handler_test.go
// ===================================================
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Alias
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase" // For usecase types
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use setupHandlerTest if available, ensure UserID injection
func setupUploadHandlerTest(method, path string, body interface{}, userID *domain.UserID) (*http.Request, *httptest.ResponseRecorder) {
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
    if userID != nil {
        ctx = context.WithValue(ctx, middleware.UserIDKey, *userID)
    }
    req = req.WithContext(ctx)
    rr := httptest.NewRecorder()
    return req, rr
}


func TestUploadHandler_RequestUpload_Success(t *testing.T) {
	mockUploadUC := mocks.NewMockUploadUseCase(t) // Use mock for usecase.UploadUseCase
	validator := validation.New()
	handler := adapter.NewUploadHandler(mockUploadUC, validator)

	userID := domain.NewUserID()
	reqBody := dto.RequestUploadRequestDTO{
		Filename:    "audio.mp3",
		ContentType: "audio/mpeg",
	}
	expectedResult := &usecase.RequestUploadResult{
		UploadURL: "http://presigned.url/put",
		ObjectKey: fmt.Sprintf("user-uploads/%s/some-uuid.mp3", userID.String()),
	}

	// Expect use case
	mockUploadUC.On("RequestUpload", mock.Anything, userID, reqBody.Filename, reqBody.ContentType).Return(expectedResult, nil).Once()

	req, rr := setupUploadHandlerTest(http.MethodPost, "/api/v1/uploads/audio/request", reqBody, &userID)
	handler.RequestUpload(rr, req)

	// Assert
	require.Equal(t, http.StatusOK, rr.Code)
	var actualResp dto.RequestUploadResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &actualResp)
	require.NoError(t, err)
	assert.Equal(t, expectedResult.UploadURL, actualResp.UploadURL)
	assert.Equal(t, expectedResult.ObjectKey, actualResp.ObjectKey)

	mockUploadUC.AssertExpectations(t)
}

func TestUploadHandler_RequestUpload_ValidationError(t *testing.T) {
	mockUploadUC := mocks.NewMockUploadUseCase(t)
	validator := validation.New()
	handler := adapter.NewUploadHandler(mockUploadUC, validator)

	userID := domain.NewUserID()
	reqBody := dto.RequestUploadRequestDTO{ /* Filename missing */ ContentType: "audio/mpeg" }

	req, rr := setupUploadHandlerTest(http.MethodPost, "/api/v1/uploads/audio/request", reqBody, &userID)
	handler.RequestUpload(rr, req)

	// Assert
	require.Equal(t, http.StatusBadRequest, rr.Code)
	mockUploadUC.AssertNotCalled(t, "RequestUpload")
}


func TestUploadHandler_CompleteUpload_Success(t *testing.T) {
    mockUploadUC := mocks.NewMockUploadUseCase(t)
    validator := validation.New()
    handler := adapter.NewUploadHandler(mockUploadUC, validator)

    userID := domain.NewUserID()
    objectKey := fmt.Sprintf("user-uploads/%s/completed-upload.mp3", userID)
    reqBody := dto.CompleteUploadRequestDTO{
        ObjectKey:    objectKey,
        Title:        "Completed Track",
        LanguageCode: "es-ES",
        DurationMs:   210000,
        IsPublic:     false,
        Level:        "A2",
    }
    expectedUsecaseReq := usecase.CompleteUploadRequest{ // Map DTO -> Usecase Req
        ObjectKey:     reqBody.ObjectKey,
        Title:         reqBody.Title,
        LanguageCode:  reqBody.LanguageCode,
        Level:         reqBody.Level,
        DurationMs:    reqBody.DurationMs,
        IsPublic:      reqBody.IsPublic,
		// Description, Tags, CoverImageURL are nil/empty in this test case
    }

    // Mock domain object returned by use case
    returnedTrack := &domain.AudioTrack{
        ID:             domain.NewTrackID(),
        Title:          reqBody.Title,
        MinioObjectKey: reqBody.ObjectKey,
		Language: domain.Language{}, // Need to mock creation or setup
        Level: domain.LevelA2,
        Duration: time.Duration(reqBody.DurationMs) * time.Millisecond,
        UploaderID: &userID,
		IsPublic: reqBody.IsPublic,
		MinioBucket: "test-bucket", // Assume from config
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
	// Populate Language VO for the mock return
	returnedTrack.Language, _ = domain.NewLanguage(reqBody.LanguageCode, "")
    expectedResp := dto.MapDomainTrackToResponseDTO(returnedTrack)


    // Expect use case
    mockUploadUC.On("CompleteUpload", mock.Anything, userID, expectedUsecaseReq).Return(returnedTrack, nil).Once()

    req, rr := setupUploadHandlerTest(http.MethodPost, "/api/v1/audio/tracks", reqBody, &userID)
    handler.CompleteUploadAndCreateTrack(rr, req)

    // Assert
    require.Equal(t, http.StatusCreated, rr.Code)
    var actualResp dto.AudioTrackResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &actualResp)
    require.NoError(t, err)
    // Compare timestamps carefully
    assert.WithinDuration(t, expectedResp.CreatedAt, actualResp.CreatedAt, time.Second)
    assert.WithinDuration(t, expectedResp.UpdatedAt, actualResp.UpdatedAt, time.Second)
    actualResp.CreatedAt = expectedResp.CreatedAt // Zero out for direct comparison
    actualResp.UpdatedAt = expectedResp.UpdatedAt
    assert.Equal(t, expectedResp, actualResp)

    mockUploadUC.AssertExpectations(t)
}


func TestUploadHandler_CompleteUpload_ValidationError(t *testing.T) {
    mockUploadUC := mocks.NewMockUploadUseCase(t)
    validator := validation.New()
    handler := adapter.NewUploadHandler(mockUploadUC, validator)

    userID := domain.NewUserID()
    reqBody := dto.CompleteUploadRequestDTO{
        ObjectKey:    "key",
        Title:        "", // Missing title
        LanguageCode: "en",
        DurationMs:   0, // Invalid duration
    }

    req, rr := setupUploadHandlerTest(http.MethodPost, "/api/v1/audio/tracks", reqBody, &userID)
    handler.CompleteUploadAndCreateTrack(rr, req)

    // Assert
    require.Equal(t, http.StatusBadRequest, rr.Code)
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "INVALID_INPUT", errResp.Code)
    assert.Contains(t, errResp.Message, "'title' failed validation on the 'required' rule")
    assert.Contains(t, errResp.Message, "'durationMs' failed validation on the 'gt' rule")

    mockUploadUC.AssertNotCalled(t, "CompleteUpload")
}

func TestUploadHandler_CompleteUpload_UseCaseError(t *testing.T) {
    mockUploadUC := mocks.NewMockUploadUseCase(t)
    validator := validation.New()
    handler := adapter.NewUploadHandler(mockUploadUC, validator)

    userID := domain.NewUserID()
    objectKey := fmt.Sprintf("user-uploads/%s/key.mp3", userID.String())
    reqBody := dto.CompleteUploadRequestDTO{ /* valid data */ ObjectKey: objectKey, Title:"T", LanguageCode:"fr", DurationMs:1000 }
    useCaseError := domain.ErrConflict // Simulate conflict error

    // Expect use case
    mockUploadUC.On("CompleteUpload", mock.Anything, userID, mock.AnythingOfType("usecase.CompleteUploadRequest")).Return(nil, useCaseError).Once()

    req, rr := setupUploadHandlerTest(http.MethodPost, "/api/v1/audio/tracks", reqBody, &userID)
    handler.CompleteUploadAndCreateTrack(rr, req)

    // Assert
    require.Equal(t, http.StatusConflict, rr.Code) // Mapped from ErrConflict
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "RESOURCE_CONFLICT", errResp.Code)

    mockUploadUC.AssertExpectations(t)
}