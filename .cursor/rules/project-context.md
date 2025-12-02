# Project Context

## What is GoConnect?
GoConnect is a cross-platform application that creates secure virtual LANs over the internet, enabling devices to communicate as if they were on the same local network.

## Tech Stack

### Backend (core/)
- **Language**: Go 1.24+
- **Database**: PostgreSQL (production), SQLite (development)
- **Framework**: Gin (HTTP), gRPC (IPC)
- **Auth**: JWT + 2FA (TOTP)

### CLI/Daemon (cli/)
- **Language**: Go 1.24+
- **TUI**: Bubbletea + Lipgloss
- **IPC**: gRPC over Unix sockets (Linux/macOS) or Named Pipes (Windows)
- **VPN**: WireGuard (userspace)

### Desktop (desktop/)
- **Framework**: Tauri 2.x (Rust)
- **Frontend**: React + TypeScript + Vite
- **Styling**: Tailwind CSS
- **State**: TanStack Query

## Module Structure

```
goconnect/
├── go.work              # Go workspace file
├── core/                # Backend server
│   ├── go.mod           # module: github.com/orhaniscoding/goconnect/server
│   ├── cmd/daemon/      # Daemon entry point
│   ├── internal/        # Business logic
│   └── proto/           # Protobuf definitions
├── cli/                 # CLI/Daemon client
│   ├── go.mod           # module: github.com/orhaniscoding/goconnect/cli
│   ├── cmd/goconnect/   # CLI entry point
│   └── internal/        # Client logic
└── desktop/             # Desktop app
    ├── src/             # React frontend
    └── src-tauri/       # Rust backend
```

## Key Dependencies
- `github.com/kardianos/service` - Cross-platform service management
- `golang.zx2c4.com/wireguard` - WireGuard implementation
- `google.golang.org/grpc` - gRPC framework
- `github.com/charmbracelet/bubbletea` - TUI framework

