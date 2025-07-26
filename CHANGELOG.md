# Changelog

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
