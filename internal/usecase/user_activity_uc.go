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
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                      // Import pagination
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
	// 1. Validate track exists (optional but good practice)
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for progress update", "error", err, "trackID", trackID, "userID", userID)
		return fmt.Errorf("failed to validate track: %w", err) // Internal error
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to record progress for non-existent track", "trackID", trackID, "userID", userID)
		return domain.ErrNotFound // Track not found
	}

	// 2. Create domain object (or just pass values to repo if simple)
	prog, err := domain.NewOrUpdatePlaybackProgress(userID, trackID, progress)
	if err != nil {
		// Should only be invalid argument error (negative progress)
		uc.logger.WarnContext(ctx, "Invalid progress value provided", "error", err, "userID", userID, "trackID", trackID, "progress", progress)
		return err
	}

	// 3. Call repo Upsert
	if err := uc.progressRepo.Upsert(ctx, prog); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to upsert playback progress", "error", err, "userID", userID, "trackID", trackID)
		return fmt.Errorf("failed to save progress: %w", err) // Internal error
	}

	uc.logger.InfoContext(ctx, "Playback progress recorded", "userID", userID, "trackID", trackID, "progress", progress)
	return nil
}


// GetPlaybackProgress retrieves the user's progress for a specific track.
func (uc *UserActivityUseCase) GetPlaybackProgress(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
    progress, err := uc.progressRepo.Find(ctx, userID, trackID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { // Log only unexpected errors
			uc.logger.ErrorContext(ctx, "Failed to get playback progress", "error", err, "userID", userID, "trackID", trackID)
		}
		// Return ErrNotFound or other errors directly
		return nil, err
	}
	return progress, nil
}


// ListUserProgress retrieves a paginated list of all progress records for a user.
func (uc *UserActivityUseCase) ListUserProgress(ctx context.Context, userID domain.UserID, limit, offset int) ([]*domain.PlaybackProgress, int, pagination.Page, error) {
    // Create Page object, applying defaults and constraints
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
	// 1. Validate track exists
	exists, err := uc.trackRepo.Exists(ctx, trackID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to check track existence for bookmark creation", "error", err, "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("failed to validate track: %w", err) // Internal error
	}
	if !exists {
		uc.logger.WarnContext(ctx, "Attempt to create bookmark for non-existent track", "trackID", trackID, "userID", userID)
		return nil, fmt.Errorf("%w: track not found", domain.ErrNotFound) // More specific than just ErrNotFound
	}

	// 2. Create domain object
	bookmark, err := domain.NewBookmark(userID, trackID, timestamp, note)
	if err != nil {
		// Should only be invalid argument error (negative timestamp)
		uc.logger.WarnContext(ctx, "Invalid bookmark timestamp provided", "error", err, "userID", userID, "trackID", trackID, "timestamp", timestamp)
		return nil, err
	}

	// 3. Save to repository
	if err := uc.bookmarkRepo.Create(ctx, bookmark); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save bookmark", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("failed to create bookmark: %w", err) // Internal error
	}

	uc.logger.InfoContext(ctx, "Bookmark created", "bookmarkID", bookmark.ID, "userID", userID, "trackID", trackID)
	return bookmark, nil
}

// ListBookmarks retrieves bookmarks for a user, optionally filtered by track.
func (uc *UserActivityUseCase) ListBookmarks(ctx context.Context, userID domain.UserID, trackID *domain.TrackID, limit, offset int) ([]*domain.Bookmark, int, pagination.Page, error) {
	var bookmarks []*domain.Bookmark
	var total int
	var err error
	var pageParams pagination.Page

	if trackID != nil {
		// Listing for a specific track (repo method doesn't support pagination, returns all)
		bookmarks, err = uc.bookmarkRepo.ListByUserAndTrack(ctx, userID, *trackID)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user and track", "error", err, "userID", userID, "trackID", *trackID)
			// Still return a default Page struct on error?
			// Let's return a Page reflecting the requested (but failed) fetch - could be misleading.
			// Returning a zero Page might be better.
			pageParams = pagination.NewPageFromOffset(0, 0) // Or just pagination.Page{}
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks for track: %w", err)
		}
		total = len(bookmarks)
		// Create Page reflecting the *actual* result (all bookmarks for the track)
		// NewPageFromOffset applies limits, which is not correct here.
		pageParams = pagination.Page{Limit: total, Offset: 0}
		// Ensure limit does not exceed MaxLimit if that's a strict requirement downstream,
		// although NewPaginatedResponse will handle this again.
		if pageParams.Limit > pagination.MaxLimit {
			pageParams.Limit = pagination.MaxLimit // Optional: Apply constraint here too?
		}

	} else {
		// Listing all bookmarks for the user (paginated)
		pageParams = pagination.NewPageFromOffset(limit, offset)
		bookmarks, total, err = uc.bookmarkRepo.ListByUser(ctx, userID, pageParams)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to list bookmarks by user", "error", err, "userID", userID, "page", pageParams)
			return nil, 0, pageParams, fmt.Errorf("failed to retrieve bookmarks: %w", err)
		}
	}

	// Success case for both scenarios (trackID nil or not nil)
	return bookmarks, total, pageParams, nil
}

// DeleteBookmark deletes a bookmark owned by the user.
func (uc *UserActivityUseCase) DeleteBookmark(ctx context.Context, userID domain.UserID, bookmarkID domain.BookmarkID) error {
	// 1. Verify ownership (fetch bookmark first)
	bookmark, err := uc.bookmarkRepo.FindByID(ctx, bookmarkID)
	if err != nil {
		// Handles ErrNotFound
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to find bookmark for deletion", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err
	}

	if bookmark.UserID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete bookmark not owned by user", "bookmarkID", bookmarkID, "ownerID", bookmark.UserID, "userID", userID)
		return domain.ErrPermissionDenied
	}

	// 2. Delete from repository
	if err := uc.bookmarkRepo.Delete(ctx, bookmarkID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) { // Handle potential race condition where it was already deleted
			uc.logger.ErrorContext(ctx, "Failed to delete bookmark from repository", "error", err, "bookmarkID", bookmarkID, "userID", userID)
		}
		return err // Return original error (includes ErrNotFound if race occurred)
	}

	uc.logger.InfoContext(ctx, "Bookmark deleted", "bookmarkID", bookmarkID, "userID", userID)
	return nil
}