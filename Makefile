.PHONY: build test lint clean install

# Build variables
BINARY_NAME=cost-blame
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) .

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Run all checks (format, test, lint)
check: fmt test lint
	@echo "All checks passed!"

# Development build with race detector
dev:
	@echo "Building for development..."
	$(GO) build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev .

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy go.mod"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  check         - Run fmt, test, and lint"
	@echo "  dev           - Build with race detector"
	@echo "  help          - Show this help message"
