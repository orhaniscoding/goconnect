# Specification: Network Join Flow

## Goal
Enable users to join networks via invite codes or `gc://` deep links, with automatic WireGuard tunnel configuration and support for approval-based networks.

## User Stories
- As a **gamer**, I want to join my friend's network using an invite code so I can play games together
- As a **user**, I want to click a `gc://` link and automatically join a network without manual steps
- As a **user**, I want to see the status of my pending join requests for approval-based networks

## Specific Requirements

**Invite Code Input**
- Text input field with 8-character limit
- Auto-uppercase transformation (case-insensitive matching)
- "Paste from Clipboard" button with 📋 icon
- Validation: alphanumeric only (0-9, A-Z)
- Real-time validation feedback

**Deep Link Protocol**
- Register `gc://` as custom protocol handler via Tauri
- Format: `gc://join?code=ABC12XYZ`
- Cross-platform registration:
  - Windows: Registry keys (handled by Tauri)
  - macOS: Info.plist bundle (handled by Tauri)
  - Linux: .desktop file MIME types (handled by Tauri)
- On app startup, check for pending deep link
- Auto-open JoinNetworkModal with pre-filled code

**Join Flow (Open Networks)**
1. User enters/pastes invite code
2. Click "Join" → Loading state
3. Success → Show "Connect now?" dialog
4. User chooses → Connect immediately or later
5. Network appears in sidebar

**Join Flow (Approval-Based Networks)**
1. User enters invite code
2. Click "Join" → Loading state
3. Success → Show "Request Sent" toast
4. Network appears in sidebar with "Awaiting approval" badge
5. Owner approves → Notification to user
6. Network becomes active, user can connect

**Auto-Connect Prompt**
- After successful join, show dialog:
  - "Successfully joined [Network Name]"
  - Button: "Connect to Network Now"
  - Button: "I'll connect later"
- "Connect now" → Trigger WireGuard tunnel setup
- "Later" → Add to network list only

**Error Handling**
| Error | User Message |
|-------|-------------|
| Invalid/expired code | "This invite code is invalid or has expired. Please check and try again." |
| Already member | "You're already a member of this network." (redirect to network) |
| Network at capacity | "This network has reached its maximum number of members." |
| Network deleted | "This network has been deleted by its owner." |
| Rate limiting | "Too many join attempts. Please wait a moment." |
| Banned | "You've been banned from this network." |
| Server offline | "Unable to connect to server. Check your connection." |

**Sidebar Updates**
- Pending networks show with "⏳ Awaiting approval" badge
- Active networks show normally
- Connection status indicator (green/gray dot)

## Visual Design
No mockups provided. Use existing modal patterns from `NetworkModals.tsx`.

## Existing Code to Leverage

**`core/internal/handler/network.go`**
- `JoinNetworkByInvite` endpoint already implemented
- Handles invite code validation and network resolution
- Returns membership or join request based on network policy

**`core/internal/service/membership.go`**
- `JoinByInviteCode` method fully implemented
- Handles idempotency, validation, approval flow

**`desktop/src/components/NetworkModals.tsx`**
- `JoinNetworkModal` exists (needs enhancement)
- Modal styling patterns established

**Tauri Deep Link Support**
- Configure in `tauri.conf.json` for protocol registration
- Handle via Tauri event listener on app startup

## Out of Scope
- Email invitations (future feature)
- QR code generation/scanning (Phase 3+)
- Bulk invite multiple users (Member Management UI)
- Join via username/search (discovery feature)
- Network details preview before joining
- "Don't ask again" for connect prompt (future enhancement)
