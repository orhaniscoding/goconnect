# Requirements: Network Join Flow

## User Answers Summary

### 1. Join Input Methods
**Answer: (b) Text input + paste from clipboard button**
- Text input for manual code entry
- Paste button with icon for codes shared via chat/email
- Auto-uppercase on input (case-insensitive 8-char codes)
- QR code scanner deferred to Phase 3+

### 2. Deep Link Registration
**Answer: (c) Both - register protocol + manual paste support**
- Register `gc://` as custom protocol (Tauri handles via config)
- Deep link format: `gc://join?code=ABC12XYZ`
- App opens automatically when clicking links
- Manual paste as fallback
- Cross-platform: Windows registry, macOS Info.plist, Linux .desktop

### 3. Approval-Based Networks
**Answer: (c) Both - "Request Sent" message + pending state in sidebar**
- Approval required: Show "Request Sent" toast, network in sidebar with "Awaiting approval" badge
- Open network: Join immediately, show success message
- Owner approves → User notified, network becomes active

### 4. WireGuard Tunnel Auto-Start
**Answer: (c) Ask user preference: "Connect now?"**
- Show dialog after successful join with two options
- "Connect to Network Now" → Auto-start WireGuard tunnel
- "I'll connect later" → Network in list, user connects manually
- Future: "Don't ask again" checkbox

### 5. Error Scenarios
All error states require specific UI handling:

| Error | Message |
|-------|---------|
| Invalid/expired code | "This invite code is invalid or has expired" |
| Already a member | "You're already a member of this network" |
| Network at capacity | "This network has reached maximum members" |
| Network deleted | "This network has been deleted by its owner" |
| Rate limiting | "Too many attempts. Please wait..." |
| Network not found | "Network not found" |
| Banned | "You've been banned from this network" |
| Server offline | "Unable to connect to server" |

### 6. Join Confirmation
**Answer: (b) No preview, join directly (simpler flow)**
- MVP simplicity - skip network details preview
- Show network info after joining in list/details
- Preview is future enhancement (Phase 3+)

### 7. Out of Scope
Confirmed OUT OF SCOPE:
- Email invitations
- QR code generation/scanning
- Bulk invite multiple users
- Join via username/search
- Network details preview before join

---

## Existing Code to Reference

### Backend (Already Implemented)
- `core/internal/handler/network.go` - `JoinNetworkByInvite` endpoint
- `core/internal/service/membership.go` - `JoinByInviteCode` method
- `core/internal/service/invite.go` - Invite validation

### Desktop (Partial)
- `desktop/src/components/NetworkModals.tsx` - `JoinNetworkModal` exists
- `desktop/src/lib/tauri-api.ts` - API integration layer
- Deep link: `gc://join?code=XYZ` format defined

---

## Technical Considerations

### Deep Link Registration (Tauri)
```json
// tauri.conf.json
{
  "tauri": {
    "bundle": {
      "deepLink": {
        "scheme": ["gc"]
      }
    }
  }
}
```

### UI Components Needed
1. Enhanced `JoinNetworkModal` with paste button
2. `JoinSuccessModal` with "Connect now?" options
3. Sidebar pending badge component
4. Error state displays in modal

### State Management
- Track pending join requests
- Update network list on successful join
- Handle deep link on app startup
