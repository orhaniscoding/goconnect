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
