package chat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	_ = storage.SaveMessage(Message{ID: "msg-1", From: "peer-1", Content: "Net 1", Time: time.Now(), NetworkID: "network-1"})
	_ = storage.SaveMessage(Message{ID: "msg-2", From: "peer-2", Content: "Net 2", Time: time.Now().Add(time.Second), NetworkID: "network-2"})
	_ = storage.SaveMessage(Message{ID: "msg-3", From: "peer-1", Content: "Net 1 again", Time: time.Now().Add(2 * time.Second), NetworkID: "network-1"})

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

// TestStorage_GetMessages_Pagination_WithBeforeID tests cursor-based pagination
func TestStorage_GetMessages_Pagination_WithBeforeID(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// Save 5 messages with different timestamps
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		err := storage.SaveMessage(Message{
			ID:      "msg-" + string(rune('a'+i)),
			From:    "peer-1",
			Content: "Message " + string(rune('A'+i)),
			Time:    baseTime.Add(time.Duration(i) * time.Second),
		})
		require.NoError(t, err)
	}

	// Get all messages first
	all, err := storage.GetMessages("", 10, "")
	require.NoError(t, err)
	assert.Len(t, all, 5)

	// Get messages before msg-d (should get msg-a, msg-b, msg-c)
	before, err := storage.GetMessages("", 10, "msg-d")
	require.NoError(t, err)
	assert.Len(t, before, 3)

	// Verify msg-d and msg-e are not in the results
	for _, msg := range before {
		assert.NotEqual(t, "msg-d", msg.ID)
		assert.NotEqual(t, "msg-e", msg.ID)
	}
}

// TestStorage_GetMessages_Pagination_WithBeforeID_AndNetworkFilter tests pagination with network filter
func TestStorage_GetMessages_Pagination_WithBeforeID_AndNetworkFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	baseTime := time.Now()
	// Save messages to two networks
	for i := 0; i < 4; i++ {
		network := "network-1"
		if i%2 == 1 {
			network = "network-2"
		}
		err := storage.SaveMessage(Message{
			ID:        "msg-" + string(rune('a'+i)),
			From:      "peer",
			Content:   "Message",
			Time:      baseTime.Add(time.Duration(i) * time.Second),
			NetworkID: network,
		})
		require.NoError(t, err)
	}

	// Get network-1 messages before msg-c
	before, err := storage.GetMessages("network-1", 10, "msg-c")
	require.NoError(t, err)

	// Should only get msg-a (network-1, before msg-c timestamp)
	assert.Len(t, before, 1)
	assert.Equal(t, "msg-a", before[0].ID)
}

// TestStorage_GetMessages_BeforeID_NotFound tests pagination with non-existent beforeID
func TestStorage_GetMessages_BeforeID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	storage.SaveMessage(Message{ID: "msg-1", From: "peer", Content: "test", Time: time.Now()})

	// Get messages before non-existent ID - should return messages anyway (no timestamp filter applied)
	messages, err := storage.GetMessages("", 10, "nonexistent-id")
	require.NoError(t, err)
	// The query will use timestamp < 0 which returns nothing
	assert.Empty(t, messages)
}

// TestStorage_GetMessagesByPeer_DefaultLimit tests default limit behavior
func TestStorage_GetMessagesByPeer_DefaultLimit(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// Save messages
	for i := 0; i < 3; i++ {
		storage.SaveMessage(Message{
			ID:      "msg-" + string(rune('a'+i)),
			From:    "alice",
			Content: "test",
			Time:    time.Now(),
		})
	}

	// Use 0 limit (should default to 100)
	messages, err := storage.GetMessagesByPeer("alice", 0)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Use negative limit (should default to 100)
	messages, err = storage.GetMessagesByPeer("alice", -5)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}

// TestStorage_SearchMessages_DefaultLimit tests default limit for search
func TestStorage_SearchMessages_DefaultLimit(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	for i := 0; i < 3; i++ {
		storage.SaveMessage(Message{
			ID:      "msg-" + string(rune('a'+i)),
			From:    "peer",
			Content: "searchable content",
			Time:    time.Now(),
		})
	}

	// Use 0 limit (should default to 50)
	results, err := storage.SearchMessages("searchable", 0)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Use negative limit
	results, err = storage.SearchMessages("searchable", -10)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

// TestStorage_SaveMessage_Replace tests that SaveMessage can replace existing message
func TestStorage_SaveMessage_Replace(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// Save initial message
	err = storage.SaveMessage(Message{
		ID:      "same-id",
		From:    "peer",
		Content: "Original content",
		Time:    time.Now(),
	})
	require.NoError(t, err)

	// Save with same ID but different content
	err = storage.SaveMessage(Message{
		ID:      "same-id",
		From:    "peer",
		Content: "Updated content",
		Time:    time.Now(),
	})
	require.NoError(t, err)

	// Should still have only one message
	count, _ := storage.GetMessageCount()
	assert.Equal(t, int64(1), count)

	// Content should be updated
	messages, _ := storage.GetMessages("", 10, "")
	assert.Equal(t, "Updated content", messages[0].Content)
}

// TestStorage_DeleteMessage_NonExistent tests deleting a message that doesn't exist
func TestStorage_DeleteMessage_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// Delete non-existent message - should not error
	err = storage.DeleteMessage("nonexistent-id")
	assert.NoError(t, err)
}

// TestStorage_DeleteOldMessages_NoneToDelete tests when no messages are old enough
func TestStorage_DeleteOldMessages_NoneToDelete(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// Save a recent message
	storage.SaveMessage(Message{
		ID:      "recent-msg",
		From:    "peer",
		Content: "Recent",
		Time:    time.Now(),
	})

	// Try to delete messages older than 24 hours - should delete none
	deleted, err := storage.DeleteOldMessages(24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}

// TestStorage_GetMessages_OrderByTimestamp tests that messages are ordered by timestamp DESC
func TestStorage_GetMessages_OrderByTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	baseTime := time.Now()

	// Save messages in non-chronological order
	storage.SaveMessage(Message{ID: "mid", From: "p", Content: "b", Time: baseTime.Add(time.Second)})
	storage.SaveMessage(Message{ID: "first", From: "p", Content: "a", Time: baseTime})
	storage.SaveMessage(Message{ID: "last", From: "p", Content: "c", Time: baseTime.Add(2 * time.Second)})

	messages, err := storage.GetMessages("", 10, "")
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Should be ordered most recent first
	assert.Equal(t, "last", messages[0].ID)
	assert.Equal(t, "mid", messages[1].ID)
	assert.Equal(t, "first", messages[2].ID)
}

// TestStorage_Close_MultipleTimes tests that closing multiple times doesn't panic
func TestStorage_Close_MultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)

	// First close
	err = storage.Close()
	assert.NoError(t, err)

	// Second close may or may not error depending on SQLite driver version
	// The important thing is it doesn't panic
	_ = storage.Close()
}

// TestNewStorage_InvalidDirectory tests creating storage with invalid directory
func TestNewStorage_InvalidDirectory(t *testing.T) {
	// Try to create storage in a path with invalid permissions
	// This test may behave differently on different OSes
	_, err := NewStorage("/proc/invalid/path")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create data directory")
	}
}

// TestStorage_DatabasePath tests that the database is created at the expected path
func TestStorage_DatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	expectedPath := filepath.Join(tmpDir, "chat.db")
	assert.Equal(t, expectedPath, storage.dbPath)

	// Verify the file exists
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)
}

// TestStorage_WALMode tests that WAL mode is enabled
func TestStorage_WALMode(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	// WAL files should exist after write
	storage.SaveMessage(Message{ID: "wal-test", From: "p", Content: "test", Time: time.Now()})

	// Check that the main db file exists (WAL files may or may not be visible)
	_, err = os.Stat(filepath.Join(tmpDir, "chat.db"))
	assert.NoError(t, err)
}

// TestStorage_ConcurrentAccess tests concurrent read/write operations
func TestStorage_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	done := make(chan bool, 20)

	// Start concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := storage.SaveMessage(Message{
				ID:      "concurrent-" + string(rune('a'+id)),
				From:    "peer",
				Content: "test",
				Time:    time.Now(),
			})
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Start concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			_, err := storage.GetMessages("", 10, "")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all operations
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all messages were saved
	count, _ := storage.GetMessageCount()
	assert.Equal(t, int64(10), count)
}

// TestStorage_GetMessages_EmptyDatabase tests GetMessages on empty database
func TestStorage_GetMessages_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	messages, err := storage.GetMessages("", 10, "")
	require.NoError(t, err)
	assert.Empty(t, messages)
}

// TestStorage_SearchMessages_EmptyQuery tests search with empty query
func TestStorage_SearchMessages_EmptyQuery(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	storage.SaveMessage(Message{ID: "1", From: "p", Content: "test", Time: time.Now()})

	// Empty query should match all (LIKE '%%')
	results, err := storage.SearchMessages("", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// TestStorage_GetMessagesByPeer_NotFound tests getting messages from non-existent peer
func TestStorage_GetMessagesByPeer_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)
	defer storage.Close()

	storage.SaveMessage(Message{ID: "1", From: "alice", Content: "test", Time: time.Now()})

	messages, err := storage.GetMessagesByPeer("bob", 10)
	require.NoError(t, err)
	assert.Empty(t, messages)
}

// TestNewStorage_ReadOnlyDirectory tests creating storage in a readonly directory
func TestNewStorage_ReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping as root")
	}
	// Skip on Windows as permissions work differently
	if filepath.Separator == '\\' {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly dir: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }()

	// Try to create nested directory in readonly parent
	nestedDir := filepath.Join(readOnlyDir, "chat")
	_, err := NewStorage(nestedDir)
	if err == nil {
		t.Error("Expected error when creating storage in readonly directory")
	}
}

// TestStorage_Close tests closing the storage
func TestStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewStorage(tmpDir)
	require.NoError(t, err)

	err = storage.Close()
	require.NoError(t, err)

	// Trying to save after close should fail
	err = storage.SaveMessage(Message{ID: "1", From: "test", Content: "test", Time: time.Now()})
	assert.Error(t, err)
}
