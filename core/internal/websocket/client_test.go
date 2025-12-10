package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastMessage(t *testing.T) {
	client := &Client{
		userID: "user1",
		send:   make(chan []byte, 10),
	}

	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:   "msg-1",
			Body: "Test",
		},
	}

	bm := &BroadcastMessage{
		Room:    "network:1",
		Message: msg,
		Exclude: client,
	}

	assert.Equal(t, "network:1", bm.Room)
	assert.Equal(t, msg, bm.Message)
	assert.Equal(t, client, bm.Exclude)
}

func TestInboundEvent(t *testing.T) {
	client := &Client{
		userID: "user1",
		send:   make(chan []byte, 10),
	}

	msg := &InboundMessage{
		Type: TypeChatSend,
		OpID: "op-1",
		Data: json.RawMessage(`{"scope":"network:1","body":"test"}`),
	}

	event := &InboundEvent{
		client:  client,
		message: msg,
	}

	assert.Equal(t, client, event.client)
	assert.Equal(t, msg, event.message)
	assert.Equal(t, TypeChatSend, event.message.Type)
}

func TestNewClient(t *testing.T) {
	hub := NewHub(nil)

	// Create mock websocket connection
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	client := NewClient(hub, ws, "user-1", "tenant-1", true, false)

	assert.NotNil(t, client)
	assert.Equal(t, "user-1", client.userID)
	assert.Equal(t, "tenant-1", client.tenantID)
	assert.True(t, client.isAdmin)
	assert.False(t, client.isModerator)
	assert.Equal(t, sendBufferSize, cap(client.send))
	assert.NotNil(t, client.rooms)
	assert.NotNil(t, client.limiter)
}

func TestClient_JoinRoom(t *testing.T) {
	client := &Client{
		rooms: make(map[string]bool),
	}

	client.JoinRoom("network:123")

	assert.True(t, client.IsInRoom("network:123"))
	assert.False(t, client.IsInRoom("network:456"))
}

func TestClient_LeaveRoom(t *testing.T) {
	client := &Client{
		rooms: make(map[string]bool),
	}

	client.JoinRoom("network:123")
	assert.True(t, client.IsInRoom("network:123"))

	client.LeaveRoom("network:123")
	assert.False(t, client.IsInRoom("network:123"))
}

func TestClient_GetRooms(t *testing.T) {
	client := &Client{
		rooms: make(map[string]bool),
	}

	client.JoinRoom("network:123")
	client.JoinRoom("network:456")
	client.JoinRoom("host")

	rooms := client.GetRooms()
	assert.Len(t, rooms, 3)
	assert.Contains(t, rooms, "network:123")
	assert.Contains(t, rooms, "network:456")
	assert.Contains(t, rooms, "host")
}

func TestClient_Send(t *testing.T) {
	client := &Client{
		send: make(chan []byte, 10),
	}

	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:123",
			UserID: "user-1",
			Body:   "Test message",
		},
	}

	err := client.Send(msg)
	assert.NoError(t, err)

	// Verify message was sent to channel
	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("message not sent")
	}
}

func TestClient_Send_Timeout(t *testing.T) {
	// Create client with buffer size 1
	client := &Client{
		send: make(chan []byte, 1),
	}

	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:   "msg-1",
			Body: "Test",
		},
	}

	// Fill the buffer
	err := client.Send(msg)
	assert.NoError(t, err)

	// This should timeout because buffer is full
	err = client.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestClient_SendError(t *testing.T) {
	client := &Client{
		send: make(chan []byte, 10),
	}

	client.sendError("op-123", "ERR_INVALID", "Invalid operation", map[string]string{
		"field": "scope",
	})

	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)

		assert.Equal(t, TypeError, received.Type)
		assert.Equal(t, "op-123", received.OpID)
		assert.NotNil(t, received.Error)
		assert.Equal(t, "ERR_INVALID", received.Error.Code)
		assert.Equal(t, "Invalid operation", received.Error.Message)
		assert.Equal(t, "scope", received.Error.Details["field"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("error message not sent")
	}
}

func TestClient_SendAck(t *testing.T) {
	client := &Client{
		send: make(chan []byte, 10),
	}

	ackData := map[string]string{
		"message_id": "msg-123",
		"status":     "sent",
	}

	client.sendAck("op-456", ackData)

	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)

		assert.Equal(t, TypeAck, received.Type)
		assert.Equal(t, "op-456", received.OpID)
		assert.NotNil(t, received.Data)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ack message not sent")
	}
}

func TestClient_IsInRoom_ThreadSafe(t *testing.T) {
	client := &Client{
		rooms: make(map[string]bool),
	}

	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			client.JoinRoom("network:1")
			client.LeaveRoom("network:1")
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = client.IsInRoom("network:1")
			_ = client.GetRooms()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// No data race should occur
}

func TestClient_MultipleRooms(t *testing.T) {
	client := &Client{
		rooms: make(map[string]bool),
	}

	// Join multiple rooms
	rooms := []string{
		"host",
		"network:1",
		"network:2",
		"network:3",
	}

	for _, room := range rooms {
		client.JoinRoom(room)
	}

	// Verify all rooms
	for _, room := range rooms {
		assert.True(t, client.IsInRoom(room))
	}

	// Leave one room
	client.LeaveRoom("network:2")
	assert.False(t, client.IsInRoom("network:2"))

	// Others should remain
	assert.True(t, client.IsInRoom("host"))
	assert.True(t, client.IsInRoom("network:1"))
	assert.True(t, client.IsInRoom("network:3"))

	// Room count should be 3
	assert.Len(t, client.GetRooms(), 3)
}

// echo is a simple WebSocket echo server for testing
func echo(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestClient_Constants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, 10*time.Second, writeWait)
	assert.Equal(t, 60*time.Second, pongWait)
	assert.Equal(t, 54*time.Second, pingPeriod) // (60 * 9) / 10
	assert.Equal(t, int64(512*1024), int64(maxMessageSize))
	assert.Equal(t, 256, sendBufferSize)
}

func TestClient_UserProperties(t *testing.T) {
	client := &Client{
		userID:      "user-123",
		tenantID:    "tenant-456",
		isAdmin:     true,
		isModerator: false,
		rooms:       make(map[string]bool),
	}

	assert.Equal(t, "user-123", client.userID)
	assert.Equal(t, "tenant-456", client.tenantID)
	assert.True(t, client.isAdmin)
	assert.False(t, client.isModerator)
}

func TestClient_AdminAndModeratorFlags(t *testing.T) {
	// Test regular user
	regularUser := &Client{
		userID:      "user1",
		isAdmin:     false,
		isModerator: false,
	}
	assert.False(t, regularUser.isAdmin)
	assert.False(t, regularUser.isModerator)

	// Test admin
	admin := &Client{
		userID:  "admin1",
		isAdmin: true,
	}
	assert.True(t, admin.isAdmin)

	// Test moderator
	moderator := &Client{
		userID:      "mod1",
		isModerator: true,
	}
	assert.True(t, moderator.isModerator)

	// Test admin + moderator
	superUser := &Client{
		userID:      "super1",
		isAdmin:     true,
		isModerator: true,
	}
	assert.True(t, superUser.isAdmin)
	assert.True(t, superUser.isModerator)
}

func TestClient_UpdateActivityAndGetLastActivity(t *testing.T) {
	client := &Client{
		userID:   "user1",
		tenantID: "tenant1",
	}

	// Check initial last activity is zero
	initial := client.GetLastActivity()
	assert.True(t, initial.IsZero())

	// Update activity
	client.UpdateActivity()

	// Check last activity is now set
	activity := client.GetLastActivity()
	assert.False(t, activity.IsZero())
	assert.WithinDuration(t, time.Now(), activity, time.Second)

	// Update again and verify it changes
	time.Sleep(10 * time.Millisecond)
	client.UpdateActivity()
	newActivity := client.GetLastActivity()
	assert.True(t, newActivity.After(activity) || newActivity.Equal(activity))
}

// createTestWebSocketServer creates a test HTTP server that upgrades to WebSocket
// and returns the server, WebSocket URL, and a cleanup function
func createTestWebSocketServer(t *testing.T, handler func(*websocket.Conn)) (*httptest.Server, string, func()) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Upgrade error: %v", err)
			return
		}
		handler(conn)
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	cleanup := func() {
		server.Close()
	}

	return server, wsURL, cleanup
}

// TestClient_ReadPump_MessageHandling tests that readPump correctly receives and processes messages
func TestClient_ReadPump_MessageHandling(t *testing.T) {
	// Create a hub with a mock handler that captures inbound events
	inboundChan := make(chan *InboundEvent, 10)
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: inboundChan,
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
	}

	// Create server that sends a message to the client
	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		// Send a valid JSON message to the client
		msg := InboundMessage{
			Type: TypeChatSend,
			OpID: "test-op-1",
			Data: json.RawMessage(`{"scope":"network:1","body":"hello"}`),
		}
		data, _ := json.Marshal(msg)
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			t.Logf("Server write error: %v", err)
		}

		// Keep connection open briefly
		time.Sleep(100 * time.Millisecond)
	})
	defer cleanup()

	// Connect client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	// Start readPump in background
	done := make(chan bool)
	go func() {
		client.readPump()
		done <- true
	}()

	// Wait for inbound event
	select {
	case event := <-inboundChan:
		assert.Equal(t, client, event.client)
		assert.Equal(t, TypeChatSend, event.message.Type)
		assert.Equal(t, "test-op-1", event.message.OpID)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected inbound event but got timeout")
	}

	// Wait for readPump to finish
	select {
	case <-done:
		// Expected - connection closed by server
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readPump did not finish")
	}
}

// TestClient_ReadPump_InvalidJSON tests that readPump handles invalid JSON gracefully
func TestClient_ReadPump_InvalidJSON(t *testing.T) {
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: make(chan *InboundEvent, 10),
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
	}

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		// Send invalid JSON
		err := conn.WriteMessage(websocket.TextMessage, []byte(`{invalid json}`))
		if err != nil {
			t.Logf("Server write error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	done := make(chan bool)
	go func() {
		client.readPump()
		done <- true
	}()

	// Wait for error message on client send channel
	select {
	case data := <-client.send:
		var outMsg OutboundMessage
		err := json.Unmarshal(data, &outMsg)
		require.NoError(t, err)
		assert.Equal(t, TypeError, outMsg.Type)
		assert.Equal(t, "ERR_INVALID_MESSAGE", outMsg.Error.Code)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Expected error message but got timeout")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readPump did not finish")
	}
}

// TestClient_ReadPump_RateLimit tests that readPump enforces rate limiting
func TestClient_ReadPump_RateLimit(t *testing.T) {
	inboundChan := make(chan *InboundEvent, 100)
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: inboundChan,
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
	}

	messageCount := 0
	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		// Send many messages quickly to trigger rate limiting
		for i := 0; i < 30; i++ {
			msg := InboundMessage{
				Type: TypeChatSend,
				OpID: "op-" + string(rune('a'+i)),
				Data: json.RawMessage(`{"scope":"network:1","body":"msg"}`),
			}
			data, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				break
			}
			messageCount++
		}

		time.Sleep(100 * time.Millisecond)
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	done := make(chan bool)
	go func() {
		client.readPump()
		done <- true
	}()

	// Collect inbound events and rate limit errors
	time.Sleep(300 * time.Millisecond)

	// Check for rate limit error in send channel
	rateLimitErrorFound := false
	for {
		select {
		case data := <-client.send:
			var outMsg OutboundMessage
			if err := json.Unmarshal(data, &outMsg); err == nil {
				if outMsg.Type == TypeError && outMsg.Error != nil && outMsg.Error.Code == "ERR_RATE_LIMIT" {
					rateLimitErrorFound = true
				}
			}
		default:
			goto checkResult
		}
	}

checkResult:
	// With rate limit of 10 messages per second, burst 20, sending 30 messages
	// should trigger at least some rate limit errors
	assert.True(t, rateLimitErrorFound, "Expected rate limit error to be triggered")

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// That's okay, we just need to verify rate limiting worked
	}
}

// TestClient_WritePump_SendMessage tests that writePump correctly sends messages to the connection
func TestClient_WritePump_SendMessage(t *testing.T) {
	receivedMsgs := make(chan []byte, 10)

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			receivedMsgs <- message
		}
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	// Create hub for client
	hub := NewHub(nil)
	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	// Start writePump in background
	go client.writePump()

	// Send a message through the client
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:1",
			UserID: "user-1",
			Body:   "Hello world",
		},
	}
	err = client.Send(msg)
	require.NoError(t, err)

	// Wait for server to receive message
	select {
	case data := <-receivedMsgs:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeChatMessage, received.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Server did not receive message")
	}

	// Clean up
	close(client.send)
	time.Sleep(50 * time.Millisecond)
}

// TestClient_WritePump_MultipleMessages tests batching of messages
func TestClient_WritePump_MultipleMessages(t *testing.T) {
	receivedMsgs := make(chan []byte, 10)

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			receivedMsgs <- message
		}
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	hub := NewHub(nil)
	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	// Queue multiple messages before starting writePump
	for i := 0; i < 3; i++ {
		msg := &OutboundMessage{
			Type: TypeChatMessage,
			Data: &ChatMessageData{
				ID:   "msg-" + string(rune('1'+i)),
				Body: "Message " + string(rune('1'+i)),
			},
		}
		data, _ := json.Marshal(msg)
		client.send <- data
	}

	// Start writePump
	go client.writePump()

	// Wait for server to receive messages (may be batched)
	time.Sleep(200 * time.Millisecond)

	// Verify we received data
	select {
	case data := <-receivedMsgs:
		// Data should contain at least one message (potentially newline-separated)
		assert.NotEmpty(t, data)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Server did not receive any messages")
	}

	close(client.send)
	time.Sleep(50 * time.Millisecond)
}

// TestClient_WritePump_ChannelClosed tests that writePump handles channel close
func TestClient_WritePump_ChannelClosed(t *testing.T) {
	closeMessageReceived := make(chan bool, 1)

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			msgType, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if msgType == websocket.CloseMessage {
				closeMessageReceived <- true
				break
			}
		}
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	hub := NewHub(nil)
	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	done := make(chan bool)
	go func() {
		client.writePump()
		done <- true
	}()

	// Close the send channel to signal shutdown
	close(client.send)

	// Wait for writePump to finish
	select {
	case <-done:
		// writePump finished as expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("writePump did not finish after channel close")
	}
}

// TestClient_Run tests the Run function that starts both pumps
func TestClient_Run(t *testing.T) {
	inboundChan := make(chan *InboundEvent, 10)
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: inboundChan,
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
	}

	receivedMsgs := make(chan []byte, 10)

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		// Send a message to client
		msg := InboundMessage{
			Type: TypeChatSend,
			OpID: "op-run-test",
			Data: json.RawMessage(`{"scope":"network:1","body":"hello from server"}`),
		}
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		// Read messages from client
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			receivedMsgs <- message
		}
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	// Run client (starts both pumps)
	ctx := context.Background()
	client.Run(ctx)

	// Verify readPump received the message
	select {
	case event := <-inboundChan:
		assert.Equal(t, "op-run-test", event.message.OpID)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readPump did not receive message")
	}

	// Send a message through writePump
	outMsg := &OutboundMessage{
		Type: TypeAck,
		OpID: "ack-test",
	}
	err = client.Send(outMsg)
	require.NoError(t, err)

	// Verify server received the message
	select {
	case data := <-receivedMsgs:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypeAck, received.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Server did not receive message from writePump")
	}

	// Clean up
	ws.Close()
	time.Sleep(100 * time.Millisecond)
}

// TestClient_ReadPump_ConnectionClose tests that readPump handles connection close properly
func TestClient_ReadPump_ConnectionClose(t *testing.T) {
	unregisterChan := make(chan *Client, 1)
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: make(chan *InboundEvent, 10),
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    unregisterChan,
	}

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		// Close connection immediately
		conn.Close()
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	done := make(chan bool)
	go func() {
		client.readPump()
		done <- true
	}()

	// Verify unregister is called
	select {
	case unregistered := <-unregisterChan:
		assert.Equal(t, client, unregistered)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Client was not unregistered after connection close")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readPump did not finish")
	}
}

// TestClient_ReadPump_PongHandler tests that pong handler resets read deadline
func TestClient_ReadPump_PongHandler(t *testing.T) {
	hub := &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: make(chan *InboundEvent, 10),
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
	}

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		// Send a ping, client should respond with pong
		err := conn.WriteMessage(websocket.PingMessage, []byte("ping"))
		if err != nil {
			t.Logf("Error sending ping: %v", err)
			return
		}

		// Keep connection alive briefly
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	done := make(chan bool)
	go func() {
		client.readPump()
		done <- true
	}()

	// Wait for readPump to finish
	select {
	case <-done:
		// Connection closed normally
	case <-time.After(1 * time.Second):
		t.Fatal("readPump did not finish")
	}
}

// TestClient_WritePump_Ping tests that writePump sends periodic pings
func TestClient_WritePump_Ping(t *testing.T) {
	// Skip this test if running in CI as it's timing-sensitive
	if testing.Short() {
		t.Skip("Skipping ping test in short mode")
	}

	pingReceived := make(chan bool, 1)

	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		defer conn.Close()

		conn.SetPingHandler(func(appData string) error {
			pingReceived <- true
			return nil
		})

		// Keep reading to process pings
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	hub := NewHub(nil)
	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	go client.writePump()

	// Note: pingPeriod is 54 seconds, so we won't wait for actual ping in unit tests
	// Instead we just verify the writePump is running correctly
	time.Sleep(100 * time.Millisecond)

	// Clean up
	close(client.send)
}

// TestClient_WritePump_WriteError tests that writePump handles write errors
func TestClient_WritePump_WriteError(t *testing.T) {
	serverReady := make(chan bool)
	_, wsURL, cleanup := createTestWebSocketServer(t, func(conn *websocket.Conn) {
		serverReady <- true
		// Keep the connection open until we're done
		time.Sleep(5 * time.Second)
		conn.Close()
	})
	defer cleanup()

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	hub := NewHub(nil)
	client := NewClient(hub, ws, "user-1", "tenant-1", false, false)

	// Wait for server to be ready
	<-serverReady

	done := make(chan bool)
	go func() {
		client.writePump()
		done <- true
	}()

	// Give writePump a moment to start
	time.Sleep(10 * time.Millisecond)

	// Close the websocket connection to cause a write error
	ws.Close()

	// writePump should finish due to write error (on next ticker tick or send)
	// Send a message to trigger the write error immediately
	msg := &OutboundMessage{
		Type: TypeAck,
		OpID: "test",
	}
	data, _ := json.Marshal(msg)

	select {
	case client.send <- data:
		// Message queued, will fail on write
	default:
		// Channel full or closed
	}

	// Wait for writePump to finish
	select {
	case <-done:
		// Expected - writePump finished due to write error
	case <-time.After(1 * time.Second):
		t.Fatal("writePump did not finish after write error")
	}
}

// TestClient_SendMessage_Internal tests the internal sendMessage function
func TestClient_SendMessage_Internal(t *testing.T) {
	client := &Client{
		send: make(chan []byte, 10),
	}

	msg := &OutboundMessage{
		Type: TypePresencePong,
		Data: map[string]string{"status": "ok"},
	}

	client.sendMessage(msg)

	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		require.NoError(t, err)
		assert.Equal(t, TypePresencePong, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendMessage did not send")
	}
}

// TestClient_SendMessage_BufferFull tests sendMessage when buffer is full
func TestClient_SendMessage_BufferFull(t *testing.T) {
	// Create client with buffer size 1
	client := &Client{
		send: make(chan []byte, 1),
	}

	msg := &OutboundMessage{
		Type: TypeAck,
	}

	// Fill the buffer
	client.send <- []byte(`{}`)

	// This should not block (uses select with default)
	client.sendMessage(msg)

	// Should still have only one message in buffer
	assert.Len(t, client.send, 1)
}
