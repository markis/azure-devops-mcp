.PHONY: help build build-prod test test-verbose test-coverage test-coverage-html test-ci lint lint-fix format deps clean all

# Default target
help:
	@echo "Available targets:"
	@echo "  make build              - Build the binary to ./bin/azure-devops-mcp"
	@echo "  make build-prod         - Build optimized binary (stripped, no debug info)"
	@echo "  make test               - Run all tests"
	@echo "  make test-verbose       - Run all tests with verbose output"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make test-coverage-html - Generate HTML coverage report and open in browser"
	@echo "  make test-ci            - Run tests with CI flags (race detector, coverage)"
	@echo "  make lint               - Run golangci-lint"
	@echo "  make lint-fix           - Run golangci-lint with auto-fix"
	@echo "  make format             - Format code with gofumpt"
	@echo "  make deps               - Download dependencies"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make all                - Format, lint, test, and build"

# Build binary
build:
	go build -o ./bin/azure-devops-mcp ./cmd/azure-devops-mcp/...

# Build production binary (optimized, stripped)
build-prod:
	go build -ldflags="-s -w" -o ./bin/azure-devops-mcp ./cmd/azure-devops-mcp/...

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -cover

# Generate HTML coverage report
test-coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Run tests with CI configuration (race detector, full coverage)
test-ci:
	go test ./... -v -race -coverprofile=coverage.out -coverpkg=./...
	go tool cover -func=coverage.out

# Run linter
lint:
	golangci-lint run

# Run linter with auto-fix
lint-fix:
	golangci-lint run --fix

# Format code
format:
	gofumpt -l -w .

# Download dependencies
deps:
	go mod download

# Clean up dependencies
tidy:
	go mod tidy

# Remove build artifacts
clean:
	rm -rf ./bin
	rm -f coverage.out

# Run everything: format, lint, test, build
all: format lint test build
