package domain

import (
	"testing"
	"time"

	// 确保导入 slices 包
	"github.com/stretchr/testify/assert"
)

func TestNewCollectionID(t *testing.T) {
	id1 := NewCollectionID()
	id2 := NewCollectionID()
	assert.NotEqual(t, CollectionID{}, id1) // Not zero value
	assert.NotEqual(t, id1, id2)            // IDs should be unique
}

func TestCollectionIDFromString(t *testing.T) {
	validUUID := "a4a5b418-2150-4d1b-9c0a-4b8f8e7a8e21"
	invalidUUID := "not-a-uuid"

	id, err := CollectionIDFromString(validUUID)
	assert.NoError(t, err)
	assert.Equal(t, validUUID, id.String())

	_, err = CollectionIDFromString(invalidUUID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CollectionID format")
}

func TestNewCollection(t *testing.T) {
	ownerID := NewUserID()

	tests := []struct {
		name        string
		title       string
		description string
		ownerID     UserID
		colType     CollectionType
		wantErr     bool
		errType     error
	}{
		{"Valid Course", "Intro Course", "Description", ownerID, TypeCourse, false, nil},
		{"Valid Playlist", "My Playlist", "", ownerID, TypePlaylist, false, nil}, // Empty description allowed
		{"Empty Title", "", "Description", ownerID, TypeCourse, true, ErrInvalidArgument},
		{"Invalid Type", "Title", "Desc", ownerID, CollectionType("INVALID"), true, ErrInvalidArgument},
		{"Unknown Type", "Title", "Desc", ownerID, TypeUnknown, true, ErrInvalidArgument}, // Unknown type not allowed by constructor
		// OwnerID validation usually done by DB FK constraint, not here
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewAudioCollection(tt.title, tt.description, tt.ownerID, tt.colType)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, CollectionID{}, got.ID)
				assert.Equal(t, tt.title, got.Title)
				assert.Equal(t, tt.description, got.Description)
				assert.Equal(t, tt.ownerID, got.OwnerID)
				assert.Equal(t, tt.colType, got.Type)
				assert.NotNil(t, got.TrackIDs) // Should be initialized
				assert.Empty(t, got.TrackIDs)  // Should be empty initially
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestAudioCollection_TrackManagement(t *testing.T) {
	ownerID := NewUserID()
	collection, _ := NewAudioCollection("Test Collection", "", ownerID, TypePlaylist)
	track1 := NewTrackID()
	track2 := NewTrackID()
	track3 := NewTrackID()
	initialTime := collection.UpdatedAt

	// --- AddTrack ---
	t.Run("AddTrack", func(t *testing.T) {
		// Add first track
		err := collection.AddTrack(track1, 0)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track1}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(initialTime))
		timeAfterAdd1 := collection.UpdatedAt

		// Add second track at beginning
		err = collection.AddTrack(track2, 0)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track2, track1}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterAdd1))
		timeAfterAdd2 := collection.UpdatedAt

		// Add third track at end (position out of bounds)
		err = collection.AddTrack(track3, 10) // Position 10 is > len(2)
		assert.NoError(t, err)
		assert.Equal(t, []TrackID{track2, track1, track3}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterAdd2))
		timeAfterAdd3 := collection.UpdatedAt

		// Add existing track (should error)
		err = collection.AddTrack(track1, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrConflict)
		assert.Equal(t, []TrackID{track2, track1, track3}, collection.TrackIDs) // No change
		assert.Equal(t, timeAfterAdd3, collection.UpdatedAt)                    // Timestamp unchanged
	})

	// Reset state for next tests if needed, or use subtests carefully
	collection.TrackIDs = []TrackID{track2, track1, track3} // Reset state
	timeBeforeRemove := collection.UpdatedAt

	// --- RemoveTrack ---
	t.Run("RemoveTrack", func(t *testing.T) {
		// Remove middle track
		removed := collection.RemoveTrack(track1)
		assert.True(t, removed)
		assert.Equal(t, []TrackID{track2, track3}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeBeforeRemove))
		timeAfterRemove1 := collection.UpdatedAt

		// Remove non-existent track
		removed = collection.RemoveTrack(NewTrackID()) // New random ID
		assert.False(t, removed)
		assert.Equal(t, []TrackID{track2, track3}, collection.TrackIDs) // No change
		assert.Equal(t, timeAfterRemove1, collection.UpdatedAt)         // Timestamp unchanged

		// Remove last track
		removed = collection.RemoveTrack(track3)
		assert.True(t, removed)
		assert.Equal(t, []TrackID{track2}, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterRemove1))
		timeAfterRemove2 := collection.UpdatedAt

		// Remove first track (last remaining)
		removed = collection.RemoveTrack(track2)
		assert.True(t, removed)
		assert.Empty(t, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeAfterRemove2))
	})

	// --- ReorderTracks ---
	t.Run("ReorderTracks", func(t *testing.T) {
		// Reset state
		collection.TrackIDs = []TrackID{track1, track2, track3}
		timeBeforeReorder := time.Now()
		collection.UpdatedAt = timeBeforeReorder // Set known time

		// Valid reorder
		newOrder := []TrackID{track3, track1, track2}
		err := collection.ReorderTracks(newOrder)
		assert.NoError(t, err)
		assert.Equal(t, newOrder, collection.TrackIDs)
		assert.True(t, collection.UpdatedAt.After(timeBeforeReorder))
		timeAfterReorder1 := collection.UpdatedAt

		// Error: Incorrect number of tracks
		wrongCountOrder := []TrackID{track1, track3}
		err = collection.ReorderTracks(wrongCountOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "number of provided track IDs")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Error: Track not originally present
		track4 := NewTrackID()
		notPresentOrder := []TrackID{track1, track2, track4}
		err = collection.ReorderTracks(notPresentOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "is not part of the original collection")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Error: Duplicate track in new order
		duplicateOrder := []TrackID{track1, track1, track2}
		err = collection.ReorderTracks(duplicateOrder)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "appears multiple times")
		assert.Equal(t, newOrder, collection.TrackIDs)           // Order unchanged
		assert.Equal(t, timeAfterReorder1, collection.UpdatedAt) // Timestamp unchanged

		// Reorder to empty list (should fail if original list not empty)
		collection.TrackIDs = []TrackID{track1} // Set to non-empty
		err = collection.ReorderTracks([]TrackID{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidArgument)
		assert.Contains(t, err.Error(), "number of provided track IDs")

		// Reorder empty list to empty list (should succeed)
		collection.TrackIDs = []TrackID{} // Set to empty
		timeBeforeEmptyReorder := time.Now()
		collection.UpdatedAt = timeBeforeEmptyReorder
		err = collection.ReorderTracks([]TrackID{})
		assert.NoError(t, err)
		assert.Empty(t, collection.TrackIDs)
		// assert.True(t, collection.UpdatedAt.After(timeBeforeEmptyReorder)) // debatable if timestamp should change here
	})
}
