// internal/adapter/handler/http/audio_handler.go
package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
	"your_project/internal/adapter/handler/http/dto" // Adjust import path
	"your_project/pkg/httputil"    // Adjust import path
	"your_project/pkg/validation"  // Adjust import path
)

// AudioHandler handles HTTP requests related to audio tracks and collections.
type AudioHandler struct {
	audioUseCase AudioContentUseCase // Use interface defined below
	validator    *validation.Validator
}

// AudioContentUseCase defines the methods expected from the use case layer.
// Input Port for this handler.
type AudioContentUseCase interface {
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error)
	ListTracks(ctx context.Context, params port.ListTracksParams, page port.Page) ([]*domain.AudioTrack, int, error)
	CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error)
	GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error)
    GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) // New method needed
	UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error
	UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error
}


// NewAudioHandler creates a new AudioHandler.
func NewAudioHandler(uc AudioContentUseCase, v *validation.Validator) *AudioHandler {
	return &AudioHandler{
		audioUseCase: uc,
		validator:    v,
	}
}


// --- Track Handlers ---

// GetTrackDetails handles GET /api/v1/audio/tracks/{trackId}
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
func (h *AudioHandler) ListTracks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters manually or using a library like schema
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }

	page := port.Page{Limit: limit, Offset: offset}

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

	tracks, total, err := h.audioUseCase.ListTracks(r.Context(), params, page)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	respData := make([]dto.AudioTrackResponseDTO, len(tracks))
	for i, track := range tracks {
		respData[i] = dto.MapDomainTrackToResponseDTO(track)
	}

	resp := dto.PaginatedTracksResponseDTO{
		Data:   respData,
		Total:  total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}


// --- Collection Handlers ---

// CreateCollection handles POST /api/v1/audio/collections
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