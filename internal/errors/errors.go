package errors

import (
	"fmt"
	"os"
)

type ErrorType string

const (
	AuthError       ErrorType = "AUTH_ERROR"
	APIError        ErrorType = "API_ERROR"
	ConfigError     ErrorType = "CONFIG_ERROR"
	NetworkError    ErrorType = "NETWORK_ERROR"
	ParseError      ErrorType = "PARSE_ERROR"
	ValidationError ErrorType = "VALIDATION_ERROR"
	UsageError      ErrorType = "USAGE_ERROR"
)

const (
	ExitSuccess    = 0
	ExitGeneral    = 1
	ExitUsage      = 2
	ExitNetwork    = 3
	ExitAuth       = 4
	ExitConfig     = 5
	ExitValidation = 6
)

type LogBassetError struct {
	Type       ErrorType
	Message    string
	Suggestion string
	Cause      error
	ExitCode   int
}

func (e *LogBassetError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Type, e.Message)
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\nSuggestion: %s", e.Suggestion)
	}
	if e.Cause != nil {
		msg += fmt.Sprintf("\nCaused by: %v", e.Cause)
	}
	return msg
}

func (e *LogBassetError) Unwrap() error {
	return e.Cause
}

func (e *LogBassetError) GetExitCode() int {
	if e.ExitCode != 0 {
		return e.ExitCode
	}
	switch e.Type {
	case AuthError:
		return ExitAuth
	case NetworkError:
		return ExitNetwork
	case ConfigError:
		return ExitConfig
	case ValidationError, UsageError:
		return ExitValidation
	default:
		return ExitGeneral
	}
}

func NewAuthError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       AuthError,
		Message:    message,
		Suggestion: "Please check your API token. You can find API tokens at https://www.scalyr.com/keys",
		Cause:      cause,
		ExitCode:   ExitAuth,
	}
}

func NewAPIError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       APIError,
		Message:    message,
		Suggestion: "Please check your query syntax and try again. Use --verbose for more details",
		Cause:      cause,
		ExitCode:   ExitGeneral,
	}
}

func NewConfigError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       ConfigError,
		Message:    message,
		Suggestion: "Check your configuration file syntax or create one with 'logbasset config init'",
		Cause:      cause,
		ExitCode:   ExitConfig,
	}
}

func NewNetworkError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       NetworkError,
		Message:    message,
		Suggestion: "Check your internet connection and server URL. Use --verbose for more details",
		Cause:      cause,
		ExitCode:   ExitNetwork,
	}
}

func NewParseError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       ParseError,
		Message:    message,
		Suggestion: "Check the response format or try again later",
		Cause:      cause,
		ExitCode:   ExitGeneral,
	}
}

func NewValidationError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       ValidationError,
		Message:    message,
		Suggestion: "Please check the command usage with --help",
		Cause:      cause,
		ExitCode:   ExitValidation,
	}
}

func NewUsageError(message string, cause error) *LogBassetError {
	return &LogBassetError{
		Type:       UsageError,
		Message:    message,
		Suggestion: "Use --help to see available commands and options",
		Cause:      cause,
		ExitCode:   ExitUsage,
	}
}

func HandleErrorAndExit(err error) {
	if err == nil {
		os.Exit(ExitSuccess)
	}

	if logbassetErr, ok := err.(*LogBassetError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", logbassetErr.Error())
		os.Exit(logbassetErr.GetExitCode())
	}

	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(ExitGeneral)
}
