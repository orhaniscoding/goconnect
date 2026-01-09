# Feature: Peer Discovery & Connection

## Raw Description
Automatic discovery of peers on the same network with visual status indicators (online/offline/connecting).

## Source
Product Roadmap - Phase 1: Core Networking (MVP), Item #3

## Existing Implementation Analysis

### Backend (CLI Daemon) - ✅ COMPLETE
- `cli/internal/daemon/grpc_server.go` - `GetPeers` handler implemented
- `cli/internal/p2p/manager.go` - P2P connection management, peer status tracking
- `cli/internal/engine/engine.go` - Engine connects peer list with P2P status
- `cli/internal/proto/daemon.pb.go` - gRPC protobuf definitions for Peer service

### Desktop UI - ✅ MOSTLY COMPLETE
- `desktop/src/App.tsx` - Peer list rendering with:
  - Connection status (green/gray dot)
  - Peer name and virtual IP
  - Latency display (ms)
  - Message button for private chat
- `desktop/src/lib/tauri-api.ts` - `getPeers()` API call
- Auto-refresh on 3-second interval

### What's Already Working
1. ✅ Peers visible in list when connected to network
2. ✅ Status indicators (online/offline based on `connected` field)
3. ✅ Latency display for connected peers
4. ✅ Self-peer identification (`is_self` field)
5. ✅ Polling-based refresh (3 seconds)

## Gap Analysis

### Missing/Could Improve
1. **"Connecting" state** - No visual difference between "offline" and "connecting"
2. **Real-time updates** - Currently polling; WebSocket would be smoother
3. **Peer details** - No expandable details (OS, version, connection type)
4. **Search/filter** - No way to search peers in large networks
5. **Sorting** - No sorting options (by name, status, latency)

## Recommendation

**Status: Feature is 90%+ complete.**

The core peer discovery functionality is already fully implemented. Remaining work is polish:
- Add "connecting" visual state
- Consider WebSocket for real-time updates (future enhancement)
- Add peer count to network list

## Scope for This Spec
Since feature is largely complete, this spec will focus on:
1. ✅ Verify existing implementation works
2. Add "connecting" status indicator
3. Add peer count badge to sidebar
4. Polish empty state and loading states
