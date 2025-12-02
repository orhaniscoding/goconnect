# 💻 GoConnect CLI

GoConnect's terminal application with an interactive TUI interface. Create networks or join existing ones from the command line.

## ✨ Features

- 🖥️ **Interactive TUI** - Modern terminal interface with Bubbletea
- 🌐 **Create Network** - Create and manage networks from terminal
- 🔗 **Join Network** - Connect with invite link
- 💬 **Chat** - Full chat functionality in terminal
- 📁 **File Transfer** - P2P file sharing between peers
- 📊 **Status Dashboard** - Connection status, members, IP addresses
- 🔧 **Daemon Mode** - Run as background service

## 🚀 Quick Start

### Download

```bash
# Linux (x64)
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_*_linux_amd64.tar.gz
tar -xzf goconnect_*_linux_amd64.tar.gz
sudo mv goconnect /usr/local/bin/

# macOS (Apple Silicon)
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_*_darwin_arm64.tar.gz
tar -xzf goconnect_*_darwin_arm64.tar.gz
sudo mv goconnect /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_*_windows_amd64.zip" -OutFile "goconnect.zip"
Expand-Archive -Path "goconnect.zip" -DestinationPath "."
.\goconnect.exe
```

### Usage

```bash
# Interactive mode (TUI)
goconnect

# Quick commands
goconnect create "Network Name"  # Create network
goconnect join <link>            # Join network
goconnect list                   # List networks
goconnect status                 # Connection status
goconnect chat                   # Open chat
goconnect disconnect             # Disconnect
```

## 🎨 TUI Interface

```
┌──────────────────────────────────────────────────────────────┐
│                    🔗 GoConnect CLI v3.0.0                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   ? What would you like to do?                               │
│                                                              │
│   ❯ 🌐 Create Network                                        │
│     🔗 Join Network                                          │
│     📋 My Networks                                           │
│     💬 Chat                                                  │
│     📁 File Transfer                                         │
│     ⚙️  Settings                                              │
│     ❌ Exit                                                   │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│   ↑/↓: select  •  Enter: confirm  •  q: quit                │
└──────────────────────────────────────────────────────────────┘
```

## 🏗️ Architecture

```
cli/
├── cmd/
│   └── goconnect/
│       └── main.go           # Entry point
├── internal/
│   ├── tui/                  # Terminal UI (Bubbletea)
│   │   ├── model.go          # TUI model
│   │   ├── views.go          # Screens
│   │   └── styles.go         # Lipgloss styles
│   ├── daemon/               # Background service
│   │   ├── server.go         # gRPC server
│   │   ├── ipc_unix.go       # Unix socket IPC
│   │   └── ipc_windows.go    # Named Pipes IPC
│   ├── chat/                 # Chat functionality
│   │   ├── manager.go        # Chat manager
│   │   └── storage.go        # SQLite persistence
│   ├── transfer/             # File transfer
│   │   ├── manager.go        # Transfer manager
│   │   └── types.go          # Transfer types
│   ├── p2p/                  # Peer-to-peer networking
│   ├── wireguard/            # WireGuard integration
│   └── config/               # Configuration
└── go.mod
```

## 🛠️ Development

### Requirements

- Go 1.24+
- WireGuard tools (`wg`, `wg-quick`)
- protoc (Protocol Buffers compiler)

### Build

```bash
# Development build
go build -o goconnect ./cmd/goconnect

# Production build with version
VERSION=v3.0.0
go build -ldflags="-s -w -X main.version=${VERSION}" -o goconnect ./cmd/goconnect

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o goconnect-linux ./cmd/goconnect
GOOS=darwin GOARCH=arm64 go build -o goconnect-macos ./cmd/goconnect
GOOS=windows GOARCH=amd64 go build -o goconnect.exe ./cmd/goconnect
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/daemon/...
go test ./internal/chat/...
```

## ⚙️ Configuration

Configuration file location:
- **Linux/macOS**: `~/.config/goconnect/config.yaml`
- **Windows**: `%APPDATA%\goconnect\config.yaml`

```yaml
# config.yaml
server:
  url: "https://api.goconnect.io"
  
daemon:
  socket_path: "/tmp/goconnect.sock"  # Unix
  pipe_name: "goconnect"              # Windows

logging:
  level: "info"
  file: "~/.config/goconnect/goconnect.log"
```

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
