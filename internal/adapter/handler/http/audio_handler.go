// ============================================
// FILE: internal/adapter/handler/http/audio_handler.go (MODIFIED)
// ============================================
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings" // Import strings

	"github.com/go-chi/chi/v5"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"        // Import for GetUserIDFromContext
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware" // Correct middleware import path
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// AudioHandler handles HTTP requests related to audio tracks and collections.
type AudioHandler struct {
	audioUseCase port.AudioContentUseCase
	validator    *validation.Validator
	// No longer needs activityUseCase here, logic moved into audioUseCase
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
// @Description Retrieves details for a specific audio track, including metadata, playback URL, and user-specific progress/bookmarks if authenticated.
// @ID get-track-details
// @Tags Audio Tracks
// @Produce json
// @Param trackId path string true "Audio Track UUID" Format(uuid)
// @Security BearerAuth // Optional: Indicate that auth affects the response (user data)
// @Success 200 {object} dto.AudioTrackDetailsResponseDTO "Audio track details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Track ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized (if accessing private track without auth)"
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

	// Usecase now returns a result struct containing track, URL, and optional user data
	result, err := h.audioUseCase.GetAudioTrackDetails(r.Context(), trackID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied, Unauthenticated, internal errors
		return
	}

	// Point 4: Mapping moved here, using a mapper function in dto package
	resp := dto.MapDomainTrackToDetailsResponseDTO(result) // Pass the result struct to the mapper

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Point 6: Refactored parameter parsing into helper function parseListTracksInput
func parseListTracksInput(r *http.Request) (port.ListTracksInput, error) {
	q := r.URL.Query()
	input := port.ListTracksInput{} // Use usecase input struct

	// Parse simple string filters
	if query := q.Get("q"); query != "" {
		input.Query = &query
	}
	if lang := q.Get("lang"); lang != "" {
		input.LanguageCode = &lang
	}
	input.Tags = q["tags"] // Already handles empty slice

	// Parse Level (and validate)
	if levelStr := q.Get("level"); levelStr != "" {
		level := domain.AudioLevel(strings.ToUpper(levelStr)) // Normalize case
		if level.IsValid() {
			input.Level = &level
		} else {
			return input, fmt.Errorf("%w: invalid level query parameter '%s'", domain.ErrInvalidArgument, levelStr)
		}
	}

	// Parse isPublic (boolean)
	if isPublicStr := q.Get("isPublic"); isPublicStr != "" {
		isPublic, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			input.IsPublic = &isPublic
		} else {
			return input, fmt.Errorf("%w: invalid isPublic query parameter (must be true or false)", domain.ErrInvalidArgument)
		}
	}

	// Parse Sort parameters
	input.SortBy = q.Get("sortBy")
	input.SortDirection = q.Get("sortDir")

	// Parse Pagination parameters
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")
	limit, errLimit := strconv.Atoi(limitStr)
	offset, errOffset := strconv.Atoi(offsetStr)
	if (limitStr != "" && errLimit != nil) || (offsetStr != "" && errOffset != nil) {
		return input, fmt.Errorf("%w: invalid limit or offset query parameter", domain.ErrInvalidArgument)
	}
	// Let the usecase/pagination package handle defaults/constraints for limit/offset
	input.Page = pagination.Page{Limit: limit, Offset: offset}

	return input, nil
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
// @Success 200 {object} dto.PaginatedResponseDTO{data=[]dto.AudioTrackResponseDTO} "Paginated list of audio tracks"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Query Parameter Format"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /audio/tracks [get]
func (h *AudioHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	// Point 6: Use helper function for parsing
	ucInput, err := parseListTracksInput(r)
	if err != nil {
		httputil.RespondError(w, r, err) // Return 400 Bad Request if parsing fails
		return
	}

	// Point 5: Call use case with the ListTracksInput struct
	tracks, total, actualPageInfo, err := h.audioUseCase.ListTracks(r.Context(), ucInput)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.AudioTrackResponseDTO, len(tracks))
	for i, track := range tracks {
		respData[i] = dto.MapDomainTrackToResponseDTO(track) // Point 1: DTO mapping uses ms
	}

	// Use the pagination info returned by the usecase (which applied constraints)
	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)
	// Map to the common DTO structure
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

// ListMyCollections handles GET /api/v1/users/me/collections
// @Summary List my audio collections
// @Description Retrieves a paginated list of audio collections owned by the currently authenticated user.
// @ID list-my-collections
// @Tags Audio Collections
// @Produce json
// @Security BearerAuth
// @Param sortBy query string false "Sort field (e.g., createdAt, title, updatedAt)" default(updatedAt)
// @Param sortDir query string false "Sort direction (asc or desc)" default(desc) Enums(asc, desc)
// @Param limit query int false "Pagination limit" default(20) minimum(1) maximum(100)
// @Param offset query int false "Pagination offset" default(0) minimum(0)
// @Success 200 {object} dto.PaginatedResponseDTO{data=[]dto.AudioCollectionResponseDTO} "Paginated list of user's audio collections"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Query Parameter Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me/collections [get]
// ADDED: New handler method
func (h *AudioHandler) ListMyCollections(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	q := r.URL.Query()
	limitStr := q.Get("limit")
	offsetStr := q.Get("offset")
	limit, errLimit := strconv.Atoi(limitStr)
	offset, errOffset := strconv.Atoi(offsetStr)
	if (limitStr != "" && errLimit != nil) || (offsetStr != "" && errOffset != nil) {
		httputil.RespondError(w, r, fmt.Errorf("%w: invalid limit or offset query parameter", domain.ErrInvalidArgument))
		return
	}

	// Create params struct for use case
	ucParams := port.ListUserCollectionsParams{
		UserID:        userID,
		SortBy:        q.Get("sortBy"),
		SortDirection: q.Get("sortDir"),
		Page:          pagination.Page{Limit: limit, Offset: offset}, // Let use case validate/default pagination
	}

	// Call use case
	collections, total, actualPageInfo, err := h.audioUseCase.ListUserCollections(r.Context(), ucParams)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// Map domain collections to response DTOs (without detailed tracks for list view)
	respData := make([]dto.AudioCollectionResponseDTO, len(collections))
	for i, col := range collections {
		// Pass nil for tracks slice to indicate it's a list view
		respData[i] = dto.MapDomainCollectionToResponseDTO(col, nil)
	}

	// Construct paginated response using actual page info from use case
	paginatedResult := pagination.NewPaginatedResponse(respData, total, actualPageInfo)
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

	initialTrackIDs := make([]domain.TrackID, 0, len(req.InitialTrackIDs))
	for _, idStr := range req.InitialTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid initial track ID format '%s'", domain.ErrInvalidArgument, idStr))
			return
		}
		initialTrackIDs = append(initialTrackIDs, id)
	}

	collectionType := domain.CollectionType(req.Type)

	collection, err := h.audioUseCase.CreateCollection(r.Context(), req.Title, req.Description, collectionType, initialTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	// Map the created collection (potentially with initial track IDs, but not full track details)
	resp := dto.MapDomainCollectionToResponseDTO(collection, nil) // Pass nil for tracks here
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// GetCollectionDetails handles GET /api/v1/audio/collections/{collectionId}
// @Summary Get audio collection details
// @Description Retrieves details for a specific audio collection, including its metadata and ordered list of tracks. Backend verifies if user can view it.
// @ID get-collection-details
// @Tags Audio Collections
// @Produce json
// @Param collectionId path string true "Audio Collection UUID" Format(uuid)
// @Security BearerAuth // Indicate auth might affect response or access
// @Success 200 {object} dto.AudioCollectionResponseDTO "Audio collection details found"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Collection ID Format"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized (if collection is private and not owned)"
// @Failure 403 {object} httputil.ErrorResponseDTO "Forbidden (if user cannot access)"
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

	// Use case should perform permission checks internally based on context userId
	collection, err := h.audioUseCase.GetCollectionDetails(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err) // Handles NotFound, PermissionDenied etc.
		return
	}

	// Fetch tracks associated with the collection
	tracks, err := h.audioUseCase.GetCollectionTracks(r.Context(), collectionID)
	if err != nil {
		// Log error but might still return collection metadata
		slog.Default().ErrorContext(r.Context(), "Failed to fetch tracks for collection details", "error", err, "collectionID", collectionID)
		// Depending on requirements, could return error here or just the collection without tracks
		// Let's return the collection metadata anyway
	}

	// Map collection and its tracks to the response DTO
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
	// Use case handles ownership check
	err = h.audioUseCase.UpdateCollectionMetadata(r.Context(), collectionID, req.Title, req.Description)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}
	orderedTrackIDs := make([]domain.TrackID, 0, len(req.OrderedTrackIDs))
	for _, idStr := range req.OrderedTrackIDs {
		id, err := domain.TrackIDFromString(idStr)
		if err != nil {
			httputil.RespondError(w, r, fmt.Errorf("%w: invalid track ID format '%s' in ordered list", domain.ErrInvalidArgument, idStr))
			return
		}
		orderedTrackIDs = append(orderedTrackIDs, id)
	}
	// Use case handles ownership check
	err = h.audioUseCase.UpdateCollectionTracks(r.Context(), collectionID, orderedTrackIDs)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	// Use case handles ownership check
	err = h.audioUseCase.DeleteCollection(r.Context(), collectionID)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
