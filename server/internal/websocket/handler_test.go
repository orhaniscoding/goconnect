package websocket

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test handler
func newTestHandler() *DefaultMessageHandler {
	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: nil, // Will be set per test if needed
	}
	return handler
}

func newTestClient(userID string) *Client {
	return &Client{
		userID:      userID,
		tenantID:    "tenant-1",
		isAdmin:     false,
		isModerator: false,
		send:        make(chan []byte, 256),
		rooms:       make(map[string]bool),
	}
}

func TestHandler_UnknownMessageType(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")
	msg := &InboundMessage{
		Type: "unknown.type",
		OpID: "op-1",
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
}

func TestHandleAuthRefresh_SendsAck(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")
	msg := &InboundMessage{
		Type: TypeAuthRefresh,
		OpID: "op-123",
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.NoError(t, err)

	// Should receive an ack message
	select {
	case ackMsg := <-client.send:
		var outbound OutboundMessage
		err := json.Unmarshal(ackMsg, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeAck, outbound.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack message but got timeout")
	}
}

func TestHandleChatSend_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		data      ChatSendData
		wantError string
	}{
		{
			name: "Missing scope",
			data: ChatSendData{
				Scope: "",
				Body:  "Hello",
			},
			wantError: "scope is required",
		},
		{
			name: "Missing body",
			data: ChatSendData{
				Scope: "network:net-1",
				Body:  "",
			},
			wantError: "body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()
			client := newTestClient("user-1")

			dataBytes, _ := json.Marshal(tt.data)
			msg := &InboundMessage{
				Type: TypeChatSend,
				OpID: "op-1",
				Data: dataBytes,
			}

			err := handler.HandleMessage(context.Background(), client, msg)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestHandleChatSend_InvalidJSON(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")

	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-1",
		Data: []byte("invalid json {{{"),
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chat.send data")
}

func TestHandleChatEdit_InvalidJSON(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")

	msg := &InboundMessage{
		Type: TypeChatEdit,
		OpID: "op-1",
		Data: []byte("not valid json"),
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chat.edit data")
}

func TestHandleChatDelete_InvalidJSON(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")

	msg := &InboundMessage{
		Type: TypeChatDelete,
		OpID: "op-1",
		Data: []byte("bad json"),
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chat.delete data")
}

func TestHandleChatDelete_InvalidModeValidation(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")

	data := ChatDeleteData{
		MessageID: "msg-1",
		Mode:      "invalid-mode", // Not "soft" or "hard"
	}
	dataBytes, _ := json.Marshal(data)

	msg := &InboundMessage{
		Type: TypeChatDelete,
		OpID: "op-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode")
	assert.Contains(t, err.Error(), "must be soft or hard")
}

func TestHandleChatRedact_InvalidJSON(t *testing.T) {
	handler := newTestHandler()
	client := newTestClient("user-1")

	msg := &InboundMessage{
		Type: TypeChatRedact,
		OpID: "op-1",
		Data: []byte("corrupt json"),
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chat.redact data")
}

// Test with nil chatService to ensure error handling
func TestHandleChatSend_NilChatServicePanics(t *testing.T) {
	handler := newTestHandler()
	handler.chatService = nil // Explicitly nil
	client := newTestClient("user-1")

	data := ChatSendData{
		Scope: "network:net-1",
		Body:  "Hello World",
	}
	dataBytes, _ := json.Marshal(data)

	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-1",
		Data: dataBytes,
	}

	// Should panic because chatService is nil
	assert.Panics(t, func() {
		handler.HandleMessage(context.Background(), client, msg)
	})
}

// Verify that handler implements MessageHandler interface
func TestHandler_ImplementsMessageHandlerInterface(t *testing.T) {
	handler := newTestHandler()
	handler.chatService = service.NewChatService(nil, nil)

	// This should compile - proving DefaultMessageHandler implements MessageHandler
	var _ MessageHandler = handler

	assert.NotNil(t, handler)
}

// Test message type routing with table-driven tests
func TestHandleMessage_RoutesCorrectly(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		data        []byte
		expectError bool
		errorText   string
	}{
		{
			name:        "AuthRefresh routes correctly",
			messageType: TypeAuthRefresh,
			data:        []byte("{}"),
			expectError: false,
		},
		{
			name:        "ChatSend with empty JSON fails validation",
			messageType: TypeChatSend,
			data:        []byte("{}"),
			expectError: true,
			errorText:   "scope is required",
		},
		{
			name:        "ChatDelete with empty JSON fails parsing",
			messageType: TypeChatDelete,
			data:        []byte("{}"),
			expectError: true,
			errorText:   "invalid mode",
		},
		{
			name:        "Unknown type returns error",
			messageType: "unknown.type",
			data:        []byte("{}"),
			expectError: true,
			errorText:   "unknown message type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()
			client := newTestClient("user-1")
			msg := &InboundMessage{
				Type: tt.messageType,
				OpID: "op-1",
				Data: tt.data,
			}

			err := handler.HandleMessage(context.Background(), client, msg)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test data struct validation
func TestChatDeleteData_Modes(t *testing.T) {
	invalidModes := []string{"medium", "delete", "remove", ""}

	handler := newTestHandler()
	client := newTestClient("user-1")

	// Test invalid modes DO error on validation (before chatService is called)
	for _, mode := range invalidModes {
		data := ChatDeleteData{
			MessageID: "msg-1",
			Mode:      mode,
		}
		dataBytes, _ := json.Marshal(data)
		msg := &InboundMessage{
			Type: TypeChatDelete,
			OpID: "op-1",
			Data: dataBytes,
		}

		err := handler.HandleMessage(context.Background(), client, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mode")
	}
}

// Success case tests with real service

func setupChatTestService() (*service.ChatService, repository.UserRepository) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()

	// Create test user
	userRepo.Create(context.Background(), &domain.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "test@example.com",
	})

	return service.NewChatService(chatRepo, userRepo), userRepo
}

func TestHandleChatSend_Success(t *testing.T) {
	chatService, _ := setupChatTestService()
	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("user-1")

	data := ChatSendData{
		Scope: "network:net-1",
		Body:  "Hello from WebSocket!",
	}
	dataBytes, _ := json.Marshal(data)

	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-send-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case ackMsg := <-client.send:
		var outbound OutboundMessage
		json.Unmarshal(ackMsg, &outbound)
		assert.Equal(t, TypeAck, outbound.Type)

		// Verify ack contains message_id
		ackData := outbound.Data.(map[string]interface{})
		assert.NotEmpty(t, ackData["message_id"])
		assert.Equal(t, "sent", ackData["status"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack but got timeout")
	}
}

func TestHandleChatSend_WithAttachments(t *testing.T) {
	chatService, _ := setupChatTestService()
	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("user-1")

	data := ChatSendData{
		Scope:       "network:net-1",
		Body:        "Check out these files",
		Attachments: []string{"file1.pdf", "image.png"},
	}
	dataBytes, _ := json.Marshal(data)

	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-attach-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, msg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case <-client.send:
		// Success - ack received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack")
	}
}

func TestHandleChatEdit_Success(t *testing.T) {
	chatService, _ := setupChatTestService()

	// First create a message
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "network:net-1", "Original message", nil)

	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("user-1")

	data := ChatEditData{
		MessageID: msg.ID,
		NewBody:   "Edited message",
	}
	dataBytes, _ := json.Marshal(data)

	editMsg := &InboundMessage{
		Type: TypeChatEdit,
		OpID: "op-edit-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, editMsg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case <-client.send:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack")
	}
}

func TestHandleChatDelete_SuccessSoftMode(t *testing.T) {
	chatService, _ := setupChatTestService()

	// Create a message
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "network:net-1", "Message to delete", nil)

	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("user-1")

	data := ChatDeleteData{
		MessageID: msg.ID,
		Mode:      "soft",
	}
	dataBytes, _ := json.Marshal(data)

	deleteMsg := &InboundMessage{
		Type: TypeChatDelete,
		OpID: "op-del-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, deleteMsg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case <-client.send:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack")
	}
}

func TestHandleChatDelete_SuccessHardMode(t *testing.T) {
	chatService, _ := setupChatTestService()

	// Create a message
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "network:net-1", "Message to hard delete", nil)

	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("user-1")
	client.isAdmin = true // Admin can hard delete

	data := ChatDeleteData{
		MessageID: msg.ID,
		Mode:      "hard",
	}
	dataBytes, _ := json.Marshal(data)

	deleteMsg := &InboundMessage{
		Type: TypeChatDelete,
		OpID: "op-del-hard-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, deleteMsg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case <-client.send:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack")
	}
}

func TestHandleChatRedact_Success(t *testing.T) {
	chatService, _ := setupChatTestService()

	// Create a message
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "network:net-1", "Message with inappropriate content", nil)

	hub := NewHub(nil)
	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: chatService,
	}
	client := newTestClient("moderator-1")
	client.isModerator = true // Moderator can redact

	data := ChatRedactData{
		MessageID: msg.ID,
		Mask:      "[REDACTED]",
	}
	dataBytes, _ := json.Marshal(data)

	redactMsg := &InboundMessage{
		Type: TypeChatRedact,
		OpID: "op-redact-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(context.Background(), client, redactMsg)

	require.NoError(t, err)

	// Should receive ack
	select {
	case <-client.send:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack")
	}
}
