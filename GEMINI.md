# GoConnect - AI Agent Rules

> [!CAUTION]
> **MANDATORY COMPLIANCE**: All rules in this file and the referenced `.agent/rules/` files are **REQUIRED**. The AI Agent MUST follow these guidelines without exception. Violations will result in inconsistent code, security vulnerabilities, and broken builds.

> [!IMPORTANT]
> Before generating ANY code, the AI Agent MUST:
> 1. Read and understand the relevant rule files in `.agent/rules/`
> 2. Follow the coding standards for the target language (Go, TypeScript, or Rust)
> 3. Include proper error handling, tests, and documentation
> 4. Never skip security validations or log sensitive data

## Project Overview
GoConnect is a cross-platform networking solution that creates secure virtual LANs over the internet using WireGuard encryption and peer-to-peer connections.

## Architecture

### Modules
- **core/** - Go backend server (module: `github.com/orhaniscoding/goconnect/server`)
- **cli/** - Go CLI/daemon (module: `github.com/orhaniscoding/goconnect/cli`)
- **desktop/** - Tauri + React desktop app

### Tech Stack
| Component | Technologies |
|-----------|-------------|
| Backend | Go 1.24+, Gin, PostgreSQL, SQLite |
| CLI/Daemon | Go 1.24+, Bubbletea, gRPC, WireGuard |
| Desktop | Tauri 2.x, React, TypeScript, Tailwind |
| IPC | gRPC over Unix sockets / Named Pipes |

## Code Style

### Go
```go
// Exported: PascalCase
// Internal: camelCase
// Error wrapping: always add context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

### TypeScript/React
```typescript
// Functional components with types
interface Props {
  title: string;
}
export function Component({ title }: Props) {
  return <div>{title}</div>;
}
```

### Rust
```rust
// snake_case for functions, PascalCase for types
// Always handle Result with ?
fn process(data: &str) -> Result<Output, Error> {
    let parsed = parse(data)?;
    Ok(transform(parsed))
}
```

## Detailed Guidelines

See `.agent/rules/` for comprehensive instructions:

| File | Content |
|------|---------|
| [01-project-structure.md](.agent/rules/01-project-structure.md) | Architecture, modules, directory conventions |
| [02-code-standards.md](.agent/rules/02-code-standards.md) | Go/TS/Rust naming, formatting, linting |
| [03-testing.md](.agent/rules/03-testing.md) | Unit/integration tests, coverage goals |
| [04-security.md](.agent/rules/04-security.md) | WireGuard, secrets, input validation |
| [05-error-handling.md](.agent/rules/05-error-handling.md) | Error wrapping, logging, user messages |
| [06-api-protocols.md](.agent/rules/06-api-protocols.md) | REST, gRPC, WebSocket, IPC |
| [07-git-workflow.md](.agent/rules/07-git-workflow.md) | Commits, branches, CI/CD |
| [08-documentation.md](.agent/rules/08-documentation.md) | Comments, GoDoc, changelog |
| [09-mcp-usage.md](.agent/rules/09-mcp-usage.md) | Guidelines for using MCP servers |

## Security Quick Reference
- Never log secrets, tokens, or personal data
- Always use parameterized queries
- Validate all external input
- Use TLS 1.2+ for all connections

## Commit Convention
Use Conventional Commits:
```
feat(cli): add network join command
fix(core): prevent SQL injection
docs: update README
```

## Testing
```bash
make test           # Run all tests
cd core && go test ./... -short
cd cli && go test ./... -short
```

## Important Notes
- This is a Go workspace (see `go.work`)
- Proto files in `core/proto/`, generated code in `cli/internal/proto/`
- Run `go mod tidy` after adding dependencies