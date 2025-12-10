package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pion/ice/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSignalService is a mock implementation of SignalService
type MockSignalService struct {
	mock.Mock
}

func (m *MockSignalService) SendOffer(targetID, ufrag, pwd string) error {
	args := m.Called(targetID, ufrag, pwd)
	return args.Error(0)
}

func (m *MockSignalService) SendAnswer(targetID, ufrag, pwd string) error {
	args := m.Called(targetID, ufrag, pwd)
	return args.Error(0)
}

func (m *MockSignalService) SendCandidate(targetID, candidate string) error {
	args := m.Called(targetID, candidate)
	return args.Error(0)
}

func (m *MockSignalService) OnOffer(f func(sourceID, ufrag, pwd string)) {
	m.Called(f)
}

func (m *MockSignalService) OnAnswer(f func(sourceID, ufrag, pwd string)) {
	m.Called(f)
}

func (m *MockSignalService) OnCandidate(f func(sourceID, candidate string)) {
	m.Called(f)
}

// ==================== NewManager Tests ====================

func TestNewManager(t *testing.T) {
	t.Run("Creates Manager With Signal Service", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "stun:stun.l.google.com:19302")

		require.NotNil(t, mgr)
		assert.NotNil(t, mgr.agents)
		assert.NotNil(t, mgr.pendingAnswers)
		assert.Equal(t, "stun:stun.l.google.com:19302", mgr.stunURL)
	})

	t.Run("Creates Manager With Empty STUN URL", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		require.NotNil(t, mgr)
		assert.Equal(t, "", mgr.stunURL)
	})
}

// ==================== Manager.Start Tests ====================

func TestManager_Start(t *testing.T) {
	t.Run("Registers Signal Handlers", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mockSignal.On("OnOffer", mock.Anything).Return()
		mockSignal.On("OnAnswer", mock.Anything).Return()
		mockSignal.On("OnCandidate", mock.Anything).Return()

		mgr := NewManager(mockSignal, "")
		mgr.Start()

		mockSignal.AssertCalled(t, "OnOffer", mock.Anything)
		mockSignal.AssertCalled(t, "OnAnswer", mock.Anything)
		mockSignal.AssertCalled(t, "OnCandidate", mock.Anything)
	})
}

func TestManager_Connect_AlreadyConnected(t *testing.T) {
	mockSignal := new(MockSignalService)
	mgr := NewManager(mockSignal, "")

	// Manually add an agent to simulate existing connection
	mgr.mu.Lock()
	mgr.agents["peer1"] = &Agent{}
	mgr.mu.Unlock()

	ctx := context.Background()
	_, err := mgr.Connect(ctx, "peer1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_IsConnected(t *testing.T) {
	mockSignal := new(MockSignalService)
	mgr := NewManager(mockSignal, "")

	assert.False(t, mgr.IsConnected("peer1"))

	mgr.mu.Lock()
	mgr.agents["peer1"] = &Agent{}
	mgr.mu.Unlock()

	assert.True(t, mgr.IsConnected("peer1"))
}

func TestManager_RemovePeer(t *testing.T) {
	mockSignal := new(MockSignalService)
	mgr := NewManager(mockSignal, "")

	// Create a real agent to test Close
	agent, err := NewAgent("")
	assert.NoError(t, err)

	mgr.mu.Lock()
	mgr.agents["peer1"] = agent
	mgr.mu.Unlock()

	assert.True(t, mgr.IsConnected("peer1"))

	mgr.RemovePeer("peer1")

	assert.False(t, mgr.IsConnected("peer1"))
}

// ==================== Manager.GetPeerStatus Tests ====================

func TestManager_GetPeerStatus(t *testing.T) {
	t.Run("Returns Disconnected Status For Unknown Peer", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		status := mgr.GetPeerStatus("unknown-peer")
		assert.False(t, status.Connected)
		assert.Equal(t, "disconnected", status.ConnectionState)
	})

	t.Run("Returns Status For Existing Peer", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		// Add a real agent
		agent, err := NewAgent("")
		require.NoError(t, err)
		defer agent.Close()

		mgr.mu.Lock()
		mgr.agents["peer-1"] = agent
		mgr.mu.Unlock()

		status := mgr.GetPeerStatus("peer-1")
		// Agent is in new state, not connected (ConnectionState returns "New" capitalized)
		assert.Equal(t, "New", status.ConnectionState)
	})
}

// ==================== Manager.SetNewConnectionCallback Tests ====================

func TestManager_SetNewConnectionCallback(t *testing.T) {
	t.Run("Sets Callback", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		mgr.SetNewConnectionCallback(func(peerID string, conn *ice.Conn) {
			// Callback set
		})

		assert.NotNil(t, mgr.onNewConnection)
	})
}

// ==================== PeerStatus Tests ====================

func TestPeerStatus(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		status := PeerStatus{
			Connected:       true,
			ConnectionState: "connected",
			LocalCandidate:  "local-candidate",
			RemoteCandidate: "remote-candidate",
			LatencyMs:       50,
		}

		assert.True(t, status.Connected)
		assert.Equal(t, "connected", status.ConnectionState)
		assert.Equal(t, "local-candidate", status.LocalCandidate)
		assert.Equal(t, "remote-candidate", status.RemoteCandidate)
		assert.Equal(t, int64(50), status.LatencyMs)
	})

	t.Run("Default Values", func(t *testing.T) {
		status := PeerStatus{}

		assert.False(t, status.Connected)
		assert.Empty(t, status.ConnectionState)
		assert.Equal(t, int64(0), status.LatencyMs)
	})
}

// ==================== answerData Tests ====================

func TestAnswerData(t *testing.T) {
	t.Run("Has Ufrag And Pwd Fields", func(t *testing.T) {
		data := answerData{
			ufrag: "test-ufrag",
			pwd:   "test-pwd",
		}

		assert.Equal(t, "test-ufrag", data.ufrag)
		assert.Equal(t, "test-pwd", data.pwd)
	})
}

// ==================== Constants Tests ====================

func TestP2PConstants(t *testing.T) {
	t.Run("Message Types Are Defined", func(t *testing.T) {
		assert.Equal(t, 0x01, msgPing)
		assert.Equal(t, 0x02, msgPong)
	})
}

// ==================== Concurrent Access Tests ====================

func TestManager_ConcurrentAccess(t *testing.T) {
	t.Run("Safe Concurrent Agent Map Access", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		done := make(chan bool, 20)

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func() {
				_ = mgr.IsConnected("peer-1")
				done <- true
			}()
		}

		// Concurrent GetPeerStatus
		for i := 0; i < 10; i++ {
			go func() {
				_ = mgr.GetPeerStatus("peer-1")
				done <- true
			}()
		}

		for i := 0; i < 20; i++ {
			<-done
		}
	})
}

// ==================== RemovePeer Edge Cases ====================

func TestManager_RemovePeer_EdgeCases(t *testing.T) {
	t.Run("Safe To Remove Non-Existent Peer", func(t *testing.T) {
		mockSignal := new(MockSignalService)
		mgr := NewManager(mockSignal, "")

		// Should not panic
		mgr.RemovePeer("non-existent")
	})
}

// ==================== Integration Tests ====================

// ForwardingSignalService routes messages between managers in memory
type ForwardingSignalService struct {
	peers map[string]*ForwardingSignalService // Routing table: peerID -> service
	id    string                              // ID of the peer using this service
	
	// Callbacks set by the manager
	onOffer     func(sourceID, ufrag, pwd string)
	onAnswer    func(sourceID, ufrag, pwd string)
	onCandidate func(sourceID, candidate string)
}

func NewForwardingSignalService(id string) *ForwardingSignalService {
	return &ForwardingSignalService{
		peers: make(map[string]*ForwardingSignalService),
		id:    id,
	}
}

func (s *ForwardingSignalService) RegisterPeer(id string, peer *ForwardingSignalService) {
	s.peers[id] = peer
}

func (s *ForwardingSignalService) SendOffer(targetID, ufrag, pwd string) error {
	if peer, ok := s.peers[targetID]; ok {
		// Run in goroutine to simulate async network and avoid deadlocks
		go func() {
			if peer.onOffer != nil {
				peer.onOffer(s.id, ufrag, pwd)
			}
		}()
		return nil
	}
	return fmt.Errorf("peer not found: %s", targetID)
}

func (s *ForwardingSignalService) SendAnswer(targetID, ufrag, pwd string) error {
	if peer, ok := s.peers[targetID]; ok {
		go func() {
			if peer.onAnswer != nil {
				peer.onAnswer(s.id, ufrag, pwd)
			}
		}()
		return nil
	}
	return fmt.Errorf("peer not found: %s", targetID)
}

func (s *ForwardingSignalService) SendCandidate(targetID, candidate string) error {
	if peer, ok := s.peers[targetID]; ok {
		go func() {
			if peer.onCandidate != nil {
				peer.onCandidate(s.id, candidate)
			}
		}()
		return nil
	}
	return fmt.Errorf("peer not found: %s", targetID)
}

func (s *ForwardingSignalService) OnOffer(f func(sourceID, ufrag, pwd string)) {
	s.onOffer = f
}

func (s *ForwardingSignalService) OnAnswer(f func(sourceID, ufrag, pwd string)) {
	s.onAnswer = f
}

func (s *ForwardingSignalService) OnCandidate(f func(sourceID, candidate string)) {
	s.onCandidate = f
}

func TestIntegration_P2P_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup signal services
	sig1 := NewForwardingSignalService("peer1")
	sig2 := NewForwardingSignalService("peer2")

	// Register peers with each other
	sig1.RegisterPeer("peer2", sig2)
	sig2.RegisterPeer("peer1", sig1)

	// Create Managers (using empty STUN to force host candidates which works locally)
	// Note: We need a valid STUN URL format or empty? NewAgent handles empty by using default.
	// But default uses google stun. We want to avoid external deps if possible, but for integration 
	// we rely on ICE. Host candidates should be gathered regardless.
	mgr1 := NewManager(sig1, "")
	mgr2 := NewManager(sig2, "")

	mgr1.Start()
	mgr2.Start()

	// Metrics for mgr2 to verify callback
	mgr2Connected := make(chan bool)
	mgr2.SetNewConnectionCallback(func(peerID string, conn *ice.Conn) {
		if peerID == "peer1" {
			mgr2Connected <- true
		}
	})

	// Initiate connection from peer1 to peer2
	ctx := context.Background()
	mgr1Connected := make(chan bool)
	
	// We run Connect in goroutine because it blocks until connected
	go func() {
		conn, err := mgr1.Connect(ctx, "peer2")
		if err != nil {
			t.Logf("Connect failed: %v", err)
		} else if conn != nil {
			mgr1Connected <- true
		}
	}()

	// Wait with timeout for BOTH sides
	timeout := time.After(10 * time.Second)
	m1Done := false
	m2Done := false

	for !m1Done || !m2Done {
		select {
		case <-mgr1Connected:
			m1Done = true
		case <-mgr2Connected:
			m2Done = true
		case <-timeout:
			t.Fatal("Timeout waiting for P2P connection")
		}
	}

	assert.True(t, m1Done, "Peer1 should be connected")
	assert.True(t, m2Done, "Peer2 should be connected")

	// Verify status
	status1 := mgr1.GetPeerStatus("peer2")
	assert.True(t, status1.Connected)
	// ConnectionState can be "Connected" or "Checking" depending on timing
	assert.Contains(t, []string{"Connected", "Checking"}, status1.ConnectionState)

	status2 := mgr2.GetPeerStatus("peer1")
	assert.True(t, status2.Connected)
}
