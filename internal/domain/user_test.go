// internal/domain/user_test.go
package domain_test

import (
	"testing"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewLocalUser(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	user, err := domain.NewLocalUser(email, name, string(hashedPassword))

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEqual(t, domain.UserID{}, user.ID) // Check ID is generated
	assert.Equal(t, email, user.Email.String())
	assert.Equal(t, name, user.Name)
	require.NotNil(t, user.HashedPassword)
	assert.Equal(t, string(hashedPassword), *user.HashedPassword)
	assert.Nil(t, user.GoogleID)
	assert.Equal(t, domain.AuthProviderLocal, user.AuthProvider)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestNewLocalUser_InvalidEmail(t *testing.T) {
	_, err := domain.NewLocalUser("invalid-email", "Test", "hash")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func TestNewLocalUser_EmptyHash(t *testing.T) {
	_, err := domain.NewLocalUser("test@example.com", "Test", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidArgument)
}


func TestUser_ValidatePassword(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	wrongPassword := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	// Create a local user
	localUser, err := domain.NewLocalUser(email, name, string(hashedPassword))
	require.NoError(t, err)

	// Create a Google user
	googleUser, err := domain.NewGoogleUser(email, name, "google123", nil)
	require.NoError(t, err)


	// Test cases
	match, err := localUser.ValidatePassword(plainPassword)
	assert.NoError(t, err)
	assert.True(t, match, "Correct password should match")

	match, err = localUser.ValidatePassword(wrongPassword)
	assert.NoError(t, err)
	assert.False(t, match, "Incorrect password should not match")

	match, err = localUser.ValidatePassword("")
	assert.NoError(t, err)
	assert.False(t, match, "Empty password should not match")

	// Test on Google user (should fail)
	match, err = googleUser.ValidatePassword(plainPassword)
	assert.Error(t, err, "Should return error for non-local user")
	assert.False(t, match, "Match should be false on error")

    // Test with nil hash (should not happen with NewLocalUser, but test defensively)
    localUser.HashedPassword = nil
    match, err = localUser.ValidatePassword(plainPassword)
	assert.Error(t, err, "Should return error if hash is nil")
	assert.False(t, match, "Match should be false on error")

}

func TestUser_LinkGoogleID(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	plainPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    googleID := "google12345"
    googleID2 := "google67890"

	localUser, err := domain.NewLocalUser(email, name, string(hashedPassword))
	require.NoError(t, err)

    // Initial state: No Google ID
    assert.Nil(t, localUser.GoogleID)

    // Link valid ID
    err = localUser.LinkGoogleID(googleID)
    assert.NoError(t, err)
    require.NotNil(t, localUser.GoogleID)
    assert.Equal(t, googleID, *localUser.GoogleID)
    // AuthProvider should remain local (as per current implementation)
    assert.Equal(t, domain.AuthProviderLocal, localUser.AuthProvider)

    // Try linking again (should fail)
    err = localUser.LinkGoogleID(googleID2)
    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict)
    assert.Equal(t, googleID, *localUser.GoogleID) // ID should not change

    // Try linking empty ID (should fail)
    localUser.GoogleID = nil // Reset for test
    err = localUser.LinkGoogleID("")
    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidArgument)
    assert.Nil(t, localUser.GoogleID)

}

// TODO: Add tests for other entities like AudioCollection (AddTrack, RemoveTrack, ReorderTracks)