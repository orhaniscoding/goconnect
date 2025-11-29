# GoConnect Architecture

## Overview

GoConnect follows a **Tenant → Network → Client** hierarchy that mirrors Discord's server/channel structure:

```
TENANT (Organization/Community)
├── Multiple Networks (Virtual LANs)
├── Member Roles & Permissions
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

## Core Components

### 1. Server (Go Backend)
- **Central Management Hub**: Single source of truth for all state
- **REST API**: `/v1/*` endpoints for all operations
- **WebSocket**: Real-time chat and status updates
- **Database**: PostgreSQL (production) / SQLite (development)
- **Authentication**: JWT-based with role-based access control

### 2. Client Daemon (Go Agent)
- **Lightweight Agent**: Runs on user devices
- **WireGuard Integration**: Manages VPN interfaces
- **Multi-Network Support**: Connect to multiple networks simultaneously
- **Auto-Reconnection**: Maintains stable connections
- **Local Bridge**: HTTP server for web UI communication

### 3. Web UI (Next.js)
- **Unified Dashboard**: Single interface for all operations
- **Real-time Updates**: WebSocket integration
- **Cross-Platform**: Works on all modern browsers
- **Responsive Design**: Mobile-friendly interface

## Data Model

### Entities

#### Tenant
```go
type Tenant struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Visibility  string    `json:"visibility"` // public, private
    Access      string    `json:"access"`     // open, password, invite_only
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
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Membership
```go
type Membership struct {
    ID        string    `json:"id"`
    TenantID  string    `json:"tenant_id"`
    NetworkID string    `json:"network_id"`
    UserID    string    `json:"user_id"`
    Role      string    `json:"role"` // owner, admin, moderator, member
    Status    string    `json:"status"` // active, pending, banned
    JoinedAt  time.Time `json:"joined_at"`
}
```

#### Device
```go
type Device struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    Name       string    `json:"name"`
    Platform   string    `json:"platform"` // windows, linux, macos
    PublicKey  string    `json:"public_key"`
    LastSeen   time.Time `json:"last_seen"`
    CreatedAt  time.Time `json:"created_at"`
}
```

## Network Flow

### 1. User Registration & Login
```
User → Web UI → Server API → JWT Token → Web UI Storage
```

### 2. Joining a Network
```
User → Web UI → Server API → Membership Creation → WireGuard Config → Client Daemon
```

### 3. Network Connection
```
Client Daemon → Server API → WireGuard Profile → Local Interface → P2P Connection
```

### 4. Real-time Communication
```
Client/Server → WebSocket → Real-time Updates → Web UI
```

## Security Architecture

### Authentication
- **JWT Tokens**: Stateless authentication with refresh tokens
- **Role-Based Access Control**: Granular permissions per tenant/network
- **Multi-Factor Authentication**: TOTP support (optional)

### Network Security
- **WireGuard Encryption**: End-to-end encrypted tunnels
- **Key Management**: Automatic key rotation and distribution
- **NAT Traversal**: UDP hole punching with fallback relays

### Data Security
- **Tenant Isolation**: Complete data separation between tenants
- **Audit Logging**: Comprehensive audit trail for all operations
- **Rate Limiting**: Per-IP and per-user rate limiting

## Scalability Considerations

### Horizontal Scaling
- **Stateless API**: Multiple server instances behind load balancer
- **Database Sharding**: Tenant-based data partitioning
- **CDN Integration**: Static asset distribution

### Performance Optimization
- **Connection Pooling**: Database connection management
- **Caching Strategy**: Redis for frequently accessed data
- **WebSocket Scaling**: Multiple WebSocket servers with pub/sub

## Deployment Architecture

### Development Environment
```
Local Machine:
├── Server (Go) :8080
├── Client Daemon :12000-13000
├── Web UI (Next.js) :3000
└── SQLite Database
```

### Production Environment
```
Cloud Infrastructure:
├── Load Balancer
├── Server Cluster (Go)
├── PostgreSQL Database
├── Redis Cache
├── WebSocket Servers
└── CDN for Static Assets
```

## API Design Principles

### RESTful Design
- **Resource-Based URLs**: `/v1/tenants/{id}/networks`
- **HTTP Methods**: Proper GET, POST, PATCH, DELETE usage
- **Status Codes**: Consistent HTTP status code responses
- **Error Handling**: Structured error responses

### Real-time Communication
- **WebSocket Endpoints**: `/v1/ws` for real-time updates
- **Event Types**: Chat messages, status changes, network events
- **Authentication**: JWT-based WebSocket authentication

## Monitoring & Observability

### Metrics
- **Application Metrics**: Request rates, response times, error rates
- **Business Metrics**: Active users, network counts, connection status
- **Infrastructure Metrics**: CPU, memory, network usage

### Logging
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Log Levels**: Debug, Info, Warn, Error with appropriate filtering
- **Audit Trail**: Immutable audit logs for security events

### Health Checks
- **Readiness Probes**: Database connectivity, external dependencies
- **Liveness Probes**: Application health and responsiveness
- **Dependency Health**: External service availability checks
