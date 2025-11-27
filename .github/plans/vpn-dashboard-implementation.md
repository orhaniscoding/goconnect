# Plan: VPN Management Dashboard Implementation

**Created**: 2025-11-27  
**Status**: Ready for Implementation  
**Assigned to**: @goconnect-lead

---

## Goal

GoConnect'in core VPN özelliklerini kullanılabilir hale getirmek. Backend API'leri zaten hazır, sadece Frontend UI eksik. Kullanıcılar network'lerini, device'larını ve peer'larını yönetebilecek.

## Architecture Impact

Frontend-only (web-ui/src/), Backend'de eksik endpoint varsa minimal ekleme.

---

## Implementation Steps

### Backend Validation (15 dakika)

1. **Check Existing Endpoints**
   - Verify `GET /v1/networks` returns user's networks
   - Verify `POST /v1/networks` creates network
   - Verify `GET /v1/devices` returns user's devices
   - Verify `POST /v1/devices` registers device
   - Verify `GET /v1/networks/:id/config` returns WireGuard config
   - Check if `GET /v1/peers/:id/stats` exists for statistics
   - Check if `GET /v1/admin/*` endpoints exist for admin panel

2. **Add Missing Endpoints** (if needed)
   - Add peer statistics endpoint if not exists
   - Add admin bulk operations if not exists

### Frontend Pages (4-5 hours)

3. **Networks List Page** (`web-ui/src/app/networks/page.tsx`)
   - Fetch and display user's networks with `getNetworks()`
   - Show network cards with: name, CIDR, visibility badge, member count
   - Add "Create Network" floating action button
   - Link each card to network detail page

4. **Network Detail Page** (`web-ui/src/app/networks/[id]/page.tsx`)
   - Fetch network details with `getNetwork(id)`
   - Display network info: CIDR, DNS, MTU, split_tunnel
   - Show members list with roles (owner/admin/moderator/member)
   - Show devices list for this network
   - Add join request approval section for admins

5. **Devices Page** (`web-ui/src/app/devices/page.tsx`)
   - Fetch user's devices with `getDevices()`
   - Display device cards: name, platform icon, public key (truncated), status
   - Add "Register New Device" button
   - Show peers for each device

6. **Device Detail Modal** (`web-ui/src/components/DeviceDetailModal.tsx`)
   - Show full device info and peer list
   - Add "Download Config" button calling `downloadConfig(networkId)`
   - Add "Delete Device" button
   - Show QR code for mobile config (optional)

7. **Admin Dashboard** (`web-ui/src/app/admin/page.tsx`)
   - Show stats: total users, networks, devices, active peers
   - List recent activities
   - Links to user/network management pages
   - Restrict access to admin users only

### Frontend Components (2-3 hours)

8. **CreateNetworkDialog Component** (`web-ui/src/components/CreateNetworkDialog.tsx`)
   - Form fields: name, CIDR (default: 10.0.0.0/24), visibility, join_policy
   - Validation: CIDR format, name length
   - Call `createNetwork()` API

9. **RegisterDeviceDialog Component** (`web-ui/src/components/RegisterDeviceDialog.tsx`)
   - Form fields: device_name, platform dropdown, public_key textarea
   - Generate keypair button (optional, using webcrypto)
   - Call `registerDevice()` API

10. **NetworkCard Component** (`web-ui/src/components/NetworkCard.tsx`)
    - Display: network name, CIDR, member count, visibility badge
    - Click to navigate to detail page
    - Role badge if user is owner/admin

11. **DeviceCard Component** (`web-ui/src/components/DeviceCard.tsx`)
    - Display: device name, platform icon, online status
    - Last seen timestamp
    - Click to open detail modal

12. **PeerStatsWidget Component** (`web-ui/src/components/PeerStatsWidget.tsx`)
    - Display: RX/TX bytes, last handshake, connection status
    - Auto-refresh every 30 seconds

### Navigation & Routing (30 minutes)

13. **Update Navbar** (`web-ui/src/components/Navbar.tsx`)
    - Add "Networks" link
    - Add "Devices" link
    - Add "Admin" link (only for admin users)
    - Remove or de-prioritize "Feed" and "Profile"

14. **Add Protected Routes**
    - Wrap admin pages with admin-only middleware
    - Redirect non-authenticated users to login

### API Client Updates (1 hour)

15. **Extend API Client** (`web-ui/src/lib/api.ts`)
    - Add TypeScript interfaces for Network, Device, Peer, Membership
    - Verify existing functions: `listNetworks()`, `createNetwork()`, etc.
    - Add missing functions if any: `getPeerStats()`, `getAdminDashboard()`

---

## Technical Considerations

### Backend API Coverage
- Most endpoints already exist in `server/internal/handler/`
- Check: peer statistics, admin dashboard stats
- If missing: add simple aggregation queries

### Frontend State Management
- Use React useState/useEffect for now
- Consider Zustand or Context if state grows complex
- Polling for real-time updates (WebSocket later)

### Authentication & Authorization
- Use existing `getAuthHeader()` for JWT
- Hide admin UI based on user role from token
- Backend already handles RBAC validation

### UX/UI Decisions
- Use Tailwind utility classes (consistent with project)
- Platform icons: emoji or SVG icons (🪟💻📱🤖🍎)
- Online status: green dot if last_seen < 5 min
- Config download: trigger browser download with Blob

### Performance
- Paginate networks/devices if count > 50
- Cache network/device lists for 30 seconds
- Lazy load peer statistics

---

## Dependencies & Risks

### New Dependencies
- None required (use existing stack)
- Optional: `qrcode.react` for QR code generation

### Known Risks
- Peer statistics endpoint might not exist (check backend)
- Admin endpoints might be incomplete (check backend)
- WireGuard config generation must be server-side (security)

### Compatibility
- Next.js 14 App Router (already in use)
- Existing auth flow (JWT + refresh)

---

## Implementation Priority

1. ✅ Backend API validation (verify endpoints)
2. 🔥 Networks page (highest priority - core feature)
3. 🔥 Devices page (second priority)
4. 📊 Peer statistics widget
5. 👑 Admin dashboard (last)

---

## Success Criteria

- ✅ User can view all networks
- ✅ User can create new network
- ✅ User can view network details and members
- ✅ User can register new device
- ✅ User can download WireGuard config
- ✅ User can see peer connection status
- ✅ Admin can view dashboard stats

---

## Next Steps

1. Developer validates backend API coverage
2. Start with Networks List page
3. Add Network Detail page
4. Add Devices page
5. Add components and navigation
6. Test end-to-end flow
7. Document any backend changes needed
