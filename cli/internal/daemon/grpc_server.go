package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	pb "github.com/orhaniscoding/goconnect/cli/internal/proto"
	"github.com/orhaniscoding/goconnect/cli/internal/voice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCServer wraps the gRPC server for the daemon.
type GRPCServer struct {
	pb.UnimplementedDaemonServiceServer
	pb.UnimplementedNetworkServiceServer
	pb.UnimplementedPeerServiceServer
	pb.UnimplementedChatServiceServer
	pb.UnimplementedTransferServiceServer
	pb.UnimplementedSettingsServiceServer
	pb.UnimplementedVoiceServiceServer

	daemon     *DaemonService
	grpcServer *grpc.Server
	listener   net.Listener
	logf       service.Logger
	ipcAuth    *IPCAuth

	// Event subscribers
	subscribers   map[chan *pb.DaemonEvent]struct{}
	subscribersMu sync.RWMutex

	// Version info
	version   string
	buildDate string
	commit    string
}

// NewGRPCServer creates a new gRPC server for the daemon.
func NewGRPCServer(daemon *DaemonService, version, buildDate, commit string) *GRPCServer {
	var ipcAuth *IPCAuth
	if daemon.config != nil && daemon.config.Daemon.IPCTokenPath != "" {
		ipcAuth = NewIPCAuthWithPath(daemon.config.Daemon.IPCTokenPath)
	} else {
		ipcAuth = NewIPCAuth()
	}

	return &GRPCServer{
		daemon:      daemon,
		logf:        daemon.logf,
		subscribers: make(map[chan *pb.DaemonEvent]struct{}),
		version:     version,
		buildDate:   buildDate,
		commit:      commit,
		ipcAuth:     ipcAuth,
	}
}

// Start starts the gRPC server.
func (s *GRPCServer) Start(ctx context.Context) error {
	// Generate and save IPC auth token
	if err := s.ipcAuth.GenerateAndSave(); err != nil {
		return fmt.Errorf("failed to initialize IPC auth: %w", err)
	}
	s.logf.Infof("IPC auth token saved to: %s", s.ipcAuth.GetTokenPath())

	// Create listener based on OS
	var err error
	s.listener, err = s.createListener()
	if err != nil {
		_ = s.ipcAuth.Cleanup() // Clean up token on failure
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Create gRPC server with chained interceptors (auth + logging)
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.ipcAuth.UnaryServerInterceptor(),
			s.loggingUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			s.ipcAuth.StreamServerInterceptor(),
			s.loggingStreamInterceptor,
		),
	)

	// Register all services
	pb.RegisterDaemonServiceServer(s.grpcServer, s)
	pb.RegisterNetworkServiceServer(s.grpcServer, s)
	pb.RegisterPeerServiceServer(s.grpcServer, s)
	pb.RegisterChatServiceServer(s.grpcServer, s)
	pb.RegisterTransferServiceServer(s.grpcServer, s)
	pb.RegisterSettingsServiceServer(s.grpcServer, s)
	pb.RegisterVoiceServiceServer(s.grpcServer, s)

	// Start server
	go func() {
		s.logf.Infof("gRPC server listening on %s", s.listener.Addr().String())
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.logf.Errorf("gRPC server error: %v", err)
		}
	}()

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	return nil
}

// Stop gracefully stops the gRPC server.
func (s *GRPCServer) Stop() {
	if s.grpcServer != nil {
		s.logf.Info("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
	// Clean up IPC auth token
	if s.ipcAuth != nil {
		if err := s.ipcAuth.Cleanup(); err != nil {
			s.logf.Warningf("Failed to cleanup IPC auth token: %v", err)
		}
	}
}

// createListener creates a platform-specific listener.
// On Linux/macOS, we create BOTH Unix socket (for CLI) and TCP (for Desktop app).
func (s *GRPCServer) createListener() (net.Listener, error) {
	switch runtime.GOOS {
	case "windows":
		// Use Windows Named Pipes for secure IPC
		// Named Pipes provide better security than TCP:
		// - Built-in Windows security descriptors
		// - Can verify client process identity
		// - No network exposure
		if IsPipeSupported() {
			listener, err := CreateWindowsListener()
			if err == nil {
				s.logf.Infof("Using Windows Named Pipe: %s", PipeName)
				return listener, nil
			}
			s.logf.Warningf("Named Pipe creation failed, falling back to TCP: %v", err)
		}
		// Fallback to TCP on localhost if Named Pipes fail
		port := s.daemon.config.Daemon.LocalPort + 1 // Use next port for gRPC
		return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	default:

		// Use Unix Domain Socket for Linux/macOS (primary - for CLI clients)
		socketPath := s.daemon.config.Daemon.SocketPath
		if socketPath == "" {
			socketPath = "/tmp/goconnect.sock"
		}
		
		// Cleanup stale socket file
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			s.logf.Warningf("Failed to remove stale socket file: %v", err)
		}

		// Also start TCP listener for Desktop app compatibility ONLY if enabled
		if s.daemon.config.Daemon.EnableDesktopIPC {
			go s.startTCPFallbackListener()
		}

		l, err := net.Listen("unix", socketPath)
		if err != nil {
			return nil, err
		}

		// Secure the socket with 0600 permissions immediately
		if err := os.Chmod(socketPath, 0600); err != nil {
			l.Close()
			return nil, fmt.Errorf("failed to set socket permissions: %w", err)
		}

		return l, nil
	}
}

// startTCPFallbackListener starts a secondary TCP listener for Desktop app compatibility.
// This runs alongside the Unix socket listener on Linux/macOS.
func (s *GRPCServer) startTCPFallbackListener() {
	const tcpPort = 34101 // Fixed port for Desktop app

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	if err != nil {
		s.logf.Warningf("TCP fallback listener failed (Desktop may not connect): %v", err)
		return
	}
	s.logf.Infof("TCP fallback listener for Desktop app on 127.0.0.1:%d", tcpPort)

	// Create a separate gRPC server for TCP (reuses same handlers)
	// We need to share the same service implementations
	tcpServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.ipcAuth.UnaryServerInterceptor(),
			s.loggingUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			s.ipcAuth.StreamServerInterceptor(),
			s.loggingStreamInterceptor,
		),
	)

	// Register all services on TCP server too
	pb.RegisterDaemonServiceServer(tcpServer, s)
	pb.RegisterNetworkServiceServer(tcpServer, s)
	pb.RegisterPeerServiceServer(tcpServer, s)
	pb.RegisterChatServiceServer(tcpServer, s)
	pb.RegisterTransferServiceServer(tcpServer, s)
	pb.RegisterSettingsServiceServer(tcpServer, s)
	pb.RegisterVoiceServiceServer(tcpServer, s)

	if err := tcpServer.Serve(listener); err != nil {
		s.logf.Warningf("TCP fallback server error: %v", err)
	}
}

// loggingUnaryInterceptor handles logging for unary RPCs.
func (s *GRPCServer) loggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	s.logf.Infof("gRPC %s took %v", info.FullMethod, time.Since(start))

	return resp, err
}

// loggingStreamInterceptor handles logging for streaming RPCs.
func (s *GRPCServer) loggingStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	s.logf.Infof("gRPC stream started: %s", info.FullMethod)

	err := handler(srv, ss)

	s.logf.Infof("gRPC stream ended: %s", info.FullMethod)

	return err
}

// =============================================================================
// DAEMON SERVICE IMPLEMENTATION
// =============================================================================

// GetStatus returns the current daemon status.
func (s *GRPCServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	engineStatus := s.daemon.engine.GetStatus()

	resp := &pb.GetStatusResponse{
		VirtualIp:   getStringValue(engineStatus, "virtual_ip"),
		ActivePeers: int32(getIntValue(engineStatus, "active_peers")),
	}

	// Map connection status
	if connected, ok := engineStatus["connected"].(bool); ok && connected {
		resp.Status = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED
	} else if connecting, ok := engineStatus["connecting"].(bool); ok && connecting {
		resp.Status = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTING
	} else {
		resp.Status = pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED
	}

	if networkID, ok := engineStatus["network_id"].(string); ok {
		resp.CurrentNetworkId = networkID
	}
	if networkName, ok := engineStatus["network_name"].(string); ok {
		resp.CurrentNetworkName = networkName
	}

	return resp, nil
}

// GetVersion returns daemon version information.
func (s *GRPCServer) GetVersion(ctx context.Context, req *emptypb.Empty) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version:   s.version,
		BuildDate: s.buildDate,
		Commit:    s.commit,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}, nil
}

// Shutdown gracefully stops the daemon.
func (s *GRPCServer) Shutdown(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	s.logf.Info("Shutdown requested via gRPC")

	// Signal daemon to stop
	go func() {
		time.Sleep(100 * time.Millisecond) // Allow response to be sent
		if s.daemon.cancel != nil {
			s.daemon.cancel()
		}
	}()

	return &emptypb.Empty{}, nil
}

// Subscribe streams daemon events to the client.
func (s *GRPCServer) Subscribe(req *pb.SubscribeRequest, stream pb.DaemonService_SubscribeServer) error {
	eventChan := make(chan *pb.DaemonEvent, 100)

	// Register subscriber
	s.subscribersMu.Lock()
	s.subscribers[eventChan] = struct{}{}
	s.subscribersMu.Unlock()

	// Cleanup on exit
	defer func() {
		s.subscribersMu.Lock()
		delete(s.subscribers, eventChan)
		s.subscribersMu.Unlock()
		close(eventChan)
	}()

	// Stream events
	for {
		select {
		case event := <-eventChan:
			// Filter events if specific types requested
			if len(req.EventTypes) > 0 {
				found := false
				for _, t := range req.EventTypes {
					if t == event.Type {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if err := stream.Send(event); err != nil {
				return err
			}

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// BroadcastEvent sends an event to all subscribers.
func (s *GRPCServer) BroadcastEvent(event *pb.DaemonEvent) {
	event.Timestamp = timestamppb.Now()

	s.subscribersMu.RLock()
	defer s.subscribersMu.RUnlock()

	for ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// =============================================================================
// NETWORK SERVICE IMPLEMENTATION
// =============================================================================

// CreateNetwork creates a new virtual network.
func (s *GRPCServer) CreateNetwork(ctx context.Context, req *pb.CreateNetworkRequest) (*pb.CreateNetworkResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "network name is required")
	}

	result, err := s.daemon.engine.CreateNetwork(req.Name, req.Cidr)
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

// JoinNetwork joins an existing network via invite code.
func (s *GRPCServer) JoinNetwork(ctx context.Context, req *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error) {
	if req.InviteCode == "" {
		return nil, status.Error(codes.InvalidArgument, "invite code is required")
	}

	result, err := s.daemon.engine.JoinNetwork(req.InviteCode)
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

// LeaveNetwork disconnects from a network.
func (s *GRPCServer) LeaveNetwork(ctx context.Context, req *pb.LeaveNetworkRequest) (*pb.LeaveNetworkResponse, error) {
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	if err := s.daemon.engine.LeaveNetwork(req.NetworkId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to leave network: %v", err)
	}

	return &pb.LeaveNetworkResponse{}, nil
}

// ListNetworks returns all networks the user is part of.
func (s *GRPCServer) ListNetworks(ctx context.Context, req *emptypb.Empty) (*pb.ListNetworksResponse, error) {
	networks, err := s.daemon.engine.GetNetworks()
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

// GetNetwork returns details of a specific network.
func (s *GRPCServer) GetNetwork(ctx context.Context, req *pb.GetNetworkRequest) (*pb.Network, error) {
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	result, err := s.daemon.engine.GetNetwork(req.NetworkId)
	if err != nil {
		if err.Error() == "network not found" {
			return nil, status.Error(codes.NotFound, "network not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get network: %v", err)
	}

	return &pb.Network{
		Id:   result.ID,
		Name: result.Name,
	}, nil
}

// DeleteNetwork deletes a network (owner only).
func (s *GRPCServer) DeleteNetwork(ctx context.Context, req *pb.DeleteNetworkRequest) (*emptypb.Empty, error) {
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	if err := s.daemon.engine.DeleteNetwork(req.NetworkId); err != nil {
		if err.Error() == "network not found" {
			return nil, status.Error(codes.NotFound, "network not found")
		}
		if err.Error() == "only the network owner can delete the network" {
			return nil, status.Error(codes.PermissionDenied, "only the network owner can delete the network")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete network: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// GenerateInvite creates an invite code for a network.
func (s *GRPCServer) GenerateInvite(ctx context.Context, req *pb.GenerateInviteRequest) (*pb.GenerateInviteResponse, error) {
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	maxUses := int(req.MaxUses)
	expiresHours := int(req.ExpiresHours)

	invite, err := s.daemon.engine.GenerateInvite(req.NetworkId, maxUses, expiresHours)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate invite: %v", err)
	}

	resp := &pb.GenerateInviteResponse{
		InviteCode: invite.Token,
		InviteUrl:  invite.InviteURL,
	}
	if !invite.ExpiresAt.IsZero() {
		resp.ExpiresAt = timestamppb.New(invite.ExpiresAt)
	}

	return resp, nil
}

// =============================================================================
// PEER SERVICE IMPLEMENTATION
// =============================================================================

// GetPeers returns all peers in the current network.
func (s *GRPCServer) GetPeers(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
	engineStatus := s.daemon.engine.GetStatus()

	pbPeers := make([]*pb.Peer, 0)

	// Get peers from engine status
	if p2pStatus, ok := engineStatus["p2p"].(map[string]interface{}); ok {
		for peerID, statusData := range p2pStatus {
			pbPeer := &pb.Peer{
				Id:     peerID,
				Status: pb.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED,
			}

			if peerStatus, ok := statusData.(map[string]interface{}); ok {
				if connected, ok := peerStatus["connected"].(bool); ok && connected {
					pbPeer.Status = pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED
				}
				if relay, ok := peerStatus["relay"].(bool); ok && relay {
					pbPeer.ConnectionType = pb.ConnectionType_CONNECTION_TYPE_RELAY
				} else {
					pbPeer.ConnectionType = pb.ConnectionType_CONNECTION_TYPE_DIRECT
				}
			}

			pbPeers = append(pbPeers, pbPeer)
		}
	}

	return &pb.GetPeersResponse{Peers: pbPeers}, nil
}

// GetPeer returns details of a specific peer.
func (s *GRPCServer) GetPeer(ctx context.Context, req *pb.GetPeerRequest) (*pb.Peer, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}

	peer, ok := s.daemon.engine.GetPeerByID(req.PeerId)
	if !ok {
		return nil, status.Error(codes.NotFound, "peer not found")
	}

	return &pb.Peer{
		Id:          peer.ID,
		Name:        peer.Name,
		DisplayName: peer.Hostname,
		VirtualIp:   getFirstIP(peer.AllowedIPs),
		Status:      pb.ConnectionStatus_CONNECTION_STATUS_CONNECTED,
	}, nil
}

// KickPeer removes a peer from the network.
func (s *GRPCServer) KickPeer(ctx context.Context, req *pb.KickPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	if err := s.daemon.engine.KickPeer(req.NetworkId, req.PeerId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to kick peer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// BanPeer permanently bans a peer from the network.
func (s *GRPCServer) BanPeer(ctx context.Context, req *pb.BanPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	if err := s.daemon.engine.BanPeer(req.NetworkId, req.PeerId, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ban peer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// UnbanPeer removes a ban.
func (s *GRPCServer) UnbanPeer(ctx context.Context, req *pb.UnbanPeerRequest) (*emptypb.Empty, error) {
	if req.PeerId == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id is required")
	}
	if req.NetworkId == "" {
		return nil, status.Error(codes.InvalidArgument, "network_id is required")
	}

	if err := s.daemon.engine.UnbanPeer(req.NetworkId, req.PeerId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unban peer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// =============================================================================
// CHAT SERVICE IMPLEMENTATION
// =============================================================================

// SendMessage sends a chat message.
func (s *GRPCServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	err := s.daemon.engine.SendChatMessage(req.RecipientId, req.Content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	return &pb.SendMessageResponse{
		Message: &pb.ChatMessage{
			Content: req.Content,
			SentAt:  timestamppb.Now(),
		},
	}, nil
}

// GetMessages retrieves chat history.
func (s *GRPCServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}

	messages := s.daemon.engine.GetChatMessages(req.NetworkId, limit, req.BeforeId)

	pbMessages := make([]*pb.ChatMessage, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.ChatMessage{
			Id:       msg.ID,
			SenderId: msg.From,
			Content:  msg.Content,
			SentAt:   timestamppb.New(msg.Time),
		}
	}

	return &pb.GetMessagesResponse{
		Messages: pbMessages,
		HasMore:  len(messages) == limit,
	}, nil
}

// SubscribeMessages streams incoming messages.
func (s *GRPCServer) SubscribeMessages(req *pb.SubscribeMessagesRequest, stream pb.ChatService_SubscribeMessagesServer) error {
	// Subscribe to chat manager
	msgChan := s.daemon.engine.SubscribeChatMessages()
	defer s.daemon.engine.UnsubscribeChatMessages(msgChan)

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			// Filter by network if specified
			if req.NetworkId != "" && msg.NetworkID != req.NetworkId {
				continue
			}

			pbMsg := &pb.ChatMessage{
				Id:       msg.ID,
				SenderId: msg.From,
				Content:  msg.Content,
				SentAt:   timestamppb.New(msg.Time),
			}

			if err := stream.Send(pbMsg); err != nil {
				return err
			}

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// =============================================================================
// TRANSFER SERVICE IMPLEMENTATION
// =============================================================================

// SendFile initiates a file transfer.
func (s *GRPCServer) SendFile(ctx context.Context, req *pb.SendFileRequest) (*pb.SendFileResponse, error) {
	if req.PeerId == "" || req.FilePath == "" {
		return nil, status.Error(codes.InvalidArgument, "peer_id and file_path are required")
	}

	session, err := s.daemon.engine.SendFileRequest(req.PeerId, req.FilePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send file: %v", err)
	}

	return &pb.SendFileResponse{TransferId: session.ID}, nil
}

// AcceptTransfer accepts an incoming file transfer.
func (s *GRPCServer) AcceptTransfer(ctx context.Context, req *pb.AcceptTransferRequest) (*emptypb.Empty, error) {
	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer_id is required")
	}

	err := s.daemon.engine.AcceptFile(req.TransferId, req.SavePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to accept transfer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// RejectTransfer rejects an incoming file transfer.
func (s *GRPCServer) RejectTransfer(ctx context.Context, req *pb.RejectTransferRequest) (*emptypb.Empty, error) {
	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer_id is required")
	}

	if err := s.daemon.engine.RejectTransfer(req.TransferId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reject transfer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// CancelTransfer cancels an ongoing transfer.
func (s *GRPCServer) CancelTransfer(ctx context.Context, req *pb.CancelTransferRequest) (*emptypb.Empty, error) {
	if req.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer_id is required")
	}

	if err := s.daemon.engine.CancelTransfer(req.TransferId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel transfer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListTransfers returns all active/recent transfers.
func (s *GRPCServer) ListTransfers(ctx context.Context, req *emptypb.Empty) (*pb.ListTransfersResponse, error) {
	transfers := s.daemon.engine.GetTransfers()

	pbTransfers := make([]*pb.FileTransfer, len(transfers))
	for i, t := range transfers {
		pbTransfers[i] = &pb.FileTransfer{
			Id:               t.ID,
			PeerId:           t.PeerID,
			Filename:         t.FileName,
			SizeBytes:        t.FileSize,
			TransferredBytes: t.SentBytes,
			Status:           mapTransferStatus(string(t.Status)),
			IsIncoming:       !t.IsSender,
		}
	}

	return &pb.ListTransfersResponse{Transfers: pbTransfers}, nil
}

// SubscribeTransfers streams transfer progress updates.
func (s *GRPCServer) SubscribeTransfers(req *emptypb.Empty, stream pb.TransferService_SubscribeTransfersServer) error {
	// Subscribe to transfer manager
	transferChan := s.daemon.engine.SubscribeTransfers()
	defer s.daemon.engine.UnsubscribeTransfers(transferChan)

	for {
		select {
		case session, ok := <-transferChan:
			if !ok {
				return nil
			}

			event := &pb.TransferEvent{
				Transfer: &pb.FileTransfer{
					Id:               session.ID,
					PeerId:           session.PeerID,
					Filename:         session.FileName,
					SizeBytes:        session.FileSize,
					TransferredBytes: session.SentBytes,
					Status:           mapTransferStatus(string(session.Status)),
					IsIncoming:       !session.IsSender,
				},
			}

			if err := stream.Send(event); err != nil {
				return err
			}

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// =============================================================================
// SETTINGS SERVICE IMPLEMENTATION
// =============================================================================

// GetSettings returns current daemon settings.
func (s *GRPCServer) GetSettings(ctx context.Context, req *emptypb.Empty) (*pb.Settings, error) {
	cfg := s.daemon.config
	return &pb.Settings{
		AutoConnect:          cfg.Settings.AutoConnect,
		NotificationsEnabled: cfg.Settings.NotificationsEnabled,
		DownloadPath:         cfg.Settings.DownloadPath,
	}, nil
}

// UpdateSettings updates daemon settings.
func (s *GRPCServer) UpdateSettings(ctx context.Context, req *pb.UpdateSettingsRequest) (*pb.Settings, error) {
	if req.Settings == nil {
		return nil, status.Error(codes.InvalidArgument, "settings are required")
	}

	cfg := s.daemon.config
	cfg.Settings.AutoConnect = req.Settings.AutoConnect
	cfg.Settings.NotificationsEnabled = req.Settings.NotificationsEnabled
	if req.Settings.DownloadPath != "" {
		cfg.Settings.DownloadPath = req.Settings.DownloadPath
	}

	// Save config to disk
	if cfg.ConfigPath != "" {
		if err := cfg.Save(cfg.ConfigPath); err != nil {
			s.logf.Warningf("Failed to save config: %v", err)
		}
	}

	return s.GetSettings(ctx, &emptypb.Empty{})
}

// ResetSettings resets settings to defaults.
func (s *GRPCServer) ResetSettings(ctx context.Context, req *emptypb.Empty) (*pb.Settings, error) {
	cfg := s.daemon.config
	cfg.Settings.AutoConnect = false
	cfg.Settings.NotificationsEnabled = true
	// Keep DownloadPath as it's system-specific

	// Save config
	if cfg.ConfigPath != "" {
		if err := cfg.Save(cfg.ConfigPath); err != nil {
			s.logf.Warningf("Failed to save config: %v", err)
		}
	}

	return s.GetSettings(ctx, &emptypb.Empty{})
}

// =============================================================================
// VOICE SERVICE IMPLEMENTATION
// =============================================================================

// SendSignal routes a WebRTC signal to a peer.
func (s *GRPCServer) SendSignal(ctx context.Context, req *pb.SendSignalRequest) (*emptypb.Empty, error) {
	if req.Signal == nil {
		return nil, status.Error(codes.InvalidArgument, "signal is required")
	}

	sig := voice.Signal{
		Type:      req.Signal.Type,
		SDP:       req.Signal.Sdp,
		Candidate: req.Signal.Candidate,
		SenderID:  req.Signal.SenderId,
		TargetID:  req.Signal.TargetId,
		NetworkID: req.Signal.NetworkId,
	}

	if err := s.daemon.engine.SendVoiceSignal(req.Signal.TargetId, sig); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send signal: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// SubscribeSignals streams incoming WebRTC signals.
func (s *GRPCServer) SubscribeSignals(req *emptypb.Empty, stream pb.VoiceService_SubscribeSignalsServer) error {
	ch := s.daemon.engine.SubscribeVoiceSignals()
	defer s.daemon.engine.UnsubscribeVoiceSignals(ch)

	for {
		select {
		case sig, ok := <-ch:
			if !ok {
				return nil
			}

			pbSig := &pb.VoiceSignal{
				Type:      sig.Type,
				Sdp:       sig.SDP,
				Candidate: sig.Candidate,
				SenderId:  sig.SenderID,
				TargetId:  sig.TargetID,
				NetworkId: sig.NetworkID,
			}

			if err := stream.Send(pbSig); err != nil {
				return err
			}

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	return 0
}

func getFirstIP(ips []string) string {
	if len(ips) == 0 {
		return ""
	}
	// Strip CIDR notation if present
	ip := ips[0]
	if idx := strings.Index(ip, "/"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

func mapTransferStatus(status string) pb.TransferStatus {
	switch status {
	case "pending":
		return pb.TransferStatus_TRANSFER_STATUS_PENDING
	case "in_progress":
		return pb.TransferStatus_TRANSFER_STATUS_IN_PROGRESS
	case "completed":
		return pb.TransferStatus_TRANSFER_STATUS_COMPLETED
	case "failed":
		return pb.TransferStatus_TRANSFER_STATUS_FAILED
	case "cancelled":
		return pb.TransferStatus_TRANSFER_STATUS_CANCELLED
	default:
		return pb.TransferStatus_TRANSFER_STATUS_UNSPECIFIED
	}
}
