# Project Structure

GoConnect is a cross-platform networking solution that creates secure virtual LANs over the internet using WireGuard encryption and peer-to-peer connections.

## Architecture Overview

```
goconnect/
├── core/           # Backend server (Go)
├── cli/            # CLI daemon (Go)
├── desktop/        # Desktop app (Tauri + React)
├── docs/           # Documentation
├── .github/        # CI/CD workflows
└── go.work         # Go workspace config
```

## Module Responsibilities

### core/ - Backend Server
**Module:** `github.com/orhaniscoding/goconnect/server`

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Application entry point |
| `internal/config/` | Configuration loading and validation |
| `internal/domain/` | Domain models and business entities |
| `internal/handler/` | HTTP and WebSocket handlers |
| `internal/repository/` | Database access layer |
| `internal/service/` | Business logic layer |
| `internal/websocket/` | Real-time communication |
| `migrations/` | PostgreSQL migrations |
| `migrations_sqlite/` | SQLite migrations |
| `proto/` | gRPC protocol definitions |
| `openapi/` | OpenAPI specifications |

**Tech Stack:** Go 1.24+, Gin, PostgreSQL, SQLite, gRPC

---

### cli/ - CLI Daemon
**Module:** `github.com/orhaniscoding/goconnect/cli`

| Directory | Purpose |
|-----------|---------|
| `cmd/goconnect/` | CLI entry point |
| `cmd/daemon/` | Background daemon entry |
| `internal/api/` | gRPC client for server communication |
| `internal/chat/` | Peer-to-peer chat functionality |
| `internal/config/` | Local configuration management |
| `internal/daemon/` | System daemon logic |
| `internal/engine/` | Core networking engine |
| `internal/identity/` | User identity and keys |
| `internal/p2p/` | Peer-to-peer connections |
| `internal/proto/` | Generated protobuf code |
| `internal/storage/` | Local database (SQLite) |
| `internal/system/` | OS-specific operations |
| `internal/transfer/` | File transfer functionality |
| `internal/tui/` | Terminal UI (Bubbletea) |
| `internal/wireguard/` | WireGuard interface management |
| `service/` | Platform-specific service configs |

**Tech Stack:** Go 1.24+, Bubbletea, gRPC, WireGuard (golang.zx2c4.com/wireguard)

---

### desktop/ - Desktop Application

| Directory | Purpose |
|-----------|---------|
| `src/` | React components and logic |
| `src-tauri/` | Rust backend (Tauri) |
| `public/` | Static assets |

**Tech Stack:** Tauri 2.x, React, TypeScript, Tailwind CSS, Vite

---

## Go Workspace

This project uses Go workspaces (`go.work`):

```go
go 1.24

use (
    ./cli
    ./core
)
```

**Rules:**
- Run `go mod tidy` in each module after adding dependencies
- Proto files live in `core/proto/`, generated code in `cli/internal/proto/`
- Never import `core` packages from `cli` or vice versa at runtime
- Shared types should be in proto definitions

---

## Directory Conventions

### internal/ Usage
The `internal/` directory is a Go toolchain feature that prevents external imports:

```
✅ core/internal/service/  → Only importable within core/
✅ cli/internal/daemon/    → Only importable within cli/
❌ External packages cannot import internal code
```

### Package Naming
- Use singular nouns: `service`, `handler`, `repository`
- Avoid generic names: prefer `userservice` over `utils`
- Keep packages focused on single responsibility

### File Naming
```
✅ user_handler.go       # snake_case for Go files
✅ UserCard.tsx          # PascalCase for React components  
✅ user_handler_test.go  # Test files end with _test.go
```

---

## Dependency Flow

```
Handler/API Layer → Service Layer → Repository Layer → Database
                           ↓
                   External Services
```

**Rules:**
- Handlers call Services (never Repositories directly)
- Services call Repositories and external APIs
- Repositories only interact with database
- Use interfaces for dependency injection
