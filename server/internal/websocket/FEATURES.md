# WebSocket Real-time Communication Features

## Overview
GoConnect WebSocket sistemi, gerçek zamanlı iletişim için kapsamlı bir altyapı sağlar. Hub-based mimari ile room subscription, presence tracking ve broadcast özellikleri sunar.

## ✅ Implemented Features

### 1. **Room Management**
- ✅ Dynamic room creation/deletion
- ✅ Client room subscription (`room.join`)
- ✅ Client room unsubscription (`room.leave`)
- ✅ Multi-room support (bir client birden fazla room'da bulunabilir)
- ✅ Automatic room cleanup (boş room'lar otomatik silinir)

### 2. **Chat Features**
- ✅ Real-time message sending (`chat.send`)
- ✅ Message editing with history (`chat.edit`)
- ✅ Message deletion (soft/hard modes) (`chat.delete`)
- ✅ Message redaction (moderation) (`chat.redact`)
- ✅ **Typing indicators** (`chat.typing` → `chat.typing.user`) ⭐ YENİ
- ✅ File attachments support
- ✅ Scope-based routing (host, network:id)

### 3. **Presence System**
- ✅ **Client activity tracking** (`lastActivity` timestamp) ⭐ YENİ
- ✅ **Ping/Pong keep-alive** (`presence.ping` → `presence.pong`) ⭐ YENİ
- ✅ Auto-disconnect on inactivity (60s timeout)
- ✅ Graceful connection handling

### 4. **Broadcast System**
- ✅ Room-specific broadcasts
- ✅ Global broadcasts (all clients)
- ✅ Selective exclusion (exclude sender from broadcast)
- ✅ Buffer overflow handling

### 5. **Security & Multi-tenancy**
- ✅ JWT-based authentication
- ✅ Tenant isolation
- ✅ Admin/Moderator role support
- ✅ User ID tracking per connection

### 6. **Performance & Scalability**
- ✅ Non-blocking message handling
- ✅ Configurable send buffer (256 messages)
- ✅ Concurrent operation support
- ✅ Thread-safe room management
- ✅ Message size limit (512 KB)

## 📋 Message Types

### Inbound (Client → Server)
```
auth.refresh      - Token yenileme
chat.send         - Mesaj gönderme
chat.edit         - Mesaj düzenleme
chat.delete       - Mesaj silme
chat.redact       - Mesaj redact etme (moderasyon)
chat.typing       - Yazıyor göstergesi ⭐ YENİ
room.join         - Room'a katılma ⭐ YENİ
room.leave        - Room'dan ayrılma ⭐ YENİ
presence.ping     - Keep-alive ping ⭐ YENİ
```

### Outbound (Server → Client)
```
chat.message         - Yeni mesaj bildirimi
chat.edited          - Mesaj düzenlendi bildirimi
chat.deleted         - Mesaj silindi bildirimi
chat.redacted        - Mesaj redact edildi bildirimi
chat.typing.user     - Kullanıcı yazıyor göstergesi ⭐ YENİ
member.joined        - Üye katıldı
member.left          - Üye ayrıldı
request.join.pending - Join isteği beklemede
request.join.approved - Join isteği onaylandı ⭐ YENİ
request.join.denied   - Join isteği reddedildi ⭐ YENİ
admin.kick           - Admin kick bildirimi
admin.ban            - Admin ban bildirimi
net.updated          - Network güncellendi
device.online        - Cihaz online oldu ⭐ YENİ
device.offline       - Cihaz offline oldu ⭐ YENİ
presence.pong        - Ping yanıtı ⭐ YENİ
presence.update      - Presence durumu güncellendi ⭐ YENİ
error                - Hata mesajı
ack                  - İstek onayı
```

## 🎯 Usage Examples

### Room Subscription
```json
// Client -> Server
{
  "type": "room.join",
  "op_id": "req-123",
  "data": {
    "room": "network:abc-123"
  }
}

// Server -> Client (Acknowledgment)
{
  "type": "ack",
  "op_id": "req-123",
  "data": {
    "room": "network:abc-123",
    "status": "joined"
  }
}
```

### Typing Indicator
```json
// Client -> Server (Started typing)
{
  "type": "chat.typing",
  "op_id": "typing-1",
  "data": {
    "scope": "network:abc-123",
    "typing": true
  }
}

// Server -> Other Clients in Room
{
  "type": "chat.typing.user",
  "data": {
    "scope": "network:abc-123",
    "user_id": "user-456",
    "typing": true
  }
}
```

### Presence Ping
```json
// Client -> Server (Every 30s)
{
  "type": "presence.ping",
  "op_id": "ping-789"
}

// Server -> Client
{
  "type": "presence.pong",
  "op_id": "ping-789",
  "data": {
    "timestamp": "2025-11-20T15:30:00Z"
  }
}
```

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
│ (WebSocket) │
└──────┬──────┘
       │
       ├─► ReadPump ──► Hub.handleInbound ──► MessageHandler
       │                                              │
       │                                              ▼
       │                                     ┌────────────────┐
       │                                     │ Business Logic │
       │                                     │ (ChatService)  │
       │                                     └───────┬────────┘
       │                                             │
       ▼                                             ▼
   WritePump ◄──── Hub.broadcast ◄──────────── Broadcast
       │                                         to Rooms
       │
       ▼
   WebSocket
```

## 📊 Metrics & Monitoring

- **Active Connections**: `Hub.GetClientCount()`
- **Active Rooms**: `Hub.GetRoomCount()`
- **Room Members**: `Hub.GetRoomClients(room)`
- **Client Activity**: `Client.GetLastActivity()`

## 🔒 Security Considerations

1. **Authentication**: JWT token zorunlu
2. **Authorization**: Room access validation (network membership check)
3. **Rate Limiting**: Message throttling (10 msg/s, burst 20)
4. **Message Size**: 512 KB limit
5. **Connection Timeout**: 60s inactivity timeout

## 🚀 Next Steps

### High Priority
- [x] Rate limiting per client
- [x] Room access validation (network membership check)
- [x] Presence status broadcast (online/away/offline)
- [x] Device online/offline events integration
- [x] Network join/leave events broadcast

### Medium Priority
- [x] Message read receipts
- [x] User status (online, away, busy, offline)
- [x] Direct messages (user-to-user)
- [x] Voice/Video call signaling

### Low Priority
- [x] Message reactions
- [x] Message threads
- [x] File upload progress
- [x] Screen sharing signaling

## 🧪 Testing

Tüm core özellikler için comprehensive test coverage:
- ✅ Hub operations (register, unregister, broadcast)
- ✅ Room management (join, leave, multi-room)
- ✅ Message handling (all message types)
- ✅ Concurrent operations
- ✅ Error handling
- ✅ Buffer overflow scenarios

```bash
go test ./internal/websocket/... -v
```

## 📝 Integration Example

```go
// Create hub and handler
hub := websocket.NewHub(handler)
go hub.Run(ctx)

// Register WebSocket endpoint
r.GET("/v1/ws", func(c *gin.Context) {
    // Upgrade connection
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    
    // Extract user info from JWT
    userID := c.GetString("user_id")
    tenantID := c.GetString("tenant_id")
    isAdmin := c.GetBool("is_admin")
    
    // Create and register client
    client := websocket.NewClient(hub, conn, userID, tenantID, isAdmin, false)
    hub.Register(client)
    client.Run(ctx)
})
```

---

**Status**: ✅ Production Ready (with PostgreSQL integration)
**Version**: 2.5.0
**Last Updated**: November 2025
