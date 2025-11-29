# GoConnect — Technical Specification

## Product Vision

GoConnect is a **cross-platform virtual LAN platform** that makes internet users appear as if they're on the same local network, with a Discord-like community structure.

**Core Philosophy**: "If you can join a Discord server, you can join a GoConnect network."

## Architecture Overview

### Hierarchy Model
```
TENANT (Organization/Community)
├── Multiple Networks (Virtual LANs)
├── Member Roles (Owner, Admin, Moderator, Member)
├── Community Chat
└── Invite System

NETWORK (Virtual LAN)
├── WireGuard Mesh Networking
├── IP Address Management
├── Network-Specific Chat
└── Member Access Control

CLIENT/DAEMON (User Device)
├── WireGuard Integration
├── Multi-Network Support
└── Auto-Connection Management
```

### Components

#### 1. Server (Go Backend)
- **Purpose**: Central management hub and single source of truth
- **Technology**: Go 1.24+ with Gin framework
- **Database**: PostgreSQL (production), SQLite (development)
- **API**: RESTful JSON API with OpenAPI 3.0 specification
- **Real-time**: WebSocket for chat and status updates

#### 2. Client Daemon (Go Agent)
- **Purpose**: Lightweight agent running on user devices
- **Technology**: Go 1.24+ with minimal dependencies
- **Networking**: WireGuard integration for VPN functionality
- **Features**: Multi-network support, auto-reconnection
- **Platform**: Windows, Linux, macOS (mobile later)

#### 3. Web UI (Next.js)
- **Purpose**: Unified dashboard for all user interactions
- **Technology**: Next.js 15 with TypeScript and Tailwind CSS
- **Features**: Real-time updates, responsive design
- **Accessibility**: WCAG 2.1 compliant with i18n support

## Data Model

### Core Entities

#### Tenant
```go
type Tenant struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Visibility  string    `json:"visibility" db:"visibility"` // public, private
    Access      string    `json:"access" db:"access"`         // open, password, invite_only
    Settings    JSONB     `json:"settings" db:"settings"`     // tenant-specific settings
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

#### Network
```go
type Network struct {
    ID           string    `json:"id" db:"id"`
    TenantID     string    `json:"tenant_id" db:"tenant_id"`
    Name         string    `json:"name" db:"name"`
    Description  string    `json:"description" db:"description"`
    CIDR         string    `json:"cidr" db:"cidr"`
    Visibility   string    `json:"visibility" db:"visibility"`   // public, private
    JoinPolicy   string    `json:"join_policy" db:"join_policy"`   // open, invite, approval
    RequiredRole string    `json:"required_role" db:"required_role"` // member, admin, owner
    Settings     JSONB     `json:"settings" db:"settings"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
```

#### User
```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Email     string    `json:"email" db:"email"`
    Password  string    `json:"-" db:"password_hash"` // hashed
    Settings  JSONB     `json:"settings" db:"settings"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

#### Membership
```go
type Membership struct {
    ID        string    `json:"id" db:"id"`
    TenantID  string    `json:"tenant_id" db:"tenant_id"`
    NetworkID *string   `json:"network_id" db:"network_id"` // nil for tenant-level membership
    UserID    string    `json:"user_id" db:"user_id"`
    Role      string    `json:"role" db:"role"` // owner, admin, moderator, member
    Status    string    `json:"status" db:"status"` // active, pending, banned
    JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}
```

#### Device
```go
type Device struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"user_id" db:"user_id"`
    Name        string    `json:"name" db:"name"`
    Platform    string    `json:"platform" db:"platform"` // windows, linux, macos
    PublicKey   string    `json:"public_key" db:"public_key"`
    PrivateKey  string    `json:"-" db:"private_key"` // encrypted at rest
    LastSeen    time.Time `json:"last_seen" db:"last_seen"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

#### Peer
```go
type Peer struct {
    ID          string    `json:"id" db:"id"`
    NetworkID   string    `json:"network_id" db:"network_id"`
    DeviceID    string    `json:"device_id" db:"device_id"`
    IPAddress   string    `json:"ip_address" db:"ip_address"`
    PublicKey   string    `json:"public_key" db:"public_key"`
    Endpoint    *string   `json:"endpoint" db:"endpoint"` // dynamic endpoint
    LastSeen    time.Time `json:"last_seen" db:"last_seen"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

## API Design

### REST Endpoints

#### Authentication
```
POST   /v1/auth/register
POST   /v1/auth/login
POST   /v1/auth/refresh
POST   /v1/auth/logout
GET    /v1/auth/me
```

#### Tenants
```
GET    /v1/tenants
POST   /v1/tenants
GET    /v1/tenants/:id
PATCH  /v1/tenants/:id
DELETE /v1/tenants/:id
```

#### Networks
```
GET    /v1/tenants/:tenant_id/networks
POST   /v1/tenants/:tenant_id/networks
GET    /v1/networks/:id
PATCH  /v1/networks/:id
DELETE /v1/networks/:id
```

#### Memberships
```
GET    /v1/tenants/:tenant_id/members
POST   /v1/tenants/:tenant_id/members
GET    /v1/networks/:id/members
POST   /v1/networks/:id/join
POST   /v1/networks/:id/approve
POST   /v1/networks/:id/deny
POST   /v1/networks/:id/kick
POST   /v1/networks/:id/ban
```

#### Devices
```
GET    /v1/devices
POST   /v1/devices
GET    /v1/devices/:id
PATCH  /v1/devices/:id
DELETE /v1/devices/:id
```

#### Peers
```
GET    /v1/networks/:id/peers
POST   /v1/networks/:id/peers
GET    /v1/peers/:id/config
DELETE /v1/peers/:id
```

#### Chat
```
GET    /v1/tenants/:tenant_id/messages
POST   /v1/tenants/:tenant_id/messages
GET    /v1/networks/:id/messages
POST   /v1/networks/:id/messages
```

### WebSocket API

#### Connection
```
WS     /v1/ws?token=<jwt_token>
```

#### Message Types
```json
{
  "type": "chat_message",
  "data": {
    "id": "msg_id",
    "tenant_id": "tenant_id",
    "network_id": "network_id",
    "user_id": "user_id",
    "content": "Hello!",
    "timestamp": "2025-01-01T00:00:00Z"
  }
}
```

```json
{
  "type": "peer_status",
  "data": {
    "network_id": "network_id",
    "peer_id": "peer_id",
    "status": "online|offline",
    "endpoint": "ip:port"
  }
}
```

## Security Architecture

### Authentication
- **JWT Tokens**: Stateless authentication with refresh tokens
- **Password Hashing**: bcrypt with cost factor 12
- **Session Management**: Secure cookie-based sessions
- **2FA Support**: TOTP-based two-factor authentication

### Authorization
- **Role-Based Access Control**: Granular permissions
- **Tenant Isolation**: Complete data separation
- **Network-Level Permissions**: Role-based network access
- **API Rate Limiting**: Per-user and per-IP limits

### Network Security
- **WireGuard Encryption**: End-to-end encrypted tunnels
- **Key Management**: Automatic key rotation
- **NAT Traversal**: UDP hole punching with relay fallback
- **Peer Authentication**: Cryptographic peer verification

### Data Protection
- **Encryption at Rest**: Sensitive data encrypted in database
- **Audit Logging**: Comprehensive audit trail
- **Data Minimization**: Only collect necessary data
- **GDPR Compliance**: Right to deletion and data export

## WireGuard Integration

### Network Management
- **Interface Creation**: Automatic WireGuard interface setup
- **Configuration Generation**: Dynamic config file creation
- **Peer Discovery**: Automatic peer endpoint detection
- **Route Management**: Automatic routing table updates

### IP Address Management
- **CIDR Allocation**: Automatic network range assignment
- **IP Assignment**: Dynamic IP address allocation
- **Conflict Detection**: Prevent IP address conflicts
- **Reservation System**: VIP IP address reservation

### Connection Management
- **Health Monitoring**: Peer connection status tracking
- **Auto-Reconnection**: Automatic reconnection on failure
- **Load Balancing**: Optimal peer selection
- **Fallback Relays**: Relay servers for NAT traversal

## Database Schema

### Tables
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    visibility TEXT NOT NULL DEFAULT 'private',
    access TEXT NOT NULL DEFAULT 'invite_only',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE networks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    cidr TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    join_policy TEXT NOT NULL DEFAULT 'invite',
    required_role TEXT NOT NULL DEFAULT 'member',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    network_id UUID REFERENCES networks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, network_id, user_id)
);

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    platform TEXT NOT NULL,
    public_key TEXT UNIQUE NOT NULL,
    private_key_encrypted TEXT NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    ip_address INET UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    endpoint TEXT,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(network_id, device_id)
);
```

### Indexes
```sql
CREATE INDEX idx_networks_tenant_id ON networks(tenant_id);
CREATE INDEX idx_memberships_tenant_id ON memberships(tenant_id);
CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_network_id ON memberships(network_id);
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_peers_network_id ON peers(network_id);
CREATE INDEX idx_peers_device_id ON peers(device_id);
```

## Performance Considerations

### Scalability
- **Horizontal Scaling**: Stateless API design
- **Database Pooling**: Connection pool management
- **Caching Strategy**: Redis for frequently accessed data
- **CDN Integration**: Static asset distribution

### Optimization
- **Query Optimization**: Efficient database queries
- **Connection Management**: Persistent connections
- **Compression**: Response compression
- **Caching Headers**: Browser caching optimization

### Monitoring
- **Application Metrics**: Request rates, response times
- **Business Metrics**: Active users, network counts
- **Infrastructure Metrics**: CPU, memory, network usage
- **Error Tracking**: Comprehensive error monitoring

## Deployment Architecture

### Development
```
Local Environment:
├── Go Server :8080
├── Next.js Dev Server :3000
├── Client Daemon :12000-13000
├── SQLite Database
└── File-based Configuration
```

### Production
```
Cloud Infrastructure:
├── Load Balancer (HTTPS termination)
├── Go Server Cluster (auto-scaling)
├── PostgreSQL Database (HA setup)
├── Redis Cluster (caching)
├── WebSocket Servers
├── CDN (static assets)
└── Monitoring Stack
```

### Container Deployment
```dockerfile
# Server
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o goconnect-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates wireguard-tools
WORKDIR /root/
COPY --from=builder /app/goconnect-server .
EXPOSE 8080
CMD ["./goconnect-server"]
```

## Testing Strategy

### Unit Testing
- **Go Tests**: 80%+ coverage requirement
- **JavaScript Tests**: Jest for frontend components
- **Mock Services**: External service mocking
- **Property Testing**: Edge case validation

### Integration Testing
- **API Tests**: End-to-end API validation
- **Database Tests**: Migration and schema validation
- **WebSocket Tests**: Real-time communication testing
- **WireGuard Tests**: VPN functionality validation

### Performance Testing
- **Load Testing**: Concurrent user simulation
- **Stress Testing**: System limit identification
- **Network Testing**: VPN performance measurement
- **Database Testing**: Query performance analysis

## Development Workflow

### Code Organization
```
goconnect/
├── server/                 # Go backend
│   ├── cmd/               # Application entry points
│   ├── internal/          # Private application code
│   │   ├── handler/       # HTTP handlers
│   │   ├── service/       # Business logic
│   │   ├── repository/    # Data access layer
│   │   └── domain/        # Domain models
│   ├── migrations/        # Database migrations
│   └── openapi/          # API specification
├── client-daemon/         # Go client agent
│   ├── cmd/              # Daemon entry point
│   ├── internal/         # Private application code
│   └── service/          # Client services
├── web-ui/               # Next.js frontend
│   ├── src/              # React components
│   ├── pages/            # Next.js pages
│   └── public/           # Static assets
└── docs/                 # Documentation
```

### Build Process
```bash
# Server
cd server && go build -o goconnect-server ./cmd/server

# Client Daemon
cd client-daemon && go build -o goconnect-daemon ./cmd/daemon

# Web UI
cd web-ui && npm run build
```

### Quality Assurance
- **Code Review**: All changes require review
- **Automated Testing**: CI/CD pipeline validation
- **Linting**: Go and JavaScript linting
- **Security Scanning**: Dependency vulnerability scanning

---

This specification serves as the authoritative technical reference for GoConnect development and implementation.
