# GoConnect Makefile

.PHONY: help build test clean dev-server dev-daemon dev-web

# Default target
help:
	@echo "GoConnect Development Commands:"
	@echo ""
	@echo "Development:"
	@echo "  dev-server    Start the GoConnect server"
	@echo "  dev-daemon    Start the client daemon"
	@echo "  dev-web       Start the web UI"
	@echo "  dev           Start all components"
	@echo ""
	@echo "Building:"
	@echo "  build         Build all components"
	@echo "  build-server  Build server binary"
	@echo "  build-daemon  Build daemon binary"
	@echo "  build-web     Build web UI"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run all tests"
	@echo "  test-server   Run server tests"
	@echo "  test-daemon   Run daemon tests"
	@echo "  test-web      Run web UI tests"
	@echo ""
	@echo "Utilities:"
	@echo "  clean         Clean build artifacts"
	@echo "  lint          Run linters"

# Development targets
dev-server:
	@echo "Starting GoConnect server..."
	cd server && go run ./cmd/server

dev-daemon:
	@echo "Starting client daemon..."
	cd client-daemon && go run ./cmd/daemon

dev-web:
	@echo "Starting web UI..."
	cd web-ui && npm run dev

dev: ## Start all development components
	@echo "Starting all GoConnect components..."
	@echo "Note: Run each service in separate terminals for development"
	@echo ""
	@echo "Terminal 1: make dev-server"
	@echo "Terminal 2: make dev-daemon"  
	@echo "Terminal 3: make dev-web"

# Build targets
build: build-server build-daemon build-web

build-server:
	@echo "Building server..."
	cd server && go build -o goconnect-server ./cmd/server

build-daemon:
	@echo "Building daemon..."
	cd client-daemon && go build -o goconnect-daemon ./cmd/daemon

build-web:
	@echo "Building web UI..."
	cd web-ui && npm run build

# Test targets
test: test-server test-daemon test-web

test-server:
	@echo "Running server tests..."
	cd server && go test ./...

test-daemon:
	@echo "Running daemon tests..."
	cd client-daemon && go test ./...

test-web:
	@echo "Running web UI tests..."
	cd web-ui && npm test

# Utility targets
clean:
	@echo "Cleaning build artifacts..."
	cd server && rm -f goconnect-server
	cd client-daemon && rm -f goconnect-daemon
	cd web-ui && rm -rf .next dist

lint:
	@echo "Running linters..."
	cd server && golangci-lint run
	cd client-daemon && golangci-lint run
	cd web-ui && npm run lint

# Installation targets
install: build
	@echo "Installing GoConnect..."
	sudo cp server/goconnect-server /usr/local/bin/
	sudo cp client-daemon/goconnect-daemon /usr/local/bin/

# Docker targets
docker-build:
	@echo "Building Docker images..."
	docker build -t goconnect-server ./server
	docker build -t goconnect-daemon ./client-daemon

docker-run:
	@echo "Running Docker containers..."
	docker-compose up -d

# Database targets
db-migrate:
	@echo "Running database migrations..."
	cd server && go run ./cmd/server --migrate

db-reset:
	@echo "Resetting database..."
	cd server && rm -f *.db
	$(MAKE) db-migrate

# Release targets
release: clean test lint build
	@echo "Creating release..."
	@echo "TODO: Add release automation"

# Development setup
setup:
	@echo "Setting up development environment..."
	cd web-ui && npm install
	cd server && go mod download
	cd client-daemon && go mod download
	@echo "Development environment ready!"

# Version info
version:
	@echo "GoConnect Version:"
	@echo "Server: $$(cd server && go run ./cmd/server --version 2>/dev/null || echo 'dev')"
	@echo "Daemon: $$(cd client-daemon && go run ./cmd/daemon --version 2>/dev/null || echo 'dev')"
