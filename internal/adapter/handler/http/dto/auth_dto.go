// internal/adapter/handler/http/dto/auth_dto.go
package dto

// --- Request DTOs ---

// RegisterRequestDTO defines the expected JSON body for user registration.
type RegisterRequestDTO struct {
    Email    string `json:"email" validate:"required,email" example:"user@example.com"` // Add example tag
    Password string `json:"password" validate:"required,min=8" format:"password" example:"Str0ngP@ssw0rd"` // Add format tag
    Name     string `json:"name" validate:"required,max=100" example:"John Doe"`
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

// REMOVED UserResponseDTO from here. It now resides in user_dto.go