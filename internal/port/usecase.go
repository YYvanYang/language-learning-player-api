// internal/port/usecase.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// AuthUseCase defines the methods for the Auth use case layer.
type AuthUseCase interface {
	RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error)
	LoginWithPassword(ctx context.Context, emailStr, password string) (string, error)
	AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error)
}

// ListTracksInput defines parameters for listing/searching tracks at the use case layer.
// It embeds pagination.Page.
// RENAMED from UseCaseListTracksParams
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

// GetAudioTrackDetailsResult holds the combined result for getting track details.
// ADDED: Struct to cleanly return multiple values from usecase.
type GetAudioTrackDetailsResult struct {
	Track         *domain.AudioTrack
	PlayURL       string
	UserProgress  *domain.PlaybackProgress // Nil if user not logged in or no progress
	UserBookmarks []*domain.Bookmark       // Empty slice if user not logged in or no bookmarks
}

// AudioContentUseCase defines the methods for the Audio Content use case layer.
type AudioContentUseCase interface {
	// CHANGED: GetAudioTrackDetails now returns a result struct
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*GetAudioTrackDetailsResult, error)
	// CHANGED: ListTracks now takes ListTracksInput and returns actual Page used
	ListTracks(ctx context.Context, input ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error)
	CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error)
	GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error)
	GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error)
	UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error
	UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error
}

// UserActivityUseCase defines the methods for the User Activity use case layer.
type UserActivityUseCase interface {
	RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error
	GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	ListUserProgress(ctx context.Context, params ListProgressParams) ([]*domain.PlaybackProgress, int, pagination.Page, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	ListBookmarks(ctx context.Context, params ListBookmarksParams) ([]*domain.Bookmark, int, pagination.Page, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
}

// UserUseCase defines the interface for user-related operations (e.g., profile)
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error)
}
