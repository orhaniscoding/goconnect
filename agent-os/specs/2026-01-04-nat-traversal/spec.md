# Specification: NAT Traversal

## Goal
Enable reliable P2P connections through firewalls and NATs by integrating TURN relay servers as fallback when direct hole punching fails.

## User Stories
- As a **user behind strict NAT**, I want to connect to other peers even when direct P2P fails
- As a **network admin**, I want to understand when connections use relay vs direct
- As a **self-hoster**, I want to run my own TURN server for privacy
- As a **user**, I want connections to just work without manual network configuration

---

## Specific Requirements

### TURN Server (coturn)

**Docker Container**
- coturn/coturn:latest image
- UDP port 3478 (standard TURN)
- TCP port 3478 (fallback)
- Secret-based authentication
- Environment configuration

**Configuration**
```
TURN_SECRET - Shared secret for HMAC auth
TURN_REALM - Domain realm (goconnect.io)
TURN_SERVER - External URL (turn:host:3478)
```

**coturn Settings**
```ini
use-auth-secret
static-auth-secret=${TURN_SECRET}
realm=goconnect.io
fingerprint
lt-cred-mech
```

---

### Core Server API

**Endpoint: GET /api/v1/networks/:id/ice-config**

Returns ICE configuration for connecting to network.

**Response:**
```json
{
  "stun_servers": [
    "stun:stun.l.google.com:19302",
    "stun:stun.cloudflare.com:3478"
  ],
  "turn_server": {
    "url": "turn:turn.goconnect.io:3478?transport=udp",
    "username": "net-abc123_1704412800",
    "credential": "hmac_sha256_signature",
    "ttl": 600
  }
}
```

**Credential Generation:**
- Username: `{networkID}_{unix_timestamp}`
- Credential: HMAC-SHA256(username, TURN_SECRET)
- TTL: 600 seconds (10 minutes)
- Refresh: Client fetches new credentials before expiry

---

### CLI Daemon Integration

**Agent Configuration Update**
- Fetch ICE config from core server on network join
- Configure pion/ice with STUN + TURN URLs
- Refresh TURN credentials before TTL expiry

**Code Changes (cli/internal/p2p/agent.go):**
```go
func NewAgentWithTURN(stunURLs []string, turnConfig *TURNConfig) (*Agent, error) {
    var urls []*ice.URL
    
    // Add STUN servers
    for _, stun := range stunURLs {
        uri, _ := ice.ParseURL(stun)
        urls = append(urls, uri)
    }
    
    // Add TURN server if provided
    if turnConfig != nil {
        turn, _ := ice.ParseURL(turnConfig.URL)
        turn.Username = turnConfig.Username
        turn.Password = turnConfig.Credential
        urls = append(urls, turn)
    }
    
    agent, _ := ice.NewAgent(&ice.AgentConfig{
        NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4, ice.NetworkTypeUDP6},
        Urls:         urls,
    })
    // ...
}
```

**Connection Type Detection:**
- After connection established, check selected candidate pair
- Candidate types: `host`, `srflx` (STUN), `relay` (TURN)
- Expose via gRPC: `PeerInfo.connection_type`

---

### Desktop UI Updates

**PeerInfo Type Extension (tauri-api.ts):**
```typescript
interface PeerInfo {
    // existing fields...
    connection_type: 'direct' | 'relay' | 'connecting' | 'failed';
}
```

**Peer Card Badge (App.tsx):**
```tsx
// Connection type badge
const connectionBadge = {
    'direct': { icon: '🔗', text: 'Direct', className: 'bg-green-600/20 text-green-400' },
    'relay':  { icon: '🔄', text: 'Relay', className: 'bg-yellow-600/20 text-yellow-400' },
    'connecting': { icon: '⏳', text: 'Connecting', className: 'bg-gray-600/20 text-gray-400 animate-pulse' },
    'failed': { icon: '❌', text: 'Failed', className: 'bg-red-600/20 text-red-400' },
};
```

**Toast Notifications:**
- On relay fallback: "Connected to {peer} via relay (slower connection)"
- On direct upgrade: "Connection to {peer} upgraded to direct P2P"

---

## Fallback Flow

```
1. User joins network
2. Daemon fetches ICE config from core server
3. pion/ice gathers candidates:
   - host (local interface)
   - srflx (STUN - public IP)
   - relay (TURN - fallback)
4. Connectivity checks run in parallel
5. ICE selects best working candidate pair:
   - Prefers host > srflx > relay
6. Connection established
7. UI shows badge: 🔗 Direct or 🔄 Relay
```

---

## Out of Scope (Phase 3+)

- Custom TURN server configuration in settings
- Multiple TURN server load balancing
- NAT type detection UI
- Connection quality metrics dashboard
- Force relay mode toggle
- TURN server DNS SRV lookup
