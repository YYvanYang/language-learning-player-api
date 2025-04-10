// internal/adapter/handler/http/dto/auth_dto.go
package dto

// --- Request DTOs ---

// RegisterRequestDTO defines the expected JSON body for user registration.
type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`                    // Add example tag
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

// RefreshRequestDTO defines the expected JSON body for token refresh.
// ADDED
type RefreshRequestDTO struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// LogoutRequestDTO defines the expected JSON body for logout.
// ADDED (Optional, can also just take token from body or header)
type LogoutRequestDTO struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// --- Response DTOs ---

// AuthResponseDTO defines the JSON response body for successful authentication/refresh.
// MODIFIED: Added refreshToken
type AuthResponseDTO struct {
	AccessToken  string `json:"accessToken"`         // The JWT access token
	RefreshToken string `json:"refreshToken"`        // The refresh token value
	IsNewUser    *bool  `json:"isNewUser,omitempty"` // Pointer, only included for Google callback if user is new
}
