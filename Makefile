.PHONY: build test clean fmt vet lint

# Build the binary
build:
	@mkdir -p bin
	go build -o bin/logbasset ./cmd/logbasset

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests for specific package
test-client:
	go test ./internal/client

# Run tests for specific package
test-cli:
	go test ./internal/cli

# Format code
fmt:
	go fmt ./...

# Run static analysis
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/logbasset-linux-amd64 ./cmd/logbasset
	GOOS=darwin GOARCH=amd64 go build -o bin/logbasset-darwin-amd64 ./cmd/logbasset
	GOOS=darwin GOARCH=arm64 go build -o bin/logbasset-darwin-arm64 ./cmd/logbasset
	GOOS=windows GOARCH=amd64 go build -o bin/logbasset-windows-amd64.exe ./cmd/logbasset

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  test-client  - Run client tests only"
	@echo "  test-cli     - Run CLI tests only"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run static analysis"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install and tidy dependencies"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  help         - Show this help"