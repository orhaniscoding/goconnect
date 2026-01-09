# Spec Requirements: Network Creation & Management

## Initial Description
Users can create private networks with custom names, generate invite codes, and manage basic network settings through the desktop app.

## Requirements Discussion

### First Round Questions

**Q1:** Network naming — I assume network names should be 3-50 characters, alphanumeric with spaces/hyphens allowed. Is that correct?
**Answer:** Confirmed. 3-50 characters, alphanumeric with spaces, hyphens, and underscores allowed. Trim whitespace, no leading/trailing special chars, no repeated spaces.

**Q2:** Invite code format — 8-character alphanumeric codes, case-insensitive. Should we support custom codes?
**Answer:** Auto-generated only for MVP. 8-char alphanumeric (0-9, A-Z uppercase), case-insensitive, server-side generated, unique constraint. Custom/vanity codes out of scope.

**Q3:** Network limits — Unlimited creation per user?
**Answer:** Confirmed. Unlimited for MVP. Future: rate limiting for abuse prevention.

**Q4:** IP allocation scheme — 10.x.x.x/24 per network = 254 peers. Sufficient?
**Answer:** Confirmed. 254 peers per network is sufficient for gaming use cases. Random subnet selection (10.x.y.0/24) with collision avoidance. IPAM service already implemented.

**Q5:** Network deletion — Hard-delete or soft-delete?
**Answer:** Hard-delete (permanent) with confirmation dialog. Type network name to confirm. Cascade deletion of memberships, IPAM allocations, invite codes, join requests. Close WireGuard interface.

**Q6:** WireGuard key management — Client-side or server-side generation?
**Answer:** Client-side generation (daemon). Private key never sent to server. Only public key registered via API. Zero-trust model.

**Q7:** Offline creation — Online-only or offline-capable?
**Answer:** Online-only for MVP. Server connectivity required. Offline creation out of scope.

**Q8:** What should be explicitly OUT OF SCOPE?
**Answer:** Member invitations via email, advanced role management, custom network settings (ports, DNS, MTU), network visibility settings, network cloning, templates, advanced permissions, network health monitoring.

### Existing Code to Reference

**Similar Features Identified:**

- Feature: Network CRUD API - Path: `core/internal/handler/network.go`
- Feature: Network Service Layer - Path: `core/internal/service/network.go`
- Feature: Network Domain Model - Path: `core/internal/domain/network.go`
- Feature: IPAM Service - Path: `core/internal/service/` (IP allocation)
- Feature: Database Migrations - Path: `core/migrations_sqlite/000001_base.up.sql`
- Feature: Device Registration - Path: `core/internal/handler/device.go` (key registration pattern)

**Patterns to reuse:**
- `errorResponse(c, domainErr)` for error handling
- `AuthMiddleware(authService)` for authentication
- `Idempotency-Key` header requirement for mutations
- Standard domain.Error format

### Follow-up Questions

No follow-up questions needed. User provided comprehensive answers.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
UI mockups to be created by UI team after spec completion.

## Requirements Summary

### Functional Requirements
- Users can create private networks with custom names (3-50 chars, alphanumeric + spaces/hyphens/underscores)
- Network creation auto-generates 8-char alphanumeric invite code (case-insensitive)
- Network creation auto-assigns /24 subnet from 10.x.x.x range (254 peers max)
- Network creation auto-generates WireGuard keys client-side (only public key sent to server)
- Users see list of their networks (owned and member of)
- Network owners can rename networks
- Network owners can delete networks (hard-delete with confirmation, type name to confirm)
- Network deletion cascades to memberships, IPAM allocations, invite codes

### Reusability Opportunities
- Existing `NetworkService` in `core/internal/service/network.go` — fully implemented
- Existing `NetworkHandler` in `core/internal/handler/network.go` — API endpoints ready
- Existing `Network` domain model in `core/internal/domain/network.go`
- Desktop UI patterns to be established (React Hook Form + Zod recommended)

### Scope Boundaries

**In Scope:**
- Network creation with name validation
- Auto-generated invite codes (8-char)
- IP allocation (/24 per network)
- Client-side WireGuard key generation
- Network listing (owned + member)
- Network renaming
- Network deletion with confirmation
- Owner role auto-assignment
- Desktop app UI (forms, modals, lists)
- Backend API integration

**Out of Scope:**
- Member invitations via email (Network Join Flow feature)
- Advanced role management beyond owner (Member Management UI feature)
- Custom network settings (ports, DNS, MTU)
- Network visibility settings (public/private toggle)
- Network cloning/duplication
- Network templates
- Advanced permissions
- Network health monitoring (Phase 3)
- Offline network creation
- Custom/vanity invite codes

### Technical Considerations
- Integration: Desktop → HTTP API → Core Server → PostgreSQL/SQLite
- WireGuard keys: Client-side generation, only public key transmitted
- IP allocation: IPAM service handles subnet assignment
- Idempotency: All mutation endpoints require `Idempotency-Key` header
- Authentication: JWT-based, all endpoints require auth
- Validation: Server-side validation on all inputs
