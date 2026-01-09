# Task Breakdown: Text Chat Channels

## Overview
Total Tasks: 5 Task Groups
**Status: ✅ IMPLEMENTATION COMPLETE**

---

## Task List

### Task Group 1: Markdown Formatting ✅ COMPLETE

- [x] 1.1 Create `formatMarkdown()` utility function
  - Location: `desktop/src/lib/utils.ts`
  - XSS sanitization + bold/italic/code/mentions
- [x] 1.2 Create `canEditMessage()` utility (15-min window)
- [x] 1.3 Create `formatMessageTime()` utility (Today/Yesterday/Date)
- [x] 1.4 Update ChatPanel to render formatted messages
- [x] 1.5 Add CSS styles for `.inline-code`, `.mention`

---

### Task Group 2: Edit/Delete UI ✅ COMPLETE

- [x] 2.1 Add `editMessage()` and `deleteMessage()` to tauri-api.ts
- [x] 2.2 Add hover actions to message bubbles (✏️ 🗑️)
- [x] 2.3 Implement 15-minute edit window check
- [x] 2.4 Create inline edit mode with textarea
- [x] 2.5 Add delete confirmation modal
- [x] 2.6 Add "(edited)" indicator with CSS

---

### Task Group 3: Sender Display ✅ COMPLETE

- [x] 3.1 Update ChatMessage interface
  - Added: `peer_name`, `is_edited`, `is_deleted`
- [x] 3.2 Show peer names above message bubbles
- [x] 3.3 Fallback to ID prefix if no name

---

### Task Group 4: Infinite Scroll ✅ COMPLETE

- [x] 4.1 Add `onScroll` handler to messages container
- [x] 4.2 Implement `loadMoreMessages()` with `before` cursor
- [x] 4.3 Add loading indicator at top
- [x] 4.4 Track `hasMore` state
- [x] 4.5 Show "Beginning of chat" indicator

---

### Task Group 5: Notification Integration ⏸️ DEFERRED

- [ ] 5.1 Check notification settings on new message
- [ ] 5.2 Show desktop notification
- [ ] 5.3 Add unread badge to chat tab

*Deferred to Phase 3 (Notification System) for proper integration*

---

## Files Modified

| File | Changes |
|------|---------|
| `desktop/src/lib/utils.ts` | Added formatMarkdown, canEditMessage, formatMessageTime |
| `desktop/src/lib/tauri-api.ts` | Added peer_name, is_edited, is_deleted to interface; editMessage, deleteMessage methods |
| `desktop/src/components/ChatPanel.tsx` | Full rewrite with all features |
| `desktop/src/index.css` | Added .inline-code, .mention, .message-actions, .edited-indicator styles |

---

## Status: ✅ IMPLEMENTATION COMPLETE
