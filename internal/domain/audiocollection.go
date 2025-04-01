// internal/domain/audiocollection.go
package domain

import (
	"fmt"
	"time"
	"slices"

	"github.com/google/uuid"
)

// CollectionID is the unique identifier for an AudioCollection.
type CollectionID uuid.UUID

func NewCollectionID() CollectionID {
	return CollectionID(uuid.New())
}

func CollectionIDFromString(s string) (CollectionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return CollectionID{}, fmt.Errorf("invalid CollectionID format: %w", err)
	}
	return CollectionID(id), nil
}

func (cid CollectionID) String() string {
	return uuid.UUID(cid).String()
}


// AudioCollection represents a curated list of audio tracks (e.g., a course or playlist).
type AudioCollection struct {
	ID          CollectionID
	Title       string
	Description string
	OwnerID     UserID // The user who owns/created the collection
	Type        CollectionType // Value object (COURSE or PLAYLIST)
	TrackIDs    []TrackID // Ordered list of TrackIDs in the collection
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAudioCollection creates a new audio collection.
func NewAudioCollection(title, description string, ownerID UserID, colType CollectionType) (*AudioCollection, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: collection title cannot be empty", ErrInvalidArgument)
	}
	if !colType.IsValid() || colType == TypeUnknown {
		return nil, fmt.Errorf("%w: invalid collection type '%s'", ErrInvalidArgument, colType)
	}
	// OwnerID validation happens implicitly via foreign key constraint usually

	now := time.Now()
	return &AudioCollection{
		ID:          NewCollectionID(),
		Title:       title,
		Description: description,
		OwnerID:     ownerID,
		Type:        colType,
		TrackIDs:    make([]TrackID, 0), // Initialize empty slice
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddTrack adds a track ID to the collection at a specific position.
// Position is 0-based. If position is out of bounds, it appends to the end.
func (c *AudioCollection) AddTrack(trackID TrackID, position int) error {
	// Check if track already exists
	for _, existingID := range c.TrackIDs {
		if existingID == trackID {
			return fmt.Errorf("%w: track %s already exists in collection %s", ErrConflict, trackID, c.ID)
		}
	}

	if position < 0 || position > len(c.TrackIDs) {
		position = len(c.TrackIDs) // Append to end if out of bounds
	}

	// Insert element at position
	c.TrackIDs = slices.Insert(c.TrackIDs, position, trackID)
	c.UpdatedAt = time.Now()
	return nil
}

// RemoveTrack removes a track ID from the collection.
func (c *AudioCollection) RemoveTrack(trackID TrackID) bool {
	initialLen := len(c.TrackIDs)
	c.TrackIDs = slices.DeleteFunc(c.TrackIDs, func(id TrackID) bool {
		return id == trackID
	})
	removed := len(c.TrackIDs) < initialLen
	if removed {
		c.UpdatedAt = time.Now()
	}
	return removed
}

// ReorderTracks sets the track order to the provided list of IDs.
// It ensures all provided IDs were already present in the collection.
func (c *AudioCollection) ReorderTracks(orderedTrackIDs []TrackID) error {
	if len(orderedTrackIDs) != len(c.TrackIDs) {
		return fmt.Errorf("%w: number of provided track IDs (%d) does not match current number (%d)", ErrInvalidArgument, len(orderedTrackIDs), len(c.TrackIDs))
	}

	// Check if all original tracks are present in the new order exactly once
	currentSet := make(map[TrackID]struct{}, len(c.TrackIDs))
	for _, id := range c.TrackIDs {
		currentSet[id] = struct{}{}
	}
	newSet := make(map[TrackID]struct{}, len(orderedTrackIDs))
	for _, id := range orderedTrackIDs {
		if _, exists := currentSet[id]; !exists {
			return fmt.Errorf("%w: track ID %s is not part of the original collection", ErrInvalidArgument, id)
		}
		if _, duplicate := newSet[id]; duplicate {
			return fmt.Errorf("%w: track ID %s appears multiple times in the new order", ErrInvalidArgument, id)
		}
		newSet[id] = struct{}{}
	}

	c.TrackIDs = orderedTrackIDs
	c.UpdatedAt = time.Now()
	return nil
}