// internal/port/params.go
package port

import (
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// === Use Case Layer Parameter Structs ===

// UseCaseListTracksParams defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
// RENAMED from ListTracksParams to avoid conflict with repository params.
type UseCaseListTracksParams struct {
	Query         *string            // Search query (title, description, maybe tags)
	LanguageCode  *string            // Filter by language code
	Level         *domain.AudioLevel // Filter by level
	IsPublic      *bool              // Filter by public status
	UploaderID    *domain.UserID     // Filter by uploader
	Tags          []string           // Filter by tags (match any)
	SortBy        string             // e.g., "createdAt", "title", "duration"
	SortDirection string             // "asc" or "desc"
	Page          pagination.Page    // Embed pagination parameters
}

// ListProgressParams defines parameters for listing user progress at the use case layer.
type ListProgressParams struct {
	UserID domain.UserID
	Page   pagination.Page
}

// ListBookmarksParams defines parameters for listing user bookmarks at the use case layer.
type ListBookmarksParams struct {
	UserID        domain.UserID
	TrackIDFilter *domain.TrackID // Optional filter by track
	Page          pagination.Page
}

// RequestUploadResult holds the result of requesting an upload URL.
type RequestUploadResult struct {
	UploadURL string
	ObjectKey string
}

// CompleteUploadRequest holds the data needed to finalize an upload and create a track record.
type CompleteUploadRequest struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	DurationMs    int64
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}

// --- Batch Request ---
type BatchRequestUploadItem struct {
	Filename    string
	ContentType string
}
type BatchRequestUpload struct {
	Files []BatchRequestUploadItem
}

// --- Batch URL Response ---
type BatchURLResultItem struct {
	OriginalFilename string
	ObjectKey        string
	UploadURL        string
	Error            string // Keep error string for item-level reporting
}

// --- Batch Complete Request ---
type BatchCompleteItem struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string
	DurationMs    int64
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}
type BatchCompleteRequest struct {
	Tracks []BatchCompleteItem
}

// --- Batch Complete Response ---
type BatchCompleteResultItem struct {
	ObjectKey string
	Success   bool
	TrackID   string // Use string for UUID here as it's just data
	Error     string
}

// === Audio Track Details (Used by AudioContentUseCase) ===

// GetAudioTrackDetailsResult holds the combined result for getting track details.
// MOVED from usecase package to port package.
type GetAudioTrackDetailsResult struct {
	Track         *domain.AudioTrack
	PlayURL       string
	UserProgress  *domain.PlaybackProgress // Nil if user not logged in or no progress
	UserBookmarks []*domain.Bookmark       // Empty slice if user not logged in or no bookmarks
}
