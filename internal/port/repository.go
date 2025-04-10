// internal/port/repository.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// --- Repository Interfaces ---

// RefreshTokenData holds the details of a stored refresh token.
type RefreshTokenData struct {
	TokenHash string // SHA-256 hash
	UserID    domain.UserID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Save(ctx context.Context, tokenData *RefreshTokenData) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshTokenData, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) // Returns number of tokens deleted
}

// UserRepository defines the persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	// ADDED: EmailExists method
	EmailExists(ctx context.Context, email domain.Email) (bool, error)
}

// ListTracksFilters defines parameters for filtering/searching tracks at the repository layer.
// RENAMED from ListTracksParams
type ListTracksFilters struct {
	Query         *string            // Search query (title, description, maybe tags)
	LanguageCode  *string            // Filter by language code
	Level         *domain.AudioLevel // Filter by level
	IsPublic      *bool              // Filter by public status
	UploaderID    *domain.UserID     // Filter by uploader
	Tags          []string           // Filter by tags (match any)
	SortBy        string             // e.g., "createdAt", "title", "durationMs" (DB column name might differ)
	SortDirection string             // "asc" or "desc"
}

// AudioTrackRepository defines the persistence operations for AudioTrack entities.
type AudioTrackRepository interface {
	FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)
	ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error)
	// List retrieves a paginated list of tracks based on filter and sort parameters.
	// RENAMED params type to ListTracksFilters
	List(ctx context.Context, filters ListTracksFilters, page pagination.Page) (tracks []*domain.AudioTrack, total int, err error)
	Create(ctx context.Context, track *domain.AudioTrack) error
	Update(ctx context.Context, track *domain.AudioTrack) error
	Delete(ctx context.Context, id domain.TrackID) error
	Exists(ctx context.Context, id domain.TrackID) (bool, error)
}

// AudioCollectionRepository defines the persistence operations for AudioCollection entities.
type AudioCollectionRepository interface {
	FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) (collections []*domain.AudioCollection, total int, err error)
	Create(ctx context.Context, collection *domain.AudioCollection) error
	UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error
	ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	Delete(ctx context.Context, id domain.CollectionID) error
}

// PlaybackProgressRepository defines the persistence operations for PlaybackProgress entities.
type PlaybackProgressRepository interface {
	Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	Upsert(ctx context.Context, progress *domain.PlaybackProgress) error
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (progressList []*domain.PlaybackProgress, total int, err error)
}

// BookmarkRepository defines the persistence operations for Bookmark entities.
type BookmarkRepository interface {
	FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)
	ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)
	ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) (bookmarks []*domain.Bookmark, total int, err error)
	Create(ctx context.Context, bookmark *domain.Bookmark) error
	Delete(ctx context.Context, id domain.BookmarkID) error
}

// --- Transaction Management ---

type Tx interface{}

type TransactionManager interface {
	Begin(ctx context.Context) (TxContext context.Context, err error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
}
