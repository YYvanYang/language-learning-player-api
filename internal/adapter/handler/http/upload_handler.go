// internal/adapter/handler/http/upload_handler.go
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
)

// UploadHandler handles HTTP requests related to file uploads.
type UploadHandler struct {
	uploadUseCase UploadUseCase // Use interface defined below
	validator     *validation.Validator
}

// UploadUseCase defines the methods expected from the use case layer.
// Input Port for this handler.
type UploadUseCase interface {
	RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*usecase.RequestUploadResult, error)
	CompleteUpload(ctx context.Context, userID domain.UserID, req usecase.CompleteUploadRequest) (*domain.AudioTrack, error)
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uc UploadUseCase, v *validation.Validator) *UploadHandler {
	return &UploadHandler{
		uploadUseCase: uc,
		validator:     v,
	}
}

// RequestUpload handles POST /api/v1/uploads/audio/request
// Requires authentication.
func (h *UploadHandler) RequestUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.RequestUploadRequestDTO // Define this DTO below
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	result, err := h.uploadUseCase.RequestUpload(r.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles InvalidArgument, internal errors
		return
	}

	// Map result to response DTO
	resp := dto.RequestUploadResponseDTO{ // Define this DTO below
		UploadURL: result.UploadURL,
		ObjectKey: result.ObjectKey,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteUploadAndCreateTrack handles POST /api/v1/audio/tracks
// This reuses the track creation endpoint conceptually, but specifically for uploaded files.
// Requires authentication.
func (h *UploadHandler) CompleteUploadAndCreateTrack(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CompleteUploadRequestDTO // Define this DTO below
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO to Usecase request struct
	ucReq := usecase.CompleteUploadRequest{
		ObjectKey:     req.ObjectKey,
		Title:         req.Title,
		Description:   req.Description,
		LanguageCode:  req.LanguageCode,
		Level:         req.Level,
		DurationMs:    req.DurationMs,
		IsPublic:      req.IsPublic,
		Tags:          req.Tags,
		CoverImageURL: req.CoverImageURL,
	}

	track, err := h.uploadUseCase.CompleteUpload(r.Context(), userID, ucReq)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles PermissionDenied, InvalidArgument, Conflict, internal errors
		return
	}

	// Return the newly created track details
	resp := dto.MapDomainTrackToResponseDTO(track) // Reuse existing mapping
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created
} 