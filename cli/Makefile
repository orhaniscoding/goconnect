.PHONY: dev test test-race test-coverage lint build clean help install-systemd install-launchd install-windows proto

## lint: Run linters (golangci-lint must be installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=3m -c ../.golangci.yml

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%d")
BUILT_BY ?= orhaniscoding
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)

# Coverage threshold
COVERAGE_THRESHOLD := 60

# Proto paths
PROTO_DIR := ../core/proto
PROTO_OUT := internal/proto

## help: Display this help message
help:
	@echo "GoConnect Client Daemon - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## proto: Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	protoc --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) $(PROTO_DIR)/daemon.proto
	@echo "✅ Proto generation complete!"

## dev: Run daemon in development mode
dev:
	go run -ldflags "$(LDFLAGS)" ./cmd/goconnect

## build: Build the daemon binary
build:
	go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon ./cmd/goconnect

## build-all: Build binaries for all platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon-linux-amd64 ./cmd/goconnect
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon-linux-arm64 ./cmd/goconnect
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon-darwin-amd64 ./cmd/goconnect
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon-darwin-arm64 ./cmd/goconnect
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/goconnect-daemon-windows-amd64.exe ./cmd/goconnect
	@echo "✅ All platform binaries built successfully!"

## test: Run all tests
test:
	go test ./... -short -v

## test-race: Run tests with race detector
test-race:
	go test ./... -race -short

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -short -coverprofile=coverage.out -covermode=atomic
	@echo ""
	@echo "Coverage by package:"
	@go tool cover -func=coverage.out | grep -v "total:" | tail -n +2
	@echo ""
	@TOTAL=$$(go tool cover -func=coverage.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$TOTAL%"; \
	echo "Threshold: $(COVERAGE_THRESHOLD)%"; \
	if [ "$$(echo "$$TOTAL < $(COVERAGE_THRESHOLD)" | bc -l)" -eq 1 ]; then \
		echo "❌ Coverage $$TOTAL% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$TOTAL% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## install-systemd: Install systemd service (Linux)
install-systemd:
	@echo "Installing systemd service..."
	@chmod +x service/linux/install.sh
	@sudo ./service/linux/install.sh

## install-launchd: Install launchd service (macOS)
install-launchd:
	@echo "Installing launchd service..."
	@chmod +x service/macos/install.sh
	@sudo ./service/macos/install.sh

## install-windows: Install Windows service (requires admin PowerShell)
install-windows:
	@echo "Installing Windows service..."
	@echo "Run this in an elevated PowerShell prompt:"
	@echo "  powershell -ExecutionPolicy Bypass -File service/windows/install.ps1"

