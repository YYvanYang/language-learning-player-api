// internal/adapter/handler/http/user_handler.go
package http

import (
	// REMOVED: "context" - Not needed directly here
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto" // Import dto package
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/internal/port" // Import port package for UserUseCase interface
)

// UserHandler handles HTTP requests related to user profiles.
type UserHandler struct {
	userUseCase port.UserUseCase // Use interface from port package
	// validator *validation.Validator // Add validator if needed for PUT/PATCH later
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc port.UserUseCase) *UserHandler {
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

	resp := dto.MapDomainUserToResponseDTO(user) // Use mapping function from DTO package
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}