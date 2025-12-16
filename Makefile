# GoConnect Makefile

.PHONY: all build test test-core test-cli clean proto

# Default target
all: build

# Build all components
build: build-core build-cli

build-core:
	@echo "Building Core Server..."
	cd core && go build -o ../bin/goconnect-server ./cmd/server

build-cli:
	@echo "Building CLI..."
	cd cli && go build -o ../bin/goconnect ./cmd/goconnect

# Run standard tests (short)
test: test-core test-cli

# Run all tests (including integration)
test-all:
	@echo "Running ALL Core tests..."
	cd core && go test ./...
	@echo "Running ALL CLI tests..."
	cd cli && go test ./...

test-core:
	@echo "Running Core tests..."
	cd core && go test ./... -short

test-cli:
	@echo "Running CLI tests..."
	cd cli && go test ./... -short

# Lint code
lint:
	@echo "Linting Core..."
	cd core && golangci-lint run
	@echo "Linting CLI..."
	cd cli && golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Generate Protobuf code (delegates to CLI which has correct paths)
proto:
	@echo "Generating Protobufs..."
	cd cli && $(MAKE) proto
