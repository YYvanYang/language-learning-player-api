// ==================================================
// FILE: internal/adapter/handler/http/user_activity_handler_test.go
// ==================================================
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5"
	adapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http" // Use alias
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"      // Import port for interfaces
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"

	"github.com/google/uuid" // For generating test UUIDs
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use setupHandlerTest if defined elsewhere in package http_test
func setupActivityHandlerTest(method, path string, body interface{}, userID *domain.UserID, pathParams map[string]string) (*http.Request, *httptest.ResponseRecorder) {
    var reqBody *bytes.Buffer = nil
    if body != nil {
        b, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(b)
    }

    req := httptest.NewRequest(method, path, reqBody)
    if reqBody != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    // --- Add Chi URL parameter context ---
    if len(pathParams) > 0 {
        rctx := chi.NewRouteContext()
        for key, value := range pathParams {
            rctx.URLParams.Add(key, value)
        }
        req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
    }
    // -------------------------------------

    ctx := req.Context()
    // Add UserID to context if provided
    if userID != nil {
        ctx = context.WithValue(ctx, middleware.UserIDKey, *userID)
    }
    req = req.WithContext(ctx)

    rr := httptest.NewRecorder()
    return req, rr
}


func TestUserActivityHandler_RecordProgress_Success(t *testing.T) {
	mockActivityUC := mocks.NewMockUserActivityUseCase(t)
	validator := validation.New()
	handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progressMs := int64(30500) // 30.5 seconds in milliseconds
	expectedDuration := time.Duration(progressMs) * time.Millisecond

	reqBody := dto.RecordProgressRequestDTO{
		TrackID:    trackID.String(),
		ProgressMs: progressMs,
	}

	// Expect use case to be called
	mockActivityUC.On("RecordPlaybackProgress", mock.Anything, userID, trackID, expectedDuration).Return(nil).Once()

	req, rr := setupActivityHandlerTest(http.MethodPost, "/api/v1/users/me/progress", reqBody, &userID, nil)
	handler.RecordProgress(rr, req)

	// Assert
	require.Equal(t, http.StatusNoContent, rr.Code) // Expect 204
	mockActivityUC.AssertExpectations(t)
}

func TestUserActivityHandler_RecordProgress_ValidationError(t *testing.T) {
	mockActivityUC := mocks.NewMockUserActivityUseCase(t)
	validator := validation.New()
	handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

	userID := domain.NewUserID()
	reqBody := dto.RecordProgressRequestDTO{
		TrackID:    "invalid-uuid", // Invalid UUID
		ProgressMs: -100,         // Invalid progress
	}

	req, rr := setupActivityHandlerTest(http.MethodPost, "/api/v1/users/me/progress", reqBody, &userID, nil)
	handler.RecordProgress(rr, req)

	// Assert
	require.Equal(t, http.StatusBadRequest, rr.Code)
	var errResp dto.ErrorResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_INPUT", errResp.Code)
	assert.Contains(t, errResp.Message, "'trackId' failed validation on the 'uuid' rule")
	assert.Contains(t, errResp.Message, "'progressMs' failed validation on the 'gte' rule")

	mockActivityUC.AssertNotCalled(t, "RecordPlaybackProgress")
}

func TestUserActivityHandler_RecordProgress_UseCaseError(t *testing.T) {
	mockActivityUC := mocks.NewMockUserActivityUseCase(t)
	validator := validation.New()
	handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progressMs := int64(30000)
	expectedDuration := time.Duration(progressMs) * time.Millisecond
	reqBody := dto.RecordProgressRequestDTO{ TrackID: trackID.String(), ProgressMs: progressMs }
	useCaseError := domain.ErrNotFound // Simulate track not found

	// Expect use case to return error
	mockActivityUC.On("RecordPlaybackProgress", mock.Anything, userID, trackID, expectedDuration).Return(useCaseError).Once()

	req, rr := setupActivityHandlerTest(http.MethodPost, "/api/v1/users/me/progress", reqBody, &userID, nil)
	handler.RecordProgress(rr, req)

	// Assert
	require.Equal(t, http.StatusNotFound, rr.Code) // Mapped from ErrNotFound
	var errResp dto.ErrorResponseDTO
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", errResp.Code)

	mockActivityUC.AssertExpectations(t)
}


func TestUserActivityHandler_GetProgress_Success(t *testing.T) {
    mockActivityUC := mocks.NewMockUserActivityUseCase(t)
    validator := validation.New()
    handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

    userID := domain.NewUserID()
    trackID := domain.NewTrackID()
    expectedProgress := &domain.PlaybackProgress{
        UserID:         userID,
        TrackID:        trackID,
        Progress:       45 * time.Second,
        LastListenedAt: time.Now().Add(-1 * time.Hour),
    }
    expectedResp := dto.MapDomainProgressToResponseDTO(expectedProgress)

    // Expect use case
    mockActivityUC.On("GetPlaybackProgress", mock.Anything, userID, trackID).Return(expectedProgress, nil).Once()

    // Setup request with URL param
    pathParams := map[string]string{"trackId": trackID.String()}
    req, rr := setupActivityHandlerTest(http.MethodGet, fmt.Sprintf("/api/v1/users/me/progress/%s", trackID.String()), nil, &userID, pathParams)

    handler.GetProgress(rr, req)

    // Assert
    require.Equal(t, http.StatusOK, rr.Code)
    var actualResp dto.PlaybackProgressResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &actualResp)
    require.NoError(t, err)
    // Compare timestamps carefully
    assert.WithinDuration(t, expectedResp.LastListenedAt, actualResp.LastListenedAt, time.Second)
    actualResp.LastListenedAt = expectedResp.LastListenedAt // Zero out for direct comparison
    assert.Equal(t, expectedResp, actualResp)

    mockActivityUC.AssertExpectations(t)
}

func TestUserActivityHandler_GetProgress_InvalidTrackID(t *testing.T) {
    mockActivityUC := mocks.NewMockUserActivityUseCase(t)
    validator := validation.New()
    handler := adapter.NewUserActivityHandler(mockActivityUC, validator)
    userID := domain.NewUserID()

    // Setup request with invalid URL param
    pathParams := map[string]string{"trackId": "not-a-uuid"}
    req, rr := setupActivityHandlerTest(http.MethodGet, "/api/v1/users/me/progress/not-a-uuid", nil, &userID, pathParams)

    handler.GetProgress(rr, req)

    // Assert
    require.Equal(t, http.StatusBadRequest, rr.Code)
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "INVALID_INPUT", errResp.Code)
    assert.Contains(t, errResp.Message, "invalid track ID format")

    mockActivityUC.AssertNotCalled(t, "GetPlaybackProgress")
}


// --- Bookmark Handler Tests ---

func TestUserActivityHandler_CreateBookmark_Success(t *testing.T) {
    mockActivityUC := mocks.NewMockUserActivityUseCase(t)
    validator := validation.New()
    handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

    userID := domain.NewUserID()
    trackID := domain.NewTrackID()
    timestampMs := int64(60500) // 60.5 seconds
    expectedDuration := time.Duration(timestampMs) * time.Millisecond
    note := "Bookmark note"

    reqBody := dto.CreateBookmarkRequestDTO{
        TrackID:     trackID.String(),
        TimestampMs: timestampMs,
        Note:        note,
    }

    // Mock domain object returned by use case
    returnedBookmark := &domain.Bookmark{
        ID:        domain.NewBookmarkID(),
        UserID:    userID,
        TrackID:   trackID,
        Timestamp: expectedDuration,
        Note:      note,
        CreatedAt: time.Now(),
    }
    expectedResp := dto.MapDomainBookmarkToResponseDTO(returnedBookmark)

    // Expect use case
    mockActivityUC.On("CreateBookmark", mock.Anything, userID, trackID, expectedDuration, note).Return(returnedBookmark, nil).Once()

    req, rr := setupActivityHandlerTest(http.MethodPost, "/api/v1/bookmarks", reqBody, &userID, nil)
    handler.CreateBookmark(rr, req)

    // Assert
    require.Equal(t, http.StatusCreated, rr.Code)
    var actualResp dto.BookmarkResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &actualResp)
    require.NoError(t, err)
    // Compare timestamps carefully
    assert.WithinDuration(t, expectedResp.CreatedAt, actualResp.CreatedAt, time.Second)
    actualResp.CreatedAt = expectedResp.CreatedAt // Zero out for direct comparison
    assert.Equal(t, expectedResp, actualResp)

    mockActivityUC.AssertExpectations(t)
}

func TestUserActivityHandler_DeleteBookmark_Success(t *testing.T) {
    mockActivityUC := mocks.NewMockUserActivityUseCase(t)
    validator := validation.New()
    handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

    userID := domain.NewUserID()
    bookmarkID := domain.NewBookmarkID()

    // Expect use case
    mockActivityUC.On("DeleteBookmark", mock.Anything, userID, bookmarkID).Return(nil).Once()

    // Setup request with URL param
    pathParams := map[string]string{"bookmarkId": bookmarkID.String()}
    req, rr := setupActivityHandlerTest(http.MethodDelete, fmt.Sprintf("/api/v1/bookmarks/%s", bookmarkID.String()), nil, &userID, pathParams)

    handler.DeleteBookmark(rr, req)

    // Assert
    require.Equal(t, http.StatusNoContent, rr.Code)
    mockActivityUC.AssertExpectations(t)
}

func TestUserActivityHandler_DeleteBookmark_UseCaseError(t *testing.T) {
    mockActivityUC := mocks.NewMockUserActivityUseCase(t)
    validator := validation.New()
    handler := adapter.NewUserActivityHandler(mockActivityUC, validator)

    userID := domain.NewUserID()
    bookmarkID := domain.NewBookmarkID()
    useCaseError := domain.ErrPermissionDenied // Simulate permission denied

    // Expect use case to return error
    mockActivityUC.On("DeleteBookmark", mock.Anything, userID, bookmarkID).Return(useCaseError).Once()

    pathParams := map[string]string{"bookmarkId": bookmarkID.String()}
    req, rr := setupActivityHandlerTest(http.MethodDelete, fmt.Sprintf("/api/v1/bookmarks/%s", bookmarkID.String()), nil, &userID, pathParams)
    handler.DeleteBookmark(rr, req)

    // Assert
    require.Equal(t, http.StatusForbidden, rr.Code) // Mapped from ErrPermissionDenied
    var errResp dto.ErrorResponseDTO
    err := json.Unmarshal(rr.Body.Bytes(), &errResp)
    require.NoError(t, err)
    assert.Equal(t, "FORBIDDEN", errResp.Code)

    mockActivityUC.AssertExpectations(t)
}

// TODO: Add tests for ListProgress, ListBookmarks (success, errors)
// TODO: Add validation error tests for CreateBookmark, DeleteBookmark (invalid UUID)