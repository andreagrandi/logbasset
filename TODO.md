# LogBasset Improvement Plan

This document outlines improvements to make LogBasset follow Go CLI application best practices, organized by priority.

## HIGH PRIORITY (Core Functionality)

### 1. Project Structure Reorganization
**Current Issues**: Monolithic 408-line client.go, mixed responsibilities in commands, missing standard directories, tight coupling, no interfaces, global state variables.

**Target Structure**:
```
logbasset/
├── cmd/logbasset/main.go   # Main application (moved from root)
├── internal/
│   ├── app/                # Application logic & version info
│   ├── cli/                # CLI command implementations  
│   ├── client/             # API client (split: query.go, tail.go, auth.go)
│   ├── config/             # Configuration management
│   ├── output/             # Output formatting (JSON, CSV, table)
│   └── errors/             # Centralized error handling
├── pkg/                    # Public APIs (if needed)
├── configs/                # Config templates
├── scripts/                # Build scripts
├── docs/                   # Documentation
└── examples/               # Usage examples
```

**Tasks**:
- [x] **Add standard Go directories** - Create directory structure above
- [x] **Move main.go to cmd/logbasset/** - Follow standard Go project layout
- [x] **Split monolithic client package** - Break down 408-line client.go into focused files
- [x] **Create dedicated packages** - Implement internal packages as shown above

### 2. Configuration & Environment Management
- [x] **Add structured configuration management** - Replace scattered environment variable reads with a config package using libraries like Viper
- [x] **Configuration validation** - Add validation for config values (e.g., server URLs, token format)
- [x] **Eliminate global variables** - Remove package-level vars (token, server, verbose) from cmd package

### 3. Error Handling Improvements
- [x] **Structured error types** - Define custom error types for different failure modes (network, auth, validation)
- [x] **Better error messages** - More descriptive errors with suggestions for resolution
- [x] **Exit codes** - Use standard exit codes (0=success, 1=general error, 2=misuse, etc.)

### 4. Input Validation & CLI Best Practices
- [x] **Input validation** - Validate all user inputs (time formats, counts, etc.) before API calls
- [x] **Flag consistency** - Standardize flag naming conventions across commands  
- [x] **Global flags** - Move common flags (server, token) to persistent flags

### 5. Testing Foundation
- [x] **Increase test coverage** - Add comprehensive unit tests for all packages
- [x] **Interface abstractions** - Add interfaces for the client to improve testability
- [x] **Mock HTTP responses** - Use httptest for testing client functionality
- [x] **Dependency injection** - Make HTTP client configurable/injectable for better testing

### 6. Context Propagation
- [x] **Context propagation** - Ensure context is properly passed through all operations

## MEDIUM PRIORITY (User Experience & Architecture)

### 7. Output & Formatting Improvements
- [ ] **Extract output logic** - Move formatting logic from individual commands to shared formatters
- [ ] **Separate business logic** - Move formatting/output logic out of command handlers
- [ ] **Consistent output** - Standardize output formats across all commands

### 8. Architecture Improvements
- [ ] **Reduce tight coupling** - Commands shouldn't directly instantiate client
- [ ] **Config file support** - Support for config files (YAML/JSON) in addition to environment variables

### 9. User Experience Enhancements
- [ ] **Progress indicators** - Add spinners/progress bars for long-running operations
- [ ] **Graceful shutdown** - Handle SIGINT/SIGTERM properly in tail command
- [ ] **Auto-completion** - Add shell completion support (bash, zsh, fish)
- [ ] **Help improvements** - Better examples in help text and man page generation

### 10. Logging & Observability
- [ ] **Structured logging** - Use a logging library (logrus/zap) instead of fmt.Fprintf to stderr
- [ ] **Log levels** - Support different verbosity levels (debug, info, warn, error)

### 11. Security Improvements
- [ ] **Token masking** - Mask tokens in verbose output and logs
- [ ] **Secure defaults** - Use secure HTTP client settings (timeouts, TLS verification)
- [ ] **Input sanitization** - Sanitize user inputs before sending to API

### 12. Enhanced Testing
- [ ] **Integration tests** - Add tests that verify CLI behavior end-to-end
- [ ] **Table-driven tests** - Convert existing tests to table-driven format where applicable

## LOW PRIORITY (Polish & Distribution)

### 13. Performance & Reliability
- [ ] **Connection pooling** - Configure HTTP client with appropriate connection limits
- [ ] **Retry logic** - Add exponential backoff for transient failures
- [ ] **Rate limiting** - Implement client-side rate limiting to respect API limits
- [ ] **Memory optimization** - Stream large responses instead of loading everything into memory

### 14. Advanced Output Features
- [ ] **Color support** - Add colored output with ability to disable
- [ ] **Paging support** - Add built-in paging for long output
- [ ] **Template support** - Allow custom output templates

### 15. Development Experience
- [ ] **Pre-commit hooks** - Add git hooks for formatting, linting, and testing
- [ ] **Development tooling** - Add air/realize for hot reloading during development
- [ ] **Linting improvements** - Add more linters (golangci-lint config) and fix all issues
- [ ] **Debugging support** - Add pprof endpoints for performance debugging

### 16. Documentation & Maintenance
- [ ] **Add godoc comments** - Document all exported functions, types, and packages
- [ ] **Version information** - Embed build info (commit, date) in binary
- [ ] **License headers** - Add license headers to source files
- [ ] **Changelog automation** - Use conventional commits and automated changelog generation

### 17. Build & Distribution
- [ ] **Goreleaser optimization** - Improve release configuration for better artifacts
- [ ] **Docker support** - Add Dockerfile and multi-stage builds
- [ ] **Package managers** - Add support for more package managers (Scoop for Windows, etc.)
- [ ] **Binary size optimization** - Use build flags to reduce binary size

### 18. Advanced Configuration
- [ ] **XDG Base Directory compliance** - Store config in standard locations (`~/.config/logbasset/`)
- [ ] **Request ID tracking** - Add correlation IDs for API requests when verbose mode is enabled

## Notes

- LogBasset has a solid foundation but will benefit significantly from implementing Go CLI best practices
- The priority ordering ensures foundational changes are made first, enabling cleaner implementation of features later
- Many improvements can be implemented incrementally without breaking existing functionality