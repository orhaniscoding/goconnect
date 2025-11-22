package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// DefaultMessageHandler implements MessageHandler interface
type DefaultMessageHandler struct {
	hub               *Hub
	chatService       *service.ChatService
	membershipService *service.MembershipService
	deviceService     *service.DeviceService
	authService       *service.AuthService
}

// NewDefaultMessageHandler creates a new default message handler
func NewDefaultMessageHandler(hub *Hub, chatService *service.ChatService, membershipService *service.MembershipService, deviceService *service.DeviceService, authService *service.AuthService) *DefaultMessageHandler {
	h := &DefaultMessageHandler{
		hub:               hub,
		chatService:       chatService,
		membershipService: membershipService,
		deviceService:     deviceService,
		authService:       authService,
	}
	if membershipService != nil {
		membershipService.SetNotifier(h)
	}
	if deviceService != nil {
		deviceService.SetNotifier(h)
	}
	return h
}

// SetHub sets the hub for the message handler
func (h *DefaultMessageHandler) SetHub(hub *Hub) {
	h.hub = hub
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
	case TypePresenceSet:
		return h.handlePresenceSet(ctx, client, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleAuthRefresh handles auth.refresh messages
func (h *DefaultMessageHandler) handleAuthRefresh(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data AuthRefreshData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid auth.refresh data: %w", err)
	}

	if data.RefreshToken == "" {
		return fmt.Errorf("refresh_token is required")
	}

	// Call AuthService to refresh tokens
	// We need to construct a domain.RefreshRequest
	req := &domain.RefreshRequest{
		RefreshToken: data.RefreshToken,
	}

	resp, err := h.authService.Refresh(ctx, req)
	if err != nil {
		return err
	}

	// Send new tokens back in ack
	client.sendAck(msg.OpID, map[string]interface{}{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"expires_in":    resp.ExpiresIn,
		"status":        "refreshed",
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

	// Validate user has access to this room
	if strings.HasPrefix(data.Room, "network:") {
		networkID := strings.TrimPrefix(data.Room, "network:")
		isMember, err := h.membershipService.IsMember(ctx, networkID, client.userID)
		if err != nil {
			return fmt.Errorf("failed to verify membership: %w", err)
		}
		if !isMember {
			return fmt.Errorf("user is not a member of this network")
		}
	} else if data.Room == "host" {
		if !client.isAdmin {
			return fmt.Errorf("access denied to host room")
		}
	} else {
		return fmt.Errorf("unknown room type or access denied: %s", data.Room)
	}

	// Join room
	h.hub.JoinRoom(client, data.Room)

	// Broadcast presence update
	h.hub.Broadcast(data.Room, &OutboundMessage{
		Type: TypePresenceUpdate,
		Data: &PresenceUpdateData{
			UserID: client.userID,
			Status: "online",
			Since:  time.Now().Format(time.RFC3339),
		},
	}, client)

	// Send list of existing members to the joining client
	existingClients := h.hub.GetRoomClients(data.Room)
	for _, existingClient := range existingClients {
		if existingClient.userID == client.userID {
			continue
		}
		client.sendMessage(&OutboundMessage{
			Type: TypePresenceUpdate,
			Data: &PresenceUpdateData{
				UserID: existingClient.userID,
				Status: "online",
				Since:  existingClient.lastActivity.Format(time.RFC3339),
			},
		})
	}

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

	// Broadcast presence update
	h.hub.Broadcast(data.Room, &OutboundMessage{
		Type: TypePresenceUpdate,
		Data: &PresenceUpdateData{
			UserID: client.userID,
			Status: "offline",
			Since:  time.Now().Format(time.RFC3339),
		},
	}, client)

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

// handlePresenceSet handles presence.set messages
func (h *DefaultMessageHandler) handlePresenceSet(ctx context.Context, client *Client, msg *InboundMessage) error {
	var data PresenceSetData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("invalid presence.set data: %w", err)
	}

	// Validate status
	status := PresenceStatus(data.Status)
	switch status {
	case StatusOnline, StatusAway, StatusBusy, StatusOffline:
		// valid
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update client status
	client.mu.Lock()
	client.status = status
	client.mu.Unlock()

	// Broadcast update to all rooms the client is in
	updateMsg := &OutboundMessage{
		Type: TypePresenceUpdate,
		Data: PresenceUpdateData{
			UserID: client.userID,
			Status: string(status),
			Since:  time.Now().Format(time.RFC3339),
		},
	}

	client.mu.RLock()
	rooms := make([]string, 0, len(client.rooms))
	for room := range client.rooms {
		rooms = append(rooms, room)
	}
	client.mu.RUnlock()

	for _, room := range rooms {
		h.hub.Broadcast(room, updateMsg, client)
	}

	// Send ack
	client.sendAck(msg.OpID, nil)

	return nil
}

// MemberJoined implements MembershipNotifier
func (h *DefaultMessageHandler) MemberJoined(networkID, userID string) {
	room := "network:" + networkID
	msg := &OutboundMessage{
		Type: TypeMemberJoined,
		Data: MemberEventData{
			NetworkID: networkID,
			UserID:    userID,
		},
	}
	h.hub.Broadcast(room, msg, nil)
}

// MemberLeft implements MembershipNotifier
func (h *DefaultMessageHandler) MemberLeft(networkID, userID string) {
	room := "network:" + networkID
	msg := &OutboundMessage{
		Type: TypeMemberLeft,
		Data: MemberEventData{
			NetworkID: networkID,
			UserID:    userID,
		},
	}
	h.hub.Broadcast(room, msg, nil)
}

// DeviceOnline implements DeviceNotifier
func (h *DefaultMessageHandler) DeviceOnline(deviceID, userID string) {
	msg := &OutboundMessage{
		Type: TypeDeviceOnline,
		Data: DeviceEventData{
			DeviceID: deviceID,
			UserID:   userID,
		},
	}
	h.hub.BroadcastToUser(userID, msg)
}

// DeviceOffline implements DeviceNotifier
func (h *DefaultMessageHandler) DeviceOffline(deviceID, userID string) {
	msg := &OutboundMessage{
		Type: TypeDeviceOffline,
		Data: DeviceEventData{
			DeviceID: deviceID,
			UserID:   userID,
		},
	}
	h.hub.BroadcastToUser(userID, msg)
}
