# 🔗 GoConnect

> **"Virtual LAN made simple."**

GoConnect is a user-friendly virtual LAN platform that makes devices on the internet appear as if they're on the same local network.

[![Release](https://img.shields.io/github/v/release/orhaniscoding/goconnect?style=flat-square)](https://github.com/orhaniscoding/goconnect/releases)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev)

---

## 📖 Table of Contents

- [What is GoConnect?](#-what-is-goconnect)
- [Who is it for?](#-who-is-it-for)
- [How it Works](#-how-it-works)
- [Installation](#-installation)
- [Usage](#-usage)
- [Features](#-features)
- [Architecture](#-architecture)
- [Development](#-development)
- [FAQ](#-faq)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🤔 What is GoConnect?

GoConnect is a **single application** that lets you:

- 🌐 **Create a network** - Host your own private LAN party
- 🔗 **Join networks** - Connect with one click via invite link
- 💬 **Chat** - Modern text channels
- 🎮 **Play games** - LAN games over the internet

### What Makes GoConnect Different?

| Traditional VPN | GoConnect |
|-----------------|-----------|
| Complex setup | **One-click setup** |
| Central server bottleneck | **Peer-to-peer** |
| Technical knowledge required | **User-friendly** |
| Single network | **Multiple networks** |
| No built-in chat | **Integrated chat** |

---

## 👥 Who is it for?

### 🎮 Gamers
- Share Minecraft LAN worlds with friends
- Play old LAN games over the internet
- Low-latency gaming experience

### 💼 Remote Workers
- Secure access to office resources
- Team file sharing
- Simple VPN alternative

### 🏠 Home Users
- Access home devices from anywhere
- Secure family file sharing
- Remote NAS connection

### 👨‍💻 Developers
- Create test environments
- Microservice communication
- Container networking

---

## ⚙️ How it Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        GoConnect App                             │
│                                                                  │
│  ┌──────────────────┐          ┌──────────────────┐             │
│  │  Create Network  │          │   Join Network   │             │
│  │       🌐         │          │       🔗         │             │
│  │                  │          │                  │             │
│  │ Start your own   │          │ Join someone's   │             │
│  │ server and       │          │ network with     │             │
│  │ invite friends   │          │ invite link      │             │
│  └────────┬─────────┘          └────────┬─────────┘             │
│           │                             │                        │
│           ▼                             ▼                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              WireGuard Secure Tunnel                     │    │
│  │         (Automatic configuration - you don't             │    │
│  │          need to do anything!)                           │    │
│  └─────────────────────────────────────────────────────────┘    │
│           │                             │                        │
│           ▼                             ▼                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Virtual Local Network                  │   │
│  │                                                           │   │
│  │   👤 You          👤 Friend 1       👤 Friend 2          │   │
│  │   10.0.1.1        10.0.1.2          10.0.1.3             │   │
│  │                                                           │   │
│  │   Now you're all on the same LAN!                        │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Simple Steps

1. **Download** → Get the GoConnect app
2. **Open** → Run the application
3. **Choose** → "Create Network" or "Join Network"
4. **Connect** → One click to connect!

---

## 📥 Installation

### Option 1: Desktop Application (Recommended)

The easiest way! Do everything with a single app.

| Platform | Download |
|----------|----------|
| **Windows** | [GoConnect-Windows.exe](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **macOS (Intel)** | [GoConnect-macOS-Intel.dmg](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **macOS (Apple Silicon)** | [GoConnect-macOS-ARM.dmg](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **Linux (Debian/Ubuntu)** | [GoConnect-Linux.deb](https://github.com/orhaniscoding/goconnect/releases/latest) |
| **Linux (AppImage)** | [GoConnect-Linux.AppImage](https://github.com/orhaniscoding/goconnect/releases/latest) |

### Option 2: Terminal Application

For those who prefer the command line.

```bash
# Linux/macOS
curl -fsSL https://get.goconnect.io | sh

# Or manual download
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-linux-amd64
chmod +x goconnect-cli-linux-amd64
./goconnect-cli-linux-amd64
```

```powershell
# Windows (PowerShell)
irm https://get.goconnect.io/windows | iex

# Or manual download
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli-windows-amd64.exe" -OutFile "goconnect.exe"
.\goconnect.exe
```

### Option 3: Docker

For running as a server.

```bash
docker run -d \
  --name goconnect \
  --cap-add NET_ADMIN \
  -p 8080:8080 \
  -p 51820:51820/udp \
  ghcr.io/orhaniscoding/goconnect:latest
```

---

## 🎯 Usage

### Creating a Network (Host)

**Desktop App:**
1. Open GoConnect
2. Click "Create Network"
3. Enter network name (e.g., "My Minecraft Server")
4. Click "Create"
5. Share the invite link with friends!

**Terminal:**
```bash
$ goconnect

  🔗 GoConnect - Discord, but for networks

  ? What would you like to do?
  ❯ Create Network
    Join Network
    Settings
    Exit

# Select "Create Network" and follow the prompts
```

### Joining a Network (Client)

**Desktop App:**
1. Open GoConnect
2. Click "Join Network"
3. Paste the invite link
4. Click "Connect"
5. You're in!

**Terminal:**
```bash
$ goconnect join gc://invite.goconnect.io/abc123

✓ Connected successfully!
  Network: My Minecraft Server
  Your IP: 10.0.1.5
  Online: 3 members
```

### Quick Commands (Terminal)

| Command | Description |
|---------|-------------|
| `goconnect` | Interactive mode |
| `goconnect create "Name"` | Quick create network |
| `goconnect join <link>` | Quick join |
| `goconnect list` | List your networks |
| `goconnect status` | Connection status |
| `goconnect disconnect` | Disconnect |
| `goconnect help` | Help |

---

## ✨ Features

### Core Features (Free)

| Feature | Description |
|---------|-------------|
| 🌐 **Create Network** | Create your own virtual LAN |
| 🔗 **Join Network** | One-click join via invite link |
| 💬 **Text Chat** | Modern chat channels |
| 👥 **Member Management** | Invite, kick, ban |
| 🔒 **Secure Connection** | WireGuard encryption |
| 🖥️ **Cross-Platform** | Windows, macOS, Linux |
| 📱 **Multi-Device** | Multiple devices per account |

### Coming Soon

| Feature | Status |
|---------|--------|
| 📱 Mobile App | 🔜 Coming Soon |
| 🎤 Voice Chat | 📋 Planned |
| 📹 Video Call | 📋 Planned |
| 🎮 Game Integration | 📋 Planned |

---

## 🏗️ Architecture

GoConnect consists of three main components:

```
┌─────────────────────────────────────────────────────────────┐
│                    GoConnect Architecture                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect App (Tauri)                   │    │
│  │                                                      │    │
│  │  • Desktop application (Windows/macOS/Linux)        │    │
│  │  • Can be both host and client                      │    │
│  │  • Modern user-friendly interface             │    │
│  │  • Runs in system tray                              │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│                            │                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect CLI                           │    │
│  │                                                      │    │
│  │  • Terminal application                             │    │
│  │  • Interactive TUI interface                        │    │
│  │  • Same features, from terminal                     │    │
│  │  • Ideal for servers/headless environments          │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GoConnect Core (Go)                     │    │
│  │                                                      │    │
│  │  • WireGuard management                             │    │
│  │  • Network creation and management                  │    │
│  │  • User authentication                              │    │
│  │  • P2P connection coordination                      │    │
│  │  • Chat and messaging                               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

| Layer | Technology | Why? |
|-------|------------|------|
| **Desktop App** | Tauri + React | Small size, native performance |
| **CLI** | Go + Bubbletea | Cross-platform, single binary |
| **Core** | Go | Fast, secure, cross-platform |
| **Networking** | WireGuard | Modern, fast VPN protocol |
| **Database** | SQLite/PostgreSQL | Embedded or scalable |

---

## 🛠️ Development

### Requirements

- Go 1.24+
- Node.js 20+ (for Desktop App)
- Rust (for Desktop App)

### Building from Source

```bash
# Clone the repo
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Build CLI
cd cli
go build -o goconnect ./cmd/goconnect

# Build Desktop App
cd ../desktop
npm install
npm run tauri build
```

### Project Structure

```
goconnect/
├── desktop/               # Tauri desktop application
│   ├── src/               # React frontend
│   ├── src-tauri/         # Rust backend
│   └── package.json
├── cli/                   # Terminal application (Go)
│   ├── cmd/goconnect/     # Main command
│   ├── internal/          # Internal packages
│   └── go.mod
├── core/                  # Core library (Go)
│   ├── internal/          # Business logic
│   ├── cmd/server/        # Core entry point
│   └── go.mod
├── docs/                  # Documentation
├── README.md              # This file
└── LICENSE                # MIT License
```

---

## ❓ FAQ

### General Questions

<details>
<summary><b>Is GoConnect free?</b></summary>

Yes! Core features are completely free. Premium features may be added in the future, but core functionality will always remain free.
</details>

<details>
<summary><b>What platforms are supported?</b></summary>

- ✅ Windows 10/11
- ✅ macOS 11+ (Intel and Apple Silicon)
- ✅ Linux (Ubuntu 20.04+, Debian 11+, Fedora 35+)
- 🔜 Android (coming soon)
- 🔜 iOS (coming soon)
</details>

<details>
<summary><b>What's the difference from a VPN?</b></summary>

GoConnect is not a VPN, it's a virtual LAN platform:
- **VPN**: Routes all traffic through a server
- **GoConnect**: Direct peer-to-peer connections only between network devices

This results in lower latency and higher speeds.
</details>

<details>
<summary><b>Is it secure?</b></summary>

Yes! GoConnect uses industry-standard WireGuard encryption:
- ChaCha20 symmetric encryption
- Curve25519 key exchange
- Blake2s hash function
- Poly1305 message authentication
</details>

### Technical Questions

<details>
<summary><b>Do I need port forwarding?</b></summary>

Usually no! GoConnect uses NAT traversal techniques:
- UDP hole punching
- STUN/TURN servers
- Relay servers (last resort)

If direct connection fails, relay is used automatically.
</details>

<details>
<summary><b>Is there a bandwidth limit?</b></summary>

No limits on traffic through GoConnect servers because traffic flows directly between devices. Some limits may apply when using relay.
</details>

<details>
<summary><b>How many devices can connect?</b></summary>

Theoretically 65,534 devices per network (/16 subnet). Practical limit depends on your hardware and bandwidth.
</details>

---

## 🤝 Contributing

Contributions are welcome!

### How to Contribute

1. **Report Bugs**: [Open an issue](https://github.com/orhaniscoding/goconnect/issues/new)
2. **Suggest Features**: [Start a discussion](https://github.com/orhaniscoding/goconnect/discussions)
3. **Code Contributions**: Fork → Branch → PR

### Development Guidelines

- Use Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)
- Run tests: `make test`
- Lint check: `make lint`

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

```
MIT License - Use, modify, and distribute freely!
```

---

## 🙏 Acknowledgments

- [WireGuard](https://www.wireguard.com/) - Modern VPN protocol
- [Tauri](https://tauri.app/) - Desktop application framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- All open-source contributors

---

## 📞 Contact

- **GitHub**: [@orhaniscoding](https://github.com/orhaniscoding)
- **Issues**: [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues)
- **Discussions**: [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)

---

<div align="center">

**[⬆ Back to Top](#-goconnect)**

Made with ❤️

</div>
