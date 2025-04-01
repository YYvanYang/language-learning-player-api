// internal/domain/playbackprogress.go
package domain

import (
	"fmt"
	"time"
)

// PlaybackProgress represents a user's listening progress on a specific audio track.
// UserID and TrackID together form the primary key.
type PlaybackProgress struct {
	UserID         UserID
	TrackID        TrackID
	Progress       time.Duration // Current listening position
	LastListenedAt time.Time     // When the progress was last updated
}

// NewOrUpdatePlaybackProgress creates or updates a playback progress record.
func NewOrUpdatePlaybackProgress(userID UserID, trackID TrackID, progress time.Duration) (*PlaybackProgress, error) {
	if progress < 0 {
		return nil, fmt.Errorf("%w: progress duration cannot be negative", ErrInvalidArgument)
	}
	// UserID and TrackID validity assumed to be checked elsewhere (e.g., foreign keys or usecase)

	return &PlaybackProgress{
		UserID:         userID,
		TrackID:        trackID,
		Progress:       progress,
		LastListenedAt: time.Now(),
	}, nil
}