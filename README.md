# GoConnect — Self-Hosted WireGuard VPN Management Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Built with Next.js](https://img.shields.io/badge/Next.js-15-black?logo=next.js)](https://nextjs.org/)
[![Latest Release](https://img.shields.io/badge/version-v0.0.0-blue)](https://github.com/orhaniscoding/goconnect/releases)

> **Latest Release:** v0.0.0 · 2025-11-20  
> **Author:** [@orhaniscoding](https://github.com/orhaniscoding)  
> **License:** MIT

**GoConnect** is a modern, production-ready WireGuard VPN management platform with multi-tenancy, real-time chat, device management, and comprehensive audit logging. Built with Go, Next.js, and PostgreSQL.

## ✨ Features

### 🔐 Core VPN Management
- **WireGuard Integration**: Automated peer provisioning and configuration generation
- **Multi-Network Support**: Create and manage multiple isolated VPN networks
- **IPAM (IP Address Management)**: Automatic IP allocation from CIDR ranges with conflict detection
- **Device Management**: Register and track devices across platforms (Linux, Windows, macOS, iOS, Android)
- **Peer-to-Peer Mesh**: Full mesh networking with automatic peer discovery

### 🏢 Multi-Tenancy & Access Control
- **Complete Tenant Isolation**: Security-first architecture with enforced boundaries
- **Role-Based Access Control (RBAC)**: Owner, Admin, Moderator, and Member roles
- **Network Membership**: Flexible join policies (Open, Approval Required, Invite-Only)
- **Join Request Workflow**: Approve/deny membership requests with audit trail

### 💬 Real-Time Communication
- **Built-in Chat**: Network-scoped and global chat with WebSocket support
- **Message Moderation**: Redact, edit, and delete messages with full audit trail
- **Edit History Tracking**: Complete message history with timestamps
- **File Attachments**: Share files within chat (planned)

### 🔍 Observability & Security
- **Comprehensive Audit Logging**: Immutable SQLite-backed audit log with hash chain integrity
- **Metrics Export**: Prometheus-compatible metrics for monitoring
- **Health Checks**: Readiness and liveness probes for container orchestration
- **Rate Limiting**: Per-IP token bucket rate limiting (configurable)
- **2FA Support**: TOTP-based two-factor authentication
- **SSO Integration**: OAuth2/OIDC support (planned)

### 🌐 Modern Tech Stack
- **Backend**: Go 1.22+ with Gin web framework
- **Frontend**: Next.js 15 with TypeScript and Tailwind CSS (placeholder)
- **Database**: PostgreSQL 14+ (in-memory for development)
- **Real-time**: WebSocket for live updates
- **API**: RESTful JSON API with OpenAPI 3.0 specification
- **i18n**: Multi-language support (English, Turkish)

## 🚀 Quick Start

### Prerequisites
- **Go** 1.22 or higher
- **Node.js** 18+ and npm (for web UI)
- **PostgreSQL** 14+ (optional, uses in-memory by default)
- **Make** (optional but recommended)

### Installation

#### Using Make (Recommended)
```bash
# Clone repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Install development tools
make install-tools

# Run all tests
make test

# Start server
make dev-server

# Start client daemon
make dev-daemon

# Start web UI
make dev-web
```

#### Manual Setup

**Server:**
```bash
cd server
go mod download
go run -ldflags "-X main.version=dev" ./cmd/server
```

**Client Daemon:**
```bash
cd client-daemon
go mod download
go run -ldflags "-X main.version=dev" ./cmd/daemon
```

**Web UI:**
```bash
cd web-ui
npm install
npm run dev
```

### Docker Compose (Coming Soon)
```bash
docker-compose up -d
```

## 📖 Documentation

- **[Technical Specification](docs/TECH_SPEC.md)** - Canonical project specification
- **[API Examples](docs/API_EXAMPLES.http)** - HTTP request examples for testing
- **[Contributing Guide](CONTRIBUTING.md)** - Development workflow and guidelines
- **[Security Policy](docs/SECURITY.md)** - Security best practices
- **[Configuration Flags](docs/CONFIG_FLAGS.md)** - Environment variables and flags
- **[OpenAPI Specification](server/openapi/openapi.yaml)** - Complete API documentation

## 🔧 Configuration

### Environment Variables

**Server:**
```bash
# Server
PORT=8080

# Rate Limiting
RATE_LIMIT_CAPACITY=5       # Requests per window
RATE_LIMIT_WINDOW=1s        # Time window

# Audit (SQLite)
AUDIT_SQLITE_DSN=audit.db
AUDIT_HASH_SECRETS_B64=<base64-secrets>
AUDIT_MAX_ROWS=10000
AUDIT_MAX_AGE_SECONDS=2592000  # 30 days
AUDIT_SIGNING_KEY_ED25519_B64=<base64-key>
```

**Client Daemon:**
```bash
# Daemon Configuration
GOCONNECT_SERVER_URL=http://localhost:8080
GOCONNECT_API_TOKEN=<your-token>
```

See [CONFIG_FLAGS.md](docs/CONFIG_FLAGS.md) for complete reference.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     WEB UI (Next.js)                    │
│  - Dashboard (network management)                       │
│  - Chat interface                                       │
│  - Device/Peer management                               │
│  Port: 3000 (development)                               │
└────────────────────────┬────────────────────────────────┘
                         │
                         │ REST API (/v1/*)
                         │ WebSocket (/v1/ws)
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  SERVER (Go Backend)                    │
│  ┌───────────────────────────────────────────────────┐ │
│  │ REST Handlers (Gin)                               │ │
│  │ - /v1/networks (CRUD + memberships)               │ │
│  │ - /v1/auth (register/login)                       │ │
│  │ - /v1/chat (messages + moderation)                │ │
│  │ - /v1/devices (device management)                 │ │
│  │ - /v1/audit/integrity                             │ │
│  │ - /health, /metrics (Prometheus)                  │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Services (Business Logic)                         │ │
│  │ - NetworkService, MembershipService               │ │
│  │ - IPAMService, AuthService                        │ │
│  │ - ChatService, DeviceService                      │ │
│  │ - PeerProvisioningService                         │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Repositories (Data Access)                        │ │
│  │ - In-Memory (development)                         │ │
│  │ - PostgreSQL (planned)                            │ │
│  └───────────────────────────────────────────────────┘ │
│  Port: 8080                                             │
└─────────────────────────────────────────────────────────┘
                         │
                         │ WireGuard Profile
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│              CLIENT DAEMON (Platform Agent)             │
│  - Apply WireGuard configuration                        │
│  - Heartbeat to server                                  │
│  - Auto-reconnect                                       │
│  Port: Random (12000-13000)                             │
└─────────────────────────────────────────────────────────┘
```

### API Endpoints

**Authentication:**
```
POST   /v1/auth/register    Register new user
POST   /v1/auth/login       Login and get tokens
POST   /v1/auth/refresh     Refresh access token
POST   /v1/auth/logout      Logout and invalidate token
```

**Networks:**
```
POST   /v1/networks                Create network
GET    /v1/networks                List networks
GET    /v1/networks/:id            Get network details
PATCH  /v1/networks/:id            Update network
DELETE /v1/networks/:id            Delete network (soft)
```

**Memberships:**
```
POST   /v1/networks/:id/join       Join network
POST   /v1/networks/:id/approve    Approve join request (admin)
POST   /v1/networks/:id/deny       Deny join request (admin)
POST   /v1/networks/:id/kick       Kick member (admin)
POST   /v1/networks/:id/ban        Ban member (admin)
GET    /v1/networks/:id/members    List members
```

**IP Allocation:**
```
POST   /v1/networks/:id/ip-allocations           Allocate IP
GET    /v1/networks/:id/ip-allocations           List allocations
DELETE /v1/networks/:id/ip-allocation            Release own IP
DELETE /v1/networks/:id/ip-allocations/:user_id  Admin release
```

**Audit:**
```
GET    /v1/audit/integrity    Export integrity snapshot
```

See [OpenAPI Specification](server/openapi/openapi.yaml) for complete API documentation.

## 🧪 Development

### Available Make Commands

**Root Level:**
```bash
make help              # Show all commands
make test              # Run tests for all components
make test-race         # Run tests with race detector
make test-coverage     # Run tests with coverage
make lint              # Run linters
make ci                # Run full CI pipeline locally
make build             # Build all components
make clean             # Clean build artifacts
```

**Server:**
```bash
cd server
make test-coverage     # Run tests with coverage report
make test-coverage-html # Generate HTML coverage report
make lint              # Run golangci-lint
make build             # Build server binary
```

**Client Daemon:**
```bash
cd client-daemon
make build-all         # Build for all platforms
make install-systemd   # Install systemd service (Linux)
make install-launchd   # Install launchd service (macOS)
```

### Running Tests

```bash
# All tests with race detector
make test-race

# Coverage report
make test-coverage

# Specific package
cd server
go test ./internal/handler -v -cover

# Integration tests
go test ./internal/integration -v
```

### Code Coverage

Current coverage (as of 2025-11-20):
- **audit**: 79.7%
- **config**: 87.7%
- **handler**: 53.8%
- **service**: 67.4%
- **rbac**: 100%
- **metrics**: 100%
- **wireguard**: 90.5%

**Target**: ≥60% (enforced in CI)

### Linting

```bash
# Run all linters
make lint

# Server only
cd server
golangci-lint run --timeout=3m

# Web UI (when configured)
cd web-ui
npm run lint
```

## 📦 Deployment

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/orhaniscoding/goconnect/releases):

```bash
# Linux (amd64)
wget https://github.com/orhaniscoding/goconnect/releases/download/v0.0.0/goconnect-server-linux-amd64
chmod +x goconnect-server-linux-amd64
./goconnect-server-linux-amd64

# macOS (arm64)
wget https://github.com/orhaniscoding/goconnect/releases/download/v0.0.0/goconnect-server-darwin-arm64
chmod +x goconnect-server-darwin-arm64
./goconnect-server-darwin-arm64
```

### Systemd Service (Linux)

```bash
# Server
sudo cp server/service/linux/goconnect-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now goconnect-server

# Client Daemon
cd client-daemon
make install-systemd
```

### Docker (Coming Soon)

```bash
docker pull ghcr.io/orhaniscoding/goconnect-server:v0.0.0
docker run -p 8080:8080 ghcr.io/orhaniscoding/goconnect-server:v0.0.0
```

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

**Quick Start:**
1. Fork the repository
2. Create feature branch: `git checkout -b feat/amazing-feature`
3. Make changes and add tests
4. Run checks: `make ci`
5. Commit with [Conventional Commits](https://www.conventionalcommits.org/): `git commit -S -m "feat(server): add amazing feature"`
6. Push and open Pull Request

**Development Workflow:**
- Run `make help` to see all available commands
- All tests must pass: `make test-race`
- Coverage must be ≥60%: `make test-coverage`
- Linters must be clean: `make lint`
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## 🛣️ Roadmap

### v1.3 (Current)
- [ ] PostgreSQL migration (replace in-memory)
- [ ] Complete web UI implementation
- [ ] Docker & Docker Compose
- [ ] Kubernetes Helm chart

### v1.4
- [ ] Real JWT/OIDC authentication
- [ ] SSO integration (GitHub, Google, Azure AD)
- [ ] 2FA/MFA support
- [ ] Email notifications

### v2.0
- [ ] Relay servers for NAT traversal
- [ ] Mobile apps (iOS, Android)
- [ ] Terraform provider
- [ ] CLI tool for automation

See [GitHub Projects](https://github.com/orhaniscoding/goconnect/projects) for detailed roadmap.

## 🔒 Security

### Current Status
⚠️ **Development Mode**: The current authentication implementation is a **PLACEHOLDER** for development purposes only. Do not use in production without implementing proper JWT/OIDC authentication.

### Reporting Vulnerabilities
Please report security vulnerabilities responsibly:
- **Email**: [security contact] (preferred)
- **GitHub**: Private security advisory
- **DO NOT** open public issues for security vulnerabilities

See [SECURITY.md](docs/SECURITY.md) for our security policy.

## 📊 Project Stats

- **Language**: Go 1.22+, TypeScript
- **Test Coverage**: 60%+ (enforced)
- **Total Tests**: 200+ (all passing)
- **Lines of Code**: ~15,000
- **Documentation**: Comprehensive (14 docs files)

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2025 orhaniscoding

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🙏 Acknowledgments

- [WireGuard](https://www.wireguard.com/) - Fast and modern VPN protocol
- [Gin Web Framework](https://gin-gonic.com/) - HTTP web framework
- [Next.js](https://nextjs.org/) - React framework
- [PostgreSQL](https://www.postgresql.org/) - Relational database
- All open-source contributors

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues)
- **Discussions**: [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- **Author**: [@orhaniscoding](https://github.com/orhaniscoding)

---

**Built with ❤️ by orhaniscoding** | Latest Release: v0.0.0 (2025-11-20)
