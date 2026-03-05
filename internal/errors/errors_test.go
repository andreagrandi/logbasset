package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLogBassetError_ToJSON(t *testing.T) {
	tests := []struct {
		name         string
		err          *LogBassetError
		expectType   string
		expectMsg    string
		expectSugg   string
		expectCode   int
	}{
		{
			name:       "auth error",
			err:        NewAuthError("API token is required", nil),
			expectType: "AUTH_ERROR",
			expectMsg:  "API token is required",
			expectCode: ExitAuth,
		},
		{
			name:       "validation error",
			err:        NewValidationError("invalid count", fmt.Errorf("too high")),
			expectType: "VALIDATION_ERROR",
			expectMsg:  "invalid count",
			expectCode: ExitValidation,
		},
		{
			name:       "network error",
			err:        NewNetworkError("connection refused", nil),
			expectType: "NETWORK_ERROR",
			expectMsg:  "connection refused",
			expectCode: ExitNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.err.ToJSON()

			var payload struct {
				Error struct {
					Type       string `json:"type"`
					Message    string `json:"message"`
					Suggestion string `json:"suggestion"`
					ExitCode   int    `json:"exit_code"`
				} `json:"error"`
			}

			err := json.Unmarshal(data, &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expectType, payload.Error.Type)
			assert.Equal(t, tt.expectMsg, payload.Error.Message)
			assert.Equal(t, tt.expectCode, payload.Error.ExitCode)
			assert.NotEmpty(t, payload.Error.Suggestion)
		})
	}
}

func TestToJSON_ValidJSON(t *testing.T) {
	err := NewAPIError("something failed", fmt.Errorf("cause"))
	data := err.ToJSON()

	assert.True(t, json.Valid(data), "ToJSON should produce valid JSON")
}

func TestNewContextError(t *testing.T) {
	tests := []struct {
		name     string
		cause    error
		expected string
	}{
		{
			name:     "context cancelled",
			cause:    context.Canceled,
			expected: "Operation was cancelled by user (Ctrl+C)",
		},
		{
			name:     "context deadline exceeded",
			cause:    context.DeadlineExceeded,
			expected: "Operation timed out. Try increasing the timeout with --timeout flag",
		},
		{
			name:     "other error",
			cause:    fmt.Errorf("other error"),
			expected: "Operation was cancelled or timed out",
		},
		{
			name:     "nil cause",
			cause:    nil,
			expected: "Operation was cancelled or timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewContextError("test message", tt.cause)
			assert.Equal(t, ContextError, err.Type)
			assert.Equal(t, "test message", err.Message)
			assert.Equal(t, tt.expected, err.Suggestion)
			assert.Equal(t, tt.cause, err.Cause)
			assert.Equal(t, ExitGeneral, err.ExitCode)
		})
	}
}
