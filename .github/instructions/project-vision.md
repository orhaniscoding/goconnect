# GoConnect Project Vision & Architecture

## Core Vision

GoConnect is a **cross-platform virtual LAN platform** that makes internet users appear as if they're on the same local network, with a Discord-like community structure.

**Core Philosophy**: "If you can join a Discord server, you can join a GoConnect network."

## Product Concept

### Target Use Cases
- **Gaming**: Minecraft LAN, older LAN-only games, multiplayer sessions
- **File Sharing**: Easy peer-to-peer file transfer
- **Remote Access**: Secure access to home/work resources
- **Development**: Local development environments across teams

### User Experience Principles
1. **Zero Configuration**: Paste server URL, login, connect
2. **Visual Feedback**: Real-time status, online indicators
3. **Cross-Platform**: Windows, Linux, macOS (mobile later)
4. **Free Core**: Basic networking always free

## Architecture Model

### Hierarchy: Tenant → Network → Client

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
- **API**: RESTful JSON API with WebSocket support
- **Features**: Authentication, network management, chat, audit logging

#### 2. Client Daemon (Go Agent)
- **Purpose**: Lightweight agent running on user devices
- **Technology**: Go 1.24+ with minimal dependencies
- **Networking**: WireGuard integration for VPN functionality
- **Features**: Multi-network support, auto-reconnection, local bridge
- **Platform**: Windows, Linux, macOS (mobile later)

#### 3. Web UI (Next.js)
- **Purpose**: Unified dashboard for all user interactions
- **Technology**: Next.js 15 with TypeScript and Tailwind CSS
- **Features**: Real-time updates, responsive design, chat interface
- **Accessibility**: WCAG 2.1 compliant with i18n support

## Data Model

### Core Entities

#### Tenant
```go
type Tenant struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Visibility  string    `json:"visibility"` // public, private
    Access      string    `json:"access"`     // open, password, invite_only
    Settings    JSONB     `json:"settings"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Network
```go
type Network struct {
    ID           string    `json:"id"`
    TenantID     string    `json:"tenant_id"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    CIDR         string    `json:"cidr"`
    Visibility   string    `json:"visibility"`   // public, private
    JoinPolicy   string    `json:"join_policy"`   // open, invite, approval
    RequiredRole string    `json:"required_role"` // member, admin, owner
    Settings     JSONB     `json:"settings"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### User
```go
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Password  string    `json:"-"` // hashed
    Settings  JSONB     `json:"settings"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Membership
```go
type Membership struct {
    ID        string    `json:"id"`
    TenantID  string    `json:"tenant_id"`
    NetworkID *string   `json:"network_id"` // nil for tenant-level
    UserID    string    `json:"user_id"`
    Role      string    `json:"role"` // owner, admin, moderator, member
    Status    string    `json:"status"` // active, pending, banned
    JoinedAt  time.Time `json:"joined_at"`
}
```

#### Device
```go
type Device struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    Name        string    `json:"name"`
    Platform    string    `json:"platform"` // windows, linux, macos
    PublicKey   string    `json:"public_key"`
    PrivateKey  string    `json:"-"` // encrypted at rest
    LastSeen    time.Time `json:"last_seen"`
    CreatedAt   time.Time `json:"created_at"`
}
```

#### Peer
```go
type Peer struct {
    ID          string    `json:"id"`
    NetworkID   string    `json:"network_id"`
    DeviceID    string    `json:"device_id"`
    IPAddress   string    `json:"ip_address"`
    PublicKey   string    `json:"public_key"`
    Endpoint    *string   `json:"endpoint"` // dynamic
    LastSeen    time.Time `json:"last_seen"`
    CreatedAt   time.Time `json:"created_at"`
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

#### Devices & Peers
```
GET    /v1/devices
POST   /v1/devices
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
```
WS     /v1/ws?token=<jwt_token>
```

## Security Architecture

### Authentication
- **JWT Tokens**: Stateless authentication with refresh tokens
- **Password Hashing**: bcrypt with cost factor 12
- **2FA Support**: TOTP-based two-factor authentication
- **Session Management**: Secure cookie-based sessions

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

## Development Guidelines

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

### Quality Standards
- **Testing**: 80%+ coverage requirement
- **Code Review**: All changes require review
- **Linting**: Go and JavaScript linting
- **Security**: Dependency vulnerability scanning

### Build Process
```bash
# Development
make dev-server    # Start Go server
make dev-daemon    # Start client daemon
make dev-web       # Start web UI

# Production
make build         # Build all components
make test          # Run all tests
make lint          # Run linters
```

## Deployment Architecture

### Development Environment
```
Local Machine:
├── Go Server :8080
├── Next.js Dev Server :3000
├── Client Daemon :12000-13000
├── SQLite Database
└── File-based Configuration
```

### Production Environment
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

## Key Features

### Core Functionality
- **Multi-Tenant Support**: Multiple organizations per instance
- **Network Management**: Create and manage virtual LANs
- **Real-time Chat**: Tenant and network-level chat
- **Device Management**: Register and manage user devices
- **Invite System**: Easy onboarding with invite links/codes

### Advanced Features
- **NAT Traversal**: Automatic peer discovery and connection
- **Relay Support**: Fallback relays for difficult network scenarios
- **Audit Logging**: Comprehensive activity tracking
- **Metrics & Monitoring**: Prometheus-compatible metrics
- **API Documentation**: OpenAPI 3.0 specification

---

This vision document serves as the authoritative reference for GoConnect development, ensuring all components align with the core product vision of "Discord for networks."
