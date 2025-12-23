package chat

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/cli/internal/logger"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"` // Peer ID or IP
	Content   string    `json:"content"`
	Time      time.Time `json:"time"`
	NetworkID string    `json:"network_id,omitempty"`
}

// Manager handles chat operations
type Manager struct {
	listener  net.Listener
	onMessage func(msg Message)
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex

	// Message history (in-memory cache)
	messages    []Message
	messagesMu  sync.RWMutex
	maxMessages int

	// Persistent storage (optional)
	storage *Storage

	// Subscribers for real-time updates
	subscribers   map[chan Message]struct{}
	subscribersMu sync.RWMutex
}

// NewManager creates a new Chat Manager
func NewManager() *Manager {
	return &Manager{
		stopChan:    make(chan struct{}),
		messages:    make([]Message, 0, 100),
		maxMessages: 1000, // Keep last 1000 messages in memory
		subscribers: make(map[chan Message]struct{}),
	}
}

// NewManagerWithStorage creates a new Chat Manager with persistent storage
func NewManagerWithStorage(dataDir string) (*Manager, error) {
	storage, err := NewStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat storage: %w", err)
	}

	m := &Manager{
		stopChan:    make(chan struct{}),
		messages:    make([]Message, 0, 100),
		maxMessages: 1000,
		subscribers: make(map[chan Message]struct{}),
		storage:     storage,
	}

	// Load recent messages into memory cache
	if err := m.loadRecentMessages(); err != nil {
		logger.Warn("Failed to load recent messages", "error", err)
	}

	return m, nil
}

// loadRecentMessages loads recent messages from storage into memory cache
func (m *Manager) loadRecentMessages() error {
	if m.storage == nil {
		return nil
	}

	messages, err := m.storage.GetMessages("", m.maxMessages, "")
	if err != nil {
		return err
	}

	m.messagesMu.Lock()
	defer m.messagesMu.Unlock()

	// Messages are returned newest first, reverse for chronological order
	for i := len(messages) - 1; i >= 0; i-- {
		m.messages = append(m.messages, messages[i])
	}

	return nil
}

// Start starts the TCP listener on the given IP and port
func (m *Manager) Start(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start chat listener: %w", err)
	}
	m.listener = l
	logger.Info("Chat listener started", "addr", addr)

	m.wg.Add(1)
	go m.acceptLoop()

	return nil
}

// Stop stops the chat manager
func (m *Manager) Stop() {
	close(m.stopChan)
	if m.listener != nil {
		m.listener.Close()
	}
	if m.storage != nil {
		m.storage.Close()
	}
	m.wg.Wait()
}

// OnMessage sets the callback for incoming messages
func (m *Manager) OnMessage(f func(msg Message)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onMessage = f
}

func (m *Manager) acceptLoop() {
	defer m.wg.Done()

	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-m.stopChan:
				return
			default:
				logger.Error("Chat accept error", "error", err)
				continue
			}
		}

		go m.handleConnection(conn)
	}
}

func (m *Manager) handleConnection(conn net.Conn) {
	defer conn.Close()

	var msg Message
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		logger.Error("Failed to decode chat message", "error", err)
		return
	}

	// Override time with local reception time
	msg.Time = time.Now()

	m.mu.Lock()
	callback := m.onMessage
	m.mu.Unlock()

	if callback != nil {
		callback(msg)
	}

	// Store message in history
	m.storeMessage(msg)

	// Notify subscribers
	m.notifySubscribers(msg)
}

// Subscribe returns a channel for receiving real-time messages
func (m *Manager) Subscribe() chan Message {
	ch := make(chan Message, 100)
	m.subscribersMu.Lock()
	m.subscribers[ch] = struct{}{}
	m.subscribersMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel
func (m *Manager) Unsubscribe(ch chan Message) {
	m.subscribersMu.Lock()
	delete(m.subscribers, ch)
	m.subscribersMu.Unlock()
	close(ch)
}

// GetMessages returns message history with optional filtering
func (m *Manager) GetMessages(networkID string, limit int, beforeID string) []Message {
	// Use persistent storage if available for better pagination
	if m.storage != nil {
		messages, err := m.storage.GetMessages(networkID, limit, beforeID)
		if err != nil {
			logger.Warn("Failed to get messages from storage", "error", err)
			// Fall through to memory cache
		} else {
			return messages
		}
	}

	// Fall back to in-memory cache
	m.messagesMu.RLock()
	defer m.messagesMu.RUnlock()

	if limit <= 0 || limit > len(m.messages) {
		limit = len(m.messages)
	}

	result := make([]Message, 0, limit)
	foundBefore := beforeID == ""

	// Iterate in reverse to get most recent first
	for i := len(m.messages) - 1; i >= 0 && len(result) < limit; i-- {
		msg := m.messages[i]

		// Skip until we find the beforeID
		if !foundBefore {
			if msg.ID == beforeID {
				foundBefore = true
			}
			continue
		}

		// Filter by network if specified
		if networkID != "" && msg.NetworkID != networkID {
			continue
		}

		result = append(result, msg)
	}

	return result
}

// SearchMessages searches for messages containing the query string
func (m *Manager) SearchMessages(query string, limit int) []Message {
	if m.storage != nil {
		messages, err := m.storage.SearchMessages(query, limit)
		if err != nil {
			logger.Warn("Failed to search messages", "error", err)
			return nil
		}
		return messages
	}
	return nil
}

// GetStorage returns the underlying storage (for advanced operations)
func (m *Manager) GetStorage() *Storage {
	return m.storage
}

func (m *Manager) storeMessage(msg Message) {
	m.messagesMu.Lock()
	defer m.messagesMu.Unlock()

	// Generate ID if not set
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("%d-%s", msg.Time.UnixNano(), msg.From)
	}

	// Store in memory cache
	m.messages = append(m.messages, msg)

	// Trim if over limit
	if len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}

	// Persist to storage if available
	if m.storage != nil {
		if err := m.storage.SaveMessage(msg); err != nil {
			logger.Warn("Failed to persist message", "error", err)
		}
	}
}

func (m *Manager) notifySubscribers(msg Message) {
	m.subscribersMu.RLock()
	defer m.subscribersMu.RUnlock()

	for ch := range m.subscribers {
		select {
		case ch <- msg:
		default:
			// Channel full, skip
		}
	}
}

// SendMessage sends a message to a peer at the given IP
func (m *Manager) SendMessage(peerIP string, port int, content string, fromID string) error {
	addr := net.JoinHostPort(peerIP, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial peer: %w", err)
	}
	defer conn.Close()

	msg := Message{
		From:    fromID,
		Content: content,
		Time:    time.Now(),
	}

	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	return nil
}
