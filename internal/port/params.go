// internal/port/params.go
package port

import (
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
)

// === Use Case Layer Parameter Structs ===

// UseCaseListTracksParams defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
// RENAMED from ListTracksParams to avoid conflict with repository params.
type UseCaseListTracksParams struct {
	Query         *string           // Search query (title, description, maybe tags)
	LanguageCode  *string           // Filter by language code
	Level         *domain.AudioLevel// Filter by level
	IsPublic      *bool             // Filter by public status
	UploaderID    *domain.UserID    // Filter by uploader
	Tags          []string          // Filter by tags (match any)
	SortBy        string            // e.g., "createdAt", "title", "duration"
	SortDirection string            // "asc" or "desc"
	Page          pagination.Page   // Embed pagination parameters
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