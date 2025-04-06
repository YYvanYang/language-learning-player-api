// =============================================
// FILE: internal/domain/audiotrack_test.go
// =============================================
package domain_test

import (
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAudioTrack_Success(t *testing.T) {
	title := "Test Track"
	desc := "A description"
	bucket := "test-bucket"
	key := "test/key.mp3"
	lang, _ := domain.NewLanguage("FR", "French")
	level := domain.LevelB1
	duration := 180 * time.Second
	isPublic := true
	tags := []string{"podcast", "intermediate"}

	track, err := domain.NewAudioTrack(title, desc, bucket, key, lang, level, duration, nil, isPublic, tags, nil)

	require.NoError(t, err)
	require.NotNil(t, track)
	assert.NotEqual(t, domain.TrackID{}, track.ID)
	assert.Equal(t, title, track.Title)
	assert.Equal(t, desc, track.Description)
	assert.Equal(t, bucket, track.MinioBucket)
	assert.Equal(t, key, track.MinioObjectKey)
	assert.Equal(t, lang, track.Language)
	assert.Equal(t, level, track.Level)
	assert.Equal(t, duration, track.Duration)
	assert.Nil(t, track.UploaderID)
	assert.True(t, track.IsPublic)
	assert.Equal(t, tags, track.Tags)
	assert.Nil(t, track.CoverImageURL)
	assert.WithinDuration(t, time.Now(), track.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), track.UpdatedAt, time.Second)
}

func TestNewAudioTrack_ValidationErrors(t *testing.T) {
	lang, _ := domain.NewLanguage("DE", "")

	testCases := []struct {
		name        string
		title       string
		bucket      string
		key         string
		level       domain.AudioLevel
		duration    time.Duration
		expectError bool
		errorType   error
	}{
		{"Empty Title", "", "b", "k", domain.LevelA1, 1, true, domain.ErrInvalidArgument},
		{"Empty Bucket", "t", "", "k", domain.LevelA1, 1, true, domain.ErrInvalidArgument},
		{"Empty Key", "t", "b", "", domain.LevelA1, 1, true, domain.ErrInvalidArgument},
		{"Negative Duration", "t", "b", "k", domain.LevelA1, -1, true, domain.ErrInvalidArgument},
		{"Invalid Level", "t", "b", "k", domain.AudioLevel("XYZ"), 1, true, domain.ErrInvalidArgument},
		{"Valid Unknown Level", "t", "b", "k", domain.LevelUnknown, 1, false, nil}, // Unknown level is allowed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewAudioTrack(tc.title, "", tc.bucket, tc.key, lang, tc.level, tc.duration, nil, true, nil, nil)
			if tc.expectError {
				require.Error(t, err)
				if tc.errorType != nil {
					assert.ErrorIs(t, err, tc.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}