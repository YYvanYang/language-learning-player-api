// internal/adapter/handler/http/auth_handler.go
package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" 
	"github.com/yvanyang/language-learning-player-backend/internal/port"   
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto" 
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"    
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"  
)

// AuthHandler handles HTTP requests related to authentication.
type AuthHandler struct {
	authUseCase port.AuthUseCase // Use interface defined in port package
	validator   *validation.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc port.AuthUseCase, v *validation.Validator) *AuthHandler {
	return &AuthHandler{
		authUseCase: uc,
		validator:   v,
	}
}

// Register handles user registration requests.
// @Summary Register a new user
// @Description Registers a new user account using email and password.
// @ID register-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequestDTO true "User Registration Info" // Input parameter: name, in, type, required, description
// @Success 201 {object} dto.AuthResponseDTO "Registration successful, returns JWT" // Success response: code, type, description
// @Failure 400 {object} httputil.ErrorResponseDTO "Invalid Input"                // Failure response: code, type, description
// @Failure 409 {object} httputil.ErrorResponseDTO "Conflict - Email Exists"
// @Failure 500 {object} httputil.ErrorResponseDTO "Internal Server Error"
// @Router /auth/register [post]  // Route path and method
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, "invalid request body"))
		return
	}
	defer r.Body.Close()

	// Validate input DTO
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case
	_, token, err := h.authUseCase.RegisterWithPassword(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		// UseCase should return domain errors, let RespondError map them
		httputil.RespondError(w, r, err)
		return
	}

	// Return JWT token
	resp := dto.AuthResponseDTO{Token: token}
	httputil.RespondJSON(w, r, http.StatusCreated, resp) // 201 Created for successful registration
}

// Login handles user login requests.
// @Summary Login a user
// @Description Authenticates a user with email and password, returns a JWT token.
// @ID login-user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequestDTO true "User Login Credentials"
// @Success 200 {object} dto.AuthResponseDTO "Login successful, returns JWT"
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

	// Validate input DTO
	if err := h.validator.ValidateStruct(req); err != nil {
		httputil.RespondError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
		return
	}

	// Call use case
	token, err := h.authUseCase.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		// Handles domain.ErrAuthenticationFailed, domain.ErrNotFound (mapped to auth failed), etc.
		httputil.RespondError(w, r, err)
		return
	}

	// Return JWT token
	resp := dto.AuthResponseDTO{Token: token}
	httputil.RespondJSON(w, r, http.StatusOK, resp)
}

// GoogleCallback handles the callback from Google OAuth flow.
// @Summary Handle Google OAuth callback
// @Description Receives the ID token from the frontend after Google sign-in, verifies it, and performs user registration or login, returning a JWT.
// @ID google-callback
// @Tags Authentication
// @Accept json
// @Produce json
// @Param googleCallback body dto.GoogleCallbackRequestDTO true "Google ID Token"
// @Success 200 {object} dto.AuthResponseDTO "Authentication successful, returns JWT. isNewUser field indicates if a new account was created."
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

	token, isNew, err := h.authUseCase.AuthenticateWithGoogle(r.Context(), req.IDToken)
	if err != nil {
		// Handles domain.ErrAuthenticationFailed, domain.ErrConflict, etc.
		httputil.RespondError(w, r, err)
		return
	}

	resp := dto.AuthResponseDTO{Token: token}
	// Only include isNewUser field if it's true, otherwise omit it
	if isNew {
		isNewPtr := true
		resp.IsNewUser = &isNewPtr
	}

	httputil.RespondJSON(w, r, http.StatusOK, resp)
}