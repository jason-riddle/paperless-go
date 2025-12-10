package paperless

import (
	"errors"
	"fmt"
)

// Error represents an API error.
type Error struct {
	StatusCode int
	Message    string
	Op         string // Operation that failed (e.g., "GetDocument")
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %d %s", e.Op, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%d %s", e.StatusCode, e.Message)
}

// IsNotFound reports whether err indicates a 404 response.
func IsNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}
