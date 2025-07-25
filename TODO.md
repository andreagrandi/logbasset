# LogBasset Improvement Plan

This document outlines improvements to make LogBasset follow Go CLI application best practices.

## Configuration & Environment Management

- [ ] **Add structured configuration management** - Replace scattered environment variable reads with a config package using libraries like Viper
- [ ] **Configuration validation** - Add validation for config values (e.g., server URLs, token format)
- [ ] **Config file support** - Support for config files (YAML/JSON) in addition to environment variables
- [ ] **XDG Base Directory compliance** - Store config in standard locations (`~/.config/logbasset/`)

## Error Handling & User Experience

- [ ] **Structured error types** - Define custom error types for different failure modes (network, auth, validation)
- [ ] **Better error messages** - More descriptive errors with suggestions for resolution
- [ ] **Exit codes** - Use standard exit codes (0=success, 1=general error, 2=misuse, etc.)
- [ ] **Progress indicators** - Add spinners/progress bars for long-running operations
- [ ] **Graceful shutdown** - Handle SIGINT/SIGTERM properly in tail command

## Logging & Observability

- [ ] **Structured logging** - Use a logging library (logrus/zap) instead of fmt.Fprintf to stderr
- [ ] **Log levels** - Support different verbosity levels (debug, info, warn, error)
- [ ] **Request ID tracking** - Add correlation IDs for API requests when verbose mode is enabled

## Code Organization & Architecture

### Project Structure Improvements
- [ ] **Add standard Go directories** - Create `pkg/`, `cmd/`, `configs/`, `scripts/`, `docs/`, `examples/` directories
- [ ] **Move main.go to cmd/logbasset/** - Follow standard Go project layout
- [ ] **Split monolithic client package** - Break down 408-line client.go into focused files (query.go, tail.go, auth.go)
- [ ] **Create dedicated packages**:
  - [ ] `internal/config/` - Configuration management and validation
  - [ ] `internal/output/` - Output formatting (JSON, CSV, table formatters)
  - [ ] `internal/errors/` - Centralized error handling
  - [ ] `internal/app/` - Application logic and version info

### Architecture Improvements
- [ ] **Dependency injection** - Make HTTP client configurable/injectable for better testing
- [ ] **Interface abstractions** - Add interfaces for the client to improve testability
- [ ] **Context propagation** - Ensure context is properly passed through all operations
- [ ] **Separate business logic** - Move formatting/output logic out of command handlers
- [ ] **Eliminate global variables** - Remove package-level vars (token, server, verbose) from cmd package
- [ ] **Reduce tight coupling** - Commands shouldn't directly instantiate client
- [ ] **Extract output logic** - Move formatting logic from individual commands to shared formatters

## CLI Best Practices

- [ ] **Input validation** - Validate all user inputs (time formats, counts, etc.) before API calls
- [ ] **Auto-completion** - Add shell completion support (bash, zsh, fish)
- [ ] **Help improvements** - Better examples in help text and man page generation
- [ ] **Flag consistency** - Standardize flag naming conventions across commands
- [ ] **Global flags** - Move common flags (server, token) to persistent flags

## Security

- [ ] **Token masking** - Mask tokens in verbose output and logs
- [ ] **Secure defaults** - Use secure HTTP client settings (timeouts, TLS verification)
- [ ] **Input sanitization** - Sanitize user inputs before sending to API

## Testing

- [ ] **Increase test coverage** - Add comprehensive unit tests for all packages
- [ ] **Integration tests** - Add tests that verify CLI behavior end-to-end
- [ ] **Mock HTTP responses** - Use httptest for testing client functionality
- [ ] **Table-driven tests** - Convert existing tests to table-driven format where applicable

## Performance & Reliability

- [ ] **Connection pooling** - Configure HTTP client with appropriate connection limits
- [ ] **Retry logic** - Add exponential backoff for transient failures
- [ ] **Rate limiting** - Implement client-side rate limiting to respect API limits
- [ ] **Memory optimization** - Stream large responses instead of loading everything into memory

## Documentation & Maintenance

- [ ] **Add godoc comments** - Document all exported functions, types, and packages
- [ ] **Version information** - Embed build info (commit, date) in binary
- [ ] **Changelog automation** - Use conventional commits and automated changelog generation
- [ ] **License headers** - Add license headers to source files

## Build & Distribution

- [ ] **Goreleaser optimization** - Improve release configuration for better artifacts
- [ ] **Docker support** - Add Dockerfile and multi-stage builds
- [ ] **Package managers** - Add support for more package managers (Scoop for Windows, etc.)
- [ ] **Binary size optimization** - Use build flags to reduce binary size

## Output & Formatting

- [ ] **Consistent output** - Standardize output formats across all commands
- [ ] **Color support** - Add colored output with ability to disable
- [ ] **Paging support** - Add built-in paging for long output
- [ ] **Template support** - Allow custom output templates

## Development Experience

- [ ] **Pre-commit hooks** - Add git hooks for formatting, linting, and testing
- [ ] **Development tooling** - Add air/realize for hot reloading during development
- [ ] **Debugging support** - Add pprof endpoints for performance debugging
- [ ] **Linting improvements** - Add more linters (golangci-lint config) and fix all issues

## Priority Recommendations

### High Priority (Core Functionality)
1. **Project structure reorganization** - Add standard directories and split monolithic packages
2. Structured configuration management
3. Better error handling and exit codes
4. Input validation
5. Increase test coverage
6. Context propagation

### Medium Priority (User Experience)
1. **Extract output formatters** - Create dedicated formatting package
2. **Eliminate global variables** - Remove package-level state
3. Progress indicators
4. Auto-completion
5. Structured logging
6. Graceful shutdown
7. Security improvements

### Low Priority (Polish)
1. Color support
2. Template support
3. Development tooling
4. Documentation improvements
5. Build optimizations

## Implementation Strategy

### Phase 1: Structural Foundation
1. Create standard Go project directories
2. Move main.go to cmd/logbasset/
3. Extract configuration package
4. Create error handling package

### Phase 2: Package Reorganization  
1. Split client package into focused files
2. Extract output formatting logic
3. Add interfaces for better testability
4. Eliminate global variables

### Phase 3: Enhanced Features
1. Improve error handling and validation
2. Add comprehensive testing
3. Implement security improvements
4. Add user experience enhancements

## Current Structure Issues

The project currently has these structural problems:
- **Monolithic client package**: Single 408-line file handling all API operations
- **Mixed responsibilities**: Commands handle both CLI logic and output formatting  
- **Global state**: Package-level variables for configuration
- **Missing standard directories**: No `pkg/`, `cmd/`, `configs/`, etc.
- **Tight coupling**: Commands directly instantiate and configure client
- **No interfaces**: Concrete types make testing and mocking difficult

## Recommended Structure

```
logbasset/
├── cmd/
│   └── logbasset/          # Main application
│       └── main.go
├── internal/
│   ├── app/                # Application logic
│   ├── cli/                # CLI command implementations  
│   ├── client/             # API client (split into focused files)
│   ├── config/             # Configuration management
│   ├── output/             # Output formatting
│   └── errors/             # Error handling
├── pkg/                    # Public APIs (if needed)
├── configs/                # Config templates
├── scripts/                # Build scripts
├── docs/                   # Documentation
└── examples/               # Usage examples
```

## Notes

- This analysis shows LogBasset is a solid foundation but could benefit significantly from implementing these Go CLI best practices
- Focus on high-priority items first for maximum impact on maintainability, user experience, and reliability
- Many improvements can be implemented incrementally without breaking existing functionality