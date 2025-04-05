// internal/domain/value_objects.go
package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

// --- Basic Value Objects ---

// Language represents a language with its code and name. Immutable.
type Language struct {
	code string // e.g., "en-US"
	name string // e.g., "English (US)"
}

func NewLanguage(code, name string) (Language, error) {
	if code == "" {
		return Language{}, fmt.Errorf("%w: language code cannot be empty", ErrInvalidArgument)
	}
	// Add more validation if needed (e.g., format check)
	return Language{code: strings.ToUpper(code), name: name}, nil
}
func (l Language) Code() string { return l.code }
func (l Language) Name() string { return l.name }

// AudioLevel represents the difficulty level of an audio track. Immutable.
type AudioLevel string // e.g., "A1", "B2", "NATIVE"

const (
	LevelA1 AudioLevel = "A1"
	LevelA2 AudioLevel = "A2"
	LevelB1 AudioLevel = "B1"
	LevelB2 AudioLevel = "B2"
	LevelC1 AudioLevel = "C1"
	LevelC2 AudioLevel = "C2"
	LevelNative AudioLevel = "NATIVE"
	LevelUnknown AudioLevel = "" // Or "UNKNOWN"
)

func (l AudioLevel) IsValid() bool {
	switch l {
	case LevelA1, LevelA2, LevelB1, LevelB2, LevelC1, LevelC2, LevelNative, LevelUnknown:
		return true
	default:
		return false
	}
}
func (l AudioLevel) String() string { return string(l) }

// CollectionType represents the type of an audio collection. Immutable.
type CollectionType string // e.g., "COURSE", "PLAYLIST"

const (
	TypeCourse   CollectionType = "COURSE"
	TypePlaylist CollectionType = "PLAYLIST"
	TypeUnknown  CollectionType = "" // Or "UNKNOWN"
)

func (t CollectionType) IsValid() bool {
	switch t {
	case TypeCourse, TypePlaylist, TypeUnknown:
		return true
	default:
		return false
	}
}
func (t CollectionType) String() string { return string(t) }

// --- Email Value Object (Example with Validation) ---

// Email represents a validated email address. Immutable.
type Email struct {
	address string
}

func NewEmail(address string) (Email, error) {
	if address == "" {
		return Email{}, fmt.Errorf("%w: email address cannot be empty", ErrInvalidArgument)
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil {
		return Email{}, fmt.Errorf("%w: invalid email address format: %v", ErrInvalidArgument, err)
	}
	// Potentially add more domain-specific validation (e.g., allowed domains)
	return Email{address: parsed.Address}, nil // Use the parsed address for canonical form
}

func (e Email) String() string { return e.address }

// --- Time-related Value Objects (Go's time types often suffice) ---
// time.Duration can represent playback progress or track duration.
// time.Time can represent timestamps like CreatedAt, UpdatedAt, LastListenedAt.