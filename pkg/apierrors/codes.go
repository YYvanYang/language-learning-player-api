// ============================================
// FILE: pkg/apierrors/codes.go (NEW FILE)
// ============================================
package apierrors

// Standard API Error Codes
const (
	CodeNotFound          = "NOT_FOUND"
	CodeConflict          = "RESOURCE_CONFLICT"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeForbidden         = "FORBIDDEN"
	CodeUnauthenticated   = "UNAUTHENTICATED"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED" // Added for potential use
)
