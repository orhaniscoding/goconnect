package websocket

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Inbound message types (client -> server)
	TypeAuthRefresh  MessageType = "auth.refresh"
	TypeChatSend     MessageType = "chat.send"
	TypeChatEdit     MessageType = "chat.edit"
	TypeChatDelete   MessageType = "chat.delete"
	TypeChatRedact   MessageType = "chat.redact"
	TypeChatTyping   MessageType = "chat.typing"   // Typing indicator
	TypeRoomJoin     MessageType = "room.join"     // Join a room
	TypeRoomLeave    MessageType = "room.leave"    // Leave a room
	TypePresencePing MessageType = "presence.ping" // Keep-alive ping

	// Outbound message types (server -> client)
	TypeChatMessage    MessageType = "chat.message"
	TypeChatEdited     MessageType = "chat.edited"
	TypeChatDeleted    MessageType = "chat.deleted"
	TypeChatRedacted   MessageType = "chat.redacted"
	TypeChatTypingUser MessageType = "chat.typing.user" // User typing indicator
	TypeMemberJoined   MessageType = "member.joined"
	TypeMemberLeft     MessageType = "member.left"
	TypeJoinPending    MessageType = "request.join.pending"
	TypeJoinApproved   MessageType = "request.join.approved"
	TypeJoinDenied     MessageType = "request.join.denied"
	TypeAdminKick      MessageType = "admin.kick"
	TypeAdminBan       MessageType = "admin.ban"
	TypeNetUpdated     MessageType = "net.updated"
	TypeDeviceOnline   MessageType = "device.online"
	TypeDeviceOffline  MessageType = "device.offline"
	TypePresencePong   MessageType = "presence.pong"   // Response to ping
	TypePresenceUpdate MessageType = "presence.update" // Presence status change

	// Control messages
	TypeError MessageType = "error"
	TypeAck   MessageType = "ack"
)

// InboundMessage represents a message from client to server
type InboundMessage struct {
	Type MessageType     `json:"type"`
	OpID string          `json:"op_id"` // Client-provided operation ID for response correlation
	Data json.RawMessage `json:"data"`
}

// OutboundMessage represents a message from server to client
type OutboundMessage struct {
	Type  MessageType `json:"type"`
	OpID  string      `json:"op_id,omitempty"` // Echo back for request/response
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorData  `json:"error,omitempty"`
}

// ErrorData represents error information in WebSocket messages
type ErrorData struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// AuthRefreshData represents data for auth.refresh messages
type AuthRefreshData struct {
	RefreshToken string `json:"refresh_token"`
}

// ChatSendData represents data for chat.send messages
type ChatSendData struct {
	Scope       string   `json:"scope"` // "host" or "network:<id>"
	Body        string   `json:"body"`
	Attachments []string `json:"attachments,omitempty"`
}

// ChatEditData represents data for chat.edit messages
type ChatEditData struct {
	MessageID string `json:"message_id"`
	NewBody   string `json:"new_body"`
}

// ChatDeleteData represents data for chat.delete messages
type ChatDeleteData struct {
	MessageID string `json:"message_id"`
	Mode      string `json:"mode"` // "soft" or "hard"
}

// ChatRedactData represents data for chat.redact messages
type ChatRedactData struct {
	MessageID string `json:"message_id"`
	Mask      string `json:"mask"` // Pattern for redaction
}

// ChatMessageData represents data for chat.message events
type ChatMessageData struct {
	ID          string   `json:"id"`
	Scope       string   `json:"scope"`
	UserID      string   `json:"user_id"`
	Body        string   `json:"body"`
	Redacted    bool     `json:"redacted"`
	DeletedAt   *string  `json:"deleted_at,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// MemberEventData represents data for member.joined/left events
type MemberEventData struct {
	NetworkID string `json:"network_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role,omitempty"`
}

// JoinPendingData represents data for request.join.pending events
type JoinPendingData struct {
	NetworkID string `json:"network_id"`
	UserID    string `json:"user_id"`
	RequestID string `json:"request_id"`
}

// AdminActionData represents data for admin.kick/ban events
type AdminActionData struct {
	NetworkID string `json:"network_id"`
	UserID    string `json:"user_id"`
	Reason    string `json:"reason,omitempty"`
}

// NetUpdatedData represents data for net.updated events
type NetUpdatedData struct {
	NetworkID  string                 `json:"network_id"`
	Changes    []string               `json:"changes"`    // List of changed fields
	UpdatedBy  string                 `json:"updated_by"` // User who made the change
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TypingData represents data for chat.typing messages
type TypingData struct {
	Scope  string `json:"scope"`  // "host" or "network:<id>"
	Typing bool   `json:"typing"` // true = started typing, false = stopped typing
}

// TypingUserData represents data for chat.typing.user events
type TypingUserData struct {
	Scope  string `json:"scope"`
	UserID string `json:"user_id"`
	Typing bool   `json:"typing"`
}

// RoomJoinData represents data for room.join messages
type RoomJoinData struct {
	Room string `json:"room"` // Room name to join (e.g., "network:<id>", "host")
}

// RoomLeaveData represents data for room.leave messages
type RoomLeaveData struct {
	Room string `json:"room"` // Room name to leave
}

// DeviceEventData represents data for device.online/offline events
type DeviceEventData struct {
	DeviceID  string `json:"device_id"`
	UserID    string `json:"user_id"`
	NetworkID string `json:"network_id,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

// PresenceUpdateData represents data for presence.update events
type PresenceUpdateData struct {
	UserID string `json:"user_id"`
	Status string `json:"status"` // "online", "away", "offline"
	Since  string `json:"since"`  // ISO 8601 timestamp
}
