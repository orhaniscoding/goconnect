package transfer

import "time"

// Status represents the state of a transfer
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// Session represents a file transfer session
type Session struct {
	ID        string    `json:"id"`
	PeerID    string    `json:"peer_id"`
	FilePath  string    `json:"file_path"` // Local path (source or dest)
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	SentBytes int64     `json:"sent_bytes"`
	Status    Status    `json:"status"`
	IsSender  bool      `json:"is_sender"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error,omitempty"`
}

// Progress returns the transfer progress as a percentage (0-100)
func (s *Session) Progress() float64 {
	if s.FileSize == 0 {
		return 0
	}
	return float64(s.SentBytes) / float64(s.FileSize) * 100
}

// Speed returns the current transfer speed in bytes per second
func (s *Session) Speed() float64 {
	elapsed := s.Elapsed()
	if elapsed == 0 {
		return 0
	}
	return float64(s.SentBytes) / elapsed.Seconds()
}

// Elapsed returns the duration of the transfer
func (s *Session) Elapsed() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

// ETA returns the estimated time to completion
func (s *Session) ETA() time.Duration {
	speed := s.Speed()
	if speed == 0 {
		return 0
	}
	remaining := s.FileSize - s.SentBytes
	return time.Duration(float64(remaining) / speed * float64(time.Second))
}

// IsActive returns true if the transfer is pending or in progress
func (s *Session) IsActive() bool {
	return s.Status == StatusPending || s.Status == StatusInProgress
}

// IsFinished returns true if the transfer is completed, failed, or cancelled
func (s *Session) IsFinished() bool {
	return s.Status == StatusCompleted || s.Status == StatusFailed || s.Status == StatusCancelled
}

// Request represents a file transfer request sent over the control channel
type Request struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

// ListOptions configures how sessions are listed
type ListOptions struct {
	// Filter options
	Status   []Status // Filter by status (empty = all)
	IsSender *bool    // Filter by direction (nil = both)
	PeerID   string   // Filter by peer ID (empty = all)

	// Pagination
	Limit  int // Max results (0 = no limit)
	Offset int // Skip first N results

	// Sorting
	SortBy    SortField // Field to sort by
	SortOrder SortOrder // Ascending or descending
}

// SortField specifies which field to sort by
type SortField string

const (
	SortByStartTime SortField = "start_time"
	SortByEndTime   SortField = "end_time"
	SortByFileSize  SortField = "file_size"
	SortByProgress  SortField = "progress"
	SortByFileName  SortField = "file_name"
)

// SortOrder specifies sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// ListResult contains paginated session results
type ListResult struct {
	Sessions    []Session `json:"sessions"`
	Total       int       `json:"total"`        // Total matching sessions
	HasMore     bool      `json:"has_more"`     // More results available
	ActiveCount int       `json:"active_count"` // Number of active transfers
}

// Stats contains transfer statistics
type Stats struct {
	TotalTransfers     int     `json:"total_transfers"`
	ActiveTransfers    int     `json:"active_transfers"`
	CompletedTransfers int     `json:"completed_transfers"`
	FailedTransfers    int     `json:"failed_transfers"`
	TotalBytesSent     int64   `json:"total_bytes_sent"`
	TotalBytesReceived int64   `json:"total_bytes_received"`
	AverageSpeed       float64 `json:"average_speed"`
}
