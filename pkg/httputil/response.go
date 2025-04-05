// pkg/httputil/response.go
package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // For Request ID
)

// ErrorResponseDTO defines the standard JSON error response body.
type ErrorResponseDTO struct {
	Code      string `json:"code"`                // Application-specific error code (e.g., "INVALID_INPUT", "NOT_FOUND")
	Message   string `json:"message"`             // User-friendly error message
	RequestID string `json:"requestId,omitempty"` // Include request ID for tracing
	// Details   interface{} `json:"details,omitempty"` // Optional: For more detailed error info (e.g., validation fields)
}

// RespondJSON writes a JSON response with the given status code and payload.
func RespondJSON(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			// Log error, but can't write header again
			logger := slog.Default()
			reqID := middleware.GetReqID(r.Context()) // Get request ID from context
			logger.ErrorContext(r.Context(), "Failed to encode JSON response", "error", err, "status", status, "request_id", reqID)
			// Attempt to write a plain text error if JSON encoding fails *after* header was written
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// MapDomainErrorToHTTP maps domain errors to HTTP status codes and error codes.
// This is a central place to define the mapping.
func MapDomainErrorToHTTP(err error) (status int, code string, message string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "The requested resource was not found."
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "RESOURCE_CONFLICT", "A conflict occurred with the current state of the resource." // Generic message
		// Consider more specific messages based on context if err wraps more info
		// if strings.Contains(err.Error(), "email") { message = "Email already exists."} ...
	case errors.Is(err, domain.ErrInvalidArgument):
		return http.StatusBadRequest, "INVALID_INPUT", err.Error() // Use error message directly for validation details
	case errors.Is(err, domain.ErrPermissionDenied):
		return http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action."
	case errors.Is(err, domain.ErrAuthenticationFailed):
		return http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication failed. Please check your credentials." // Use 401 for auth failure
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required. Please log in." // Also 401
	default:
		// Any other error is treated as an internal server error
		return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected internal error occurred."
	}
}

// RespondError maps a domain error to an HTTP status code and JSON error response.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := MapDomainErrorToHTTP(err)

	reqID := middleware.GetReqID(r.Context()) // Get request ID from context
	logger := slog.Default()

	// Log internal server errors with more detail
	if status >= 500 {
		logger.ErrorContext(r.Context(), "Internal server error occurred", "error", err, "status", status, "code", code, "request_id", reqID)
		// Avoid leaking internal error details in the response message for 500 errors
		message = "An unexpected internal error occurred."
	} else {
		// Log client errors (4xx) at a lower level, e.g., Warn or Info
		logger.WarnContext(r.Context(), "Client error occurred", "error", err, "status", status, "code", code, "request_id", reqID)
	}

	errorResponse := ErrorResponseDTO{
		Code:      code,
		Message:   message,
		RequestID: reqID,
	}

	RespondJSON(w, r, status, errorResponse)
}