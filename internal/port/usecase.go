// internal/port/usecase.go
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
)

// AuthUseCase defines the methods for the Auth use case layer.
type AuthUseCase interface {
	RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error)
	LoginWithPassword(ctx context.Context, emailStr, password string) (string, error)
	AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (authToken string, isNewUser bool, err error)
}

// AudioContentUseCase defines the methods for the Audio Content use case layer.
type AudioContentUseCase interface {
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error)
	ListTracks(ctx context.Context, params ListTracksParams, page Page) ([]*domain.AudioTrack, int, error)
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
	ListUserProgress(ctx context.Context, userID domain.UserID, page Page) ([]*domain.PlaybackProgress, int, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	ListBookmarks(ctx context.Context, userID domain.UserID, trackID *domain.TrackID, page Page) ([]*domain.Bookmark, int, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
} 