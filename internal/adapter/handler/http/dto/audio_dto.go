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
// This struct isn't directly used for request binding but documents parameters.
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
	InitialTrackIDs []string `json:"initialTrackIds" validate:"omitempty,dive,uuid"` // Add validation for slice elements
}

// UpdateCollectionRequestDTO defines the JSON body for updating collection metadata.
type UpdateCollectionRequestDTO struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCollectionTracksRequestDTO defines the JSON body for updating tracks in a collection.
type UpdateCollectionTracksRequestDTO struct {
	OrderedTrackIDs []string `json:"orderedTrackIds" validate:"omitempty,dive,uuid"` // Add validation
}


// --- Response DTOs ---

// AudioTrackResponseDTO defines the JSON representation of a single audio track.
type AudioTrackResponseDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	LanguageCode  string    `json:"languageCode"`
	Level         string    `json:"level,omitempty"` // Domain type maps to string here
	DurationMs    int64     `json:"durationMs"`   // CORRECTED: Use milliseconds
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
		LanguageCode:  track.Language.Code(), // Use Code() method
		Level:         string(track.Level),   // Convert domain level to string
		DurationMs:    track.Duration.Milliseconds(), // CORRECTED: Get milliseconds
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
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description,omitempty"`
	OwnerID     string                  `json:"ownerId"`
	Type        string                  `json:"type"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	Tracks      []AudioTrackResponseDTO `json:"tracks,omitempty"` // Include full track details if needed by frontend
}

// MapDomainCollectionToResponseDTO converts a domain collection to its response DTO.
func MapDomainCollectionToResponseDTO(collection *domain.AudioCollection, tracks []*domain.AudioTrack) AudioCollectionResponseDTO {
	dto := AudioCollectionResponseDTO{
		ID:          collection.ID.String(),
		Title:       collection.Title,
		Description: collection.Description,
		OwnerID:     collection.OwnerID.String(),
		Type:        string(collection.Type),
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
		Tracks:      make([]AudioTrackResponseDTO, 0),
	}
	if tracks != nil {
		dto.Tracks = make([]AudioTrackResponseDTO, len(tracks))
		for i, t := range tracks {
			dto.Tracks[i] = MapDomainTrackToResponseDTO(t)
		}
	}
	return dto
}


// PaginatedTracksResponseDTO defines the paginated response for track list.
// Use the generic one from common_dto.go if preferred
type PaginatedTracksResponseDTO struct {
	Data       []AudioTrackResponseDTO `json:"data"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	Page       int                   `json:"page"`
	TotalPages int                   `json:"totalPages"`
}

// PaginatedCollectionsResponseDTO defines the paginated response for collection list.
// Use the generic one from common_dto.go if preferred
type PaginatedCollectionsResponseDTO struct {
    Data       []AudioCollectionResponseDTO `json:"data"`
    Total      int                        `json:"total"`
    Limit      int                        `json:"limit"`
    Offset     int                        `json:"offset"`
    Page       int                        `json:"page"`
    TotalPages int                        `json:"totalPages"`
}