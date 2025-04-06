// =============================================
// FILE: internal/domain/value_objects_test.go
// =============================================
// This file was already provided in all_code.md and looks correct.
// No changes needed.
package domain_test // Use _test package to test only exported identifiers

import (
	"testing"
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Import the actual package

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedVal domain.Email
		expectError bool
		errorType   error // Optional: Check for specific error type
	}{
		{
			name:        "Valid Email",
			input:       "test@example.com",
			expectedVal: domain.Email{}, // We'll check string representation
			expectError: false,
		},
		{
			name:        "Valid Email with Subdomain",
			input:       "test.sub@example.co.uk",
			expectedVal: domain.Email{},
			expectError: false,
		},
		{
			name:        "Valid Email with Display Name (ignored)",
			input:       "Test User <test@example.com>",
			expectedVal: domain.Email{}, // Should parse out just the address
			expectError: false,
		},
		{
			name:        "Invalid Email - Missing @",
			input:       "testexample.com",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
		{
			name:        "Invalid Email - Missing Domain",
			input:       "test@",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
		{
			name:        "Empty Email",
			input:       "",
			expectError: true,
			errorType:   domain.ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			email, err := domain.NewEmail(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error but got none")
				if tc.errorType != nil {
					assert.ErrorIs(t, err, tc.errorType, "Expected error type mismatch")
				}
			} else {
				require.NoError(t, err, "Expected no error but got one: %v", err)
				// Check the string representation
				expectedEmailStr := tc.input
				if tc.name == "Valid Email with Display Name (ignored)" {
					expectedEmailStr = "test@example.com" // Check the parsed value
				}
				assert.Equal(t, expectedEmailStr, email.String(), "Email string representation mismatch")
			}
		})
	}
}

func TestAudioLevel_IsValid(t *testing.T) {
	assert.True(t, domain.LevelA1.IsValid())
	assert.True(t, domain.LevelB2.IsValid())
	assert.True(t, domain.LevelNative.IsValid())
	assert.True(t, domain.LevelUnknown.IsValid())
	assert.False(t, domain.AudioLevel("D1").IsValid())
	assert.False(t, domain.AudioLevel("").IsValid()) // Should use LevelUnknown
}

func TestCollectionType_IsValid(t *testing.T) {
    assert.True(t, domain.TypeCourse.IsValid())
    assert.True(t, domain.TypePlaylist.IsValid())
    assert.True(t, domain.TypeUnknown.IsValid())
    assert.False(t, domain.CollectionType("BOOK").IsValid())
    assert.False(t, domain.CollectionType("").IsValid())
}

func TestNewLanguage(t *testing.T) {
    lang, err := domain.NewLanguage("en-US", "English (US)")
    require.NoError(t, err)
    assert.Equal(t, "EN-US", lang.Code()) // Should be uppercase
    assert.Equal(t, "English (US)", lang.Name())

    _, err = domain.NewLanguage("", "Spanish")
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}