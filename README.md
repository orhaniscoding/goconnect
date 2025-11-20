# GoConnect — Self-Hosted VPN Platform# GoConnect — by orhaniscoding (Orhan Tüzer)



[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)Latest: {LATEST_TAG} · {RELEASE_DATE}

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/orhaniscoding/goconnect)Binaries: goconnect-server · goconnect-daemon

© 2025 orhaniscoding — MIT

**GoConnect** is a modern, self-hosted VPN platform with multi-tenancy support, real-time chat, and comprehensive network management. Built with Go, PostgreSQL, and Next.js.

> **Author**: [orhaniscoding](https://github.com/orhaniscoding) (Orhan Tüzer)  
> **License**: MIT  
> **Status**: Active Development (v0.1.0)

---

## 🚀 Features

### ✅ Completed
- **🔐 Authentication & Authorization**
  - JWT-based auth with access + refresh tokens
  - Argon2id password hashing
  - Multi-tenant support
  - Role-based access control (Admin & Moderator roles)

- **🖥️ Device Management**
  - Device registration with WireGuard public keys
  - Multi-platform support (Windows, macOS, Linux, Android, iOS)
  - Device heartbeat and activity tracking
  - Soft enable/disable functionality

- **🔐 WireGuard Profile Generation**
  - Automatic WireGuard configuration file generation
  - Per-device IP allocation from network CIDR
  - Configurable DNS, MTU, and keepalive settings
  - Secure profile rendering with audit logging

- **🌐 Network Management**
  - Create and manage virtual networks
  - Public/private network visibility
  - Join request approval workflow

- **💬 Real-Time Chat**
  - WebSocket-based messaging
  - Message editing with time limits
  - Soft/hard delete modes
  - Edit history tracking
  - **Content moderation** (redaction by moderators/admins)

- **📊 IP Address Management (IPAM)**
  - Automatic IP allocation
  - CIDR overlap detection

- **🔍 Audit Trail**
  - SQLite-backed audit logging
  - SHA-256 hash chain integrity

- **📈 Observability**
  - Prometheus metrics export

---

## 📚 Quick Start

```bash
# Clone repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect/server

# Run migrations
go run cmd/server/main.go -migrate

# Start server
go run cmd/server/main.go
```

See [docs/](./docs/) for full documentation.

---

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

**Author**: Orhan Tüzer ([@orhaniscoding](https://github.com/orhaniscoding))
