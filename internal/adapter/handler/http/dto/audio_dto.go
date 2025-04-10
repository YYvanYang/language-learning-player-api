// internal/adapter/handler/http/dto/audio_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port" // Import port for result struct
)

// --- Request DTOs ---

// ListTracksRequestDTO (Documents query parameters, not used for binding)
type ListTracksRequestDTO struct {
	Query         *string  `query:"q"`
	LanguageCode  *string  `query:"lang"`
	Level         *string  `query:"level"`
	IsPublic      *bool    `query:"isPublic"`
	Tags          []string `query:"tags"`
	SortBy        string   `query:"sortBy"`
	SortDirection string   `query:"sortDir"`
	Limit         int      `query:"limit"`
	Offset        int      `query:"offset"`
}

// CreateCollectionRequestDTO defines the JSON body for creating a collection.
type CreateCollectionRequestDTO struct {
	Title           string   `json:"title" validate:"required,max=255"`
	Description     string   `json:"description"`
	Type            string   `json:"type" validate:"required,oneof=COURSE PLAYLIST"`
	InitialTrackIDs []string `json:"initialTrackIds" validate:"omitempty,dive,uuid"`
}

// UpdateCollectionRequestDTO defines the JSON body for updating collection metadata.
type UpdateCollectionRequestDTO struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description"`
}

// UpdateCollectionTracksRequestDTO defines the JSON body for updating tracks in a collection.
type UpdateCollectionTracksRequestDTO struct {
	OrderedTrackIDs []string `json:"orderedTrackIds" validate:"omitempty,dive,uuid"`
}

// --- Response DTOs ---

// AudioTrackResponseDTO defines the JSON representation of a single audio track's basic info.
type AudioTrackResponseDTO struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	LanguageCode  string    `json:"languageCode"`
	Level         string    `json:"level,omitempty"`
	DurationMs    int64     `json:"durationMs"` // Point 1: Use milliseconds (int64)
	CoverImageURL *string   `json:"coverImageUrl,omitempty"`
	UploaderID    *string   `json:"uploaderId,omitempty"`
	IsPublic      bool      `json:"isPublic"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AudioTrackDetailsResponseDTO includes the track metadata, playback URL, and user-specific info.
type AudioTrackDetailsResponseDTO struct {
	AudioTrackResponseDTO                       // Embed basic track info
	PlayURL               string                `json:"playUrl"`                  // Presigned URL
	UserProgressMs        *int64                `json:"userProgressMs,omitempty"` // Point 1: User progress in ms
	UserBookmarks         []BookmarkResponseDTO `json:"userBookmarks,omitempty"`  // Array of user bookmarks for this track
}

// UploaderInfoDTO - embedded within AudioTrackDetailsResponseDTO if needed
type UploaderInfoDTO struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// MapDomainTrackToResponseDTO converts a domain track to its basic response DTO.
func MapDomainTrackToResponseDTO(track *domain.AudioTrack) AudioTrackResponseDTO {
	if track == nil {
		return AudioTrackResponseDTO{}
	} // Handle nil track gracefully

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
		DurationMs:    track.Duration.Milliseconds(), // Convert Duration to ms
		CoverImageURL: track.CoverImageURL,
		UploaderID:    uploaderIDStr,
		IsPublic:      track.IsPublic,
		Tags:          track.Tags,
		CreatedAt:     track.CreatedAt,
		UpdatedAt:     track.UpdatedAt,
	}
}

// MapDomainTrackToDetailsResponseDTO converts the result from the usecase to the detailed response DTO.
func MapDomainTrackToDetailsResponseDTO(result *port.GetAudioTrackDetailsResult) AudioTrackDetailsResponseDTO {
	if result == nil || result.Track == nil {
		return AudioTrackDetailsResponseDTO{} // Return empty DTO if track is nil
	}

	baseDTO := MapDomainTrackToResponseDTO(result.Track)
	detailsDTO := AudioTrackDetailsResponseDTO{
		AudioTrackResponseDTO: baseDTO,
		PlayURL:               result.PlayURL,
		UserProgressMs:        nil,
		UserBookmarks:         make([]BookmarkResponseDTO, 0), // Initialize with correct type
	}

	if result.UserProgress != nil {
		progressMs := result.UserProgress.Progress.Milliseconds()
		detailsDTO.UserProgressMs = &progressMs
	}

	if len(result.UserBookmarks) > 0 {
		// UPDATED: Ensure we create BookmarkResponseDTO instances using the function from activity_dto.go scope
		detailsDTO.UserBookmarks = make([]BookmarkResponseDTO, len(result.UserBookmarks))
		for i, b := range result.UserBookmarks {
			// Call the mapping function (which is accessible within the dto package)
			detailsDTO.UserBookmarks[i] = MapDomainBookmarkToResponseDTO(b)
		}
	}

	return detailsDTO
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
	Tracks      []AudioTrackResponseDTO `json:"tracks,omitempty"`
}

// MapDomainCollectionToResponseDTO converts a domain collection to its response DTO.
func MapDomainCollectionToResponseDTO(collection *domain.AudioCollection, tracks []*domain.AudioTrack) AudioCollectionResponseDTO {
	if collection == nil {
		return AudioCollectionResponseDTO{}
	} // Handle nil gracefully

	dto := AudioCollectionResponseDTO{
		ID:          collection.ID.String(),
		Title:       collection.Title,
		Description: collection.Description,
		OwnerID:     collection.OwnerID.String(),
		Type:        string(collection.Type),
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
		Tracks:      make([]AudioTrackResponseDTO, 0), // Initialize empty
	}
	if tracks != nil {
		dto.Tracks = make([]AudioTrackResponseDTO, len(tracks))
		for i, t := range tracks {
			dto.Tracks[i] = MapDomainTrackToResponseDTO(t) // Use the basic track mapper
		}
	}
	return dto
}

// PaginatedTracksResponseDTO - Using common PaginatedResponseDTO instead
// type PaginatedTracksResponseDTO struct { ... }

// PaginatedCollectionsResponseDTO - Using common PaginatedResponseDTO instead
// type PaginatedCollectionsResponseDTO struct { ... }
