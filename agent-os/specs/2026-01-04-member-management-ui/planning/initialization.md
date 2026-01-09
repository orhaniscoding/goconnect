# Feature: Member Management UI

## Raw Description
Network owners can view members, remove users, and assign basic roles (admin/member) from the desktop app.

## Source
Product Roadmap - Phase 1: Core Networking (MVP), Item #5

## Existing Implementation Analysis

### Backend (Core Server) - ✅ COMPLETE
- `core/internal/handler/network.go` - `ListMembers` endpoint: `GET /networks/:id/members`
- `core/internal/handler/tenant.go` - `RemoveMember` endpoint: `DELETE /tenants/:tenantId/members/:memberId`
- `core/internal/service/membership.go` - Full membership service with status filtering
- `core/internal/service/tenant_membership.go` - Role-based member management
- Roles: Owner, Admin, Moderator, Member
- Statuses: Pending, Approved, Banned

### CLI Daemon - Needs Verification
- Likely has gRPC endpoints for member operations
- Desktop uses Tauri commands that proxy to daemon

### Desktop App - ✅ API LAYER READY
- `desktop/src/lib/tauri-api.ts` already has:
  - `kickPeer(network_id, peer_id)` - Remove member
  - `banPeer(network_id, peer_id, reason)` - Ban member
  - `unbanPeer(network_id, peer_id)` - Unban member
- Missing:
  - List members API
  - Role assignment API
  - Member details view

## Gap Analysis

### What Needs to Be Built
1. **Member List UI** - View all network members
2. **Member Actions Dropdown** - Kick, Ban, Unban (owner/admin only)
3. **Role Assignment** - Promote/demote members (owner only)
4. **Banned Members View** - List banned users, allow unban
5. **Pending Join Requests** - For approval-based networks

### What Exists
- Backend APIs fully implemented
- Tauri API has kick/ban/unban methods
- Peer list shows current connected peers (not full member list)

## Scope for This Spec
1. Member list panel (separate from Connected Peers)
2. Owner-only management actions
3. Role badges in member list
4. Ban/unban functionality
5. Pending approval list (if network requires approval)

## Dependencies
- Network Creation & Management (complete)
- Network Join Flow (complete)
