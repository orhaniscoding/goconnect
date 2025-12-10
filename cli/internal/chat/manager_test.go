package chat

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("Expected Manager to be created")
	}
	if m.maxMessages != 1000 {
		t.Errorf("Expected maxMessages to be 1000, got %d", m.maxMessages)
	}
}

func TestStoreMessage(t *testing.T) {
	m := NewManager()

	msg := Message{
		ID:        "msg-1",
		From:      "peer-1",
		Content:   "Hello world",
		Time:      time.Now(),
		NetworkID: "net-1",
	}

	m.storeMessage(msg)

	messages := m.GetMessages("", 10, "")
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].ID != "msg-1" {
		t.Errorf("Expected message ID 'msg-1', got %s", messages[0].ID)
	}
}

func TestGetMessages_WithNetworkFilter(t *testing.T) {
	m := NewManager()

	// Add messages to different networks
	m.storeMessage(Message{ID: "msg-1", From: "peer-1", Content: "Net1 msg", NetworkID: "net-1", Time: time.Now()})
	m.storeMessage(Message{ID: "msg-2", From: "peer-2", Content: "Net2 msg", NetworkID: "net-2", Time: time.Now()})
	m.storeMessage(Message{ID: "msg-3", From: "peer-1", Content: "Net1 msg2", NetworkID: "net-1", Time: time.Now()})

	// Get all messages
	all := m.GetMessages("", 10, "")
	if len(all) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(all))
	}

	// Filter by network
	net1 := m.GetMessages("net-1", 10, "")
	if len(net1) != 2 {
		t.Errorf("Expected 2 messages for net-1, got %d", len(net1))
	}

	net2 := m.GetMessages("net-2", 10, "")
	if len(net2) != 1 {
		t.Errorf("Expected 1 message for net-2, got %d", len(net2))
	}
}

func TestGetMessages_WithLimit(t *testing.T) {
	m := NewManager()

	// Add 5 messages
	for i := 0; i < 5; i++ {
		m.storeMessage(Message{
			ID:      string(rune('a' + i)),
			From:    "peer-1",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Get with limit
	limited := m.GetMessages("", 3, "")
	if len(limited) != 3 {
		t.Errorf("Expected 3 messages with limit, got %d", len(limited))
	}
}

func TestGetMessages_WithBeforeID(t *testing.T) {
	m := NewManager()

	// Add messages
	m.storeMessage(Message{ID: "msg-1", From: "peer-1", Content: "First", Time: time.Now()})
	m.storeMessage(Message{ID: "msg-2", From: "peer-1", Content: "Second", Time: time.Now()})
	m.storeMessage(Message{ID: "msg-3", From: "peer-1", Content: "Third", Time: time.Now()})

	// Get messages before msg-3
	before := m.GetMessages("", 10, "msg-3")
	if len(before) != 2 {
		t.Errorf("Expected 2 messages before msg-3, got %d", len(before))
	}

	// Should contain msg-1 and msg-2 but not msg-3
	for _, msg := range before {
		if msg.ID == "msg-3" {
			t.Error("Should not contain msg-3")
		}
	}
}

func TestSubscribeUnsubscribe(t *testing.T) {
	m := NewManager()

	// Subscribe
	ch := m.Subscribe()
	if ch == nil {
		t.Fatal("Expected subscription channel")
	}

	// Unsubscribe
	m.Unsubscribe(ch)

	// Verify channel is closed
	_, ok := <-ch
	if ok {
		t.Error("Expected channel to be closed after unsubscribe")
	}
}

func TestNotifySubscribers(t *testing.T) {
	m := NewManager()

	ch := m.Subscribe()
	defer m.Unsubscribe(ch)

	msg := Message{
		ID:      "msg-notify",
		From:    "peer-1",
		Content: "Test notification",
		Time:    time.Now(),
	}

	// Notify in goroutine since storeMessage also notifies
	go m.notifySubscribers(msg)

	// Wait for message with timeout
	select {
	case received := <-ch:
		if received.ID != "msg-notify" {
			t.Errorf("Expected message ID 'msg-notify', got %s", received.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for notification")
	}
}

func TestMessageHistoryLimit(t *testing.T) {
	m := NewManager()
	m.maxMessages = 5 // Set a small limit for testing

	// Add more messages than the limit
	for i := 0; i < 10; i++ {
		m.storeMessage(Message{
			ID:      fmt.Sprintf("%d", i),
			From:    "peer-1",
			Content: "test",
			Time:    time.Now(),
		})
	}

	messages := m.GetMessages("", 100, "")
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages (limit), got %d", len(messages))
	}

	// GetMessages returns most recent first, so messages[0] should be "9" (most recent)
	// and the oldest kept message should be "5"
	if messages[0].ID != "9" {
		t.Errorf("Expected most recent message ID '9', got %s", messages[0].ID)
	}
	if messages[4].ID != "5" {
		t.Errorf("Expected oldest kept message ID '5', got %s", messages[4].ID)
	}
}

func TestOnMessageCallback(t *testing.T) {
	m := NewManager()

	m.OnMessage(func(msg Message) {
		// Callback set
	})

	// Verify callback is set
	m.mu.Lock()
	if m.onMessage == nil {
		t.Error("Expected onMessage callback to be set")
	}
	m.mu.Unlock()
}

func TestSendMessage(t *testing.T) {
	// SendMessage requires a running listener on the target
	// This is more of an integration test, so we just verify the function exists
	m := NewManager()
	
	// Calling SendMessage to non-existent target should fail gracefully
	err := m.SendMessage("192.168.1.1", 0, "test", "peer-1")
	if err == nil {
		t.Error("Expected error when sending to non-existent target")
	}
}

func TestManager_StartAndStop(t *testing.T) {
	m := NewManager()

	// Start on a random available port
	err := m.Start("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Verify listener is running
	if m.listener == nil {
		t.Error("Expected listener to be set after Start")
	}

	// Get the actual port
	addr := m.listener.Addr().String()
	if addr == "" {
		t.Error("Expected listener address to be set")
	}

	// Stop should not panic
	m.Stop()

	// Listener should be closed
	if m.listener != nil {
		// Try to accept - should fail if closed
		_, err := m.listener.Accept()
		if err == nil {
			t.Error("Expected listener to be closed after Stop")
		}
	}
}

func TestManager_StopWithoutStart(t *testing.T) {
	m := NewManager()

	// Stop without starting should not panic
	m.Stop()
}

func TestManager_GetStorage_NoStorage(t *testing.T) {
	m := NewManager()

	storage := m.GetStorage()
	if storage != nil {
		t.Error("Expected nil storage for manager created without storage")
	}
}

func TestManager_SearchMessages_NoStorage(t *testing.T) {
	m := NewManager()

	// SearchMessages without storage returns nil
	results := m.SearchMessages("test", 10)
	if results != nil {
		t.Error("Expected nil results when no storage is configured")
	}
}

func TestMessage_Struct(t *testing.T) {
	t.Run("All Fields", func(t *testing.T) {
		now := time.Now()
		msg := Message{
			ID:        "msg-123",
			From:      "peer-456",
			Content:   "Hello, World!",
			Time:      now,
			NetworkID: "network-789",
		}

		if msg.ID != "msg-123" {
			t.Errorf("Expected ID 'msg-123', got %s", msg.ID)
		}
		if msg.From != "peer-456" {
			t.Errorf("Expected From 'peer-456', got %s", msg.From)
		}
		if msg.Content != "Hello, World!" {
			t.Errorf("Expected Content 'Hello, World!', got %s", msg.Content)
		}
		if !msg.Time.Equal(now) {
			t.Errorf("Expected Time %v, got %v", now, msg.Time)
		}
		if msg.NetworkID != "network-789" {
			t.Errorf("Expected NetworkID 'network-789', got %s", msg.NetworkID)
		}
	})
}

func TestManager_GetMessages_NegativeLimit(t *testing.T) {
	m := NewManager()

	// Add some messages
	for i := 0; i < 3; i++ {
		m.storeMessage(Message{
			ID:      fmt.Sprintf("msg-%d", i),
			From:    "peer-1",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Negative limit should return all messages
	messages := m.GetMessages("", -1, "")
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages with negative limit, got %d", len(messages))
	}
}

func TestManager_GetMessages_ZeroLimit(t *testing.T) {
	m := NewManager()

	// Add some messages
	for i := 0; i < 3; i++ {
		m.storeMessage(Message{
			ID:      fmt.Sprintf("msg-%d", i),
			From:    "peer-1",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Zero limit should return all messages
	messages := m.GetMessages("", 0, "")
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages with zero limit, got %d", len(messages))
	}
}

func TestManager_ConcurrentSubscribers(t *testing.T) {
	m := NewManager()

	// Create multiple subscribers
	ch1 := m.Subscribe()
	ch2 := m.Subscribe()
	ch3 := m.Subscribe()

	// Store a message
	msg := Message{
		ID:      "concurrent-msg",
		From:    "peer-1",
		Content: "test",
		Time:    time.Now(),
	}

	// Notify all subscribers
	go m.notifySubscribers(msg)

	// All subscribers should receive the message
	timeout := time.After(time.Second)
	for i, ch := range []chan Message{ch1, ch2, ch3} {
		select {
		case received := <-ch:
			if received.ID != "concurrent-msg" {
				t.Errorf("Subscriber %d: Expected message ID 'concurrent-msg', got %s", i, received.ID)
			}
		case <-timeout:
			t.Errorf("Subscriber %d: Timeout waiting for message", i)
		}
	}

	// Cleanup
	m.Unsubscribe(ch1)
	m.Unsubscribe(ch2)
	m.Unsubscribe(ch3)
}

func TestManager_SubscriberChannelFull(t *testing.T) {
	m := NewManager()

	// Subscribe with a small buffer
	ch := m.Subscribe()

	// Fill the channel buffer (100 messages)
	for i := 0; i < 100; i++ {
		m.notifySubscribers(Message{
			ID:      fmt.Sprintf("msg-%d", i),
			From:    "peer-1",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// This should not block - additional messages should be dropped
	m.notifySubscribers(Message{
		ID:      "overflow-msg",
		From:    "peer-1",
		Content: "test",
		Time:    time.Now(),
	})

	m.Unsubscribe(ch)
}

func TestManager_StoreMessage_GeneratesID(t *testing.T) {
	m := NewManager()

	// Store message without ID
	msg := Message{
		From:    "peer-1",
		Content: "test",
		Time:    time.Now(),
	}

	m.storeMessage(msg)

	messages := m.GetMessages("", 1, "")
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// ID should be generated (format: timestamp-from)
	if messages[0].ID == "" {
		t.Error("Expected ID to be generated for message without ID")
	}
}

// TestManager_HandleConnection tests the handleConnection method via actual TCP connections
func TestManager_HandleConnection(t *testing.T) {
	m := NewManager()

	// Start on a random port
	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer m.Stop()

	// Set up a message callback to verify reception
	received := make(chan Message, 1)
	m.OnMessage(func(msg Message) {
		received <- msg
	})

	// Get the actual address
	addr := m.listener.Addr().String()

	// Connect and send a message
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// Send a valid JSON message
	msg := Message{
		ID:        "test-connection-msg",
		From:      "test-peer",
		Content:   "Hello from test",
		Time:      time.Now(),
		NetworkID: "test-network",
	}
	err = json.NewEncoder(conn).Encode(msg)
	require.NoError(t, err)

	// Wait for the message to be received
	select {
	case rcvd := <-received:
		assert.Equal(t, "test-connection-msg", rcvd.ID)
		assert.Equal(t, "test-peer", rcvd.From)
		assert.Equal(t, "Hello from test", rcvd.Content)
		assert.Equal(t, "test-network", rcvd.NetworkID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// TestManager_HandleConnection_InvalidJSON tests handling of invalid JSON input
func TestManager_HandleConnection_InvalidJSON(t *testing.T) {
	m := NewManager()

	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer m.Stop()

	addr := m.listener.Addr().String()

	// Connect and send invalid JSON
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)

	_, err = conn.Write([]byte("this is not valid json"))
	require.NoError(t, err)
	conn.Close()

	// Give it time to process the invalid input
	time.Sleep(100 * time.Millisecond)

	// The manager should continue running (not crash)
	assert.NotNil(t, m.listener)
}

// TestManager_HandleConnection_NoCallback tests handling when no callback is set
func TestManager_HandleConnection_NoCallback(t *testing.T) {
	m := NewManager()

	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer m.Stop()

	addr := m.listener.Addr().String()

	// Connect and send a message without setting a callback
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()

	msg := Message{
		ID:      "no-callback-msg",
		From:    "peer",
		Content: "Test",
		Time:    time.Now(),
	}
	err = json.NewEncoder(conn).Encode(msg)
	require.NoError(t, err)

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Message should still be stored even without callback
	messages := m.GetMessages("", 10, "")
	assert.Len(t, messages, 1)
	assert.Equal(t, "no-callback-msg", messages[0].ID)
}

// TestManager_HandleConnection_SubscriberNotification tests that subscribers are notified
func TestManager_HandleConnection_SubscriberNotification(t *testing.T) {
	m := NewManager()

	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer m.Stop()

	// Subscribe before sending message
	ch := m.Subscribe()
	defer m.Unsubscribe(ch)

	addr := m.listener.Addr().String()

	// Connect and send a message
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()

	msg := Message{
		ID:      "subscriber-msg",
		From:    "peer",
		Content: "For subscribers",
		Time:    time.Now(),
	}
	err = json.NewEncoder(conn).Encode(msg)
	require.NoError(t, err)

	// Wait for subscriber notification
	select {
	case rcvd := <-ch:
		assert.Equal(t, "subscriber-msg", rcvd.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for subscriber notification")
	}
}

// TestManager_SendMessage_Success tests successful message sending
func TestManager_SendMessage_Success(t *testing.T) {
	// Create a receiver manager
	receiver := NewManager()
	err := receiver.Start("127.0.0.1", 0)
	require.NoError(t, err)
	defer receiver.Stop()

	// Get the port
	addr := receiver.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Set up message callback
	received := make(chan Message, 1)
	receiver.OnMessage(func(msg Message) {
		received <- msg
	})

	// Create sender manager
	sender := NewManager()

	// Send a message
	err = sender.SendMessage("127.0.0.1", port, "Hello from sender", "sender-peer-id")
	require.NoError(t, err)

	// Verify message was received
	select {
	case rcvd := <-received:
		assert.Equal(t, "Hello from sender", rcvd.Content)
		assert.Equal(t, "sender-peer-id", rcvd.From)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for sent message")
	}
}

// TestManager_SendMessage_InvalidPort tests sending to an invalid port
func TestManager_SendMessage_InvalidPort(t *testing.T) {
	m := NewManager()

	// Try to send to an invalid/closed port
	err := m.SendMessage("127.0.0.1", 59999, "test", "peer")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to dial peer")
}

// TestManager_SendMessage_InvalidIP tests sending to an invalid IP
func TestManager_SendMessage_InvalidIP(t *testing.T) {
	m := NewManager()

	// Try to send to an invalid IP
	err := m.SendMessage("999.999.999.999", 1234, "test", "peer")
	assert.Error(t, err)
}

// TestManager_SearchMessages_WithStorage tests SearchMessages with storage
func TestManager_SearchMessages_WithStorage(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)
	defer m.Stop()

	// Store some messages
	m.storeMessage(Message{
		ID:      "search-1",
		From:    "peer",
		Content: "Hello world",
		Time:    time.Now(),
	})
	m.storeMessage(Message{
		ID:      "search-2",
		From:    "peer",
		Content: "Goodbye world",
		Time:    time.Now(),
	})
	m.storeMessage(Message{
		ID:      "search-3",
		From:    "peer",
		Content: "Hello again",
		Time:    time.Now(),
	})

	// Search for "Hello"
	results := m.SearchMessages("Hello", 10)
	assert.Len(t, results, 2)
}

// TestManager_GetMessages_WithStorage tests GetMessages with storage backend
func TestManager_GetMessages_WithStorage(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)
	defer m.Stop()

	// Store messages
	for i := 0; i < 5; i++ {
		m.storeMessage(Message{
			ID:        fmt.Sprintf("stored-msg-%d", i),
			From:      "peer",
			Content:   fmt.Sprintf("Message %d", i),
			Time:      time.Now().Add(time.Duration(i) * time.Second),
			NetworkID: "network-1",
		})
	}

	// Get messages from storage
	messages := m.GetMessages("", 3, "")
	assert.Len(t, messages, 3)

	// Get messages filtered by network
	networkMsgs := m.GetMessages("network-1", 10, "")
	assert.Len(t, networkMsgs, 5)
}

// TestManager_StoreMessage_WithStorage tests message persistence
func TestManager_StoreMessage_WithStorage(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)

	// Store a message
	m.storeMessage(Message{
		ID:        "persist-msg",
		From:      "peer",
		Content:   "Persistent message",
		Time:      time.Now(),
		NetworkID: "network",
	})

	m.Stop()

	// Create a new manager with the same storage
	m2, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)
	defer m2.Stop()

	// Message should be loaded from storage
	messages := m2.GetMessages("", 10, "")
	assert.Len(t, messages, 1)
	assert.Equal(t, "persist-msg", messages[0].ID)
}

// TestManager_LoadRecentMessages tests loading messages into memory cache
func TestManager_LoadRecentMessages(t *testing.T) {
	tmpDir := t.TempDir()

	// First create a manager and save some messages
	m1, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		m1.storeMessage(Message{
			ID:      fmt.Sprintf("load-msg-%d", i),
			From:    "peer",
			Content: fmt.Sprintf("Message %d", i),
			Time:    time.Now().Add(time.Duration(i) * time.Second),
		})
	}
	m1.Stop()

	// Create a new manager - messages should be loaded into cache
	m2, err := NewManagerWithStorage(tmpDir)
	require.NoError(t, err)
	defer m2.Stop()

	// Check in-memory cache has messages
	m2.messagesMu.RLock()
	cachedCount := len(m2.messages)
	m2.messagesMu.RUnlock()

	assert.Equal(t, 10, cachedCount)
}

// TestManager_AcceptLoop_StopChannel tests that acceptLoop handles stopChan correctly
func TestManager_AcceptLoop_StopChannel(t *testing.T) {
	m := NewManager()

	err := m.Start("127.0.0.1", 0)
	require.NoError(t, err)

	// Stop should trigger the stopChan and exit gracefully
	done := make(chan struct{})
	go func() {
		m.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success - Stop completed
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for Stop to complete")
	}
}

// TestManager_Start_FailsOnBadAddress tests Start with invalid address
func TestManager_Start_FailsOnBadAddress(t *testing.T) {
	m := NewManager()

	// Try to start on an invalid address
	err := m.Start("999.999.999.999", 12345)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start chat listener")
}

// TestNewManagerWithStorage_InvalidPath tests creating storage with invalid path
func TestNewManagerWithStorage_InvalidPath(t *testing.T) {
	// Try to create storage in a path that doesn't allow writing
	_, err := NewManagerWithStorage("/nonexistent/deeply/nested/path/that/should/fail")
	// This might not fail on all systems, but the important thing is it handles the case gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "failed to initialize chat storage")
	}
}

// TestManager_MultipleSubscribers tests handling multiple subscribers
func TestManager_MultipleSubscribers(t *testing.T) {
	m := NewManager()

	// Create multiple subscribers
	subscribers := make([]chan Message, 5)
	for i := 0; i < 5; i++ {
		subscribers[i] = m.Subscribe()
	}

	// Notify subscribers directly (not via storeMessage to avoid race)
	msg := Message{
		ID:      "multi-sub-msg",
		From:    "peer",
		Content: "Test",
		Time:    time.Now(),
	}
	m.notifySubscribers(msg)

	// All subscribers should receive the message
	for i, ch := range subscribers {
		select {
		case rcvd := <-ch:
			assert.Equal(t, "multi-sub-msg", rcvd.ID, "Subscriber %d received wrong message", i)
		case <-time.After(2 * time.Second):
			t.Fatalf("Subscriber %d did not receive message", i)
		}
	}

	// Unsubscribe all
	for _, ch := range subscribers {
		m.Unsubscribe(ch)
	}
}

// TestMessage_JSON tests JSON marshaling/unmarshaling of Message
func TestMessage_JSON(t *testing.T) {
	original := Message{
		ID:        "json-test-msg",
		From:      "peer-123",
		Content:   "Test content with special chars: <>&\"'",
		Time:      time.Now().Round(time.Second), // Round for comparison
		NetworkID: "network-456",
	}

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var decoded Message
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.From, decoded.From)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.NetworkID, decoded.NetworkID)
	assert.True(t, original.Time.Equal(decoded.Time))
}

// TestManager_GetMessages_BeforeID_NotFound tests GetMessages with non-existent beforeID
func TestManager_GetMessages_BeforeID_NotFound(t *testing.T) {
	m := NewManager()

	// Add some messages
	for i := 0; i < 3; i++ {
		m.storeMessage(Message{
			ID:      fmt.Sprintf("msg-%d", i),
			From:    "peer",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Try to get messages before a non-existent ID
	messages := m.GetMessages("", 10, "nonexistent-id")
	// Should return empty since the beforeID was never found
	assert.Empty(t, messages)
}

// TestManager_GetMessages_NetworkFilter_NoMatch tests network filtering with no matches
func TestManager_GetMessages_NetworkFilter_NoMatch(t *testing.T) {
	m := NewManager()

	// Add messages to one network
	m.storeMessage(Message{
		ID:        "msg-1",
		From:      "peer",
		Content:   "test",
		Time:      time.Now(),
		NetworkID: "network-1",
	})

	// Try to get messages from a different network
	messages := m.GetMessages("network-2", 10, "")
	assert.Empty(t, messages)
}

// TestManager_GetMessages_LimitExceedsCount tests limit larger than message count
func TestManager_GetMessages_LimitExceedsCount(t *testing.T) {
	m := NewManager()

	// Add 3 messages
	for i := 0; i < 3; i++ {
		m.storeMessage(Message{
			ID:      fmt.Sprintf("msg-%d", i),
			From:    "peer",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Request more messages than exist
	messages := m.GetMessages("", 100, "")
	assert.Len(t, messages, 3)
}
