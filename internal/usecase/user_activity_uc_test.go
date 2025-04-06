// ==================================================
// FILE: internal/usecase/user_activity_uc_test.go
// ==================================================
// This file was already provided in all_code.md and looks relatively complete.
// Needs mock definitions from `internal/port/mocks`.
package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use actual mocks import path
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use logger helper if available
func newTestLoggerForActivityUC() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestUserActivityUseCase_RecordPlaybackProgress_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t) // Needed for constructor
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()

	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := 30 * time.Second

	// Expect TrackRepo.Exists to be called and return true
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(true, nil).Once()

	// Expect ProgressRepo.Upsert to be called
	mockProgressRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(p *domain.PlaybackProgress) bool {
		return p.UserID == userID && p.TrackID == trackID && p.Progress == progress
	})).Return(nil).Once()

	// Execute
	err := uc.RecordPlaybackProgress(context.Background(), userID, trackID, progress)

	// Assert
	require.NoError(t, err)
	mockTrackRepo.AssertExpectations(t)
	mockProgressRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_RecordPlaybackProgress_TrackNotFound(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := 30 * time.Second

	// Expect TrackRepo.Exists to return false
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(false, nil).Once()

	// Execute
	err := uc.RecordPlaybackProgress(context.Background(), userID, trackID, progress)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound) // Expect NotFound because track doesn't exist

	mockTrackRepo.AssertExpectations(t)
	// Ensure Upsert was not called
	mockProgressRepo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
}

func TestUserActivityUseCase_GetPlaybackProgress_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	expectedProgress := &domain.PlaybackProgress{
		UserID:         userID,
		TrackID:        trackID,
		Progress:       45 * time.Second,
		LastListenedAt: time.Now().Add(-5 * time.Minute),
	}

	// Expect Find to be called and return progress
	mockProgressRepo.On("Find", mock.Anything, userID, trackID).Return(expectedProgress, nil).Once()

	// Execute
	progress, err := uc.GetPlaybackProgress(context.Background(), userID, trackID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProgress, progress)
	mockProgressRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_GetPlaybackProgress_NotFound(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()

	// Expect Find to return NotFound
	mockProgressRepo.On("Find", mock.Anything, userID, trackID).Return(nil, domain.ErrNotFound).Once()

	// Execute
	progress, err := uc.GetPlaybackProgress(context.Background(), userID, trackID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, progress)
	mockProgressRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_ListUserProgress_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	reqPageParams := pagination.Page{Limit: 10, Offset: 0} // Request pagination
	ucParams := port.ListProgressParams{UserID: userID, Page: reqPageParams}
	// Page params after validation/defaults (assuming they match requested)
	expectedPageParams := pagination.NewPageFromOffset(reqPageParams.Limit, reqPageParams.Offset)

	expectedProgress := []*domain.PlaybackProgress{
		{UserID: userID, TrackID: domain.NewTrackID(), Progress: 10 * time.Second},
		{UserID: userID, TrackID: domain.NewTrackID(), Progress: 20 * time.Second},
	}
	expectedTotal := 5

	// Expect ListByUser to be called with constrained page params
	mockProgressRepo.On("ListByUser", mock.Anything, userID, expectedPageParams).Return(expectedProgress, expectedTotal, nil).Once()

	// Execute
	progressList, total, actualPage, err := uc.ListUserProgress(context.Background(), ucParams)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProgress, progressList)
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, expectedPageParams, actualPage) // Ensure returned page matches expected used page
	mockProgressRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_CreateBookmark_Success(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	timestamp := 60 * time.Second
	note := "Important point"

	// Expect TrackRepo.Exists to return true
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(true, nil).Once()

	// Expect BookmarkRepo.Create to be called
	// We capture the bookmark to check its details
	var createdBookmark *domain.Bookmark
	mockBookmarkRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Bookmark) bool {
		createdBookmark = b
		return b.UserID == userID && b.TrackID == trackID && b.Timestamp == timestamp && b.Note == note
	})).Return(nil).Once()


	// Execute
	bookmark, err := uc.CreateBookmark(context.Background(), userID, trackID, timestamp, note)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, bookmark)
	// Check if the returned bookmark matches the one captured (or has expected fields)
	// ID is generated, so we might only check other fields match the input/capture
    if createdBookmark != nil { // Check if capture was successful
	    assert.Equal(t, createdBookmark.ID, bookmark.ID) // ID is generated by domain/repo
    } else {
        assert.NotEqual(t, domain.BookmarkID{}, bookmark.ID) // Ensure ID was generated
    }
	assert.Equal(t, userID, bookmark.UserID)
	assert.Equal(t, trackID, bookmark.TrackID)
	assert.Equal(t, timestamp, bookmark.Timestamp)
	assert.Equal(t, note, bookmark.Note)

	mockTrackRepo.AssertExpectations(t)
	mockBookmarkRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_CreateBookmark_TrackNotFound(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()

	// Expect TrackRepo.Exists to return false
	mockTrackRepo.On("Exists", mock.Anything, trackID).Return(false, nil).Once()

	// Execute
	bookmark, err := uc.CreateBookmark(context.Background(), userID, trackID, 10*time.Second, "")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, bookmark)
	mockTrackRepo.AssertExpectations(t)
	mockBookmarkRepo.AssertNotCalled(t, "Create") // Create should not be called
}

func TestUserActivityUseCase_DeleteBookmark_SuccessOwned(t *testing.T) {
    mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
    mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    logger := newTestLoggerForActivityUC()
    uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

    userID := domain.NewUserID()    // The user performing the action
    bookmarkID := domain.NewBookmarkID()
    trackID := domain.NewTrackID()

    // The bookmark to be deleted, owned by the current user
    ownedBookmark := &domain.Bookmark{
        ID:        bookmarkID,
        UserID:    userID, // Belongs to the user
        TrackID:   trackID,
        Timestamp: 10 * time.Second,
    }

    // Expect FindByID to find the owned bookmark
    mockBookmarkRepo.On("FindByID", mock.Anything, bookmarkID).Return(ownedBookmark, nil).Once()
    // Expect Delete to be called with the correct ID
    mockBookmarkRepo.On("Delete", mock.Anything, bookmarkID).Return(nil).Once()

    // Execute
    err := uc.DeleteBookmark(context.Background(), userID, bookmarkID)

    // Assert
    require.NoError(t, err)
    mockBookmarkRepo.AssertExpectations(t)
}

func TestUserActivityUseCase_DeleteBookmark_PermissionDenied(t *testing.T) {
    mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
    mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    logger := newTestLoggerForActivityUC()
    uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

    requestingUserID := domain.NewUserID() // The user trying to delete
    ownerUserID := domain.NewUserID()      // The actual owner (different)
    bookmarkID := domain.NewBookmarkID()
    trackID := domain.NewTrackID()

    // Bookmark owned by someone else
    otherUsersBookmark := &domain.Bookmark{
        ID:        bookmarkID,
        UserID:    ownerUserID, // Belongs to a different user
        TrackID:   trackID,
        Timestamp: 10 * time.Second,
    }

    // Expect FindByID to find the bookmark
    mockBookmarkRepo.On("FindByID", mock.Anything, bookmarkID).Return(otherUsersBookmark, nil).Once()

    // Execute
    err := uc.DeleteBookmark(context.Background(), requestingUserID, bookmarkID)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrPermissionDenied) // Expect permission denied

    mockBookmarkRepo.AssertExpectations(t) // FindByID was called
    // Ensure Delete was NOT called
    mockBookmarkRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestUserActivityUseCase_DeleteBookmark_NotFound(t *testing.T) {
    mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
    mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    logger := newTestLoggerForActivityUC()
    uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

    userID := domain.NewUserID()
    bookmarkID := domain.NewBookmarkID()

    // Expect FindByID to return NotFound
    mockBookmarkRepo.On("FindByID", mock.Anything, bookmarkID).Return(nil, domain.ErrNotFound).Once()

    // Execute
    err := uc.DeleteBookmark(context.Background(), userID, bookmarkID)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrNotFound)
    mockBookmarkRepo.AssertExpectations(t)
    mockBookmarkRepo.AssertNotCalled(t, "Delete") // Delete should not be called
}

func TestUserActivityUseCase_ListBookmarks_SuccessAll(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	reqPageParams := pagination.Page{Limit: 15, Offset: 0}
	ucParams := port.ListBookmarksParams{UserID: userID, Page: reqPageParams}
	expectedPageParams := pagination.NewPageFromOffset(reqPageParams.Limit, reqPageParams.Offset) // Validated params

	expectedBookmarks := []*domain.Bookmark{
		{ID: domain.NewBookmarkID(), UserID: userID},
		{ID: domain.NewBookmarkID(), UserID: userID},
	}
	expectedTotal := 12

	// Expect ListByUser (paginated)
	mockBookmarkRepo.On("ListByUser", mock.Anything, userID, expectedPageParams).Return(expectedBookmarks, expectedTotal, nil).Once()

	// Execute
	bookmarks, total, actualPage, err := uc.ListBookmarks(context.Background(), ucParams)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBookmarks, bookmarks)
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, expectedPageParams, actualPage)
	mockBookmarkRepo.AssertExpectations(t)
	mockBookmarkRepo.AssertNotCalled(t, "ListByUserAndTrack") // Should not be called
}

func TestUserActivityUseCase_ListBookmarks_SuccessFilteredByTrack(t *testing.T) {
	mockProgressRepo := mocks.NewMockPlaybackProgressRepository(t)
	mockBookmarkRepo := mocks.NewMockBookmarkRepository(t)
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	logger := newTestLoggerForActivityUC()
	uc := usecase.NewUserActivityUseCase(mockProgressRepo, mockBookmarkRepo, mockTrackRepo, logger)

	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	reqPageParams := pagination.Page{Limit: 10, Offset: 0} // Note: Pagination is ignored for track filter in UC
	ucParams := port.ListBookmarksParams{UserID: userID, TrackIDFilter: &trackID, Page: reqPageParams}

	expectedBookmarks := []*domain.Bookmark{ // Assume repo returns all for the track
		{ID: domain.NewBookmarkID(), UserID: userID, TrackID: trackID},
		{ID: domain.NewBookmarkID(), UserID: userID, TrackID: trackID},
	}
	expectedTotal := len(expectedBookmarks)
	// Page returned reflects all results were returned
	expectedPageResult := pagination.Page{Limit: expectedTotal, Offset: 0}
    if expectedTotal == 0 { expectedPageResult.Limit = pagination.DefaultLimit } // Handle zero case


	// Expect ListByUserAndTrack
	mockBookmarkRepo.On("ListByUserAndTrack", mock.Anything, userID, trackID).Return(expectedBookmarks, nil).Once()

	// Execute
	bookmarks, total, actualPage, err := uc.ListBookmarks(context.Background(), ucParams)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBookmarks, bookmarks)
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, expectedPageResult, actualPage) // Page reflects all results returned
	mockBookmarkRepo.AssertExpectations(t)
	mockBookmarkRepo.AssertNotCalled(t, "ListByUser") // Should not be called
}