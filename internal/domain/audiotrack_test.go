package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAudioTrack(t *testing.T) {
	langEn, _ := NewLanguage("en-US", "English (US)")
	langFr, _ := NewLanguage("fr-FR", "French (France)")
	uploaderID := NewUserID()
	coverURL := "http://example.com/cover.jpg"

	tests := []struct {
		name       string
		title      string
		desc       string
		bucket     string
		objectKey  string
		lang       Language
		level      AudioLevel
		duration   time.Duration
		uploaderID *UserID
		isPublic   bool
		tags       []string
		coverURL   *string
		wantErr    bool
		errType    error
	}{
		{"Valid public track", "Track 1", "Desc 1", "bucket", "key1", langEn, LevelB1, 120 * time.Second, &uploaderID, true, []string{"news", "politics"}, &coverURL, false, nil},
		{"Valid private track no uploader", "Track 2", "", "bucket", "key2", langFr, LevelA2, 10 * time.Millisecond, nil, false, nil, nil, false, nil},
		{"Empty title", "", "Desc", "bucket", "key3", langEn, LevelC1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Empty bucket", "Track 3", "Desc", "", "key4", langEn, LevelA1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Empty object key", "Track 4", "Desc", "bucket", "", langEn, LevelA1, time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Negative duration", "Track 5", "Desc", "bucket", "key5", langEn, LevelNative, -time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Invalid Level", "Track 6", "Desc", "bucket", "key6", langEn, AudioLevel("XYZ"), time.Second, nil, true, nil, nil, true, ErrInvalidArgument},
		{"Unknown Level", "Track 7", "Desc", "bucket", "key7", langEn, LevelUnknown, time.Second, nil, true, nil, nil, false, nil}, // Unknown level is allowed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewAudioTrack(tt.title, tt.desc, tt.bucket, tt.objectKey, tt.lang, tt.level, tt.duration, tt.uploaderID, tt.isPublic, tt.tags, tt.coverURL)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, TrackID{}, got.ID)
				assert.Equal(t, tt.title, got.Title)
				assert.Equal(t, tt.desc, got.Description)
				assert.Equal(t, tt.bucket, got.MinioBucket)
				assert.Equal(t, tt.objectKey, got.MinioObjectKey)
				assert.Equal(t, tt.lang, got.Language)
				assert.Equal(t, tt.level, got.Level)
				assert.Equal(t, tt.duration, got.Duration)
				assert.Equal(t, tt.uploaderID, got.UploaderID)
				assert.Equal(t, tt.isPublic, got.IsPublic)
				assert.Equal(t, tt.tags, got.Tags) // Check slice equality
				assert.Equal(t, tt.coverURL, got.CoverImageURL)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}
