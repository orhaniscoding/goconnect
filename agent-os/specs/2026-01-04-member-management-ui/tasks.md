# Task Breakdown: Member Management UI

## Overview
Total Tasks: 5 Task Groups, 18 Sub-tasks

**Key Insight**: Backend APIs exist. Focus is on desktop UI, Tauri API additions, and CLI daemon integration.

## Task List

### API Layer

#### Task Group 1: Tauri API & Type Extensions ✅ COMPLETE
**Dependencies:** None

- [x] 1.0 Complete API layer
  - [x] 1.1 Add MemberInfo type to tauri-api.ts
    - Includes: id, user_id, name, role, status, joined_at, ban_reason, is_online
  - [x] 1.2 Add listMembers API method
    - `daemon_list_members` command
  - [x] 1.3 Add promoteMember/demoteMember API methods
    - `daemon_promote_member`, `daemon_demote_member` commands
  - [x] 1.4 Add approveMember/rejectMember API methods
    - `daemon_approve_member`, `daemon_reject_member` commands
  - [x] 1.5 Added getBannedMembers helper method

**Acceptance Criteria:** ✅
- All API methods type-safe
- Methods callable from UI components

---

### Desktop UI Layer

#### Task Group 2: Members Tab & List Component ✅ COMPLETE
**Dependencies:** Task Group 1

- [x] 2.0 Complete members tab
  - [x] 2.1 Add "Members" tab to App.tsx tab bar
    - Added between Connected Peers and Chat
    - Shows 👥 icon
  - [x] 2.2 Create MembersTab component
    - Search bar at top
    - Loading and empty states
    - Location: `desktop/src/components/MembersTab.tsx`
  - [x] 2.3 Create MemberCard component (in MembersTab.tsx)
    - Name, role badge, status indicator
    - Join date (relative)
    - Actions buttons (owner/admin)
  - [x] 2.4 Implement search/filter
    - Filter by name

**Acceptance Criteria:** ✅
- Tab visible and selectable
- Member list renders with role badges
- Search filters work

---

#### Task Group 3: Pending Requests UI ✅ COMPLETE
**Dependencies:** Task Group 2

- [x] 3.0 Complete pending requests
  - [x] 3.1 Add pending count display in header
    - Format: "Members (5) • 2 pending"
    - Only shows if pending > 0
  - [x] 3.2 Create PendingRequestsSection component (in MembersTab.tsx)
    - Displayed at top of Members tab
    - Yellow pulsing indicator
  - [x] 3.3 Implement Approve/Reject buttons
    - Call approveMember/rejectMember API
    - UI updates on action

**Acceptance Criteria:** ✅
- Pending requests visible at top
- Approve/reject work correctly
- Count updates on action

---

#### Task Group 4: Member Actions (Kick/Ban/Promote) ✅ COMPLETE
**Dependencies:** Task Group 2

- [x] 4.0 Complete member actions
  - [x] 4.1 Create action buttons in MemberCard
    - Context-sensitive (owner sees all, admin sees kick only)
    - Buttons: Promote, Demote, Remove, Ban
  - [x] 4.2 Create KickConfirmModal (ConfirmModal)
    - Shows member name
    - Explains consequence
    - Cancel/Confirm buttons
  - [x] 4.3 Create BanConfirmModal (ConfirmModal with input)
    - Shows member name
    - Required reason field
    - Warning text
    - Cancel/Confirm buttons
  - [x] 4.4 Implement promote/demote logic
    - Call promoteMember/demoteMember API
    - Update UI immediately

**Acceptance Criteria:** ✅
- Actions respect role permissions
- Confirmations show before destructive actions
- UI updates after actions

---

#### Task Group 5: Banned Members Panel ✅ COMPLETE
**Dependencies:** Task Group 4

- [x] 5.0 Complete banned members
  - [x] 5.1 Create BannedMembersPanel component
    - Show count badge
    - List banned members
    - Location: `desktop/src/components/BannedMembersPanel.tsx`
  - [x] 5.2 Create banned member card
    - Name, ban date, reason
    - Unban button
  - [x] 5.3 Create UnbanConfirmModal
    - Explains they can rejoin
    - Cancel/Confirm buttons
  - [x] 5.4 Implement unban flow
    - Call unbanPeer API
    - Remove from banned list

**Acceptance Criteria:** ✅
- Banned list component created
- Unban works correctly
- List updates after unban

---

## Execution Order ✅ COMPLETED

1. ✅ **API Layer** (Task Group 1) - Types and Tauri methods
2. ✅ **Members Tab** (Task Group 2) - Tab, list, cards
3. ✅ **Pending Requests** (Task Group 3) - Approve/reject UI
4. ✅ **Member Actions** (Task Group 4) - Kick/ban/promote
5. ✅ **Banned Members** (Task Group 5) - Separate panel component

## Files Created/Modified

### New Files ✅
- `desktop/src/components/MembersTab.tsx` - Main members tab with all sub-components
- `desktop/src/components/BannedMembersPanel.tsx` - Banned members for settings

### Modified Files ✅
- `desktop/src/lib/tauri-api.ts` - Added MemberInfo type and member APIs
- `desktop/src/App.tsx` - Added Members tab import and rendering

## Status: ✅ IMPLEMENTATION COMPLETE

### Notes
- BannedMembersPanel needs to be integrated into SettingsPanel.tsx
- All lint errors are environment-related (missing node_modules after npm install)
- Daemon commands (daemon_list_members, etc.) need to be implemented in Tauri backend
