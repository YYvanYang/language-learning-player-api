// internal/domain/bookmark.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BookmarkID is the unique identifier for a Bookmark.
type BookmarkID uuid.UUID

func NewBookmarkID() BookmarkID {
	return BookmarkID(uuid.New())
}

func BookmarkIDFromString(s string) (BookmarkID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return BookmarkID{}, fmt.Errorf("invalid BookmarkID format: %w", err)
	}
	return BookmarkID(id), nil
}

func (bid BookmarkID) String() string {
	return uuid.UUID(bid).String()
}

// Bookmark represents a specific point in an audio track saved by a user.
type Bookmark struct {
	ID        BookmarkID
	UserID    UserID
	TrackID   TrackID
	Timestamp time.Duration // Position in the audio track where the bookmark is placed
	Note      string        // Optional user note about the bookmark
	CreatedAt time.Time
}

// NewBookmark creates a new bookmark instance.
func NewBookmark(userID UserID, trackID TrackID, timestamp time.Duration, note string) (*Bookmark, error) {
	if timestamp < 0 {
		return nil, fmt.Errorf("%w: bookmark timestamp cannot be negative", ErrInvalidArgument)
	}
	// UserID and TrackID validity assumed

	return &Bookmark{
		ID:        NewBookmarkID(),
		UserID:    userID,
		TrackID:   trackID,
		Timestamp: timestamp,
		Note:      note,
		CreatedAt: time.Now(),
	}, nil
}