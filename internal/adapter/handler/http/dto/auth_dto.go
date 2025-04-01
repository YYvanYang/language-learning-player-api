// internal/adapter/handler/http/dto/auth_dto.go
package dto

// --- Request DTOs ---

// RegisterRequestDTO defines the expected JSON body for user registration.
type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"` // Min length validation
	Name     string `json:"name" validate:"required,max=100"`
}

// LoginRequestDTO defines the expected JSON body for user login.
type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GoogleCallbackRequestDTO defines the expected JSON body for Google OAuth callback.
type GoogleCallbackRequestDTO struct {
	IDToken string `json:"idToken" validate:"required"`
}


// --- Response DTOs ---

// AuthResponseDTO defines the JSON response body for successful authentication.
type AuthResponseDTO struct {
	Token     string `json:"token"`                // The JWT access token
	IsNewUser *bool  `json:"isNewUser,omitempty"` // Pointer, only included for Google callback if user is new
}

// UserResponseDTO defines the JSON representation of a user profile (example).
type UserResponseDTO struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	Name            string  `json:"name"`
	AuthProvider    string  `json:"authProvider"`
	ProfileImageURL *string `json:"profileImageUrl,omitempty"`
	CreatedAt       string  `json:"createdAt"` // Use string format like RFC3339
	UpdatedAt       string  `json:"updatedAt"`
}