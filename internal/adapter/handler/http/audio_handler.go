// internal/adapter/handler/http/audio_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
)

// AudioHandler handles HTTP requests related to audio tracks and collections.
type AudioHandler struct {
	audioUseCase port.AudioContentUseCase // 使用port包中定义的接口
	validator    *validation.Validator
}

// NewAudioHandler creates a new AudioHandler.
func NewAudioHandler(uc port.AudioContentUseCase, v *validation.Validator) *AudioHandler {
	return &AudioHandler{
		audioUseCase: uc,
		validator:    v,
	}
}

// --- Track Handlers ---

// GetTrackDetails handles GET /api/v1/audio/tracks/{trackId}
// @Summary Get audio track details
// @Description Retrieves details for a specific audio track, including metadata and a temporary playback URL.
// @ID get-track-details
// @Tags Audio Tracks
// @Produce json
// @Param trackId path string true "Audio Track UUID" Format(uuid)
// @Success 200 {object} dto.AudioTrackDetailsResponseDTO "Audio track details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format"
// @Failure 404 {object} httputil.ErrorResponseDTO "Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks/{trackId} [get]
func (h *AudioHandler) GetTrackDetails(w http.ResponseWriter, r *http.Request) {
	trackIDStr := chi.URLParam(r, "trackId")
	trackID, err := domain.TrackIDFromString(trackIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format", domain.ErrInvalidArgument))
		return
	}

	track, playURL, err := h.audioUseCase.GetAudioTrackDetails(r.Context(), trackID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, internal errors
		return
	}

	resp := dto.AudioTrackDetailsResponseDTO{
		AudioTrackResponseDTO: dto.MapDomainTrackToResponseDTO(track),
		PlayURL:               playURL,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// ListTracks handles GET /api/v1/audio/tracks
// @Summary List audio tracks
// @Description Retrieves a paginated list of audio tracks, supporting filtering and sorting.
// @ID list-audio-tracks
// @Tags Audio Tracks
// @Produce json
// @Param q query string false "Search query (searches title, description)"
// @Param lang query string false "Filter by language code (e.g., en-US)"
// @Param level query string false "Filter by audio level (e.g., A1, B2)" Enums(A1, A2, B1, B2, C1, C2, NATIVE)
// @Param isPublic query boolean false "Filter by public status (true or false)"
// @Param tags query []string false "Filter by tags (e.g., ?tags=news&tags=podcast)" collectionFormat(multi)
// @Param sortBy query string false "Sort field (e.g., createdAt, title, durationMs)" default(createdAt)
// @Param sortDir query string false "Sort direction (asc or desc)" default(desc) Enums(asc, desc)
// @Param limit query int false "Pagination limit" default(20) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedTracksResponseDTO "Paginated list of audio tracks"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Query Parameter Format"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [get]
func (h *AudioHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters manually or using a library like schema
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	// Let the usecase handle defaults and constraints
	// Remove old page creation: // if limit <= 0 { limit = 20 }
    // Remove old page creation: // if offset < 0 { offset = 0 }
	// Remove old page creation: // page := port.Page{Limit: limit, Offset: offset}

	params := port.ListTracksParams{
		SortBy:        q.Get("sortBy"),
		SortDirection: q.Get("sortDir"),
		Tags:          q["tags"], // Get array of tags ?tags=a&tags=b
	}
	// Assign optional filters if present
	if query := q.Get("q"); query != "" { params.Query = &query }
	if lang := q.Get("lang"); lang != "" { params.LanguageCode = &lang }
	if levelStr := q.Get("level"); levelStr != "" {
		level := domain.AudioLevel(levelStr)
		if level.IsValid() { // Validate level input
			params.Level = &level
		} else {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid level query parameter", domain.ErrInvalidArgument))
			return
		}
	}
	if isPublicStr := q.Get("isPublic"); isPublicStr != "" {
		isPublic, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			params.IsPublic = &isPublic
		} else {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid isPublic query parameter (must be true or false)", domain.ErrInvalidArgument))
			return
		}
	}

	// Call use case with raw limit/offset, expect pageInfo back
	tracks, total, pageInfo, err := h.audioUseCase.ListTracks(r.Context(), params, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.AudioTrackResponseDTO, len(tracks))
	for i, track := range tracks {
		respData[i] = dto.MapDomainTrackToResponseDTO(track)
	}

	// Create paginated response DTO using the helper from pkg/pagination
	paginatedResult := pagination.NewPaginatedResponse(respData, total, pageInfo)

	// Use the generic PaginatedResponseDTO from the DTO package
	resp := dto.PaginatedResponseDTO{
		Data:       paginatedResult.Data,
		Total:      paginatedResult.Total,
		Limit:      paginatedResult.Limit,
		Offset:     paginatedResult.Offset,
		Page:       paginatedResult.Page,
		TotalPages: paginatedResult.TotalPages,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}


// --- Collection Handlers ---

// CreateCollection handles POST /api/v1/audio/collections
// @Summary Create an audio collection
// @Description Creates a new audio collection (playlist or course) for the authenticated user.
// @ID create-audio-collection
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collection body dto.CreateCollectionRequestDTO true "Collection details"
// @Success 201 {object} dto.AudioCollectionResponseDTO "Collection created successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Track ID Format / Collection Type"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections [post]
func (h *AudioHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCollectionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Convert string IDs to domain IDs
	initialTrackIDs := make([]domain.TrackID, 0, len(req.InitialTrackIDs))
	for _, idStr := range req.InitialTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid initial track ID format '%s'", domain.ErrInvalidArgument, idStr))
			return
		}
		initialTrackIDs = append(initialTrackIDs, id)
	}

	collectionType := domain.CollectionType(req.Type) // Already validated by tag

	collection, err := h.audioUseCase.CreateCollection(r.Context(), req.Title, req.Description, collectionType, initialTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles Unauthenticated, internal errors
		return
	}

	// Fetch track details if needed for response (or adjust DTO mapping)
	// For now, map without full track details unless collection included them
	resp := dto.MapDomainCollectionToResponseDTO(collection, nil)

	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}


// GetCollectionDetails handles GET /api/v1/audio/collections/{collectionId}
// @Summary Get audio collection details
// @Description Retrieves details for a specific audio collection, including its metadata and ordered list of tracks.
// @ID get-collection-details
// @Tags Audio Collections
// @Produce json
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Success 200 {object} dto.AudioCollectionResponseDTO "Audio collection details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Collection ID Format"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error (e.g., failed to fetch tracks)"
// @Router /audio/collections/{collectionId} [get]
func (h *AudioHandler) GetCollectionDetails(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}

	// Get collection metadata
	collection, err := h.audioUseCase.GetCollectionDetails(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied
		return
	}

	// Get associated track details (ordered)
	tracks, err := h.audioUseCase.GetCollectionTracks(r.Context(), collectionID)
    if err != nil {
        // Log the error but might still return the collection metadata?
        slog.Default().ErrorContext(r.Context(), "Failed to fetch tracks for collection details", "error", err, "collectionID", collectionID)
        // Decide if this is critical or okay to return collection without tracks
        // Let's return an error for now, as the frontend likely expects tracks.
        httputil.RespondError(w, r, fmt.Errorf("failed to retrieve collection tracks: %w", err))
        return
    }

	// Map domain object and tracks to response DTO
	resp := dto.MapDomainCollectionToResponseDTO(collection, tracks)

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// UpdateCollectionMetadata handles PUT /api/v1/audio/collections/{collectionId}
// @Summary Update collection metadata
// @Description Updates the title and description of an audio collection owned by the authenticated user.
// @ID update-collection-metadata
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Param collection body dto.UpdateCollectionRequestDTO true "Updated collection metadata"
// @Success 204 "Collection metadata updated successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Collection ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId} [put]
func (h *AudioHandler) UpdateCollectionMetadata(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}

	var req dto.UpdateCollectionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	err = h.audioUseCase.UpdateCollectionMetadata(r.Context(), collectionID, req.Title, req.Description)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful update with no body
}

// UpdateCollectionTracks handles PUT /api/v1/audio/collections/{collectionId}/tracks
// @Summary Update collection tracks
// @Description Updates the ordered list of tracks within a specific collection owned by the authenticated user. Replaces the entire list.
// @ID update-collection-tracks
// @Tags Audio Collections
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Param tracks body dto.UpdateCollectionTracksRequestDTO true "Ordered list of track UUIDs"
// @Success 204 "Collection tracks updated successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input / Collection or Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found / Track Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId}/tracks [put]
func (h *AudioHandler) UpdateCollectionTracks(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}

	var req dto.UpdateCollectionTracksRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid request body", domain.ErrInvalidArgument))
		return
	}
	defer r.Body.Close()

	// Validate the request DTO itself (e.g., if array shouldn't be null, though empty is okay)
	// validator doesn't easily validate contents of array here, usecase does that.

	// Convert string IDs to domain IDs
	orderedTrackIDs := make([]domain.TrackID, 0, len(req.OrderedTrackIDs))
	for _, idStr := range req.OrderedTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format '%s' in ordered list", domain.ErrInvalidArgument, idStr))
			return
		}
		orderedTrackIDs = append(orderedTrackIDs, id)
	}

	err = h.audioUseCase.UpdateCollectionTracks(r.Context(), collectionID, orderedTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated, InvalidArgument (bad track id)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}


// DeleteCollection handles DELETE /api/v1/audio/collections/{collectionId}
// @Summary Delete an audio collection
// @Description Deletes an audio collection owned by the authenticated user.
// @ID delete-audio-collection
// @Tags Audio Collections
// @Produce json
// @Security BearerAuth
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Success 204 "Collection deleted successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (Not Owner)"
// @Failure 404 {object} httputil.ErrorResponseDTO "Collection Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/collections/{collectionId} [delete]
func (h *AudioHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	collectionIDStr := chi.URLParam(r, "collectionId")
	collectionID, err := domain.CollectionIDFromString(collectionIDStr)
	if err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid collection ID format", domain.ErrInvalidArgument))
		return
	}

	err = h.audioUseCase.DeleteCollection(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}