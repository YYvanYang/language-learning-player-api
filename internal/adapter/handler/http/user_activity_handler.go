// internal/adapter/handler/http/user_activity_handler.go
package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
	"your_project/internal/adapter/handler/http/dto" // Adjust import path
	"your_project/internal/adapter/handler/http/middleware" // Adjust import path
	"your_project/pkg/httputil"    // Adjust import path
	"your_project/pkg/validation"  // Adjust import path
)

// UserActivityHandler handles HTTP requests related to user progress and bookmarks.
type UserActivityHandler struct {
	activityUseCase UserActivityUseCase // Use interface defined below
	validator       *validation.Validator
}

// UserActivityUseCase defines the methods expected from the use case layer.
// Input Port for this handler.
type UserActivityUseCase interface {
	RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error
	GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
    ListUserProgress(ctx context.Context, userID domain.UserID, page port.Page) ([]*domain.PlaybackProgress, int, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	ListBookmarks(ctx context.Context, userID domain.UserID, trackID *domain.TrackID, page port.Page) ([]*domain.Bookmark, int, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
}

// NewUserActivityHandler creates a new UserActivityHandler.
func NewUserActivityHandler(uc UserActivityUseCase, v *validation.Validator) *UserActivityHandler {
	return &UserActivityHandler{
		activityUseCase: uc,
		validator:       v,
	}
}

// --- Progress Handlers ---

// RecordProgress handles POST /api/v1/users/me/progress
func (h *UserActivityHandler) RecordProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.RecordProgressRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	trackID, err := domain.TrackIDFromString(req.TrackID)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	// Convert float seconds to duration
	progressDuration := time.Duration(req.ProgressSeconds * float64(time.Second))

	err = h.activityUseCase.RecordPlaybackProgress(r.Context(), userID, trackID, progressDuration)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound (track), internal errors
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content is suitable for successful upsert
}

// GetProgress handles GET /api/v1/users/me/progress/{trackId}
func (h *UserActivityHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        httputil.RespondError(w, r, domain.ErrUnauthenticated)
        return
    }

	trackIDStr := chi.URLParam(r, "trackId")
	trackID, err := domain.TrackIDFromString(trackIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	progress, err := h.activityUseCase.GetPlaybackProgress(r.Context(), userID, trackID)
    if err != nil {
		// Handles ErrNotFound correctly via RespondError
        httputil.RespondError(w, r, err)
        return
    }

	resp := dto.MapDomainProgressToResponseDTO(progress)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// ListProgress handles GET /api/v1/users/me/progress
func (h *UserActivityHandler) ListProgress(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        httputil.RespondError(w, r, domain.ErrUnauthenticated)
        return
    }

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
    if limit <= 0 { limit = 50 }
    if offset < 0 { offset = 0 }
	page := port.Page{Limit: limit, Offset: offset}

	progressList, total, err := h.activityUseCase.ListUserProgress(r.Context(), userID, page)
    if err != nil {
        httputil.RespondError(w, r, err)
        return
    }

	respData := make([]dto.PlaybackProgressResponseDTO, len(progressList))
	for i, p := range progressList {
		respData[i] = dto.MapDomainProgressToResponseDTO(p)
	}

	resp := dto.PaginatedProgressResponseDTO{
		Data:   respData,
		Total:  total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// --- Bookmark Handlers ---

// CreateBookmark handles POST /api/v1/bookmarks
func (h *UserActivityHandler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CreateBookmarkRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	trackID, err := domain.TrackIDFromString(req.TrackID)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	timestampDuration := time.Duration(req.TimestampSeconds * float64(time.Second))

	bookmark, err := h.activityUseCase.CreateBookmark(r.Context(), userID, trackID, timestampDuration, req.Note)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound (track), internal errors
		return
	}

	resp := dto.MapDomainBookmarkToResponseDTO(bookmark)
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created
}

// ListBookmarks handles GET /api/v1/bookmarks
func (h *UserActivityHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
    if limit <= 0 { limit = 50 }
    if offset < 0 { offset = 0 }
	page := port.Page{Limit: limit, Offset: offset}

	// Check for optional trackId filter
	var trackIDFilter *domain.TrackID
	if trackIDStr := q.Get("trackId"); trackIDStr != "" {
		tid, err := domain.TrackIDFromString(trackIDStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid trackId query parameter format", domain.ErrInvalidArgument))
			return
		}
		trackIDFilter = &tid
	}


	bookmarks, total, err := h.activityUseCase.ListBookmarks(r.Context(), userID, trackIDFilter, page)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.BookmarkResponseDTO, len(bookmarks))
	for i, b := range bookmarks {
		respData[i] = dto.MapDomainBookmarkToResponseDTO(b)
	}

	resp := dto.PaginatedBookmarksResponseDTO{
		Data:   respData,
		Total:  total,
		Limit:  page.Limit, // Use the potentially adjusted limit from usecase
		Offset: page.Offset,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// DeleteBookmark handles DELETE /api/v1/bookmarks/{bookmarkId}
// @Summary Delete a bookmark
// @Description Deletes a specific bookmark owned by the current user.
// @Tags User Activity
// @Produce json
// @Security BearerAuth // Apply the security definition defined in main.go
// @Param bookmarkId path string true "Bookmark UUID" Format(uuid)
// @Success 204 "Bookmark deleted successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Bookmark Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /bookmarks/{bookmarkId} [delete]
func (h *UserActivityHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	bookmarkIDStr := chi.URLParam(r, "bookmarkId")
	bookmarkID, err := domain.BookmarkIDFromString(bookmarkIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid bookmark ID format", domain.ErrInvalidArgument))
		return
	}

	err = h.activityUseCase.DeleteBookmark(r.Context(), userID, bookmarkID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}