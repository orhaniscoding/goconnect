package daemon

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/orhaniscoding/goconnect/cli/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockEngine implements a minimal engine for testing
type mockEngine struct {
	status        map[string]interface{}
	networks      []NetworkResult
	peers         []PeerResult
	createErr     error
	joinErr       error
	getNetworkErr error
	getPeersErr   error
	kickErr       error
	banErr        error
	unbanErr      error
	rejectErr     error
	cancelErr     error
}

// NetworkResult for mock engine
type NetworkResult struct {
	ID         string
	Name       string
	InviteCode string
}

// PeerResult for mock engine
type PeerResult struct {
	ID        string
	Name      string
	VirtualIP string
	Online    bool
}

// InviteResult for mock engine
type InviteResult struct {
	Code      string
	URL       string
	ExpiresAt time.Time
}

func (m *mockEngine) GetStatus() map[string]interface{} {
	if m.status == nil {
		return map[string]interface{}{
			"connected":    true,
			"virtual_ip":   "10.0.0.1",
			"active_peers": 5,
			"network_id":   "net-123",
			"network_name": "Test Network",
		}
	}
	return m.status
}

func (m *mockEngine) CreateNetwork(name string) (*NetworkResult, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &NetworkResult{
		ID:         "net-" + name,
		Name:       name,
		InviteCode: "invite-abc123",
	}, nil
}

func (m *mockEngine) JoinNetwork(inviteCode string) (*NetworkResult, error) {
	if m.joinErr != nil {
		return nil, m.joinErr
	}
	return &NetworkResult{
		ID:   "net-joined",
		Name: "Joined Network",
	}, nil
}

func (m *mockEngine) GetNetworks() ([]NetworkResult, error) {
	if m.getNetworkErr != nil {
		return nil, m.getNetworkErr
	}
	if m.networks != nil {
		return m.networks, nil
	}
	return []NetworkResult{
		{ID: "net-1", Name: "Network 1"},
		{ID: "net-2", Name: "Network 2"},
	}, nil
}

func (m *mockEngine) GetPeers() ([]PeerResult, error) {
	if m.getPeersErr != nil {
		return nil, m.getPeersErr
	}
	if m.peers != nil {
		return m.peers, nil
	}
	return []PeerResult{
		{ID: "peer-1", Name: "Peer 1", VirtualIP: "10.0.0.2", Online: true},
		{ID: "peer-2", Name: "Peer 2", VirtualIP: "10.0.0.3", Online: false},
	}, nil
}

func (m *mockEngine) GetPeerByID(peerID string) (*PeerResult, error) {
	peers, err := m.GetPeers()
	if err != nil {
		return nil, err
	}
	for _, p := range peers {
		if p.ID == peerID {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("peer not found: %s", peerID)
}

func (m *mockEngine) GenerateInvite(networkID string, maxUses int, expiresHours int) (*InviteResult, error) {
	if networkID == "" {
		return nil, fmt.Errorf("networkID required")
	}
	return &InviteResult{
		Code:      "INV-TEST-123",
		URL:       "https://example.com/invite/INV-TEST-123",
		ExpiresAt: time.Now().Add(time.Duration(expiresHours) * time.Hour),
	}, nil
}

func (m *mockEngine) KickPeer(networkID, peerID, reason string) error {
	if networkID == "" || peerID == "" {
		return fmt.Errorf("networkID and peerID required")
	}
	if m.kickErr != nil {
		return m.kickErr
	}
	return nil
}

func (m *mockEngine) BanPeer(networkID, peerID, reason string) error {
	if networkID == "" || peerID == "" {
		return fmt.Errorf("networkID and peerID required")
	}
	if m.banErr != nil {
		return m.banErr
	}
	return nil
}

func (m *mockEngine) UnbanPeer(networkID, peerID string) error {
	if networkID == "" || peerID == "" {
		return fmt.Errorf("networkID and peerID required")
	}
	if m.unbanErr != nil {
		return m.unbanErr
	}
	return nil
}

func (m *mockEngine) RejectTransfer(transferID string) error {
	if transferID == "" {
		return fmt.Errorf("transferID required")
	}
	if m.rejectErr != nil {
		return m.rejectErr
	}
	return nil
}

func (m *mockEngine) CancelTransfer(transferID string) error {
	if transferID == "" {
		return fmt.Errorf("transferID required")
	}
	if m.cancelErr != nil {
		return m.cancelErr
	}
	return nil
}

// mockLogger implements service.Logger
type mockLogger struct{}

func (m *mockLogger) Error(args ...interface{})                   {}
func (m *mockLogger) Warning(args ...interface{})                 {}
func (m *mockLogger) Info(args ...interface{})                    {}
func (m *mockLogger) Errorf(format string, args ...interface{})   {}
func (m *mockLogger) Warningf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})    {}

// testGRPCServer is a simplified server for testing without full daemon
type testGRPCServer struct {
	pb.UnimplementedDaemonServiceServer
	pb.UnimplementedNetworkServiceServer
	pb.UnimplementedPeerServiceServer
	pb.UnimplementedTransferServiceServer

	engine     *mockEngine
	grpcServer *grpc.Server
	listener   net.Listener
	ipcAuth    *IPCAuth
	version    string
}

func newTestGRPCServer(t *testing.T) (*testGRPCServer, string) {
	t.Helper()

	// Create temp dir for token
	tokenPath := t.TempDir() + "/goconnect-ipc-token"
	ipcAuth := NewIPCAuthWithPath(tokenPath)

	if err := ipcAuth.GenerateAndSave(); err != nil {
		t.Fatalf("Failed to generate IPC token: %v", err)
	}

	// Create test listener on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	server := &testGRPCServer{
		engine:   &mockEngine{},
		listener: listener,
		ipcAuth:  ipcAuth,
		version:  "1.0.0-test",
	}

	// Create gRPC server with auth interceptor
	server.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(ipcAuth.UnaryServerInterceptor()),
	)

	// Register services
	pb.RegisterDaemonServiceServer(server.grpcServer, server)
	pb.RegisterNetworkServiceServer(server.grpcServer, server)
	pb.RegisterPeerServiceServer(server.grpcServer, server)
	pb.RegisterTransferServiceServer(server.grpcServer, server)

	// Start server in background
	go func() {
		server.grpcServer.Serve(listener)
	}()

	return server, listener.Addr().String()
}

func (s *testGRPCServer) stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
	if s.ipcAuth != nil {
		s.ipcAuth.Cleanup()
	}
}

// Implement DaemonService methods for testing
func (s *testGRPCServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	engineStatus := s.engine.GetStatus()

	resp := &pb.GetStatusResponse{
		VirtualIp:   engineStatus["virtual_ip"].(string),
		ActivePeers: int32(engineStatus["active_peers"].(int)),
	}

	if connected, ok := engineStatus["connected"].(bool); ok && connected {
		resp.Status = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED
	}
	if networkID, ok := engineStatus["network_id"].(string); ok {
		resp.CurrentNetworkId = networkID
	}
	if networkName, ok := engineStatus["network_name"].(string); ok {
		resp.CurrentNetworkName = networkName
	}

	return resp, nil
}

func (s *testGRPCServer) GetVersion(ctx context.Context, req *emptypb.Empty) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version: s.version,
	}, nil
}

// Implement NetworkService methods for testing
func (s *testGRPCServer) CreateNetwork(ctx context.Context, req *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "network name is required")
	}

	result, err := s.engine.CreateNetwork(req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create network: %v", err)
	}

	return &pb.CreateNetworkResponse{
		Network: &pb.Network{
			Id:   result.ID,
			Name: result.Name,
		},
		InviteCode: result.InviteCode,
	}, nil
}

func (s *testGRPCServer) JoinNetwork(ctx context.Context, req *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error) {
	if req.InviteCode == "" {
		return nil, status.Error(codes.InvalidArgument, "invite code is required")
	}

	result, err := s.engine.JoinNetwork(req.InviteCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to join network: %v", err)
	}

	return &pb.JoinNetworkResponse{
		Network: &pb.Network{
			Id:   result.ID,
			Name: result.Name,
		},
	}, nil
}

func (s *testGRPCServer) ListNetworks(ctx context.Context, req *emptypb.Empty) (*pb.ListNetworksResponse, error) {
	networks, err := s.engine.GetNetworks()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get networks: %v", err)
	}

	pbNetworks := make([]*pb.Network, len(networks))
	for i, n := range networks {
		pbNetworks[i] = &pb.Network{
			Id:   n.ID,
			Name: n.Name,
		}
	}

	return &pb.ListNetworksResponse{
		Networks: pbNetworks,
	}, nil
}

func (s *testGRPCServer) GenerateInvite(ctx context.Context, req *pb.GenerateInviteRequest) (*pb.GenerateInviteResponse, error) {
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}
	invite, err := s.engine.GenerateInvite(req.NetworkId, int(req.MaxUses), int(req.ExpiresHours))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate invite: %v", err)
	}
	resp := &pb.GenerateInviteResponse{
		InviteCode: invite.Code,
		InviteUrl:  invite.URL,
	}
	if !invite.ExpiresAt.IsZero() {
		resp.ExpiresAt = timestamppb.New(invite.ExpiresAt)
	}
	return resp, nil
}

// Implement PeerService methods for testing
func (s *testGRPCServer) GetPeers(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
	peers, err := s.engine.GetPeers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get peers: %v", err)
	}

	pbPeers := make([]*pb.Peer, len(peers))
	for i, p := range peers {
		peerStatus := pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED
		if p.Online {
			peerStatus = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED
		}
		pbPeers[i] = &pb.Peer{
			Id:        p.ID,
			Name:      p.Name,
			VirtualIp: p.VirtualIP,
			Status:    peerStatus,
		}
	}

	return &pb.GetPeersResponse{
		Peers: pbPeers,
	}, nil
}

func (s *testGRPCServer) GetPeer(ctx context.Context, req *pb.GetPeerRequest) (*pb.Peer, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}

	peer, err := s.engine.GetPeerByID(req.PeerId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "peer not found: %s", req.PeerId)
	}

	peerStatus := pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED
	if peer.Online {
		peerStatus = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED
	}

	return &pb.Peer{
		Id:        peer.ID,
		Name:      peer.Name,
		VirtualIp: peer.VirtualIP,
		Status:    peerStatus,
	}, nil
}

func (s *testGRPCServer) KickPeer(ctx context.Context, req *pb.KickPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}
	if err := s.engine.KickPeer(req.NetworkId, req.PeerId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to kick peer: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *testGRPCServer) BanPeer(ctx context.Context, req *pb.BanPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}
	if err := s.engine.BanPeer(req.NetworkId, req.PeerId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ban peer: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *testGRPCServer) UnbanPeer(ctx context.Context, req *pb.UnbanPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}
	if err := s.engine.UnbanPeer(req.NetworkId, req.PeerId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unban peer: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// Implement TransferService methods for testing
func (s *testGRPCServer) RejectTransfer(ctx context.Context, req *pb.RejectTransferRequest) (*emptypb.Empty, error) {
	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer_id is required")
	}
	if err := s.engine.RejectTransfer(req.TransferId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reject transfer: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *testGRPCServer) CancelTransfer(ctx context.Context, req *pb.CancelTransferRequest) (*emptypb.Empty, error) {
	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer_id is required")
	}
	if err := s.engine.CancelTransfer(req.TransferId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel transfer: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// =============================================================================
// TESTS
// =============================================================================

func TestGRPCServer_AuthRequired(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	// Connect WITHOUT auth token
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.GetStatus(ctx, &pb.GetStatusRequest{})
	if err == nil {
		t.Fatal("Expected error when calling without auth token")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected Unauthenticated, got: %v", st.Code())
	}
}

func TestGRPCServer_AuthWithValidToken(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	// Load the token
	token, err := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	// Connect WITH auth token
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if resp.Status != pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED {
		t.Errorf("Expected connected status, got: %v", resp.Status)
	}
	if resp.VirtualIp != "10.0.0.1" {
		t.Errorf("Expected VirtualIP 10.0.0.1, got: %s", resp.VirtualIp)
	}
	if resp.ActivePeers != 5 {
		t.Errorf("Expected 5 active peers, got: %d", resp.ActivePeers)
	}
}

func TestGRPCServer_AuthWithInvalidToken(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	// Connect with WRONG token
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials("wrong-token")),
	)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.GetStatus(ctx, &pb.GetStatusRequest{})
	if err == nil {
		t.Fatal("Expected error with invalid token")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected Unauthenticated, got: %v", st.Code())
	}
}

func TestGRPCServer_GetVersion(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetVersion(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if resp.Version != "1.0.0-test" {
		t.Errorf("Expected version 1.0.0-test, got: %s", resp.Version)
	}
}

func TestGRPCServer_CreateNetwork(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name        string
		networkName string
		wantErr     codes.Code
	}{
		{
			name:        "valid_network",
			networkName: "MyNetwork",
			wantErr:     codes.OK,
		},
		{
			name:        "empty_name",
			networkName: "",
			wantErr:     codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.CreateNetwork(ctx, &pb.CreateNetworkRequest{
				Name: tt.networkName,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("CreateNetwork failed: %v", err)
				}
				if resp.Network.Name != tt.networkName {
					t.Errorf("Expected name %s, got: %s", tt.networkName, resp.Network.Name)
				}
				if resp.InviteCode == "" {
					t.Error("Expected invite code to be set")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_JoinNetwork(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name       string
		inviteCode string
		wantErr    codes.Code
	}{
		{
			name:       "valid_invite",
			inviteCode: "invite-abc123",
			wantErr:    codes.OK,
		},
		{
			name:       "empty_invite",
			inviteCode: "",
			wantErr:    codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.JoinNetwork(ctx, &pb.JoinNetworkRequest{
				InviteCode: tt.inviteCode,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("JoinNetwork failed: %v", err)
				}
				if resp.Network.Id == "" {
					t.Error("Expected network ID to be set")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_ListNetworks(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.ListNetworks(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListNetworks failed: %v", err)
	}

	if len(resp.Networks) != 2 {
		t.Errorf("Expected 2 networks, got: %d", len(resp.Networks))
	}
}

func TestGRPCServer_GenerateInvite(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Success case
	resp, err := client.GenerateInvite(ctx, &pb.GenerateInviteRequest{NetworkId: "net-1", MaxUses: 5, ExpiresHours: 1})
	if err != nil {
		t.Fatalf("GenerateInvite failed: %v", err)
	}
	if resp.InviteCode == "" {
		t.Errorf("Expected invite code")
	}
	if resp.InviteUrl == "" {
		t.Errorf("Expected invite URL")
	}
	if resp.ExpiresAt == nil {
		t.Errorf("Expected expires_at timestamp")
	}

	// Invalid argument
	_, err = client.GenerateInvite(ctx, &pb.GenerateInviteRequest{NetworkId: ""})
	if err == nil {
		t.Fatal("Expected error for empty network_id")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument, got %v", st.Code())
	}
}

func TestGRPCServer_GetPeers(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetPeers(ctx, &pb.GetPeersRequest{})
	if err != nil {
		t.Fatalf("GetPeers failed: %v", err)
	}

	if len(resp.Peers) != 2 {
		t.Errorf("Expected 2 peers, got: %d", len(resp.Peers))
	}

	// Check first peer
	if resp.Peers[0].Name != "Peer 1" {
		t.Errorf("Expected Peer 1, got: %s", resp.Peers[0].Name)
	}
	if resp.Peers[0].Status != pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED {
		t.Errorf("Expected connected status for Peer 1")
	}

	// Check second peer
	if resp.Peers[1].Status != pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED {
		t.Errorf("Expected disconnected status for Peer 2")
	}
}

func TestGRPCServer_GetPeer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name    string
		peerID  string
		wantErr codes.Code
	}{
		{
			name:    "existing_peer",
			peerID:  "peer-1",
			wantErr: codes.OK,
		},
		{
			name:    "nonexistent_peer",
			peerID:  "peer-999",
			wantErr: codes.NotFound,
		},
		{
			name:    "empty_peer_id",
			peerID:  "",
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetPeer(ctx, &pb.GetPeerRequest{
				PeerId: tt.peerID,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("GetPeer failed: %v", err)
				}
				if resp.Id != tt.peerID {
					t.Errorf("Expected ID %s, got: %s", tt.peerID, resp.Id)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_MetadataToken(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())

	// Connect without PerRPCCredentials, use metadata directly
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	// Add token via metadata using the correct header key
	ctx := metadata.AppendToOutgoingContext(
		context.Background(),
		"x-goconnect-ipc-token", token,
	)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	resp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus with metadata token failed: %v", err)
	}

	if resp.Status != pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED {
		t.Errorf("Expected connected status")
	}
}

func TestGRPCServer_KickPeer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name      string
		networkID string
		peerID    string
		wantErr   codes.Code
	}{
		{name: "success", networkID: "net-1", peerID: "peer-1", wantErr: codes.OK},
		{name: "missing_peer_id", networkID: "net-1", peerID: "", wantErr: codes.InvalidArgument},
		{name: "missing_network_id", networkID: "", peerID: "peer-1", wantErr: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.KickPeer(ctx, &pb.KickPeerRequest{
				NetworkId: tt.networkID,
				PeerId:    tt.peerID,
				Reason:    "test reason",
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("KickPeer failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_BanPeer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name      string
		networkID string
		peerID    string
		wantErr   codes.Code
	}{
		{name: "success", networkID: "net-1", peerID: "peer-1", wantErr: codes.OK},
		{name: "missing_peer_id", networkID: "net-1", peerID: "", wantErr: codes.InvalidArgument},
		{name: "missing_network_id", networkID: "", peerID: "peer-1", wantErr: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.BanPeer(ctx, &pb.BanPeerRequest{
				NetworkId: tt.networkID,
				PeerId:    tt.peerID,
				Reason:    "test ban reason",
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("BanPeer failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_UnbanPeer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name      string
		networkID string
		peerID    string
		wantErr   codes.Code
	}{
		{name: "success", networkID: "net-1", peerID: "peer-1", wantErr: codes.OK},
		{name: "missing_peer_id", networkID: "net-1", peerID: "", wantErr: codes.InvalidArgument},
		{name: "missing_network_id", networkID: "", peerID: "peer-1", wantErr: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.UnbanPeer(ctx, &pb.UnbanPeerRequest{
				NetworkId: tt.networkID,
				PeerId:    tt.peerID,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("UnbanPeer failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_RejectTransfer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name       string
		transferID string
		wantErr    codes.Code
	}{
		{name: "success", transferID: "transfer-123", wantErr: codes.OK},
		{name: "missing_transfer_id", transferID: "", wantErr: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.RejectTransfer(ctx, &pb.RejectTransferRequest{
				TransferId: tt.transferID,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("RejectTransfer failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

func TestGRPCServer_CancelTransfer(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name       string
		transferID string
		wantErr    codes.Code
	}{
		{name: "success", transferID: "transfer-456", wantErr: codes.OK},
		{name: "missing_transfer_id", transferID: "", wantErr: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CancelTransfer(ctx, &pb.CancelTransferRequest{
				TransferId: tt.transferID,
			})

			if tt.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("CancelTransfer failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantErr {
					t.Errorf("Expected %v, got: %v", tt.wantErr, st.Code())
				}
			}
		})
	}
}

// TestGRPCServer_BanPeer_EngineError tests engine error handling
func TestGRPCServer_BanPeer_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.banErr = fmt.Errorf("engine ban error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.BanPeer(ctx, &pb.BanPeerRequest{
		NetworkId: "net-1",
		PeerId:    "peer-1",
		Reason:    "test",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_UnbanPeer_EngineError tests engine error handling
func TestGRPCServer_UnbanPeer_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.unbanErr = fmt.Errorf("engine unban error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.UnbanPeer(ctx, &pb.UnbanPeerRequest{
		NetworkId: "net-1",
		PeerId:    "peer-1",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_KickPeer_EngineError tests engine error handling
func TestGRPCServer_KickPeer_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.kickErr = fmt.Errorf("engine kick error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.KickPeer(ctx, &pb.KickPeerRequest{
		NetworkId: "net-1",
		PeerId:    "peer-1",
		Reason:    "test",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_RejectTransfer_EngineError tests engine error handling
func TestGRPCServer_RejectTransfer_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.rejectErr = fmt.Errorf("engine reject error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.RejectTransfer(ctx, &pb.RejectTransferRequest{
		TransferId: "transfer-123",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_CancelTransfer_EngineError tests engine error handling
func TestGRPCServer_CancelTransfer_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.cancelErr = fmt.Errorf("engine cancel error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.CancelTransfer(ctx, &pb.CancelTransferRequest{
		TransferId: "transfer-456",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_CreateNetwork_EngineError tests engine error handling
func TestGRPCServer_CreateNetwork_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.createErr = fmt.Errorf("engine create error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.CreateNetwork(ctx, &pb.CreateNetworkRequest{
		Name: "TestNetwork",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_JoinNetwork_EngineError tests engine error handling
func TestGRPCServer_JoinNetwork_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.joinErr = fmt.Errorf("engine join error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.JoinNetwork(ctx, &pb.JoinNetworkRequest{
		InviteCode: "invite-abc",
	})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_ListNetworks_EngineError tests engine error handling
func TestGRPCServer_ListNetworks_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.getNetworkErr = fmt.Errorf("engine list error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.ListNetworks(ctx, &emptypb.Empty{})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_GetPeers_EngineError tests engine error handling
func TestGRPCServer_GetPeers_EngineError(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.getPeersErr = fmt.Errorf("engine peers error")
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.GetPeers(ctx, &pb.GetPeersRequest{})

	if err == nil {
		t.Fatal("Expected error from engine")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error, got: %v", st.Code())
	}
}

// TestGRPCServer_GetPeer_NotFound tests peer not found scenario
func TestGRPCServer_GetPeer_NotFound(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.GetPeer(ctx, &pb.GetPeerRequest{
		PeerId: "nonexistent-peer",
	})

	if err == nil {
		t.Fatal("Expected error for nonexistent peer")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("Expected NotFound error, got: %v", st.Code())
	}
}

// TestGRPCServer_GetStatus_DisconnectedState tests disconnected status
func TestGRPCServer_GetStatus_DisconnectedState(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.status = map[string]interface{}{
		"connected":    false,
		"virtual_ip":   "",
		"active_peers": 0,
		"network_id":   "",
		"network_name": "",
	}
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if resp.Status == pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED {
		t.Error("Expected disconnected status")
	}
	if resp.ActivePeers != 0 {
		t.Errorf("Expected 0 active peers, got: %d", resp.ActivePeers)
	}
}

// TestGRPCServer_GenerateInvite_EmptyNetworkID tests empty network ID
func TestGRPCServer_GenerateInvite_EmptyNetworkID(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.GenerateInvite(ctx, &pb.GenerateInviteRequest{
		NetworkId:    "",
		MaxUses:      5,
		ExpiresHours: 24,
	})

	if err == nil {
		t.Fatal("Expected error for empty network ID")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error, got: %v", st.Code())
	}
}

// TestGRPCServer_ListNetworks_EmptyResult tests empty network list
func TestGRPCServer_ListNetworks_EmptyResult(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.networks = []NetworkResult{}
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewNetworkServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.ListNetworks(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListNetworks failed: %v", err)
	}

	if len(resp.Networks) != 0 {
		t.Errorf("Expected 0 networks, got: %d", len(resp.Networks))
	}
}

// TestGRPCServer_GetPeers_EmptyResult tests empty peer list
func TestGRPCServer_GetPeers_EmptyResult(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	server.engine.peers = []PeerResult{}
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetPeers(ctx, &pb.GetPeersRequest{})
	if err != nil {
		t.Fatalf("GetPeers failed: %v", err)
	}

	if len(resp.Peers) != 0 {
		t.Errorf("Expected 0 peers, got: %d", len(resp.Peers))
	}
}

// TestGRPCServer_RejectTransfer_MissingID tests missing transfer_id validation
func TestGRPCServer_RejectTransfer_MissingID(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.RejectTransfer(ctx, &pb.RejectTransferRequest{
		TransferId: "",
	})

	if err == nil {
		t.Fatal("Expected error for empty transfer_id")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error, got: %v", st.Code())
	}
}

// TestGRPCServer_CancelTransfer_MissingID tests missing transfer_id validation
func TestGRPCServer_CancelTransfer_MissingID(t *testing.T) {
	server, addr := newTestGRPCServer(t)
	defer server.stop()

	token, _ := LoadClientTokenFromPath(server.ipcAuth.GetTokenPath())
	conn, _ := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	defer conn.Close()

	client := pb.NewTransferServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.CancelTransfer(ctx, &pb.CancelTransferRequest{
		TransferId: "",
	})

	if err == nil {
		t.Fatal("Expected error for empty transfer_id")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error, got: %v", st.Code())
	}
}
