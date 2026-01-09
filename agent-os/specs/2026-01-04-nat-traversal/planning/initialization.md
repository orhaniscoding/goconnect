# Feature: NAT Traversal

## Raw Description
STUN/TURN integration and UDP hole punching to establish P2P connections through firewalls and NATs.

## Source
Product Roadmap - Phase 1: Core Networking (MVP), Item #4

---

## Existing Implementation Analysis

### Current P2P Architecture (cli/internal/p2p/)

**STUN is ALREADY IMPLEMENTED** ✅

- `agent.go` - ICE agent using `pion/ice/v2`
  - Default STUN: `stun:stun.l.google.com:19302`
  - Creates ICE agents with UDP4/UDP6 support
  - Candidate gathering, dial/accept flows

- `manager.go` - P2P connection manager
  - Signaling interface (offer/answer/candidate exchange)
  - Connection state monitoring
  - Auto-reconnect with exponential backoff
  - Latency measurement via ping/pong

- Proto definitions already have `CONNECTION_TYPE_RELAY`

### What's Missing

1. **TURN Server Support**
   - No TURN configuration in agent.go
   - Relay fallback not implemented
   - TURN credentials management needed

2. **Configurable STUN/TURN Servers**
   - Currently hardcoded Google STUN
   - Need user-configurable servers
   - Environment-based configuration

3. **Connection Type Detection & Display**
   - Backend has `CONNECTION_TYPE_RELAY` proto value
   - UI doesn't show direct vs relay status
   - No analytics on connection types

4. **NAT Type Detection**
   - No logic to detect NAT type (Full Cone, Symmetric, etc.)
   - Could help optimize hole punching strategy

---

## Gap Analysis Summary

| Feature | Status | Effort |
|---------|--------|--------|
| STUN basic support | ✅ Done | - |
| TURN integration | ❌ Missing | M |
| Configurable servers | ❌ Missing | S |
| Relay fallback logic | ❌ Missing | M |
| Connection type UI | ❌ Missing | S |
| NAT type detection | ❌ (nice to have) | L |

---

## Scope for This Spec

### Must Have (MVP)
1. **TURN Server Integration** - Add TURN URL/credentials to agent config
2. **Configuration** - Environment vars or config file for STUN/TURN
3. **Relay Fallback** - Detect when direct fails, use TURN
4. **Connection Type Display** - Show "Direct" vs "Relay" in UI

### Nice to Have (Future)
- NAT type detection algorithm
- Multiple STUN/TURN server failover
- Self-hosted TURN server instructions

---

## Technical Research Needed

### TURN Integration with pion/ice
```go
// Current (STUN only)
uri, _ := ice.ParseURL("stun:stun.l.google.com:19302")
agent, _ := ice.NewAgent(&ice.AgentConfig{
    Urls: []*ice.URL{uri},
})

// With TURN
stunURL, _ := ice.ParseURL("stun:stun.l.google.com:19302")
turnURL, _ := ice.ParseURL("turn:turn.example.com:3478?transport=udp")
turnURL.Username = "user"
turnURL.Password = "pass"

agent, _ := ice.NewAgent(&ice.AgentConfig{
    Urls: []*ice.URL{stunURL, turnURL},
})
```

### Relay Detection
- ICE agent provides candidate pair info after connection
- Check local/remote candidate types:
  - `host` - Direct local network
  - `srflx` - Server reflexive (STUN hole punch success)
  - `relay` - TURN relay (fallback)

---

## Questions for User

1. **TURN Server** - Use a public TURN server, self-hosted, or cloud service (Twilio/Xirsys)?
2. **Configuration** - Store STUN/TURN in settings panel or config file only?
3. **Default Behavior** - Always try direct first, then fallback to relay?
4. **Credential Security** - TURN credentials stored locally or fetched from core server?

---

## Dependencies
- Peer Discovery & Connection (complete)
- Network Join Flow (complete)
