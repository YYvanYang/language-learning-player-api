// ============================================
// FILE: internal/port/usecase.go (MODIFIED)
// ============================================
package port

import (
	"context"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// AuthResult holds both access and refresh tokens, and registration status.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	IsNewUser    bool         // Only relevant for external auth methods like Google sign-in
	User         *domain.User // ADDED: Include the authenticated user
}

// AuthUseCase defines the methods for the Auth use case layer.
type AuthUseCase interface {
	// RegisterWithPassword registers a new user with email/password.
	// Returns the created user, auth tokens, and error.
	RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, AuthResult, error)

	// LoginWithPassword authenticates a user with email/password.
	// Returns auth tokens and error.
	LoginWithPassword(ctx context.Context, emailStr, password string) (AuthResult, error)

	// AuthenticateWithGoogle handles login or registration via Google ID Token.
	// Returns auth tokens and error. The IsNewUser field in AuthResult indicates
	// if a new account was created during this process.
	AuthenticateWithGoogle(ctx context.Context, googleIdToken string) (AuthResult, error)

	// RefreshAccessToken validates a refresh token and issues a new pair of access/refresh tokens.
	RefreshAccessToken(ctx context.Context, refreshTokenValue string) (AuthResult, error)

	// Logout invalidates the provided refresh token.
	Logout(ctx context.Context, refreshTokenValue string) error
}

// AudioContentUseCase defines the methods for the Audio Content use case layer.
type AudioContentUseCase interface {
	GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*GetAudioTrackDetailsResult, error)
	ListTracks(ctx context.Context, input ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error)
	CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error)
	GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error)
	GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error)
	UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error
	UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error
	DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error
	// ADDED: List collections for the authenticated user
	ListUserCollections(ctx context.Context, params ListUserCollectionsParams) ([]*domain.AudioCollection, int, pagination.Page, error)
}

// UserActivityUseCase defines the methods for the User Activity use case layer.
type UserActivityUseCase interface {
	RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error
	GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error)
	ListUserProgress(ctx context.Context, params ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error)
	CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error)
	ListBookmarks(ctx context.Context, params ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error)
	DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error
}

// UserUseCase defines the interface for user-related operations (e.g., profile)
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error)
}

// UploadUseCase defines the methods for the Upload use case layer.
type UploadUseCase interface {
	RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*RequestUploadResult, error)
	CompleteUpload(ctx context.Context, userID domain.UserID, req CompleteUploadInput) (*domain.AudioTrack, error)
	RequestBatchUpload(ctx context.Context, userID domain.UserID, req BatchRequestUploadInput) ([]BatchURLResultItem, error)
	CompleteBatchUpload(ctx context.Context, userID domain.UserID, req BatchCompleteInput) ([]BatchCompleteResultItem, error)
}
