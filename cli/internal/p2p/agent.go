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

	// Connection type tracking (for relay detection)
	isRelay bool
}

// ICEConfig holds STUN/TURN configuration
type ICEConfig struct {
	STUNServers []string
	TURNServer  *TURNCredentials
}

// TURNCredentials holds time-limited TURN server credentials
type TURNCredentials struct {
	URL        string
	Username   string
	Credential string
}

// NewAgent creates a new ICE agent with STUN only (legacy)
func NewAgent(stunURL string) (*Agent, error) {
	config := ICEConfig{
		STUNServers: []string{stunURL},
	}
	return NewAgentWithConfig(config)
}

// NewAgentWithConfig creates a new ICE agent with full ICE configuration
func NewAgentWithConfig(config ICEConfig) (*Agent, error) {
	var urls []*ice.URL

	// Add STUN servers
	for _, stunURL := range config.STUNServers {
		if stunURL == "" {
			continue
		}
		uri, err := ice.ParseURL(stunURL)
		if err != nil {
			// Log but don't fail - try remaining servers
			continue
		}
		urls = append(urls, uri)
	}

	// Add default STUN if none provided
	if len(urls) == 0 {
		defaultSTUN, _ := ice.ParseURL("stun:stun.l.google.com:19302")
		urls = append(urls, defaultSTUN)
	}

	// Add TURN server if provided
	if config.TURNServer != nil && config.TURNServer.URL != "" {
		turnURI, err := ice.ParseURL(config.TURNServer.URL)
		if err == nil {
			turnURI.Username = config.TURNServer.Username
			turnURI.Password = config.TURNServer.Credential
			urls = append(urls, turnURI)
		}
	}

	agent, err := ice.NewAgent(&ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4, ice.NetworkTypeUDP6},
		Urls:         urls,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %w", err)
	}

	a := &Agent{
		agent:   agent,
		state:   ice.ConnectionStateNew,
		isRelay: false,
	}

	// Track state changes
	if err := agent.OnConnectionStateChange(func(s ice.ConnectionState) {
		a.stateMu.Lock()
		a.state = s
		a.stateMu.Unlock()

		// Check for relay connection when connected
		if s == ice.ConnectionStateConnected || s == ice.ConnectionStateCompleted {
			a.detectRelayConnection()
		}
	}); err != nil {
		return nil, fmt.Errorf("failed to set state change handler: %w", err)
	}

	return a, nil
}

// detectRelayConnection checks if the current connection uses relay
func (a *Agent) detectRelayConnection() {
	pair, err := a.agent.GetSelectedCandidatePair()
	if err != nil || pair == nil {
		return
	}

	a.stateMu.Lock()
	defer a.stateMu.Unlock()

	// Check if either local or remote candidate is a relay
	if pair.Local != nil && pair.Local.Type() == ice.CandidateTypeRelay {
		a.isRelay = true
		return
	}
	if pair.Remote != nil && pair.Remote.Type() == ice.CandidateTypeRelay {
		a.isRelay = true
		return
	}
	a.isRelay = false
}

// IsRelay returns true if the connection is using a TURN relay
func (a *Agent) IsRelay() bool {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.isRelay
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
	// Check relay status after dial
	a.detectRelayConnection()
	return conn, nil
}

// Accept accepts an incoming connection from a remote peer
func (a *Agent) Accept(ctx context.Context, ufrag, pwd string) (*ice.Conn, error) {
	conn, err := a.agent.Accept(ctx, ufrag, pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to accept: %w", err)
	}
	// Check relay status after accept
	a.detectRelayConnection()
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

// GetCandidatePairInfo returns information about the current candidate pair
func (a *Agent) GetCandidatePairInfo() (localType, remoteType string, err error) {
	pair, err := a.agent.GetSelectedCandidatePair()
	if err != nil || pair == nil {
		return "", "", fmt.Errorf("no selected candidate pair")
	}

	if pair.Local != nil {
		localType = pair.Local.Type().String()
	}
	if pair.Remote != nil {
		remoteType = pair.Remote.Type().String()
	}

	return localType, remoteType, nil
}
