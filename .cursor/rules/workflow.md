# Development Workflow

## Building

### All Components
```bash
make build          # Build core + cli
```

### Individual Components
```bash
make build-core     # Build core daemon
make build-cli      # Build CLI
cd desktop && npm run tauri build  # Build desktop app
```

## Testing

### Run All Tests
```bash
make test           # Run core + cli tests
```

### Individual Tests
```bash
cd core && go test ./... -short    # Core tests
cd cli && go test ./... -short     # CLI tests
```

### With Coverage
```bash
cd core && go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Proto Generation

Proto files are in `core/proto/`. Generate Go code:
```bash
cd cli && make proto
```

## Adding Dependencies

```bash
# In core/
cd core && go get github.com/example/package && go mod tidy

# In cli/
cd cli && go get github.com/example/package && go mod tidy
```

## Common Tasks

### Start Development
```bash
# Terminal 1: Run daemon
cd cli && go run ./cmd/goconnect

# Terminal 2: Run desktop app
cd desktop && npm run tauri dev
```

### Database Migrations
```bash
# Create new migration
# Add files to core/migrations/ or core/migrations_sqlite/
```

### Linting
```bash
cd core && golangci-lint run
cd cli && golangci-lint run
```

