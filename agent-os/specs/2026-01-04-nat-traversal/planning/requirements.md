# Requirements: NAT Traversal

## User Answers Summary

### 1. TURN Server Approach
**Answer: (c) Self-hosted coturn (Docker)**
- Zero cost, full control, privacy-focused
- Docker Compose setup alongside existing services
- Self-hosting aligns with GoConnect philosophy

### 2. Configuration Location
**Answer: (c) Core server provides dynamic credentials**
- No manual config for users
- Server generates time-limited TURN credentials
- API endpoint: `GET /api/v1/networks/:id/ice-config`
- Environment override for self-hosters: `GC_TURN_SERVER`

### 3. Connection Type Display
**Answer: (a) Badge on peer card**
- 🔗 Direct - P2P connection (best)
- 🔄 Relay - Using TURN (slower but works)
- ⏳ Connecting
- ❌ Failed
- Badge background color: green (direct), yellow (relay)

### 4. TURN Credentials Security
**Answer: (a) Time-limited tokens (10 min TTL)**
- HMAC-based credential generation
- Username format: `networkID_timestamp`
- coturn validates with `use-auth-secret`
- No credential storage needed

### 5. Fallback Strategy
**Answer: (a) Automatic (ICE handles it)**
- pion/ICE tries candidates in priority order
- Direct P2P first, then STUN, then TURN
- User sees connection type badge
- Toast notification on relay fallback

### 6. MVP Priority
**Answer: (b) TURN integration + basic config (defer UI polish)**

**In Scope:**
- coturn Docker setup
- TURN credential API
- pion/ICE TURN integration
- Basic connection badges
- Documentation

**Deferred:**
- Advanced UI settings
- Multiple TURN servers
- Connection metrics dashboard

---

## Technical Decisions

### TURN Server: coturn
```yaml
# docker-compose.yml addition
coturn:
  image: coturn/coturn:latest
  ports:
    - "3478:3478/udp"
    - "3478:3478/tcp"
  environment:
    - TURN_SECRET=${TURN_SECRET}
    - REALM=goconnect.io
```

### Credential Generation
```go
// Server-side
func generateTURNCredentials(networkID string) TURNConfig {
    username := fmt.Sprintf("%s_%d", networkID, time.Now().Unix())
    credential := hmacSHA256(username, TURN_SECRET)
    return TURNConfig{
        URL:        "turn:turn.goconnect.io:3478",
        Username:   username,
        Credential: credential,
        TTL:        600, // 10 minutes
    }
}
```

### API Response Format
```json
{
  "stun_servers": ["stun.l.google.com:19302"],
  "turn_server": {
    "url": "turn:turn.goconnect.io:3478",
    "username": "net_123456_1704412800",
    "credential": "abc123hmac",
    "ttl": 600
  }
}
```

---

## Effort Estimate

| Task | Days |
|------|------|
| TURN server setup | 2 |
| Credential API | 2 |
| pion/ICE integration | 3 |
| UI badges | 2 |
| Testing & docs | 3 |
| **Total** | **12 days** |
