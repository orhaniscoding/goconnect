# Contributing to GoConnect

We welcome contributions to GoConnect! This guide will help you get started.

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** - For CLI and Server
- **Node.js 20+** - For Desktop App
- **Rust** - For Desktop App (Tauri backend)
- **protoc** - Protocol Buffers compiler

### Development Setup

```bash
# Clone the repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Start the server (in one terminal)
cd core
cp config.example.env .env
go run ./cmd/server

# Start the CLI (in another terminal)
cd cli
go run ./cmd/goconnect

# Start the desktop app (in another terminal)
cd desktop
npm install
npm run tauri dev
```

## 📁 Project Structure

```
goconnect/
├── desktop/               # Tauri desktop application
│   ├── src/               # React frontend (TypeScript)
│   │   ├── components/    # React components
│   │   ├── lib/           # Utilities and hooks
│   │   └── App.tsx        # Main app component
│   ├── src-tauri/         # Rust backend
│   │   ├── src/           # Rust source
│   │   └── Cargo.toml     # Rust dependencies
│   └── package.json       # Node dependencies
│
├── cli/                   # Terminal application (Go)
│   ├── cmd/goconnect/     # CLI entry point
│   └── internal/          # Private packages
│       ├── tui/           # Terminal UI (Bubbletea)
│       ├── daemon/        # Background service & IPC
│       ├── chat/          # Chat functionality
│       ├── transfer/      # File transfer
│       ├── p2p/           # Peer-to-peer networking
│       └── wireguard/     # WireGuard integration
│
├── core/                  # Server backend (Go)
│   ├── cmd/server/        # Server entry point
│   ├── internal/          # Business logic
│   │   ├── handler/       # HTTP handlers
│   │   ├── service/       # Business services
│   │   ├── repository/    # Database layer
│   │   ├── websocket/     # Real-time communication
│   │   └── wireguard/     # WireGuard management
│   ├── migrations/        # Database migrations
│   └── openapi/           # API specification
│
├── docs/                  # Documentation
├── .github/workflows/     # CI/CD pipelines
└── Makefile               # Build automation
```

## 🔧 Development Workflow

### Running Tests

```bash
# Run all Go tests
make test

# Run CLI tests only
cd cli && go test ./...

# Run Server tests only
cd core && go test ./...

# Run with coverage
cd cli && go test -cover ./...
```

### Linting

```bash
# Run Go linter
make lint

# Or manually
golangci-lint run ./...
```

### Building

```bash
# Build CLI
cd cli && go build -o goconnect ./cmd/goconnect

# Build Server
cd core && go build -o goconnect-server ./cmd/server

# Build Desktop App
cd desktop && npm run tauri build
```

## 📝 Code Style

### Go Code

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write tests for new functionality

```go
// Good: Clear, documented function
// CreateNetwork creates a new virtual network with the given name.
func (s *Service) CreateNetwork(ctx context.Context, name string) (*Network, error) {
    // Implementation
}
```

### TypeScript/React Code

- Use TypeScript strictly (no `any` types)
- Follow ESLint rules
- Use Prettier for formatting
- Prefer functional components with hooks

```typescript
// Good: Typed, functional component
interface NetworkCardProps {
  network: Network;
  onConnect: (id: string) => void;
}

export function NetworkCard({ network, onConnect }: NetworkCardProps) {
  return (
    // JSX
  );
}
```

### Rust Code

- Follow Rust conventions
- Use `cargo fmt` for formatting
- Run `cargo clippy` for linting

## 🔀 Git Workflow

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation
- `refactor/description` - Code refactoring

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add network creation wizard
fix: resolve connection timeout issue
docs: update installation guide
refactor: simplify daemon IPC logic
test: add chat service unit tests
chore: update dependencies
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add tests for new functionality
5. Run tests: `make test`
6. Run linter: `make lint`
7. Commit with conventional commit message
8. Push to your fork
9. Open a Pull Request

### PR Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] Conventional commit message used
- [ ] PR description explains the changes

## 🧪 Testing Guidelines

### Unit Tests

```go
func TestCreateNetwork(t *testing.T) {
    // Arrange
    svc := NewService(mockRepo)
    
    // Act
    network, err := svc.CreateNetwork(ctx, "Test Network")
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, "Test Network", network.Name)
}
```

### Integration Tests

Place integration tests in `*_integration_test.go` files with build tag:

```go
//go:build integration

package service_test

func TestNetworkIntegration(t *testing.T) {
    // Test with real database
}
```

### Coverage Goals

- **Core packages**: 80%+ coverage
- **Handlers**: 70%+ coverage
- **Utilities**: 90%+ coverage

## 📚 Documentation

- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update OpenAPI spec for API changes
- Add examples for new features

## 🐛 Reporting Issues

### Bug Reports

Include:
1. GoConnect version
2. Operating system
3. Steps to reproduce
4. Expected vs actual behavior
5. Logs (if applicable)

### Feature Requests

Include:
1. Problem description
2. Proposed solution
3. Alternatives considered
4. Use cases

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

<div align="center">

Thank you for contributing to GoConnect! 🎉

</div>
