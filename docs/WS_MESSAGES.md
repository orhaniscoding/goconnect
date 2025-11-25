# WebSocket Message Protocol

## Inbound Messages (Client -> Server)

| Type            | Description                  | Data Payload                                                                  |
| --------------- | ---------------------------- | ----------------------------------------------------------------------------- |
| `auth.refresh`  | Refresh authentication token | `{ "refresh_token": "..." }`                                                  |
| `chat.send`     | Send a chat message          | `{ "scope": "...", "body": "...", "attachments": [...], "parent_id": "..." }` |
| `chat.edit`     | Edit a message               | `{ "message_id": "...", "new_body": "..." }`                                  |
| `chat.delete`   | Delete a message             | `{ "message_id": "...", "mode": "soft                                         | hard" }`                   |
| `chat.redact`   | Redact a message (Mod/Admin) | `{ "message_id": "...", "mask": "..." }`                                      |
| `chat.read`     | Mark message as read         | `{ "message_id": "...", "room": "..." }`                                      |
| `chat.reaction` | Add/remove reaction          | `{ "message_id": "...", "reaction": "...", "action": "add                     | remove", "scope": "..." }` |
| `chat.typing`   | Typing indicator             | `{ "scope": "...", "typing": true                                             | false }`                   |
| `file.upload`   | File upload progress         | `{ "scope": "...", "fileId": "...", "progress": 0-100, "isComplete": bool }`  |
| `call.offer`    | WebRTC Offer                 | `{ "target_id": "...", "call_type": "audio                                    | video                      | screen", "sdp": ... }` |
| `call.answer`   | WebRTC Answer                | `{ "target_id": "...", "sdp": ... }`                                          |
| `call.ice`      | WebRTC ICE Candidate         | `{ "target_id": "...", "candidate": ... }`                                    |
| `call.end`      | End call                     | `{ "target_id": "...", "reason": "..." }`                                     |
| `room.join`     | Join a room                  | `{ "room": "..." }`                                                           |
| `room.leave`    | Leave a room                 | `{ "room": "..." }`                                                           |
| `presence.set`  | Set presence status          | `{ "status": "online                                                          | away                       | busy                   | offline" }` |
| `presence.ping` | Keep-alive ping              | (empty)                                                                       |

## Outbound Messages (Server -> Client)

| Type                   | Description            | Data Payload                                                                    |
| ---------------------- | ---------------------- | ------------------------------------------------------------------------------- |
| `chat.message`         | New message received   | `ChatMessage` object                                                            |
| `chat.edited`          | Message edited         | `ChatMessage` object                                                            |
| `chat.deleted`         | Message deleted        | `{ "message_id": "...", "deleted_at": "..." }`                                  |
| `chat.redacted`        | Message redacted       | `{ "message_id": "...", "new_body": "..." }`                                    |
| `chat.read.update`     | Read receipt update    | `{ "message_id": "...", "user_id": "...", "room": "..." }`                      |
| `chat.reaction.update` | Reaction update        | `{ "message_id": "...", "user_id": "...", "reaction": "...", "action": "..." }` |
| `chat.typing.user`     | User typing status     | `{ "scope": "...", "user_id": "...", "typing": true                             | false }` |
| `file.upload.event`    | File upload progress   | `{ "scope": "...", "user_id": "...", "fileId": "...", "progress": ... }`        |
| `call.offer`           | Incoming call offer    | `{ "from_user": "...", "call_type": "...", "sdp": ... }`                        |
| `call.answer`          | Incoming call answer   | `{ "from_user": "...", "sdp": ... }`                                            |
| `call.ice`             | Incoming ICE candidate | `{ "from_user": "...", "candidate": ... }`                                      |
| `call.end`             | Call ended             | `{ "from_user": "...", "reason": "..." }`                                       |
| `member.joined`        | User joined network    | `{ "network_id": "...", "user_id": "..." }`                                     |
| `member.left`          | User left network      | `{ "network_id": "...", "user_id": "..." }`                                     |
| `device.online`        | Device came online     | `{ "device_id": "...", "user_id": "..." }`                                      |
| `device.offline`       | Device went offline    | `{ "device_id": "...", "user_id": "..." }`                                      |
| `presence.update`      | User presence changed  | `{ "user_id": "...", "status": "..." }`                                         |
| `presence.pong`        | Pong response          | `{ "timestamp": "..." }`                                                        |
| `ack`                  | Acknowledgment         | `{ "op_id": "...", "data": ... }`                                               |
| `error`                | Error response         | `{ "code": "...", "message": "..." }`                                           |

## Tenant-Specific Messages (v2.9.0+)

### Inbound Messages (Client -> Server)

| Type                      | Description                    | Data Payload                                                                           |
| ------------------------- | ------------------------------ | -------------------------------------------------------------------------------------- |
| `tenant.chat.send`        | Send message to tenant chat    | `{ "tenant_id": "...", "content": "..." }`                                             |
| `tenant.chat.edit`        | Edit tenant chat message       | `{ "tenant_id": "...", "message_id": "...", "content": "..." }`                        |
| `tenant.chat.delete`      | Delete tenant chat message     | `{ "tenant_id": "...", "message_id": "..." }`                                          |
| `tenant.chat.typing`      | Typing indicator in tenant     | `{ "tenant_id": "...", "typing": true \| false }`                                      |
| `tenant.join`             | Join tenant room for updates   | `{ "tenant_id": "..." }`                                                               |
| `tenant.leave`            | Leave tenant room              | `{ "tenant_id": "..." }`                                                               |

### Outbound Messages (Server -> Client)

| Type                       | Description                    | Data Payload                                                                          |
| -------------------------- | ------------------------------ | ------------------------------------------------------------------------------------- |
| `tenant.chat.message`      | New tenant chat message        | `TenantChatMessage` object                                                            |
| `tenant.chat.edited`       | Tenant chat message edited     | `TenantChatMessage` object                                                            |
| `tenant.chat.deleted`      | Tenant chat message deleted    | `{ "tenant_id": "...", "message_id": "...", "deleted_at": "..." }`                    |
| `tenant.chat.typing.user`  | User typing in tenant chat     | `{ "tenant_id": "...", "user_id": "...", "typing": true \| false }`                   |
| `tenant.announcement`      | New/updated announcement       | `TenantAnnouncement` object                                                           |
| `tenant.announcement.deleted` | Announcement deleted        | `{ "tenant_id": "...", "announcement_id": "..." }`                                    |
| `tenant.member.joined`     | User joined tenant             | `{ "tenant_id": "...", "user_id": "...", "role": "member" }`                          |
| `tenant.member.left`       | User left tenant               | `{ "tenant_id": "...", "user_id": "..." }`                                            |
| `tenant.member.kicked`     | User was kicked                | `{ "tenant_id": "...", "user_id": "...", "kicked_by": "...", "reason": "..." }`       |
| `tenant.member.banned`     | User was banned                | `{ "tenant_id": "...", "user_id": "...", "banned_by": "...", "reason": "..." }`       |
| `tenant.member.role_changed` | Member role updated          | `{ "tenant_id": "...", "user_id": "...", "old_role": "...", "new_role": "..." }`      |
| `tenant.owner.transferred` | Ownership transferred          | `{ "tenant_id": "...", "old_owner": "...", "new_owner": "..." }`                      |
| `tenant.settings.updated`  | Tenant settings changed        | `{ "tenant_id": "...", "changes": {...} }`                                            |

---

## Tenant Chat Message Object

```json
{
  "id": "msg_01JQKP7F4ZWX...",
  "tenant_id": "tenant_01JQKP7F...",
  "user_id": "user_01JQKP7F...",
  "user": {
    "id": "user_01JQKP7F...",
    "display_name": "JohnDoe",
    "nickname": "Johnny"  // Tenant-specific nickname
  },
  "content": "Hello everyone!",
  "created_at": "2025-01-15T10:30:00Z",
  "edited_at": null
}
```

## Tenant Announcement Object

```json
{
  "id": "ann_01JQKP7F4ZWX...",
  "tenant_id": "tenant_01JQKP7F...",
  "title": "Server Maintenance",
  "content": "We will be performing maintenance...",
  "author_id": "user_01JQKP7F...",
  "author": {
    "id": "user_01JQKP7F...",
    "display_name": "Admin",
    "role": "admin"
  },
  "is_pinned": true,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

## Example: Joining Tenant Room

```javascript
// Join tenant room to receive real-time updates
ws.send(JSON.stringify({
  type: "tenant.join",
  data: { tenant_id: "tenant_01JQKP7F..." }
}));

// Send chat message
ws.send(JSON.stringify({
  type: "tenant.chat.send",
  data: {
    tenant_id: "tenant_01JQKP7F...",
    content: "Hey everyone!"
  }
}));

// Handle incoming messages
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  switch(msg.type) {
    case "tenant.chat.message":
      console.log("New message:", msg.data.content);
      break;
    case "tenant.member.joined":
      console.log("User joined:", msg.data.user_id);
      break;
    case "tenant.announcement":
      console.log("Announcement:", msg.data.title);
      break;
  }
};
```

## Tenant Role Permissions

| Role       | Level | Can Kick | Can Ban | Can Manage Roles | Can Delete Messages | Can Create Announcements |
| ---------- | ----- | -------- | ------- | ---------------- | ------------------- | ------------------------ |
| `owner`    | 5     | ✅       | ✅      | ✅ (all)         | ✅                  | ✅                       |
| `admin`    | 4     | ✅       | ✅      | ✅ (below admin) | ✅                  | ✅                       |
| `moderator`| 3     | ✅       | ❌      | ❌               | ✅                  | ✅                       |
| `vip`      | 2     | ❌       | ❌      | ❌               | ❌ (own only)       | ❌                       |
| `member`   | 1     | ❌       | ❌      | ❌               | ❌ (own only)       | ❌                       |
