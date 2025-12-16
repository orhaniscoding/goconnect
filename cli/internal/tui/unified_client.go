package tui

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
)

// DaemonClient is the unified interface for communicating with the daemon.
// It abstracts both HTTP and gRPC backends.
type DaemonClient interface {
	// Connection
	CheckDaemonStatus() bool
	Close() error

	// Status
	GetStatus() (*Status, error)

	// Networks
	CreateNetwork(name, cidr string) (*Network, error)
	JoinNetwork(inviteCode string) (*Network, error)
	GetNetworks() ([]Network, error)
	LeaveNetwork(networkID string) error

	// Peers
	GetPeers() ([]Peer, error)

	// Chat
	SendChatMessage(networkID, content string) error

	// Files
	SendFile(peerID, filePath string) (string, error)

	// Settings
	GetSettings() (*pb.Settings, error)
	UpdateSettings(settings *pb.Settings) error
}

// UnifiedClient implements DaemonClient and can use either HTTP or gRPC.
type UnifiedClient struct {
	grpcClient *GRPCClient
	httpClient *Client
	useGRPC    bool
	mu         sync.RWMutex
}

// NewUnifiedClient creates a new unified client that prefers gRPC but falls back to HTTP.
func NewUnifiedClient() *UnifiedClient {
	uc := &UnifiedClient{
		httpClient: NewClient(),
		useGRPC:    false,
	}

	// Try to connect via gRPC first
	if IsGRPCAvailable() {
		grpcClient, err := NewGRPCClient()
		if err == nil {
			uc.grpcClient = grpcClient
			uc.useGRPC = true
		}
	}

	return uc
}

// NewUnifiedClientWithMode creates a client with a specific mode.
func NewUnifiedClientWithMode(preferGRPC bool) *UnifiedClient {
	uc := &UnifiedClient{
		httpClient: NewClient(),
		useGRPC:    false,
	}

	if preferGRPC {
		grpcClient, err := NewGRPCClient()
		if err == nil {
			uc.grpcClient = grpcClient
			uc.useGRPC = true
		}
	}

	return uc
}

// IsUsingGRPC returns true if the client is using gRPC.
func (u *UnifiedClient) IsUsingGRPC() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.useGRPC
}

// SwitchToGRPC attempts to switch to gRPC backend.
func (u *UnifiedClient) SwitchToGRPC() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.useGRPC && u.grpcClient != nil {
		return nil // Already using gRPC
	}

	grpcClient, err := NewGRPCClient()
	if err != nil {
		return err
	}

	u.grpcClient = grpcClient
	u.useGRPC = true
	return nil
}

// SwitchToHTTP switches to HTTP backend.
func (u *UnifiedClient) SwitchToHTTP() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.grpcClient != nil {
		u.grpcClient.Close()
		u.grpcClient = nil
	}
	u.useGRPC = false
}

// Close closes all connections.
func (u *UnifiedClient) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.grpcClient != nil {
		return u.grpcClient.Close()
	}
	return nil
}

// CheckDaemonStatus verifies if the daemon is reachable.
func (u *UnifiedClient) CheckDaemonStatus() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.CheckDaemonStatus()
	}
	return u.httpClient.CheckDaemonStatus()
}

// GetStatus fetches the current status from the daemon.
func (u *UnifiedClient) GetStatus() (*Status, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.GetStatus()
	}
	return u.httpClient.GetStatus()
}

// CreateNetwork creates a new network.
func (u *UnifiedClient) CreateNetwork(name, cidr string) (*Network, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.CreateNetwork(name, cidr)
	}
	return u.httpClient.CreateNetwork(name, cidr)
}

// JoinNetwork joins an existing network.
func (u *UnifiedClient) JoinNetwork(inviteCode string) (*Network, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.JoinNetwork(inviteCode)
	}
	return u.httpClient.JoinNetwork(inviteCode)
}

// GetNetworks fetches the list of networks.
func (u *UnifiedClient) GetNetworks() ([]Network, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.GetNetworks()
	}
	return u.httpClient.GetNetworks()
}

// LeaveNetwork disconnects from a network.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) LeaveNetwork(networkID string) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.LeaveNetwork(networkID)
	}
	// LeaveNetwork is only available via gRPC (daemon-specific operation)
	return fmt.Errorf("leave network requires gRPC connection to daemon")
}

// GetPeers returns all peers in the current network.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) GetPeers() ([]Peer, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.GetPeers()
	}
	// GetPeers is only available via gRPC (daemon-specific operation)
	return nil, fmt.Errorf("get peers requires gRPC connection to daemon")
}

// SendChatMessage sends a chat message.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) SendChatMessage(networkID, content string) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.SendChatMessage(networkID, content)
	}
	// SendChatMessage is only available via gRPC (daemon-specific operation)
	return fmt.Errorf("send chat message requires gRPC connection to daemon")
}

// SendFile initiates a file transfer.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) SendFile(peerID, filePath string) (string, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.SendFile(peerID, filePath)
	}
	// SendFile is only available via gRPC (daemon-specific operation)
	return "", fmt.Errorf("send file requires gRPC connection to daemon")
}

// GetSettings returns the current daemon settings.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) GetSettings() (*pb.Settings, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.GetSettings()
	}
	// GetSettings is only available via gRPC (daemon-specific operation)
	return nil, fmt.Errorf("get settings requires gRPC connection to daemon")
}

// UpdateSettings updates daemon settings.
// Note: This is a daemon-specific operation and requires gRPC connection.
func (u *UnifiedClient) UpdateSettings(settings *pb.Settings) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.UpdateSettings(settings)
	}
	// UpdateSettings is only available via gRPC (daemon-specific operation)
	return fmt.Errorf("update settings requires gRPC connection to daemon")
}

// Subscribe subscribes to daemon events (gRPC only).
func (u *UnifiedClient) Subscribe(ctx context.Context, eventTypes []pb.EventType) (pb.DaemonService_SubscribeClient, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.Subscribe(ctx, eventTypes)
	}
	return nil, fmt.Errorf("subscribe only available with gRPC")
}

// GetVersion returns daemon version (gRPC only).
func (u *UnifiedClient) GetVersion() (*pb.VersionResponse, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.useGRPC && u.grpcClient != nil {
		return u.grpcClient.GetVersion()
	}
	return nil, fmt.Errorf("get version only available with gRPC")
}
