# GoConnect Makefile
GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: all build test test-core test-cli clean proto

# Default target
all: build

# Build all components
build: build-core build-cli

build-core:
	@echo "Building Core Server..."
	cd core && $(GO) build -o ../bin/goconnect-server ./cmd/server

build-cli:
	@echo "Building CLI..."
	cd cli && $(GO) build -o ../bin/goconnect ./cmd/goconnect

# Run standard tests (short)
test: test-core test-cli

# Run all tests (including integration)
test-all:
	@echo "Running ALL Core tests..."
	cd core && $(GO) test ./...
	@echo "Running ALL CLI tests..."
	cd cli && $(GO) test ./...

test-core:
	@echo "Running Core tests..."
	cd core && $(GO) test ./... -short

test-cli:
	@echo "Running CLI tests..."
	cd cli && $(GO) test ./... -short

# Lint code
lint:
	@echo "Linting Core..."
	cd core && $(GOLANGCI_LINT) run -c ../.golangci.yml
	@echo "Linting CLI..."
	cd cli && $(GOLANGCI_LINT) run -c ../.golangci.yml

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Generate Protobuf code (delegates to CLI which has correct paths)
proto:
	@echo "Generating Protobufs..."
	cd cli && $(MAKE) proto

# Release management
release-dry:
	@echo "Running release dry-run..."
	goreleaser release --snapshot --clean

release:
	@echo "Creating release..."
	@read -p "Enter version (e.g., v1.0.0): " version; \
	if [ -z "$$version" ]; then echo "Version required"; exit 1; fi; \
	if git rev-parse "$$version" >/dev/null 2>&1; then echo "Tag exists"; exit 1; fi; \
	git tag -a "$$version" -m "Release $$version"; \
	git push origin "$$version"; \
	echo "Release $$version triggered"

