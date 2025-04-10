package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLanguage(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		langName string
		wantCode string
		wantErr  bool
	}{
		{"Valid language", "en-US", "English (US)", "EN-US", false},
		{"Valid language lowercase", "fr-ca", "French (Canada)", "FR-CA", false},
		{"Empty code", "", "No Language", "", true},
		{"Only name", "en", "", "EN", false}, // Allow empty name
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLanguage(tt.code, tt.langName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidArgument) // Check specific error type
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, got.Code())
				assert.Equal(t, tt.langName, got.Name())
			}
		})
	}
}

func TestAudioLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level AudioLevel
		want  bool
	}{
		{"Valid A1", LevelA1, true},
		{"Valid B2", LevelB2, true},
		{"Valid Native", LevelNative, true},
		{"Valid Unknown", LevelUnknown, true},
		{"Invalid Lowercase", AudioLevel("a1"), false},
		{"Invalid Other", AudioLevel("XYZ"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.IsValid())
			assert.Equal(t, string(tt.level), tt.level.String()) // Test String() method
		})
	}
}

func TestCollectionType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		typ  CollectionType
		want bool
	}{
		{"Valid Course", TypeCourse, true},
		{"Valid Playlist", TypePlaylist, true},
		{"Valid Unknown", TypeUnknown, true},
		{"Invalid Lowercase", CollectionType("course"), false},
		{"Invalid Other", CollectionType("XYZ"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.typ.IsValid())
			assert.Equal(t, string(tt.typ), tt.typ.String()) // Test String() method
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
		wantErr bool
	}{
		{"Valid email", "test@example.com", "test@example.com", false},
		{"Valid email with name", "Test User <test.user+alias@example.co.uk>", "test.user+alias@example.co.uk", false},
		{"Invalid format", "test@", "", true},
		{"Empty string", "", "", true},
		{"Only domain", "@example.com", "", true},
		{"Accept No TLD (mail.ParseAddress allows)", "test@example", "test@example", false}, // mail.ParseAddress requires valid TLD usually
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidArgument)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}
