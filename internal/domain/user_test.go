package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNewLocalUser(t *testing.T) {
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	validHash := string(hashedPwd)

	tests := []struct {
		name           string
		email          string
		userName       string
		hashedPassword string
		wantErr        bool
		errType        error // Expected error type (e.g., ErrInvalidArgument)
	}{
		{"Valid local user", "local@example.com", "Local User", validHash, false, nil},
		{"Invalid email", "local@", "Local User", validHash, true, ErrInvalidArgument},
		{"Empty name", "local@example.com", "", validHash, false, nil}, // Name can be empty
		{"Empty hash", "local@example.com", "Local User", "", true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewLocalUser(tt.email, tt.userName, tt.hashedPassword)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, UserID{}, got.ID)
				assert.Equal(t, tt.email, got.Email.String())
				assert.Equal(t, tt.userName, got.Name)
				assert.NotNil(t, got.HashedPassword)
				assert.Equal(t, tt.hashedPassword, *got.HashedPassword)
				assert.Nil(t, got.GoogleID)
				assert.Equal(t, AuthProviderLocal, got.AuthProvider)
				assert.Nil(t, got.ProfileImageURL)
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestNewGoogleUser(t *testing.T) {
	profileURL := "http://example.com/pic.jpg"

	tests := []struct {
		name       string
		email      string
		userName   string
		googleID   string
		profileURL *string
		wantErr    bool
		errType    error
	}{
		{"Valid google user", "google@example.com", "Google User", "google123", &profileURL, false, nil},
		{"Valid google user no pic", "google2@example.com", "Google User 2", "google456", nil, false, nil},
		{"Invalid email", "google@", "Google User", "google123", nil, true, ErrInvalidArgument},
		{"Empty name", "google@example.com", "", "google123", nil, false, nil}, // Name can be empty
		{"Empty google ID", "google@example.com", "Google User", "", nil, true, ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := NewGoogleUser(tt.email, tt.userName, tt.googleID, tt.profileURL)
			end := time.Now()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.NotEqual(t, UserID{}, got.ID)
				assert.Equal(t, tt.email, got.Email.String())
				assert.Equal(t, tt.userName, got.Name)
				assert.Nil(t, got.HashedPassword)
				assert.NotNil(t, got.GoogleID)
				assert.Equal(t, tt.googleID, *got.GoogleID)
				assert.Equal(t, AuthProviderGoogle, got.AuthProvider)
				if tt.profileURL != nil {
					assert.NotNil(t, got.ProfileImageURL)
					assert.Equal(t, *tt.profileURL, *got.ProfileImageURL)
				} else {
					assert.Nil(t, got.ProfileImageURL)
				}
				assert.WithinDuration(t, start, got.CreatedAt, end.Sub(start)+time.Millisecond)
				assert.WithinDuration(t, start, got.UpdatedAt, end.Sub(start)+time.Millisecond)
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	plainPassword := "password123"
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	validHash := string(hashedPwd)

	localUser, _ := NewLocalUser("local@example.com", "Test", validHash)
	googleUser, _ := NewGoogleUser("google@example.com", "Test", "gid1", nil)
	localUserNoHash, _ := NewLocalUser("localnohash@example.com", "Test", "dummy")
	localUserNoHash.HashedPassword = nil // Simulate case where hash is somehow nil

	tests := []struct {
		name        string
		user        *User
		password    string
		wantMatch   bool
		wantErr     bool
		errContains string
	}{
		{"Correct password local user", localUser, plainPassword, true, false, ""},
		{"Incorrect password local user", localUser, "wrongpassword", false, false, ""},
		{"Google user", googleUser, plainPassword, false, true, "password validation not applicable"},
		{"Local user with nil hash", localUserNoHash, plainPassword, false, true, "password validation not applicable"},
		{"Invalid hash format check", localUser, plainPassword, false, true, "error comparing password hash"}, // Need to provide an invalid hash to trigger this
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.user
			password := tt.password
			hash := ""
			if user.HashedPassword != nil {
				hash = *user.HashedPassword
			}

			// Special case for invalid hash format test
			if tt.name == "Invalid hash format check" {
				hash = "this-is-not-a-bcrypt-hash"
				// Recreate user with this invalid hash temporarily for the test
				user = &User{
					ID: localUser.ID, Email: localUser.Email, Name: localUser.Name, AuthProvider: localUser.AuthProvider,
					HashedPassword: &hash, CreatedAt: localUser.CreatedAt, UpdatedAt: localUser.UpdatedAt,
				}
			}

			gotMatch, err := user.ValidatePassword(password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMatch, gotMatch)
			}
		})
	}
}

// Add tests for UpdateProfile, LinkGoogleID if needed
