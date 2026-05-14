package viya

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponseErrorMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      ErrorResponse
		expected string
	}{
		{
			name:     "empty error",
			err:      ErrorResponse{},
			expected: "unknown error",
		},
		{
			name: "with message",
			err: ErrorResponse{
				Message: "something went wrong",
			},
			expected: "something went wrong",
		},
		{
			name: "with error info",
			err: ErrorResponse{
				ErrorInfo: &struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					Code:    "SAS_ERROR",
					Message: "auth failed",
				},
			},
			expected: "auth failed",
		},
		{
			name: "message takes precedence over error info",
			err: ErrorResponse{
				Message: "top level message",
				ErrorInfo: &struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					Message: "nested message",
				},
			},
			expected: "top level message",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}
