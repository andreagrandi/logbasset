package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogBassetError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *LogBassetError
		want string
	}{
		{
			name: "error with suggestion and cause",
			err: &LogBassetError{
				Type:       AuthError,
				Message:    "invalid token",
				Suggestion: "check your API token",
				Cause:      fmt.Errorf("unauthorized"),
			},
			want: "AUTH_ERROR: invalid token\nSuggestion: check your API token\nCaused by: unauthorized",
		},
		{
			name: "error with suggestion only",
			err: &LogBassetError{
				Type:       ConfigError,
				Message:    "missing config file",
				Suggestion: "create a config file",
			},
			want: "CONFIG_ERROR: missing config file\nSuggestion: create a config file",
		},
		{
			name: "error with cause only",
			err: &LogBassetError{
				Type:    NetworkError,
				Message: "connection failed",
				Cause:   fmt.Errorf("timeout"),
			},
			want: "NETWORK_ERROR: connection failed\nCaused by: timeout",
		},
		{
			name: "error with message only",
			err: &LogBassetError{
				Type:    ParseError,
				Message: "invalid JSON",
			},
			want: "PARSE_ERROR: invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestLogBassetError_GetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *LogBassetError
		expected int
	}{
		{
			name:     "auth error",
			err:      &LogBassetError{Type: AuthError},
			expected: ExitAuth,
		},
		{
			name:     "network error",
			err:      &LogBassetError{Type: NetworkError},
			expected: ExitNetwork,
		},
		{
			name:     "config error",
			err:      &LogBassetError{Type: ConfigError},
			expected: ExitConfig,
		},
		{
			name:     "validation error",
			err:      &LogBassetError{Type: ValidationError},
			expected: ExitValidation,
		},
		{
			name:     "usage error",
			err:      &LogBassetError{Type: UsageError},
			expected: ExitValidation,
		},
		{
			name:     "custom exit code",
			err:      &LogBassetError{Type: APIError, ExitCode: 99},
			expected: 99,
		},
		{
			name:     "default exit code",
			err:      &LogBassetError{Type: ParseError},
			expected: ExitGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.GetExitCode())
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string, error) *LogBassetError
		errType     ErrorType
		exitCode    int
	}{
		{"NewAuthError", NewAuthError, AuthError, ExitAuth},
		{"NewAPIError", NewAPIError, APIError, ExitGeneral},
		{"NewConfigError", NewConfigError, ConfigError, ExitConfig},
		{"NewNetworkError", NewNetworkError, NetworkError, ExitNetwork},
		{"NewParseError", NewParseError, ParseError, ExitGeneral},
		{"NewValidationError", NewValidationError, ValidationError, ExitValidation},
		{"NewUsageError", NewUsageError, UsageError, ExitUsage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := fmt.Errorf("test cause")
			err := tt.constructor("test message", cause)

			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, "test message", err.Message)
			assert.Equal(t, cause, err.Cause)
			assert.Equal(t, tt.exitCode, err.ExitCode)
			assert.NotEmpty(t, err.Suggestion)
		})
	}
}

func TestLogBassetError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := NewAuthError("test", cause)

	assert.Equal(t, cause, err.Unwrap())
}

func TestLogBassetError_NilCause(t *testing.T) {
	err := NewAuthError("test", nil)
	assert.Nil(t, err.Unwrap())
}
