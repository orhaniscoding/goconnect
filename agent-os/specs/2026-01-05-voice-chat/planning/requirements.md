# Requirements: Voice Chat (WebRTC)

## Existing Implementation

### Backend ✅
- `core/internal/handler/voice.go` - Redis-based signaling (offer/answer/ICE)
- VoiceSignal struct, POST/GET endpoints
- 30-second ephemeral signal storage

### CLI ✅
- `cli/cmd/voice.go` - Test command to verify signaling

### Desktop UI 🔶 Partial
- `VoiceChat.tsx` - Placeholder with join/leave buttons
- Missing: WebRTC peer connections, audio capture

---

## Requirements

### 1. Audio Capture
- Browser `getUserMedia()` for microphone access
- Permission request on first join
- Audio-only (no video)

### 2. WebRTC Peer Connections
- Create RTCPeerConnection for each connected peer
- Exchange offers/answers via backend signaling
- ICE candidate trickling

### 3. Voice UI
- Mute/unmute self
- Volume indicators for speaking peers
- Push-to-talk (optional, toggle mode)
- Leave button disconnects all

### 4. Tauri API
- `getVoiceSignals(network_id)` - Poll for incoming signals
- `sendVoiceSignal(signal)` - Send offer/answer/candidate

---

## Technical Approach

**Signaling Flow:**
1. User A joins → Creates offers for all connected peers
2. Offers sent via `sendVoiceSignal` → Redis queue
3. User B polls → Gets offer → Creates answer
4. Answer sent back → Peer connection established
5. Audio streams exchanged directly via WebRTC

**Audio Routing:**
- Each peer has its own `RTCPeerConnection`
- Audio tracks added to remote audio element
- Mixed automatically by browser

---

## Effort Estimate
| Task | Days |
|------|------|
| WebRTC connection manager | 2 |
| Audio capture/playback | 0.5 |
| Tauri API for signals | 0.5 |
| UI enhancements | 1 |
| Testing | 1 |
| **Total** | **5 days** |
