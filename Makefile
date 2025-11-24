.PHONY: build test test-coverage install clean build-all fmt lint help

# Default target
.DEFAULT_GOAL := help

# Build the binary for current platform
build: ## Build apex-bench binary
	@echo "Building apex-bench..."
	go build -o apex-bench ./cmd/apex-bench
	@echo "Build complete: ./apex-bench"

# Run all tests
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install binary to GOPATH/bin
install: ## Install apex-bench to GOPATH/bin
	@echo "Installing apex-bench..."
	go install ./cmd/apex-bench
	@echo "Installed to: $(shell go env GOPATH)/bin/apex-bench"

# Clean build artifacts
clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -f apex-bench
	rm -f coverage.out coverage.html
	rm -rf bin/
	rm -f *.apex.tmp
	@echo "Clean complete"

# Build for all platforms
build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/apex-bench-linux-amd64 ./cmd/apex-bench
	GOOS=linux GOARCH=arm64 go build -o bin/apex-bench-linux-arm64 ./cmd/apex-bench
	GOOS=darwin GOARCH=amd64 go build -o bin/apex-bench-darwin-amd64 ./cmd/apex-bench
	GOOS=darwin GOARCH=arm64 go build -o bin/apex-bench-darwin-arm64 ./cmd/apex-bench
	GOOS=windows GOARCH=amd64 go build -o bin/apex-bench-windows-amd64.exe ./cmd/apex-bench
	@echo "Build complete: bin/"
	@ls -lh bin/

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run ./...

# Display help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
