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

# Run tests
test: test-core test-cli

test-core:
	@echo "Running Core tests..."
	cd core && go test ./... -short

test-cli:
	@echo "Running CLI tests..."
	cd cli && go test ./... -short

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Generate Protobuf code (delegates to CLI which has correct paths)
proto:
	@echo "Generating Protobufs..."
	cd cli && $(MAKE) proto
