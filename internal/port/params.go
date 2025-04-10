// internal/port/params.go
package port

import (
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// === Use Case Layer Input/Result Structs ===

// ListTracksInput defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
type ListTracksInput struct {
	Query         *string            // Search query (title, description, maybe tags)
	LanguageCode  *string            // Filter by language code
	Level         *domain.AudioLevel // Filter by level
	IsPublic      *bool              // Filter by public status
	UploaderID    *domain.UserID     // Filter by uploader
	Tags          []string           // Filter by tags (match any)
	SortBy        string             // e.g., "createdAt", "title", "durationMs"
	SortDirection string             // "asc" or "desc"
	Page          pagination.Page    // Embed pagination parameters
}

// ListProgressInput defines parameters for listing user progress at the use case layer.
type ListProgressInput struct {
	UserID domain.UserID
	Page   pagination.Page
}

// ListBookmarksInput defines parameters for listing user bookmarks at the use case layer.
type ListBookmarksInput struct {
	UserID        domain.UserID
	TrackIDFilter *domain.TrackID // Optional filter by track
	Page          pagination.Page
}

// RequestUploadResult holds the result of requesting an upload URL.
type RequestUploadResult struct {
	UploadURL string
	ObjectKey string
}

// CompleteUploadInput holds the data needed to finalize an upload and create a track record.
type CompleteUploadInput struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	Duration      time.Duration // Use time.Duration consistent with domain model
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}

// --- Batch Request Input ---
type BatchRequestUploadInputItem struct {
	Filename    string
	ContentType string
}
type BatchRequestUploadInput struct {
	Files []BatchRequestUploadInputItem
}

// --- Batch URL Result ---
type BatchURLResultItem struct {
	OriginalFilename string
	ObjectKey        string
	UploadURL        string
	// Using string for Error is simpler for JSON marshalling in batch results,
	// though less type-safe internally. Acknowledge this trade-off.
	Error string
}

// --- Batch Complete Input ---
type BatchCompleteItem struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	Duration      time.Duration // Use time.Duration consistent with domain model
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}
type BatchCompleteInput struct {
	Tracks []BatchCompleteItem
}

// --- Batch Complete Result ---
type BatchCompleteResultItem struct {
	ObjectKey string
	Success   bool
	TrackID   string // Use string for UUID here as it's just data
	// Using string for Error is simpler for JSON marshalling in batch results,
	// though less type-safe internally. Acknowledge this trade-off.
	Error string
}

// === Audio Track Details Result (Used by AudioContentUseCase) ===

// GetAudioTrackDetailsResult holds the combined result for getting track details.
type GetAudioTrackDetailsResult struct {
	Track         *domain.AudioTrack
	PlayURL       string
	UserProgress  *domain.PlaybackProgress // Nil if user not logged in or no progress
	UserBookmarks []*domain.Bookmark       // Empty slice if user not logged in or no bookmarks
}
