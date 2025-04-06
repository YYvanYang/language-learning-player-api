// internal/usecase/user_activity_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

// UserActivityUseCase handles business logic for user interactions like progress and bookmarks.
type UserActivityUseCase struct {
	progressRepo port.PlaybackProgressRepository
	bookmarkRepo port.BookmarkRepository
	trackRepo    port.AudioTrackRepository // Needed to validate track existence
	logger       *slog.Logger
}

// NewUserActivityUseCase creates a new UserActivityUseCase.
func NewUserActivityUseCase(
	pr port.PlaybackProgressRepository,
	br port.BookmarkRepository,
	tr port.AudioTrackRepository,
	log *slog.Logger,
) *UserActivityUseCase {
	return &UserActivityUseCase{
		progressRepo: pr,
		bookmarkRepo: br,
		trackRepo:    tr,
		logger:       log.With("usecase", "UserActivityUseCase"),
	}
}

// --- Playback Progress Use Cases ---

// RecordPlaybackProgress saves or updates the user's listening progress for a track.
func (uc *UserActivityUseCase) RecordPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID, progress time.Duration) error {
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for progress update", "error", err, "trackID", trackID, "userID", userID)
		return fmt.Errorf("failed to validate track: %w", err)
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to record progress for non-existent track", "trackID", trackID, "userID", userID)
		return domain.ErrNotFound
	}
	prog, err := domain.NewOrUpdatePlaybackProgress(userID, trackID, progress)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid progress value provided", "error", err, "userID", userID, "trackID", trackID, "progress", progress)
		return err
	}
	if err := uc.progressRepo.Upsert(ctx, prog); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to upsert playback progress", "error", err, "userID", userID, "trackID", trackID)
		return fmt.Errorf("failed to save progress: %w", err)
	}
	uc.logger.InfoContext(ctx, "Playback progress recorded", "userID", userID, "trackID", trackID, "progressMs", progress.Milliseconds())
	return nil
}

// GetPlaybackProgress retrieves the user's progress for a specific track.
func (uc *UserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
    progress, err := uc.progressRepo.Find(ctx, userID, trackID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to get playback progress", "error", err, "userID", userID, "trackID", trackID) }
		return nil, err
	}
	return progress, nil
}

// ListUserProgress retrieves a paginated list of all progress records for a user.
func (uc *UserActivityUseCase) ListUserProgress(ctx context.Context, userID domain.UserID, limit, offset int) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
	pageParams := pagination.NewPageFromOffset(limit, offset)
	progressList, total, err := uc.progressRepo.ListByUser(ctx, userID, pageParams)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list user progress", "error", err, "userID", userID, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve progress list: %w", err)
	}
	return progressList, total, pageParams, nil
}

// --- Bookmark Use Cases ---

// CreateBookmark creates a new bookmark for the user on a specific track.
func (uc *UserActivityUseCase) CreateBookmark(ctx context.Context, userID domain.UserID, trackID domain.TrackID, timestamp time.Duration, note string) (*domain.Bookmark, error) {
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for bookmark creation", "error", err, "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("failed to validate track: %w", err)
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to create bookmark for non-existent track", "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("%w: track not found", domain.ErrNotFound)
	}
	bookmark, err := domain.NewBookmark(userID, trackID, timestamp, note)
	if err != nil {
		uc.logger.WarnContext(ctx, "Invalid bookmark timestamp provided", "error", err, "userID", userID, "trackID", trackID, "timestampMs", timestamp.Milliseconds())
		return nil, err
	}
	if err := uc.bookmarkRepo.Create(ctx, bookmark); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save bookmark", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}
	uc.logger.InfoContext(ctx, "Bookmark created", "bookmarkID", bookmark.ID, "userID", userID, "trackID", trackID)
	return bookmark, nil
}

// ListBookmarks retrieves bookmarks for a user, optionally filtered by track.
// CORRECTED: Pagination is only applied when listing *all* bookmarks for a user.
// When filtering by trackId, *all* bookmarks for that track are returned.
func (uc *UserActivityUseCase) ListBookmarks(ctx context.Context, userID domain.UserID, trackIDFilter *domain.TrackID, limit, offset int) ([]*domain.Bookmark, int, pagination.Page, error) {
	var bookmarks []*domain.Bookmark
	var total int
	var err error
	var pageParams pagination.Page // Will hold the *actual* pagination info used/returned

	if trackIDFilter != nil {
		// --- Listing for a specific track: Fetch ALL, NO standard pagination applied ---
		bookmarks, err = uc.bookmarkRepo.ListByUserAndTrack(ctx, userID, *trackIDFilter)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user and track", "error", err, "userID", userID, "trackID", *trackIDFilter)
			pageParams = pagination.NewPageFromOffset(0, 0) // Return zero page on error
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks for track: %w", err)
		}
		total = len(bookmarks)
		// Return pagination info reflecting that all results were returned (limit=total, offset=0)
		// This informs the client that no further pages exist for this specific filter.
		pageParams = pagination.Page{Limit: total, Offset: 0}
		// Use DefaultLimit if total is 0 to avoid limit 0 in response
		if total == 0 { pageParams.Limit = pagination.DefaultLimit }

	} else {
		// --- Listing all bookmarks for the user (PAGINATED) ---
		pageParams = pagination.NewPageFromOffset(limit, offset) // Apply defaults/constraints
		bookmarks, total, err = uc.bookmarkRepo.ListByUser(ctx, userID, pageParams)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user", "error", err, "userID", userID, "page", pageParams)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks: %w", err)
		}
		// pageParams is already the constrained, correct page info used for the query
	}

	// Success case for both scenarios
	return bookmarks, total, pageParams, nil
}

// DeleteBookmark deletes a bookmark owned by the user.
func (uc *UserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	bookmark, err := uc.bookmarkRepo.FindByID(ctx, bookmarkID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to find bookmark for deletion", "error", err, "bookmarkID", bookmarkID, "userID", userID) }
		return err
	}
	if bookmark.UserID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete bookmark not owned by user", "bookmarkID", bookmarkID, "ownerID", bookmark.UserID, "userID", userID)
		return domain.ErrPermissionDenied
	}
	// Use bookmark ID directly for deletion, ownership already checked
	if err := uc.bookmarkRepo.Delete(ctx, bookmarkID); err != nil {
		// Handles ErrNotFound race condition gracefully
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to delete bookmark from repository", "error", err, "bookmarkID", bookmarkID, "userID", userID) }
		// Return the error from the Delete operation
		return err
	}
	uc.logger.InfoContext(ctx, "Bookmark deleted", "bookmarkID", bookmarkID, "userID", userID)
	return nil
}

// Compile-time check to ensure UserActivityUseCase satisfies the port.UserActivityUseCase interface
var _ port.UserActivityUseCase = (*UserActivityUseCase)(nil)
