package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOrUpdatePlaybackProgress(t *testing.T) {
	userID := NewUserID()
	trackID := NewTrackID()

	tests := []struct {
		name     string
		userID   UserID
		trackID  TrackID
		progress time.Duration
		wantErr  bool
		errType  error
	}{
		{"Valid progress", userID, trackID, 30 * time.Second, false, nil},
		{"Zero progress", userID, trackID, 0, false, nil},
		{"Negative progress", userID, trackID, -10 * time.Millisecond, true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewOrUpdatePlaybackProgress(tt.userID, tt.trackID, tt.progress)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.userID, got.UserID)
				assert.Equal(t, tt.trackID, got.TrackID)
				assert.Equal(t, tt.progress, got.Progress)
				assert.WithinDuration(t, start, got.LastListenedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
