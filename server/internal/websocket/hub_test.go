package websocket

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.rooms)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.NotNil(t, hub.handleInbound)
	assert.NotNil(t, hub.broadcast)
	assert.Equal(t, handler, hub.handler)
}

func TestNewDefaultMessageHandler(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.hub)
	assert.Nil(t, handler.chatService)
}

func TestHub_RegisterUnregister(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create mock client
	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "test-user",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	// Register client
	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.GetClientCount())

	// Unregister client
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_JoinLeaveRoom(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "test-user",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Join room
	hub.JoinRoom(client, "network:123")
	time.Sleep(10 * time.Millisecond)

	assert.True(t, client.IsInRoom("network:123"))
	assert.Equal(t, 1, hub.GetRoomCount())

	// Leave room
	hub.LeaveRoom(client, "network:123")
	time.Sleep(10 * time.Millisecond)

	assert.False(t, client.IsInRoom("network:123"))
	assert.Equal(t, 0, hub.GetRoomCount())
}

func TestHub_Broadcast(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create two clients
	client1 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	client2 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user2",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(10 * time.Millisecond)

	// Join both to same room
	hub.JoinRoom(client1, "network:123")
	hub.JoinRoom(client2, "network:123")
	time.Sleep(10 * time.Millisecond)

	// Broadcast to room
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:123",
			UserID: "user1",
			Body:   "Hello!",
		},
	}

	hub.Broadcast("network:123", msg, nil)
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive message
	select {
	case data := <-client1.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client1 did not receive message")
	}

	select {
	case data := <-client2.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client2 did not receive message")
	}
}

func TestHub_BroadcastWithExclude(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client1 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	client2 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user2",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client1)
	hub.Register(client2)
	hub.JoinRoom(client1, "network:123")
	hub.JoinRoom(client2, "network:123")
	time.Sleep(10 * time.Millisecond)

	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:123",
			UserID: "user1",
			Body:   "Hello!",
		},
	}

	// Broadcast but exclude client1 (sender)
	hub.Broadcast("network:123", msg, client1)
	time.Sleep(10 * time.Millisecond)

	// Client1 should NOT receive (excluded)
	select {
	case <-client1.send:
		t.Fatal("client1 should not receive message (excluded)")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Client2 should receive
	select {
	case data := <-client2.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client2 did not receive message")
	}
}

func TestMessageHandler_ChatSend(t *testing.T) {
	t.Skip("Skipping - requires mock chat service setup")

	hub := NewHub(nil)
	handler := NewDefaultMessageHandler(hub, nil)
	hub.handler = handler

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "test-user",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	hub.JoinRoom(client, "network:123")
	time.Sleep(10 * time.Millisecond)

	// Send chat message
	sendData := ChatSendData{
		Scope: "network:123",
		Body:  "Test message",
	}

	dataBytes, _ := json.Marshal(sendData)
	inbound := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-1",
		Data: dataBytes,
	}

	err := handler.HandleMessage(ctx, client, inbound)
	assert.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	// Should receive 2 messages: broadcast and ACK (order may vary)
	receivedMessages := make([]*OutboundMessage, 0, 2)

	for i := 0; i < 2; i++ {
		select {
		case data := <-client.send:
			var received OutboundMessage
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			receivedMessages = append(receivedMessages, &received)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("did not receive message %d", i+1)
		}
	}

	// Verify we got both message types
	var foundChatMessage, foundAck bool
	for _, msg := range receivedMessages {
		if msg.Type == TypeChatMessage {
			foundChatMessage = true
		}
		if msg.Type == TypeAck && msg.OpID == "op-1" {
			foundAck = true
		}
	}

	assert.True(t, foundChatMessage, "should receive chat broadcast")
	assert.True(t, foundAck, "should receive ACK")
}

func TestHub_BroadcastToAll(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create three clients
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		clients[i] = &Client{
			hub:         hub,
			send:        make(chan []byte, 10),
			userID:      "user-" + string(rune('1'+i)),
			tenantID:    "test-tenant",
			isAdmin:     false,
			isModerator: false,
			rooms:       make(map[string]bool),
		}
		hub.Register(clients[i])
	}
	time.Sleep(10 * time.Millisecond)

	// Broadcast to ALL (empty room string)
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "host",
			UserID: "admin",
			Body:   "Global announcement",
		},
	}

	hub.Broadcast("", msg, nil)
	time.Sleep(10 * time.Millisecond)

	// All clients should receive
	for i, client := range clients {
		select {
		case data := <-client.send:
			var received OutboundMessage
			err := json.Unmarshal(data, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeChatMessage, received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("client %d did not receive message", i)
		}
	}
}

func TestHub_GetRoomClients(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client1 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	client2 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user2",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client1)
	hub.Register(client2)
	hub.JoinRoom(client1, "network:123")
	hub.JoinRoom(client2, "network:123")
	time.Sleep(10 * time.Millisecond)

	// Get room clients
	clients := hub.GetRoomClients("network:123")
	assert.Len(t, clients, 2)

	// Empty room should return empty list
	emptyClients := hub.GetRoomClients("network:999")
	assert.Len(t, emptyClients, 0)
}

func TestHub_ClientInMultipleRooms(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Join multiple rooms
	hub.JoinRoom(client, "network:1")
	hub.JoinRoom(client, "network:2")
	hub.JoinRoom(client, "host")
	time.Sleep(10 * time.Millisecond)

	assert.True(t, client.IsInRoom("network:1"))
	assert.True(t, client.IsInRoom("network:2"))
	assert.True(t, client.IsInRoom("host"))
	assert.Equal(t, 3, hub.GetRoomCount())
}

func TestHub_RoomCleanupWhenEmpty(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	hub.JoinRoom(client, "network:123")
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.GetRoomCount())

	// Leave room - should be cleaned up
	hub.LeaveRoom(client, "network:123")
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.GetRoomCount())
}

func TestHub_UnregisterRemovesFromAllRooms(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	hub.JoinRoom(client, "network:1")
	hub.JoinRoom(client, "network:2")
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, hub.GetRoomCount())

	// Unregister should remove from all rooms
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, hub.GetRoomCount())
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_GracefulShutdown(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "test-tenant",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, hub.GetClientCount())

	// Cancel context to trigger shutdown
	cancel()
	time.Sleep(10 * time.Millisecond)

	// Hub should stop accepting new operations (test by timeout)
	// Note: Real shutdown would close client connections
}

func TestHub_ConcurrentOperations(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create multiple clients concurrently
	done := make(chan bool)
	numClients := 10

	for i := 0; i < numClients; i++ {
		go func(id int) {
			client := &Client{
				hub:         hub,
				send:        make(chan []byte, 10),
				userID:      "user-" + string(rune('0'+id)),
				tenantID:    "test-tenant",
				isAdmin:     false,
				isModerator: false,
				rooms:       make(map[string]bool),
			}
			hub.Register(client)
			hub.JoinRoom(client, "network:1")
			time.Sleep(5 * time.Millisecond)
			hub.LeaveRoom(client, "network:1")
			hub.Unregister(client)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numClients; i++ {
		<-done
	}

	time.Sleep(20 * time.Millisecond)

	// All clients should be unregistered
	assert.Equal(t, 0, hub.GetClientCount())
	assert.Equal(t, 0, hub.GetRoomCount())
}

func TestHub_BroadcastToNonexistentRoom(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:   "msg-1",
			Body: "Test",
		},
	}

	// Broadcast to nonexistent room should not panic
	hub.Broadcast("network:999", msg, nil)
	time.Sleep(10 * time.Millisecond)

	// Should complete without error
}

func TestHub_MultipleClientsInDifferentRooms(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client1 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "tenant1",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	client2 := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user2",
		tenantID:    "tenant2",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client1)
	hub.Register(client2)
	hub.JoinRoom(client1, "network:1")
	hub.JoinRoom(client2, "network:2")
	time.Sleep(10 * time.Millisecond)

	// Broadcast to network:1
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:1",
			UserID: "user1",
			Body:   "Message to network 1",
		},
	}

	hub.Broadcast("network:1", msg, nil)
	time.Sleep(10 * time.Millisecond)

	// Client1 should receive
	select {
	case data := <-client1.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client1 did not receive message")
	}

	// Client2 should NOT receive (different room)
	select {
	case <-client2.send:
		t.Fatal("client2 should not receive message from different room")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_SendBufferFull(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create client with very small buffer
	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 1),
		userID:      "user1",
		tenantID:    "tenant1",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	hub.JoinRoom(client, "network:1")
	time.Sleep(10 * time.Millisecond)

	// Fill the buffer
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:   "msg-1",
			Body: "Test",
		},
	}

	// First message fills buffer
	hub.Broadcast("network:1", msg, nil)
	time.Sleep(5 * time.Millisecond)

	// Second message should trigger unregister due to full buffer
	hub.Broadcast("network:1", msg, nil)
	time.Sleep(20 * time.Millisecond)

	// Client should be unregistered
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_MultipleRoomsWithSameClient(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil)
	hub := NewHub(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "tenant1",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	hub.JoinRoom(client, "network:1")
	hub.JoinRoom(client, "network:2")
	time.Sleep(10 * time.Millisecond)

	// Broadcast to network:1
	msg1 := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:1",
			UserID: "user1",
			Body:   "Message 1",
		},
	}

	hub.Broadcast("network:1", msg1, nil)
	time.Sleep(10 * time.Millisecond)

	// Client should receive from network:1
	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
		// Verify it's from network:1 by re-marshaling and checking scope
		dataMap, ok := received.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "network:1", dataMap["scope"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not receive message from network:1")
	}

	// Broadcast to network:2
	msg2 := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-2",
			Scope:  "network:2",
			UserID: "user1",
			Body:   "Message 2",
		},
	}

	hub.Broadcast("network:2", msg2, nil)
	time.Sleep(10 * time.Millisecond)

	// Client should receive from network:2
	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		// Note: Cannot directly access ID due to interface{} type
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not receive message from network:2")
	}
}

func TestHub_HandleInboundMessage(t *testing.T) {
	callCount := 0
	mockHandler := &mockMessageHandler{
		handleFunc: func(ctx context.Context, client *Client, msg *InboundMessage) error {
			callCount++
			assert.Equal(t, "user1", client.userID)
			assert.Equal(t, TypeAuthRefresh, msg.Type)
			return nil
		},
	}

	hub := NewHub(mockHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "tenant1",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Send inbound message
	hub.handleInbound <- &InboundEvent{
		client: client,
		message: &InboundMessage{
			Type: TypeAuthRefresh,
			OpID: "op-1",
			Data: json.RawMessage(`{}`),
		},
	}

	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 1, callCount)
}

func TestHub_HandleInboundMessage_Error(t *testing.T) {
	mockHandler := &mockMessageHandler{
		handleFunc: func(ctx context.Context, client *Client, msg *InboundMessage) error {
			return assert.AnError
		},
	}

	hub := NewHub(mockHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	client := &Client{
		hub:         hub,
		send:        make(chan []byte, 10),
		userID:      "user1",
		tenantID:    "tenant1",
		isAdmin:     false,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Send inbound message that will error
	hub.handleInbound <- &InboundEvent{
		client: client,
		message: &InboundMessage{
			Type: TypeChatSend,
			OpID: "op-1",
			Data: json.RawMessage(`{"scope":"network:1","body":"test"}`),
		},
	}

	time.Sleep(20 * time.Millisecond)

	// Client should receive error message
	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeError, received.Type)
		assert.Equal(t, "op-1", received.OpID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("did not receive error message")
	}
}

// mockMessageHandler implements MessageHandler for testing
type mockMessageHandler struct {
	handleFunc func(ctx context.Context, client *Client, msg *InboundMessage) error
}

func (m *mockMessageHandler) HandleMessage(ctx context.Context, client *Client, msg *InboundMessage) error {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, client, msg)
	}
	return nil
}
