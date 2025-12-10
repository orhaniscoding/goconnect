package chat

import (
	"fmt"
	"testing"
	"time"
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
