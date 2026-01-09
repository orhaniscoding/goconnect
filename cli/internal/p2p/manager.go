package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/cli/internal/logger"
	"github.com/pion/ice/v2"
)

// SignalService defines the interface for signaling
type SignalService interface {
	SendOffer(targetID string, ufrag, pwd string) error
	SendAnswer(targetID string, ufrag, pwd string) error
	SendCandidate(targetID string, candidate string) error
	OnOffer(func(sourceID, ufrag, pwd string))
	OnAnswer(func(sourceID, ufrag, pwd string))
	OnCandidate(func(sourceID, candidate string))
}

// Manager manages P2P connections
type Manager struct {
	signal SignalService
	agents map[string]*Agent
	mu     sync.RWMutex

	// Pending answers for outbound connections
	pendingAnswers map[string]chan answerData
	answersMu      sync.Mutex

	// Callback for new connections
	onNewConnection func(peerID string, conn *ice.Conn)

	// Latency storage
	latencies   map[string]int64
	latenciesMu sync.RWMutex

	// ICE configuration (legacy: stunURL)
	stunURL   string
	iceConfig *ICEConfig
}

type answerData struct {
	ufrag string
	pwd   string
}

// NewManager creates a new P2P manager
func NewManager(signal SignalService, stunURL string) *Manager {
	return &Manager{
		signal:         signal,
		agents:         make(map[string]*Agent),
		pendingAnswers: make(map[string]chan answerData),
		latencies:      make(map[string]int64),
		stunURL:        stunURL,
	}
}

// SetNewConnectionCallback sets the callback for established connections
func (m *Manager) SetNewConnectionCallback(f func(peerID string, conn *ice.Conn)) {
	m.onNewConnection = f
}

// Start registers signal handlers
func (m *Manager) Start() {
	m.signal.OnOffer(m.handleOffer)
	m.signal.OnAnswer(m.handleAnswer)
	m.signal.OnCandidate(m.handleCandidate)
}

func (m *Manager) handleOffer(sourceID, ufrag, pwd string) {
	logger.Info("Received offer", "sourceID", sourceID)
	// Handle incoming offer in a separate goroutine to not block the signal loop
	go func() {
		conn, err := m.Accept(context.Background(), sourceID, ufrag, pwd)
		if err != nil {
			logger.Error("Failed to accept connection", "source", sourceID, "error", err)
			return
		}
		if m.onNewConnection != nil {
			m.onNewConnection(sourceID, conn)
		}
	}()
}

func (m *Manager) handleAnswer(sourceID, ufrag, pwd string) {
	logger.Info("Received answer", "sourceID", sourceID)
	m.answersMu.Lock()
	ch, ok := m.pendingAnswers[sourceID]
	m.answersMu.Unlock()

	if ok {
		ch <- answerData{ufrag, pwd}
	} else {
		logger.Warn("No pending connection for answer", "sourceID", sourceID)
	}
}

func (m *Manager) handleCandidate(sourceID, candidate string) {
	logger.Debug("Received candidate", "sourceID", sourceID, "candidate", candidate)
	m.mu.RLock()
	agent, ok := m.agents[sourceID]
	m.mu.RUnlock()

	if ok {
		c, err := ice.UnmarshalCandidate(candidate)
		if err != nil {
			logger.Error("Failed to unmarshal candidate", "source", sourceID, "error", err)
			return
		}
		if err := agent.AddRemoteCandidate(c); err != nil {
			logger.Error("Failed to add remote candidate", "source", sourceID, "error", err)
		}
	} else {
		logger.Warn("No agent found for candidate", "sourceID", sourceID)
	}
}

// Connect initiates a P2P connection to a peer
func (m *Manager) Connect(ctx context.Context, peerID string) (*ice.Conn, error) {
	m.mu.Lock()
	if _, exists := m.agents[peerID]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("connection to %s already exists", peerID)
	}

	agent, err := NewAgent(m.stunURL)
	if err != nil {
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	m.agents[peerID] = agent
	m.monitorConnectionState(peerID, agent)
	m.mu.Unlock()

	// Setup answer channel
	answerCh := make(chan answerData, 1)
	m.answersMu.Lock()
	m.pendingAnswers[peerID] = answerCh
	m.answersMu.Unlock()

	defer func() {
		m.answersMu.Lock()
		delete(m.pendingAnswers, peerID)
		m.answersMu.Unlock()
	}()

	// Get local credentials
	ufrag, pwd, err := agent.GetLocalCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get local credentials: %w", err)
	}

	// Start gathering candidates
	_ = agent.OnCandidate(func(c ice.Candidate) {
		if c != nil {
			if err := m.signal.SendCandidate(peerID, c.Marshal()); err != nil {
				logger.Error("Failed to send candidate", "peer", peerID, "error", err)
			}
		}
	})

	if err := agent.GatherCandidates(); err != nil {
		return nil, fmt.Errorf("failed to gather candidates: %w", err)
	}

	// Send offer
	if err := m.signal.SendOffer(peerID, ufrag, pwd); err != nil {
		return nil, fmt.Errorf("failed to send offer: %w", err)
	}

	// Wait for answer
	select {
	case answer := <-answerCh:
		// Received answer, now Dial
		conn, err := agent.Dial(ctx, answer.ufrag, answer.pwd)
		if err == nil {
			go m.startMetricsLoop(peerID, conn)
		}
		return conn, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Accept accepts an incoming P2P connection from a peer
func (m *Manager) Accept(ctx context.Context, peerID, ufrag, pwd string) (*ice.Conn, error) {
	m.mu.Lock()
	if _, exists := m.agents[peerID]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("connection to %s already exists", peerID)
	}

	agent, err := NewAgent(m.stunURL)
	if err != nil {
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	m.agents[peerID] = agent
	m.monitorConnectionState(peerID, agent)
	m.mu.Unlock()

	localUfrag, localPwd, err := agent.GetLocalCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get local credentials: %w", err)
	}

	// Handle candidates
	_ = agent.OnCandidate(func(c ice.Candidate) {
		if c != nil {
			if err := m.signal.SendCandidate(peerID, c.Marshal()); err != nil {
				logger.Error("Failed to send candidate", "peer", peerID, "error", err)
			}
		}
	})

	// Start gathering candidates
	if err := agent.GatherCandidates(); err != nil {
		return nil, fmt.Errorf("failed to gather candidates: %w", err)
	}

	// Send answer
	if err := m.signal.SendAnswer(peerID, localUfrag, localPwd); err != nil {
		return nil, fmt.Errorf("failed to send answer: %w", err)
	}

	// Accept connection
	conn, err := agent.Accept(ctx, ufrag, pwd)
	if err == nil {
		go m.startMetricsLoop(peerID, conn)
	}
	return conn, err
}

// IsConnected checks if a connection to the peer exists
func (m *Manager) IsConnected(peerID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.agents[peerID]
	return exists
}

// RemovePeer removes a peer from the manager
func (m *Manager) RemovePeer(peerID string) {
	m.mu.Lock()
	if agent, exists := m.agents[peerID]; exists {
		_ = agent.Close()
		delete(m.agents, peerID)
	}
	m.mu.Unlock()

	m.latenciesMu.Lock()
	delete(m.latencies, peerID)
	m.latenciesMu.Unlock()
}

func (m *Manager) monitorConnectionState(peerID string, agent *Agent) {
	_ = agent.OnConnectionStateChange(func(state ice.ConnectionState) {
		logger.Info("Connection state changed", "peer", peerID, "state", state.String())
		if state == ice.ConnectionStateFailed || state == ice.ConnectionStateClosed {
			m.RemovePeer(peerID)
			// Trigger auto-reconnect in a separate goroutine
			go m.reconnect(peerID)
		}
	})
}

func (m *Manager) reconnect(peerID string) {
	backoff := 1 * time.Second
	maxBackoff := 60 * time.Second

	for {
		logger.Info("Attempting to reconnect", "peer", peerID, "backoff", backoff)
		time.Sleep(backoff)

		// Check if already connected (race condition check)
		if m.IsConnected(peerID) {
			logger.Info("Already reconnected, stopping retry loop", "peer", peerID)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := m.Connect(ctx, peerID)
		cancel()

		if err == nil {
			logger.Info("Successfully reconnected", "peer", peerID)
			return
		}

		logger.Error("Reconnection failed", "peer", peerID, "error", err)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// PeerStatus represents the status of a P2P connection
type PeerStatus struct {
	Connected       bool   `json:"connected"`
	ConnectionState string `json:"connection_state"`
	LocalCandidate  string `json:"local_candidate,omitempty"`
	RemoteCandidate string `json:"remote_candidate,omitempty"`
	LatencyMs       int64  `json:"latency_ms"`
	IsRelay         bool   `json:"is_relay"`
}

// GetPeerStatus returns the status of a peer connection
func (m *Manager) GetPeerStatus(peerID string) PeerStatus {
	m.mu.RLock()
	agent, exists := m.agents[peerID]
	m.mu.RUnlock()

	if !exists {
		return PeerStatus{Connected: false, ConnectionState: "disconnected"}
	}

	m.latenciesMu.RLock()
	latency := m.latencies[peerID]
	m.latenciesMu.RUnlock()

	// Get actual ICE connection state
	state := agent.ConnectionState()
	isRelay := agent.IsRelay()

	return PeerStatus{
		Connected:       state == ice.ConnectionStateConnected || state == ice.ConnectionStateCompleted,
		ConnectionState: state.String(),
		LatencyMs:       latency,
		IsRelay:         isRelay,
	}
}

const (
	msgPing = 0x01
	msgPong = 0x02
)

func (m *Manager) startMetricsLoop(peerID string, conn *ice.Conn) {
	// Start reader loop
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				logger.Debug("Metrics read loop stopped", "peer", peerID, "error", err)
				return
			}
			if n < 1 {
				continue
			}

			switch buf[0] {
			case msgPing:
				if n >= 9 {
					// Echo back as pong
					out := make([]byte, 9)
					out[0] = msgPong
					copy(out[1:], buf[1:9])
					if _, err := conn.Write(out); err != nil {
						logger.Error("Failed to send pong", "peer", peerID, "error", err)
					}
				}
			case msgPong:
				if n >= 9 {
					ts := int64(0)
					// Manual binary decoding to avoid importing binary package if not needed,
					// but binary is safer. Let's assume we import binary.
					// Actually, let's just use a simple shift for 8 bytes
					for i := 0; i < 8; i++ {
						ts |= int64(buf[1+i]) << (8 * i)
					}

					now := time.Now().UnixNano()
					rtt := now - ts
					if rtt > 0 {
						m.latenciesMu.Lock()
						m.latencies[peerID] = rtt / 1e6 // Convert to ms
						m.latenciesMu.Unlock()
					}
				}
			}
		}
	}()

	// Start pinger loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			buf := make([]byte, 9)
			buf[0] = msgPing
			for i := 0; i < 8; i++ {
				buf[1+i] = byte(now >> (8 * i))
			}
			if _, err := conn.Write(buf); err != nil {
				logger.Error("Failed to send ping", "peer", peerID, "error", err)
				return
			}
		}
	}
}
