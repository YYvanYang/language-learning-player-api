// ================================================
// FILE: internal/domain/audiocollection_test.go
// ================================================
package domain_test

import (
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAudioCollection_Success(t *testing.T) {
	title := "My French Course"
	desc := "Beginner French lessons"
	ownerID := domain.NewUserID()
	colType := domain.TypeCourse

	collection, err := domain.NewAudioCollection(title, desc, ownerID, colType)

	require.NoError(t, err)
	require.NotNil(t, collection)
	assert.NotEqual(t, domain.CollectionID{}, collection.ID)
	assert.Equal(t, title, collection.Title)
	assert.Equal(t, desc, collection.Description)
	assert.Equal(t, ownerID, collection.OwnerID)
	assert.Equal(t, colType, collection.Type)
	assert.Empty(t, collection.TrackIDs) // Starts empty
	assert.WithinDuration(t, time.Now(), collection.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), collection.UpdatedAt, time.Second)
}

func TestNewAudioCollection_ValidationErrors(t *testing.T) {
	ownerID := domain.NewUserID()
	testCases := []struct {
		name      string
		title     string
		colType   domain.CollectionType
		expectErr bool
	}{
		{"Empty Title", "", domain.TypePlaylist, true},
		{"Invalid Type", "Valid Title", domain.CollectionType("INVALID"), true},
		{"Unknown Type", "Valid Title", domain.TypeUnknown, true}, // Should not allow unknown type on creation
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewAudioCollection(tc.title, "", ownerID, tc.colType)
			if tc.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidArgument)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAudioCollection_AddTrack(t *testing.T) {
	collection, _ := domain.NewAudioCollection("Test", "", domain.NewUserID(), domain.TypePlaylist)
	track1 := domain.NewTrackID()
	track2 := domain.NewTrackID()
	track3 := domain.NewTrackID()

	// Add first track
	err := collection.AddTrack(track1, 0)
	require.NoError(t, err)
	assert.Equal(t, []domain.TrackID{track1}, collection.TrackIDs)
	initialUpdate := collection.UpdatedAt

	// Add second track at end (invalid position)
	time.Sleep(1 * time.Millisecond)
	err = collection.AddTrack(track2, 99)
	require.NoError(t, err)
	assert.Equal(t, []domain.TrackID{track1, track2}, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))
	initialUpdate = collection.UpdatedAt

	// Add third track at beginning
	time.Sleep(1 * time.Millisecond)
	err = collection.AddTrack(track3, 0)
	require.NoError(t, err)
	assert.Equal(t, []domain.TrackID{track3, track1, track2}, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))

	// Try adding existing track
	err = collection.AddTrack(track1, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.Equal(t, []domain.TrackID{track3, track1, track2}, collection.TrackIDs) // Order unchanged
}

func TestAudioCollection_RemoveTrack(t *testing.T) {
	collection, _ := domain.NewAudioCollection("Test", "", domain.NewUserID(), domain.TypePlaylist)
	track1 := domain.NewTrackID()
	track2 := domain.NewTrackID()
	track3 := domain.NewTrackID()
	collection.TrackIDs = []domain.TrackID{track1, track2, track3}
	initialUpdate := collection.UpdatedAt

	// Remove middle track
	time.Sleep(1 * time.Millisecond)
	removed := collection.RemoveTrack(track2)
	assert.True(t, removed)
	assert.Equal(t, []domain.TrackID{track1, track3}, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))
	initialUpdate = collection.UpdatedAt

	// Remove non-existent track
	time.Sleep(1 * time.Millisecond)
	removed = collection.RemoveTrack(domain.NewTrackID())
	assert.False(t, removed)
	assert.Equal(t, []domain.TrackID{track1, track3}, collection.TrackIDs)
	assert.Equal(t, initialUpdate, collection.UpdatedAt) // Time should not change

	// Remove first track
	time.Sleep(1 * time.Millisecond)
	removed = collection.RemoveTrack(track1)
	assert.True(t, removed)
	assert.Equal(t, []domain.TrackID{track3}, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))
	initialUpdate = collection.UpdatedAt

	// Remove last track
	time.Sleep(1 * time.Millisecond)
	removed = collection.RemoveTrack(track3)
	assert.True(t, removed)
	assert.Empty(t, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))
}

func TestAudioCollection_ReorderTracks(t *testing.T) {
	collection, _ := domain.NewAudioCollection("Test", "", domain.NewUserID(), domain.TypePlaylist)
	track1 := domain.NewTrackID()
	track2 := domain.NewTrackID()
	track3 := domain.NewTrackID()
	collection.TrackIDs = []domain.TrackID{track1, track2, track3}
	initialUpdate := collection.UpdatedAt

	// Valid reorder
	time.Sleep(1 * time.Millisecond)
	newOrder := []domain.TrackID{track3, track1, track2}
	err := collection.ReorderTracks(newOrder)
	require.NoError(t, err)
	assert.Equal(t, newOrder, collection.TrackIDs)
	assert.True(t, collection.UpdatedAt.After(initialUpdate))

	// Invalid - incorrect number of tracks
	err = collection.ReorderTracks([]domain.TrackID{track1, track2})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
	assert.Equal(t, newOrder, collection.TrackIDs) // Order unchanged

	// Invalid - contains duplicate track
	err = collection.ReorderTracks([]domain.TrackID{track3, track1, track1})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
	assert.Equal(t, newOrder, collection.TrackIDs)

	// Invalid - contains track not in original collection
	err = collection.ReorderTracks([]domain.TrackID{track3, track1, domain.NewTrackID()})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
	assert.Equal(t, newOrder, collection.TrackIDs)
}