// ============================================
// FILE: internal/usecase/user_activity_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

type UserActivityUseCase struct {
	progressRepo port.PlaybackProgressRepository
	bookmarkRepo port.BookmarkRepository
	trackRepo    port.AudioTrackRepository
	logger       *slog.Logger
}

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

func (uc *UserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	progress, err := uc.progressRepo.Find(ctx, userID, trackID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to get playback progress", "error", err, "userID", userID, "trackID", trackID)
		}
		return nil, err
	}
	return progress, nil
}

// ListUserProgress retrieves a paginated list of all progress records for a user.
// CHANGED: Parameter type to port.ListProgressInput
func (uc *UserActivityUseCase) ListUserProgress(ctx context.Context, input port.ListProgressInput) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
	// CHANGED: Use input.Page
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)
	// CHANGED: Use input.UserID
	progressList, total, err := uc.progressRepo.ListByUser(ctx, input.UserID, pageParams)
	if err != nil {
		// CHANGED: Use input.UserID
		uc.logger.ErrorContext(ctx, "Failed to list user progress", "error", err, "userID", input.UserID, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve progress list: %w", err)
	}
	return progressList, total, pageParams, nil
}

// --- Bookmark Use Cases ---

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
// CHANGED: Parameter type to port.ListBookmarksInput
func (uc *UserActivityUseCase) ListBookmarks(ctx context.Context, input port.ListBookmarksInput) ([]*domain.Bookmark, int, pagination.Page, error) {
	var bookmarks []*domain.Bookmark
	var total int
	var err error
	// CHANGED: Use input.Page
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)

	// CHANGED: Use input.TrackIDFilter and input.UserID
	if input.TrackIDFilter != nil {
		bookmarks, err = uc.bookmarkRepo.ListByUserAndTrack(ctx, input.UserID, *input.TrackIDFilter)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user and track", "error", err, "userID", input.UserID, "trackID", *input.TrackIDFilter)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks for track: %w", err)
		}
		total = len(bookmarks)
		pageParams = pagination.Page{Limit: total, Offset: 0}
		if total == 0 {
			pageParams.Limit = pagination.DefaultLimit
		}
	} else {
		// CHANGED: Use input.UserID
		bookmarks, total, err = uc.bookmarkRepo.ListByUser(ctx, input.UserID, pageParams)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user", "error", err, "userID", input.UserID, "page", pageParams)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks: %w", err)
		}
	}
	return bookmarks, total, pageParams, nil
}

func (uc *UserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	bookmark, err := uc.bookmarkRepo.FindByID(ctx, bookmarkID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to find bookmark for deletion", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err
	}
	if bookmark.UserID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete bookmark not owned by user", "bookmarkID", bookmarkID, "ownerID", bookmark.UserID, "userID", userID)
		return domain.ErrPermissionDenied
	}
	// CHANGED: Pass only bookmarkID to Delete, ownership check is done above
	if err := uc.bookmarkRepo.Delete(ctx, bookmarkID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to delete bookmark from repository", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err // Return NotFound if repo returns it (double deletion attempt)
	}
	uc.logger.InfoContext(ctx, "Bookmark deleted", "bookmarkID", bookmarkID, "userID", userID)
	return nil
}

var _ port.UserActivityUseCase = (*UserActivityUseCase)(nil)
