# AGENTS.md - LogBasset Development Guide

## Project Structure
LogBasset follows the standard Go project layout:
```
logbasset/
├── cmd/logbasset/main.go   # Main application entry point
├── internal/
│   ├── app/                # Application logic & version info
│   ├── cli/                # CLI command implementations  
│   ├── client/             # API client (split: client.go, basic_query.go, power_query.go, numeric_query.go, facet_query.go, timeseries_query.go, tail.go, types.go)
│   ├── config/             # Configuration management
│   ├── output/             # Output formatting (JSON, CSV, table)
│   └── errors/             # Centralized error handling
├── pkg/                    # Public APIs (if needed)
├── configs/                # Config templates
├── scripts/                # Build scripts
├── docs/                   # Documentation
└── examples/               # Usage examples
```

## Configuration Management

LogBasset now uses a structured configuration system with Viper:

### Configuration Sources
- Environment variables (prefixed with `scalyr_`)
- Configuration files (YAML format)
- Command-line flags
- Default values

### Configuration Locations
- `./logbasset.yaml` (current directory)
- `~/.logbasset/logbasset.yaml` (user home)  
- `~/.config/logbasset/logbasset.yaml` (XDG config)

### Configuration Validation
- API token presence validation
- Server URL format validation
- Priority value validation (high|low)
- Automatic type conversion and defaults

### Usage in Code
```go
// Get configuration with validation
cfg, err := config.New()
if err != nil {
    return err
}

// Override with flags
cfg.SetFromFlags(token, server, verbose, priority)

// Validate final config
if err := cfg.Validate(); err != nil {
    return err
}

// Get client
client := cfg.GetClient()
```

## Build/Test Commands
- `make build` - Build the CLI tool (output to bin/)
- `make test` - Run all tests
- `make test-verbose` - Run tests with verbose output
- `make test-client` - Run tests for client package only
- `make test-cli` - Run tests for CLI package only
- `make fmt` - Format code
- `make vet` - Static analysis
- `make lint` - Run linter (requires golangci-lint)
- `make clean` - Clean build artifacts
- `make deps` - Install and tidy dependencies
- `make build-all` - Build for multiple platforms
- `go test -v ./internal/client -run TestNew` - Run a single test function

## Error Handling

LogBasset uses a structured error handling system with custom error types and standard exit codes:

### Error Types
- `AuthError` - API token authentication issues (exit code 4)
- `NetworkError` - Network/connection failures (exit code 3)
- `ConfigError` - Configuration problems (exit code 5)
- `ValidationError` - Input validation failures (exit code 6)
- `UsageError` - Command usage errors (exit code 2)
- `APIError` - API response errors (exit code 1)
- `ParseError` - JSON parsing failures (exit code 1)

### Error Handling Best Practices
- Use structured errors from `internal/errors` package instead of `fmt.Errorf`
- Include helpful suggestions in error messages
- Use appropriate exit codes for different error types
- Wrap underlying errors with context using the `Cause` field

### Usage Example
```go
// Instead of:
return fmt.Errorf("API token is required")

// Use:
return errors.NewAuthError("API token is required", nil)

// With cause:
return errors.NewNetworkError("failed to connect", err)
```

### Exit Code Reference
- `0` - Success
- `1` - General error
- `2` - Usage error/command misuse
- `3` - Network error
- `4` - Authentication error
- `5` - Configuration error
- `6` - Validation error

## Code Style Guidelines
- Use Go standard formatting (gofmt)
- Package names: lowercase, single word (e.g., `client`, `cli`)
- Types: PascalCase (e.g., `QueryParams`, `Client`)
- Functions/methods: PascalCase for exported, camelCase for unexported
- Variables: camelCase (e.g., `httpClient`, `requestParams`)
- Constants: PascalCase or ALL_CAPS (e.g., `DefaultServer`, `APIVersion`)
- Use structured errors from `internal/errors` package instead of `fmt.Errorf`
- Import ordering: standard library, third-party, local packages
- Use interfaces for testability (e.g., HTTP client abstractions)
- Struct initialization: use field names for clarity
- Context propagation: pass `context.Context` as first parameter to functions making external calls
- Test files: use `testify/assert` and `testify/require` for assertions
- Test function names: `TestFunctionName` or `TestFunctionName_Scenario`
- Always defer `resp.Body.Close()` immediately after checking error for HTTP responses
- Use table-driven tests with `tests := []struct{}` pattern for multiple test cases