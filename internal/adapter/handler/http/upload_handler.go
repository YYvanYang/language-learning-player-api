// internal/adapter/handler/http/upload_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	// Assuming module path is updated
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port" // Import port package
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// UploadHandler handles HTTP requests related to file uploads.
type UploadHandler struct {
	uploadUseCase port.UploadUseCase // Use interface from port package
	validator     *validation.Validator
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uc port.UploadUseCase, v *validation.Validator) *UploadHandler {
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

	var req dto.RequestUploadRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case - returns port.RequestUploadResult
	result, err := h.uploadUseCase.RequestUpload(r.Context(), userID, req.Filename, req.ContentType)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map port result to response DTO
	resp := dto.RequestUploadResponseDTO{
		UploadURL: result.UploadURL,
		ObjectKey: result.ObjectKey,
	}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteUploadAndCreateTrack handles POST /api/v1/audio/tracks
// @Summary Complete audio upload and create track metadata (Single File)
// @Description After the client successfully uploads a file using the presigned URL, this endpoint is called to create the corresponding audio track metadata record in the database. Use `/audio/tracks/batch/complete` for batch uploads.
// @ID complete-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param completeUpload body dto.CompleteUploadInputDTO true "Track metadata and object key"
// @Success 201 {object} dto.AudioTrackResponseDTO "Track metadata created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors, object key not found, file not in storage)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Object key mismatch)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., object key already used in DB)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [post]
func (h *UploadHandler) CompleteUploadAndCreateTrack(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.CompleteUploadInputDTO // DTO from adapter layer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO to port.CompleteUploadInput
	portReq := port.CompleteUploadInput{
		ObjectKey:     req.ObjectKey,
		Title:         req.Title,
		Description:   req.Description,
		LanguageCode:  req.LanguageCode,
		Level:         req.Level,
		Duration:      time.Duration(req.DurationMs) * time.Millisecond, // Convert ms to duration
		IsPublic:      req.IsPublic,
		Tags:          req.Tags,
		CoverImageURL: req.CoverImageURL,
	}

	// Call use case with port params
	track, err := h.uploadUseCase.CompleteUpload(r.Context(), userID, portReq)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map domain.AudioTrack to response DTO
	resp := dto.MapDomainTrackToResponseDTO(track) // Mapper converts duration back to ms
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// --- Batch Upload Handlers ---

// RequestBatchUpload handles POST /api/v1/uploads/audio/batch/request
// @Summary Request presigned URLs for batch audio upload
// @Description Requests multiple presigned URLs for uploading several audio files in parallel.
// @ID request-batch-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param batchUploadRequest body dto.BatchRequestUploadInputRequestDTO true "List of files to request URLs for"
// @Success 200 {object} dto.BatchRequestUploadInputResponseDTO "List of generated presigned URLs and object keys, including potential errors per item."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., empty file list)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /uploads/audio/batch/request [post]
func (h *UploadHandler) RequestBatchUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.BatchRequestUploadInputRequestDTO // Use batch DTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO request to port request
	portReq := port.BatchRequestUploadInput{Files: make([]port.BatchRequestUploadInputItem, len(req.Files))}
	for i, f := range req.Files {
		portReq.Files[i] = port.BatchRequestUploadInputItem{
			Filename:    f.Filename,
			ContentType: f.ContentType,
		}
	}

	// Call use case - returns []port.BatchURLResultItem
	results, err := h.uploadUseCase.RequestBatchUpload(r.Context(), userID, portReq)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map port results to response DTO results
	respItems := make([]dto.BatchRequestUploadInputResponseItemDTO, len(results))
	for i, res := range results {
		respItems[i] = dto.BatchRequestUploadInputResponseItemDTO{
			OriginalFilename: res.OriginalFilename,
			ObjectKey:        res.ObjectKey,
			UploadURL:        res.UploadURL,
			Error:            res.Error,
		}
	}
	resp := dto.BatchRequestUploadInputResponseDTO{Results: respItems}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// CompleteBatchUploadAndCreateTracks handles POST /api/v1/audio/tracks/batch/complete
// @Summary Complete batch audio upload and create track metadata
// @Description After clients successfully upload multiple files using presigned URLs, this endpoint is called to create the corresponding audio track metadata records in the database within a single transaction.
// @ID complete-batch-audio-upload
// @Tags Uploads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param batchCompleteUpload body dto.BatchCompleteUploadInputDTO true "List of track metadata and object keys for uploaded files"
// @Success 201 {object} dto.BatchCompleteUploadResponseDTO "Batch processing attempted. Results indicate success/failure per item. If overall transaction succeeded, status is 201."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (e.g., validation errors in items, files not in storage)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Object key mismatch)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict (e.g., duplicate object key during processing)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., transaction failure)"
// @Router /audio/tracks/batch/complete [post]
func (h *UploadHandler) CompleteBatchUploadAndCreateTracks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	var req dto.BatchCompleteUploadInputDTO // Use batch DTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Map DTO request to port request
	portReq := port.BatchCompleteInput{Tracks: make([]port.BatchCompleteItem, len(req.Tracks))}
	for i, t := range req.Tracks {
		portReq.Tracks[i] = port.BatchCompleteItem{
			ObjectKey:     t.ObjectKey,
			Title:         t.Title,
			Description:   t.Description,
			LanguageCode:  t.LanguageCode,
			Level:         t.Level,
			Duration:      time.Duration(t.DurationMs) * time.Millisecond, // Convert ms to duration
			IsPublic:      t.IsPublic,
			Tags:          t.Tags,
			CoverImageURL: t.CoverImageURL,
		}
	}

	// Call use case - returns []port.BatchCompleteResultItem
	results, err := h.uploadUseCase.CompleteBatchUpload(r.Context(), userID, portReq)
	// `err` represents an overall transactional or fundamental failure

	// Map port results to response DTO results
	respItems := make([]dto.BatchCompleteUploadResponseItemDTO, len(results))
	for i, res := range results {
		respItems[i] = dto.BatchCompleteUploadResponseItemDTO{
			ObjectKey: res.ObjectKey,
			Success:   res.Success,
			TrackID:   res.TrackID,
			Error:     res.Error,
		}
	}
	resp := dto.BatchCompleteUploadResponseDTO{Results: respItems}

	if err != nil {
		// Overall transaction or critical pre-check failed
		httputil.RespondError(w, r, err) // Map the overall error
		return
	}

	// Pre-checks passed and transaction committed (even if some DB inserts failed within the rolled-back tx)
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // Use 201 Created
}
