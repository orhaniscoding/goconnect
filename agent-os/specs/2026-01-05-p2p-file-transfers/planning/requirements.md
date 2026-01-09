# Requirements: P2P File Transfers

## User Answers Summary

### 1. Send File Button Location
**Answer: (a) On peer cards**
- "Send File" button on each connected peer card
- File picker dialog after click
- Contextual and discoverable

### 2. File Size Limit
**Answer: (b) 1 GB**
- Reasonable for gaming use cases (maps, mods, saves)
- Client-side validation before transfer
- Error dialog if exceeded

### 3. Resume Support
**Answer: (b) No for MVP**
- Restart transfer if disconnected
- Simpler implementation
- Resume deferred to Phase 3

### 4. Multiple Files
**Answer: (a) Single file only**
- MVP simplicity
- One transfer = one progress bar
- Multi-file deferred to Phase 3

---

## Technical Decisions

### Transfer Protocol
```
1. Sender → file_transfer_offer (filename, size, hash)
2. Receiver → file_transfer_accept / file_transfer_reject
3. Sender → stream data chunks over P2P
4. Sender → file_transfer_complete (final hash)
5. Receiver → verify hash, show success/error
```

### File Validation
```typescript
const MAX_FILE_SIZE = 1024 * 1024 * 1024; // 1GB
```

### Security
- SHA-256 hash verification
- Already encrypted via WireGuard tunnel

---

## Effort Estimate

| Task | Days |
|------|------|
| Send File UI + button | 0.5 |
| CLI transfer manager | 2 |
| P2P data streaming | 1 |
| Tauri command handlers | 0.5 |
| Testing & polish | 1 |
| **Total** | **5 days** |
