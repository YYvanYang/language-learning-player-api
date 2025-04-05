// internal/adapter/handler/http/user_handler.go
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
)

// UserUseCase defines the interface for user-related operations needed by the handler.
type UserUseCase interface {
	GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error)
}

// UserHandler handles HTTP requests related to user profiles.
type UserHandler struct {
	userUseCase UserUseCase
	// validator *validation.Validator // Add validator if needed for PUT/PATCH later
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: uc,
	}
}

// GetMyProfile handles GET /api/v1/users/me
// @Summary Get current user's profile
// @Description Retrieves the profile information for the currently authenticated user.
// @ID get-my-profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponseDTO "User profile retrieved successfully"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized"
// @Failure 404 {object} httputil.ErrorResponseDTO "User Not Found"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /users/me [get]
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	user, err := h.userUseCase.GetUserProfile(r.Context(), userID)
	if err != nil {
		// Handles domain.ErrNotFound, etc.
		httputil.RespondError(w, r, err)
		return
	}

	resp := MapDomainUserToResponseDTO(user) // Use mapping function
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// MapDomainUserToResponseDTO converts a domain user to its DTO representation.
// Consider moving this to the dto package or keeping it here if only used by this handler.
func MapDomainUserToResponseDTO(user *domain.User) dto.UserResponseDTO {
	return dto.UserResponseDTO{
		ID:              user.ID.String(),
		Email:           user.Email.String(),
		Name:            user.Name,
		AuthProvider:    string(user.AuthProvider),
		ProfileImageURL: user.ProfileImageURL,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339), // Format time
		UpdatedAt:       user.UpdatedAt.Format(time.RFC3339),
	}
} 