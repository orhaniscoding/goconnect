# Contributing to GoConnect

Thank you for your interest in contributing to GoConnect! This document provides guidelines and workflows for contributing to the project.

## 🚀 Quick Start

### Prerequisites
- **Go** 1.24 or higher
- **Node.js** 18+ and npm
- **Git** with commit signing configured
- **golangci-lint** (optional but recommended)

### Setup Development Environment

1. Clone the repository:
```bash
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect
```

2. Install development tools:
```bash
make install-tools
```

3. Verify setup:
```bash
make status
```

## 🔧 Development Workflow

### 1. Work on Main Branch (No Feature Branches)

This project follows a **strictly no branching** policy. All development occurs directly on `main`:

```bash
# Stay on main
git checkout main

# Keep up to date
git pull --rebase
```

**Commit message prefixes (Conventional Commits):**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code refactoring
- `test:` - Test additions
- `chore:` - Build process, tooling

### 2. Make Your Changes

Follow the project architecture:
- **Server (Go)**: handler → service → repository pattern
- **Web UI (Next.js)**: App Router structure with i18n
- **Client Daemon (Go)**: Minimal, platform-agnostic core

### 3. Run Local Checks

#### Quick Test
```bash
# Test all components
make test

# Test with race detector
make test-race
```

#### Full CI Pipeline (Recommended)
```bash
# Run all checks like CI does
make ci
```

This runs:
- `go vet` - Static analysis
- `go test -race` - Race condition detection
- Coverage check (≥60% threshold)
- `golangci-lint` - Comprehensive linting

#### Component-Specific Commands
```bash
# Server
cd server
make help              # See all commands
make test              # Run tests
make test-coverage     # Run tests with coverage report
make lint              # Run linters
make build             # Build binary

# Client Daemon
cd client-daemon
make test-coverage
make build-all         # Build for all platforms

# Web UI
cd web-ui
npm run typecheck
npm run build
```

### 4. Update Documentation

If your changes affect:
- **API contracts**: Update `server/openapi/openapi.yaml`
- **Architecture**: Update `docs/TECH_SPEC.md`
- **User-facing features**: Update `README.md`
- **Developer workflow**: Update this file

### 5. Write Tests

Testing requirements:
- **Minimum coverage**: 60% (enforced in CI)
- **Happy path**: At least one success case
- **Edge cases**: Error handling, validation, boundaries
- **Race conditions**: Use `-race` flag for concurrent code
- **Table-driven tests**: Preferred for multiple scenarios

Example test structure:
```go
func TestFeature(t *testing.T) {
	tests := []struct {
		name    string
		input   Input
		want    Output
		wantErr bool
	}{
		{
			name:  "Success case",
			input: validInput,
			want:  expectedOutput,
		},
		{
			name:    "Invalid input",
			input:   invalidInput,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Feature(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
```

### 6. Commit Your Changes

We use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .
git commit -S -m "feat(server): add IP allocation endpoint"
```

Commit message format:
```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `security`: Security fixes

**Scopes:**
- `server`: Backend changes
- `daemon`: Client daemon changes
- `web-ui`: Frontend changes
- `tests`: Test-specific changes
- `docs`: Documentation
- `deps`: Dependency updates

**Examples:**
```bash
feat(server): add rate limiting to network endpoints
fix(auth): resolve token expiration edge case
docs(api): update OpenAPI spec for chat endpoints
refactor(service): extract membership logic to separate service
test(handler): add comprehensive RBAC test coverage
```

### 7. Push and Submit

1. Push your changes:
```bash
git push origin main
```

2. For external contributors (fork workflow):
   - Push to your fork's main branch
   - Create PR on GitHub with:
     - Clear title following conventional commits
     - Description of changes
     - Link to related issues
     - Screenshots (if UI changes)

## ✅ PR Checklist

Before submitting, ensure:

### Tests & Quality
- [ ] All tests pass: `make test-race`
- [ ] Coverage maintained: `make test-coverage`
- [ ] Linters clean: `make lint`
- [ ] Go vet passes: `make vet`
- [ ] No race conditions detected

### Code Standards
- [ ] Follows project architecture (handler → service → repo)
- [ ] Error handling uses domain errors (`domain.NewError`)
- [ ] Idempotency-Key required for all mutations
- [ ] RBAC checks return `ERR_NOT_AUTHORIZED` consistently
- [ ] Audit events added for important actions

### Documentation
- [ ] OpenAPI spec updated (if API changed)
- [ ] `docs/TECH_SPEC.md` updated (if architecture changed)
- [ ] Code comments added for complex logic
- [ ] README updated (if user-facing changes)

### Security
- [ ] No secrets or credentials in code
- [ ] No PII in logs or audit events (use hashes/redaction)
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (when DB implemented)

### Database (when applicable)
- [ ] Migrations created (up and down)
- [ ] Migration tested
- [ ] Rollback tested

## 🏗️ Architecture Guidelines

### Server (Go)

**Layered Architecture:**
```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

- **Handlers**: HTTP request/response, validation, auth middleware
- **Services**: Business rules, orchestration, audit logging
- **Repositories**: Data persistence, in-memory or PostgreSQL

**Error Handling:**
```go
// Use domain errors
return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)

// In handlers
if domainErr, ok := err.(*domain.Error); ok {
    errorResponse(c, domainErr)
    return
}
```

**Idempotency:**
```go
// All mutations require Idempotency-Key header
idempotencyKey := c.GetHeader("Idempotency-Key")
if idempotencyKey == "" {
    return domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key required", nil)
}
```

### Web UI (Next.js)

**Structure:**
```
src/app/[locale]/
  (public)/   - Unauthenticated pages
  (protected)/ - Authenticated pages
```

**i18n:**
```typescript
import { useTranslation } from '@/lib/i18n-context'

const { t } = useTranslation()
return <h1>{t('dashboard.title')}</h1>
```

## 🧪 Testing Guidance

### Run Specific Tests
```bash
# Server - specific package
cd server
go test ./internal/handler -v -run TestAuthHandler

# Server - with coverage
go test ./internal/service -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Client Daemon
cd client-daemon
go test ./internal/smoke -v
```

### Race Detection
```bash
# Critical for concurrent code
make test-race

# Or specific package
go test ./internal/service -race -count=100
```

### Coverage Reports
```bash
# Generate HTML report
make -C server test-coverage-html
open server/coverage.html  # or xdg-open on Linux
```

## 🔐 Security Guidelines

### Authentication & Authorization
- Current auth is **PLACEHOLDER ONLY** - not production-ready
- Token validation always returns admin in dev mode
- Real JWT/OIDC implementation required before production

### Sensitive Data
- Never log passwords, tokens, or PII
- Use audit-safe hashing for user IDs in logs
- Redact email addresses in public logs

### Dependencies
- Evaluate license compatibility (MIT preferred)
- Check for known vulnerabilities
- Minimize dependency count

## 📦 Release Process

Releases are automated via [Release Please](https://github.com/googleapis/release-please):

1. **Merge to main**: Commits trigger release-please bot
2. **Version PR**: Bot creates/updates PR with changelog
3. **Merge version PR**: Creates git tag
4. **GoReleaser**: Tag triggers binary builds and GitHub release

Version bumping:
- `feat:` → Minor version (1.1.0 → 1.2.0)
- `fix:` → Patch version (1.1.0 → 1.1.1)
- `feat!:` or `BREAKING CHANGE:` → Major version (1.1.0 → 2.0.0)

## 🎨 Code Style

### Go
- Use `gofmt` (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Short receiver names: `(s *Service)` not `(service *Service)`
- Explicit error wrapping: `fmt.Errorf("failed to create: %w", err)`

### TypeScript
- Strict mode enabled
- No implicit `any`
- Prefer interfaces over types for object shapes
- Use functional components with hooks

### General
- Comments explain **why**, not **what**
- Keep functions small and focused
- Avoid premature optimization
- Write self-documenting code

## 🐛 Reporting Issues

### Bug Reports
Include:
- Go version: `go version`
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

### Feature Requests
Include:
- Use case description
- Proposed solution
- Alternative solutions considered
- Impact on existing features

## 💬 Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/orhaniscoding/goconnect/discussions)
- **Bugs**: Create an [Issue](https://github.com/orhaniscoding/goconnect/issues)
- **Security**: Email security concerns (don't open public issues)

## 📚 Additional Resources

- [Technical Specification](docs/TECH_SPEC.md)
- [API Examples](docs/API_EXAMPLES.http)
- [Security Policy](docs/SECURITY.md)
- [Architecture Overview](README.md#architecture)

## 🙏 Thank You!

Your contributions make GoConnect better for everyone. We appreciate your time and effort!

