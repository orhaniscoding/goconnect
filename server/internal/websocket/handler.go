package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// DefaultMessageHandler implements MessageHandler interface
type DefaultMessageHandler struct {
	hub         *Hub
	chatService *service.ChatService
}

// NewDefaultMessageHandler creates a new default message handler
func NewDefaultMessageHandler(hub *Hub, chatService *service.ChatService) *DefaultMessageHandler {
	return &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
}

// HandleMessage processes inbound WebSocket messages
func (h *DefaultMessageHandler) HandleMessage(ctx context.Context, client *Client, msg *InboundMessage) error {
	switch msg.Type {
	case TypeAuthRefresh:
		return h.handleAuthRefresh(ctx, client, msg)
	case TypeChatSend:
		return h.handleChatSend(ctx, client, msg)
	case TypeChatEdit:
		return h.handleChatEdit(ctx, client, msg)
	case TypeChatDelete:
		return h.handleChatDelete(ctx, client, msg)
	case TypeChatRedact:
		return h.handleChatRedact(ctx, client, msg)
	case TypeChatTyping:
		return h.handleChatTyping(ctx, client, msg)
	case TypeRoomJoin:
		return h.handleRoomJoin(ctx, client, msg)
	case TypeRoomLeave:
		return h.handleRoomLeave(ctx, client, msg)
	case TypePresencePing:
		return h.handlePresencePing(ctx, client, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleAuthRefresh handles auth.refresh messages
func (h *DefaultMessageHandler) handleAuthRefresh(ctx context.Context, client *Client, msg *InboundMessage) error {
	// TODO: Implement token refresh logic
	// For now, just acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"status": "refresh_not_implemented",
	})
	return nil
}

// handleChatSend handles chat.send messages
func (h *DefaultMessageHandler) handleChatSend(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data ChatSendData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid chat.send data: %w", err)
	}

	// Validate scope
	if data.Scope == "" {
		return fmt.Errorf("scope is required")
	}

	if data.Body == "" {
		return fmt.Errorf("body is required")
	}

	// Create message via chat service
	chatMsg, err := h.chatService.SendMessage(ctx, client.userID, client.tenantID, data.Scope, data.Body, data.Attachments)
	if err != nil {
		return err
	}

	// Broadcast to scope
	h.hub.Broadcast(data.Scope, &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:          chatMsg.ID,
			Scope:       chatMsg.Scope,
			UserID:      chatMsg.UserID,
			Body:        chatMsg.Body,
			Redacted:    chatMsg.Redacted,
			Attachments: chatMsg.Attachments,
			CreatedAt:   chatMsg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil)

	// Acknowledge to sender
	client.sendAck(msg.OpID, map[string]string{
		"message_id": chatMsg.ID,
		"status":     "sent",
	})

	return nil
}

// handleChatEdit handles chat.edit messages
func (h *DefaultMessageHandler) handleChatEdit(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data ChatEditData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid chat.edit data: %w", err)
	}

	// Edit message via chat service
	chatMsg, err := h.chatService.EditMessage(ctx, data.MessageID, client.userID, client.tenantID, data.NewBody, client.isAdmin)
	if err != nil {
		return err
	}

	// Broadcast edit event
	h.hub.Broadcast(chatMsg.Scope, &OutboundMessage{
		Type: TypeChatEdited,
		Data: map[string]interface{}{
			"message_id": chatMsg.ID,
			"new_body":   chatMsg.Body,
			"edited_at":  chatMsg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil)

	// Acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"status": "edited",
	})

	return nil
}

// handleChatDelete handles chat.delete messages
func (h *DefaultMessageHandler) handleChatDelete(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data ChatDeleteData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid chat.delete data: %w", err)
	}

	// Validate mode
	if data.Mode != "soft" && data.Mode != "hard" {
		return fmt.Errorf("invalid mode: %s (must be soft or hard)", data.Mode)
	}

	// Get message first to get scope
	chatMsg, err := h.chatService.GetMessage(ctx, data.MessageID, client.tenantID)
	if err != nil {
		return err
	}

	// Delete message via chat service
	if err := h.chatService.DeleteMessage(ctx, data.MessageID, client.userID, client.tenantID, data.Mode, client.isAdmin, client.isModerator); err != nil {
		return err
	}

	// Broadcast delete event
	h.hub.Broadcast(chatMsg.Scope, &OutboundMessage{
		Type: TypeChatDeleted,
		Data: map[string]interface{}{
			"message_id": data.MessageID,
			"mode":       data.Mode,
		},
	}, nil)

	// Acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"status": "deleted",
	})

	return nil
}

// handleChatRedact handles chat.redact messages
func (h *DefaultMessageHandler) handleChatRedact(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data ChatRedactData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid chat.redact data: %w", err)
	}

	// Get message first to get scope
	chatMsg, err := h.chatService.GetMessage(ctx, data.MessageID, client.tenantID)
	if err != nil {
		return err
	}

	// Redact message via chat service
	redactedMsg, err := h.chatService.RedactMessage(ctx, data.MessageID, client.userID, client.tenantID, client.isAdmin, client.isModerator, data.Mask)
	if err != nil {
		return err
	}

	// Broadcast redact event
	h.hub.Broadcast(chatMsg.Scope, &OutboundMessage{
		Type: TypeChatRedacted,
		Data: map[string]interface{}{
			"message_id":     data.MessageID,
			"redaction_mask": data.Mask,
			"redacted_by":    client.userID,
		},
	}, nil)

	// Acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"status":   "redacted",
		"new_body": redactedMsg.Body,
	})

	return nil
}

// handleChatTyping handles chat.typing messages
func (h *DefaultMessageHandler) handleChatTyping(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data TypingData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid chat.typing data: %w", err)
	}

	// Validate scope
	if data.Scope == "" {
		return fmt.Errorf("scope is required")
	}

	// Broadcast typing indicator to room (exclude sender)
	h.hub.Broadcast(data.Scope, &OutboundMessage{
		Type: TypeChatTypingUser,
		Data: &TypingUserData{
			Scope:  data.Scope,
			UserID: client.userID,
			Typing: data.Typing,
		},
	}, client)

	return nil
}

// handleRoomJoin handles room.join messages
func (h *DefaultMessageHandler) handleRoomJoin(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data RoomJoinData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid room.join data: %w", err)
	}

	// Validate room
	if data.Room == "" {
		return fmt.Errorf("room is required")
	}

	// TODO: Validate user has access to this room (check network membership)
	// For now, allow joining any room

	// Join room
	h.hub.JoinRoom(client, data.Room)

	// Acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"room":   data.Room,
		"status": "joined",
	})

	return nil
}

// handleRoomLeave handles room.leave messages
func (h *DefaultMessageHandler) handleRoomLeave(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data RoomLeaveData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid room.leave data: %w", err)
	}

	// Validate room
	if data.Room == "" {
		return fmt.Errorf("room is required")
	}

	// Leave room
	h.hub.LeaveRoom(client, data.Room)

	// Acknowledge
	client.sendAck(msg.OpID, map[string]string{
		"room":   data.Room,
		"status": "left",
	})

	return nil
}

// handlePresencePing handles presence.ping messages (keep-alive)
func (h *DefaultMessageHandler) handlePresencePing(ctx context.Context, client *Client, msg *InboundMessage) error {
	// Update client's last activity timestamp
	client.UpdateActivity()

	// Send pong response
	client.sendMessage(&OutboundMessage{
		Type: TypePresencePong,
		OpID: msg.OpID,
		Data: map[string]interface{}{
			"timestamp": client.lastActivity.Format("2006-01-02T15:04:05Z07:00"),
		},
	})

	return nil
}
