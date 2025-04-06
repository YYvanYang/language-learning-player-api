// internal/domain/audiotrack.go
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrackID is the unique identifier for an AudioTrack.
type TrackID uuid.UUID

func NewTrackID() TrackID {
	return TrackID(uuid.New())
}

func TrackIDFromString(s string) (TrackID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TrackID{}, fmt.Errorf("invalid TrackID format: %w", err)
	}
	return TrackID(id), nil
}

func (tid TrackID) String() string {
	return uuid.UUID(tid).String()
}

// AudioTrack represents a single audio file with metadata.
type AudioTrack struct {
	ID              TrackID
	Title           string
	Description     string
	Language        Language     // CORRECTED TYPE: Use Language value object
	Level           AudioLevel   // CORRECTED TYPE: Use AudioLevel value object
	Duration        time.Duration // Store as duration for easier use
	MinioBucket     string
	MinioObjectKey  string
	CoverImageURL   *string
	UploaderID      *UserID // Optional link to the user who uploaded it
	IsPublic        bool
	Tags            []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	// TranscriptionID *TranscriptionID // Optional if transcriptions are separate entities
}

// NewAudioTrack creates a new audio track instance.
func NewAudioTrack(
	title, description, bucket, objectKey string,
	lang Language, // CORRECTED TYPE: Accept Language VO
	level AudioLevel, // CORRECTED TYPE: Accept AudioLevel VO
	duration time.Duration,
	uploaderID *UserID, isPublic bool, tags []string, coverURL *string,
) (*AudioTrack, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: track title cannot be empty", ErrInvalidArgument)
	}
	if bucket == "" || objectKey == "" {
		return nil, fmt.Errorf("%w: minio bucket and key cannot be empty", ErrInvalidArgument)
	}
	if duration < 0 {
		return nil, fmt.Errorf("%w: duration cannot be negative", ErrInvalidArgument)
	}
	// Allow LevelUnknown, but validate others
	if level != LevelUnknown && !level.IsValid() {
		return nil, fmt.Errorf("%w: invalid audio level '%s'", ErrInvalidArgument, level)
	}
	// Language VO creation handles its own validation if called externally before this

	now := time.Now()
	return &AudioTrack{
		ID:             NewTrackID(),
		Title:          title,
		Description:    description,
		Language:       lang, // Assign VO directly
		Level:          level, // Assign VO directly
		Duration:       duration,
		MinioBucket:    bucket,
		MinioObjectKey: objectKey,
		UploaderID:     uploaderID,
		IsPublic:       isPublic,
		Tags:           tags,
		CoverImageURL:  coverURL,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}