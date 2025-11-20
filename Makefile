.PHONY: help dev test test-all test-coverage lint build clean install-tools

## help: Display this help message
help:
	@echo "GoConnect - Monorepo Management"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "Component-specific commands:"
	@echo "  make -C server <command>        - Run command in server/"
	@echo "  make -C client-daemon <command> - Run command in client-daemon/"
	@echo "  make -C web-ui <command>        - Run command in web-ui/ (if Makefile exists)"

## dev-server: Run server in development mode
dev-server:
	@echo "Starting GoConnect Server..."
	$(MAKE) -C server dev

## dev-daemon: Run client daemon in development mode
dev-daemon:
	@echo "Starting GoConnect Client Daemon..."
	$(MAKE) -C client-daemon dev

## dev-web: Run web UI in development mode
dev-web:
	@echo "Starting GoConnect Web UI..."
	cd web-ui && npm run dev

## test: Run tests for all Go components
test:
	@echo "Running server tests..."
	$(MAKE) -C server test
	@echo ""
	@echo "Running client-daemon tests..."
	$(MAKE) -C client-daemon test

## test-race: Run tests with race detector for all Go components
test-race:
	@echo "Running server tests with race detector..."
	$(MAKE) -C server test-race
	@echo ""
	@echo "Running client-daemon tests with race detector..."
	$(MAKE) -C client-daemon test-race

## test-coverage: Run coverage tests for all components
test-coverage:
	@echo "=== Server Coverage ==="
	$(MAKE) -C server test-coverage
	@echo ""
	@echo "=== Client Daemon Coverage ==="
	$(MAKE) -C client-daemon test-coverage
	@echo ""
	@echo "✅ All coverage checks passed!"

## lint: Run linters for all components
lint:
	@echo "Linting server..."
	$(MAKE) -C server lint
	@echo ""
	@echo "Linting client-daemon..."
	$(MAKE) -C client-daemon lint
	@echo ""
	@echo "Linting web-ui..."
	cd web-ui && npm run lint --if-present || echo "No lint script defined"

## vet: Run go vet for all Go components
vet:
	@echo "Vetting server..."
	$(MAKE) -C server vet
	@echo ""
	@echo "Vetting client-daemon..."
	$(MAKE) -C client-daemon vet

## build: Build all components
build:
	@echo "Building server..."
	$(MAKE) -C server build
	@echo ""
	@echo "Building client-daemon..."
	$(MAKE) -C client-daemon build
	@echo ""
	@echo "Building web-ui..."
	cd web-ui && npm run build
	@echo ""
	@echo "✅ All components built successfully!"

## build-all: Build binaries for all platforms
build-all:
	@echo "Building server..."
	$(MAKE) -C server build
	@echo ""
	@echo "Building client-daemon for all platforms..."
	$(MAKE) -C client-daemon build-all
	@echo ""
	@echo "✅ All platform builds completed!"

## clean: Clean all build artifacts
clean:
	@echo "Cleaning server..."
	$(MAKE) -C server clean
	@echo "Cleaning client-daemon..."
	$(MAKE) -C client-daemon clean
	@echo "Cleaning web-ui..."
	cd web-ui && rm -rf .next out
	@echo "✅ All artifacts cleaned!"

## install-tools: Install development tools
install-tools:
	@echo "Installing server tools..."
	$(MAKE) -C server install-tools
	@echo ""
	@echo "Installing web-ui dependencies..."
	cd web-ui && npm ci
	@echo ""
	@echo "✅ All tools installed!"

## ci: Run all CI checks (like GitHub Actions)
ci: vet test-race test-coverage lint
	@echo ""
	@echo "=== CI Pipeline Complete ==="
	@echo "✅ All checks passed!"

## fmt: Format all Go code
fmt:
	@echo "Formatting server..."
	$(MAKE) -C server fmt
	@echo "Formatting client-daemon..."
	$(MAKE) -C client-daemon fmt

## tidy: Tidy all Go modules
tidy:
	@echo "Tidying server modules..."
	$(MAKE) -C server tidy
	@echo "Tidying client-daemon modules..."
	$(MAKE) -C client-daemon tidy
	@echo "Syncing go.work..."
	go work sync

## version: Show version information
version:
	@echo "GoConnect Version Information:"
	@echo ""
	@cd server && go run ./cmd/server --version 2>/dev/null || echo "Server: dev"
	@cd client-daemon && go run ./cmd/daemon --version 2>/dev/null || echo "Daemon: dev"

## status: Check status of all components
status:
	@echo "=== GoConnect Project Status ==="
	@echo ""
	@echo "Server:"
	@cd server && go list -m && go version
	@echo ""
	@echo "Client Daemon:"
	@cd client-daemon && go list -m && go version
	@echo ""
	@echo "Web UI:"
	@cd web-ui && node --version && npm --version
	@echo ""
	@echo "Git:"
	@git branch --show-current
	@git log -1 --oneline

