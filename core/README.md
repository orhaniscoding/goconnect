# ⚙️ GoConnect Core (Server)

GoConnect's backend server for self-hosted deployments. Provides network coordination, user management, and real-time communication.

> **Note:** Most users don't need this. Use the [Desktop App](../desktop/) or [CLI](../cli/) instead. This is for running your own GoConnect infrastructure.

---

## 🎯 What is it?

GoConnect Core is the backend server providing:

- **Network Management**: Virtual LAN creation and coordination
- **User Management**: Authentication, authorization, RBAC
- **Real-time Communication**: WebSocket-based chat and signaling
- **WireGuard Integration**: Key management and configuration
- **Multi-tenant Support**: Isolated tenant spaces

---

## 🚀 Quick Start

### Docker (Recommended)

```bash
# Pull and run
docker run -d \
  --name goconnect-server \
  --cap-add NET_ADMIN \
  -p 8080:8080 \
  -p 51820:51820/udp \
  -v goconnect-data:/data \
  ghcr.io/orhaniscoding/goconnect-server:latest

# Open setup wizard
open http://localhost:8080/setup
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    ports:
      - "8080:8080"
      - "51820:51820/udp"
    volumes:
      - goconnect-data:/data
    environment:
      - JWT_SECRET=your-secret-key-here
      - DATABASE_URL=sqlite:///data/goconnect.db
    cap_add:
      - NET_ADMIN

volumes:
  goconnect-data:
```

### Binary

```bash
# Download
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_linux_amd64.tar.gz
tar -xzf goconnect-server_linux_amd64.tar.gz

# Configure
cp config.example.env .env
# Edit .env with your settings

# Run
./goconnect-server
```

---

## ⚙️ Configuration

All configuration via environment variables:

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secret key for JWT tokens (32+ chars) | `openssl rand -base64 48` |
| `DATABASE_URL` | Database connection string | `postgres://user:pass@host/db` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | HTTP API port |
| `WG_PORT` | `51820` | WireGuard UDP port |
| `WG_SUBNET` | `10.0.0.0/8` | VPN subnet |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |

### Example `.env`

```bash
# Required
JWT_SECRET=your-super-secret-key-at-least-32-characters
DATABASE_URL=postgres://goconnect:password@localhost:5432/goconnect?sslmode=disable

# WireGuard
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=YourServerPublicKey=
WG_SERVER_PRIVKEY=YourServerPrivateKey=

# Optional
HTTP_PORT=8080
LOG_LEVEL=info
```

---

## 🏗️ Architecture

```
core/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── handler/              # HTTP handlers
│   │   ├── auth.go           # Authentication
│   │   ├── network.go        # Network management
│   │   └── user.go           # User management
│   ├── service/              # Business logic
│   │   ├── auth_service.go
│   │   ├── network_service.go
│   │   └── user_service.go
│   ├── repository/           # Database layer
│   │   ├── postgres/
│   │   └── sqlite/
│   ├── websocket/            # Real-time communication
│   ├── wireguard/            # WireGuard management
│   ├── rbac/                 # Role-based access control
│   └── audit/                # Audit logging
├── migrations/               # Database migrations
│   ├── postgres/
│   └── sqlite/
├── openapi/
│   └── openapi.yaml          # API specification
└── go.mod
```

---

## 🔒 Security Features

- **JWT Authentication** - Secure token-based auth with refresh tokens
- **Argon2id Passwords** - OWASP-recommended password hashing
- **RBAC** - Admin, Moderator, User roles
- **2FA Support** - TOTP-based two-factor authentication
- **Audit Logging** - Comprehensive event tracking
- **Rate Limiting** - Protection against abuse

---

## 📚 API Documentation

API documentation available at `/docs` when running the server, or see [openapi/openapi.yaml](openapi/openapi.yaml).

### Key Endpoints

```
POST   /v1/auth/register         # Register new user
POST   /v1/auth/login            # Login
POST   /v1/auth/refresh          # Refresh token

GET    /v1/networks              # List networks
POST   /v1/networks              # Create network
GET    /v1/networks/:id          # Get network
DELETE /v1/networks/:id          # Delete network

POST   /v1/networks/:id/join     # Join network
POST   /v1/networks/:id/leave    # Leave network

WS     /v1/ws                    # WebSocket connection
```

---

## 🛠️ Development

### Requirements

- Go 1.24+
- PostgreSQL 15+ (or SQLite for development)
- WireGuard

### Build

```bash
# Development
go run ./cmd/server

# Production build
go build -ldflags="-s -w" -o goconnect-server ./cmd/server

# With version info
VERSION=v3.0.0
go build -ldflags="-s -w -X main.version=${VERSION}" -o goconnect-server ./cmd/server
```

### Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests (requires database)
go test -tags=integration ./...
```

### Database Migrations

```bash
# Run migrations
go run ./cmd/server -migrate

# Or using goose directly
goose -dir migrations/postgres postgres "your-connection-string" up
```

---

## 🐳 Docker Build

```bash
# Build image
docker build -t goconnect-server .

# Run
docker run -p 8080:8080 goconnect-server
```

---

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
