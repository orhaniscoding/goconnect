package deeplink

import (
	"context"
	"fmt"
	"time"

	pb "github.com/orhaniscoding/goconnect/cli/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Result represents the outcome of a deep link action
type Result struct {
	Success bool
	Message string
	Data    map[string]interface{}
}

// Handler processes deep links by communicating with the daemon
type Handler struct {
	grpcTarget     string
	connectTimeout time.Duration
	requestTimeout time.Duration
}

// HandlerOption configures a Handler
type HandlerOption func(*Handler)

// WithGRPCTarget sets the gRPC target address
func WithGRPCTarget(target string) HandlerOption {
	return func(h *Handler) {
		h.grpcTarget = target
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(d time.Duration) HandlerOption {
	return func(h *Handler) {
		h.connectTimeout = d
	}
}

// WithRequestTimeout sets the request timeout
func WithRequestTimeout(d time.Duration) HandlerOption {
	return func(h *Handler) {
		h.requestTimeout = d
	}
}

// NewHandler creates a new deep link handler
func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{
		grpcTarget:     getDefaultGRPCTarget(),
		connectTimeout: 5 * time.Second,
		requestTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Handle processes a deep link and returns the result
func (h *Handler) Handle(dl *DeepLink) (*Result, error) {
	if dl == nil {
		return nil, fmt.Errorf("deep link is nil")
	}

	switch dl.Action {
	case ActionJoin:
		return h.handleJoin(dl)
	case ActionNetwork:
		return h.handleNetwork(dl)
	case ActionConnect:
		return h.handleConnect(dl)
	case ActionLogin:
		return h.handleLogin(dl)
	default:
		return nil, fmt.Errorf("unsupported action: %s", dl.Action)
	}
}

// handleJoin joins a network using an invite code
func (h *Handler) handleJoin(dl *DeepLink) (*Result, error) {
	conn, err := h.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	resp, err := client.JoinNetwork(ctx, &pb.JoinNetworkRequest{
		InviteCode: dl.Target,
	})
	if err != nil {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Failed to join network: %v", err),
		}, nil
	}

	return &Result{
		Success: true,
		Message: fmt.Sprintf("Successfully joined network: %s", resp.Network.Name),
		Data: map[string]interface{}{
			"network_id":   resp.Network.Id,
			"network_name": resp.Network.Name,
			"role":         resp.Network.MyRole.String(),
		},
	}, nil
}

// handleNetwork views or connects to a specific network
func (h *Handler) handleNetwork(dl *DeepLink) (*Result, error) {
	conn, err := h.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// GetNetwork returns *Network directly (not a response wrapper)
	network, err := client.GetNetwork(ctx, &pb.GetNetworkRequest{
		NetworkId: dl.Target,
	})
	if err != nil {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Failed to get network: %v", err),
		}, nil
	}

	result := &Result{
		Success: true,
		Message: fmt.Sprintf("Network: %s", network.Name),
		Data: map[string]interface{}{
			"network_id":   network.Id,
			"network_name": network.Name,
			"role":         network.MyRole.String(),
		},
	}

	return result, nil
}

// handleConnect shows info about connecting to a specific peer
// Note: Direct peer connection is not yet implemented in the daemon
func (h *Handler) handleConnect(dl *DeepLink) (*Result, error) {
	conn, err := h.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Get peer info (connect action shows peer details for now)
	peer, err := client.GetPeer(ctx, &pb.GetPeerRequest{
		PeerId: dl.Target,
	})
	if err != nil {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Failed to find peer: %v", err),
		}, nil
	}

	return &Result{
		Success: true,
		Message: fmt.Sprintf("Peer found: %s", peer.DisplayName),
		Data: map[string]interface{}{
			"peer_id":      peer.Id,
			"display_name": peer.DisplayName,
			"status":       peer.Status.String(),
		},
	}, nil
}

// handleLogin is handled separately (uses keyring, not gRPC)
func (h *Handler) handleLogin(dl *DeepLink) (*Result, error) {
	// Login is handled by the main.go directly since it needs keyring access
	// This is just a placeholder that returns the parsed params
	return &Result{
		Success: true,
		Message: "Login parameters parsed",
		Data: map[string]interface{}{
			"token":  dl.Params["token"],
			"server": dl.Params["server"],
		},
	}, nil
}

// connect establishes a gRPC connection to the daemon
func (h *Handler) connect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.connectTimeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	conn, err := grpc.DialContext(ctx, h.grpcTarget, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial daemon at %s: %w", h.grpcTarget, err)
	}

	return conn, nil
}
