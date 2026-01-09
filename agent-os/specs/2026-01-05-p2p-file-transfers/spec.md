# Specification: P2P File Transfers

## Goal
Enable direct file sharing between connected peers with progress indicators, 1GB limit, and file integrity verification.

## User Stories
- As a **user**, I want to send a file to a connected peer by clicking "Send File"
- As a **user**, I want to see transfer progress with percentage and speed
- As a **user**, I want to accept or reject incoming file transfers
- As a **user**, I want to cancel an ongoing transfer

---

## Specific Requirements

### 1. Send File Button on Peer Cards

**Location:** Add "📁 Send File" button next to existing "Message" button on peer cards.

**Trigger:** Opens file picker dialog via Tauri.

```tsx
// In App.tsx peer card render
{peer.connected && !peer.is_self && (
  <>
    <button onClick={() => handleSendFile(peer)}>📁 Send</button>
    <button onClick={() => openChat(peer)}>💬</button>
  </>
)}
```

---

### 2. File Picker & Validation

**Dialog:**
```typescript
import { open } from '@tauri-apps/plugin-dialog';

const file = await open({
  title: 'Select file to send',
  multiple: false,
});
```

**Validation:**
```typescript
const MAX_SIZE = 1024 * 1024 * 1024; // 1GB

if (file.size > MAX_SIZE) {
  toast.error(`File too large (max 1GB). Selected: ${formatBytes(file.size)}`);
  return;
}
```

---

### 3. Transfer Protocol

**Offer Message:**
```json
{
  "type": "file_transfer_offer",
  "transfer_id": "uuid",
  "filename": "game_save.zip",
  "size": 256000000,
  "sha256": "abc123..."
}
```

**Accept/Reject:**
```json
{ "type": "file_transfer_accept", "transfer_id": "uuid" }
{ "type": "file_transfer_reject", "transfer_id": "uuid" }
```

**Data Streaming:**
- Use existing ICE P2P connection
- Stream in 64KB chunks
- Report progress every 1MB or 1 second

**Complete:**
```json
{
  "type": "file_transfer_complete",
  "transfer_id": "uuid",
  "final_sha256": "abc123..."
}
```

---

### 4. CLI Transfer Manager

**Location:** `cli/internal/transfer/`

**manager.go:**
```go
type TransferManager struct {
    active   map[string]*Transfer
    mu       sync.RWMutex
}

type Transfer struct {
    ID          string
    PeerID      string
    Filename    string
    Size        int64
    Transferred int64
    Direction   string // "upload" | "download"
    Status      string // "pending" | "active" | "completed" | "failed"
    Error       string
    StartedAt   time.Time
}
```

**Methods:**
- `StartUpload(peerID, filePath)` → Transfer
- `AcceptDownload(transferID, savePath)` → error
- `RejectDownload(transferID)` → error
- `CancelTransfer(transferID)` → error
- `ListTransfers()` → []Transfer
- `GetStats()` → TransferStats

---

### 5. Tauri Command Handlers

**New commands in `src-tauri/src/`:**
```rust
#[tauri::command]
async fn daemon_send_file(peer_id: String, file_path: String) -> Result<String, String>;

#[tauri::command]
async fn daemon_accept_transfer(transfer_id: String, save_path: String) -> Result<(), String>;

#[tauri::command]
async fn daemon_reject_transfer(transfer_id: String) -> Result<(), String>;

#[tauri::command]
async fn daemon_cancel_transfer(transfer_id: String) -> Result<(), String>;
```

---

### 6. Tauri API Updates

**Add to tauri-api.ts:**
```typescript
sendFile: (peer_id: string, file_path: string) => 
  invoke<string>('daemon_send_file', { peer_id, file_path }),

// Already exists:
// acceptTransfer, rejectTransfer, cancelTransfer
```

---

## UI Components

### Peer Card Update
```tsx
<button 
  onClick={() => handleSendFile(peer.id)}
  className="px-2 py-1 bg-gc-dark-700 text-gray-300 rounded text-xs"
>
  📁 Send
</button>
```

### Transfer Progress (already exists in FileTransferPanel)
- Progress bar with percentage
- Upload/Download badge
- Speed indicator (future)
- Accept/Reject for pending incoming
- Cancel for active

---

## Out of Scope
- Multiple file selection
- Folder transfer
- Resume on disconnect
- Transfer history persistence
- File preview/thumbnails
