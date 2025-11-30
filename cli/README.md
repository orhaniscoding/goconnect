# 💻 GoConnect CLI

GoConnect's terminal application. Create networks or join existing ones with an interactive TUI interface.

## ✨ Features

- 🖥️ **Interactive TUI** - Modern terminal interface with Bubbletea
- 🌐 **Create Network** - Create and manage networks from terminal
- 🔗 **Join Network** - Connect with invite link
- 📊 **View Status** - Connection status, members, IP addresses
- 🔧 **Headless Mode** - Run in background on servers

## 🚀 Quick Start

### Download

```bash
# Linux
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-linux-amd64
chmod +x goconnect-cli-linux-amd64
sudo mv goconnect-cli-linux-amd64 /usr/local/bin/goconnect

# macOS
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-darwin-arm64
chmod +x goconnect-cli-darwin-arm64
sudo mv goconnect-cli-darwin-arm64 /usr/local/bin/goconnect
```

### Usage

```bash
# Interactive mode
goconnect

# Quick commands
goconnect create "Network Name"  # Create network
goconnect join <link>            # Join network
goconnect list                   # List networks
goconnect status                 # Connection status
goconnect disconnect             # Disconnect
```

## 🎨 TUI Interface

```
┌──────────────────────────────────────────────────────────────┐
│                    🔗 GoConnect CLI                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   ? What would you like to do?                               │
│                                                              │
│   ❯ 🌐 Create Network                                        │
│     🔗 Join Network                                          │
│     📋 My Networks                                           │
│     ⚙️  Settings                                              │
│     ❌ Exit                                                   │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│   ↑/↓: select  •  Enter: confirm  •  q: quit                │
└──────────────────────────────────────────────────────────────┘
```

## 🛠️ Development

### Requirements

- Go 1.24+
- WireGuard tools (`wg`, `wg-quick`)

### Build

```bash
# Single platform
go build -o goconnect ./cmd/daemon

# All platforms
make build-all
```

### Project Structure

```
cli/
├── cmd/
│   └── daemon/
│       └── main.go         # Entry point
├── internal/
│   ├── tui/                # Terminal UI
│   │   ├── model.go        # TUI model
│   │   ├── views.go        # Screens
│   │   └── styles.go       # Styles
│   ├── network/            # Network management
│   ├── wireguard/          # WireGuard integration
│   └── config/             # Configuration
└── go.mod
```

## ⚙️ Configuration

Configuration file locations:
- **Linux:** `~/.config/goconnect/config.yaml`
- **macOS:** `~/Library/Application Support/GoConnect/config.yaml`
- **Windows:** `%APPDATA%\GoConnect\config.yaml`

### Example Configuration

```yaml
# GoConnect CLI Configuration
server:
  url: ""  # Empty = create new network

wireguard:
  interface_name: goconnect0

daemon:
  local_port: 12345
  health_check_interval: 30s
```

## 🔧 System Service

### Linux (systemd)

```bash
sudo ./goconnect install
sudo systemctl enable goconnect
sudo systemctl start goconnect
```

### macOS (launchd)

```bash
sudo ./goconnect install
```

### Windows (Windows Service)

```powershell
# Run as Administrator
.\goconnect.exe install
```

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
