// internal/adapter/handler/http/dto/user_dto.go
package dto

import (
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
)

// UserResponseDTO defines the JSON representation of a user profile.
type UserResponseDTO struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	Name            string  `json:"name"`
	AuthProvider    string  `json:"authProvider"`
	ProfileImageURL *string `json:"profileImageUrl,omitempty"`
	CreatedAt       string  `json:"createdAt"` // Use string format like RFC3339
	UpdatedAt       string  `json:"updatedAt"`
}

// MapDomainUserToResponseDTO converts a domain user to its DTO representation.
func MapDomainUserToResponseDTO(user *domain.User) UserResponseDTO {
	return UserResponseDTO{
		ID:              user.ID.String(),
		Email:           user.Email.String(),
		Name:            user.Name,
		AuthProvider:    string(user.AuthProvider),
		ProfileImageURL: user.ProfileImageURL,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339), // Format time
		UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
	}
}

// Note: Auth Request/Response DTOs remain in auth_dto.go