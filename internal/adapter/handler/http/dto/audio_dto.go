// internal/adapter/handler/http/dto/audio_dto.go
package dto

import (
	"time"
	
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
)

// --- Request DTOs ---

// ListTracksRequestDTO holds query parameters for listing tracks.
// We'll typically bind these from r.URL.Query() in the handler, not from request body.
// Validation tags aren't standard for query params via validator, often handled manually.
type ListTracksRequestDTO struct {
	Query         *string   `query:"q"`           // Search query
	LanguageCode  *string   `query:"lang"`      // Filter by language
	Level         *string   `query:"level"`     // Filter by level
	IsPublic      *bool     `query:"isPublic"`  // Filter by public status
	Tags          []string  `query:"tags"`      // Filter by tags (e.g., ?tags=news&tags=podcast)
	SortBy        string    `query:"sortBy"`    // e.g., "createdAt", "title"
	SortDirection string    `query:"sortDir"`   // "asc" or "desc"
	Limit         int       `query:"limit"`     // Pagination limit
	Offset        int       `query:"offset"`    // Pagination offset
}

// CreateCollectionRequestDTO defines the JSON body for creating a collection.
type CreateCollectionRequestDTO struct {
	Title           string   `json:"title" validate:"required,max=255"`
	Description     string   `json:"description"`
	Type            string   `json:"type" validate:"required,oneof=COURSE PLAYLIST"` // Matches domain.CollectionType
	InitialTrackIDs []string `json:"initialTrackIds"` // Optional list of track UUIDs
}

// UpdateCollectionRequestDTO defines the JSON body for updating collection metadata.
type UpdateCollectionRequestDTO struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCollectionTracksRequestDTO defines the JSON body for updating tracks in a collection.
type UpdateCollectionTracksRequestDTO struct {
	OrderedTrackIDs []string `json:"orderedTrackIds"` // Full ordered list of track UUIDs
}


// --- Response DTOs ---

// AudioTrackResponseDTO defines the JSON representation of a single audio track.
type AudioTrackResponseDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	LanguageCode  string    `json:"languageCode"`
	Level         string    `json:"level,omitempty"`
	DurationMs    int64     `json:"durationMs"` // Use int64 for milliseconds
	CoverImageURL *string   `json:"coverImageUrl,omitempty"`
	UploaderID    *string   `json:"uploaderId,omitempty"` // Use string UUID
	IsPublic      bool      `json:"isPublic"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"createdAt"` // Use time.Time, will marshal to RFC3339
	UpdatedAt     time.Time `json:"updatedAt"`
}


// AudioTrackDetailsResponseDTO includes the track metadata and playback URL.
type AudioTrackDetailsResponseDTO struct {
	AudioTrackResponseDTO        // Embed basic track info
	PlayURL               string `json:"playUrl"` // Presigned URL
	// TODO: Add fields for user progress, bookmarks, transcription status if needed later
	// UserProgressSeconds *int      `json:"userProgressSeconds,omitempty"`
	// Bookmarks           []BookmarkResponseDTO `json:"bookmarks,omitempty"`
	// TranscriptionAvailable bool `json:"transcriptionAvailable"`
}

// MapDomainTrackToResponseDTO converts a domain track to its response DTO.
func MapDomainTrackToResponseDTO(track *domain.AudioTrack) AudioTrackResponseDTO {
	var uploaderIDStr *string
	if track.UploaderID != nil {
		s := track.UploaderID.String()
		uploaderIDStr = &s
	}
	return AudioTrackResponseDTO{
		ID:            track.ID.String(),
		Title:         track.Title,
		Description:   track.Description,
		LanguageCode:  track.Language.Code(),
		Level:         string(track.Level),
		DurationMs:    track.Duration.Milliseconds(),
		CoverImageURL: track.CoverImageURL,
		UploaderID:    uploaderIDStr,
		IsPublic:      track.IsPublic,
		Tags:          track.Tags,
		CreatedAt:     track.CreatedAt,
		UpdatedAt:     track.UpdatedAt,
	}
}

// AudioCollectionResponseDTO defines the JSON representation of a collection.
type AudioCollectionResponseDTO struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"ownerId"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	// Optionally include track details or just IDs:
	// TrackIDs []string `json:"trackIds"` // Just the ordered IDs
	Tracks []AudioTrackResponseDTO `json:"tracks,omitempty"` // Include full track details if needed by frontend
}

// MapDomainCollectionToResponseDTO converts a domain collection to its response DTO.
// Optionally includes track details if they are loaded.
func MapDomainCollectionToResponseDTO(collection *domain.AudioCollection, tracks []*domain.AudioTrack) AudioCollectionResponseDTO {
	dto := AudioCollectionResponseDTO{
		ID:          collection.ID.String(),
		Title:       collection.Title,
		Description: collection.Description,
		OwnerID:     collection.OwnerID.String(),
		Type:        string(collection.Type),
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
		// TrackIDs: make([]string, len(collection.TrackIDs)), // Map TrackIDs if only sending IDs
		Tracks: make([]AudioTrackResponseDTO, 0), // Initialize empty slice for tracks
	}
	// for i, id := range collection.TrackIDs {
	// 	dto.TrackIDs[i] = id.String()
	// }

	// If full track details were provided (e.g., after fetching them based on TrackIDs)
	if tracks != nil {
		dto.Tracks = make([]AudioTrackResponseDTO, len(tracks))
		for i, t := range tracks {
			dto.Tracks[i] = MapDomainTrackToResponseDTO(t)
		}
	}

	return dto
}