# GoConnect — Self-Hosted WireGuard VPN Management Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Built with Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)](https://nextjs.org/)

**GoConnect** is a modern, production-ready WireGuard VPN management platform with multi-tenancy, real-time chat, device management, and comprehensive audit logging. Built with Go, Next.js, and PostgreSQL.

## ✨ Features

### 🔐 Core VPN Management
- **WireGuard Integration**: Automated peer provisioning and configuration generation
- **Multi-Network Support**: Create and manage multiple isolated VPN networks
- **IPAM (IP Address Management)**: Automatic IP allocation from CIDR ranges
- **Device Management**: Register and track devices across platforms (Linux, Windows, macOS, iOS, Android)
- **Peer-to-Peer Mesh**: Full mesh networking with peer discovery

### 🏢 Multi-Tenancy & Access Control
- **Complete Tenant Isolation**: Security-first architecture with enforced boundaries
- **Role-Based Access Control (RBAC)**: Admin, Moderator, and Member roles
- **Network Membership**: Join policies (Open, Approval Required, Invite-Only, Closed)
- **Join Request Workflow**: Approve/deny membership requests

### 💬 Real-Time Communication
- **Built-in Chat**: Network-scoped and global chat with WebSocket support
- **Message Moderation**: Redact, edit, and delete messages
- **Audit Trail**: Complete edit history tracking
- **File Attachments**: Share files within chat (planned)

### 🔍 Observability & Security
- **Comprehensive Audit Logging**: Track all actions with detailed context
- **Metrics Export**: Prometheus-compatible metrics
- **Health Checks**: Readiness and liveness probes
- **2FA Support**: TOTP-based two-factor authentication (planned)
- **SSO Integration**: OAuth2/OIDC support (planned)

### 🌐 Modern Tech Stack
- **Backend**: Go 1.21+ with Gin web framework
- **Frontend**: Next.js 14 with TypeScript and Tailwind CSS
- **Database**: PostgreSQL 14+ with connection pooling
- **Real-time**: WebSocket for live updates
- **API**: RESTful JSON API with OpenAPI 3.0 spec
- **i18n**: Multi-language support (English, Turkish)

## 🚀 Quick Start

### Prerequisites
- **Go** 1.21 or higher
- **Node.js** 18+ and npm/yarn
- **PostgreSQL** 14 or higher
- **WireGuard** kernel module (for server host)

### Installation

#### 1. Clone the Repository
```bash
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect
```

#### 2. Set Up Database
```bash
# Create PostgreSQL database
createdb goconnect

# Run migrations (using go-migrate or similar)
cd server
make migrate-up
```

#### 3. Configure Server
```bash
cd server
cp .env.example .env
# Edit .env with your settings:
# - DATABASE_URL
# - JWT_SECRET
# - SERVER_PORT
```

#### 4. Run Server
```bash
# Development mode
make dev

# Production build
make build
./bin/goconnect-server
```

#### 5. Set Up Web UI
```bash
cd web-ui
npm install
npm run dev
# Access at http://localhost:3000
```

### Docker Compose (Recommended)
```bash
# Start all services (server, database, web UI)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 📖 Documentation

### Core Concepts
- **[Technical Specification](docs/TECH_SPEC.md)**: Architecture and design decisions
- **[API Examples](docs/API_EXAMPLES.http)**: HTTP request examples
- **[Configuration Flags](docs/CONFIG_FLAGS.md)**: Environment variables and CLI flags
- **[WebSocket Messages](docs/WS_MESSAGES.md)**: Real-time event formats

### Security
- **[Security Policy](docs/SECURITY.md)**: Vulnerability reporting and best practices
- **[Threat Model](docs/THREAT_MODEL.md)**: Security considerations and mitigations
- **[SSO & 2FA](docs/SSO_2FA.md)**: Authentication integration guide

### Operations
- **[Runbooks](docs/RUNBOOKS.md)**: Operational procedures and troubleshooting
- **[Local Bridge](docs/LOCAL_BRIDGE.md)**: Bridge mode for local network access

## 🔧 Configuration

### Server Environment Variables
```bash
# Database
DATABASE_URL="postgres://user:pass@localhost:5432/goconnect?sslmode=disable"

# JWT Authentication
JWT_SECRET="your-secret-key-change-this"
JWT_ACCESS_TTL="15m"
JWT_REFRESH_TTL="168h"  # 7 days

# Server
SERVER_PORT="8080"
SERVER_HOST="0.0.0.0"

# WireGuard
WG_SERVER_ENDPOINT="vpn.example.com:51820"
WG_SERVER_PUBLIC_KEY="server-public-key"
WG_SERVER_PRIVATE_KEY="server-private-key"

# CORS
CORS_ORIGINS="http://localhost:3000,https://yourdomain.com"

# Observability
METRICS_ENABLED="true"
AUDIT_LOG_PATH="/var/log/goconnect/audit.jsonl"
```

### Web UI Configuration
```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/v1/ws
```

## 🏗️ Architecture

### System Components
```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Web UI    │─────▶│   Server    │─────▶│  PostgreSQL │
│  (Next.js)  │◀─────│    (Go)     │◀─────│             │
└─────────────┘      └─────────────┘      └─────────────┘
                           │
                           │ WireGuard
                           ▼
                     ┌─────────────┐
                     │   Clients   │
                     │  (Devices)  │
                     └─────────────┘
```

### API Structure
- `POST /v1/auth/register` - User registration
- `POST /v1/auth/login` - Authentication
- `GET /v1/networks` - List networks
- `POST /v1/networks` - Create network
- `GET /v1/networks/:id/wg/profile` - Download WireGuard config
- `POST /v1/networks/:id/join` - Request to join network
- `GET /v1/chat/messages` - Retrieve chat messages
- `WebSocket /v1/ws` - Real-time updates

See [API_EXAMPLES.http](docs/API_EXAMPLES.http) for complete API documentation.

## 🧪 Development

### Running Tests
```bash
# Server tests
cd server
make test

# With coverage
make test-coverage

# Integration tests
make test-integration

# Web UI tests
cd web-ui
npm test
```

### Code Quality
```bash
# Linting
make lint

# Formatting
make fmt

# Security scan
make security-check
```

### Database Migrations
```bash
# Create new migration
make migrate-create NAME=add_feature

# Apply migrations
make migrate-up

# Rollback
make migrate-down
```

## 📦 Deployment

### Binary Releases
Download pre-built binaries from [Releases](https://github.com/orhaniscoding/goconnect/releases).

### systemd Service (Linux)
```bash
# Server service
sudo cp server/service/linux/goconnect-server.service /etc/systemd/system/
sudo systemctl enable --now goconnect-server

# Daemon service (client-side)
cd client-daemon
make install
```

### Production Checklist
- [ ] Set strong JWT_SECRET
- [ ] Enable HTTPS with valid certificates
- [ ] Configure firewall rules (allow WireGuard UDP port)
- [ ] Set up database backups
- [ ] Configure log rotation
- [ ] Enable metrics collection
- [ ] Review CORS origins
- [ ] Set up monitoring and alerting

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Commit Convention
We follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [WireGuard](https://www.wireguard.com/) - Fast, modern VPN protocol
- [Gin](https://gin-gonic.com/) - High-performance Go web framework
- [Next.js](https://nextjs.org/) - React framework for production
- [PostgreSQL](https://www.postgresql.org/) - Advanced open-source database

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues)
- **Discussions**: [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- **Security**: See [SECURITY.md](docs/SECURITY.md) for vulnerability reporting

---

**Built with ❤️ by [orhaniscoding](https://github.com/orhaniscoding)**

Latest Release: {LATEST_TAG} · {RELEASE_DATE}
