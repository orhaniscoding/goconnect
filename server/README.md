# GoConnect Server

Enterprise-grade self-hosted VPN management platform with WireGuard integration, multi-tenancy, and comprehensive audit logging.

## 🚀 Features

### Core Capabilities
- **Multi-Tenant Architecture**: Isolated tenant spaces with RBAC
- **WireGuard VPN**: Automated profile generation and device management
- **Real-time Chat**: WebSocket-based messaging with moderation
- **Device Management**: Registration, heartbeat tracking, and lifecycle management
- **Network Management**: CIDR-based network creation with IP allocation (IPAM)
- **Audit Logging**: Comprehensive event tracking with SQLite persistence

### Security & Authentication
- **JWT Authentication**: Secure token-based auth with refresh tokens
- **Argon2id Password Hashing**: OWASP-recommended parameters
- **RBAC**: Admin, Moderator, and User roles
- **Content Moderation**: Message redaction and edit history
- **2FA Support**: TOTP-ready architecture (implementation pending)

### Developer Experience
- **Environment-based Configuration**: 12-factor app compliant
- **PostgreSQL Integration**: Production-ready database layer
- **Migration System**: Goose-based database migrations
- **Comprehensive Testing**: 87.7% config coverage, 68.8% service coverage
- **OpenAPI Specification**: Documented REST API

## 📦 Quick Start

### Prerequisites
- **Go**: 1.21+
- **PostgreSQL**: 14+
- **WireGuard**: Server installation

### Installation

```bash
# Clone repository
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect/server

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env  # Set JWT_SECRET, WG_SERVER_ENDPOINT, WG_SERVER_PUBKEY, DB_PASSWORD

# Run migrations
go run cmd/server/main.go -migrate

# Start server
go run cmd/server/main.go
```

### Docker Deployment (Coming Soon)

```bash
docker-compose up -d
```

## ⚙️ Configuration

All configuration is done via environment variables. See `.env.example` for all options.

### Critical Variables

| Variable             | Description                                  | Example                   |
| -------------------- | -------------------------------------------- | ------------------------- |
| `JWT_SECRET`         | **REQUIRED** - 32+ char random key           | `openssl rand -base64 48` |
| `WG_SERVER_ENDPOINT` | **REQUIRED** - Public VPN endpoint           | `vpn.example.com:51820`   |
| `WG_SERVER_PUBKEY`   | **REQUIRED** - Server's WireGuard public key | `wg genkey \| wg pubkey`  |
| `DB_HOST`            | PostgreSQL host                              | `localhost`               |
| `DB_PASSWORD`        | Database password                            | `your_password`           |

### Optional Features

```bash
# Enable SQLite audit logs
AUDIT_SQLITE_DSN=./audit.db

# Enable async audit processing
AUDIT_ASYNC=true
AUDIT_QUEUE_SIZE=1024

# Configure CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com
```

## 🏗️ Architecture

```
server/
├── cmd/server/              # Application entry point
├── internal/
│   ├── audit/              # Audit logging (stdout + SQLite)
│   ├── config/             # Environment-based configuration ⭐ NEW
│   ├── database/           # PostgreSQL connection & migrations
│   ├── domain/             # Business entities & validation
│   ├── handler/            # HTTP handlers (REST + WebSocket)
│   ├── repository/         # Data access layer (in-memory + PostgreSQL)
│   ├── service/            # Business logic layer
│   ├── websocket/          # WebSocket hub & client management
│   └── wireguard/          # WireGuard profile generation
├── migrations/             # Database migration files
│   ├── 000001_initial_schema.sql
│   ├── 000002_chat_tables.sql
│   └── 000003_device_tables.sql ⭐ NEW
└── .env.example            # Environment template ⭐ NEW
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/config -v

# Run integration tests (requires PostgreSQL)
go test ./internal/repository -tags=integration
```

### Test Coverage Summary
- **Config**: 87.7%
- **WireGuard**: 91.8%
- **RBAC**: 100.0%
- **Audit**: 79.7%
- **Service Layer**: 68.8%

## 📚 API Documentation

### Authentication
```http
POST /v1/auth/register
POST /v1/auth/login
POST /v1/auth/refresh
```

### Networks
```http
POST   /v1/networks
GET    /v1/networks
GET    /v1/networks/:id
PATCH  /v1/networks/:id
DELETE /v1/networks/:id
```

### Devices
```http
POST   /v1/devices
GET    /v1/devices
GET    /v1/devices/:id
PATCH  /v1/devices/:id
DELETE /v1/devices/:id
POST   /v1/devices/:id/heartbeat
POST   /v1/devices/:id/disable
POST   /v1/devices/:id/enable
GET    /v1/devices/:id/config
```

### Chat (REST + WebSocket)
```http
GET    /v1/chat
POST   /v1/chat
PATCH  /v1/chat/:id
DELETE /v1/chat/:id
POST   /v1/chat/:id/redact  # Moderator only
```

```javascript
// WebSocket connection
ws://localhost:8080/ws?token=<jwt_token>

// Send message
{"type": "chat.send", "op_id": "1", "data": {"scope": "host", "body": "Hello!"}}
```

See [docs/API_EXAMPLES.http](../docs/API_EXAMPLES.http) for full examples.

## 🔒 Security

### Password Storage
- **Algorithm**: Argon2id
- **Parameters**: time=3, memory=64MB, threads=4, keyLen=32

### Token Security
- **Access Token**: 15 minutes (default)
- **Refresh Token**: 7 days (default)
- **Algorithm**: HS256 (HMAC-SHA256)

### RBAC Roles
- **Admin**: Full system access, can edit any message
- **Moderator**: Content moderation, can delete/redact messages
- **User**: Standard access, can edit own messages (15min limit)

## 🐛 Troubleshooting

### "JWT_SECRET is required"
Generate a secure key:
```bash
openssl rand -base64 48
```

### "WG_SERVER_PUBKEY must be exactly 44 characters"
Generate WireGuard keys:
```bash
# Generate private key
wg genkey > server_private.key

# Generate public key
wg pubkey < server_private.key > server_public.key

# Use the public key in .env
cat server_public.key
```

### "invalid pubkey format (must be 44 characters)"
Device public keys must be base64-encoded WireGuard keys (32 bytes = 44 chars).

### Database Connection Issues
```bash
# Check PostgreSQL is running
systemctl status postgresql

# Test connection
psql -h localhost -U postgres -d goconnect

# Check environment variables
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME
```

## 📈 Performance

### Recommended Production Settings

```bash
# Server
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=10m

# Audit
AUDIT_ASYNC=true
AUDIT_QUEUE_SIZE=2048
AUDIT_WORKER_COUNT=2
```

### Monitoring
- Prometheus metrics endpoint: `/metrics`
- Health check: `/health`

## 🛠️ Development

### Project Structure
- **Domain-Driven Design**: Clear separation of concerns
- **Repository Pattern**: Swappable data sources (in-memory ↔ PostgreSQL)
- **Service Layer**: Business logic isolated from handlers
- **Middleware**: Authentication, RBAC, CORS, Request ID

### Adding a New Feature

1. **Define domain model**: `internal/domain/`
2. **Create repository interface**: `internal/repository/`
3. **Implement service logic**: `internal/service/`
4. **Add HTTP handlers**: `internal/handler/`
5. **Write tests**: `*_test.go`
6. **Update migrations**: `migrations/`

### Code Quality
```bash
# Format code
gofmt -w .

# Lint
golangci-lint run

# Security scan
gosec ./...
```

## 📝 Migration Status

| File                        | Description                           | Status     |
| --------------------------- | ------------------------------------- | ---------- |
| `000001_initial_schema.sql` | Users, tenants, networks, memberships | ✅ Complete |
| `000002_chat_tables.sql`    | Chat messages, edit history           | ✅ Complete |
| `000003_device_tables.sql`  | Device management                     | ✅ Complete |

## 🗺️ Roadmap

### Phase 1: Foundation ✅ (Current)
- [x] Multi-tenant architecture
- [x] JWT authentication
- [x] Network & IP management
- [x] Device registration
- [x] WireGuard integration
- [x] Chat system
- [x] RBAC (Admin, Moderator)
- [x] Environment configuration
- [x] PostgreSQL migrations

### Phase 2: Production Readiness 🚧
- [ ] PostgreSQL repository implementations
- [ ] Redis caching layer
- [ ] Token revocation (blacklist)
- [ ] Rate limiting
- [ ] Admin dashboard API
- [ ] Comprehensive logging (structured JSON)

### Phase 3: Advanced Features
- [ ] 2FA (TOTP)
- [ ] SSO integration (OIDC)
- [ ] Advanced analytics
- [ ] Multi-region support
- [ ] Kubernetes deployment

## 🤝 Contributing

Contributions welcome! Please read [CONTRIBUTING.md](../CONTRIBUTING.md) first.

## 📄 License

MIT License - see [LICENSE](../LICENSE) for details.

## 🙏 Acknowledgments

- **WireGuard**: Fast, modern VPN protocol
- **Gin**: High-performance HTTP framework
- **PostgreSQL**: Reliable RDBMS
- **Goose**: Database migration tool

---

**Built with ❤️ by [@orhaniscoding](https://github.com/orhaniscoding)**

For questions or support, open an issue or reach out via [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions).
