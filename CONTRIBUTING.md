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

# Start the server
cd server
go run ./cmd/server

# Start the client daemon
cd ../client-daemon
go run ./cmd/daemon

# Start the web UI
cd ../web-ui
npm install
npm run dev
```

## Project Structure

```
goconnect/
├── server/                 # Go backend
│   ├── cmd/               # Application entry points
│   ├── internal/          # Private application code
│   │   ├── handler/       # HTTP handlers
│   │   ├── service/       # Business logic
│   │   ├── repository/    # Data access layer
│   │   └── domain/        # Domain models
│   ├── migrations/        # Database migrations
│   └── openapi/          # API specification
├── client-daemon/         # Go client agent
│   ├── cmd/              # Daemon entry point
│   ├── internal/         # Private application code
│   └── service/          # Client services
├── web-ui/               # Next.js frontend
│   ├── src/              # React components
│   ├── pages/            # Next.js pages
│   └── public/           # Static assets
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
cd server && go test ./...
cd client-daemon && go test ./...

# Frontend tests
cd web-ui && npm test
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
