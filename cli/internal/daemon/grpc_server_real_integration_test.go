package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/voice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// setupIntegrationServer creates a real GRPCServer with isolated file paths
func setupIntegrationServer(t *testing.T) (*GRPCServer, *MockEngine, string) {
	t.Helper()

	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "daemon.sock")
	tokenPath := filepath.Join(tmpDir, "ipc.token")
	identityPath := filepath.Join(tmpDir, "identity.json")

	cfg := &config.Config{}
	cfg.IdentityPath = identityPath
	cfg.Daemon.SocketPath = socketPath
	cfg.Daemon.IPCTokenPath = tokenPath
	cfg.Settings.AutoConnect = true
	cfg.Settings.NotificationsEnabled = true

	daemon := &DaemonService{
		config:        cfg,
		logf:          &fallbackServiceLogger{}, // Use stdout/stderr logger or mock
		daemonVersion: "1.0.0-integration",
		sseClients:    make(map[chan string]bool),
	}
	daemon.idManager = identity.NewManager(cfg.IdentityPath)
	_, _ = daemon.idManager.LoadOrCreateIdentity()

	mockEng := new(MockEngine)
	daemon.engine = mockEng

	// Mock Start/Stop for engine since daemon might call them
	mockEng.On("Start").Return().Maybe()
	mockEng.On("Stop").Return().Maybe()

	grpcSrv := NewGRPCServer(daemon, "1.0.0", "2024", "deadbeef")

	return grpcSrv, mockEng, socketPath
}

func TestGRPCServer_Integration_StartAndConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, mockEng, socketPath := setupIntegrationServer(t)

	// Context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	err := srv.Start(ctx)
	assert.NoError(t, err, "Failed to start gRPC server")
	defer srv.Stop()

	// Wait for socket to appear
	assert.Eventually(t, func() bool {
		_, err := os.Stat(socketPath)
		return err == nil
	}, 2*time.Second, 100*time.Millisecond, "Socket file was not created")

	// Verify token was created
	tokenPath := srv.daemon.config.Daemon.IPCTokenPath
	assert.FileExists(t, tokenPath)

	// Load token for client
	token, err := LoadClientTokenFromPath(tokenPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Connect client
	target := "unix://" + socketPath
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	// Make a request - verifying Interceptor logic implicitly by success
	mockEng.On("GetStatus").Return(map[string]interface{}{
		"connected":  true,
		"virtual_ip": "10.0.0.99",
	}).Once()

	resp, err := client.GetStatus(ctx, &pb.GetStatusRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.99", resp.VirtualIp)
}

func TestGRPCServer_Integration_SubscribeMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, mockEng, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup subscription channels
	msgChan := make(chan chat.Message, 10)
	mockEng.On("SubscribeChatMessages").Return(msgChan)
	mockEng.On("UnsubscribeChatMessages", msgChan).Return()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Client setup
	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	chatClient := pb.NewChatServiceClient(conn)
	stream, err := chatClient.SubscribeMessages(ctx, &pb.SubscribeMessagesRequest{})
	assert.NoError(t, err)

	// Send message to channel
	go func() {
		time.Sleep(100 * time.Millisecond)
		msgChan <- chat.Message{
			ID:      "msg-1",
			From:    "peer-A",
			Content: "Integration Test",
			Time:    time.Now(),
		}
	}()

	// Receive from stream
	received, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, "peer-A", received.SenderId)
	assert.Equal(t, "Integration Test", received.Content)
}

func TestGRPCServer_Integration_AuthFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, mockEng, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Mock GetStatus just in case (should not be reached if auth fails correct, but interceptor runs first)
	// Actually, if auth fails, handler won't be called.
	// But mocking it just to be safe if failure mode is wrong.
	mockEng.On("GetStatus").Return(map[string]interface{}{}).Maybe()

	// Connect with WRONG token
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials("bad-token")),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)

	_, err = client.GetStatus(ctx, &pb.GetStatusRequest{})
	assert.Error(t, err)
	// Logic: Interceptor should reject.
}

func TestGRPCServer_Integration_SubscribeTransfers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, mockEng, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup subscription channels
	transferChan := make(chan transfer.Session, 10)
	mockEng.On("SubscribeTransfers").Return(transferChan)
	mockEng.On("UnsubscribeTransfers", transferChan).Return()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Client setup
	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	transferClient := pb.NewTransferServiceClient(conn)
	stream, err := transferClient.SubscribeTransfers(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	// Send message to channel
	go func() {
		time.Sleep(100 * time.Millisecond)
		transferChan <- transfer.Session{
			ID:       "transfer-Z",
			PeerID:   "peer-B",
			FileName: "secret_plans.txt",
			FileSize: 1024,
			Status:   transfer.StatusPending,
		}
	}()

	// Receive from stream
	received, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, "transfer-Z", received.Transfer.Id)
	assert.Equal(t, "secret_plans.txt", received.Transfer.Filename)
}

func TestGRPCServer_Integration_Subscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, _, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Client setup
	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{})
	assert.NoError(t, err)

	// Broadcast an event
	go func() {
		time.Sleep(100 * time.Millisecond)
		srv.BroadcastEvent(&pb.DaemonEvent{
			Type: pb.EventType_EVENT_TYPE_NOTIFICATION,
			Payload: &pb.DaemonEvent_Notification{
				Notification: &pb.Notification{
					Title:   "Test",
					Message: "Hello Subscribers",
				},
			},
		})
	}()

	// Receive from stream
	received, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, pb.EventType_EVENT_TYPE_NOTIFICATION, received.Type)
	payload, ok := received.Payload.(*pb.DaemonEvent_Notification)
	assert.True(t, ok)
	assert.Equal(t, "Hello Subscribers", payload.Notification.Message)
}

func TestGRPCServer_Integration_VoiceSignals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, mockEng, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal channel
	sigChan := make(chan voice.Signal, 10)
	mockEng.On("SubscribeVoiceSignals").Return(sigChan)
	mockEng.On("UnsubscribeVoiceSignals", sigChan).Return()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Client setup
	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	voiceClient := pb.NewVoiceServiceClient(conn)

	// 1. Test SendSignal
	mockEng.On("SendVoiceSignal", "peer-1", mock.MatchedBy(func(s voice.Signal) bool {
		return s.TargetID == "peer-1" && s.SDP == "v=0..."
	})).Return(nil)

	_, err = voiceClient.SendSignal(ctx, &pb.SendSignalRequest{
		Signal: &pb.VoiceSignal{
			TargetId: "peer-1",
			Sdp:      "v=0...",
		},
	})
	assert.NoError(t, err)

	// 2. Test SubscribeSignals
	stream, err := voiceClient.SubscribeSignals(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- voice.Signal{
			SenderID: "peer-X",
			SDP:      "answer-sdp",
			Type:     "answer",
		}
	}()

	received, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, "peer-X", received.SenderId)
	assert.Equal(t, "answer-sdp", received.Sdp)
}

func TestGRPCServer_Integration_Subscribe_Filtered(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, _, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	// Wait for socket
	time.Sleep(200 * time.Millisecond)

	// Client setup
	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewDaemonServiceClient(conn)
	// Request ONLY status changed events
	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{
		EventTypes: []pb.EventType{pb.EventType_EVENT_TYPE_STATUS_CHANGED},
	})
	assert.NoError(t, err)

	// Broadcast different events
	go func() {
		time.Sleep(100 * time.Millisecond)
		// This should be filtered out
		srv.BroadcastEvent(&pb.DaemonEvent{
			Type: pb.EventType_EVENT_TYPE_NOTIFICATION,
			Payload: &pb.DaemonEvent_Notification{
				Notification: &pb.Notification{Message: "Ignored"},
			},
		})
		time.Sleep(50 * time.Millisecond)
		// This should pass
		srv.BroadcastEvent(&pb.DaemonEvent{
			Type: pb.EventType_EVENT_TYPE_STATUS_CHANGED,
			Payload: &pb.DaemonEvent_StatusChanged{
				StatusChanged: &pb.StatusChangedEvent{
					NewStatus: pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
					NetworkId: "net-123",
				},
			},
		})
	}()

	// Receive from stream - should skip Notification and get StatusChanged
	received, err := stream.Recv()
	assert.NoError(t, err)
	assert.Equal(t, pb.EventType_EVENT_TYPE_STATUS_CHANGED, received.Type)
}

func TestGRPCServer_Integration_VoiceSignals_Errors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	srv, _, socketPath := setupIntegrationServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	defer srv.Stop()

	time.Sleep(200 * time.Millisecond)

	token, _ := LoadClientTokenFromPath(srv.daemon.config.Daemon.IPCTokenPath)
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(NewTokenCredentials(token)),
	)
	assert.NoError(t, err)
	defer conn.Close()

	voiceClient := pb.NewVoiceServiceClient(conn)

	// 1. Missing signal object
	_, err = voiceClient.SendSignal(ctx, &pb.SendSignalRequest{
		Signal: nil,
	})
	assert.Error(t, err)
}
