package tui

import (
	"context"
	"fmt"
	"net"
	"testing"

	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

// MockDaemonServer implements all required gRPC services for testing
type MockDaemonServer struct {
	pb.UnimplementedDaemonServiceServer
	pb.UnimplementedNetworkServiceServer
	pb.UnimplementedPeerServiceServer
	pb.UnimplementedChatServiceServer
	pb.UnimplementedTransferServiceServer
	pb.UnimplementedSettingsServiceServer

	// Test hooks
	GetStatusFunc     func(context.Context, *pb.GetStatusRequest) (*pb.GetStatusResponse, error)
	CreateNetworkFunc func(context.Context, *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error)
	JoinNetworkFunc   func(context.Context, *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error)
	ListNetworksFunc  func(context.Context, *emptypb.Empty) (*pb.ListNetworksResponse, error)
	GetPeersFunc      func(context.Context, *pb.GetPeersRequest) (*pb.GetPeersResponse, error)
}

func (s *MockDaemonServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	if s.GetStatusFunc != nil {
		return s.GetStatusFunc(ctx, req)
	}
	return &pb.GetStatusResponse{Status: pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED}, nil
}

func (s *MockDaemonServer) CreateNetwork(ctx context.Context, req *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error) {
	if s.CreateNetworkFunc != nil {
		return s.CreateNetworkFunc(ctx, req)
	}
	return &pb.CreateNetworkResponse{}, nil
}

func (s *MockDaemonServer) JoinNetwork(ctx context.Context, req *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error) {
	if s.JoinNetworkFunc != nil {
		return s.JoinNetworkFunc(ctx, req)
	}
	return &pb.JoinNetworkResponse{}, nil
}

func (s *MockDaemonServer) ListNetworks(ctx context.Context, req *emptypb.Empty) (*pb.ListNetworksResponse, error) {
	if s.ListNetworksFunc != nil {
		return s.ListNetworksFunc(ctx, req)
	}
	return &pb.ListNetworksResponse{}, nil
}

func (s *MockDaemonServer) GetPeers(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
	if s.GetPeersFunc != nil {
		return s.GetPeersFunc(ctx, req)
	}
	return &pb.GetPeersResponse{}, nil
}

// setupMockGRPCServer sets up a bufconn listener and a gRPC server
func setupMockGRPCServer(t *testing.T) (*grpc.Server, *bufconn.Listener, *MockDaemonServer) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	mock := &MockDaemonServer{}

	pb.RegisterDaemonServiceServer(s, mock)
	pb.RegisterNetworkServiceServer(s, mock)
	pb.RegisterPeerServiceServer(s, mock)
	pb.RegisterChatServiceServer(s, mock)
	pb.RegisterTransferServiceServer(s, mock)
	pb.RegisterSettingsServiceServer(s, mock)

	go func() {
		if err := s.Serve(lis); err != nil {
			// We can't t.Fatal here properly as it's a goroutine, but it shouldn't happen in tests usually
			panic(fmt.Sprintf("Server exited with error: %v", err))
		}
	}()

	return s, lis, mock
}

// setupTestGRPCClient creates a client connected to the mock server
func setupTestGRPCClient(t *testing.T, mock *MockDaemonServer, lis *bufconn.Listener) *GRPCClient {
	// Override helpers
	loadAuthTokenFunc = func() (string, error) {
		return "test-token", nil
	}
	getGRPCTargetFunc = func() string {
		return "bufnet"
	}

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Connect
	client, err := NewGRPCClient(
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Override any auth creds if needed, but NewGRPCClient adds token creds.
		// Since our mock server doesn't use interceptors to check tokens, sending token creds is fine.
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}

	return client
}

func TestNewGRPCClient(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	if client == nil {
		t.Fatal("Client is nil")
	}
}

func TestGRPCClient_GetStatus(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetStatusFunc = func(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
		return &pb.GetStatusResponse{
			Status:             pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
			CurrentNetworkName: "Test Net",
			VirtualIp:          "10.0.0.1",
			ActivePeers:        5,
		}, nil
	}
	// Also mock GetNetworks as GetStatus calls it
	mock.ListNetworksFunc = func(ctx context.Context, req *emptypb.Empty) (*pb.ListNetworksResponse, error) {
		return &pb.ListNetworksResponse{Networks: []*pb.Network{}}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	status, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.Connected {
		t.Error("Expected connected status")
	}
	if status.NetworkName != "Test Net" {
		t.Errorf("Expected network name 'Test Net', got %s", status.NetworkName)
	}
	if status.IP != "10.0.0.1" {
		t.Errorf("Expected IP '10.0.0.1', got %s", status.IP)
	}
	if status.OnlineMembers != 5 {
		t.Errorf("Expected 5 online members, got %d", status.OnlineMembers)
	}
}

func TestGRPCClient_CreateNetwork(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.CreateNetworkFunc = func(ctx context.Context, req *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error) {
		if req.Name != "New Net" {
			return nil, fmt.Errorf("unexpected name: %s", req.Name)
		}
		return &pb.CreateNetworkResponse{
			Network: &pb.Network{
				Id:     "net-123",
				Name:   req.Name,
				MyRole: pb.NetworkRole_NETWORK_ROLE_OWNER,
			},
		}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	net, err := client.CreateNetwork("New Net")
	if err != nil {
		t.Fatalf("CreateNetwork failed: %v", err)
	}

	if net.ID != "net-123" {
		t.Errorf("Expected ID 'net-123', got %s", net.ID)
	}
	if net.Role != "host" {
		t.Errorf("Expected role 'host', got %s", net.Role)
	}
}

func TestGRPCClient_JoinNetwork(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.JoinNetworkFunc = func(ctx context.Context, req *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error) {
		if req.InviteCode != "INVITE" {
			return nil, fmt.Errorf("unexpected invite code")
		}
		return &pb.JoinNetworkResponse{
			Network: &pb.Network{
				Id:     "net-456",
				Name:   "Joined Net",
				MyRole: pb.NetworkRole_NETWORK_ROLE_MEMBER,
			},
		}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	net, err := client.JoinNetwork("INVITE")
	if err != nil {
		t.Fatalf("JoinNetwork failed: %v", err)
	}

	if net.ID != "net-456" {
		t.Errorf("Expected ID 'net-456', got %s", net.ID)
	}
	if net.Role != "client" {
		t.Errorf("Expected role 'client', got %s", net.Role)
	}
}

func TestGRPCClient_GetNetworks(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.ListNetworksFunc = func(ctx context.Context, req *emptypb.Empty) (*pb.ListNetworksResponse, error) {
		return &pb.ListNetworksResponse{
			Networks: []*pb.Network{
				{Id: "n1", Name: "N1", MyRole: pb.NetworkRole_NETWORK_ROLE_ADMIN},
				{Id: "n2", Name: "N2", MyRole: pb.NetworkRole_NETWORK_ROLE_MEMBER},
			},
		}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	nets, err := client.GetNetworks()
	if err != nil {
		t.Fatalf("GetNetworks failed: %v", err)
	}

	if len(nets) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(nets))
	}
	if nets[0].Role != "admin" {
		t.Errorf("Expected N1 role 'admin', got %s", nets[0].Role)
	}
}

func TestGRPCClient_GetPeers(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetPeersFunc = func(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
		return &pb.GetPeersResponse{
			Peers: []*pb.Peer{
				{
					Id:             "p1",
					Name:           "Peer 1",
					Status:         pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
					ConnectionType: pb.ConnectionType_CONNECTION_TYPE_DIRECT,
					LatencyMs:      15,
				},
			},
		}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	peers, err := client.GetPeers()
	if err != nil {
		t.Fatalf("GetPeers failed: %v", err)
	}

	if len(peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(peers))
	}
	if peers[0].Status != "connected" {
		t.Errorf("Expected status 'connected', got %s", peers[0].Status)
	}
	if peers[0].ConnectionType != "direct" {
		t.Errorf("Expected connection 'direct', got %s", peers[0].ConnectionType)
	}
}
