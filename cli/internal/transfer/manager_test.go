package transfer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	tmpfile.Write(content)
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
	tmpfile.Write([]byte("test"))
	tmpfile.Close()

	// Initially empty
	sessions := tm.GetSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}

	// Create a session
	tm.CreateSendSession("peer1", tmpfile.Name())

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
		func(s Session) { progressCalled = true },
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
