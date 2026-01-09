# Requirements: Text Chat Channels

## User Answers Summary

### 1. Message Formatting
**Answer: (a) Simple Markdown**
- Bold (`**text**`), italic (`*text*`), code blocks
- @mentions highlight users
- No images/embeds for MVP

### 2. Edit Time Window  
**Answer: (a) 15 minutes**
- Discord-like behavior
- Edit button hidden after window expires
- Edit history tracked on backend

### 3. New Message Notifications
**Answer: (d) Honor existing settings**
- Reuse notification preferences from Settings
- Desktop toast + sound based on user choice
- Badge/highlight for minimized window

### 4. Sender Display
**Answer: (a) Display name**
- Use configured peer name
- Fallback: first 8 chars of peer ID
- Tooltip with full ID on hover (optional)

### 5. Message History
**Answer: (c) Infinite scroll**
- Load 50 messages initially
- Batch load on scroll up
- Loading indicator at top
- Client cap: 500 messages in memory

---

## Technical Decisions

### Markdown Parsing
```typescript
// Use simple regex-based parsing, no heavy library
const formatMessage = (text: string): string => {
  return text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code>$1</code>')
    .replace(/@(\w+)/g, '<span class="mention">@$1</span>');
};
```

### Edit Window Check
```typescript
const EDIT_WINDOW_MS = 15 * 60 * 1000; // 15 minutes

const canEdit = (msg: ChatMessage): boolean => {
  return msg.is_self && Date.now() - new Date(msg.timestamp).getTime() < EDIT_WINDOW_MS;
};
```

### Infinite Scroll
```typescript
// Intersection Observer for loading more
const observer = new IntersectionObserver(entries => {
  if (entries[0].isIntersecting) loadMoreMessages();
}, { threshold: 0.1 });
```

---

## Effort Estimate

| Task | Days |
|------|------|
| Markdown formatting | 0.5 |
| Edit/Delete UI | 1 |
| Sender names | 0.5 |
| Infinite scroll | 1 |
| Notification integration | 0.5 |
| Testing & polish | 1.5 |
| **Total** | **5 days** |
