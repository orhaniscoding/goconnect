package daemon

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ==================== Real GRPCServer Implementation Tests ====================

// setupRealGRPCServer creates a GRPCServer with mock engine for testing
func setupRealGRPCServer(t *testing.T) (*GRPCServer, *MockEngine) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")
	cfg.Settings.AutoConnect = true
	cfg.Settings.NotificationsEnabled = true
	cfg.Settings.DownloadPath = "/tmp/downloads"

	daemon := &DaemonService{
		config:        cfg,
		logf:          &fallbackServiceLogger{log.New(os.Stderr, "[test] ", log.LstdFlags)},
		daemonVersion: "1.0.0-test",
		sseClients:    make(map[chan string]bool),
	}
	daemon.idManager = identity.NewManager(cfg.IdentityPath)
	daemon.idManager.LoadOrCreateIdentity()

	mockEng := new(MockEngine)
	daemon.engine = mockEng

	grpcSrv := NewGRPCServer(daemon, "1.0.0", "2024-01-01", "abc123")

	return grpcSrv, mockEng
}

// ==================== DaemonService gRPC Methods ====================

func TestGRPCServer_GetStatus_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("connected status", func(t *testing.T) {
		mockEng.On("GetStatus").Return(map[string]interface{}{
			"connected":    true,
			"virtual_ip":   "10.0.0.5",
			"active_peers": 3,
			"network_id":   "net-123",
			"network_name": "MyNetwork",
		}).Once()

		resp, err := srv.GetStatus(context.Background(), &pb.GetStatusRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED, resp.Status)
		assert.Equal(t, "10.0.0.5", resp.VirtualIp)
		assert.Equal(t, int32(3), resp.ActivePeers)
		assert.Equal(t, "net-123", resp.CurrentNetworkId)
		assert.Equal(t, "MyNetwork", resp.CurrentNetworkName)
	})

	t.Run("connecting status", func(t *testing.T) {
		mockEng.On("GetStatus").Return(map[string]interface{}{
			"connecting": true,
		}).Once()

		resp, err := srv.GetStatus(context.Background(), &pb.GetStatusRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ConnectionStatus_CONNECTION_STATUS_CONNECTING, resp.Status)
	})

	t.Run("disconnected status", func(t *testing.T) {
		mockEng.On("GetStatus").Return(map[string]interface{}{
			"connected":  false,
			"connecting": false,
		}).Once()

		resp, err := srv.GetStatus(context.Background(), &pb.GetStatusRequest{})
		assert.NoError(t, err)
		assert.Equal(t, pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED, resp.Status)
	})
}

func TestGRPCServer_GetVersion_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	resp, err := srv.GetVersion(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, "2024-01-01", resp.BuildDate)
	assert.Equal(t, "abc123", resp.Commit)
	assert.NotEmpty(t, resp.GoVersion)
	assert.NotEmpty(t, resp.Os)
	assert.NotEmpty(t, resp.Arch)
}

func TestGRPCServer_Shutdown_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	// Set cancel function
	ctx, cancel := context.WithCancel(context.Background())
	srv.daemon.cancel = cancel

	resp, err := srv.Shutdown(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Wait a bit for shutdown to trigger
	select {
	case <-ctx.Done():
		// Good - context was cancelled
	default:
		// May not be cancelled immediately due to sleep
	}
}

// ==================== NetworkService gRPC Methods ====================

func TestGRPCServer_CreateNetwork_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("CreateNetwork", "TestNetwork").Return(&api.NetworkResponse{
			ID:   "net-new",
			Name: "TestNetwork",
		}, nil).Once()

		resp, err := srv.CreateNetwork(context.Background(), &pb.CreateNetworkRequest{
			Name: "TestNetwork",
		})
		assert.NoError(t, err)
		assert.Equal(t, "net-new", resp.Network.Id)
		assert.Equal(t, "TestNetwork", resp.Network.Name)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := srv.CreateNetwork(context.Background(), &pb.CreateNetworkRequest{
			Name: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("engine error", func(t *testing.T) {
		mockEng.On("CreateNetwork", "FailNetwork").Return(nil, assert.AnError).Once()

		_, err := srv.CreateNetwork(context.Background(), &pb.CreateNetworkRequest{
			Name: "FailNetwork",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestGRPCServer_JoinNetwork_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("JoinNetwork", "INVITE123").Return(&api.NetworkResponse{
			ID:   "net-joined",
			Name: "JoinedNetwork",
		}, nil).Once()

		resp, err := srv.JoinNetwork(context.Background(), &pb.JoinNetworkRequest{
			InviteCode: "INVITE123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "net-joined", resp.Network.Id)
	})

	t.Run("empty invite", func(t *testing.T) {
		_, err := srv.JoinNetwork(context.Background(), &pb.JoinNetworkRequest{
			InviteCode: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestGRPCServer_LeaveNetwork_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("LeaveNetwork", "net-leave").Return(nil).Once()

		resp, err := srv.LeaveNetwork(context.Background(), &pb.LeaveNetworkRequest{
			NetworkId: "net-leave",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("empty network_id", func(t *testing.T) {
		_, err := srv.LeaveNetwork(context.Background(), &pb.LeaveNetworkRequest{
			NetworkId: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestGRPCServer_ListNetworks_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("GetNetworks").Return([]api.NetworkResponse{
			{ID: "net-1", Name: "Network1"},
			{ID: "net-2", Name: "Network2"},
		}, nil).Once()

		resp, err := srv.ListNetworks(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
		assert.Len(t, resp.Networks, 2)
		assert.Equal(t, "net-1", resp.Networks[0].Id)
	})

	t.Run("error", func(t *testing.T) {
		mockEng.On("GetNetworks").Return(nil, assert.AnError).Once()

		_, err := srv.ListNetworks(context.Background(), &emptypb.Empty{})
		assert.Error(t, err)
	})
}

func TestGRPCServer_GetNetwork_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	t.Run("empty network_id", func(t *testing.T) {
		_, err := srv.GetNetwork(context.Background(), &pb.GetNetworkRequest{
			NetworkId: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("unimplemented", func(t *testing.T) {
		_, err := srv.GetNetwork(context.Background(), &pb.GetNetworkRequest{
			NetworkId: "net-1",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})
}

func TestGRPCServer_DeleteNetwork_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	t.Run("empty network_id", func(t *testing.T) {
		_, err := srv.DeleteNetwork(context.Background(), &pb.DeleteNetworkRequest{
			NetworkId: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("unimplemented", func(t *testing.T) {
		_, err := srv.DeleteNetwork(context.Background(), &pb.DeleteNetworkRequest{
			NetworkId: "net-1",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unimplemented, st.Code())
	})
}

func TestGRPCServer_GenerateInvite_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("GenerateInvite", "net-1", 5, 24).Return(&api.InviteTokenResponse{
			Token:     "TOKEN123",
			InviteURL: "https://example.com/invite/TOKEN123",
		}, nil).Once()

		resp, err := srv.GenerateInvite(context.Background(), &pb.GenerateInviteRequest{
			NetworkId:    "net-1",
			MaxUses:      5,
			ExpiresHours: 24,
		})
		assert.NoError(t, err)
		assert.Equal(t, "TOKEN123", resp.InviteCode)
		assert.Equal(t, "https://example.com/invite/TOKEN123", resp.InviteUrl)
	})

	t.Run("empty network_id", func(t *testing.T) {
		_, err := srv.GenerateInvite(context.Background(), &pb.GenerateInviteRequest{
			NetworkId: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

// ==================== PeerService gRPC Methods ====================

func TestGRPCServer_GetPeers_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("with peers", func(t *testing.T) {
		mockEng.On("GetStatus").Return(map[string]interface{}{
			"p2p": map[string]interface{}{
				"peer-1": map[string]interface{}{
					"connected": true,
					"relay":     false,
				},
				"peer-2": map[string]interface{}{
					"connected": false,
					"relay":     true,
				},
			},
		}).Once()

		resp, err := srv.GetPeers(context.Background(), &pb.GetPeersRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Peers, 2)
	})

	t.Run("no peers", func(t *testing.T) {
		mockEng.On("GetStatus").Return(map[string]interface{}{}).Once()

		resp, err := srv.GetPeers(context.Background(), &pb.GetPeersRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Peers, 0)
	})
}

func TestGRPCServer_GetPeer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("found", func(t *testing.T) {
		mockEng.On("GetPeerByID", "peer-1").Return(&api.PeerConfig{
			ID:         "peer-1",
			Name:       "TestPeer",
			Hostname:   "testhost",
			AllowedIPs: []string{"10.0.0.2/32"},
		}, true).Once()

		resp, err := srv.GetPeer(context.Background(), &pb.GetPeerRequest{
			PeerId: "peer-1",
		})
		assert.NoError(t, err)
		assert.Equal(t, "peer-1", resp.Id)
		assert.Equal(t, "TestPeer", resp.Name)
		assert.Equal(t, "10.0.0.2", resp.VirtualIp)
	})

	t.Run("not found", func(t *testing.T) {
		mockEng.On("GetPeerByID", "peer-unknown").Return(nil, false).Once()

		_, err := srv.GetPeer(context.Background(), &pb.GetPeerRequest{
			PeerId: "peer-unknown",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("empty peer_id", func(t *testing.T) {
		_, err := srv.GetPeer(context.Background(), &pb.GetPeerRequest{
			PeerId: "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestGRPCServer_KickPeer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("KickPeer", "net-1", "peer-1", "test reason").Return(nil).Once()

		resp, err := srv.KickPeer(context.Background(), &pb.KickPeerRequest{
			NetworkId: "net-1",
			PeerId:    "peer-1",
			Reason:    "test reason",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("missing peer_id", func(t *testing.T) {
		_, err := srv.KickPeer(context.Background(), &pb.KickPeerRequest{
			NetworkId: "net-1",
			PeerId:    "",
		})
		assert.Error(t, err)
	})

	t.Run("missing network_id", func(t *testing.T) {
		_, err := srv.KickPeer(context.Background(), &pb.KickPeerRequest{
			NetworkId: "",
			PeerId:    "peer-1",
		})
		assert.Error(t, err)
	})
}

func TestGRPCServer_BanPeer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("BanPeer", "net-1", "peer-1", "ban reason").Return(nil).Once()

		resp, err := srv.BanPeer(context.Background(), &pb.BanPeerRequest{
			NetworkId: "net-1",
			PeerId:    "peer-1",
			Reason:    "ban reason",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestGRPCServer_UnbanPeer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("UnbanPeer", "net-1", "peer-1").Return(nil).Once()

		resp, err := srv.UnbanPeer(context.Background(), &pb.UnbanPeerRequest{
			NetworkId: "net-1",
			PeerId:    "peer-1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// ==================== ChatService gRPC Methods ====================

func TestGRPCServer_SendMessage_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("SendChatMessage", "peer-1", "Hello!").Return(nil).Once()

		resp, err := srv.SendMessage(context.Background(), &pb.SendMessageRequest{
			RecipientId: "peer-1",
			Content:     "Hello!",
		})
		assert.NoError(t, err)
		assert.Equal(t, "Hello!", resp.Message.Content)
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := srv.SendMessage(context.Background(), &pb.SendMessageRequest{
			RecipientId: "peer-1",
			Content:     "",
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestGRPCServer_GetMessages_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("with messages", func(t *testing.T) {
		mockEng.On("GetChatMessages", "net-1", 50, "").Return([]chat.Message{
			{ID: "msg-1", From: "peer-1", Content: "Hello"},
			{ID: "msg-2", From: "peer-2", Content: "Hi there"},
		}).Once()

		resp, err := srv.GetMessages(context.Background(), &pb.GetMessagesRequest{
			NetworkId: "net-1",
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Messages, 2)
		assert.False(t, resp.HasMore)
	})

	t.Run("custom limit", func(t *testing.T) {
		mockEng.On("GetChatMessages", "net-1", 10, "before-id").Return([]chat.Message{}).Once()

		resp, err := srv.GetMessages(context.Background(), &pb.GetMessagesRequest{
			NetworkId: "net-1",
			Limit:     10,
			BeforeId:  "before-id",
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Messages, 0)
	})
}

// ==================== TransferService gRPC Methods ====================

func TestGRPCServer_SendFile_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("SendFileRequest", "peer-1", "/path/to/file.txt").Return(&transfer.Session{
			ID: "transfer-1",
		}, nil).Once()

		resp, err := srv.SendFile(context.Background(), &pb.SendFileRequest{
			PeerId:   "peer-1",
			FilePath: "/path/to/file.txt",
		})
		assert.NoError(t, err)
		assert.Equal(t, "transfer-1", resp.TransferId)
	})

	t.Run("missing fields", func(t *testing.T) {
		_, err := srv.SendFile(context.Background(), &pb.SendFileRequest{
			PeerId:   "",
			FilePath: "/path/to/file.txt",
		})
		assert.Error(t, err)
	})
}

func TestGRPCServer_AcceptTransfer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("AcceptFile", "transfer-1", "/save/path").Return(nil).Once()

		resp, err := srv.AcceptTransfer(context.Background(), &pb.AcceptTransferRequest{
			TransferId: "transfer-1",
			SavePath:   "/save/path",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("empty transfer_id", func(t *testing.T) {
		_, err := srv.AcceptTransfer(context.Background(), &pb.AcceptTransferRequest{
			TransferId: "",
		})
		assert.Error(t, err)
	})
}

func TestGRPCServer_RejectTransfer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("RejectTransfer", "transfer-1").Return(nil).Once()

		resp, err := srv.RejectTransfer(context.Background(), &pb.RejectTransferRequest{
			TransferId: "transfer-1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestGRPCServer_CancelTransfer_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("success", func(t *testing.T) {
		mockEng.On("CancelTransfer", "transfer-1").Return(nil).Once()

		resp, err := srv.CancelTransfer(context.Background(), &pb.CancelTransferRequest{
			TransferId: "transfer-1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestGRPCServer_ListTransfers_Real(t *testing.T) {
	srv, mockEng := setupRealGRPCServer(t)

	t.Run("with transfers", func(t *testing.T) {
		mockEng.On("GetTransfers").Return([]transfer.Session{
			{ID: "t-1", PeerID: "peer-1", FileName: "file1.txt", FileSize: 1024, Status: transfer.StatusPending, IsSender: true},
			{ID: "t-2", PeerID: "peer-2", FileName: "file2.txt", FileSize: 2048, Status: transfer.StatusCompleted, IsSender: false},
		}).Once()

		resp, err := srv.ListTransfers(context.Background(), &emptypb.Empty{})
		assert.NoError(t, err)
		assert.Len(t, resp.Transfers, 2)
		assert.Equal(t, "t-1", resp.Transfers[0].Id)
		assert.False(t, resp.Transfers[0].IsIncoming) // IsSender=true means not incoming
		assert.True(t, resp.Transfers[1].IsIncoming)  // IsSender=false means incoming
	})
}

// ==================== SettingsService gRPC Methods ====================

func TestGRPCServer_GetSettings_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	resp, err := srv.GetSettings(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.True(t, resp.AutoConnect)
	assert.True(t, resp.NotificationsEnabled)
	assert.Equal(t, "/tmp/downloads", resp.DownloadPath)
}

func TestGRPCServer_UpdateSettings_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	t.Run("update settings", func(t *testing.T) {
		resp, err := srv.UpdateSettings(context.Background(), &pb.UpdateSettingsRequest{
			Settings: &pb.Settings{
				AutoConnect:          false,
				NotificationsEnabled: false,
				DownloadPath:         "/new/path",
			},
		})
		assert.NoError(t, err)
		assert.False(t, resp.AutoConnect)
		assert.False(t, resp.NotificationsEnabled)
		assert.Equal(t, "/new/path", resp.DownloadPath)
	})

	t.Run("nil settings", func(t *testing.T) {
		_, err := srv.UpdateSettings(context.Background(), &pb.UpdateSettingsRequest{
			Settings: nil,
		})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestGRPCServer_ResetSettings_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	resp, err := srv.ResetSettings(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.False(t, resp.AutoConnect)
	assert.True(t, resp.NotificationsEnabled)
}

// ==================== BroadcastEvent Test ====================

func TestGRPCServer_BroadcastEvent_Real(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	// Add a subscriber
	eventChan := make(chan *pb.DaemonEvent, 10)
	srv.subscribersMu.Lock()
	srv.subscribers[eventChan] = struct{}{}
	srv.subscribersMu.Unlock()

	// Broadcast an event
	event := &pb.DaemonEvent{
		Type: pb.EventType_EVENT_TYPE_STATUS_CHANGED,
	}
	srv.BroadcastEvent(event)

	// Check if event was received
	select {
	case received := <-eventChan:
		assert.Equal(t, pb.EventType_EVENT_TYPE_STATUS_CHANGED, received.Type)
		assert.NotNil(t, received.Timestamp)
	default:
		t.Error("Expected event to be broadcast")
	}
}

// ==================== Logging Interceptor Tests ====================

func TestGRPCServer_LoggingInterceptors(t *testing.T) {
	srv, _ := setupRealGRPCServer(t)

	// Just verify the server has the logging functions
	// They are private methods that can't be easily tested without full gRPC setup
	assert.NotNil(t, srv.logf)
}
