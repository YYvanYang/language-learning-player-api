// ==========================================================
// FILE: internal/adapter/handler/http/auth_handler.go (MODIFIED)
// ==========================================================
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/httputil"
	"github.com/yvanyang/language-learning-player-api/pkg/validation"
)

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authUseCase port.AuthUseCase
	validator   *validation.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc port.AuthUseCase, v *validation.Validator) *AuthHandler {
	return &AuthHandler{
		authUseCase: uc,
		validator:   v,
	}
}

// mapAuthResultToDTO maps the use case AuthResult to the response DTO.
func mapAuthResultToDTO(result port.AuthResult) dto.AuthResponseDTO {
	resp := dto.AuthResponseDTO{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	if result.IsNewUser { // Only include if true for external auth
		isNewPtr := true
		resp.IsNewUser = &isNewPtr
	}
	// **ADDED**: Map the user if present in the AuthResult
	if result.User != nil {
		// Use the existing DTO mapping function
		userDto := dto.MapDomainUserToResponseDTO(result.User)
		resp.User = &userDto // Assign the mapped DTO
	}
	return resp
}

// Register handles user registration requests.
// @Summary Register a new user
// @Description Registers a new user account using email and password.
// @ID register-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequestDTO true "User Registration Info"
// @Success 201 {object} dto.AuthResponseDTO "Registration successful, returns user details, access token, and refresh token."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email Exists"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	_, authResult, err := h.authUseCase.RegisterWithPassword(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusCreated, resp)
}

// Login handles user login requests.
// @Summary Login a user
// @Description Authenticates a user with email and password, returns user details, access token, and refresh token.
// @ID login-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequestDTO true "User Login Credentials"
// @Success 200 {object} dto.AuthResponseDTO "Login successful, returns user details, access token, and refresh token."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	authResult, err := h.authUseCase.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// GoogleCallback handles the callback from Google OAuth flow.
// @Summary Handle Google OAuth callback
// @Description Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning user details, access token, and refresh token.
// @ID google-callback
// @Tags Authentication
// @Accept json
// @Produce json
// @Param googleCallback body dto.GoogleCallbackRequestDTO true "Google ID Token"
// @Success 200 {object} dto.AuthResponseDTO "Authentication successful, returns user details, access/refresh tokens. isNewUser indicates new account creation."
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (Missing or Invalid ID Token)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed (Invalid Google Token)"
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email already exists with a different login method"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/google/callback [post]
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req dto.GoogleCallbackRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// MODIFIED: Use case returns AuthResult
	authResult, err := h.authUseCase.AuthenticateWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		httputil.RespondError(w, r, err)
		return
	}

	// MODIFIED: Map AuthResult to DTO
	resp := mapAuthResultToDTO(authResult)
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Refresh handles token refresh requests.
// @Summary Refresh access token
// @Description Provides a valid refresh token to get a new pair of access and refresh tokens.
// @ID refresh-token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshRequestDTO true "Refresh Token"
// @Success 200 {object} dto.AuthResponseDTO "Tokens refreshed successfully"
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input (Missing Refresh Token)"
// @Failure 401 {object} httputil.ErrorResponseDTO "Authentication Failed (Invalid or Expired Refresh Token)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/refresh [post]
// ADDED METHOD
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	authResult, err := h.authUseCase.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Use case should return domain.ErrAuthenticationFailed for invalid/expired tokens
		httputil.RespondError(w, r, err)
		return
	}

	resp := mapAuthResultToDTO(authResult) // Reuse mapping
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// Logout handles user logout requests by invalidating the refresh token.
// @Summary Logout user
// @Description Invalidates the current user's session and refresh tokens on the backend.
// @ID logout-user
// @Tags Authentication
// @Produce json
// @Security BearerAuth // Requires a valid access token
// @Success 204 "Logout successful"
// @Failure 401 {object} httputil.ErrorResponseDTO "Unauthorized (No valid access token)"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/logout [post] // Keep POST for semantic logout action
// MODIFIED: No request body needed
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get UserID from the context (validated by Authenticator middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		// This shouldn't happen if Authenticator runs first, but handle defensively
		httputil.RespondError(w, r, domain.ErrUnauthenticated)
		return
	}

	// Call use case with UserID
	err := h.authUseCase.Logout(r.Context(), userID)
	if err != nil {
		// Logout use case should generally return nil even if DB cleanup fails,
		// but handle potential configuration errors etc.
		httputil.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful logout initiation
}
