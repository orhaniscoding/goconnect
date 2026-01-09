# GoConnect Tech Stack

## Architecture Overview

GoConnect follows a multi-module architecture with three main components:

| Module | Purpose | Location |
|--------|---------|----------|
| **Core Server** | Backend API, signaling, coordination | `core/` |
| **CLI/Daemon** | System service, WireGuard management, IPC | `cli/` |
| **Desktop App** | User interface, user-facing features | `desktop/` |

---

## Backend (Core Server)

| Category | Technology | Notes |
|----------|------------|-------|
| **Language** | Go 1.24+ | Performance, concurrency, static binary |
| **Web Framework** | Gin | Fast HTTP router with middleware support |
| **Database** | PostgreSQL | Production database |
| **Database** | SQLite | Local/embedded database option |
| **Authentication** | JWT + Argon2id | Stateless auth with secure password hashing |
| **2FA** | TOTP | Time-based one-time passwords |
| **Real-time** | WebSocket | Signaling and live updates |

---

## CLI & Daemon

| Category | Technology | Notes |
|----------|------------|-------|
| **Language** | Go 1.24+ | Same as backend for code sharing |
| **TUI Framework** | Bubbletea | Interactive terminal UI |
| **IPC Protocol** | gRPC | Desktop ↔ Daemon communication |
| **IPC Transport** | Unix Sockets / Named Pipes | Platform-specific transport |
| **VPN Protocol** | WireGuard | Kernel module or userspace implementation |
| **NAT Traversal** | STUN/TURN | UDP hole punching for P2P |

---

## Desktop Application

| Category | Technology | Notes |
|----------|------------|-------|
| **Framework** | Tauri 2.x | Lightweight native app (~15MB) |
| **Backend** | Rust | Tauri core, secure IPC, system integration |
| **Frontend** | React 19 | Component-based UI |
| **Language** | TypeScript | Type-safe frontend code |
| **Styling** | Tailwind CSS | Utility-first CSS framework |
| **State Management** | Zustand / React Context | Lightweight state management |
| **Voice Chat** | WebRTC | Browser-native real-time communication |

---

## Communication Protocols

| Protocol | Purpose |
|----------|---------|
| **REST API** | Core server HTTP endpoints |
| **WebSocket** | Real-time signaling, chat, presence |
| **gRPC** | Desktop app ↔ Daemon IPC |
| **WireGuard** | Encrypted peer-to-peer tunneling |
| **WebRTC** | Voice communication |

---

## Development & Quality

| Category | Technology |
|----------|------------|
| **Go Testing** | `go test` + testify |
| **Go Linting** | golangci-lint |
| **TypeScript Testing** | Vitest |
| **TypeScript Linting** | ESLint + Prettier |
| **Rust Linting** | Clippy + rustfmt |
| **CI/CD** | GitHub Actions |
| **Versioning** | Conventional Commits |

---

## Security

| Aspect | Implementation |
|--------|----------------|
| **Encryption** | WireGuard (ChaCha20, Curve25519) |
| **TLS** | TLS 1.2+ for all HTTP connections |
| **Password Storage** | Argon2id hashing |
| **Input Validation** | Server-side validation on all endpoints |
| **Secrets** | Environment variables, never logged |

---

## Supported Platforms

### Currently Supported ✅
- **Windows** 10/11 (x64)
- **macOS** 11+ (Intel & Apple Silicon)
- **Linux** Debian/Ubuntu (.deb), AppImage

### Planned 🔜
- **Android** (Kotlin/Compose)
- **iOS** (Swift/SwiftUI)
- **Web Browser** (management dashboard)
