package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBookmarkID(t *testing.T) {
	id1 := NewBookmarkID()
	id2 := NewBookmarkID()
	assert.NotEqual(t, BookmarkID{}, id1)
	assert.NotEqual(t, id1, id2)
}

func TestBookmarkIDFromString(t *testing.T) {
	validUUID := "d4e5f418-3150-4d1b-9c0a-4b8f8e7a8e21"
	invalidUUID := "not-a-uuid"

	id, err := BookmarkIDFromString(validUUID)
	assert.NoError(t, err)
	assert.Equal(t, validUUID, id.String())

	_, err = BookmarkIDFromString(invalidUUID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid BookmarkID format")
}

func TestNewBookmark(t *testing.T) {
	userID := NewUserID()
	trackID := NewTrackID()

	tests := []struct {
		name      string
		userID    UserID
		trackID   TrackID
		timestamp time.Duration
		note      string
		wantErr   bool
		errType   error
	}{
		{"Valid bookmark", userID, trackID, 15 * time.Second, "Interesting point", false, nil},
		{"Valid bookmark zero timestamp", userID, trackID, 0, "Start", false, nil},
		{"Valid bookmark empty note", userID, trackID, 1 * time.Minute, "", false, nil},
		{"Invalid negative timestamp", userID, trackID, -5 * time.Second, "Should fail", true, ErrInvalidArgument},
		// UserID and TrackID validity assumed to be handled elsewhere
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewBookmark(tt.userID, tt.trackID, tt.timestamp, tt.note)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, BookmarkID{}, got.ID)
				assert.Equal(t, tt.userID, got.UserID)
				assert.Equal(t, tt.trackID, got.TrackID)
				assert.Equal(t, tt.timestamp, got.Timestamp)
				assert.Equal(t, tt.note, got.Note)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
