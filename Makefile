# GoConnect Makefile

.PHONY: help build test clean dev

# Default target
help:
	@echo "GoConnect Development Commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev-core      Start GoConnect Core (backend)"
	@echo "  dev-cli       Start GoConnect CLI"
	@echo "  dev-desktop   Start GoConnect Desktop (dev mode)"
	@echo ""
	@echo "Building:"
	@echo "  build         Build all components"
	@echo "  build-core    Build core library"
	@echo "  build-cli     Build CLI binary"
	@echo "  build-desktop Build desktop app"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run all tests"
	@echo "  test-core     Run core tests"
	@echo "  test-cli      Run CLI tests"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         Clean build artifacts"
	@echo "  lint          Run linters"
	@echo "  setup         Setup development environment"

# =============================================================================
# Development
# =============================================================================

dev-core:
	@echo "Starting GoConnect Core..."
	cd server && go run ./cmd/server

dev-cli:
	@echo "Starting GoConnect CLI..."
	cd client-daemon && go run ./cmd/daemon

dev-desktop:
	@echo "Starting GoConnect Desktop (dev mode)..."
	cd desktop-client && npm run tauri dev

# =============================================================================
# Build
# =============================================================================

build: build-core build-cli

build-core:
	@echo "Building GoConnect Core..."
	cd server && go build -o ../dist/goconnect-core ./cmd/server

build-cli:
	@echo "Building GoConnect CLI..."
	cd client-daemon && go build -o ../dist/goconnect-cli ./cmd/daemon

build-desktop:
	@echo "Building GoConnect Desktop..."
	cd desktop-client && npm run tauri build

build-all: build build-desktop
	@echo "All builds complete!"

# Cross-platform CLI builds
build-cli-all:
	@echo "Building CLI for all platforms..."
	@mkdir -p dist
	cd client-daemon && GOOS=linux GOARCH=amd64 go build -o ../dist/goconnect-cli-linux-amd64 ./cmd/daemon
	cd client-daemon && GOOS=linux GOARCH=arm64 go build -o ../dist/goconnect-cli-linux-arm64 ./cmd/daemon
	cd client-daemon && GOOS=darwin GOARCH=amd64 go build -o ../dist/goconnect-cli-darwin-amd64 ./cmd/daemon
	cd client-daemon && GOOS=darwin GOARCH=arm64 go build -o ../dist/goconnect-cli-darwin-arm64 ./cmd/daemon
	cd client-daemon && GOOS=windows GOARCH=amd64 go build -o ../dist/goconnect-cli-windows-amd64.exe ./cmd/daemon

# =============================================================================
# Testing
# =============================================================================

test: test-core test-cli

test-core:
	@echo "Running Core tests..."
	cd server && go test ./...

test-cli:
	@echo "Running CLI tests..."
	cd client-daemon && go test ./...

# =============================================================================
# Utilities
# =============================================================================

clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/
	cd server && rm -f goconnect-core
	cd client-daemon && rm -f goconnect-cli
	cd desktop-client && rm -rf src-tauri/target dist

lint:
	@echo "Running linters..."
	cd server && golangci-lint run
	cd client-daemon && golangci-lint run

setup:
	@echo "Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	cd server && go mod download
	cd client-daemon && go mod download
	@echo ""
	@echo "2. Installing Desktop dependencies..."
	cd desktop-client && npm install
	@echo ""
	@echo "Development environment ready!"
	@echo ""
	@echo "Quick start:"
	@echo "  make dev-desktop  - Start desktop app"
	@echo "  make dev-cli      - Start CLI app"

# =============================================================================
# Database
# =============================================================================

db-migrate:
	@echo "Running database migrations..."
	cd server && go run ./cmd/server --migrate

db-reset:
	@echo "Resetting database..."
	cd server && rm -f *.db
	$(MAKE) db-migrate

# =============================================================================
# Docker
# =============================================================================

docker-build:
	@echo "Building Docker image..."
	docker build -t goconnect-core ./server

docker-run:
	@echo "Running Docker container..."
	docker run -d -p 8080:8080 goconnect-core

# =============================================================================
# Release
# =============================================================================

release: clean test lint build-all
	@echo "Release build complete!"
	@echo "Artifacts in dist/ directory"

version:
	@echo "GoConnect v3.0.0"
