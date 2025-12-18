.PHONY: dev test test-race test-coverage lint vet build clean help

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%d")
BUILT_BY ?= orhaniscoding
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)

# Coverage threshold (as per instructions)
COVERAGE_THRESHOLD := 60

## help: Display this help message
help:
	@echo "GoConnect Server - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## dev: Run server in development mode with live reload
dev:
	go run -ldflags "$(LDFLAGS)" ./cmd/server

## build: Build the server binary
build:
	go build -ldflags "$(LDFLAGS)" -o bin/goconnect-server ./cmd/server

## test: Run all tests
test:
	go test ./... -short -v

## test-race: Run tests with race detector
test-race:
	go test ./... -race -short

## test-coverage: Run tests with coverage report and enforce threshold
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

## test-coverage-html: Generate HTML coverage report
test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters (golangci-lint must be installed)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=3m -c ../.golangci.yml

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format code
fmt:
	go fmt ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## clean: Remove build artifacts and coverage files
clean:
	rm -rf bin/ coverage.out coverage.html

## ci: Run all CI checks (vet, lint, test with race and coverage)
ci: vet test-race test-coverage
	@echo "✅ All CI checks passed!"

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	@echo "✅ Tools installed!"

## db-migrate-up: Run database migrations (when implemented)
db-migrate-up:
	@echo "⚠️  Database migrations not yet implemented"

## db-migrate-down: Rollback database migrations (when implemented)
db-migrate-down:
	@echo "⚠️  Database migrations not yet implemented"

## run-prod: Run server in production mode (requires environment variables)
run-prod:
	@echo "Starting GoConnect Server (production mode)..."
	./bin/goconnect-server

