package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
