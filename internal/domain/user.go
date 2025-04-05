// internal/domain/user.go
package domain

import (
	"time"
	"fmt"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserID is the unique identifier for a User.
type UserID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func UserIDFromString(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid UserID format: %w", err)
	}
	return UserID(id), nil
}

func (uid UserID) String() string {
	return uuid.UUID(uid).String()
}

// AuthProvider represents the method used for authentication.
type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle AuthProvider = "google"
	// Add other providers like Facebook, Apple etc. here
)

// User represents a user in the system.
type User struct {
	ID              UserID
	Email           Email // Using validated Email value object
	Name            string
	HashedPassword  *string // Pointer allows null for external auth users
	GoogleID        *string // Unique ID from Google (subject claim)
	AuthProvider    AuthProvider
	ProfileImageURL *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewLocalUser creates a new user who registered with email and password.
func NewLocalUser(emailAddr, name, hashedPassword string) (*User, error) {
	emailVO, err := NewEmail(emailAddr)
	if err != nil {
		return nil, err
	}
	if hashedPassword == "" {
		return nil, fmt.Errorf("%w: password hash cannot be empty for local user", ErrInvalidArgument)
	}

	now := time.Now()
	return &User{
		ID:             NewUserID(),
		Email:          emailVO,
		Name:           name,
		HashedPassword: &hashedPassword, // Store the already hashed password
		GoogleID:       nil,
		AuthProvider:   AuthProviderLocal,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// NewGoogleUser creates a new user authenticated via Google.
func NewGoogleUser(emailAddr, name, googleID string, profileImageURL *string) (*User, error) {
	emailVO, err := NewEmail(emailAddr)
	if err != nil {
		return nil, err
	}
	if googleID == "" {
		return nil, fmt.Errorf("%w: google ID cannot be empty for google user", ErrInvalidArgument)
	}

	now := time.Now()
	return &User{
		ID:              NewUserID(),
		Email:           emailVO,
		Name:            name,
		HashedPassword:  nil, // No password for Google users initially
		GoogleID:        &googleID,
		AuthProvider:    AuthProviderGoogle,
		ProfileImageURL: profileImageURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// ValidatePassword checks if the provided plain password matches the user's hashed password.
// Returns true if the password matches or if the user is not a local user (no password set).
// Returns false if the password does not match or an error occurred during comparison.
func (u *User) ValidatePassword(plainPassword string) (bool, error) {
	if u.AuthProvider != AuthProviderLocal || u.HashedPassword == nil {
		// Cannot validate password for non-local users or if hash is missing
		// Consider returning an error or specific bool value based on desired behavior.
		// Returning false might be misleading if the intent is "password auth not applicable".
		// Let's return an error for clarity.
		return false, fmt.Errorf("password validation not applicable for provider %s", u.AuthProvider)
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.HashedPassword), []byte(plainPassword))
	if err == nil {
		return true, nil // Passwords match
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil // Passwords don't match
	}
	// Some other error occurred (e.g., invalid hash format)
	return false, fmt.Errorf("error comparing password hash: %w", err)
}

// UpdateProfile updates mutable profile fields.
func (u *User) UpdateProfile(name string, profileImageURL *string) {
	u.Name = name
	u.ProfileImageURL = profileImageURL
	u.UpdatedAt = time.Now()
}

// LinkGoogleID links a Google account ID to an existing local user.
// Use this cautiously, ensure proper verification beforehand in usecase.
func (u *User) LinkGoogleID(googleID string) error {
	if u.GoogleID != nil && *u.GoogleID != "" {
		return fmt.Errorf("%w: user already linked to a google account", ErrConflict)
	}
	if googleID == "" {
		return fmt.Errorf("%w: google ID cannot be empty", ErrInvalidArgument)
	}
	u.GoogleID = &googleID
	// Potentially change auth provider or keep it? Depends on login flow.
	// If user can now login via EITHER method, maybe keep 'local' or add a list?
	// For simplicity, let's assume linking doesn't change the primary AuthProvider if it was 'local'.
	// If the user was created via Google first, AuthProvider would already be 'google'.
	u.UpdatedAt = time.Now()
	return nil
}