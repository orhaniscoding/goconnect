package tui

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient wraps the gRPC connection to the daemon.
type GRPCClient struct {
	conn           *grpc.ClientConn
	daemonClient   pb.DaemonServiceClient
	networkClient  pb.NetworkServiceClient
	peerClient     pb.PeerServiceClient
	chatClient     pb.ChatServiceClient
	transferClient pb.TransferServiceClient
	settingsClient pb.SettingsServiceClient
	defaultTimeout time.Duration
}

// Variables for testing to allow mocking
var (
	loadAuthTokenFunc = daemon.LoadClientToken
	getGRPCTargetFunc = getGRPCTarget
)

// NewGRPCClient creates a new gRPC client connected to the daemon.
func NewGRPCClient(opts ...grpc.DialOption) (*GRPCClient, error) {
	target := getGRPCTargetFunc()

	// Load IPC auth token
	token, err := loadAuthTokenFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to load IPC token: %w", err)
	}

	// Set up connection with IPC token authentication
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(daemon.NewTokenCredentials(token)),
		grpc.WithBlock(),
	}

	defaultOpts = append(defaultOpts, opts...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, defaultOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}

	return &GRPCClient{
		conn:           conn,
		daemonClient:   pb.NewDaemonServiceClient(conn),
		networkClient:  pb.NewNetworkServiceClient(conn),
		peerClient:     pb.NewPeerServiceClient(conn),
		chatClient:     pb.NewChatServiceClient(conn),
		transferClient: pb.NewTransferServiceClient(conn),
		settingsClient: pb.NewSettingsServiceClient(conn),
		defaultTimeout: 5 * time.Second,
	}, nil
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getGRPCTarget returns the daemon gRPC address based on OS.
func getGRPCTarget() string {
	switch runtime.GOOS {
	case "windows":
		// TCP on localhost for Windows
		return "127.0.0.1:34101"
	default:
		// Unix socket for Linux/macOS
		return "unix:///tmp/goconnect-daemon.sock"
	}
}

// CheckDaemonStatus verifies if the daemon is reachable via gRPC.
func (c *GRPCClient) CheckDaemonStatus() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.daemonClient.GetStatus(ctx, &pb.GetStatusRequest{})
	return err == nil
}

// GetStatus fetches the current status from the daemon.
func (c *GRPCClient) GetStatus() (*Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.daemonClient.GetStatus(ctx, &pb.GetStatusRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Get networks
	networks, _ := c.GetNetworks()

	return &Status{
		Connected:     resp.Status == pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
		NetworkName:   resp.CurrentNetworkName,
		IP:            resp.VirtualIp,
		OnlineMembers: int(resp.ActivePeers),
		Networks:      networks,
	}, nil
}

// GetVersion returns the daemon version information.
func (c *GRPCClient) GetVersion() (*pb.VersionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)

	defer cancel()

	return c.daemonClient.GetVersion(ctx, nil)
}

// CreateNetwork creates a new network.
func (c *GRPCClient) CreateNetwork(name, cidr string) (*Network, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.networkClient.CreateNetwork(ctx, &pb.CreateNetworkRequest{
		Name: name,
		Cidr: cidr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	return &Network{
		ID:   resp.Network.Id,
		Name: resp.Network.Name,
		Role: mapNetworkRole(resp.Network.MyRole),
	}, nil
}

// JoinNetwork joins an existing network via invite code.
func (c *GRPCClient) JoinNetwork(inviteCode string) (*Network, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.networkClient.JoinNetwork(ctx, &pb.JoinNetworkRequest{
		InviteCode: inviteCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to join network: %w", err)
	}

	return &Network{
		ID:   resp.Network.Id,
		Name: resp.Network.Name,
		Role: mapNetworkRole(resp.Network.MyRole),
	}, nil
}

// GetNetworks fetches the list of networks.
func (c *GRPCClient) GetNetworks() ([]Network, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.networkClient.ListNetworks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	networks := make([]Network, 0, len(resp.Networks))
	for _, n := range resp.Networks {
		networks = append(networks, Network{
			ID:   n.Id,
			Name: n.Name,
			Role: mapNetworkRole(n.MyRole),
		})
	}

	return networks, nil
}

// LeaveNetwork disconnects from a network.
func (c *GRPCClient) LeaveNetwork(networkID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	_, err := c.networkClient.LeaveNetwork(ctx, &pb.LeaveNetworkRequest{
		NetworkId: networkID,
	})
	return err
}

// Peer represents a peer in the network.
type Peer struct {
	ID             string
	Name           string
	VirtualIP      string
	Status         string
	ConnectionType string
	LatencyMs      int64
}

// GetPeers returns all peers in the current network.
func (c *GRPCClient) GetPeers() ([]Peer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.peerClient.GetPeers(ctx, &pb.GetPeersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	peers := make([]Peer, 0, len(resp.Peers))
	for _, p := range resp.Peers {
		peers = append(peers, Peer{
			ID:             p.Id,
			Name:           p.Name,
			VirtualIP:      p.VirtualIp,
			Status:         mapConnectionStatus(p.Status),
			ConnectionType: mapConnectionType(p.ConnectionType),
			LatencyMs:      p.LatencyMs,
		})
	}

	return peers, nil
}

// SendChatMessage sends a chat message.
func (c *GRPCClient) SendChatMessage(networkID, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	_, err := c.chatClient.SendMessage(ctx, &pb.SendMessageRequest{
		NetworkId: networkID,
		Content:   content,
	})
	return err
}

// SendFile initiates a file transfer.
func (c *GRPCClient) SendFile(peerID, filePath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	resp, err := c.transferClient.SendFile(ctx, &pb.SendFileRequest{
		PeerId:   peerID,
		FilePath: filePath,
	})
	if err != nil {
		return "", err
	}
	return resp.TransferId, nil
}

// GetSettings returns the current daemon settings.
func (c *GRPCClient) GetSettings() (*pb.Settings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	return c.settingsClient.GetSettings(ctx, nil)
}

// UpdateSettings updates daemon settings.
func (c *GRPCClient) UpdateSettings(settings *pb.Settings) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	_, err := c.settingsClient.UpdateSettings(ctx, &pb.UpdateSettingsRequest{
		Settings: settings,
	})
	return err
}

// Subscribe subscribes to daemon events.
func (c *GRPCClient) Subscribe(ctx context.Context, eventTypes []pb.EventType) (pb.DaemonService_SubscribeClient, error) {
	return c.daemonClient.Subscribe(ctx, &pb.SubscribeRequest{
		EventTypes: eventTypes,
	})
}

// TryConnect attempts to connect to the daemon with retries.
func TryConnectGRPC(maxRetries int, retryDelay time.Duration) (*GRPCClient, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		client, err := NewGRPCClient()
		if err == nil {
			return client, nil
		}
		lastErr = err
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, lastErr)
}

// IsGRPCAvailable checks if gRPC daemon is available.
func IsGRPCAvailable() bool {
	target := getGRPCTarget()

	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		conn, err = net.DialTimeout("tcp", "127.0.0.1:34101", 500*time.Millisecond)
	} else {
		conn, err = net.DialTimeout("unix", "/tmp/goconnect-daemon.sock", 500*time.Millisecond)
	}

	if err != nil {
		_ = target // suppress unused warning
		return false
	}
	conn.Close()
	return true
}

// Helper functions

func mapNetworkRole(role pb.NetworkRole) string {
	switch role {
	case pb.NetworkRole_NETWORK_ROLE_OWNER:
		return "host"
	case pb.NetworkRole_NETWORK_ROLE_ADMIN:
		return "admin"
	case pb.NetworkRole_NETWORK_ROLE_MEMBER:
		return "client"
	default:
		return "unknown"
	}
}

func mapConnectionStatus(status pb.ConnectionStatus) string {
	switch status {
	case pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED:
		return "connected"
	case pb.ConnectionStatus_CONNECTION_STATUS_CONNECTING:
		return "connecting"
	case pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED:
		return "disconnected"
	case pb.ConnectionStatus_CONNECTION_STATUS_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}

func mapConnectionType(connType pb.ConnectionType) string {
	switch connType {
	case pb.ConnectionType_CONNECTION_TYPE_DIRECT:
		return "direct"
	case pb.ConnectionType_CONNECTION_TYPE_RELAY:
		return "relay"
	default:
		return "unknown"
	}
}
