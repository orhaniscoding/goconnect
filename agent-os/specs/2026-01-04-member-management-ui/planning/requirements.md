# Requirements: Member Management UI

## User Answers Summary

### 1. Member List Location
**Answer: (a) New "Members" tab next to "Connected Peers"**
- Dedicated tab for discoverability
- Consistent with existing tab pattern
- Scalable for large member lists
- Search/filter bar at top
- Role badges and action buttons

### 2. Role Assignment
**Answer: (b) Owner can only make Admins, no moderator role for MVP**
- Three roles: Owner, Admin, Member
- Permission matrix:
  - Owner: Full control (promote/demote, kick, ban, delete)
  - Admin: Can kick members only
  - Member: View-only access
- Owner sees promote/demote buttons
- Admin sees kick button only

### 3. Banned Members
**Answer: (c) Accessible from Network Settings**
- Separate "Banned Members" section in settings
- Shows ban reason and date
- Unban button with confirmation
- Keep banned members out of main list

### 4. Join Request Approval
**Answer: (c) Inline in member list with Approve/Reject buttons**
- Pending requests section at top of Members tab
- Badge showing pending count
- Quick approve/reject buttons
- Toast notification for new requests

### 5. Kick Confirmation
**Answer: (a) Yes, show confirmation dialog**
- Dialog shows member name
- Explains consequence (can rejoin with invite)
- Cancel and confirm buttons
- Ban requires stronger confirmation

### 6. Out of Scope
Confirmed:
- Transfer network ownership
- Audit log
- Bulk operations
- Email notifications
- Moderator role
- Member search by IP

---

## Existing Code to Reference

### Backend APIs
- `GET /networks/:id/members` - List members with status filter
- `DELETE /tenants/:tenantId/members/:memberId` - Remove member
- Membership statuses: Pending, Approved, Banned

### Desktop Tauri API (`tauri-api.ts`)
- `kickPeer(network_id, peer_id)` - Already exists
- `banPeer(network_id, peer_id, reason)` - Already exists  
- `unbanPeer(network_id, peer_id)` - Already exists
- **Missing**: List members, role assignment

### UI Patterns
- Tab structure from App.tsx (Connected Peers, Chat, etc.)
- Modal components from NetworkModals.tsx
- Peer list card design from peer rendering

---

## Technical Considerations

### New API Methods Needed
```typescript
// In tauri-api.ts
listMembers: (network_id: string) => Promise<MemberInfo[]>
promoteMember: (network_id: string, member_id: string, role: 'admin' | 'member') => Promise<void>
approveMember: (network_id: string, member_id: string) => Promise<void>
rejectMember: (network_id: string, member_id: string) => Promise<void>
```

### New Types Needed
```typescript
interface MemberInfo {
    id: string;
    user_id: string;
    name: string;
    role: 'owner' | 'admin' | 'member';
    status: 'pending' | 'approved' | 'banned';
    joined_at: string;
    banned_at?: string;
    ban_reason?: string;
    is_online: boolean;
}
```

### Component Structure
```
MembersTab/
├── MembersTab.tsx (main tab component)
├── PendingRequestsSection.tsx
├── MemberList.tsx
├── MemberCard.tsx
├── MemberActionsDropdown.tsx
├── KickConfirmModal.tsx
├── BanConfirmModal.tsx
└── BannedMembersPanel.tsx (for settings)
```
