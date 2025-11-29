# GoConnect — Virtual LAN Platform

**GoConnect** is a cross-platform virtual LAN/overlay network system that makes people on the internet appear as if they are on the same local network, with a Discord-like community structure.

> **Vision**: "Discord, but for networks."  
> **Latest Release:** v0.0.0 · 2025-11-29  
> **Author:** [@orhaniscoding](https://github.com/orhaniscoding)  
> **License:** MIT

## 🎯 Product Concept

### Core Architecture
```
TENANT (Server/Organization)
├── Multiple Networks (Virtual LANs)
├── Member Roles (Owner, Admin, Moderator, Member)
├── Community Chat
└── Invite System

NETWORK (Virtual LAN)
├── WireGuard Mesh Networking
├── IP Address Management
├── Network Chat
└── Member Access Control

CLIENT/DAEMON (User Device)
├── WireGuard Integration
├── Auto-connection
└── Multi-network Support
```

### User Experience
- **Zero Configuration**: Paste server URL, login, connect
- **Gaming Focus**: Perfect for Minecraft LAN, older games, file sharing
- **Cross-Platform**: Windows, Linux, macOS (mobile later)
- **Free Core**: Basic networking and tenant/network creation is free

## 🏗️ Architecture

### Components
1. **Server (Go)**: Central management hub with REST API
2. **Client Daemon (Go)**: Lightweight agent running on user devices
3. **Web UI (Next.js)**: Unified dashboard for management and chat

### Technology Stack
- **Backend**: Go 1.24+ with Gin framework
- **Frontend**: Next.js 15 with TypeScript and Tailwind CSS
- **Database**: PostgreSQL (production), SQLite (development)
- **Networking**: WireGuard for secure P2P connections
- **Real-time**: WebSocket for chat and status updates

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 18+ and npm
- PostgreSQL 15+ (optional)

### Development Setup

```bash
# Clone repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Start server
cd server
go run ./cmd/server

# Start client daemon
cd ../client-daemon
go run ./cmd/daemon

# Start web UI
cd ../web-ui
npm install
npm run dev
```

### Production Deployment

```bash
# Build server
cd server
go build -o goconnect-server ./cmd/server

# Build daemon
cd ../client-daemon
go build -o goconnect-daemon ./cmd/daemon

# Build web UI
cd ../web-ui
npm run build
```

## 📖 Documentation

- **[Technical Specification](docs/TECH_SPEC.md)** - Complete technical details
- **[API Documentation](server/openapi/openapi.yaml)** - REST API reference
- **[Security Policy](docs/SECURITY.md)** - Security best practices
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment

## 🔧 Configuration

### Environment Variables

**Server:**
```bash
PORT=8080
DATABASE_URL=postgresql://user:pass@localhost/goconnect
WIREGUARD_INTERFACE=wg0
```

**Client Daemon:**
```bash
GOCONNECT_SERVER_URL=http://localhost:8080
GOCONNECT_API_TOKEN=<your-token>
```

## 🏢 License

<<<<<<< HEAD
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

Current coverage (as of 2025-11-29):
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
2. Work directly on `main` branch (no feature branches)
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

- **Language**: Go 1.24+, TypeScript
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

**Built with ❤️ by orhaniscoding** | Latest Release: v0.0.0 (2025-11-29)
=======
MIT License - see [LICENSE](LICENSE) file for details.

---

**Built with ❤️ for gamers and communities**
>>>>>>> aeb0c86 (feat: Complete GoConnect architecture cleanup and product-ready implementation)
