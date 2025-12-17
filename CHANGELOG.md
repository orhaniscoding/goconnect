# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0](https://github.com/orhaniscoding/goconnect/releases/tag/v1.1.0) (2025-12-17)

### ✨ New Features

* **Push Notifications** - Cross-platform native notifications
  - Linux: `notify-send` integration
  - macOS: `osascript` AppleScript
  - Windows: PowerShell toast notifications
  - Settings: Enable/Disable, Do Not Disturb, Sound

* **Metrics Dashboard** - Real-time monitoring
  - 6 Prometheus metrics (WebSocket connections, rooms, networks, peers)
  - JSON summary endpoint (`/v1/metrics/summary`)
  - React dashboard component with auto-refresh

* **Auto-Update System** - Keep your client updated
  - CLI: `goconnect update` command with GitHub release checking
  - Desktop: Tauri updater plugin integration
  - Automatic version comparison and download

### 🔧 Technical

* Added `tauri-plugin-notification` and `tauri-plugin-updater`
* New Metrics tab in desktop app
* Extended Settings panel with notification controls

---

## [1.0.0](https://github.com/orhaniscoding/goconnect/releases/tag/v1.0.0) (2025-12-17)

### 🎉 Initial Release

Complete rewrite and fresh start with production-ready features.

### ✨ Core Features

* **Virtual LAN** - WireGuard-based secure networking
* **Desktop App** - Tauri 2.0 + React with native performance
* **CLI Application** - Go + Bubbletea interactive TUI
* **Server** - Self-hosted backend with PostgreSQL/SQLite

### 💬 Communication

* **Text Chat** - Real-time messaging with WebSocket
* **Voice Chat** - WebRTC-based voice communication
* **Presence System** - Online/offline status tracking

### 📁 File Sharing

* **P2P Transfers** - Direct file sharing between peers
* **Progress Tracking** - Real-time transfer status

### 👥 Network Management

* **Create Networks** - Host your own virtual LAN
* **Join via Invite** - One-click connection
* **Member Management** - Invite, kick, ban capabilities
* **RBAC** - Admin, Moderator, User roles

### 🔒 Security

* WireGuard encryption (ChaCha20, Curve25519)
* JWT authentication
* Optional 2FA (TOTP)
* Audit logging

### 🏗️ Architecture

| Component | Technology |
|-----------|------------|
| Desktop | Tauri 2.0, React, TypeScript |
| CLI | Go 1.24+, Bubbletea, gRPC |
| Server | Go, Gin, PostgreSQL/SQLite |
| Networking | WireGuard, WebRTC |

---

## Previous Development Versions

This is a fresh v1.0.0 release. For historical development notes, see [GitHub Commits](https://github.com/orhaniscoding/goconnect/commits/main).
