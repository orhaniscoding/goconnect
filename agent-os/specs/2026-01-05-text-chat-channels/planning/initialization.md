# Feature: Text Chat Channels

## Raw Description
Each network has a built-in chat channel with message history, timestamps, and basic formatting.

## Source
Product Roadmap - Phase 2: Communication, Item #6

---

## Existing Implementation Analysis

### Backend (Core) ✅ COMPLETE

**Domain Model** (`core/internal/domain/chat.go`)
- `ChatMessage` struct with full features:
  - ID, Scope, TenantID, UserID, Body
  - Attachments, Redacted, DeletedAt
  - ParentID (threading support)
  - Edit/Redact/SoftDelete methods

**Service** (`core/internal/service/chat.go`)
- Full CRUD operations
- Edit history tracking
- Redaction support

**Handler** (`core/internal/handler/chat.go`)
- GET /v1/chat - List messages
- POST /v1/chat - Send message
- PATCH /v1/chat/:id - Edit message
- DELETE /v1/chat/:id - Delete message
- POST /v1/chat/:id/redact - Redact
- GET /v1/chat/:id/edits - Edit history

**Database**
- `000002_chat_tables.up.sql` - PostgreSQL + SQLite support

### Frontend (Desktop) ⚠️ PARTIAL

**ChatPanel.tsx** - Exists but limited:
- ✅ Basic message list display
- ✅ Send message functionality
- ✅ Polling for updates (3s interval)
- ✅ Self vs other message styling
- ❌ Private DMs not working (backend missing recipient_id)
- ❌ No message formatting
- ❌ No typing indicators
- ❌ No real-time updates (WebSocket)
- ❌ No message reactions

**Tauri API** (`tauri-api.ts`):
- ✅ `getMessages(network_id, limit, before)`
- ✅ `sendMessage(network_id, content)`
- ❌ No edit/delete methods exposed
- ❌ No WebSocket subscription

---

## Gap Analysis Summary

| Feature | Backend | Frontend | Status |
|---------|---------|----------|--------|
| Send/receive messages | ✅ | ✅ | Working |
| Message history | ✅ | ✅ | Working |
| Edit messages | ✅ | ❌ | Needs UI |
| Delete messages | ✅ | ❌ | Needs UI |
| Message formatting | ❌ | ❌ | Not implemented |
| Private DMs | ⚠️ | ⚠️ | Mock only |
| Typing indicators | ❌ | ❌ | Not implemented |
| Real-time updates | ❌ | ❌ | Polling only |
| Message reactions | ❌ | ❌ | Not implemented |

---

## Scope for This Spec

### Must Have (MVP for Phase 2)
1. **Basic formatting** - Markdown bold, italic, code
2. **Edit/Delete UI** - Integrate existing backend
3. **Sender names** - Show peer names instead of IDs
4. **Improved polling** - Reduce interval when active
5. **Unread indicators** - Mark new messages

### Nice to Have (Future)
- WebSocket real-time updates
- Typing indicators
- Private DMs
- Message reactions
- Read receipts
- Message threads

---

## Questions for User

1. **Formatting** - Markdown or rich text editor?
2. **Edit window** - Time limit for editing messages? (e.g., 15 min)
3. **Notification** - Show desktop notification for new messages?
4. **History** - How far back to load messages? (default 50)
5. **Sender display** - Use display_name or peer ID prefix?

---

## Dependencies
- NAT Traversal (complete) - P2P messaging possible
- Member Management (complete) - User names available
