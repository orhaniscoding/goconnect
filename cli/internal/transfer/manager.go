package transfer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	TransferPort = 3001
	ChunkSize    = 32 * 1024 // 32KB chunks
)

// Manager handles file transfers
type Manager struct {
	sessions map[string]*Session // ID -> Session
	mu       sync.RWMutex

	listener net.Listener
	stopChan chan struct{}

	onProgress func(Session)
	onRequest  func(Request, string) // Request, SenderID

	pendingReqs map[string]pendingRequest // ID -> pendingRequest
	reqTimers   map[string]*time.Timer
	
	// Subscribers for real-time updates
	subscribers   map[chan Session]struct{}
	subscribersMu sync.RWMutex
}

type pendingRequest struct {
	Request  Request
	SenderID string
}

// NewManager creates a new transfer manager
func NewManager() *Manager {
	m := &Manager{
		sessions:    make(map[string]*Session),
		pendingReqs: make(map[string]pendingRequest),
		reqTimers:   make(map[string]*time.Timer),
		stopChan:    make(chan struct{}),
		subscribers: make(map[chan Session]struct{}),
	}
	return m
}

// Start starts the transfer listener
func (m *Manager) Start(bindIP string) error {
	addr := fmt.Sprintf("%s:%d", bindIP, TransferPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.listener = ln

	go m.acceptLoop()
	return nil
}

// Stop stops the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	if m.listener != nil {
		m.listener.Close()
	}

	m.mu.Lock()
	for _, timer := range m.reqTimers {
		timer.Stop()
	}
	m.mu.Unlock()
}

func (m *Manager) acceptLoop() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-m.stopChan:
				return
			default:
				continue
			}
		}
		go m.handleConnection(conn)
	}
}

// handleConnection handles incoming connections (Receiver connecting to Sender)
// Protocol:
// 1. Receiver sends TransferID (UUID string) + newline
// 2. Sender verifies ID
// 3. Sender streams file
func (m *Manager) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read Transfer ID
	buf := make([]byte, 36) // UUID length
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return
	}
	id := string(buf)

	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok || !session.IsSender {
		return // Unknown session or we are not the sender
	}

	// Start sending file
	m.sendFileData(conn, session)
}

func (m *Manager) sendFileData(conn net.Conn, session *Session) {
	file, err := os.Open(session.FilePath)
	if err != nil {
		m.failSession(session.ID, err.Error())
		return
	}
	defer file.Close()

	m.updateStatus(session.ID, StatusInProgress)

	buf := make([]byte, ChunkSize)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			m.failSession(session.ID, err.Error())
			return
		}

		_, wErr := conn.Write(buf[:n])
		if wErr != nil {
			m.failSession(session.ID, wErr.Error())
			return
		}

		m.mu.Lock()
		session.SentBytes += int64(n)
		m.mu.Unlock()
		m.notifyProgress(session.ID)
	}

	m.updateStatus(session.ID, StatusCompleted)
}

// CreateSendSession creates a new session for sending a file
func (m *Manager) CreateSendSession(peerID, filePath string) (*Session, error) {
	// Validate file path
	cleanPath := filepath.Clean(filePath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("cannot transfer directories")
	}

	id := uuid.New().String()
	session := &Session{
		ID:        id,
		PeerID:    peerID,
		FilePath:  cleanPath,
		FileName:  filepath.Base(cleanPath),
		FileSize:  info.Size(),
		Status:    StatusPending,
		IsSender:  true,
		StartTime: time.Now(),
	}

	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()

	return session, nil
}

// CreateReceiveSession creates a new session for receiving a file
func (m *Manager) CreateReceiveSession(req Request, peerID, savePath string) (*Session, error) {
	// Validate save path
	cleanPath := filepath.Clean(savePath)
	dir := filepath.Dir(cleanPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("destination directory does not exist")
	}

	// Remove from pending requests
	m.mu.Lock()
	delete(m.pendingReqs, req.ID)
	if timer, ok := m.reqTimers[req.ID]; ok {
		timer.Stop()
		delete(m.reqTimers, req.ID)
	}
	m.mu.Unlock()

	session := &Session{
		ID:        req.ID,
		PeerID:    peerID,
		FilePath:  cleanPath,
		FileName:  req.FileName,
		FileSize:  req.FileSize,
		Status:    StatusPending,
		IsSender:  false,
		StartTime: time.Now(),
	}

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session, nil
}

// StartDownload initiates the download by connecting to the sender
func (m *Manager) StartDownload(sessionID string, senderIP string) error {
	m.mu.RLock()
	session, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}

	go func() {
		addr := net.JoinHostPort(senderIP, fmt.Sprintf("%d", TransferPort))
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			m.failSession(sessionID, err.Error())
			return
		}
		defer conn.Close()

		// Send Transfer ID
		_, err = conn.Write([]byte(sessionID))
		if err != nil {
			m.failSession(sessionID, err.Error())
			return
		}

		// Create destination file
		out, err := os.Create(session.FilePath)
		if err != nil {
			m.failSession(sessionID, err.Error())
			return
		}
		defer out.Close()

		m.updateStatus(sessionID, StatusInProgress)

		// Read loop
		buf := make([]byte, ChunkSize)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				m.failSession(sessionID, err.Error())
				return
			}

			if _, wErr := out.Write(buf[:n]); wErr != nil {
				m.failSession(sessionID, wErr.Error())
				return
			}

			m.mu.Lock()
			session.SentBytes += int64(n)
			m.mu.Unlock()
			m.notifyProgress(sessionID)
		}

		// Verify size
		if session.SentBytes != session.FileSize {
			m.failSession(sessionID, "incomplete transfer")
			return
		}

		m.updateStatus(sessionID, StatusCompleted)
	}()

	return nil
}

func (m *Manager) updateStatus(id string, status Status) {
	m.mu.Lock()
	if s, ok := m.sessions[id]; ok {
		s.Status = status
		if status == StatusCompleted || status == StatusFailed {
			s.EndTime = time.Now()
		}
	}
	m.mu.Unlock()
	m.notifyProgress(id)
}

func (m *Manager) failSession(id, err string) {
	m.mu.Lock()
	if s, ok := m.sessions[id]; ok {
		s.Status = StatusFailed
		s.Error = err
		s.EndTime = time.Now()
	}
	m.mu.Unlock()
	m.notifyProgress(id)
}

func (m *Manager) notifyProgress(id string) {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if ok {
		// Notify callback
		if m.onProgress != nil {
			m.onProgress(*session)
		}
		// Notify subscribers
		m.notifySubscribers(session)
	}
}

// SetCallbacks sets the callbacks
func (m *Manager) SetCallbacks(onProgress func(Session), onRequest func(Request, string)) {
	m.onProgress = onProgress
	m.onRequest = onRequest
}

// GetSession returns a session by ID
func (m *Manager) GetSession(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// HandleSignalingMessage handles incoming control messages
func (m *Manager) HandleSignalingMessage(payload string, fromPeerID string) {
	var req Request
	if err := json.Unmarshal([]byte(payload), &req); err == nil {
		// Store pending request with timeout
		m.mu.Lock()
		m.pendingReqs[req.ID] = pendingRequest{
			Request:  req,
			SenderID: fromPeerID,
		}

		// Set 5 minute timeout
		timer := time.AfterFunc(5*time.Minute, func() {
			m.mu.Lock()
			delete(m.pendingReqs, req.ID)
			delete(m.reqTimers, req.ID)
			m.mu.Unlock()
		})
		m.reqTimers[req.ID] = timer
		m.mu.Unlock()

		if m.onRequest != nil {
			m.onRequest(req, fromPeerID)
		}
	}
}

// GetPendingRequest returns a pending request by ID
func (m *Manager) GetPendingRequest(id string) (Request, string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pr, ok := m.pendingReqs[id]
	return pr.Request, pr.SenderID, ok
}

// Subscribe returns a channel for receiving transfer progress updates
func (m *Manager) Subscribe() chan Session {
	ch := make(chan Session, 100)
	m.subscribersMu.Lock()
	m.subscribers[ch] = struct{}{}
	m.subscribersMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel
func (m *Manager) Unsubscribe(ch chan Session) {
	m.subscribersMu.Lock()
	delete(m.subscribers, ch)
	m.subscribersMu.Unlock()
	close(ch)
}

// GetSessions returns all active sessions
func (m *Manager) GetSessions() []Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	sessions := make([]Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, *s)
	}
	return sessions
}

// ListSessions returns sessions with filtering, pagination, and sorting
func (m *Manager) ListSessions(opts ListOptions) ListResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Collect all sessions that match filters
	var filtered []*Session
	activeCount := 0
	
	for _, s := range m.sessions {
		if s.IsActive() {
			activeCount++
		}
		
		// Apply filters
		if !m.matchesFilter(s, opts) {
			continue
		}
		filtered = append(filtered, s)
	}
	
	total := len(filtered)
	
	// Sort
	m.sortSessions(filtered, opts.SortBy, opts.SortOrder)
	
	// Apply pagination
	start := opts.Offset
	if start > len(filtered) {
		start = len(filtered)
	}
	
	end := len(filtered)
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}
	
	// Copy to result
	result := make([]Session, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, *filtered[i])
	}
	
	return ListResult{
		Sessions:    result,
		Total:       total,
		HasMore:     end < total,
		ActiveCount: activeCount,
	}
}

func (m *Manager) matchesFilter(s *Session, opts ListOptions) bool {
	// Filter by status
	if len(opts.Status) > 0 {
		found := false
		for _, st := range opts.Status {
			if s.Status == st {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Filter by direction
	if opts.IsSender != nil && s.IsSender != *opts.IsSender {
		return false
	}
	
	// Filter by peer
	if opts.PeerID != "" && s.PeerID != opts.PeerID {
		return false
	}
	
	return true
}

func (m *Manager) sortSessions(sessions []*Session, by SortField, order SortOrder) {
	if by == "" {
		by = SortByStartTime
	}
	if order == "" {
		order = SortDesc
	}
	
	// Simple bubble sort (fine for small lists, could use sort.Slice for larger)
	n := len(sessions)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if m.shouldSwap(sessions[j], sessions[j+1], by, order) {
				sessions[j], sessions[j+1] = sessions[j+1], sessions[j]
			}
		}
	}
}

func (m *Manager) shouldSwap(a, b *Session, by SortField, order SortOrder) bool {
	var less bool
	
	switch by {
	case SortByStartTime:
		less = a.StartTime.Before(b.StartTime)
	case SortByEndTime:
		less = a.EndTime.Before(b.EndTime)
	case SortByFileSize:
		less = a.FileSize < b.FileSize
	case SortByProgress:
		less = a.Progress() < b.Progress()
	case SortByFileName:
		less = a.FileName < b.FileName
	default:
		less = a.StartTime.Before(b.StartTime)
	}
	
	if order == SortDesc {
		return less
	}
	return !less
}

// GetStats returns transfer statistics
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var stats Stats
	var totalSpeed float64
	speedCount := 0
	
	for _, s := range m.sessions {
		stats.TotalTransfers++
		
		switch s.Status {
		case StatusPending, StatusInProgress:
			stats.ActiveTransfers++
		case StatusCompleted:
			stats.CompletedTransfers++
		case StatusFailed, StatusCancelled:
			stats.FailedTransfers++
		}
		
		if s.IsSender {
			stats.TotalBytesSent += s.SentBytes
		} else {
			stats.TotalBytesReceived += s.SentBytes
		}
		
		if s.Status == StatusCompleted && s.Elapsed() > 0 {
			totalSpeed += s.Speed()
			speedCount++
		}
	}
	
	if speedCount > 0 {
		stats.AverageSpeed = totalSpeed / float64(speedCount)
	}
	
	return stats
}

// CleanupOld removes finished sessions older than the specified duration
func (m *Manager) CleanupOld(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	removed := 0
	
	for id, s := range m.sessions {
		if s.IsFinished() && !s.EndTime.IsZero() && s.EndTime.Before(cutoff) {
			delete(m.sessions, id)
			removed++
		}
	}
	
	return removed
}

// GetActiveCount returns the number of active transfers
func (m *Manager) GetActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	count := 0
	for _, s := range m.sessions {
		if s.IsActive() {
			count++
		}
	}
	return count
}

func (m *Manager) notifySubscribers(session *Session) {
	m.subscribersMu.RLock()
	defer m.subscribersMu.RUnlock()
	
	for ch := range m.subscribers {
		select {
		case ch <- *session:
		default:
			// Channel full, skip
		}
	}
}

// RejectTransfer rejects a pending incoming transfer request
func (m *Manager) RejectTransfer(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from pending requests
	if _, ok := m.pendingReqs[id]; ok {
		delete(m.pendingReqs, id)
		if timer, ok := m.reqTimers[id]; ok {
			timer.Stop()
			delete(m.reqTimers, id)
		}
		return nil
	}

	// Check if it's an active session
	if session, ok := m.sessions[id]; ok {
		if session.Status == StatusPending {
			session.Status = StatusCancelled
			session.Error = "rejected by user"
			session.EndTime = time.Now()
			go m.notifyProgress(id)
			return nil
		}
		return fmt.Errorf("transfer already in progress or completed")
	}

	return fmt.Errorf("transfer not found: %s", id)
}

// CancelTransfer cancels an ongoing transfer
func (m *Manager) CancelTransfer(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("transfer not found: %s", id)
	}

	if session.Status == StatusCompleted || session.Status == StatusFailed || session.Status == StatusCancelled {
		return fmt.Errorf("transfer already finished")
	}

	session.Status = StatusCancelled
	session.Error = "cancelled by user"
	session.EndTime = time.Now()
	go m.notifyProgress(id)

	return nil
}
