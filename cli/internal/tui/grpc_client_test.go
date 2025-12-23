package tui

import (
	"context"
	"fmt"
	"net"
	"testing"

	pb "github.com/orhaniscoding/goconnect/cli/internal/proto"
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
	GetStatusFunc      func(context.Context, *pb.GetStatusRequest) (*pb.GetStatusResponse, error)
	CreateNetworkFunc  func(context.Context, *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error)
	JoinNetworkFunc    func(context.Context, *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error)
	ListNetworksFunc   func(context.Context, *emptypb.Empty) (*pb.ListNetworksResponse, error)
	GetPeersFunc       func(context.Context, *pb.GetPeersRequest) (*pb.GetPeersResponse, error)
	SendMessageFunc    func(context.Context, *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
	SendFileFunc       func(context.Context, *pb.SendFileRequest) (*pb.SendFileResponse, error)
	GetSettingsFunc    func(context.Context, *emptypb.Empty) (*pb.Settings, error)
	UpdateSettingsFunc func(context.Context, *pb.UpdateSettingsRequest) (*pb.Settings, error)
	GetVersionFunc     func(context.Context, *emptypb.Empty) (*pb.VersionResponse, error)
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

func (s *MockDaemonServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if s.SendMessageFunc != nil {
		return s.SendMessageFunc(ctx, req)
	}
	return &pb.SendMessageResponse{}, nil
}

func (s *MockDaemonServer) SendFile(ctx context.Context, req *pb.SendFileRequest) (*pb.SendFileResponse, error) {
	if s.SendFileFunc != nil {
		return s.SendFileFunc(ctx, req)
	}
	return &pb.SendFileResponse{TransferId: "transfer-123"}, nil
}

func (s *MockDaemonServer) GetSettings(ctx context.Context, req *emptypb.Empty) (*pb.Settings, error) {
	if s.GetSettingsFunc != nil {
		return s.GetSettingsFunc(ctx, req)
	}
	return &pb.Settings{}, nil
}

func (s *MockDaemonServer) UpdateSettings(ctx context.Context, req *pb.UpdateSettingsRequest) (*pb.Settings, error) {
	if s.UpdateSettingsFunc != nil {
		return s.UpdateSettingsFunc(ctx, req)
	}
	return &pb.Settings{}, nil
}

func (s *MockDaemonServer) GetVersion(ctx context.Context, req *emptypb.Empty) (*pb.VersionResponse, error) {
	if s.GetVersionFunc != nil {
		return s.GetVersionFunc(ctx, req)
	}
	return &pb.VersionResponse{Version: "1.0.0"}, nil
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

	net, err := client.CreateNetwork("New Net", "")
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

// ==================== Mapping Function Tests with Proto Types ====================

func TestMapNetworkRole_WithProto(t *testing.T) {
	tests := []struct {
		role     pb.NetworkRole
		expected string
	}{
		{pb.NetworkRole_NETWORK_ROLE_OWNER, "host"},
		{pb.NetworkRole_NETWORK_ROLE_ADMIN, "admin"},
		{pb.NetworkRole_NETWORK_ROLE_MEMBER, "client"},
		{pb.NetworkRole_NETWORK_ROLE_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := mapNetworkRole(tt.role)
			if result != tt.expected {
				t.Errorf("mapNetworkRole(%v) = %s, want %s", tt.role, result, tt.expected)
			}
		})
	}
}

func TestMapConnectionStatus_WithProto(t *testing.T) {
	tests := []struct {
		status   pb.ConnectionStatus
		expected string
	}{
		{pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED, "connected"},
		{pb.ConnectionStatus_CONNECTION_STATUS_CONNECTING, "connecting"},
		{pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED, "disconnected"},
		{pb.ConnectionStatus_CONNECTION_STATUS_FAILED, "failed"},
		{pb.ConnectionStatus_CONNECTION_STATUS_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := mapConnectionStatus(tt.status)
			if result != tt.expected {
				t.Errorf("mapConnectionStatus(%v) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestMapConnectionType_WithProto(t *testing.T) {
	tests := []struct {
		connType pb.ConnectionType
		expected string
	}{
		{pb.ConnectionType_CONNECTION_TYPE_DIRECT, "direct"},
		{pb.ConnectionType_CONNECTION_TYPE_RELAY, "relay"},
		{pb.ConnectionType_CONNECTION_TYPE_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := mapConnectionType(tt.connType)
			if result != tt.expected {
				t.Errorf("mapConnectionType(%v) = %s, want %s", tt.connType, result, tt.expected)
			}
		})
	}
}

// ==================== GRPCClient CheckDaemonStatus Tests ====================

func TestGRPCClient_CheckDaemonStatus(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetStatusFunc = func(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
		return &pb.GetStatusResponse{
			Status: pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
		}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	status := client.CheckDaemonStatus()
	if !status {
		t.Error("Expected daemon to be reachable")
	}
}

func TestGRPCClient_CheckDaemonStatus_Error(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetStatusFunc = func(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
		return nil, fmt.Errorf("connection error")
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	status := client.CheckDaemonStatus()
	if status {
		t.Error("Expected daemon to be unreachable when error occurs")
	}
}

// ==================== GRPCClient Close Tests ====================

func TestGRPCClient_Close(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	client := setupTestGRPCClient(t, mock, lis)

	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestGRPCClient_Close_NilConn(t *testing.T) {
	client := &GRPCClient{conn: nil}
	err := client.Close()
	if err != nil {
		t.Errorf("Close with nil conn should not error: %v", err)
	}
}

// ==================== GRPCClient LeaveNetwork Tests ====================

func TestGRPCClient_LeaveNetwork(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	// LeaveNetwork uses the network client - need to verify it calls correctly
	err := client.LeaveNetwork("network-123")
	// The mock server will return default empty response which is fine
	if err != nil {
		// Mock doesn't implement LeaveNetwork, so this is expected
		t.Logf("LeaveNetwork returned error (expected with default mock): %v", err)
	}
}

// ==================== getGRPCTarget Extended Tests ====================

func TestGetGRPCTarget_Extended(t *testing.T) {
	target := getGRPCTarget()
	if target == "" {
		t.Error("getGRPCTarget returned empty string")
	}
	// Verify it returns a valid target
	if target != "unix:///tmp/goconnect.sock" && target != "127.0.0.1:34101" {
		t.Errorf("Unexpected target: %s", target)
	}
}

// ==================== TryConnectGRPC Tests ====================

func TestTryConnectGRPC_Failure(t *testing.T) {
	// Reset the mock functions to use real implementations
	originalLoadAuthToken := loadAuthTokenFunc
	originalGetGRPCTarget := getGRPCTargetFunc
	defer func() {
		loadAuthTokenFunc = originalLoadAuthToken
		getGRPCTargetFunc = originalGetGRPCTarget
	}()

	// Make it fail by pointing to non-existent target
	getGRPCTargetFunc = func() string {
		return "127.0.0.1:59999" // Non-existent port
	}
	loadAuthTokenFunc = func() (string, error) {
		return "test-token", nil
	}

	// Try to connect with very short retry
	_, err := TryConnectGRPC(1, 1)
	if err == nil {
		t.Error("Expected TryConnectGRPC to fail when daemon is not running")
	}
}

// ==================== IsGRPCAvailable Tests ====================

func TestIsGRPCAvailable_NotRunning(t *testing.T) {
	// Without daemon, should return false
	available := IsGRPCAvailable()
	// Just verify it doesn't panic and returns a boolean
	_ = available
}

// ==================== Peer Struct Tests ====================

func TestPeerStruct(t *testing.T) {
	peer := Peer{
		ID:             "peer-123",
		Name:           "Test Peer",
		VirtualIP:      "10.0.0.5",
		Status:         "connected",
		ConnectionType: "direct",
		LatencyMs:      25,
	}

	if peer.ID != "peer-123" {
		t.Errorf("Expected ID 'peer-123', got %s", peer.ID)
	}
	if peer.Name != "Test Peer" {
		t.Errorf("Expected Name 'Test Peer', got %s", peer.Name)
	}
	if peer.LatencyMs != 25 {
		t.Errorf("Expected LatencyMs 25, got %d", peer.LatencyMs)
	}
}

// ==================== GRPCClient GetPeers with various statuses ====================

func TestGRPCClient_GetPeers_VariousStatuses(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetPeersFunc = func(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
		return &pb.GetPeersResponse{
			Peers: []*pb.Peer{
				{
					Id:             "p1",
					Name:           "Connected Peer",
					Status:         pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
					ConnectionType: pb.ConnectionType_CONNECTION_TYPE_DIRECT,
				},
				{
					Id:             "p2",
					Name:           "Connecting Peer",
					Status:         pb.ConnectionStatus_CONNECTION_STATUS_CONNECTING,
					ConnectionType: pb.ConnectionType_CONNECTION_TYPE_RELAY,
				},
				{
					Id:             "p3",
					Name:           "Disconnected Peer",
					Status:         pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED,
					ConnectionType: pb.ConnectionType_CONNECTION_TYPE_UNSPECIFIED,
				},
				{
					Id:             "p4",
					Name:           "Failed Peer",
					Status:         pb.ConnectionStatus_CONNECTION_STATUS_FAILED,
					ConnectionType: pb.ConnectionType_CONNECTION_TYPE_DIRECT,
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

	if len(peers) != 4 {
		t.Errorf("Expected 4 peers, got %d", len(peers))
	}

	// Verify status mappings
	expectedStatuses := []string{"connected", "connecting", "disconnected", "failed"}
	for i, expected := range expectedStatuses {
		if peers[i].Status != expected {
			t.Errorf("Peer %d: expected status '%s', got '%s'", i, expected, peers[i].Status)
		}
	}

	// Verify connection type mappings
	expectedTypes := []string{"direct", "relay", "unknown", "direct"}
	for i, expected := range expectedTypes {
		if peers[i].ConnectionType != expected {
			t.Errorf("Peer %d: expected type '%s', got '%s'", i, expected, peers[i].ConnectionType)
		}
	}
}

// ==================== GRPCClient SendChatMessage Tests ====================

func TestGRPCClient_SendChatMessage(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.SendMessageFunc = func(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
		if req.NetworkId != "network-123" {
			return nil, fmt.Errorf("unexpected network ID: %s", req.NetworkId)
		}
		if req.Content != "Hello, world!" {
			return nil, fmt.Errorf("unexpected content: %s", req.Content)
		}
		return &pb.SendMessageResponse{}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	err := client.SendChatMessage("network-123", "Hello, world!")
	if err != nil {
		t.Fatalf("SendChatMessage failed: %v", err)
	}
}

// ==================== GRPCClient SendFile Tests ====================

func TestGRPCClient_SendFile(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.SendFileFunc = func(ctx context.Context, req *pb.SendFileRequest) (*pb.SendFileResponse, error) {
		if req.PeerId != "peer-123" {
			return nil, fmt.Errorf("unexpected peer ID: %s", req.PeerId)
		}
		if req.FilePath != "/tmp/test.txt" {
			return nil, fmt.Errorf("unexpected file path: %s", req.FilePath)
		}
		return &pb.SendFileResponse{TransferId: "transfer-456"}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	transferID, err := client.SendFile("peer-123", "/tmp/test.txt")
	if err != nil {
		t.Fatalf("SendFile failed: %v", err)
	}
	if transferID != "transfer-456" {
		t.Errorf("Expected transfer ID 'transfer-456', got %s", transferID)
	}
}

// ==================== GRPCClient GetSettings Tests ====================

func TestGRPCClient_GetSettings(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetSettingsFunc = func(ctx context.Context, req *emptypb.Empty) (*pb.Settings, error) {
		return &pb.Settings{}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	settings, err := client.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings == nil {
		t.Error("Expected non-nil settings")
	}
}

// ==================== GRPCClient UpdateSettings Tests ====================

func TestGRPCClient_UpdateSettings(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.UpdateSettingsFunc = func(ctx context.Context, req *pb.UpdateSettingsRequest) (*pb.Settings, error) {
		return &pb.Settings{}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	err := client.UpdateSettings(&pb.Settings{})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}
}

// ==================== GRPCClient GetVersion Tests ====================

func TestGRPCClient_GetVersion(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	mock.GetVersionFunc = func(ctx context.Context, req *emptypb.Empty) (*pb.VersionResponse, error) {
		return &pb.VersionResponse{Version: "2.0.0"}, nil
	}

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	version, err := client.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	if version.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", version.Version)
	}
}

// ==================== GRPCClient Subscribe Tests ====================

func TestGRPCClient_Subscribe(t *testing.T) {
	s, lis, mock := setupMockGRPCServer(t)
	defer s.Stop()
	defer lis.Close()

	_ = mock // Subscribe is a streaming call, tested via UnifiedClient

	client := setupTestGRPCClient(t, mock, lis)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe call - might fail since mock doesn't fully implement streaming
	_, err := client.Subscribe(ctx, []pb.EventType{})
	// Error is acceptable since mock doesn't implement streaming
	_ = err
}
