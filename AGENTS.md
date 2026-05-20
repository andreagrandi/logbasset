# AGENTS.md - LogBasset Development Guide

## Session Start Workflow

Before making code or documentation changes in this repo:

1. Switch back to `master`.
2. Pull the latest changes with `git pull --ff-only`.
3. Create a new branch with a short, descriptive name related to the feature being added or the bug being fixed.
4. Make the requested changes on that branch.

Do not start work from an old feature branch unless the user explicitly asks to continue that branch.

## Adding a Ticket, Issue, or Bug

When the user asks to add a "ticket", "issue", or "bug" to this repo, all of the
steps below are required — creating the GitHub issue alone is not enough.

**1. Create the issue** with the repo label and an area label:

```bash
gh issue create --repo andreagrandi/logbasset \
  --title "<concise title>" \
  --body "<description>" \
  --label "logbasset" \
  --label "<area>"
```

- Always apply the `logbasset` label — every work item in this repo carries it.
- Add the matching area label when one exists: `feature`, `ux`, `agent`,
  `docs`, `security`, `testing`, `release`. The `Reliability` and `Packaging`
  areas have no label; for those, set only the project Area field (step 3).
- Add a type label when it fits: `bug` (bug report), `enhancement` (feature
  request), or `documentation` (docs-only work).

**2. Add the issue to the "CLI Tools" project** and capture the item ID
(https://github.com/users/andreagrandi/projects/1):

```bash
ITEM_ID=$(gh project item-add 1 --owner andreagrandi \
  --url <issue-url> --format json --jq .id)
```

**3. Set Priority, Area, and Status** on the project item. The project and
field IDs are identical for every repo on this board:

- Project ID: `PVT_kwHOAAm1584BYDlZ`
- Priority — field `PVTSSF_lAHOAAm1584BYDlZzhTMDck`:
  High `ed3787e3`, Medium `3e3ea407`, Low `994234f4`
- Area — field `PVTSSF_lAHOAAm1584BYDlZzhTMDco`:
  Reliability `6595432d`, Packaging `6895c50a`, UX `2bc024bb`,
  Testing `0d5bc016`, Feature `6390f97d`, Agent `3a2d6f7e`,
  Docs `f5c50514`, Security `062d12a3`, Release `b344aeab`
- Status — field `PVTSSF_lAHOAAm1584BYDlZzhTMDNQ`:
  Todo `f75ad846`, In Progress `47fc9ee4`, Done `98236657`

```bash
# Priority — always set it
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDck \
  --single-select-option-id <priority-option-id>

# Area — match the area label from step 1
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDco \
  --single-select-option-id <area-option-id>

# Status — new tickets start as Todo
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDNQ \
  --single-select-option-id f75ad846
```

If the user does not state a priority or area, ask before creating the issue.
Follow the conventions of existing project issues — do not invent new labels or
fields.

## When Adding User-Facing Features

Any change that adds, renames, or removes a flag, command, output format, or
exit code must update every place users (and agents) discover it. Skipping any
of these leaves the docs lying:

- `README.md` — user-facing examples and option lists.
- `CHANGELOG.md` — entry under `[Unreleased]` referencing the issue/PR.
- `CONTEXT.md` **and** `internal/cli/context_embed.md` — both must stay
  byte-identical; the embed file is compiled into the `context` command.
- `internal/cli/schema.go` — enum/default/description entries for the
  affected command so `logbasset schema` reports the new state.
- Help strings on the relevant `cobra.Command` flag definitions.

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
│   ├── logging/            # Structured logging with logrus
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

### Logging Configuration
LogBasset uses structured logging with logrus:
- Log levels: debug, info, warn, error (default: info)
- Configurable via `--log-level` flag, `scalyr_log_level` env var, or config file
- Structured output with timestamps and contextual fields
- Debug level shows API request/response details when verbose mode is enabled

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
- `make smoke-test` - Smoke-test the built binary (no API credentials needed)
- `make fmt` - Format code
- `make vet` - Static analysis
- `make lint` - Run linter (requires golangci-lint)
- `make clean` - Clean build artifacts
- `make deps` - Install and tidy dependencies
- `make build-all` - Build for multiple platforms
- `go test -v ./internal/client -run TestNew` - Run a single test function

## CLI Usage with Context Support
All commands now support timeout and cancellation:

```bash
# Set custom timeout (examples)
logbasset query "error" --timeout 60s
logbasset tail "error" --timeout 2m --lines 100
logbasset power-query "count by server" --start 1h --timeout 30s

# Graceful cancellation with Ctrl+C
logbasset tail ".*" --lines 1000
# Press Ctrl+C to stop cleanly

# Default timeout is 30 seconds for all operations
```

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
- `ContextError` - Context cancellation/timeout errors (exit code 1)

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

// Context-related errors:
return errors.NewContextError("operation timed out", context.DeadlineExceeded)
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

## Context Propagation
LogBasset implements comprehensive context propagation for:

### Timeout Control
- Global `--timeout` flag available for all commands (default: 30s)
- Context deadlines are properly enforced in all HTTP operations
- Clear error messages when operations time out

### Cancellation Support
- All commands support graceful cancellation with Ctrl+C (SIGINT/SIGTERM)
- Tail command exits cleanly when cancelled without error messages
- Context cancellation is properly propagated through the entire call stack

### Implementation
```go
// CLI commands create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
defer cancel()

// Signal handling for graceful cancellation
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sigChan
    cancel()
}()

// All client methods accept and use context
result, err := client.Query(ctx, params)
```

### Error Handling
- Context cancellation errors provide user-friendly messages
- Timeout errors suggest using `--timeout` flag to increase duration
- Distinguished between user cancellation (Ctrl+C) and timeout scenarios

## Testing Guidelines

### Test Coverage
LogBasset has comprehensive test coverage across all packages:
- **Client package**: 69.5% coverage with HTTP client abstraction
- **Output package**: 98.9% coverage for all formatters 
- **Config package**: 89.1% coverage for configuration management
- **Errors package**: 75.9% coverage for error handling
- **App package**: 100.0% coverage for version information

### Interface-Based Testing
The codebase uses interface abstractions for better testability:

```go
// HTTPClient interface allows dependency injection for testing
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// ClientInterface defines the contract for the API client
type ClientInterface interface {
    Query(ctx context.Context, params QueryParams) (*QueryResponse, error)
    // ... other methods
}

// MockHTTPClient for testing
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}
```

### HTTP Client Dependency Injection
The client supports dependency injection for better testing:

```go
// Standard constructor for production use
client := client.New("token", "server", false)

// Constructor with custom HTTP client for testing
mockClient := &MockHTTPClient{...}
client := client.NewWithHTTPClient("token", "server", false, mockClient)
```

### Mock Testing with httptest
Tests use Go's `httptest` package for realistic HTTP testing:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Verify request structure
    assert.Equal(t, "POST", r.Method)
    assert.Equal(t, "/api/query", r.URL.Path)
    
    // Return mock response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "success"}`))
}))
defer server.Close()

client := client.New("test-token", server.URL, false)
// Test client methods...
```

### Test Organization
- Use table-driven tests for multiple test cases
- Test both success and error scenarios  
- Verify request structure and response parsing
- Test interface implementations with compile-time checks
- Cover edge cases like empty data, invalid input, network errors

## Project style
- When you generate or update the CHANGELOD.md, you must be concise
- Unless I ask you to bump the version, new additions to the CHANGELOG.md must be filled under [Unreleased] section on top.
