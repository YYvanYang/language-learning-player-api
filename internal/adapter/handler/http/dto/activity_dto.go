// internal/adapter/handler/http/dto/activity_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
)

// --- Request DTOs ---

// RecordProgressRequestDTO defines the JSON body for recording playback progress.
type RecordProgressRequestDTO struct {
	TrackID    string `json:"trackId" validate:"required,uuid"`
	ProgressMs int64  `json:"progressMs" validate:"required,gte=0"` // Point 1: Already uses ms
}

// CreateBookmarkRequestDTO defines the JSON body for creating a bookmark.
type CreateBookmarkRequestDTO struct {
	TrackID     string `json:"trackId" validate:"required,uuid"`
	TimestampMs int64  `json:"timestampMs" validate:"required,gte=0"` // Point 1: Already uses ms
	Note        string `json:"note"`
}

// --- Response DTOs ---

// PlaybackProgressResponseDTO defines the JSON representation of playback progress.
type PlaybackProgressResponseDTO struct {
	UserID         string    `json:"userId"`
	TrackID        string    `json:"trackId"`
	ProgressMs     int64     `json:"progressMs"` // Point 1: Already uses ms
	LastListenedAt time.Time `json:"lastListenedAt"`
}

// Point 1: MapDomainProgressToResponseDTO converts domain progress (with time.Duration) to DTO (with int64 ms).
func MapDomainProgressToResponseDTO(p *domain.PlaybackProgress) PlaybackProgressResponseDTO {
	if p == nil {
		return PlaybackProgressResponseDTO{}
	} // Handle nil gracefully
	return PlaybackProgressResponseDTO{
		UserID:         p.UserID.String(),
		TrackID:        p.TrackID.String(),
		ProgressMs:     p.Progress.Milliseconds(), // Convert duration to ms
		LastListenedAt: p.LastListenedAt,
	}
}

// BookmarkResponseDTO defines the JSON representation of a bookmark.
// NOTE: Also defined in audio_dto.go for embedding. Ensure consistency.
type BookmarkResponseDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	TrackID     string    `json:"trackId"`
	TimestampMs int64     `json:"timestampMs"` // Point 1: Already uses ms
	Note        string    `json:"note,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Point 1: MapDomainBookmarkToResponseDTO converts domain bookmark (with time.Duration) to DTO (with int64 ms).
func MapDomainBookmarkToResponseDTO(b *domain.Bookmark) BookmarkResponseDTO {
	if b == nil {
		return BookmarkResponseDTO{}
	} // Handle nil gracefully
	return BookmarkResponseDTO{
		ID:          b.ID.String(),
		UserID:      b.UserID.String(),
		TrackID:     b.TrackID.String(),
		TimestampMs: b.Timestamp.Milliseconds(), // Convert duration to ms
		Note:        b.Note,
		CreatedAt:   b.CreatedAt,
	}
}

// --- Paginated Response DTOs ---
// REMOVED PaginatedProgressResponseDTO and PaginatedBookmarksResponseDTO
// Use common_dto.PaginatedResponseDTO instead.
