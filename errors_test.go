package paperless

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with operation",
			err: &Error{
				StatusCode: 404,
				Message:    "Not Found",
				Op:         "GetDocument",
			},
			want: "GetDocument: 404 Not Found",
		},
		{
			name: "without operation",
			err: &Error{
				StatusCode: 500,
				Message:    "Internal Server Error",
			},
			want: "500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "404 error",
			err: &Error{
				StatusCode: 404,
				Message:    "Not Found",
			},
			want: true,
		},
		{
			name: "500 error",
			err: &Error{
				StatusCode: 500,
				Message:    "Internal Server Error",
			},
			want: false,
		},
		{
			name: "wrapped 404 error",
			err:  errors.New("wrapped: " + (&Error{StatusCode: 404}).Error()),
			want: false, // Not wrapped with %w, so errors.As won't find it
		},
		{
			name: "non-API error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
