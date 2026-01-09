# Feature: P2P File Transfers

## Raw Description
Direct file sharing between peers with progress indicators, pause/resume, and transfer history.

## Source
Product Roadmap - Phase 2: Communication, Item #8

---

## Existing Implementation Analysis

### Desktop UI ✅ COMPLETE

**FileTransferPanel.tsx** - Fully implemented:
- Stats header (active, up, down, completed)
- Transfer list with progress bars
- Accept/Reject for incoming
- Cancel for active
- Upload/download badges
- Polling every 2s

**Tauri API** (`tauri-api.ts`):
- `TransferInfo` interface (id, peer_id, file_name, file_size, transferred, status, direction)
- `TransferStats` interface (totals, active, completed, failed)
- `listTransfers()` ✅
- `getTransferStats()` ✅
- `acceptTransfer(id, save_path)` ✅
- `rejectTransfer(id)` ❓ Not visible in shown code
- `cancelTransfer(id)` ❓ Not visible in shown code

### Backend (CLI Daemon) ❌ NOT IMPLEMENTED

No transfer-related Go files found in `cli/internal/`. Need to implement:
- Transfer manager (track active transfers)
- P2P data channel over ICE connection
- File chunking and reassembly
- Resume capability (track progress)

---

## Gap Analysis

| Component | Status | Notes |
|-----------|--------|-------|
| UI Panel | ✅ Complete | FileTransferPanel.tsx |
| Tauri API | ✅ Complete | Methods defined |
| Daemon commands | ❌ Missing | Need Rust-side handlers |
| CLI transfer logic | ❌ Missing | Need Go implementation |
| P2P data channel | ❌ Missing | Use ICE connection |

---

## Missing Features for MVP

### 1. "Send File" Button
- Add send button somewhere (peer card? dedicated button?)
- File picker dialog
- Select recipient peer

### 2. CLI Transfer Manager
Location: `cli/internal/transfer/`
- `manager.go` - Track transfers
- `sender.go` - Chunk and send file
- `receiver.go` - Receive and reassemble

### 3. P2P Data Channel
- Reuse existing ICE connection from P2P agent
- Protocol: `[type:1][transfer_id:16][chunk_num:4][data:N]`

### 4. Tauri Command Handlers
Location: `desktop/src-tauri/src/`
- Handle `daemon_*_transfer` commands
- Bridge to CLI daemon via IPC

---

## Questions for User

1. **Initiation UI** - Where should "Send File" button be?
   - (a) On peer cards in main view
   - (b) In a dedicated "File Transfer" tab
   - (c) Context menu on right-click

2. **File Size Limit** - Maximum file size?
   - (a) No limit (streaming)
   - (b) 1 GB limit
   - (c) 100 MB limit for MVP

3. **Resume Support** - Require resume on disconnect?
   - (a) Yes, save partial progress
   - (b) No, restart transfer on reconnect

4. **Multiple Files** - Send folder/multiple files?
   - (a) Single file only for MVP
   - (b) Multiple file selection
   - (c) Folder support

---

## Effort Estimate

| Task | Days |
|------|------|
| CLI transfer manager | 2 |
| P2P data channel | 1.5 |
| Tauri handlers | 1 |
| Send file UI | 0.5 |
| Testing & polish | 1 |
| **Total** | **~6 days** |

---

## Dependencies
- NAT Traversal (complete) - ICE connection available
- P2P Agent (complete) - Can send/receive data
