package websocket

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

// PresenceStatus defines user presence status
type PresenceStatus string

const (
	StatusOnline  PresenceStatus = "online"
	StatusAway    PresenceStatus = "away"
	StatusBusy    PresenceStatus = "busy"
	StatusOffline PresenceStatus = "offline"
)

const (
	// Inbound message types (client -> server)
	TypeAuthRefresh  MessageType = "auth.refresh"
	TypeChatSend     MessageType = "chat.send"
	TypeChatEdit     MessageType = "chat.edit"
	TypeChatDelete   MessageType = "chat.delete"
	TypeChatRedact   MessageType = "chat.redact"
	TypeChatTyping   MessageType = "chat.typing"   // Typing indicator
	TypeChatRead     MessageType = "chat.read"     // Mark message as read
	TypeChatReaction MessageType = "chat.reaction" // Add/remove reaction
	TypeRoomJoin     MessageType = "room.join"     // Join a room
	TypeRoomLeave    MessageType = "room.leave"    // Leave a room
	TypePresencePing MessageType = "presence.ping" // Keep-alive ping
	TypePresenceSet  MessageType = "presence.set"  // Set presence status

	// Outbound message types (server -> client)
	TypeChatMessage        MessageType = "chat.message"
	TypeChatEdited         MessageType = "chat.edited"
	TypeChatDeleted        MessageType = "chat.deleted"
	TypeChatRedacted       MessageType = "chat.redacted"
	TypeChatTypingUser     MessageType = "chat.typing.user"     // User typing indicator
	TypeChatReadUpdate     MessageType = "chat.read.update"     // Read receipt update
	TypeChatReactionUpdate MessageType = "chat.reaction.update" // Reaction update
	TypeCallOffer          MessageType = "call.offer"           // WebRTC Offer
	TypeCallAnswer         MessageType = "call.answer"          // WebRTC Answer
	TypeCallICE            MessageType = "call.ice"             // WebRTC ICE Candidate
	TypeCallEnd            MessageType = "call.end"             // End call
	TypeCallSignal         MessageType = "call.signal"          // Generic signaling (offer/answer/ice)
	TypeCallSignalEvent    MessageType = "call.signal.event"    // Generic signaling event
	TypeFileUpload         MessageType = "file.upload"          // File upload progress
	TypeFileUploadEvent    MessageType = "file.upload.event"    // File upload progress event
	TypeMemberJoined       MessageType = "member.joined"
	TypeMemberLeft         MessageType = "member.left"
	TypeJoinPending        MessageType = "request.join.pending"
	TypeJoinApproved       MessageType = "request.join.approved"
	TypeJoinDenied         MessageType = "request.join.denied"
	TypeAdminKick          MessageType = "admin.kick"
	TypeAdminBan           MessageType = "admin.ban"
	TypeNetUpdated         MessageType = "net.updated"
	TypeDeviceOnline       MessageType = "device.online"
	TypeDeviceOffline      MessageType = "device.offline"
	TypePresencePong       MessageType = "presence.pong"   // Response to ping
	TypePresenceUpdate     MessageType = "presence.update" // Presence status change
	TypeNotification       MessageType = "notification"    // Push notification

	// Control messages
	TypeError MessageType = "error"
	TypeAck   MessageType = "ack"

	// Tenant chat inbound messages (client -> server)
	TypeTenantChatSend   MessageType = "tenant.chat.send"
	TypeTenantChatEdit   MessageType = "tenant.chat.edit"
	TypeTenantChatDelete MessageType = "tenant.chat.delete"
	TypeTenantChatTyping MessageType = "tenant.chat.typing"
	TypeTenantJoin       MessageType = "tenant.join"
	TypeTenantLeave      MessageType = "tenant.leave"

	// Tenant chat outbound messages (server -> client)
	TypeTenantChatMessage       MessageType = "tenant.chat.message"
	TypeTenantChatEdited        MessageType = "tenant.chat.edited"
	TypeTenantChatDeleted       MessageType = "tenant.chat.deleted"
	TypeTenantChatTypingUser    MessageType = "tenant.chat.typing.user"
	TypeTenantAnnouncement      MessageType = "tenant.announcement"
	TypeTenantMemberJoined      MessageType = "tenant.member.joined"
	TypeTenantMemberLeft        MessageType = "tenant.member.left"
	TypeTenantMemberKicked      MessageType = "tenant.member.kicked"
	TypeTenantMemberRoleChanged MessageType = "tenant.member.role_changed"
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
	ParentID    string   `json:"parent_id,omitempty"` // For threads
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
	ParentID    string   `json:"parent_id,omitempty"` // For threads
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

// PresenceSetData represents data for presence.set messages
type PresenceSetData struct {
	Status string `json:"status"` // online, away, busy, offline
}

type PresenceUpdateData struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Since  string `json:"since,omitempty"`
}

// ChatReadData represents data for chat.read messages
type ChatReadData struct {
	MessageID string `json:"message_id"`
	Room      string `json:"room"`
}

// ChatReadUpdateData represents data for chat.read.update events
type ChatReadUpdateData struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ReadAt    string `json:"read_at"`
	Room      string `json:"room"`
}

// ChatReactionData represents data for chat.reaction messages
type ChatReactionData struct {
	Scope     string `json:"scope"` // "host" or "network:<id>"
	MessageID string `json:"message_id"`
	Reaction  string `json:"reaction"` // e.g. "👍"
	Action    string `json:"action"`   // "add" or "remove"
}

// ChatReactionUpdateData represents data for chat.reaction.update events
type ChatReactionUpdateData struct {
	Scope     string `json:"scope"`
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	Reaction  string `json:"reaction"`
	Action    string `json:"action"`
}

// CallSignalData represents data for call.* messages
type CallSignalData struct {
	TargetID string          `json:"targetId,omitempty"` // For inbound
	FromUser string          `json:"fromUser,omitempty"` // For outbound
	Signal   json.RawMessage `json:"signal"`
	CallType string          `json:"callType,omitempty"` // "audio", "video", "screen"
}

// FileUploadData represents file upload progress
type FileUploadData struct {
	Scope       string  `json:"scope"` // "host" or "network:<id>"
	FileID      string  `json:"fileId"`
	FileName    string  `json:"fileName"`
	Progress    float64 `json:"progress"` // 0-100
	IsComplete  bool    `json:"isComplete"`
	DownloadURL string  `json:"downloadUrl,omitempty"`
}

// Message represents the WebSocket message structure
type Message struct {
	Type      MessageType     `json:"type"`
	OpID      string          `json:"op_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     *ErrorData      `json:"error,omitempty"`
	Signature string          `json:"signature,omitempty"`
}

// ==================== TENANT CHAT DATA TYPES ====================

// TenantChatSendData represents data for tenant.chat.send messages
type TenantChatSendData struct {
	TenantID string `json:"tenant_id"`
	Content  string `json:"content"`
}

// TenantChatEditData represents data for tenant.chat.edit messages
type TenantChatEditData struct {
	TenantID  string `json:"tenant_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
}

// TenantChatDeleteData represents data for tenant.chat.delete messages
type TenantChatDeleteData struct {
	TenantID  string `json:"tenant_id"`
	MessageID string `json:"message_id"`
}

// TenantChatTypingData represents data for tenant.chat.typing messages
type TenantChatTypingData struct {
	TenantID string `json:"tenant_id"`
	Typing   bool   `json:"typing"`
}

// TenantJoinData represents data for tenant.join messages
type TenantJoinData struct {
	TenantID string `json:"tenant_id"`
}

// TenantLeaveData represents data for tenant.leave messages
type TenantLeaveData struct {
	TenantID string `json:"tenant_id"`
}

// TenantChatMessageData represents data for tenant.chat.message events
type TenantChatMessageData struct {
	ID        string              `json:"id"`
	TenantID  string              `json:"tenant_id"`
	UserID    string              `json:"user_id"`
	User      *TenantChatUserInfo `json:"user,omitempty"`
	Content   string              `json:"content"`
	CreatedAt string              `json:"created_at"`
	EditedAt  *string             `json:"edited_at,omitempty"`
}

// TenantChatUserInfo represents user info in tenant chat messages
type TenantChatUserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname,omitempty"`
}

// TenantChatTypingUserData represents data for tenant.chat.typing.user events
type TenantChatTypingUserData struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Typing   bool   `json:"typing"`
}

// TenantMemberEventData represents data for tenant.member.* events
type TenantMemberEventData struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role,omitempty"`
	OldRole  string `json:"old_role,omitempty"`
	NewRole  string `json:"new_role,omitempty"`
	By       string `json:"by,omitempty"` // Who performed the action
	Reason   string `json:"reason,omitempty"`
}
