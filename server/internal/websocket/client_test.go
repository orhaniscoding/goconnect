package websocket

import (
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
