// ============================================
// FILE: pkg/httputil/response.go (MODIFIED)
// ============================================
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/apierrors"   // Import the new package
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const RequestIDKey ContextKey = "requestID"

// GetReqID retrieves the request ID from the context.
// Returns an empty string if not found.
func GetReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

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
			reqID := GetReqID(r.Context()) // Get request ID from context
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
		// Use constant from apierrors package
		return http.StatusNotFound, apierrors.CodeNotFound, "The requested resource was not found."
	case errors.Is(err, domain.ErrConflict):
		// Use constant from apierrors package
		return http.StatusConflict, apierrors.CodeConflict, "A conflict occurred with the current state of the resource."
	case errors.Is(err, domain.ErrInvalidArgument):
		// Use constant from apierrors package
		// Use the specific error message for validation details
		return http.StatusBadRequest, apierrors.CodeInvalidInput, err.Error()
	case errors.Is(err, domain.ErrPermissionDenied):
		// Use constant from apierrors package
		// Check for specific rate limit error message if needed, otherwise use generic forbidden
		if strings.Contains(err.Error(), "rate limit exceeded") {
			return http.StatusTooManyRequests, apierrors.CodeRateLimitExceeded, "Too many requests. Please try again later."
		}
		return http.StatusForbidden, apierrors.CodeForbidden, "You do not have permission to perform this action."
	case errors.Is(err, domain.ErrAuthenticationFailed):
		// Use constant from apierrors package
		return http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication failed. Please check your credentials."
	case errors.Is(err, domain.ErrUnauthenticated):
		// Use constant from apierrors package
		return http.StatusUnauthorized, apierrors.CodeUnauthenticated, "Authentication required. Please log in."
	default:
		// Any other error is treated as an internal server error
		// Use constant from apierrors package
		return http.StatusInternalServerError, apierrors.CodeInternalError, "An unexpected internal error occurred."
	}
}

// RespondError maps a domain error to an HTTP status code and JSON error response.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := MapDomainErrorToHTTP(err)

	reqID := GetReqID(r.Context()) // Get request ID from context
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
