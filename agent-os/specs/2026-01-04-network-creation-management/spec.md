# Specification: Network Creation & Management

## Goal
Enable users to create private virtual LAN networks with custom names, auto-generated invite codes, and manage basic network settings (rename, delete) through the desktop app.

## User Stories
- As a **gamer**, I want to create a private network so that I can play LAN games with friends online
- As a **network owner**, I want to generate an invite code so that I can easily share access with friends
- As a **network owner**, I want to delete my network so that I can clean up unused gaming sessions

## Specific Requirements

**Network Creation**
- Network name must be 3-50 characters (alphanumeric, spaces, hyphens, underscores)
- Trim leading/trailing whitespace, collapse repeated spaces
- No leading/trailing hyphens or underscores allowed
- Network creation requires authenticated user and server connectivity
- Return validation errors immediately on invalid input
- Auto-assign owner role to creating user

**Invite Code Generation**
- Auto-generate 8-character alphanumeric code (0-9, A-Z uppercase only)
- Case-insensitive matching when users input codes
- Server-side generation with unique constraint in database
- Display code prominently after network creation
- Support `gc://join?code=XYZ` deep link format for sharing

**IP Allocation**
- Allocate /24 subnet from 10.x.x.x range per network (254 usable IPs)
- Random subnet selection with collision avoidance
- Use existing IPAM service for allocation logic
- First IP (.1) reserved for network gateway

**WireGuard Key Generation**
- Generate WireGuard key pair client-side (in daemon)
- Only send public key to server via API
- Never transmit or store private key on server (zero-trust)
- Store private key locally in encrypted storage

**Network Listing**
- Show all networks user owns or is member of
- Visually distinguish owned networks (crown/star icon)
- Display for each: network name, member count, connection status
- Sort by most recently used

**Network Renaming**
- Only network owner can rename
- Same validation rules as creation (3-50 chars, alphanumeric + special)
- Immediate UI update after successful rename

**Network Deletion**
- Only network owner can delete
- Hard-delete (permanent, not soft-delete)
- Two-step confirmation: dialog + type network name to confirm
- Cascade delete: memberships, IP allocations, invite codes, join requests
- Disconnect all connected peers via daemon

**Desktop UI Integration**
- Extend existing `CreateNetworkModal` component for validation
- Add "Manage Network" button in `NetworkDetails` for owner actions
- Show toast notification on success/error
- Disable submit button during API call (loading state)

## Visual Design
No visual mockups provided. Use existing component patterns from `NetworkModals.tsx` and `NetworkDetails.tsx` as reference.

## Existing Code to Leverage

**`core/internal/handler/network.go` (NetworkHandler)**
- Complete REST API: CreateNetwork, ListNetworks, GetNetwork, UpdateNetwork, DeleteNetwork
- JWT authentication middleware already applied
- Idempotency-Key header support for mutations
- Error response pattern: `errorResponse(c, domainErr)`

**`core/internal/service/network.go` (NetworkService)**
- Business logic for CRUD operations fully implemented
- Idempotency handling built-in
- Audit logging integration ready

**`core/internal/domain/network.go` (Network model)**
- Domain model with all fields defined
- `CreateNetworkRequest` struct with validation tags
- `ApplyDefaults()` for sensible default values (visibility=private, join_policy=approval)

**`desktop/src/components/NetworkModals.tsx`**
- `CreateNetworkModal` component exists (needs validation enhancement)
- Modal styling pattern with `bg-gc-dark-800`, border colors established
- Keyboard shortcuts (Enter to submit, Escape to close)

**`core/internal/service/invite.go` (InviteService)**
- Invite code generation and validation already implemented
- `CreateInvite`, `ListInvites`, `ValidateInvite` methods ready
- Token-based invite URLs with expiration support

## Out of Scope
- Member invitations via email (separate Network Join Flow feature)
- Advanced role management beyond owner role (Member Management UI feature)
- Custom network settings (ports, DNS, MTU configuration)
- Network visibility settings (public/private toggle in UI)
- Network cloning or duplication
- Network templates for quick setup
- Advanced permissions beyond owner/member
- Network health monitoring dashboard (Phase 3 feature)
- Offline network creation capability
- Custom/vanity invite codes (premium feature for future)
