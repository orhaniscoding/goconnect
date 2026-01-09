# Specification: Text Chat Channels

## Goal
Enhance the existing chat system with markdown formatting, edit/delete capabilities, proper sender names, and infinite scroll for a polished messaging experience.

## User Stories
- As a **user**, I want to format my messages with bold/italic/code for clarity
- As a **user**, I want to edit typos within 15 minutes of sending
- As a **user**, I want to delete messages I no longer want visible
- As a **user**, I want to see who sent each message by name
- As a **user**, I want to scroll up to see older message history

---

## Specific Requirements

### 1. Markdown Formatting

**Supported Syntax:**
| Markdown | Rendered |
|----------|----------|
| `**bold**` | **bold** |
| `*italic*` | *italic* |
| `` `code` `` | `code` |
| `@username` | @username (highlighted) |

**Implementation:**
- Parse on render, store raw markdown in database
- Lightweight regex parsing, no external library
- Sanitize HTML to prevent XSS

**Code:**
```typescript
export function formatMarkdown(text: string): string {
  // Escape HTML first
  let safe = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
  
  // Apply markdown
  return safe
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    .replace(/@(\w+)/g, '<span class="mention">@$1</span>');
}
```

---

### 2. Edit/Delete UI

**Edit Button:**
- Shows on hover for own messages within 15-min window
- Opens inline edit field or modal
- Saves via `tauriApi.editMessage()`
- Shows "(edited)" indicator after edit

**Delete Button:**
- Shows on hover for own messages (no time limit)
- Confirmation modal: "Delete this message?"
- Soft delete on backend (already implemented)
- Shows "[Message deleted]" placeholder

**API Methods (add to tauri-api.ts):**
```typescript
editMessage: (message_id: string, new_content: string) => 
  invoke<void>('daemon_edit_message', { message_id, new_content }),

deleteMessage: (message_id: string) => 
  invoke<void>('daemon_delete_message', { message_id }),
```

---

### 3. Sender Display

**ChatMessage Interface Update:**
```typescript
export interface ChatMessage {
  id: string;
  peer_id: string;
  peer_name: string; // NEW: Display name
  content: string;
  timestamp: string;
  is_self: boolean;
  is_edited?: boolean; // NEW
  is_deleted?: boolean; // NEW
}
```

**Display Logic:**
```tsx
<div className="message-sender">
  {message.peer_name || message.peer_id.substring(0, 8) + '...'}
</div>
```

---

### 4. Infinite Scroll

**Initial Load:** 50 most recent messages

**Scroll Detection:**
```typescript
const handleScroll = (e: React.UIEvent) => {
  const { scrollTop } = e.currentTarget;
  if (scrollTop === 0 && !loading && hasMore) {
    loadMoreMessages();
  }
};
```

**Load More:**
```typescript
const loadMoreMessages = async () => {
  setLoading(true);
  const older = await tauriApi.getMessages(networkId, 50, oldestMessageId);
  setMessages(prev => [...older, ...prev]);
  setHasMore(older.length === 50);
  setLoading(false);
};
```

**Loading Indicator:**
```tsx
{loading && (
  <div className="text-center py-2 text-gray-500 animate-pulse">
    Loading older messages...
  </div>
)}
```

---

### 5. Notification Integration

**Check Notification Settings:**
```typescript
if (settings.notifications_enabled && newMessages.length > 0) {
  const latest = newMessages[newMessages.length - 1];
  if (!latest.is_self) {
    showNotification({
      title: `${networkName}`,
      body: `${latest.peer_name}: ${latest.content.substring(0, 50)}`,
      onClick: () => focusChatTab()
    });
  }
}
```

---

## UI Components

### Message Bubble
```tsx
<div className={`message ${msg.is_self ? 'self' : 'other'}`}>
  <div className="message-header">
    <span className="sender">{msg.peer_name}</span>
    <span className="time">{formatTime(msg.timestamp)}</span>
    {msg.is_edited && <span className="edited">(edited)</span>}
  </div>
  <div 
    className="message-content"
    dangerouslySetInnerHTML={{ __html: formatMarkdown(msg.content) }}
  />
  {msg.is_self && canEdit(msg) && (
    <div className="message-actions">
      <button onClick={() => startEdit(msg)}>✏️</button>
      <button onClick={() => deleteMessage(msg)}>🗑️</button>
    </div>
  )}
</div>
```

---

## Out of Scope

- WebSocket real-time updates (keep polling for MVP)
- Private DMs (requires protocol update)
- Message reactions/emoji picker
- Message threads/replies
- Read receipts
- Typing indicators
