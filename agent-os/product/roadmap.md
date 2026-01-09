# Product Roadmap

> **Status Legend**: ✅ Implemented & Tested | 🔧 Code Exists (Needs E2E Testing) | 🚧 Skeleton/In Progress

## Phase 1: Core Networking (MVP)

1. [x] **Network Creation & Management** — Users can create private networks with custom names, generate invite codes, and manage basic network settings through the desktop app. `M` ✅

2. [x] **Network Join Flow** — Users can join networks via invite codes or `gc://` protocol deep links, with automatic WireGuard tunnel configuration. `M` 🔧

3. [x] **Peer Discovery & Connection** — Automatic discovery of peers on the same network with visual status indicators (online/offline/connecting). `S` 🔧

4. [x] **NAT Traversal** — STUN/TURN integration and UDP hole punching to establish P2P connections through firewalls and NATs. `L` 🔧

5. [x] **Member Management UI** — Network owners can view members, remove users, and assign basic roles (admin/member) from the desktop app. `S` 🔧

---

## Phase 2: Communication

6. [x] **Text Chat Channels** — Each network has a built-in chat channel with message history, timestamps, and basic formatting. `M` 🔧
   - ChatPanel.tsx: 311 lines, edit/delete, markdown

7. [x] **Voice Chat (WebRTC)** — Push-to-talk and open mic voice communication between peers with volume controls and mute functionality. `L` 🔧
   - VoiceChat.tsx: 303 lines, RTCPeerConnection
   - Backend: VoiceHandler with Redis signaling

8. [x] **P2P File Transfers** — Direct file sharing between peers with progress indicators, pause/resume, and transfer history. `M` 🔧
   - CLI transfer/manager.go: 674 lines
   - Tests: 49KB
   - Desktop UI: FileTransferPanel.tsx

---

## Phase 3: Polish & Stability

9. [x] **Network Health Dashboard** — Visual overview showing peer latency, connection quality, bandwidth usage, and network topology. `S` 🔧
   - MetricsDashboard.tsx: 188 lines

10. [x] **Notification System** — Desktop notifications for new messages, voice calls, file transfers, and member join/leave events. `S` 🔧
    - notifications.ts created

11. [x] **Settings & Preferences** — User preferences for audio devices, hotkeys, startup behavior, and notification settings. `S` 🔧
    - SettingsPanel.tsx: 218 lines

12. [x] **Auto-Update System** — Automatic update detection and installation for the desktop app with rollback capability. `M` 🔧
    - UpdateChecker.tsx: 185 lines, Tauri updater

---

> **Notes**
> - 🔧 items have code but need end-to-end testing
> - Effort scale: `XS` 1 day, `S` 2-3 days, `M` 1 week, `L` 2 weeks, `XL` 3+ weeks
