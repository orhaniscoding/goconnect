package voice

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/logger"
)

// Manager handles voice signaling operations
type Manager struct {
	listener  net.Listener
	onSignal  func(sig Signal)
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex

	// Subscribers for real-time updates
	subscribers   map[chan Signal]struct{}
	subscribersMu sync.RWMutex
}

// NewManager creates a new Voice Manager
func NewManager() *Manager {
	return &Manager{
		stopChan:    make(chan struct{}),
		subscribers: make(map[chan Signal]struct{}),
	}
}

// Start starts the TCP listener
func (m *Manager) Start(ip string, port int) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start voice signal listener: %w", err)
	}
	m.listener = l
	logger.Info("Voice signal listener started", "addr", addr)

	m.wg.Add(1)
	go m.acceptLoop()

	return nil
}

// Stop stops the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	if m.listener != nil {
		m.listener.Close()
	}
	m.wg.Wait()
}

// OnSignal sets the callback for incoming signals
func (m *Manager) OnSignal(f func(sig Signal)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onSignal = f
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
				logger.Error("Voice accept error", "error", err)
				continue
			}
		}

		go m.handleConnection(conn)
	}
}

func (m *Manager) handleConnection(conn net.Conn) {
	defer conn.Close()

	var sig Signal
	if err := json.NewDecoder(conn).Decode(&sig); err != nil {
		logger.Error("Failed to decode voice signal", "error", err)
		return
	}

	m.mu.Lock()
	callback := m.onSignal
	m.mu.Unlock()

	if callback != nil {
		callback(sig)
	}

	// Notify subscribers
	m.notifySubscribers(sig)
}

// Subscribe returns a channel for receiving real-time signals
func (m *Manager) Subscribe() chan Signal {
	ch := make(chan Signal, 100)
	m.subscribersMu.Lock()
	m.subscribers[ch] = struct{}{}
	m.subscribersMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel
func (m *Manager) Unsubscribe(ch chan Signal) {
	m.subscribersMu.Lock()
	delete(m.subscribers, ch)
	m.subscribersMu.Unlock()
	close(ch)
}

func (m *Manager) notifySubscribers(sig Signal) {
	m.subscribersMu.RLock()
	defer m.subscribersMu.RUnlock()

	for ch := range m.subscribers {
		select {
		case ch <- sig:
		default:
			// Channel full, skip
		}
	}
}

// SendSignal sends a signal to a peer at the given IP
func (m *Manager) SendSignal(peerIP string, port int, sig Signal) error {
	addr := net.JoinHostPort(peerIP, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial peer for voice: %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(sig); err != nil {
		return fmt.Errorf("failed to encode voice signal: %w", err)
	}

	return nil
}
