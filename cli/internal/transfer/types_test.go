package transfer

import (
	"testing"
	"time"
)

// ==================== Status Tests ====================

func TestStatusConstants(t *testing.T) {
	t.Run("Status Values Are Correct", func(t *testing.T) {
		if StatusPending != "pending" {
			t.Errorf("Expected 'pending', got %s", StatusPending)
		}
		if StatusInProgress != "in_progress" {
			t.Errorf("Expected 'in_progress', got %s", StatusInProgress)
		}
		if StatusCompleted != "completed" {
			t.Errorf("Expected 'completed', got %s", StatusCompleted)
		}
		if StatusFailed != "failed" {
			t.Errorf("Expected 'failed', got %s", StatusFailed)
		}
		if StatusCancelled != "cancelled" {
			t.Errorf("Expected 'cancelled', got %s", StatusCancelled)
		}
	})
}

// ==================== Session.Speed Tests ====================

func TestSession_Speed(t *testing.T) {
	t.Run("Returns Zero When No Time Elapsed", func(t *testing.T) {
		now := time.Now()
		s := &Session{
			StartTime: now,
			EndTime:   now,
			SentBytes: 1000,
		}
		if s.Speed() != 0 {
			t.Errorf("Expected speed 0, got %f", s.Speed())
		}
	})

	t.Run("Calculates Speed With EndTime", func(t *testing.T) {
		start := time.Now().Add(-10 * time.Second)
		end := time.Now()
		s := &Session{
			StartTime: start,
			EndTime:   end,
			SentBytes: 10000,
		}
		speed := s.Speed()
		// Should be approximately 1000 bytes/sec (10000 / 10 seconds)
		if speed < 900 || speed > 1100 {
			t.Errorf("Expected speed ~1000, got %f", speed)
		}
	})

	t.Run("Calculates Speed For Ongoing Transfer", func(t *testing.T) {
		start := time.Now().Add(-5 * time.Second)
		s := &Session{
			StartTime: start,
			SentBytes: 5000,
		}
		speed := s.Speed()
		// Should be approximately 1000 bytes/sec
		if speed < 800 || speed > 1200 {
			t.Errorf("Expected speed ~1000, got %f", speed)
		}
	})
}

// ==================== Session.Elapsed Tests ====================

func TestSession_Elapsed(t *testing.T) {
	t.Run("Returns Duration For Completed Transfer", func(t *testing.T) {
		start := time.Now().Add(-30 * time.Second)
		end := time.Now()
		s := &Session{
			StartTime: start,
			EndTime:   end,
		}
		elapsed := s.Elapsed()
		if elapsed < 29*time.Second || elapsed > 31*time.Second {
			t.Errorf("Expected ~30s elapsed, got %v", elapsed)
		}
	})

	t.Run("Returns Duration For Ongoing Transfer", func(t *testing.T) {
		start := time.Now().Add(-5 * time.Second)
		s := &Session{
			StartTime: start,
		}
		elapsed := s.Elapsed()
		if elapsed < 4*time.Second || elapsed > 6*time.Second {
			t.Errorf("Expected ~5s elapsed, got %v", elapsed)
		}
	})
}

// ==================== Session.ETA Tests ====================

func TestSession_ETA(t *testing.T) {
	t.Run("Returns Zero When Speed Is Zero", func(t *testing.T) {
		now := time.Now()
		s := &Session{
			StartTime: now,
			EndTime:   now, // Zero elapsed time = zero speed
			SentBytes: 1000,
			FileSize:  10000,
		}
		eta := s.ETA()
		if eta != 0 {
			t.Errorf("Expected ETA 0, got %v", eta)
		}
	})

	t.Run("Calculates ETA For Active Transfer", func(t *testing.T) {
		start := time.Now().Add(-10 * time.Second)
		s := &Session{
			StartTime: start,
			SentBytes: 5000,  // 50% done in 10 seconds
			FileSize:  10000, // Need 5000 more bytes
		}
		eta := s.ETA()
		// At 500 bytes/sec, 5000 bytes should take ~10 seconds
		if eta < 8*time.Second || eta > 12*time.Second {
			t.Errorf("Expected ETA ~10s, got %v", eta)
		}
	})

	t.Run("Returns Zero When Transfer Complete", func(t *testing.T) {
		s := &Session{
			StartTime: time.Now().Add(-10 * time.Second),
			SentBytes: 10000,
			FileSize:  10000, // All bytes sent
		}
		eta := s.ETA()
		if eta != 0 {
			t.Errorf("Expected ETA 0 for completed transfer, got %v", eta)
		}
	})
}

// ==================== SortField Tests ====================

func TestSortFieldConstants(t *testing.T) {
	t.Run("SortField Values Are Correct", func(t *testing.T) {
		if SortByStartTime != "start_time" {
			t.Errorf("Expected 'start_time', got %s", SortByStartTime)
		}
		if SortByEndTime != "end_time" {
			t.Errorf("Expected 'end_time', got %s", SortByEndTime)
		}
		if SortByFileSize != "file_size" {
			t.Errorf("Expected 'file_size', got %s", SortByFileSize)
		}
		if SortByProgress != "progress" {
			t.Errorf("Expected 'progress', got %s", SortByProgress)
		}
		if SortByFileName != "file_name" {
			t.Errorf("Expected 'file_name', got %s", SortByFileName)
		}
	})
}

// ==================== SortOrder Tests ====================

func TestSortOrderConstants(t *testing.T) {
	t.Run("SortOrder Values Are Correct", func(t *testing.T) {
		if SortAsc != "asc" {
			t.Errorf("Expected 'asc', got %s", SortAsc)
		}
		if SortDesc != "desc" {
			t.Errorf("Expected 'desc', got %s", SortDesc)
		}
	})
}

// ==================== Request Tests ====================

func TestRequest_Fields(t *testing.T) {
	t.Run("Request Has Required Fields", func(t *testing.T) {
		req := Request{
			ID:       "test-id",
			FileName: "test.txt",
			FileSize: 12345,
		}
		if req.ID != "test-id" {
			t.Errorf("Expected ID 'test-id', got %s", req.ID)
		}
		if req.FileName != "test.txt" {
			t.Errorf("Expected FileName 'test.txt', got %s", req.FileName)
		}
		if req.FileSize != 12345 {
			t.Errorf("Expected FileSize 12345, got %d", req.FileSize)
		}
	})
}

// ==================== ListOptions Tests ====================

func TestListOptions_Defaults(t *testing.T) {
	t.Run("Default Options Are Zero Values", func(t *testing.T) {
		opts := ListOptions{}
		if opts.Limit != 0 {
			t.Errorf("Expected default Limit 0, got %d", opts.Limit)
		}
		if opts.Offset != 0 {
			t.Errorf("Expected default Offset 0, got %d", opts.Offset)
		}
		if opts.PeerID != "" {
			t.Errorf("Expected default PeerID '', got %s", opts.PeerID)
		}
		if opts.IsSender != nil {
			t.Errorf("Expected default IsSender nil, got %v", opts.IsSender)
		}
	})

	t.Run("Status Filter Can Be Set", func(t *testing.T) {
		opts := ListOptions{
			Status: []Status{StatusPending, StatusInProgress},
		}
		if len(opts.Status) != 2 {
			t.Errorf("Expected 2 status filters, got %d", len(opts.Status))
		}
	})

	t.Run("IsSender Filter Can Be Set", func(t *testing.T) {
		isSender := true
		opts := ListOptions{
			IsSender: &isSender,
		}
		if opts.IsSender == nil || *opts.IsSender != true {
			t.Error("Expected IsSender to be true")
		}
	})
}

// ==================== ListResult Tests ====================

func TestListResult_Fields(t *testing.T) {
	t.Run("ListResult Contains All Fields", func(t *testing.T) {
		result := ListResult{
			Sessions:    []Session{{ID: "1"}, {ID: "2"}},
			Total:       10,
			HasMore:     true,
			ActiveCount: 2,
		}
		if len(result.Sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(result.Sessions))
		}
		if result.Total != 10 {
			t.Errorf("Expected Total 10, got %d", result.Total)
		}
		if !result.HasMore {
			t.Error("Expected HasMore to be true")
		}
		if result.ActiveCount != 2 {
			t.Errorf("Expected ActiveCount 2, got %d", result.ActiveCount)
		}
	})
}

// ==================== Stats Tests ====================

func TestStats_Fields(t *testing.T) {
	t.Run("Stats Contains All Fields", func(t *testing.T) {
		stats := Stats{
			TotalTransfers:     100,
			ActiveTransfers:    5,
			CompletedTransfers: 90,
			FailedTransfers:    5,
			TotalBytesSent:     1000000,
			TotalBytesReceived: 500000,
			AverageSpeed:       50000.5,
		}
		if stats.TotalTransfers != 100 {
			t.Errorf("Expected TotalTransfers 100, got %d", stats.TotalTransfers)
		}
		if stats.ActiveTransfers != 5 {
			t.Errorf("Expected ActiveTransfers 5, got %d", stats.ActiveTransfers)
		}
		if stats.CompletedTransfers != 90 {
			t.Errorf("Expected CompletedTransfers 90, got %d", stats.CompletedTransfers)
		}
		if stats.FailedTransfers != 5 {
			t.Errorf("Expected FailedTransfers 5, got %d", stats.FailedTransfers)
		}
		if stats.TotalBytesSent != 1000000 {
			t.Errorf("Expected TotalBytesSent 1000000, got %d", stats.TotalBytesSent)
		}
		if stats.TotalBytesReceived != 500000 {
			t.Errorf("Expected TotalBytesReceived 500000, got %d", stats.TotalBytesReceived)
		}
		if stats.AverageSpeed != 50000.5 {
			t.Errorf("Expected AverageSpeed 50000.5, got %f", stats.AverageSpeed)
		}
	})
}

// ==================== Session Struct Tests ====================

func TestSession_AllFields(t *testing.T) {
	t.Run("Session Has All Required Fields", func(t *testing.T) {
		now := time.Now()
		s := Session{
			ID:        "test-session",
			PeerID:    "peer-1",
			FilePath:  "/path/to/file.txt",
			FileName:  "file.txt",
			FileSize:  10000,
			SentBytes: 5000,
			Status:    StatusInProgress,
			IsSender:  true,
			StartTime: now,
			EndTime:   now.Add(10 * time.Second),
			Error:     "",
		}

		if s.ID != "test-session" {
			t.Errorf("Expected ID 'test-session', got %s", s.ID)
		}
		if s.PeerID != "peer-1" {
			t.Errorf("Expected PeerID 'peer-1', got %s", s.PeerID)
		}
		if s.FileName != "file.txt" {
			t.Errorf("Expected FileName 'file.txt', got %s", s.FileName)
		}
		if !s.IsSender {
			t.Error("Expected IsSender to be true")
		}
	})
}

// ==================== Constants Tests ====================

func TestTransferConstants(t *testing.T) {
	t.Run("TransferPort Is Set", func(t *testing.T) {
		if TransferPort != 3001 {
			t.Errorf("Expected TransferPort 3001, got %d", TransferPort)
		}
	})

	t.Run("ChunkSize Is Set", func(t *testing.T) {
		if ChunkSize != 32*1024 {
			t.Errorf("Expected ChunkSize 32768, got %d", ChunkSize)
		}
	})
}
