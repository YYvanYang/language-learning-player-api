// =============================================
// FILE: internal/domain/bookmark_test.go
// =============================================
package domain_test

import (
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBookmark_Success(t *testing.T) {
	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	timestamp := 15 * time.Second + 500*time.Millisecond
	note := "Check this phrase"

	bookmark, err := domain.NewBookmark(userID, trackID, timestamp, note)

	require.NoError(t, err)
	require.NotNil(t, bookmark)
	assert.NotEqual(t, domain.BookmarkID{}, bookmark.ID)
	assert.Equal(t, userID, bookmark.UserID)
	assert.Equal(t, trackID, bookmark.TrackID)
	assert.Equal(t, timestamp, bookmark.Timestamp)
	assert.Equal(t, note, bookmark.Note)
	assert.WithinDuration(t, time.Now(), bookmark.CreatedAt, time.Second)
}

func TestNewBookmark_NegativeTimestamp(t *testing.T) {
	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	timestamp := -5 * time.Second

	_, err := domain.NewBookmark(userID, trackID, timestamp, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}