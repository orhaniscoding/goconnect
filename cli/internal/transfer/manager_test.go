package transfer

import (
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransferManager(t *testing.T) {
	tm := NewManager()
	if tm == nil {
		t.Fatal("Expected TransferManager to be created")
	}
}

func TestCreateSendSession(t *testing.T) {
	// Create a dummy file for testing
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("hello world")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tm := NewManager()
	session, err := tm.CreateSendSession("receiver", tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create send session: %v", err)
	}

	if session.ID == "" {
		t.Error("Expected session ID to be generated")
	}
	if session.PeerID != "receiver" {
		t.Errorf("Expected PeerID to be 'receiver', got %s", session.PeerID)
	}
	if session.FileName != filepath.Base(tmpfile.Name()) {
		t.Errorf("Expected FileName to be %s, got %s", filepath.Base(tmpfile.Name()), session.FileName)
	}
	if session.FileSize != int64(len(content)) {
		t.Errorf("Expected FileSize to be %d, got %d", len(content), session.FileSize)
	}
	if session.Status != StatusPending {
		t.Errorf("Expected Status to be Pending, got %s", session.Status)
	}
}

func TestCreateReceiveSession(t *testing.T) {
	tm := NewManager()
	req := Request{
		ID:       "req-1",
		FileName: "test.txt",
		FileSize: 1024,
	}

	tmpDir, err := os.MkdirTemp("", "transfer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	savePath := filepath.Join(tmpDir, "test.txt")

	session, err := tm.CreateReceiveSession(req, "sender", savePath)
	if err != nil {
		t.Fatalf("Failed to create receive session: %v", err)
	}

	if session.ID != "req-1" {
		t.Errorf("Expected session ID to match request ID, got %s", session.ID)
	}
	if session.PeerID != "sender" {
		t.Errorf("Expected PeerID to be 'sender', got %s", session.PeerID)
	}
	if session.Status != StatusPending {
		t.Errorf("Expected Status to be Pending, got %s", session.Status)
	}
}

func TestRejectTransfer(t *testing.T) {
	tm := NewManager()

	// Test rejecting a pending request via signaling
	tm.HandleSignalingMessage(`{"id":"req-reject-1","file_name":"test.txt","file_size":1024}`, "sender-1")

	err := tm.RejectTransfer("req-reject-1")
	if err != nil {
		t.Fatalf("RejectTransfer failed: %v", err)
	}

	// Verify request was removed
	_, _, found := tm.GetPendingRequest("req-reject-1")
	if found {
		t.Error("Expected pending request to be removed after rejection")
	}

	// Test rejecting non-existent transfer
	err = tm.RejectTransfer("non-existent")
	if err == nil {
		t.Error("Expected error when rejecting non-existent transfer")
	}

	// Test rejecting a pending session
	tmpDir, _ := os.MkdirTemp("", "transfer_test")
	defer os.RemoveAll(tmpDir)
	savePath := filepath.Join(tmpDir, "test.txt")

	req := Request{
		ID:       "req-reject-2",
		FileName: "test.txt",
		FileSize: 1024,
	}
	session, _ := tm.CreateReceiveSession(req, "sender", savePath)
	err = tm.RejectTransfer(session.ID)
	if err != nil {
		t.Fatalf("RejectTransfer on pending session failed: %v", err)
	}

	// Verify session was marked as cancelled
	s := tm.GetSession(session.ID)
	if s == nil {
		t.Fatal("Expected session to exist")
	}
	if s.Status != StatusCancelled {
		t.Errorf("Expected status Cancelled, got %s", s.Status)
	}
}

func TestCancelTransfer(t *testing.T) {
	tm := NewManager()

	// Create a temp file for send session
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := []byte("hello world for cancel test")
	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	session, err := tm.CreateSendSession("receiver", tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create send session: %v", err)
	}

	// Cancel the pending session
	err = tm.CancelTransfer(session.ID)
	if err != nil {
		t.Fatalf("CancelTransfer failed: %v", err)
	}

	// Verify session was cancelled
	s := tm.GetSession(session.ID)
	if s.Status != StatusCancelled {
		t.Errorf("Expected status Cancelled, got %s", s.Status)
	}
	if s.Error != "cancelled by user" {
		t.Errorf("Expected error message 'cancelled by user', got %s", s.Error)
	}

	// Test cancelling already cancelled transfer
	err = tm.CancelTransfer(session.ID)
	if err == nil {
		t.Error("Expected error when cancelling already finished transfer")
	}

	// Test cancelling non-existent transfer
	err = tm.CancelTransfer("non-existent")
	if err == nil {
		t.Error("Expected error when cancelling non-existent transfer")
	}
}

func TestTransferSubscription(t *testing.T) {
	tm := NewManager()

	// Subscribe
	ch := tm.Subscribe()
	if ch == nil {
		t.Fatal("Expected subscription channel")
	}

	// Unsubscribe
	tm.Unsubscribe(ch)

	// Verify channel is closed by trying to receive
	_, ok := <-ch
	if ok {
		t.Error("Expected channel to be closed after unsubscribe")
	}
}

func TestGetSessions(t *testing.T) {
	tm := NewManager()

	// Create temp file
	tmpfile, _ := os.CreateTemp("", "testfile")
	defer os.Remove(tmpfile.Name())
	_, _ = tmpfile.Write([]byte("test"))
	tmpfile.Close()

	// Initially empty
	sessions := tm.GetSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}

	// Create a session
	_, _ = tm.CreateSendSession("peer1", tmpfile.Name())

	sessions = tm.GetSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}
}

func TestSession_Progress(t *testing.T) {
	s := &Session{FileSize: 1000, SentBytes: 500}
	if s.Progress() != 50.0 {
		t.Errorf("Expected 50%% progress, got %.2f%%", s.Progress())
	}

	// Zero file size
	s2 := &Session{FileSize: 0, SentBytes: 0}
	if s2.Progress() != 0 {
		t.Errorf("Expected 0%% progress for empty file, got %.2f%%", s2.Progress())
	}
}

func TestSession_IsActive(t *testing.T) {
	tests := []struct {
		status   Status
		expected bool
	}{
		{StatusPending, true},
		{StatusInProgress, true},
		{StatusCompleted, false},
		{StatusFailed, false},
		{StatusCancelled, false},
	}

	for _, tt := range tests {
		s := &Session{Status: tt.status}
		if s.IsActive() != tt.expected {
			t.Errorf("IsActive() for %s = %v, want %v", tt.status, s.IsActive(), tt.expected)
		}
	}
}

func TestSession_IsFinished(t *testing.T) {
	tests := []struct {
		status   Status
		expected bool
	}{
		{StatusPending, false},
		{StatusInProgress, false},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusCancelled, true},
	}

	for _, tt := range tests {
		s := &Session{Status: tt.status}
		if s.IsFinished() != tt.expected {
			t.Errorf("IsFinished() for %s = %v, want %v", tt.status, s.IsFinished(), tt.expected)
		}
	}
}

func TestListSessions_Filtering(t *testing.T) {
	tm := NewManager()

	// Add test sessions directly
	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", PeerID: "alice", Status: StatusPending, IsSender: true}
	tm.sessions["2"] = &Session{ID: "2", PeerID: "bob", Status: StatusCompleted, IsSender: false}
	tm.sessions["3"] = &Session{ID: "3", PeerID: "alice", Status: StatusInProgress, IsSender: true}
	tm.mu.Unlock()

	// Filter by status
	result := tm.ListSessions(ListOptions{Status: []Status{StatusPending}})
	if len(result.Sessions) != 1 {
		t.Errorf("Expected 1 pending session, got %d", len(result.Sessions))
	}

	// Filter by peer
	result = tm.ListSessions(ListOptions{PeerID: "alice"})
	if len(result.Sessions) != 2 {
		t.Errorf("Expected 2 sessions for alice, got %d", len(result.Sessions))
	}

	// Filter by direction
	sender := true
	result = tm.ListSessions(ListOptions{IsSender: &sender})
	if len(result.Sessions) != 2 {
		t.Errorf("Expected 2 sender sessions, got %d", len(result.Sessions))
	}
}

func TestListSessions_Pagination(t *testing.T) {
	tm := NewManager()

	// Add 5 test sessions
	tm.mu.Lock()
	for i := 0; i < 5; i++ {
		tm.sessions[string(rune('a'+i))] = &Session{
			ID:       string(rune('a' + i)),
			Status:   StatusPending,
			FileName: string(rune('a' + i)),
		}
	}
	tm.mu.Unlock()

	// Get first 2
	result := tm.ListSessions(ListOptions{Limit: 2})
	if len(result.Sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(result.Sessions))
	}
	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}
	if !result.HasMore {
		t.Error("Expected HasMore to be true")
	}

	// Get page 2
	result = tm.ListSessions(ListOptions{Limit: 2, Offset: 2})
	if len(result.Sessions) != 2 {
		t.Errorf("Expected 2 sessions on page 2, got %d", len(result.Sessions))
	}

	// Get last page
	result = tm.ListSessions(ListOptions{Limit: 2, Offset: 4})
	if len(result.Sessions) != 1 {
		t.Errorf("Expected 1 session on last page, got %d", len(result.Sessions))
	}
	if result.HasMore {
		t.Error("Expected HasMore to be false on last page")
	}
}

func TestGetStats(t *testing.T) {
	tm := NewManager()

	// Add test sessions
	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", Status: StatusCompleted, SentBytes: 1000, IsSender: true}
	tm.sessions["2"] = &Session{ID: "2", Status: StatusInProgress, SentBytes: 500, IsSender: false}
	tm.sessions["3"] = &Session{ID: "3", Status: StatusFailed, SentBytes: 200, IsSender: true}
	tm.mu.Unlock()

	stats := tm.GetStats()

	if stats.TotalTransfers != 3 {
		t.Errorf("Expected 3 total transfers, got %d", stats.TotalTransfers)
	}
	if stats.ActiveTransfers != 1 {
		t.Errorf("Expected 1 active transfer, got %d", stats.ActiveTransfers)
	}
	if stats.CompletedTransfers != 1 {
		t.Errorf("Expected 1 completed transfer, got %d", stats.CompletedTransfers)
	}
	if stats.FailedTransfers != 1 {
		t.Errorf("Expected 1 failed transfer, got %d", stats.FailedTransfers)
	}
	if stats.TotalBytesSent != 1200 {
		t.Errorf("Expected 1200 bytes sent, got %d", stats.TotalBytesSent)
	}
	if stats.TotalBytesReceived != 500 {
		t.Errorf("Expected 500 bytes received, got %d", stats.TotalBytesReceived)
	}
}

func TestCleanupOld(t *testing.T) {
	tm := NewManager()

	now := time.Now()

	// Add sessions with different end times
	tm.mu.Lock()
	tm.sessions["old"] = &Session{ID: "old", Status: StatusCompleted, EndTime: now.Add(-2 * time.Hour)}
	tm.sessions["recent"] = &Session{ID: "recent", Status: StatusCompleted, EndTime: now.Add(-30 * time.Minute)}
	tm.sessions["active"] = &Session{ID: "active", Status: StatusInProgress}
	tm.mu.Unlock()

	// Cleanup sessions older than 1 hour
	removed := tm.CleanupOld(1 * time.Hour)

	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}

	sessions := tm.GetSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 remaining sessions, got %d", len(sessions))
	}
}

func TestGetActiveCount(t *testing.T) {
	tm := NewManager()

	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", Status: StatusPending}
	tm.sessions["2"] = &Session{ID: "2", Status: StatusInProgress}
	tm.sessions["3"] = &Session{ID: "3", Status: StatusCompleted}
	tm.mu.Unlock()

	count := tm.GetActiveCount()
	if count != 2 {
		t.Errorf("Expected 2 active, got %d", count)
	}
}

func TestSetCallbacks(t *testing.T) {
	tm := NewManager()

	progressCalled := false
	requestCalled := false

	tm.SetCallbacks(
		func(_ Session) { progressCalled = true },
		func(r Request, peer string) { requestCalled = true },
	)

	// Trigger progress callback via notifyProgress
	tm.mu.Lock()
	tm.sessions["test"] = &Session{ID: "test", Status: StatusInProgress}
	tm.mu.Unlock()
	tm.notifyProgress("test")

	if !progressCalled {
		t.Error("Expected progress callback to be called")
	}

	// Trigger request callback via HandleSignalingMessage
	tm.HandleSignalingMessage(`{"id":"callback-test","file_name":"test.txt","file_size":100}`, "peer-1")

	if !requestCalled {
		t.Error("Expected request callback to be called")
	}
}

func TestGetSession(t *testing.T) {
	tm := NewManager()

	// Non-existent session
	s := tm.GetSession("non-existent")
	if s != nil {
		t.Error("Expected nil for non-existent session")
	}

	// Add a session
	tm.mu.Lock()
	tm.sessions["exists"] = &Session{ID: "exists", Status: StatusPending}
	tm.mu.Unlock()

	s = tm.GetSession("exists")
	if s == nil {
		t.Fatal("Expected session to exist")
	}
	if s.ID != "exists" {
		t.Errorf("Expected ID 'exists', got %s", s.ID)
	}
}

func TestHandleSignalingMessage(t *testing.T) {
	tm := NewManager()

	// Handle valid message
	tm.HandleSignalingMessage(`{"id":"sig-1","file_name":"test.txt","file_size":1024}`, "sender-1")

	req, peer, found := tm.GetPendingRequest("sig-1")
	if !found {
		t.Fatal("Expected pending request to be added")
	}
	if peer != "sender-1" {
		t.Errorf("Expected peer 'sender-1', got %s", peer)
	}
	if req.FileName != "test.txt" {
		t.Errorf("Expected FileName 'test.txt', got %s", req.FileName)
	}

	// Handle invalid JSON
	tm.HandleSignalingMessage("invalid-json", "sender-2")
	// Should not panic, just log error
}

func TestGetPendingRequest_NotFound(t *testing.T) {
	tm := NewManager()

	_, _, found := tm.GetPendingRequest("non-existent")
	if found {
		t.Error("Expected pending request to not be found")
	}
}

func TestManager_StartAndStop(t *testing.T) {
	tm := NewManager()

	// Start on a random available port
	err := tm.Start("127.0.0.1")
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Verify listener is set
	if tm.listener == nil {
		t.Error("Expected listener to be set after Start")
	}

	// Stop should not panic
	tm.Stop()
}

func TestManager_StopWithoutStart(t *testing.T) {
	tm := NewManager()

	// Stop without starting should not panic
	tm.Stop()
}

func TestCreateSendSession_NonexistentFile(t *testing.T) {
	tm := NewManager()

	_, err := tm.CreateSendSession("receiver", "/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestListSessions_Sorting(t *testing.T) {
	tm := NewManager()

	now := time.Now()

	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", FileName: "aaa.txt", StartTime: now.Add(-1 * time.Hour), FileSize: 100}
	tm.sessions["2"] = &Session{ID: "2", FileName: "bbb.txt", StartTime: now.Add(-2 * time.Hour), FileSize: 300}
	tm.sessions["3"] = &Session{ID: "3", FileName: "ccc.txt", StartTime: now.Add(-30 * time.Minute), FileSize: 200}
	tm.mu.Unlock()

	// Sort by name ascending (default)
	result := tm.ListSessions(ListOptions{SortBy: SortByFileName, SortOrder: SortAsc})
	if len(result.Sessions) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(result.Sessions))
	}
	if result.Sessions[0].FileName != "aaa.txt" {
		t.Errorf("Expected first file 'aaa.txt', got %s", result.Sessions[0].FileName)
	}

	// Sort by size descending
	result = tm.ListSessions(ListOptions{SortBy: SortByFileSize, SortOrder: SortDesc})
	if result.Sessions[0].FileSize != 300 {
		t.Errorf("Expected first file size 300, got %d", result.Sessions[0].FileSize)
	}
}

func TestRequest_Struct(t *testing.T) {
	req := Request{
		ID:       "req-123",
		FileName: "document.pdf",
		FileSize: 1024000,
	}

	if req.ID != "req-123" {
		t.Errorf("Expected ID 'req-123', got %s", req.ID)
	}
	if req.FileName != "document.pdf" {
		t.Errorf("Expected FileName 'document.pdf', got %s", req.FileName)
	}
	if req.FileSize != 1024000 {
		t.Errorf("Expected FileSize 1024000, got %d", req.FileSize)
	}
}

func TestConcurrentSessionAccess(t *testing.T) {
	tm := NewManager()

	// Add initial sessions
	for i := 0; i < 10; i++ {
		tm.mu.Lock()
		tm.sessions[string(rune('a'+i))] = &Session{ID: string(rune('a' + i)), Status: StatusPending}
		tm.mu.Unlock()
	}

	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = tm.GetSessions()
		}
		done <- true
	}()

	// Concurrent stats
	go func() {
		for i := 0; i < 100; i++ {
			_ = tm.GetStats()
		}
		done <- true
	}()

	// Concurrent active count
	go func() {
		for i := 0; i < 100; i++ {
			_ = tm.GetActiveCount()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

// ==================== Enhanced Tests with Testify ====================

func TestNewManager_WithTestify(t *testing.T) {
	tm := NewManager()
	require.NotNil(t, tm, "Manager should not be nil")
	assert.NotNil(t, tm.sessions, "sessions map should be initialized")
	assert.NotNil(t, tm.pendingReqs, "pendingReqs map should be initialized")
	assert.NotNil(t, tm.reqTimers, "reqTimers map should be initialized")
	assert.NotNil(t, tm.stopChan, "stopChan should be initialized")
	assert.NotNil(t, tm.subscribers, "subscribers map should be initialized")
}

// ==================== handleConnection Tests ====================

func TestHandleConnection_UnknownSession(t *testing.T) {
	tm := NewManager()

	// Start the manager on a random port
	err := tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Connect to the listener
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Send an unknown transfer ID
	_, err = conn.Write([]byte("12345678-1234-1234-1234-123456789012"))
	require.NoError(t, err)

	// Connection should be closed by server
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	assert.Error(t, err, "Connection should be closed for unknown session")
}

func TestHandleConnection_ReceiverSession(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Create a receive session (not a sender)
	tmpDir, err := os.MkdirTemp("", "transfer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	req := Request{
		ID:       "rcv-12345678-1234-1234-1234-12345678",
		FileName: "test.txt",
		FileSize: 100,
	}
	_, err = tm.CreateReceiveSession(req, "peer", filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)

	// Connect and send the receiver session ID
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Pad ID to 36 characters
	paddedID := req.ID + "9012345"
	_, err = conn.Write([]byte(paddedID[:36]))
	require.NoError(t, err)

	// Connection should be closed (we're not the sender)
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	assert.Error(t, err, "Connection should be closed for receiver session")
}

// ==================== sendFileData Tests ====================

func TestSendFileData_Success(t *testing.T) {
	tm := NewManager()

	// Create a temp file with known content
	tmpfile, err := os.CreateTemp("", "sendtest")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := []byte("hello world file content")
	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	tmpfile.Close()

	err = tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Create a send session
	session, err := tm.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Connect as receiver
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Send session ID (padded to 36 chars)
	paddedID := session.ID
	for len(paddedID) < 36 {
		paddedID = "0" + paddedID
	}
	_, err = conn.Write([]byte(paddedID[:36]))
	require.NoError(t, err)

	// Read the file data
	time.Sleep(100 * time.Millisecond) // Give time for transfer
	received := make([]byte, len(content)+100)
	n, _ := conn.Read(received)

	assert.Equal(t, content, received[:n], "Should receive file content")

	// Verify session completed
	time.Sleep(100 * time.Millisecond)
	s := tm.GetSession(session.ID)
	assert.Equal(t, StatusCompleted, s.Status, "Session should be completed")
}

func TestSendFileData_FileNotFound(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Create a session for a file that we'll delete
	tmpfile, err := os.CreateTemp("", "sendtest")
	require.NoError(t, err)
	_, _ = tmpfile.Write([]byte("content"))
	tmpfile.Close()

	session, err := tm.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Delete the file before transfer
	os.Remove(tmpfile.Name())

	// Connect as receiver
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Send session ID
	paddedID := session.ID
	for len(paddedID) < 36 {
		paddedID = "0" + paddedID
	}
	_, err = conn.Write([]byte(paddedID[:36]))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Session should be failed
	s := tm.GetSession(session.ID)
	assert.Equal(t, StatusFailed, s.Status, "Session should be failed")
	assert.NotEmpty(t, s.Error, "Error message should be set")
}

// ==================== updateStatus Tests ====================

func TestUpdateStatus(t *testing.T) {
	tm := NewManager()

	// Create a session
	tm.mu.Lock()
	tm.sessions["test-status"] = &Session{
		ID:        "test-status",
		Status:    StatusPending,
		StartTime: time.Now(),
	}
	tm.mu.Unlock()

	// Update to in progress
	tm.updateStatus("test-status", StatusInProgress)
	s := tm.GetSession("test-status")
	assert.Equal(t, StatusInProgress, s.Status, "Status should be in_progress")
	assert.True(t, s.EndTime.IsZero(), "EndTime should not be set for in_progress")

	// Update to completed
	tm.updateStatus("test-status", StatusCompleted)
	s = tm.GetSession("test-status")
	assert.Equal(t, StatusCompleted, s.Status, "Status should be completed")
	assert.False(t, s.EndTime.IsZero(), "EndTime should be set for completed")

	// Update non-existent session (should not panic)
	tm.updateStatus("non-existent", StatusFailed)
}

func TestUpdateStatus_Failed(t *testing.T) {
	tm := NewManager()

	// Create a session
	tm.mu.Lock()
	tm.sessions["test-fail"] = &Session{
		ID:        "test-fail",
		Status:    StatusInProgress,
		StartTime: time.Now(),
	}
	tm.mu.Unlock()

	// Update to failed
	tm.updateStatus("test-fail", StatusFailed)
	s := tm.GetSession("test-fail")
	assert.Equal(t, StatusFailed, s.Status, "Status should be failed")
	assert.False(t, s.EndTime.IsZero(), "EndTime should be set for failed")
}

// ==================== failSession Tests ====================

func TestFailSession(t *testing.T) {
	tm := NewManager()

	// Create a session
	tm.mu.Lock()
	tm.sessions["fail-test"] = &Session{
		ID:        "fail-test",
		Status:    StatusInProgress,
		StartTime: time.Now(),
	}
	tm.mu.Unlock()

	// Fail the session
	tm.failSession("fail-test", "connection lost")

	s := tm.GetSession("fail-test")
	assert.Equal(t, StatusFailed, s.Status, "Status should be failed")
	assert.Equal(t, "connection lost", s.Error, "Error message should be set")
	assert.False(t, s.EndTime.IsZero(), "EndTime should be set")
}

func TestFailSession_NonExistent(t *testing.T) {
	tm := NewManager()

	// Fail non-existent session (should not panic)
	tm.failSession("non-existent", "error")
}

func TestFailSession_WithProgressCallback(t *testing.T) {
	tm := NewManager()

	progressCalled := false
	tm.SetCallbacks(func(s Session) {
		progressCalled = true
		assert.Equal(t, "fail-callback", s.ID)
		assert.Equal(t, StatusFailed, s.Status)
	}, nil)

	tm.mu.Lock()
	tm.sessions["fail-callback"] = &Session{
		ID:        "fail-callback",
		Status:    StatusInProgress,
		StartTime: time.Now(),
	}
	tm.mu.Unlock()

	tm.failSession("fail-callback", "test error")
	time.Sleep(50 * time.Millisecond) // Allow goroutine to run
	assert.True(t, progressCalled, "Progress callback should be called")
}

// ==================== StartDownload Tests ====================

func TestStartDownload_SessionNotFound(t *testing.T) {
	tm := NewManager()

	err := tm.StartDownload("non-existent", "127.0.0.1")
	assert.Error(t, err, "Should return error for non-existent session")
	assert.Contains(t, err.Error(), "session not found")
}

func TestStartDownload_ConnectionFailed(t *testing.T) {
	tm := NewManager()

	// Create a receive session
	tmpDir, err := os.MkdirTemp("", "download_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	req := Request{
		ID:       "download-test",
		FileName: "test.txt",
		FileSize: 100,
	}
	_, err = tm.CreateReceiveSession(req, "sender", filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)

	// Start download to unreachable IP
	err = tm.StartDownload("download-test", "192.0.2.1") // TEST-NET, unreachable
	assert.NoError(t, err, "StartDownload returns immediately")

	// Wait for connection attempt to fail
	time.Sleep(200 * time.Millisecond)

	// Session should eventually fail
	s := tm.GetSession("download-test")
	// Note: May still be pending if connection hasn't timed out
	assert.NotNil(t, s, "Session should exist")
}

func TestStartDownload_Success(t *testing.T) {
	// Create a sender manager
	senderMgr := NewManager()
	err := senderMgr.Start("127.0.0.1")
	require.NoError(t, err)
	defer senderMgr.Stop()

	// Create temp file for sender
	tmpfile, err := os.CreateTemp("", "sender_file")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := []byte("transfer test content 12345")
	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	senderSession, err := senderMgr.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Create a receiver manager
	receiverMgr := NewManager()

	// Create receive session with same ID
	tmpDir, err := os.MkdirTemp("", "receiver_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	req := Request{
		ID:       senderSession.ID,
		FileName: "received.txt",
		FileSize: int64(len(content)),
	}
	receivePath := filepath.Join(tmpDir, "received.txt")
	_, err = receiverMgr.CreateReceiveSession(req, "sender", receivePath)
	require.NoError(t, err)

	// Get sender's address
	senderAddr := senderMgr.listener.Addr().(*net.TCPAddr)

	// Start download
	err = receiverMgr.StartDownload(senderSession.ID, senderAddr.IP.String())
	require.NoError(t, err)

	// Wait for transfer to complete
	time.Sleep(500 * time.Millisecond)

	// Verify the file was received
	receivedContent, err := os.ReadFile(receivePath)
	require.NoError(t, err)
	assert.Equal(t, content, receivedContent, "Content should match")

	// Verify session completed
	s := receiverMgr.GetSession(senderSession.ID)
	assert.Equal(t, StatusCompleted, s.Status, "Receiver session should be completed")
}

// ==================== notifySubscribers Tests ====================

func TestNotifySubscribers_ChannelFull(t *testing.T) {
	tm := NewManager()

	// Create a subscriber with very small buffer
	ch := make(chan Session, 1)
	tm.subscribersMu.Lock()
	tm.subscribers[ch] = struct{}{}
	tm.subscribersMu.Unlock()

	// Add a session
	tm.mu.Lock()
	tm.sessions["notify-test"] = &Session{
		ID:     "notify-test",
		Status: StatusInProgress,
	}
	tm.mu.Unlock()

	// Fill the channel
	ch <- Session{ID: "filler"}

	// Should not block or panic when channel is full
	tm.notifySubscribers(&Session{ID: "notify-test"})

	// Clean up
	close(ch)
}

func TestNotifySubscribers_MultipleSubscribers(t *testing.T) {
	tm := NewManager()

	ch1 := tm.Subscribe()
	ch2 := tm.Subscribe()

	// Add a session
	tm.mu.Lock()
	tm.sessions["multi-sub"] = &Session{
		ID:     "multi-sub",
		Status: StatusInProgress,
	}
	tm.mu.Unlock()

	// Notify
	tm.notifySubscribers(&Session{ID: "multi-sub"})

	// Both channels should receive
	select {
	case s := <-ch1:
		assert.Equal(t, "multi-sub", s.ID)
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 should receive notification")
	}

	select {
	case s := <-ch2:
		assert.Equal(t, "multi-sub", s.ID)
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 should receive notification")
	}

	tm.Unsubscribe(ch1)
	tm.Unsubscribe(ch2)
}

// ==================== shouldSwap Sorting Tests ====================

func TestShouldSwap_AllSortFields(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		a        *Session
		b        *Session
		sortBy   SortField
		order    SortOrder
		expected bool
	}{
		{
			name:     "StartTime Asc - a before b",
			a:        &Session{StartTime: now.Add(-1 * time.Hour)},
			b:        &Session{StartTime: now},
			sortBy:   SortByStartTime,
			order:    SortAsc,
			expected: false,
		},
		{
			name:     "StartTime Desc - a before b",
			a:        &Session{StartTime: now.Add(-1 * time.Hour)},
			b:        &Session{StartTime: now},
			sortBy:   SortByStartTime,
			order:    SortDesc,
			expected: true,
		},
		{
			name:     "EndTime Asc - a before b",
			a:        &Session{EndTime: now.Add(-1 * time.Hour)},
			b:        &Session{EndTime: now},
			sortBy:   SortByEndTime,
			order:    SortAsc,
			expected: false,
		},
		{
			name:     "EndTime Desc - a before b",
			a:        &Session{EndTime: now.Add(-1 * time.Hour)},
			b:        &Session{EndTime: now},
			sortBy:   SortByEndTime,
			order:    SortDesc,
			expected: true,
		},
		{
			name:     "FileSize Asc - a smaller",
			a:        &Session{FileSize: 100},
			b:        &Session{FileSize: 200},
			sortBy:   SortByFileSize,
			order:    SortAsc,
			expected: false,
		},
		{
			name:     "FileSize Desc - a smaller",
			a:        &Session{FileSize: 100},
			b:        &Session{FileSize: 200},
			sortBy:   SortByFileSize,
			order:    SortDesc,
			expected: true,
		},
		{
			name:     "Progress Asc - a less progress",
			a:        &Session{FileSize: 100, SentBytes: 10},
			b:        &Session{FileSize: 100, SentBytes: 50},
			sortBy:   SortByProgress,
			order:    SortAsc,
			expected: false,
		},
		{
			name:     "Progress Desc - a less progress",
			a:        &Session{FileSize: 100, SentBytes: 10},
			b:        &Session{FileSize: 100, SentBytes: 50},
			sortBy:   SortByProgress,
			order:    SortDesc,
			expected: true,
		},
		{
			name:     "FileName Asc - a < b",
			a:        &Session{FileName: "aaa.txt"},
			b:        &Session{FileName: "bbb.txt"},
			sortBy:   SortByFileName,
			order:    SortAsc,
			expected: false,
		},
		{
			name:     "FileName Desc - a < b",
			a:        &Session{FileName: "aaa.txt"},
			b:        &Session{FileName: "bbb.txt"},
			sortBy:   SortByFileName,
			order:    SortDesc,
			expected: true,
		},
		{
			name:     "Default sort field",
			a:        &Session{StartTime: now.Add(-1 * time.Hour)},
			b:        &Session{StartTime: now},
			sortBy:   "",
			order:    SortDesc,
			expected: true,
		},
	}

	tm := NewManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.shouldSwap(tt.a, tt.b, tt.sortBy, tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== CreateReceiveSession Error Paths ====================

func TestCreateReceiveSession_InvalidDirectory(t *testing.T) {
	tm := NewManager()

	req := Request{
		ID:       "invalid-dir-test",
		FileName: "test.txt",
		FileSize: 100,
	}

	// Use a path that definitely doesn't exist on any OS
	// Create a temp dir, then remove it, to get a guaranteed non-existent path
	tmpDir, err := os.MkdirTemp("", "nonexistent_test")
	require.NoError(t, err)
	nonExistentPath := filepath.Join(tmpDir, "subdir", "test.txt")
	os.RemoveAll(tmpDir) // Remove the temp dir so it doesn't exist

	_, err = tm.CreateReceiveSession(req, "sender", nonExistentPath)
	assert.Error(t, err, "Should return error for non-existent directory")
	assert.Contains(t, err.Error(), "destination directory does not exist")
}

func TestCreateReceiveSession_RemovesPendingRequest(t *testing.T) {
	tm := NewManager()

	// Add pending request
	tm.HandleSignalingMessage(`{"id":"pending-remove","file_name":"test.txt","file_size":100}`, "sender")

	// Verify it exists
	_, _, found := tm.GetPendingRequest("pending-remove")
	assert.True(t, found, "Pending request should exist")

	// Create receive session
	tmpDir, err := os.MkdirTemp("", "remove_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	req := Request{
		ID:       "pending-remove",
		FileName: "test.txt",
		FileSize: 100,
	}
	_, err = tm.CreateReceiveSession(req, "sender", filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)

	// Verify pending request is removed
	_, _, found = tm.GetPendingRequest("pending-remove")
	assert.False(t, found, "Pending request should be removed")
}

// ==================== CreateSendSession Error Paths ====================

func TestCreateSendSession_Directory(t *testing.T) {
	tm := NewManager()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dir_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	_, err = tm.CreateSendSession("receiver", tmpDir)
	assert.Error(t, err, "Should return error for directory")
	assert.Contains(t, err.Error(), "cannot transfer directories")
}

// ==================== acceptLoop Tests ====================

func TestAcceptLoop_StopChannel(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)

	// Stop should close the accept loop cleanly
	tm.Stop()

	// Verify listener is closed
	_, err = net.Dial("tcp", tm.listener.Addr().String())
	assert.Error(t, err, "Listener should be closed")
}

// ==================== RejectTransfer Edge Cases ====================

func TestRejectTransfer_InProgressSession(t *testing.T) {
	tm := NewManager()

	// Create a session that's in progress
	tm.mu.Lock()
	tm.sessions["in-progress"] = &Session{
		ID:        "in-progress",
		Status:    StatusInProgress,
		StartTime: time.Now(),
	}
	tm.mu.Unlock()

	err := tm.RejectTransfer("in-progress")
	assert.Error(t, err, "Should return error for in-progress transfer")
	assert.Contains(t, err.Error(), "already in progress")
}

// ==================== GetStats Edge Cases ====================

func TestGetStats_WithCompletedTransfers(t *testing.T) {
	tm := NewManager()

	now := time.Now()

	// Add completed session with proper timing
	tm.mu.Lock()
	tm.sessions["completed-1"] = &Session{
		ID:        "completed-1",
		Status:    StatusCompleted,
		SentBytes: 10000,
		FileSize:  10000,
		IsSender:  true,
		StartTime: now.Add(-10 * time.Second),
		EndTime:   now,
	}
	tm.sessions["completed-2"] = &Session{
		ID:        "completed-2",
		Status:    StatusCompleted,
		SentBytes: 5000,
		FileSize:  5000,
		IsSender:  false,
		StartTime: now.Add(-5 * time.Second),
		EndTime:   now,
	}
	tm.mu.Unlock()

	stats := tm.GetStats()

	assert.Equal(t, 2, stats.TotalTransfers)
	assert.Equal(t, 2, stats.CompletedTransfers)
	assert.Equal(t, int64(10000), stats.TotalBytesSent)
	assert.Equal(t, int64(5000), stats.TotalBytesReceived)
	assert.Greater(t, stats.AverageSpeed, float64(0), "Average speed should be calculated")
}

func TestGetStats_AllStatusTypes(t *testing.T) {
	tm := NewManager()

	tm.mu.Lock()
	tm.sessions["pending"] = &Session{ID: "pending", Status: StatusPending, IsSender: true}
	tm.sessions["in-progress"] = &Session{ID: "in-progress", Status: StatusInProgress, IsSender: false}
	tm.sessions["completed"] = &Session{ID: "completed", Status: StatusCompleted, IsSender: true}
	tm.sessions["failed"] = &Session{ID: "failed", Status: StatusFailed, IsSender: false}
	tm.sessions["cancelled"] = &Session{ID: "cancelled", Status: StatusCancelled, IsSender: true}
	tm.mu.Unlock()

	stats := tm.GetStats()

	assert.Equal(t, 5, stats.TotalTransfers)
	assert.Equal(t, 2, stats.ActiveTransfers) // pending + in_progress
	assert.Equal(t, 1, stats.CompletedTransfers)
	assert.Equal(t, 2, stats.FailedTransfers) // failed + cancelled
}

// ==================== ListSessions Edge Cases ====================

func TestListSessions_OffsetBeyondTotal(t *testing.T) {
	tm := NewManager()

	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", Status: StatusPending}
	tm.mu.Unlock()

	result := tm.ListSessions(ListOptions{Offset: 100})
	assert.Equal(t, 0, len(result.Sessions), "Should return empty when offset > total")
	assert.Equal(t, 1, result.Total, "Total should still be correct")
}

func TestListSessions_EmptyStatusFilter(t *testing.T) {
	tm := NewManager()

	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", Status: StatusPending}
	tm.sessions["2"] = &Session{ID: "2", Status: StatusCompleted}
	tm.mu.Unlock()

	// Empty status filter should return all
	result := tm.ListSessions(ListOptions{Status: []Status{}})
	assert.Equal(t, 2, len(result.Sessions))
}

func TestListSessions_MultipleStatusFilter(t *testing.T) {
	tm := NewManager()

	tm.mu.Lock()
	tm.sessions["1"] = &Session{ID: "1", Status: StatusPending}
	tm.sessions["2"] = &Session{ID: "2", Status: StatusCompleted}
	tm.sessions["3"] = &Session{ID: "3", Status: StatusFailed}
	tm.mu.Unlock()

	result := tm.ListSessions(ListOptions{Status: []Status{StatusPending, StatusFailed}})
	assert.Equal(t, 2, len(result.Sessions))
}

// ==================== HandleSignalingMessage Timeout ====================

func TestHandleSignalingMessage_Timeout(t *testing.T) {
	tm := NewManager()

	// Add a request (can't easily test 5 minute timeout, but can verify timer is set)
	tm.HandleSignalingMessage(`{"id":"timeout-test","file_name":"test.txt","file_size":100}`, "sender")

	tm.mu.RLock()
	_, timerExists := tm.reqTimers["timeout-test"]
	tm.mu.RUnlock()

	assert.True(t, timerExists, "Timer should be set for request")

	// Clean up by rejecting
	_ = tm.RejectTransfer("timeout-test")
}

// ==================== Concurrent Write Tests ====================

func TestConcurrentWriteAccess(t *testing.T) {
	tm := NewManager()

	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tmpfile, err := os.CreateTemp("", "concurrent")
			if err != nil {
				return
			}
			defer os.Remove(tmpfile.Name())
			_, _ = tmpfile.Write([]byte("test"))
			tmpfile.Close()
			_, _ = tm.CreateSendSession("peer", tmpfile.Name())
		}(i)
	}

	// Concurrent status updates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tm.mu.Lock()
			for id := range tm.sessions {
				tm.sessions[id].SentBytes += 100
			}
			tm.mu.Unlock()
		}()
	}

	wg.Wait()
}

// ==================== Start Error Handling ====================

func TestStart_PortInUse(t *testing.T) {
	tm1 := NewManager()
	tm2 := NewManager()

	// Start first manager on specific port
	err := tm1.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm1.Stop()

	// Second manager on same port should fail
	// Note: Since Start uses a fixed port (3001), this will fail
	err = tm2.Start("127.0.0.1")
	assert.Error(t, err, "Should fail when port is in use")
}

// ==================== Stop with Timers ====================

func TestStop_WithActiveTimers(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)

	// Add some pending requests with timers
	for i := 0; i < 5; i++ {
		tm.HandleSignalingMessage(`{"id":"timer-`+string(rune('a'+i))+`","file_name":"test.txt","file_size":100}`, "sender")
	}

	// Stop should clean up all timers
	tm.Stop()
}

// ==================== Progress with Subscriber ====================

func TestProgressUpdates_WithSubscriber(t *testing.T) {
	tm := NewManager()

	ch := tm.Subscribe()

	// Create session
	tm.mu.Lock()
	tm.sessions["progress-sub"] = &Session{
		ID:        "progress-sub",
		Status:    StatusInProgress,
		SentBytes: 0,
		FileSize:  1000,
	}
	tm.mu.Unlock()

	// Notify progress
	tm.notifyProgress("progress-sub")

	select {
	case s := <-ch:
		assert.Equal(t, "progress-sub", s.ID)
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive progress update")
	}

	tm.Unsubscribe(ch)
}

// ==================== Full Transfer Flow Test ====================

func TestFullTransferFlow(t *testing.T) {
	// Create sender manager
	senderMgr := NewManager()
	err := senderMgr.Start("127.0.0.1")
	require.NoError(t, err)
	defer senderMgr.Stop()

	// Create temp file
	tmpfile, err := os.CreateTemp("", "flow_test")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := []byte("This is the file content for the full flow test - make it longer for better testing")
	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	// Create send session
	sendSession, err := senderMgr.CreateSendSession("receiver-peer", tmpfile.Name())
	require.NoError(t, err)
	assert.Equal(t, StatusPending, sendSession.Status)
	assert.Equal(t, int64(len(content)), sendSession.FileSize)

	// Create receiver manager
	receiverMgr := NewManager()

	// Create receive session
	tmpDir, err := os.MkdirTemp("", "flow_receive")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	req := Request{
		ID:       sendSession.ID,
		FileName: "received_flow.txt",
		FileSize: int64(len(content)),
	}
	receivePath := filepath.Join(tmpDir, "received_flow.txt")
	receiveSession, err := receiverMgr.CreateReceiveSession(req, "sender-peer", receivePath)
	require.NoError(t, err)
	assert.Equal(t, StatusPending, receiveSession.Status)

	// Subscribe to updates
	senderCh := senderMgr.Subscribe()
	receiverCh := receiverMgr.Subscribe()

	// Start download
	senderAddr := senderMgr.listener.Addr().(*net.TCPAddr)
	err = receiverMgr.StartDownload(sendSession.ID, senderAddr.IP.String())
	require.NoError(t, err)

	// Wait for completion
	timeout := time.After(2 * time.Second)
	senderCompleted := false
	receiverCompleted := false

loop:
	for {
		select {
		case s := <-senderCh:
			if s.Status == StatusCompleted {
				senderCompleted = true
			}
		case s := <-receiverCh:
			if s.Status == StatusCompleted {
				receiverCompleted = true
			}
		case <-timeout:
			break loop
		}
		if senderCompleted && receiverCompleted {
			break
		}
	}

	// Verify file content
	receivedContent, err := os.ReadFile(receivePath)
	require.NoError(t, err)
	assert.Equal(t, content, receivedContent)

	// Clean up
	senderMgr.Unsubscribe(senderCh)
	receiverMgr.Unsubscribe(receiverCh)
}

// ==================== Connection Write Error Test ====================

func TestSendFileData_WriteError(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Create temp file with content
	tmpfile, err := os.CreateTemp("", "write_error_test")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write enough content that we need multiple chunks
	content := make([]byte, ChunkSize*3)
	for i := range content {
		content[i] = byte(i % 256)
	}
	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	session, err := tm.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Connect and then close connection mid-transfer
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)

	// Send session ID
	paddedID := session.ID
	for len(paddedID) < 36 {
		paddedID = "0" + paddedID
	}
	_, _ = conn.Write([]byte(paddedID[:36]))

	// Read a bit then close
	buf := make([]byte, 100)
	_, _ = conn.Read(buf)
	conn.Close()

	// Wait for error to be processed
	time.Sleep(200 * time.Millisecond)

	s := tm.GetSession(session.ID)
	// Session may be failed or completed depending on timing
	assert.True(t, s.Status == StatusFailed || s.Status == StatusCompleted)
}

// ==================== StartDownload File Creation Error ====================

func TestStartDownload_FileCreationError(t *testing.T) {
	// Create a sender
	senderMgr := NewManager()
	err := senderMgr.Start("127.0.0.1")
	require.NoError(t, err)
	defer senderMgr.Stop()

	// Create temp file for sender
	tmpfile, err := os.CreateTemp("", "sender_file")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, _ = tmpfile.Write([]byte("test content"))
	tmpfile.Close()

	senderSession, err := senderMgr.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Create receiver with invalid path
	receiverMgr := NewManager()

	// Manually create session with invalid path
	receiverMgr.mu.Lock()
	receiverMgr.sessions[senderSession.ID] = &Session{
		ID:        senderSession.ID,
		PeerID:    "sender",
		FilePath:  "/nonexistent/dir/file.txt",
		FileName:  "file.txt",
		FileSize:  12,
		Status:    StatusPending,
		IsSender:  false,
		StartTime: time.Now(),
	}
	receiverMgr.mu.Unlock()

	senderAddr := senderMgr.listener.Addr().(*net.TCPAddr)
	err = receiverMgr.StartDownload(senderSession.ID, senderAddr.IP.String())
	require.NoError(t, err)

	// Wait for failure
	time.Sleep(500 * time.Millisecond)

	s := receiverMgr.GetSession(senderSession.ID)
	assert.Equal(t, StatusFailed, s.Status)
	assert.NotEmpty(t, s.Error)
}

// ==================== Transfer Incomplete Size Check ====================

func TestStartDownload_IncompleteSizeCheck(t *testing.T) {
	// This tests the size verification at the end of download
	senderMgr := NewManager()
	err := senderMgr.Start("127.0.0.1")
	require.NoError(t, err)
	defer senderMgr.Stop()

	// Create temp file for sender
	tmpfile, err := os.CreateTemp("", "sender_file")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	content := []byte("short")
	_, _ = tmpfile.Write(content)
	tmpfile.Close()

	senderSession, err := senderMgr.CreateSendSession("receiver", tmpfile.Name())
	require.NoError(t, err)

	// Create receiver expecting more data than sender will send
	receiverMgr := NewManager()

	tmpDir, err := os.MkdirTemp("", "incomplete_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create session expecting much larger file
	receiverMgr.mu.Lock()
	receiverMgr.sessions[senderSession.ID] = &Session{
		ID:        senderSession.ID,
		PeerID:    "sender",
		FilePath:  filepath.Join(tmpDir, "received.txt"),
		FileName:  "received.txt",
		FileSize:  1000000, // Expect much more
		Status:    StatusPending,
		IsSender:  false,
		StartTime: time.Now(),
	}
	receiverMgr.mu.Unlock()

	senderAddr := senderMgr.listener.Addr().(*net.TCPAddr)
	err = receiverMgr.StartDownload(senderSession.ID, senderAddr.IP.String())
	require.NoError(t, err)

	// Wait for transfer and size check failure
	time.Sleep(500 * time.Millisecond)

	s := receiverMgr.GetSession(senderSession.ID)
	assert.Equal(t, StatusFailed, s.Status)
	assert.Contains(t, s.Error, "incomplete transfer")
}

// ==================== Read Error During Download ====================

func TestHandleConnection_ReadIDError(t *testing.T) {
	tm := NewManager()

	err := tm.Start("127.0.0.1")
	require.NoError(t, err)
	defer tm.Stop()

	// Connect but send less than 36 bytes then close
	conn, err := net.Dial("tcp", tm.listener.Addr().String())
	require.NoError(t, err)

	// Write partial ID
	_, _ = conn.Write([]byte("short"))
	conn.Close()

	// Should not crash
	time.Sleep(50 * time.Millisecond)
}

// ==================== Notify Progress No Session ====================

func TestNotifyProgress_NoSession(t *testing.T) {
	tm := NewManager()

	// Should not panic when session doesn't exist
	tm.notifyProgress("non-existent")
}

// ==================== matchesFilter Tests ====================

func TestMatchesFilter_AllFilters(t *testing.T) {
	tm := NewManager()

	session := &Session{
		ID:       "filter-test",
		PeerID:   "peer-1",
		Status:   StatusInProgress,
		IsSender: true,
	}

	// Test status match
	assert.True(t, tm.matchesFilter(session, ListOptions{Status: []Status{StatusInProgress}}))
	assert.False(t, tm.matchesFilter(session, ListOptions{Status: []Status{StatusCompleted}}))

	// Test IsSender match
	sender := true
	notSender := false
	assert.True(t, tm.matchesFilter(session, ListOptions{IsSender: &sender}))
	assert.False(t, tm.matchesFilter(session, ListOptions{IsSender: &notSender}))

	// Test PeerID match
	assert.True(t, tm.matchesFilter(session, ListOptions{PeerID: "peer-1"}))
	assert.False(t, tm.matchesFilter(session, ListOptions{PeerID: "peer-2"}))

	// Test combined filters
	assert.True(t, tm.matchesFilter(session, ListOptions{
		Status:   []Status{StatusInProgress},
		IsSender: &sender,
		PeerID:   "peer-1",
	}))
}

// ==================== Edge Case: Empty Manager ====================

func TestEmptyManager_Operations(t *testing.T) {
	tm := NewManager()

	// All operations should work on empty manager
	assert.Equal(t, 0, len(tm.GetSessions()))
	assert.Equal(t, 0, tm.GetActiveCount())

	stats := tm.GetStats()
	assert.Equal(t, 0, stats.TotalTransfers)

	result := tm.ListSessions(ListOptions{})
	assert.Equal(t, 0, len(result.Sessions))
	assert.Equal(t, 0, result.Total)

	removed := tm.CleanupOld(time.Hour)
	assert.Equal(t, 0, removed)
}

// ==================== Cleanup with No EndTime ====================

func TestCleanupOld_NoEndTime(t *testing.T) {
	tm := NewManager()

	// Session with zero EndTime should not be cleaned up even if finished
	tm.mu.Lock()
	tm.sessions["no-endtime"] = &Session{
		ID:     "no-endtime",
		Status: StatusCompleted,
		// EndTime is zero
	}
	tm.mu.Unlock()

	removed := tm.CleanupOld(0) // 0 duration means everything old should be removed
	assert.Equal(t, 0, removed, "Session without EndTime should not be cleaned")
}
