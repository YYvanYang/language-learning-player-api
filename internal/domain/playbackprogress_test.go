// ==================================================
// FILE: internal/domain/playbackprogress_test.go
// ==================================================
package domain_test

import (
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrUpdatePlaybackProgress_Success(t *testing.T) {
	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := 45 * time.Second + 200*time.Millisecond

	pp, err := domain.NewOrUpdatePlaybackProgress(userID, trackID, progress)

	require.NoError(t, err)
	require.NotNil(t, pp)
	assert.Equal(t, userID, pp.UserID)
	assert.Equal(t, trackID, pp.TrackID)
	assert.Equal(t, progress, pp.Progress)
	assert.WithinDuration(t, time.Now(), pp.LastListenedAt, time.Second)
}

func TestNewOrUpdatePlaybackProgress_NegativeProgress(t *testing.T) {
	userID := domain.NewUserID()
	trackID := domain.NewTrackID()
	progress := -10 * time.Second

	_, err := domain.NewOrUpdatePlaybackProgress(userID, trackID, progress)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}