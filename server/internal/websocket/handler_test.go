package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test handler
func newTestHandler() *DefaultMessageHandler {
	hub := NewHub(nil)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()

	// Seed user-1 for auth tests
	if err := userRepo.Create(context.Background(), &domain.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "user1@example.com",
	}); err != nil {
		panic(err)
	}

	authService := service.NewAuthService(userRepo, tenantRepo)

	handler := &DefaultMessageHandler{
		hub:         hub,
		chatService: nil, // Will be set per test if needed
		authService: authService,
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

	// Generate valid refresh token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "user-1",
		"user_id": "user-1",
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	secret := []byte("dev-secret-change-in-production")
	tokenString, _ := token.SignedString(secret)

	msg := &InboundMessage{
		Type: TypeAuthRefresh,
		OpID: "op-123",
		Data: json.RawMessage(fmt.Sprintf(`{"refresh_token":"%s"}`, tokenString)),
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
	// Generate valid refresh token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "user-1",
		"user_id": "user-1",
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	secret := []byte("dev-secret-change-in-production")
	tokenString, _ := token.SignedString(secret)
	validRefreshData := []byte(fmt.Sprintf(`{"refresh_token":"%s"}`, tokenString))

	tests := []struct {
		name        string
		messageType MessageType
		data        json.RawMessage
		expectError bool
		errorText   string
	}{
		{
			name:        "AuthRefresh routes correctly",
			messageType: TypeAuthRefresh,
			data:        validRefreshData,
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
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "host", "Original message", nil, "")

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
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "host", "Message to delete", nil, "")

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
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "host", "Message to delete", nil, "")

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
	msg, _ := chatService.SendMessage(context.Background(), "user-1", "tenant-1", "host", "Message to redact", nil, "")

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

func setupAuthorizationTest() (*DefaultMessageHandler, *repository.InMemoryMembershipRepository, *repository.InMemoryUserRepository) {
	hub := NewHub(nil)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()

	authService := service.NewAuthService(userRepo, tenantRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idemRepo)

	handler := NewDefaultMessageHandler(hub, nil, membershipService, nil, authService)

	return handler, membershipRepo, userRepo
}

func TestHandleRoomJoin_Authorization(t *testing.T) {
	handler, membershipRepo, userRepo := setupAuthorizationTest()

	// Setup users
	adminUser := &domain.User{ID: "admin-1", TenantID: "tenant-1", Email: "admin@example.com", IsAdmin: true}
	normalUser := &domain.User{ID: "user-1", TenantID: "tenant-1", Email: "user@example.com", IsAdmin: false}
	otherUser := &domain.User{ID: "user-2", TenantID: "tenant-1", Email: "other@example.com", IsAdmin: false}

	_ = userRepo.Create(context.Background(), adminUser)
	_ = userRepo.Create(context.Background(), normalUser)
	_ = userRepo.Create(context.Background(), otherUser)

	// Setup membership
	networkID := "net-1"
	_, _ = membershipRepo.UpsertApproved(context.Background(), networkID, normalUser.ID, domain.RoleMember, time.Now())

	tests := []struct {
		name      string
		userID    string
		isAdmin   bool
		room      string
		wantError bool
	}{
		{
			name:      "Member joining network room",
			userID:    normalUser.ID,
			isAdmin:   false,
			room:      "network:" + networkID,
			wantError: false,
		},
		{
			name:      "Non-member joining network room",
			userID:    otherUser.ID,
			isAdmin:   false,
			room:      "network:" + networkID,
			wantError: true,
		},
		{
			name:      "Admin joining host room",
			userID:    adminUser.ID,
			isAdmin:   true,
			room:      "host",
			wantError: false,
		},
		{
			name:      "Non-admin joining host room",
			userID:    normalUser.ID,
			isAdmin:   false,
			room:      "host",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.userID)
			client.isAdmin = tt.isAdmin

			msg := &InboundMessage{
				Type: TypeRoomJoin,
				OpID: "op-1",
				Data: json.RawMessage(fmt.Sprintf(`{"room":"%s"}`, tt.room)),
			}

			err := handler.HandleMessage(context.Background(), client, msg)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, client.rooms[tt.room])
			}
		})
	}
}

func TestHandlePresenceSet(t *testing.T) {
	handler := newTestHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handler.hub.Run(ctx)

	client := newTestClient("user-1")
	// Register client to hub
	handler.hub.Register(client)

	// Join a room to verify broadcast
	room := "network:net-1"
	handler.hub.JoinRoom(client, room)

	// Another client in the same room
	otherClient := newTestClient("user-2")
	handler.hub.Register(otherClient)
	handler.hub.JoinRoom(otherClient, room)

	// Wait for registration and join to process
	time.Sleep(50 * time.Millisecond)

	// Set presence
	msg := &InboundMessage{
		Type: TypePresenceSet,
		OpID: "op-1",
		Data: json.RawMessage(`{"status":"away"}`),
	}

	err := handler.HandleMessage(context.Background(), client, msg)
	require.NoError(t, err)

	// Verify client status updated
	client.mu.RLock()
	assert.Equal(t, StatusAway, client.status)
	client.mu.RUnlock()

	// Verify ack
	select {
	case ackMsg := <-client.send:
		var outbound OutboundMessage
		err := json.Unmarshal(ackMsg, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeAck, outbound.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack message")
	}

	// Verify broadcast to other client
	select {
	case msgBytes := <-otherClient.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypePresenceUpdate, outbound.Type)

		var data PresenceUpdateData
		// Re-marshal data to parse into struct
		dataBytes, _ := json.Marshal(outbound.Data)
		err = json.Unmarshal(dataBytes, &data)
		require.NoError(t, err)

		assert.Equal(t, "user-1", data.UserID)
		assert.Equal(t, "away", data.Status)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected presence update broadcast")
	}
}

func setupMembershipTest() (*DefaultMessageHandler, *service.MembershipService, *repository.InMemoryMembershipRepository, *repository.InMemoryJoinRequestRepository, *repository.InMemoryNetworkRepository, *repository.InMemoryUserRepository) {
	hub := NewHub(nil)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()

	authService := service.NewAuthService(userRepo, tenantRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idemRepo)

	handler := NewDefaultMessageHandler(hub, nil, membershipService, nil, authService)

	// Start hub
	go hub.Run(context.Background())

	return handler, membershipService, membershipRepo, joinRepo, networkRepo, userRepo
}

func TestMembershipEvents(t *testing.T) {
	handler, membershipService, membershipRepo, joinRepo, networkRepo, userRepo := setupMembershipTest()

	// Setup data
	adminUser := &domain.User{ID: "admin-1", TenantID: "tenant-1", Email: "admin@example.com", IsAdmin: true}
	targetUser := &domain.User{ID: "user-1", TenantID: "tenant-1", Email: "user@example.com", IsAdmin: false}
	_ = userRepo.Create(context.Background(), adminUser)
	_ = userRepo.Create(context.Background(), targetUser)

	network := &domain.Network{ID: "net-1", TenantID: "tenant-1", Name: "Test Net", JoinPolicy: domain.JoinPolicyApproval}
	_ = networkRepo.Create(context.Background(), network)

	// Admin must be member/owner to approve
	_, _ = membershipRepo.UpsertApproved(context.Background(), network.ID, adminUser.ID, domain.RoleOwner, time.Now())

	// Create pending join request
	_, _ = joinRepo.CreatePending(context.Background(), network.ID, targetUser.ID)

	// Connect a client to the network room to receive broadcast
	client := newTestClient("observer-1")
	handler.hub.Register(client)
	handler.hub.JoinRoom(client, "network:"+network.ID)

	// Wait for join
	time.Sleep(50 * time.Millisecond)

	// Approve join request
	_, err := membershipService.Approve(context.Background(), network.ID, targetUser.ID, adminUser.ID, "tenant-1")
	require.NoError(t, err)

	// Verify broadcast
	select {
	case msgBytes := <-client.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeMemberJoined, outbound.Type)

		var data MemberEventData
		dataBytes, _ := json.Marshal(outbound.Data)
		err = json.Unmarshal(dataBytes, &data)
		require.NoError(t, err)

		assert.Equal(t, network.ID, data.NetworkID)
		assert.Equal(t, targetUser.ID, data.UserID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected member.joined broadcast")
	}
}

func setupDeviceTest() (*DefaultMessageHandler, *service.DeviceService, *repository.InMemoryDeviceRepository, *repository.InMemoryUserRepository) {
	hub := NewHub(nil)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()

	wgConfig := config.WireGuardConfig{}

	authService := service.NewAuthService(userRepo, tenantRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	handler := NewDefaultMessageHandler(hub, nil, nil, deviceService, authService)

	// Start hub
	go hub.Run(context.Background())

	return handler, deviceService, deviceRepo, userRepo
}

func TestDeviceEvents(t *testing.T) {
	handler, deviceService, deviceRepo, userRepo := setupDeviceTest()

	// Setup user
	user := &domain.User{ID: "user-1", TenantID: "tenant-1", Email: "user@example.com", IsAdmin: false}
	_ = userRepo.Create(context.Background(), user)

	// Setup device
	device := &domain.Device{
		ID: "dev-1", UserID: user.ID, TenantID: "tenant-1", Name: "Test Device",
		PubKey: "key1", Platform: "linux", LastSeen: time.Time{}, // Never seen
	}
	_ = deviceRepo.Create(context.Background(), device)

	// Connect client (user)
	client := newTestClient(user.ID)
	handler.hub.Register(client)

	// Wait for registration
	time.Sleep(50 * time.Millisecond)

	// Send heartbeat (should trigger online event because LastSeen is zero)
	req := &domain.DeviceHeartbeatRequest{IPAddress: "1.2.3.4"}
	err := deviceService.Heartbeat(context.Background(), device.ID, user.ID, "tenant-1", req)
	require.NoError(t, err)

	// Verify broadcast
	select {
	case msgBytes := <-client.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeDeviceOnline, outbound.Type)

		var data DeviceEventData
		dataBytes, _ := json.Marshal(outbound.Data)
		err = json.Unmarshal(dataBytes, &data)
		require.NoError(t, err)

		assert.Equal(t, device.ID, data.DeviceID)
		assert.Equal(t, user.ID, data.UserID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected device.online broadcast")
	}
}

func setupChatTest() (*DefaultMessageHandler, *service.ChatService, *repository.InMemoryUserRepository) {
	hub := NewHub(nil)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	chatRepo := repository.NewInMemoryChatRepository()

	authService := service.NewAuthService(userRepo, tenantRepo)
	chatService := service.NewChatService(chatRepo, userRepo)

	handler := NewDefaultMessageHandler(hub, chatService, nil, nil, authService)

	go hub.Run(context.Background())

	return handler, chatService, userRepo
}

func TestHandleChatSend_DirectMessage(t *testing.T) {
	handler, _, userRepo := setupChatTest()

	// Setup users
	sender := &domain.User{ID: "user-1", TenantID: "tenant-1", Email: "sender@example.com"}
	receiver := &domain.User{ID: "user-2", TenantID: "tenant-1", Email: "receiver@example.com"}
	other := &domain.User{ID: "user-3", TenantID: "tenant-1", Email: "other@example.com"}

	_ = userRepo.Create(context.Background(), sender)
	_ = userRepo.Create(context.Background(), receiver)
	_ = userRepo.Create(context.Background(), other)

	// Connect clients
	clientSender := newTestClient(sender.ID)
	clientReceiver := newTestClient(receiver.ID)
	clientOther := newTestClient(other.ID)

	handler.hub.Register(clientSender)
	handler.hub.Register(clientReceiver)
	handler.hub.Register(clientOther)

	time.Sleep(50 * time.Millisecond)

	// Send DM from sender to receiver
	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-dm",
		Data: json.RawMessage(fmt.Sprintf(`{"scope":"dm:%s", "body":"Hello DM"}`, receiver.ID)),
	}

	err := handler.HandleMessage(context.Background(), clientSender, msg)
	require.NoError(t, err)

	// Verify sender gets message then ack
	// 1. Message
	select {
	case msgBytes := <-clientSender.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, outbound.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message for sender")
	}
	// 2. Ack
	select {
	case msgBytes := <-clientSender.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeAck, outbound.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected ack for sender")
	}

	// Verify receiver gets message
	select {
	case msgBytes := <-clientReceiver.send:
		var outbound OutboundMessage
		err := json.Unmarshal(msgBytes, &outbound)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, outbound.Type)

		var data ChatMessageData
		dataBytes, _ := json.Marshal(outbound.Data)
		err = json.Unmarshal(dataBytes, &data)
		require.NoError(t, err)

		assert.Equal(t, "Hello DM", data.Body)
		assert.Equal(t, sender.ID, data.UserID)
		// Verify canonical scope
		assert.Equal(t, "dm:user-1:user-2", data.Scope)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message for receiver")
	}

	// Verify sender also got the message (sync across devices)
	// Note: clientSender already received the message in step 1 above (before ack)
	// The test logic above consumed 2 messages from clientSender: Message and Ack.
	// So we don't need to check again unless we expect duplicates.
	// The code above:
	// 1. Message
	// 2. Ack
	// So clientSender is done.

	// Verify third party did NOT get the message
	select {
	case <-clientOther.send:
		t.Fatal("third party should not receive DM")
	default:
		// OK
	}
}

func TestHandleChatRead(t *testing.T) {
	handler := newTestHandler()
	go handler.hub.Run(context.Background())

	// Setup clients
	sender := newTestClient("sender")
	sender.hub = handler.hub
	handler.hub.Register(sender)

	receiver := newTestClient("receiver")
	receiver.hub = handler.hub
	handler.hub.Register(receiver)

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Case 1: Room Read Receipt
	t.Run("Room Read Receipt", func(t *testing.T) {
		room := "network:test-net"
		handler.hub.JoinRoom(sender, room)
		handler.hub.JoinRoom(receiver, room)

		msgID := "msg-123"
		readData := ChatReadData{
			MessageID: msgID,
			Room:      room,
		}
		dataJSON, _ := json.Marshal(readData)

		msg := &InboundMessage{
			Type: TypeChatRead,
			Data: dataJSON,
		}

		err := handler.HandleMessage(context.Background(), receiver, msg)
		require.NoError(t, err)

		// Sender should receive update
		select {
		case raw := <-sender.send:
			var out OutboundMessage
			err := json.Unmarshal(raw, &out)
			require.NoError(t, err)
			assert.Equal(t, TypeChatReadUpdate, out.Type)

			dataBytes, _ := json.Marshal(out.Data)
			var update ChatReadUpdateData
			err = json.Unmarshal(dataBytes, &update)
			require.NoError(t, err)
			assert.Equal(t, msgID, update.MessageID)
			assert.Equal(t, receiver.userID, update.UserID)
			assert.Equal(t, room, update.Room)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for read update")
		}
	})

	// Case 2: DM Read Receipt
	t.Run("DM Read Receipt", func(t *testing.T) {
		msgID := "msg-dm-123"
		// receiver reads message from sender
		// room sent by client is "dm:sender" (target_id)
		readData := ChatReadData{
			MessageID: msgID,
			Room:      "dm:sender",
		}
		dataJSON, _ := json.Marshal(readData)

		msg := &InboundMessage{
			Type: TypeChatRead,
			Data: dataJSON,
		}

		err := handler.HandleMessage(context.Background(), receiver, msg)
		require.NoError(t, err)

		// Sender should receive update
		select {
		case raw := <-sender.send:
			var out OutboundMessage
			err := json.Unmarshal(raw, &out)
			require.NoError(t, err)
			assert.Equal(t, TypeChatReadUpdate, out.Type)

			dataBytes, _ := json.Marshal(out.Data)
			var update ChatReadUpdateData
			err = json.Unmarshal(dataBytes, &update)
			require.NoError(t, err)
			assert.Equal(t, msgID, update.MessageID)
			assert.Equal(t, receiver.userID, update.UserID)
			// Canonical room name should be dm:receiver:sender (sorted)
			// receiver < sender
			expectedRoom := fmt.Sprintf("dm:%s:%s", receiver.userID, sender.userID)
			assert.Equal(t, expectedRoom, update.Room)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for read update")
		}
	})
}

func TestHandleCallSignal(t *testing.T) {
	handler := newTestHandler()
	go handler.hub.Run(context.Background())

	caller := newTestClient("caller")
	caller.hub = handler.hub
	handler.hub.Register(caller)

	callee := newTestClient("callee")
	callee.hub = handler.hub
	handler.hub.Register(callee)

	time.Sleep(10 * time.Millisecond)

	// Test Offer
	t.Run("Call Offer", func(t *testing.T) {
		sdp := map[string]interface{}{"type": "offer", "sdp": "mock-sdp"}
		sdpBytes, _ := json.Marshal(sdp)
		offerData := CallSignalData{
			TargetID: "callee",
			Signal:   sdpBytes,
		}
		dataJSON, _ := json.Marshal(offerData)
		msg := &InboundMessage{
			Type: TypeCallOffer,
			Data: dataJSON,
		}

		err := handler.HandleMessage(context.Background(), caller, msg)
		require.NoError(t, err)

		select {
		case raw := <-callee.send:
			var out OutboundMessage
			err := json.Unmarshal(raw, &out)
			require.NoError(t, err)
			assert.Equal(t, TypeCallOffer, out.Type)

			dataBytes, _ := json.Marshal(out.Data)
			var signal CallSignalData
			err = json.Unmarshal(dataBytes, &signal)
			require.NoError(t, err)
			assert.Equal(t, "caller", signal.FromUser)

			// Verify Signal content
			var signalMap map[string]interface{}
			err = json.Unmarshal(signal.Signal, &signalMap)
			require.NoError(t, err)
			assert.Equal(t, "offer", signalMap["type"])
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for offer")
		}
	})

	// Test Screen Share Offer
	t.Run("Screen Share Offer", func(t *testing.T) {
		sdp := map[string]interface{}{"type": "offer", "sdp": "screen-sdp"}
		sdpBytes, _ := json.Marshal(sdp)
		offerData := CallSignalData{
			TargetID: "callee",
			CallType: "screen",
			Signal:   sdpBytes,
		}
		dataJSON, _ := json.Marshal(offerData)
		msg := &InboundMessage{
			Type: TypeCallOffer,
			Data: dataJSON,
		}

		err := handler.HandleMessage(context.Background(), caller, msg)
		require.NoError(t, err)

		select {
		case raw := <-callee.send:
			var out OutboundMessage
			err := json.Unmarshal(raw, &out)
			require.NoError(t, err)
			assert.Equal(t, TypeCallOffer, out.Type)

			dataBytes, _ := json.Marshal(out.Data)
			var signal CallSignalData
			err = json.Unmarshal(dataBytes, &signal)
			require.NoError(t, err)
			assert.Equal(t, "caller", signal.FromUser)
			assert.Equal(t, "screen", signal.CallType)

			// Verify Signal content
			var signalMap map[string]interface{}
			err = json.Unmarshal(signal.Signal, &signalMap)
			require.NoError(t, err)
			assert.Equal(t, "screen-sdp", signalMap["sdp"])
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for screen share offer")
		}
	})
}

func TestHandleChatReaction(t *testing.T) {
	handler := newTestHandler()
	go handler.hub.Run(context.Background())

	sender := newTestClient("sender")
	sender.hub = handler.hub
	handler.hub.Register(sender)

	receiver := newTestClient("receiver")
	receiver.hub = handler.hub
	handler.hub.Register(receiver)

	time.Sleep(10 * time.Millisecond)

	t.Run("Room Reaction", func(t *testing.T) {
		room := "network:test-net"
		handler.hub.JoinRoom(sender, room)
		handler.hub.JoinRoom(receiver, room)

		msgID := "msg-123"
		reactionData := ChatReactionData{
			MessageID: msgID,
			Reaction:  "👍",
			Action:    "add",
			Scope:     room,
		}
		dataJSON, _ := json.Marshal(reactionData)

		msg := &InboundMessage{
			Type: TypeChatReaction,
			Data: dataJSON,
		}

		err := handler.HandleMessage(context.Background(), sender, msg)
		require.NoError(t, err)

		// Receiver should get update
		select {
		case raw := <-receiver.send:
			var out OutboundMessage
			err := json.Unmarshal(raw, &out)
			require.NoError(t, err)
			assert.Equal(t, TypeChatReactionUpdate, out.Type)

			dataBytes, _ := json.Marshal(out.Data)
			var update ChatReactionUpdateData
			err = json.Unmarshal(dataBytes, &update)
			require.NoError(t, err)
			assert.Equal(t, msgID, update.MessageID)
			assert.Equal(t, sender.userID, update.UserID)
			assert.Equal(t, "👍", update.Reaction)
			assert.Equal(t, "add", update.Action)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for reaction update")
		}
	})
}

func TestHandleChatTyping(t *testing.T) {
	handler := newTestHandler()
	go handler.hub.Run(context.Background())

	sender := newTestClient("sender")
	sender.hub = handler.hub
	handler.hub.Register(sender)

	receiver := newTestClient("receiver")
	receiver.hub = handler.hub
	handler.hub.Register(receiver)

	time.Sleep(10 * time.Millisecond)

	room := "network:test-net"
	handler.hub.JoinRoom(sender, room)
	handler.hub.JoinRoom(receiver, room)

	typingData := TypingData{
		Scope:  room,
		Typing: true,
	}
	dataJSON, _ := json.Marshal(typingData)

	msg := &InboundMessage{
		Type: TypeChatTyping,
		Data: dataJSON,
	}

	err := handler.HandleMessage(context.Background(), sender, msg)
	require.NoError(t, err)

	// Receiver should get update
	select {
	case raw := <-receiver.send:
		var out OutboundMessage
		err := json.Unmarshal(raw, &out)
		require.NoError(t, err)
		assert.Equal(t, TypeChatTypingUser, out.Type)

		dataBytes, _ := json.Marshal(out.Data)
		var update TypingUserData
		err = json.Unmarshal(dataBytes, &update)
		require.NoError(t, err)
		assert.Equal(t, room, update.Scope)
		assert.Equal(t, sender.userID, update.UserID)
		assert.True(t, update.Typing)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for typing update")
	}
}

func TestHandleFileUpload(t *testing.T) {
	handler := newTestHandler()
	go handler.hub.Run(context.Background())

	sender := newTestClient("sender")
	sender.hub = handler.hub
	handler.hub.Register(sender)

	receiver := newTestClient("receiver")
	receiver.hub = handler.hub
	handler.hub.Register(receiver)

	time.Sleep(10 * time.Millisecond)

	room := "network:test-net"
	handler.hub.JoinRoom(sender, room)
	handler.hub.JoinRoom(receiver, room)

	uploadData := FileUploadData{
		Scope:       room,
		FileID:      "file-123",
		FileName:    "test.pdf",
		Progress:    45.5,
		IsComplete:  false,
		DownloadURL: "",
	}
	dataJSON, _ := json.Marshal(uploadData)

	msg := &InboundMessage{
		Type: TypeFileUpload,
		Data: dataJSON,
	}

	err := handler.HandleMessage(context.Background(), sender, msg)
	require.NoError(t, err)

	// Receiver should get update
	select {
	case raw := <-receiver.send:
		var out OutboundMessage
		err := json.Unmarshal(raw, &out)
		require.NoError(t, err)
		assert.Equal(t, TypeFileUploadEvent, out.Type)

		dataMap, ok := out.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, room, dataMap["scope"])
		assert.Equal(t, sender.userID, dataMap["user_id"])
		assert.Equal(t, "file-123", dataMap["fileId"])
		assert.Equal(t, 45.5, dataMap["progress"])
		assert.Equal(t, false, dataMap["isComplete"])

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for file upload update")
	}
}
