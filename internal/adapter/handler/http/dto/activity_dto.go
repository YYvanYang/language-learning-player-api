// internal/adapter/handler/http/dto/activity_dto.go
package dto

import "time"

// --- Request DTOs ---

// RecordProgressRequestDTO defines the JSON body for recording playback progress.
type RecordProgressRequestDTO struct {
	TrackID        string  `json:"trackId" validate:"required,uuid"`
	ProgressSeconds float64 `json:"progressSeconds" validate:"required,gte=0"` // Use float64 for seconds from JSON
}

// CreateBookmarkRequestDTO defines the JSON body for creating a bookmark.
type CreateBookmarkRequestDTO struct {
	TrackID         string  `json:"trackId" validate:"required,uuid"`
	TimestampSeconds float64 `json:"timestampSeconds" validate:"required,gte=0"` // Use float64 for seconds from JSON
	Note            string  `json:"note"`                                    // Optional note
}

// --- Response DTOs ---

// PlaybackProgressResponseDTO defines the JSON representation of playback progress.
type PlaybackProgressResponseDTO struct {
	UserID         string    `json:"userId"`
	TrackID        string    `json:"trackId"`
	ProgressSeconds float64   `json:"progressSeconds"`
	LastListenedAt time.Time `json:"lastListenedAt"`
	// Optionally include basic Track info here?
	// TrackTitle    string `json:"trackTitle,omitempty"`
}

// MapDomainProgressToResponseDTO converts domain progress to DTO.
func MapDomainProgressToResponseDTO(p *domain.PlaybackProgress) PlaybackProgressResponseDTO {
	return PlaybackProgressResponseDTO{
		UserID:         p.UserID.String(),
		TrackID:        p.TrackID.String(),
		ProgressSeconds: p.Progress.Seconds(),
		LastListenedAt: p.LastListenedAt,
	}
}

// BookmarkResponseDTO defines the JSON representation of a bookmark.
type BookmarkResponseDTO struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	TrackID         string    `json:"trackId"`
	TimestampSeconds float64   `json:"timestampSeconds"`
	Note            string    `json:"note,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	// Optionally include basic Track info here?
	// TrackTitle    string `json:"trackTitle,omitempty"`
}

// MapDomainBookmarkToResponseDTO converts domain bookmark to DTO.
func MapDomainBookmarkToResponseDTO(b *domain.Bookmark) BookmarkResponseDTO {
	return BookmarkResponseDTO{
		ID:               b.ID.String(),
		UserID:           b.UserID.String(),
		TrackID:          b.TrackID.String(),
		TimestampSeconds: b.Timestamp.Seconds(),
		Note:             b.Note,
		CreatedAt:        b.CreatedAt,
	}
}


// PaginatedProgressResponseDTO defines the structure for paginated progress lists.
type PaginatedProgressResponseDTO struct {
	Data  []PlaybackProgressResponseDTO `json:"data"`
	Total int                           `json:"total"`
	Limit int                           `json:"limit"`
	Offset int                          `json:"offset"`
}

// PaginatedBookmarksResponseDTO defines the structure for paginated bookmark lists.
type PaginatedBookmarksResponseDTO struct {
	Data  []BookmarkResponseDTO `json:"data"`
	Total int                   `json:"total"`
	Limit int                   `json:"limit"`
	Offset int                  `json:"offset"`
}