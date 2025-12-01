# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [3.0.0](https://github.com/orhaniscoding/goconnect/compare/v2.29.1...v3.0.0) (2025-12-02)

### ⚠️ BREAKING CHANGE - Complete Architecture Overhaul

This release completely redesigns GoConnect with a modern, unified architecture. Three separate applications for different use cases: Desktop App (GUI), CLI (Terminal TUI), and Server (Self-Hosted).

### ✨ New Features

#### Desktop Application (Tauri)
* **Cross-platform GUI** - Native app for Windows, macOS, and Linux
* **Modern React UI** - Clean, user-friendly interface with Tailwind CSS
* **System Tray** - Quick access to networks and status
* **Tauri 2.0** - Small binary size (~15MB), native performance

#### CLI (Terminal TUI)
* **Interactive TUI** - Beautiful terminal interface with Bubbletea
* **Daemon Architecture** - Background service with gRPC IPC
* **Cross-platform IPC** - Unix sockets (Linux/macOS) and Named Pipes (Windows)
* **Chat Support** - Full chat functionality from terminal
* **File Transfer** - P2P file transfer between peers

#### Server (Self-Hosted)
* **Multi-tenant Architecture** - Support for multiple organizations
* **RBAC** - Admin, Moderator, and User roles
* **2FA Support** - TOTP-based two-factor authentication
* **Audit Logging** - Comprehensive event tracking
* **Docker Support** - Easy deployment with Docker

### 🔧 Technical Improvements

* **Go 1.24+** - Updated to latest Go version
* **gRPC IPC** - Type-safe inter-process communication
* **SQLite + PostgreSQL** - Flexible database options
* **172 Tests** - Comprehensive test coverage
* **GitHub Actions** - Automated CI/CD with multi-platform releases

### 📦 Release Assets

Now includes three types of releases:
1. **Desktop App** - `.exe`, `.msi`, `.dmg`, `.deb`, `.AppImage`
2. **CLI** - Standalone binaries for all platforms
3. **Server** - Binaries and Docker images

### 🔄 Migration from v2.x

The v3.0 architecture is completely different:
- `goconnect-daemon` → `goconnect-cli` (renamed, now includes TUI)
- `goconnect-server` → Same name, updated internals
- New: `GoConnect` desktop app

---

## [2.29.1](https://github.com/orhaniscoding/goconnect/compare/v2.29.0...v2.29.1) (2025-12-01)

### 🔧 Fixes

* **ci:** Fix GitHub Actions workflows for new folder structure
* **ci:** Restore deleted workflow files from commit 1085dba
* **release:** Update build paths from `server/` to `core/` and `client-daemon/` to `cli/`

---

## [2.29.0](https://github.com/orhaniscoding/goconnect/compare/v2.28.2...v2.29.0) (2025-12-01)

### ✨ Features

* **cli:** Add daemon architecture with gRPC IPC
* **cli:** Implement Windows Named Pipes support
* **cli:** Add chat storage with SQLite persistence
* **cli:** Add file transfer manager with progress tracking
* **desktop:** Add Tauri commands for daemon communication
* **desktop:** Implement connection status hooks

### 🧪 Testing

* Add 172 comprehensive tests across all packages
* Implement daemon IPC tests for Windows and Unix
* Add chat manager and storage tests
* Add file transfer tests with progress callbacks

---

## [2.28.2](https://github.com/orhaniscoding/goconnect/compare/v2.28.0...v2.28.2) (2025-11-30)

### 🔧 Fixes

* **migrations:** Fix PostgreSQL schema for posts, devices, peers tables
* **server:** Simplify tenant CREATE query for registration flow

---

## [2.28.0](https://github.com/orhaniscoding/goconnect/compare/v2.27.0...v2.28.0) (2025-11-30)

### ✨ Features

* **server:** Add interactive setup wizard with web UI
* **daemon:** Add interactive CLI setup command
* **web-ui:** Fix Next.js 15+ params async compatibility

### 🔧 Fixes

* **migrations:** Add proper up/down migration files for Goose format

---

## [2.27.0](https://github.com/orhaniscoding/goconnect/compare/v2.26.0...v2.27.0) (2025-11-29)

### ✨ Features

* Complete GoConnect architecture cleanup and product-ready implementation
* **daemon,web:** Implement localhost bridge, deep linking, and daemon discovery

---

## Previous Releases

For older releases, see [GitHub Releases](https://github.com/orhaniscoding/goconnect/releases).
