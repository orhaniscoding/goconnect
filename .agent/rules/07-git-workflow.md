# Git Workflow & CI/CD

Standards for version control and continuous integration.

## Commit Convention

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types
| Type | Usage |
|------|-------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no code change |
| `refactor` | Code change that neither fixes bug nor adds feature |
| `perf` | Performance improvement |
| `test` | Adding or fixing tests |
| `chore` | Build, CI, tools, etc. |

### Scope
| Scope | Usage |
|-------|-------|
| `core` | Backend server |
| `cli` | CLI/daemon |
| `desktop` | Desktop app |
| `proto` | Protocol buffers |
| `ci` | GitHub Actions |
| `deps` | Dependencies |

### Examples
```bash
# Feature
feat(cli): add network join command with TUI progress

# Bug fix
fix(core): prevent SQL injection in user search

# Documentation
docs: update installation instructions for macOS

# Breaking change (use ! or footer)
feat(api)!: change network response format

# Multi-line with body
fix(cli): resolve connection timeout issue

The daemon was not properly handling slow network connections.
Increased timeout from 5s to 30s and added retry logic.

Fixes #123
```

---

## Branch Strategy

### Branch Naming
```
main              # Production-ready code
develop           # Integration branch (optional)
feature/<name>    # New features
fix/<name>        # Bug fixes
docs/<name>       # Documentation updates
refactor/<name>   # Code refactoring
```

### Examples
```bash
feature/network-discovery
fix/websocket-reconnection
docs/api-reference
refactor/service-layer
```

---

## Pull Request Guidelines

### PR Title
Follow commit convention: `feat(core): add user authentication`

### PR Description Template
```markdown
## Summary
Brief description of what this PR does.

## Changes
- Added X functionality
- Fixed Y bug
- Updated Z documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests (if applicable)
- [ ] Manual testing performed

## Screenshots (if UI changes)
[Add screenshots here]

## Related Issues
Closes #123
```

### PR Requirements
- [ ] All tests pass
- [ ] No lint errors
- [ ] Code reviewed by at least 1 person
- [ ] Conventional commit title
- [ ] Documentation updated (if needed)

---

## CI/CD Pipeline

### GitHub Actions Workflows

Located in `.github/workflows/`:

```yaml
# test.yml - Runs on every PR
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Test Core
        run: cd core && go test ./... -race -coverprofile=coverage.out
      - name: Test CLI
        run: cd cli && go test ./... -race -coverprofile=coverage.out
      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: core
```

### Pre-commit Checklist
Before pushing, run locally:

```bash
# Format code
cd core && go fmt ./...
cd cli && go fmt ./...

# Run linters
cd core && golangci-lint run
cd cli && golangci-lint run

# Run tests
make test

# Update dependencies
cd core && go mod tidy
cd cli && go mod tidy
```

---

## Dependency Management

### Adding Dependencies
```bash
# Add dependency
cd core
go get github.com/new/package@v1.2.3

# Always tidy after changes
go mod tidy
```

### Updating Dependencies
```bash
# Update all
go get -u ./...
go mod tidy

# Update specific package
go get -u github.com/package/name
go mod tidy
```

### Security Updates
```bash
# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Release Process

### Version Tags
Use semantic versioning: `v1.2.3`

```bash
# Create release tag
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

### Changelog
Update `CHANGELOG.md` with each release:

```markdown
## [1.2.0] - 2024-01-15

### Added
- Network discovery feature (#45)
- File transfer progress (#52)

### Fixed
- WebSocket reconnection issue (#48)
- Memory leak in daemon (#51)

### Changed
- Upgraded WireGuard library to v0.5.0
```
