# Contributing to GoConnect

We welcome contributions to GoConnect! This guide will help you get started.

## Development Setup

### Prerequisites
- Go 1.24+
- Node.js 18+ and npm
- PostgreSQL 15+ (optional, uses SQLite for development)

### Quick Start
```bash
# Clone the repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Start the core backend
cd core
go run ./cmd/server

# Start the CLI (in another terminal)
cd cli
go run ./cmd/goconnect

# Start the desktop app (in another terminal)
cd desktop
npm install
npm run tauri dev
```

## Project Structure

```
goconnect/
├── desktop/               # Tauri desktop app
│   ├── src/              # React frontend
│   ├── src-tauri/        # Rust backend
│   └── package.json
├── cli/                   # Go CLI application
│   ├── cmd/goconnect/    # CLI entry point
│   └── internal/         # Private code
├── core/                  # Go backend/library
│   ├── cmd/server/       # Server entry point
│   ├── internal/         # Business logic
│   ├── migrations/       # Database migrations
│   └── openapi/          # API specification
└── docs/                 # Documentation
```

## Code Style

### Go
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `golangci-lint` for linting

### TypeScript/React
- Use Prettier for formatting
- Follow ESLint rules
- Use TypeScript strictly

## Testing

### Running Tests
```bash
# Go tests
cd core && go test ./...
cd cli && go test ./...

# Or use Makefile
make test
```

### Coverage
- Go: Target 80%+ coverage
- Frontend: Use Jest for unit tests

## Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the test suite
6. Submit a pull request

## Pull Request Process

- Describe your changes clearly
- Include tests for new functionality
- Ensure all tests pass
- Update documentation if needed

## Issues

- Use GitHub Issues for bug reports
- Provide detailed reproduction steps
- Include environment information

## License

By contributing, you agree to license your contributions under the MIT License.
