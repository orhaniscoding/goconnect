# WebSocket API Documentation

GoConnect provides real-time communication via WebSocket for chat messages, membership events, and network updates.

## Connection

**Endpoint:** `wss://api.goconnect.example/v1/ws` (dev: `ws://localhost:8080/v1/ws`)

**Authentication:** JWT token required via `Authorization` header or query parameter

```javascript
const ws = new WebSocket('ws://localhost:8080/v1/ws', {
  headers: {
    'Authorization': 'Bearer <jwt_token>'
  }
});

// Or via query parameter
const ws = new WebSocket('ws://localhost:8080/v1/ws?token=<jwt_token>');
```

## Message Format

All WebSocket messages follow this structure:

### Inbound (Client → Server)

```json
{
  "type": "message_type",
  "op_id": "unique-operation-id",
  "data": { }
}
```

- **type**: Message type (see below)
- **op_id**: Client-provided operation ID for request/response correlation
- **data**: Message-specific data payload

### Outbound (Server → Client)

```json
{
  "type": "message_type",
  "op_id": "echoed-operation-id",
  "data": { },
  "error": {
    "code": "ERR_CODE",
    "message": "Human readable error",
    "details": {}
  }
}
```

- **op_id**: Echoed from inbound message (for ACKs/errors)
- **error**: Present only for error responses

## Message Types

### Inbound Messages (Client → Server)

#### 1. `auth.refresh` - Refresh Authentication Token

**Payload:**
```json
{
  "type": "auth.refresh",
  "op_id": "op-1",
  "data": {
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Response:**
```json
{
  "type": "ack",
  "op_id": "op-1",
  "data": {
    "access_token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "expires_in": 900,
    "status": "refreshed"
  }
}
```

**Status:** Not implemented (v0.1.0)

#### 2. `chat.send` - Send Chat Message

**Payload:**
```json
{
  "type": "chat.send",
  "op_id": "op-2",
  "data": {
    "scope": "host|network:<network_id>",
    "body": "Message text",
    "attachments": ["url1", "url2"]
  }
}
```

**Parameters:**
- **scope** (required): `"host"` for global chat or `"network:<id>"` for network-specific
- **body** (required): Message text (1-4096 characters)
- **attachments** (optional): Array of attachment URLs

**Response:**
```json
{
  "type": "ack",
  "op_id": "op-2",
  "data": {
    "message_id": "msg-123",
    "status": "sent"
  }
}
```

**Broadcast:** All clients in the same scope receive `chat.message` event

#### 3. `chat.edit` - Edit Chat Message

**Payload:**
```json
{
  "type": "chat.edit",
  "op_id": "op-3",
  "data": {
    "message_id": "msg-123",
    "new_body": "Updated message text"
  }
}
```

**Status:** Not implemented (v0.1.0)

#### 4. `chat.delete` - Delete Chat Message

**Payload:**
```json
{
  "type": "chat.delete",
  "op_id": "op-4",
  "data": {
    "message_id": "msg-123",
    "mode": "soft|hard"
  }
}
```

**Parameters:**
- **mode**: `"soft"` (mark as deleted) or `"hard"` (remove completely)

**Status:** Not implemented (v0.1.0)

#### 5. `chat.redact` - Redact Chat Message (Moderator)

**Payload:**
```json
{
  "type": "chat.redact",
  "op_id": "op-5",
  "data": {
    "message_id": "msg-123",
    "mask": "***"
  }
}
```

**Permissions:** Requires moderator/admin role

**Status:** Not implemented (v0.1.0)

#### 6. `room.join` - Join a Room

**Payload:**
```json
{
  "type": "room.join",
  "op_id": "op-6",
  "data": {
    "room": "network:<id>|host"
  }
}
```

**Response:**
```json
{
  "type": "ack",
  "op_id": "op-6",
  "data": {
    "room": "network:<id>",
    "status": "joined"
  }
}
```

#### 7. `room.leave` - Leave a Room

**Payload:**
```json
{
  "type": "room.leave",
  "op_id": "op-7",
  "data": {
    "room": "network:<id>|host"
  }
}
```

#### 8. `presence.ping` - Keep-Alive Ping

**Payload:**
```json
{
  "type": "presence.ping",
  "op_id": "op-8",
  "data": {}
}
```

**Response:**
```json
{
  "type": "presence.pong",
  "op_id": "op-8",
  "data": {
    "timestamp": "2025-11-21T10:00:00Z"
  }
}
```

### Outbound Messages (Server → Client)

#### 1. `chat.message` - New Chat Message

**Payload:**
```json
{
  "type": "chat.message",
  "data": {
    "id": "msg-123",
    "scope": "network:abc-123",
    "user_id": "user-456",
    "body": "Hello world!",
    "redacted": false,
    "deleted_at": null,
    "attachments": [],
    "created_at": "2025-10-29T12:34:56Z"
  }
}
```

#### 2. `chat.edited` - Message Edited

**Payload:**
```json
{
  "type": "chat.edited",
  "data": {
    "message_id": "msg-123",
    "new_body": "Updated text",
    "edited_at": "2025-10-29T12:35:00Z"
  }
}
```

#### 3. `chat.deleted` - Message Deleted

**Payload:**
```json
{
  "type": "chat.deleted",
  "data": {
    "message_id": "msg-123",
    "mode": "soft",
    "deleted_at": "2025-10-29T12:36:00Z"
  }
}
```

#### 4. `chat.redacted` - Message Redacted

**Payload:**
```json
{
  "type": "chat.redacted",
  "data": {
    "message_id": "msg-123",
    "redaction_mask": "***",
    "redacted_by": "moderator-789"
  }
}
```

#### 5. `presence.update` - User Presence Update

**Payload:**
```json
{
  "type": "presence.update",
  "data": {
    "user_id": "user-123",
    "status": "online|offline",
    "since": "2025-11-21T10:00:00Z"
  }
}
```

#### 5. `member.joined` - Member Joined Network

**Payload:**
```json
{
  "type": "member.joined",
  "data": {
    "network_id": "net-123",
    "user_id": "user-456",
    "role": "member"
  }
}
```

#### 6. `member.left` - Member Left Network

**Payload:**
```json
{
  "type": "member.left",
  "data": {
    "network_id": "net-123",
    "user_id": "user-456"
  }
}
```

#### 7. `request.join.pending` - Join Request Pending Approval

**Payload:**
```json
{
  "type": "request.join.pending",
  "data": {
    "network_id": "net-123",
    "user_id": "user-789",
    "request_id": "req-456"
  }
}
```

**Recipients:** Network admins/owners

#### 8. `admin.kick` - Member Kicked

**Payload:**
```json
{
  "type": "admin.kick",
  "data": {
    "network_id": "net-123",
    "user_id": "user-456",
    "reason": "Violation of rules"
  }
}
```

#### 9. `admin.ban` - Member Banned

**Payload:**
```json
{
  "type": "admin.ban",
  "data": {
    "network_id": "net-123",
    "user_id": "user-456",
    "reason": "Permanent ban"
  }
}
```

#### 10. `net.updated` - Network Settings Updated

**Payload:**
```json
{
  "type": "net.updated",
  "data": {
    "network_id": "net-123",
    "changes": ["name", "visibility"],
    "updated_by": "user-owner",
    "properties": {
      "name": "New Network Name",
      "visibility": "public"
    }
  }
}
```

#### 11. `error` - Error Response

**Payload:**
```json
{
  "type": "error",
  "op_id": "op-1",
  "error": {
    "code": "ERR_INVALID_MESSAGE",
    "message": "Failed to parse message",
    "details": {
      "field": "data.body"
    }
  }
}
```

#### 12. `ack` - Acknowledgment

**Payload:**
```json
{
  "type": "ack",
  "op_id": "op-2",
  "data": {
    "status": "success"
  }
}
```

## Room/Scope Subscription

Clients are automatically subscribed to rooms based on their permissions:

- **`host`**: Global room (all authenticated users)
- **`tenant:<tenant_id>`**: Tenant-wide broadcasts
- **`network:<network_id>`**: Network-specific events (auto-join on network membership)

## Error Codes

| Code                  | Description                       | Retry             |
| --------------------- | --------------------------------- | ----------------- |
| `ERR_UNAUTHORIZED`    | Missing or invalid authentication | No                |
| `ERR_INVALID_MESSAGE` | Malformed message payload         | No                |
| `ERR_HANDLER_FAILED`  | Message handler error             | Yes               |
| `ERR_FORBIDDEN`       | Insufficient permissions          | No                |
| `ERR_NOT_FOUND`       | Resource not found                | No                |
| `ERR_RATE_LIMIT`      | Too many requests                 | Yes (after delay) |

## Connection Lifecycle

### 1. Connect
```javascript
const ws = new WebSocket('ws://localhost:8080/v1/ws?token=<jwt>');

ws.onopen = () => {
  console.log('Connected to WebSocket');
};
```

### 2. Send Messages
```javascript
ws.send(JSON.stringify({
  type: 'chat.send',
  op_id: 'op-' + Date.now(),
  data: {
    scope: 'host',
    body: 'Hello everyone!'
  }
}));
```

### 3. Receive Messages
```javascript
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case 'chat.message':
      console.log('New message:', message.data);
      break;
    case 'ack':
      console.log('ACK for:', message.op_id);
      break;
    case 'error':
      console.error('Error:', message.error);
      break;
  }
};
```

### 4. Handle Disconnection
```javascript
ws.onclose = (event) => {
  console.log('Disconnected:', event.code, event.reason);
  // Reconnect with exponential backoff
  setTimeout(() => reconnect(), 1000);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

## Best Practices

### 1. Operation IDs
Use unique operation IDs for request/response correlation:
```javascript
const opId = `op-${Date.now()}-${Math.random()}`;
```

### 2. Reconnection Strategy
Implement exponential backoff:
```javascript
let retryDelay = 1000;
const maxDelay = 30000;

function reconnect() {
  setTimeout(() => {
    connect();
    retryDelay = Math.min(retryDelay * 2, maxDelay);
  }, retryDelay);
}
```

### 3. Message Queuing
Queue messages during disconnection:
```javascript
const messageQueue = [];

function send(message) {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
  } else {
    messageQueue.push(message);
  }
}

ws.onopen = () => {
  while (messageQueue.length > 0) {
    ws.send(JSON.stringify(messageQueue.shift()));
  }
};
```

### 4. Heartbeat/Ping-Pong
Server sends ping every 54 seconds. Client receives pong automatically (handled by browser).

### 5. Rate Limiting
Respect rate limits:
- Max 100 messages/minute per client
- Max 1000 messages/minute per network

## Configuration

### Server Settings

```bash
# Connection limits
WS_MAX_CONNECTIONS=10000
WS_MAX_CONNECTIONS_PER_USER=5

# Message limits
WS_MAX_MESSAGE_SIZE=512KB
WS_SEND_BUFFER_SIZE=256

# Timeouts
WS_WRITE_WAIT=10s
WS_PONG_WAIT=60s
WS_PING_PERIOD=54s
```

## Monitoring

### Metrics (Prometheus)

- `websocket_connections_total`: Total active connections
- `websocket_messages_sent_total`: Messages sent by type
- `websocket_messages_received_total`: Messages received by type
- `websocket_errors_total`: Errors by type
- `websocket_rooms_total`: Active rooms

### Health Check

```bash
curl http://localhost:8080/health
```

Response includes WebSocket stats:
```json
{
  "ok": true,
  "websocket": {
    "connections": 42,
    "rooms": 15
  }
}
```

## Examples

See [examples/websocket/](../../examples/websocket/) for complete client implementations:
- `simple_chat.html` - Browser-based chat client
- `node_client.js` - Node.js WebSocket client
- `go_client.go` - Go WebSocket client

## Troubleshooting

### Connection Refused
- Check JWT token validity
- Verify server is running
- Check CORS settings

### Messages Not Received
- Verify room subscription
- Check message scope matches
- Inspect network tab in browser DevTools

### Frequent Disconnections
- Check network stability
- Increase timeouts if on slow connection
- Verify server capacity

## See Also

- [TECH_SPEC.md](./TECH_SPEC.md) - Technical specification
- [API_EXAMPLES.http](./API_EXAMPLES.http) - REST API examples
- [RUNBOOKS.md](./RUNBOOKS.md) - Operational guides
