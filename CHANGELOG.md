# Changelog

## [Unreleased]

## v0.4.2 - 2025-07-30

### Added
- **Context propagation**: Complete implementation of context propagation throughout all CLI operations
- **Timeout support**: Global `--timeout` flag to configure request timeouts for all commands (default: 30s)
- **Signal handling**: Graceful cancellation with Ctrl+C for all operations, especially tail command
- **Context-aware error handling**: Enhanced error messages for timeout and cancellation scenarios
- **Comprehensive context tests**: Added test coverage for context cancellation, timeouts, and error scenarios

### Changed
- **Enhanced timeout control**: All API operations now respect context timeouts and can be cancelled by user
- **Improved error messages**: Context-related errors provide helpful suggestions for resolution
- **Better tail experience**: Tail command now supports graceful cancellation without error spam

## v0.4.1 - 2025-07-28

### Fixed
- **GoReleaser configuration**: Fixed GoReleaser v2 compatibility by renaming `homebrew` section to `brews`

## v0.4.0 - 2025-07-28

### Added
- **Comprehensive input validation system**: Implemented robust validation for time formats, counts, buckets, and query syntax across all commands
- **Validation package**: New `internal/validation` package with extensive validation functions and comprehensive test coverage
- **Enhanced CLI error handling**: All commands now provide clear validation errors with helpful suggestions for resolution
- HTTP client interfaces for better testability and dependency injection
- Comprehensive unit tests across all packages with mock HTTP responses
- `MockHTTPClient` and `NewWithHTTPClient()` constructor for testing

### Changed
- **Improved user experience**: Commands now validate inputs before making API calls, providing immediate feedback for invalid parameters
- **Standardized error handling**: All validation errors use structured error types with consistent formatting and exit codes
- **Enhanced flag handling**: Input validation is consistently applied across all CLI commands with appropriate validation flags
- Client package coverage increased from 45.2% to 69.5%
- Output package coverage increased to 98.9% with comprehensive formatter tests
- Enhanced testing infrastructure with interface-based design patterns

### Fixed
- **Time format validation**: Added support for relative time formats (24h, 1d, 30m), absolute dates, and time-only formats
- **Range validation**: Proper validation of count limits (1-5000), bucket limits (1-5000), and other numeric parameters
- **Input sanitization**: Query syntax validation prevents excessively long queries and validates column specifications

## v0.3.0 - 2025-07-26

### Added
- **Structured error handling system**: Implemented comprehensive error types with custom error structs
- **Standard Unix exit codes**: Added proper exit codes for different error conditions (auth=4, network=3, config=5, validation=6, usage=2)
- **Enhanced error messages**: All errors now include helpful suggestions for resolution
- **Comprehensive error tests**: Added full test suite for error handling with 100% coverage

### Changed
- **Error handling improvements**: Replaced generic `fmt.Errorf` with structured error types throughout codebase
- **Main application flow**: Updated main.go to use `HandleErrorAndExit()` for proper exit code handling
- **Client error handling**: Enhanced API client to use structured errors for better user experience
- **Configuration validation**: Improved config validation with specific error types and suggestions

### Technical Details
- Added `internal/errors` package with custom error types for different failure modes
- Implemented `LogBassetError` struct with Type, Message, Suggestion, Cause, and ExitCode fields
- Updated client and config packages to use new structured error system
- All errors now provide actionable suggestions to help users resolve issues

## v0.2.1 - 2025-07-26

### Fixed
- **GoReleaser configuration**: Fixed build path to reference correct main.go location (`cmd/logbasset/main.go`)
- **Version injection**: Updated ldflags to properly inject version into the correct package

## v0.2.0 - 2025-07-26

### Changed
- **Major project restructuring**: Reorganized codebase to follow standard Go project layout conventions
- **Moved main.go** to `cmd/logbasset/` directory following Go best practices
- **Split monolithic client**: Broke down large client.go into focused, single-responsibility files
- **Added new internal packages**: Introduced `app`, `config`, `output`, and `errors` packages for better organization
- **Improved modularity**: Separated concerns between CLI commands, API client, and output formatting
- **Updated build system**: Fixed Makefile to work with new project structure
- **Enhanced documentation**: Updated project guides and added comprehensive TODO roadmap

### Added
- **Structured error handling**: New error types and centralized error management
- **Output formatting system**: Dedicated package for handling different output formats
- **Configuration management**: Centralized config handling for better maintainability
- **Application metadata**: Dedicated app package for version and build information

## v0.1.5 - 2025-07-25

### Changed
- Updated release process documentation with improved instructions

## v.0.1.2 - 2025-07-25

### Fixed
- Updated all workflows to use Go 1.24 consistently
- Fixed GoReleaser configuration for proper archive formats (zip for Windows, tar.gz for Unix)

## v0.1.0 - 2025-07-25

### Added
- Complete command-line interface for Scalyr services
- **query**: Retrieve log data with filtering, columns selection, and multiple output formats
- **power-query**: Execute PowerQueries for advanced log analysis 
- **numeric-query**: Retrieve numeric/graph data with statistical functions
- **facet-query**: Find most common values for specified fields
- **timeseries-query**: Fast retrieval of precomputed numeric data
- **tail**: Live tail functionality for real-time log monitoring
- Support for multiple output formats: JSON, CSV, multiline, singleline
- Cross-platform builds for Linux, macOS, and Windows
- Environment variable configuration for API tokens and server URLs
- Comprehensive test suite and CI/CD pipeline
- Automated releases with GoReleaser

### Features
- Fast, efficient Go implementation with improved performance over Python tools
- Support for all major Scalyr query types and time range specifications
- Configurable query priorities and execution limits
- Excel-compatible CSV output format
- Pretty-printed JSON with proper formatting
- Live tail with customizable line limits and output modes
