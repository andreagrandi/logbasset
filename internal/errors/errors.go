package errors

import (
	"fmt"
)

type ErrorType string

const (
	AuthError    ErrorType = "AUTH_ERROR"
	APIError     ErrorType = "API_ERROR"
	ConfigError  ErrorType = "CONFIG_ERROR"
	NetworkError ErrorType = "NETWORK_ERROR"
	ParseError   ErrorType = "PARSE_ERROR"
)

type LogBassetError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *LogBassetError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *LogBassetError) Unwrap() error {
	return e.Cause
}

func NewAuthError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:    AuthError,
		Message: message,
		Cause:   cause,
	}
}

func NewAPIError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:    APIError,
		Message: message,
		Cause:   cause,
	}
}

func NewConfigError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:    ConfigError,
		Message: message,
		Cause:   cause,
	}
}

func NewNetworkError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:    NetworkError,
		Message: message,
		Cause:   cause,
	}
}

func NewParseError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:    ParseError,
		Message: message,
		Cause:   cause,
	}
}
