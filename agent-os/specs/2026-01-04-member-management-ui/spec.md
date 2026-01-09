# Specification: Member Management UI

## Goal
Enable network owners and admins to view all network members, manage roles, approve join requests, and remove/ban users through the desktop app.

## User Stories
- As a **network owner**, I want to see all members in my network and their roles
- As a **network owner**, I want to promote members to admin so they can help manage the network
- As a **network owner**, I want to approve or reject join requests for my network
- As a **network admin**, I want to kick disruptive members from the network
- As a **network owner**, I want to ban members who violate rules and unban them later if needed

## Specific Requirements

**Members Tab**
- New tab labeled "Members" in main content area between existing tabs
- Shows total member count in tab label
- Search/filter bar to find members by name
- Scrollable member list with cards

**Member Card Display**
- User name and avatar placeholder
- Role badge: 👑 Owner (gold), ⚡ Admin (blue), 👤 Member (gray)
- Online status indicator (reuse peer status logic)
- Join date (relative: "Joined 3 days ago")
- Actions dropdown (context-sensitive based on viewer role)

**Role Management (Owner Only)**
- "Promote to Admin" button for regular members
- "Demote to Member" button for admins
- Changes take effect immediately
- Toast notification confirms action

**Pending Requests Section**
- Appears at top of Members tab when requests exist
- Badge on tab: "Members (5) • 2 pending"
- Each request shows:
  - User name
  - Request time ("Requested 2 hours ago")
  - Approve/Reject buttons
- Approve adds to members, reject removes request

**Kick Member Flow**
- Available to Owner and Admin
- Cannot kick self, owner, or higher-role users
- Confirmation dialog:
  - Shows member name
  - "Remove Member from Network?" title
  - Explains they can rejoin with invite
  - Cancel/Confirm buttons
- Kicked member removed from list immediately
- Toast: "Member removed from network"

**Ban Member Flow**
- Available to Owner only
- Ban confirmation dialog:
  - Shows member name
  - Required text field for ban reason
  - Warning: "They cannot rejoin until unbanned"
  - Cancel/Confirm buttons
- Banned member moved to Banned list
- Toast: "Member banned"

**Banned Members (Network Settings)**
- Section in Settings panel: "Banned Members"
- Shows list of banned users with:
  - Name
  - Ban date
  - Ban reason
  - Unban button
- Unban confirmation dialog
- Unbanned user removed from banned list
- Toast: "Member unbanned - they can now rejoin"

## Visual Design
Use existing design patterns:
- Card style from peer list (bg-gc-dark-800, rounded-lg, border)
- Tab style from existing tabs
- Modal style from NetworkModals.tsx
- Role colors: Owner (yellow-500), Admin (blue-500), Member (gray-500)

## Existing Code to Leverage

**`desktop/src/lib/tauri-api.ts`**
- `kickPeer`, `banPeer`, `unbanPeer` already implemented
- Add `listMembers`, `promoteMember`, `approveMember`, `rejectMember`

**`desktop/src/App.tsx`**
- Tab rendering logic for adding Members tab
- Peer list pattern for member list

**`desktop/src/components/NetworkModals.tsx`**
- Modal patterns for confirmation dialogs

**`core/internal/handler/network.go`**
- `ListMembers` endpoint exists
- Backend fully implemented

## Out of Scope
- Transfer network ownership
- Audit log of member actions  
- Bulk member operations
- Email notifications
- Moderator role
- Member search by IP/device
- Member notes/tags
