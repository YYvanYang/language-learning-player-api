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
// @Summary Request presigned URL for audio upload
// @Description Requests a presigned URL from the object storage (MinIO/S3) that can be used by the client to directly upload an audio file.
// @ID request-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uploadRequest body dto.RequestUploadRequestDTO true "Upload Request Info (filename, content type)"
// @Success 200 {object} dto.RequestUploadResponseDTO "Presigned URL and object key generated"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., failed to generate URL)"
// @Router /uploads/audio/request [post]
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
// @Summary Complete audio upload and create track metadata
// @Description After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database.
// @ID complete-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param completeUpload body dto.CompleteUploadRequestDTO true "Track metadata and object key"
// @Success 201 {object} dto.AudioTrackResponseDTO "Track metadata created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors, object key not found)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., object key already used)" // Depending on use case logic
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [post] // Reuses POST on /audio/tracks conceptually
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