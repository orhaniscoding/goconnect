package chat

import (
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	tmpDir := t.TempDir()
	
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	if storage.db == nil {
		t.Error("Expected database connection")
	}
}

func TestStorage_SaveMessage(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	msg := Message{
		ID:        "test-msg-1",
		From:      "peer-123",
		Content:   "Hello, world!",
		Time:      time.Now(),
		NetworkID: "network-1",
	}
	
	if err := storage.SaveMessage(msg); err != nil {
		t.Fatalf("SaveMessage failed: %v", err)
	}
	
	// Verify message was saved
	messages, err := storage.GetMessages("", 10, "")
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	
	if messages[0].ID != "test-msg-1" {
		t.Errorf("Expected message ID 'test-msg-1', got %s", messages[0].ID)
	}
	if messages[0].Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got %s", messages[0].Content)
	}
}

func TestStorage_SaveMessage_GeneratesID(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	msg := Message{
		From:    "peer-123",
		Content: "No ID provided",
		Time:    time.Now(),
	}
	
	if err := storage.SaveMessage(msg); err != nil {
		t.Fatalf("SaveMessage failed: %v", err)
	}
	
	messages, _ := storage.GetMessages("", 10, "")
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	
	// ID should have been generated
	if messages[0].ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestStorage_GetMessages_FilterByNetwork(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	// Save messages to different networks
	storage.SaveMessage(Message{ID: "msg-1", From: "peer-1", Content: "Net 1", Time: time.Now(), NetworkID: "network-1"})
	storage.SaveMessage(Message{ID: "msg-2", From: "peer-2", Content: "Net 2", Time: time.Now().Add(time.Second), NetworkID: "network-2"})
	storage.SaveMessage(Message{ID: "msg-3", From: "peer-1", Content: "Net 1 again", Time: time.Now().Add(2 * time.Second), NetworkID: "network-1"})
	
	// Get all messages
	all, _ := storage.GetMessages("", 10, "")
	if len(all) != 3 {
		t.Errorf("Expected 3 total messages, got %d", len(all))
	}
	
	// Filter by network-1
	net1, _ := storage.GetMessages("network-1", 10, "")
	if len(net1) != 2 {
		t.Errorf("Expected 2 messages for network-1, got %d", len(net1))
	}
	
	// Filter by network-2
	net2, _ := storage.GetMessages("network-2", 10, "")
	if len(net2) != 1 {
		t.Errorf("Expected 1 message for network-2, got %d", len(net2))
	}
}

func TestStorage_GetMessages_Pagination(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	// Save 5 messages
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		storage.SaveMessage(Message{
			ID:      "msg-" + string(rune('a'+i)),
			From:    "peer-1",
			Content: "Message " + string(rune('A'+i)),
			Time:    baseTime.Add(time.Duration(i) * time.Second),
		})
	}
	
	// Get first 2 (most recent)
	first2, _ := storage.GetMessages("", 2, "")
	if len(first2) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(first2))
	}
	
	// Most recent should be first (msg-e)
	if first2[0].ID != "msg-e" {
		t.Errorf("Expected msg-e first, got %s", first2[0].ID)
	}
}

func TestStorage_GetMessagesByPeer(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	storage.SaveMessage(Message{ID: "msg-1", From: "alice", Content: "From Alice", Time: time.Now()})
	storage.SaveMessage(Message{ID: "msg-2", From: "bob", Content: "From Bob", Time: time.Now()})
	storage.SaveMessage(Message{ID: "msg-3", From: "alice", Content: "Alice again", Time: time.Now()})
	
	aliceMsgs, err := storage.GetMessagesByPeer("alice", 10)
	if err != nil {
		t.Fatalf("GetMessagesByPeer failed: %v", err)
	}
	
	if len(aliceMsgs) != 2 {
		t.Errorf("Expected 2 messages from alice, got %d", len(aliceMsgs))
	}
}

func TestStorage_DeleteMessage(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	storage.SaveMessage(Message{ID: "msg-to-delete", From: "peer", Content: "Delete me", Time: time.Now()})
	
	// Verify it exists
	count, _ := storage.GetMessageCount()
	if count != 1 {
		t.Errorf("Expected 1 message, got %d", count)
	}
	
	// Delete it
	if err := storage.DeleteMessage("msg-to-delete"); err != nil {
		t.Fatalf("DeleteMessage failed: %v", err)
	}
	
	// Verify it's gone
	count, _ = storage.GetMessageCount()
	if count != 0 {
		t.Errorf("Expected 0 messages after delete, got %d", count)
	}
}

func TestStorage_DeleteOldMessages(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	// Save old and new messages
	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now()
	
	storage.SaveMessage(Message{ID: "old-msg", From: "peer", Content: "Old", Time: oldTime})
	storage.SaveMessage(Message{ID: "new-msg", From: "peer", Content: "New", Time: newTime})
	
	// Delete messages older than 24 hours
	deleted, err := storage.DeleteOldMessages(24 * time.Hour)
	if err != nil {
		t.Fatalf("DeleteOldMessages failed: %v", err)
	}
	
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}
	
	// Verify only new message remains
	messages, _ := storage.GetMessages("", 10, "")
	if len(messages) != 1 {
		t.Errorf("Expected 1 message remaining, got %d", len(messages))
	}
	if messages[0].ID != "new-msg" {
		t.Errorf("Expected new-msg to remain, got %s", messages[0].ID)
	}
}

func TestStorage_GetMessageCount(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	count, _ := storage.GetMessageCount()
	if count != 0 {
		t.Errorf("Expected 0 messages initially, got %d", count)
	}
	
	storage.SaveMessage(Message{ID: "1", From: "p", Content: "a", Time: time.Now()})
	storage.SaveMessage(Message{ID: "2", From: "p", Content: "b", Time: time.Now()})
	storage.SaveMessage(Message{ID: "3", From: "p", Content: "c", Time: time.Now()})
	
	count, _ = storage.GetMessageCount()
	if count != 3 {
		t.Errorf("Expected 3 messages, got %d", count)
	}
}

func TestStorage_SearchMessages(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer storage.Close()
	
	storage.SaveMessage(Message{ID: "1", From: "p", Content: "Hello world", Time: time.Now()})
	storage.SaveMessage(Message{ID: "2", From: "p", Content: "Goodbye world", Time: time.Now()})
	storage.SaveMessage(Message{ID: "3", From: "p", Content: "Hello again", Time: time.Now()})
	
	// Search for "Hello"
	results, err := storage.SearchMessages("Hello", 10)
	if err != nil {
		t.Fatalf("SearchMessages failed: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Hello', got %d", len(results))
	}
	
	// Search for "world"
	results, _ = storage.SearchMessages("world", 10)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'world', got %d", len(results))
	}
	
	// Search for non-existent
	results, _ = storage.SearchMessages("nonexistent", 10)
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestNewManagerWithStorage(t *testing.T) {
	tmpDir := t.TempDir()
	
	manager, err := NewManagerWithStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewManagerWithStorage failed: %v", err)
	}
	
	if manager.storage == nil {
		t.Error("Expected storage to be initialized")
	}
	
	// Must stop manager to close storage properly before TempDir cleanup
	manager.Stop()
}

func TestManager_GetStorage(t *testing.T) {
	tmpDir := t.TempDir()
	
	manager, err := NewManagerWithStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewManagerWithStorage failed: %v", err)
	}
	
	storage := manager.GetStorage()
	if storage == nil {
		t.Error("Expected GetStorage to return storage")
	}
	
	// Must stop manager to close storage properly before TempDir cleanup
	manager.Stop()
}
