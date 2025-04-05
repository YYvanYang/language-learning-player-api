// internal/port/repository.go
package port

import (
	"context"
	"time"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	// Assuming a common pagination package exists or will be created in pkg/pagination
	// "github.com/yvanyang/language-learning-player-backend/pkg/pagination"
)

// --- Pagination Placeholder ---
// Replace with actual implementation in pkg/pagination later
type Page struct {
	Limit  int
	Offset int
}
// --- End Pagination Placeholder ---


// --- Repository Interfaces ---

// UserRepository defines the persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) // Use Email value object
	FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	// LinkProviderID(ctx context.Context, userID domain.UserID, provider domain.AuthProvider, providerUserID string) error // Maybe Update is sufficient? Depends on implementation. Let's rely on Update for now.
}

// ListTracksParams defines parameters for listing/searching tracks.
type ListTracksParams struct {
	Query         *string           // Search query (title, description, maybe tags)
	LanguageCode  *string           // Filter by language code
	Level         *domain.AudioLevel// Filter by level
	IsPublic      *bool             // Filter by public status
	UploaderID    *domain.UserID    // Filter by uploader
	Tags          []string          // Filter by tags (match any)
	SortBy        string            // e.g., "createdAt", "title", "duration"
	SortDirection string            // "asc" or "desc"
}

// AudioTrackRepository defines the persistence operations for AudioTrack entities.
type AudioTrackRepository interface {
	FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error)
	// ListByIDs retrieves multiple tracks efficiently, preserving order if possible.
	ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error)
	// List retrieves a paginated list of tracks based on filter and sort parameters.
	// Returns the list of tracks for the current page and the total count matching the filters.
	List(ctx context.Context, params ListTracksParams, page Page) (tracks []*domain.AudioTrack, total int, err error)
	Create(ctx context.Context, track *domain.AudioTrack) error
	Update(ctx context.Context, track *domain.AudioTrack) error
	Delete(ctx context.Context, id domain.TrackID) error
	Exists(ctx context.Context, id domain.TrackID) (bool, error) // Helper to check existence efficiently
}

// AudioCollectionRepository defines the persistence operations for AudioCollection entities.
type AudioCollectionRepository interface {
	FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	// FindWithTracks retrieves collection metadata and its ordered track list.
	// This might require a JOIN or separate queries in implementation.
	FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error)
	ListByOwner(ctx context.Context, ownerID domain.UserID, page Page) (collections []*domain.AudioCollection, total int, err error)
	Create(ctx context.Context, collection *domain.AudioCollection) error // Creates collection metadata only
	UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error // Updates title, description
	// ManageTracks persists the full ordered list of track IDs for a collection.
	// Implementations will likely clear existing associations and insert the new ones.
	ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	Delete(ctx context.Context, id domain.CollectionID) error // Assumes ownership check happens in Usecase
}

// PlaybackProgressRepository defines the persistence operations for PlaybackProgress entities.
type PlaybackProgressRepository interface {
	// Find retrieves progress for a specific user and track. Returns domain.ErrNotFound if none exists.
	Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	// Upsert creates or updates the progress record.
	Upsert(ctx context.Context, progress *domain.PlaybackProgress) error
	// ListByUser retrieves all progress records for a user, paginated, ordered by LastListenedAt descending.
	ListByUser(ctx context.Context, userID domain.UserID, page Page) (progressList []*domain.PlaybackProgress, total int, err error)
}

// BookmarkRepository defines the persistence operations for Bookmark entities.
type BookmarkRepository interface {
	FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error)
	// ListByUserAndTrack retrieves all bookmarks for a specific user on a specific track, ordered by timestamp.
	ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error)
	// ListByUser retrieves all bookmarks for a user, paginated, ordered by CreatedAt descending.
	ListByUser(ctx context.Context, userID domain.UserID, page Page) (bookmarks []*domain.Bookmark, total int, err error)
	Create(ctx context.Context, bookmark *domain.Bookmark) error
	Delete(ctx context.Context, id domain.BookmarkID) error // Assumes ownership check happens in Usecase
}


// --- Transaction Management (Optional but Recommended for Complex Usecases) ---

// Tx represents a database transaction. Specific implementation depends on the driver (e.g., pgx.Tx).
// Using interface{} allows flexibility but requires type assertions in the implementation.
// A more type-safe approach might involve generics (Go 1.18+) or specific interfaces per DB type.
type Tx interface{}

// TransactionManager defines an interface for managing database transactions.
type TransactionManager interface {
	// Begin starts a new transaction and returns a context containing the transaction handle.
	Begin(ctx context.Context) (TxContext context.Context, err error)
	// Commit commits the transaction stored in the context.
	Commit(ctx context.Context) error
	// Rollback aborts the transaction stored in the context.
	Rollback(ctx context.Context) error
	// Execute runs the given function within a transaction.
	// It automatically handles Begin, Commit, and Rollback based on the function's return error.
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
}

// --- End Transaction Management ---