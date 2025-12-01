package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/ice/v2"
)

// Agent manages the ICE agent for P2P connections
type Agent struct {
	agent *ice.Agent

	// State tracking
	state   ice.ConnectionState
	stateMu sync.RWMutex
}

// NewAgent creates a new ICE agent
func NewAgent(stunURL string) (*Agent, error) {
	if stunURL == "" {
		stunURL = "stun:stun.l.google.com:19302" // Default
	}

	uri, err := ice.ParseURL(stunURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse STUN URL: %w", err)
	}

	agent, err := ice.NewAgent(&ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4, ice.NetworkTypeUDP6},
		Urls:         []*ice.URL{uri},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %w", err)
	}

	a := &Agent{
		agent: agent,
		state: ice.ConnectionStateNew,
	}

	// Track state changes
	if err := agent.OnConnectionStateChange(func(s ice.ConnectionState) {
		a.stateMu.Lock()
		a.state = s
		a.stateMu.Unlock()
	}); err != nil {
		return nil, fmt.Errorf("failed to set state change handler: %w", err)
	}

	return a, nil
}

// GetLocalCredentials returns the local ufrag and pwd
func (a *Agent) GetLocalCredentials() (string, string, error) {
	return a.agent.GetLocalUserCredentials()
}

// GatherCandidates starts the candidate gathering process
func (a *Agent) GatherCandidates() error {
	if err := a.agent.GatherCandidates(); err != nil {
		return fmt.Errorf("failed to gather candidates: %w", err)
	}
	return nil
}

// OnCandidate sets the callback for when a new candidate is discovered
func (a *Agent) OnCandidate(f func(ice.Candidate)) error {
	return a.agent.OnCandidate(f)
}

// OnConnectionStateChange sets the callback for connection state changes
func (a *Agent) OnConnectionStateChange(f func(ice.ConnectionState)) error {
	// Wrap the callback to ensure we also update our local state
	return a.agent.OnConnectionStateChange(func(s ice.ConnectionState) {
		a.stateMu.Lock()
		a.state = s
		a.stateMu.Unlock()
		f(s)
	})
}

// ConnectionState returns the current connection state
func (a *Agent) ConnectionState() ice.ConnectionState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

// Dial initiates a connection to a remote peer
func (a *Agent) Dial(ctx context.Context, ufrag, pwd string) (*ice.Conn, error) {
	conn, err := a.agent.Dial(ctx, ufrag, pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	return conn, nil
}

// Accept accepts an incoming connection from a remote peer
func (a *Agent) Accept(ctx context.Context, ufrag, pwd string) (*ice.Conn, error) {
	conn, err := a.agent.Accept(ctx, ufrag, pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to accept: %w", err)
	}
	return conn, nil
}

// AddRemoteCandidate adds a remote candidate to the agent
func (a *Agent) AddRemoteCandidate(candidate ice.Candidate) error {
	return a.agent.AddRemoteCandidate(candidate)
}

// Close closes the agent
func (a *Agent) Close() error {
	return a.agent.Close()
}
