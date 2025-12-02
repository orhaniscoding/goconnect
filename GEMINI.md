# GoConnect - Cursor AI Rules

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

## Security Guidelines
- See `.agent/rules/security-*.md` for detailed security rules
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