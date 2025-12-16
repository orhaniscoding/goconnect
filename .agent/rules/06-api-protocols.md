# API & Protocol Guidelines

This document defines standards for APIs and communication protocols in GoConnect.

## REST API Design

### URL Structure
```
GET    /api/v1/networks              # List networks
POST   /api/v1/networks              # Create network
GET    /api/v1/networks/{id}         # Get network
PUT    /api/v1/networks/{id}         # Update network
DELETE /api/v1/networks/{id}         # Delete network
POST   /api/v1/networks/{id}/join    # Join network (action)
```

### Naming Conventions
- Use plural nouns: `/networks`, `/users`, `/peers`
- Use kebab-case: `/network-configs`, `/ip-addresses`
- Nest related resources: `/networks/{id}/peers`
- Use verbs for actions: `/networks/{id}/join`

### Response Format
```json
// Success (200, 201)
{
  "data": {
    "id": "net_123",
    "name": "My Network",
    "created_at": "2024-01-15T10:30:00Z"
  }
}

// List (200)
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150
  }
}

// Error (4xx, 5xx)
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid network name",
    "details": [
      {"field": "name", "message": "must be at least 3 characters"}
    ]
  }
}
```

### HTTP Status Codes
| Code | Usage |
|------|-------|
| 200 | Success (GET, PUT) |
| 201 | Created (POST) |
| 204 | No Content (DELETE) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

---

## gRPC/Protocol Buffers

### Proto File Location
- Definitions: `core/proto/`
- Generated code: `cli/internal/proto/`

### Proto Style Guide
```protobuf
syntax = "proto3";

package goconnect.v1;

option go_package = "github.com/orhaniscoding/goconnect/cli/internal/proto";

// Use PascalCase for message names
message Network {
  string id = 1;
  string name = 2;
  NetworkStatus status = 3;
  google.protobuf.Timestamp created_at = 4;
}

// Use UPPER_SNAKE_CASE for enum values
enum NetworkStatus {
  NETWORK_STATUS_UNSPECIFIED = 0;
  NETWORK_STATUS_ACTIVE = 1;
  NETWORK_STATUS_OFFLINE = 2;
}

// Service definitions
service NetworkService {
  // Use imperative verb for RPC names
  rpc CreateNetwork(CreateNetworkRequest) returns (CreateNetworkResponse);
  rpc ListNetworks(ListNetworksRequest) returns (ListNetworksResponse);
  rpc JoinNetwork(JoinNetworkRequest) returns (JoinNetworkResponse);
}
```

### Generating Proto Code
```bash
cd core
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/*.proto

# Copy to cli
cp proto/*.pb.go ../cli/internal/proto/
```

---

## WebSocket Protocol

### Connection Lifecycle
```
1. Client connects to /ws with auth token
2. Server validates and sends CONNECTED event
3. Bidirectional message exchange
4. Keep-alive pings every 30 seconds
5. Graceful disconnect or timeout
```

### Message Format
```json
// Client -> Server
{
  "type": "SUBSCRIBE",
  "payload": {
    "channel": "network:net_123"
  },
  "request_id": "req_abc"
}

// Server -> Client
{
  "type": "EVENT",
  "payload": {
    "event": "PEER_JOINED",
    "data": {
      "peer_id": "peer_456",
      "name": "John's Laptop"
    }
  },
  "channel": "network:net_123"
}

// Error
{
  "type": "ERROR",
  "payload": {
    "code": "INVALID_CHANNEL",
    "message": "Channel not found"
  },
  "request_id": "req_abc"
}
```

### Event Types
```go
const (
    EventPeerJoined    = "PEER_JOINED"
    EventPeerLeft      = "PEER_LEFT"
    EventNetworkUpdate = "NETWORK_UPDATE"
    EventMessageNew    = "MESSAGE_NEW"
    EventTransferStart = "TRANSFER_START"
)
```

---

## IPC (Desktop ↔ Daemon)

### Platform-Specific Channels
| Platform | Mechanism |
|----------|-----------|
| Linux/macOS | Unix Domain Socket (`/var/run/goconnect.sock`) |
| Windows | Named Pipe (`\\.\pipe\goconnect`) |

### Protocol
Uses gRPC over the platform-specific transport:

```go
// Daemon (server)
listener, _ := net.Listen("unix", "/var/run/goconnect.sock")
grpcServer := grpc.NewServer()
proto.RegisterDaemonServiceServer(grpcServer, daemonService)
grpcServer.Serve(listener)

// CLI/Desktop (client)
conn, _ := grpc.Dial(
    "unix:///var/run/goconnect.sock",
    grpc.WithInsecure(),
)
client := proto.NewDaemonServiceClient(conn)
```

### IPC Service Definition
```protobuf
service DaemonService {
  rpc GetStatus(Empty) returns (StatusResponse);
  rpc ListNetworks(Empty) returns (ListNetworksResponse);
  rpc JoinNetwork(JoinRequest) returns (JoinResponse);
  rpc LeaveNetwork(LeaveRequest) returns (Empty);
  rpc StreamEvents(Empty) returns (stream Event);
}
```

---

## Versioning Strategy

### API Versioning
- URL-based: `/api/v1/`, `/api/v2/`
- Major version for breaking changes
- Deprecation notice 6 months before removal

### Proto Versioning
- Package includes version: `goconnect.v1`
- Add new fields (don't remove/rename)
- Deprecate with `[deprecated = true]`

### Compatibility Rules
```go
// ✅ Backward compatible changes
- Adding new optional fields
- Adding new endpoints
- Adding new enum values

// ❌ Breaking changes (require version bump)
- Removing fields
- Changing field types
- Renaming fields
- Changing URL structure
```
