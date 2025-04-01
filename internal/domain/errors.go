// internal/domain/errors.go
package domain

import "errors"

// Standard business logic errors
var (
	// ErrNotFound indicates that a requested entity could not be found.
	ErrNotFound = errors.New("entity not found")
	// ErrConflict indicates a conflict, e.g., resource already exists.
	ErrConflict = errors.New("resource conflict")
	// ErrInvalidArgument indicates that the input provided was invalid.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrPermissionDenied indicates that the user does not have permission for the action.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrAuthenticationFailed indicates that user authentication failed.
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrUnauthenticated indicates the user needs to be authenticated.
	ErrUnauthenticated = errors.New("unauthenticated") // Could be used by middleware later
)