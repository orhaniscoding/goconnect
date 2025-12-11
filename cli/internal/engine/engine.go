package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/p2p"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/system"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/wireguard"
	"github.com/pion/ice/v2"
)

// Version is injected at build time
var Version = "dev"

// WireGuardClient defines the interface for WireGuard operations
type WireGuardClient interface {
	Down() error
	ApplyConfig(config *api.DeviceConfig, privateKey string) error
	GetStatus() (*wireguard.Status, error)
}

// Engine is the core logic controller
type Engine struct {
	config        *config.Config
	idMgr         *identity.Manager
	apiClient     *api.Client
	wgClient      WireGuardClient
	p2pMgr        *p2p.Manager
	chatMgr       *chat.Manager
	transferMgr   *transfer.Manager
	sysConf       system.Configurator
	hostsMgr      *system.HostsManager
	stopChan      chan struct{}
	logf          service.Logger
	daemonVersion string
	paused        bool

	// State
	peerMap  map[string]api.PeerConfig
	networks []api.NetworkResponse
	mu       sync.RWMutex

	onChatMessage      func(chat.Message)
	onTransferProgress func(transfer.Session)
	onTransferRequest  func(transfer.Request, string)
}

// SetTransferCallbacks sets the callbacks for transfer events
func (e *Engine) SetTransferCallbacks(onProgress func(transfer.Session), onRequest func(transfer.Request, string)) {
	e.onTransferProgress = onProgress
	e.onTransferRequest = onRequest
}

// SetOnChatMessage sets the callback for received chat messages
func (e *Engine) SetOnChatMessage(handler func(chat.Message)) {
	e.onChatMessage = handler
}

// NewEngine creates a new Engine instance
func NewEngine(cfg *config.Config, idMgr *identity.Manager, wgClient WireGuardClient, apiClient *api.Client, logger service.Logger) (*Engine, error) {
	sysConf := system.NewConfigurator()
	hostsMgr := system.NewHostsManager()

	// Initialize P2P Manager
	p2pMgr := p2p.NewManager(apiClient, cfg.P2P.StunServer)

	// Initialize Chat Manager
	chatMgr := chat.NewManager()

	// Initialize Transfer Manager
	transferMgr := transfer.NewManager()

	return &Engine{
		config:        cfg,
		idMgr:         idMgr,
		apiClient:     apiClient,
		wgClient:      wgClient,
		p2pMgr:        p2pMgr,
		chatMgr:       chatMgr,
		transferMgr:   transferMgr,
		sysConf:       sysConf,
		hostsMgr:      hostsMgr,
		stopChan:      make(chan struct{}),
		logf:          logger,
		daemonVersion: Version,
		peerMap:       make(map[string]api.PeerConfig),
	}, nil
}

// Start starts the engine
func (e *Engine) Start() {
	e.logf.Info("Starting daemon engine...")

	// Setup P2P callbacks
	if e.config.P2P.Enabled {
		e.p2pMgr.SetNewConnectionCallback(func(peerID string, conn *ice.Conn) {
			e.logf.Infof("P2P connection established with %s", peerID)
			e.updatePeerEndpoint(peerID, conn.RemoteAddr().String())
		})

		e.p2pMgr.Start()
	} else {
		e.logf.Info("P2P is disabled in config.")
	}

	// Start WebSocket for signaling
	go func() {
		// Attempt to start WebSocket. It might fail if not authenticated yet,
		// but the client should handle retries or we can trigger it later.
		if err := e.apiClient.StartWebSocket(context.Background()); err != nil {
			e.logf.Error("Failed to start WebSocket: ", err.Error())
		}
	}()

	go e.heartbeatLoop()
	go e.configLoop()
}

// Stop stops the engine
func (e *Engine) Stop() {
	e.logf.Info("Stopping daemon engine...")
	close(e.stopChan)
	if e.wgClient != nil {
		if err := e.wgClient.Down(); err != nil {
			e.logf.Error("Failed to bring down WireGuard interface: ", err.Error())
		}
	}
	if e.chatMgr != nil {
		e.chatMgr.Stop()
	}
	e.logf.Info("Daemon engine stopped.")
}

// Connect enables the VPN connection
func (e *Engine) Connect() {
	e.logf.Info("Connecting VPN...")
	e.paused = false
	go e.syncConfig()
}

// Disconnect disables the VPN connection
func (e *Engine) Disconnect() {
	e.logf.Info("Disconnecting VPN...")
	e.paused = true
	if e.wgClient != nil {
		if err := e.wgClient.Down(); err != nil {
			e.logf.Error("Failed to clear WireGuard config: ", err.Error())
		} else {
			e.logf.Info("WireGuard interface cleared successfully.")
		}
	}
	if e.chatMgr != nil {
		e.chatMgr.Stop()
	}
}

func (e *Engine) configLoop() {
	ticker := time.NewTicker(e.config.Daemon.HealthCheckInterval) // Use config value
	defer ticker.Stop()

	if !e.paused {
		e.syncConfig()
	}

	for {
		select {
		case <-ticker.C:
			if !e.paused {
				e.syncConfig()
			}
		case <-e.stopChan:
			e.logf.Info("Config sync loop stopped.")
			return
		}
	}
}

func (e *Engine) syncConfig() {
	id := e.idMgr.Get()
	if id.DeviceID == "" { // Token is now in keyring, not in identity struct
		e.logf.Info("Device not registered, skipping config sync.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := e.apiClient.GetConfig(ctx, id.DeviceID) // No authToken needed here
	if err != nil {
		e.logf.Error("Config sync failed: ", err.Error())
		return
	}

	e.logf.Infof("Received config: %d peers, %d addresses", len(config.Peers), len(config.Interface.Addresses))

	// Fetch networks
	networks, err := e.apiClient.GetNetworks(ctx)
	if err != nil {
		e.logf.Error("Failed to fetch networks: ", err.Error())
	} else {
		e.mu.Lock()
		e.networks = networks
		e.mu.Unlock()
	}

	// Update peer map
	e.mu.Lock()
	for _, peer := range config.Peers {
		if peer.ID != "" {
			e.peerMap[peer.ID] = peer
		}
	}
	e.mu.Unlock()

	// Trigger P2P connections
	if e.config.P2P.Enabled {
		for _, peer := range config.Peers {
			if peer.ID != "" {
				// Deterministic initiator selection to avoid glare
				if id.DeviceID < peer.ID {
					if !e.p2pMgr.IsConnected(peer.ID) {
						e.logf.Infof("Initiating P2P connection to %s", peer.ID)
						go func(peerID string) {
							ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
							defer cancel()
							if _, err := e.p2pMgr.Connect(ctx, peerID); err != nil {
								e.logf.Errorf("Failed to connect to peer %s: %v", peerID, err)
							}
						}(peer.ID)
					}
				}
			}
		}
	}

	if e.wgClient != nil {
		if err := e.wgClient.ApplyConfig(config, id.PrivateKey); err != nil {
			e.logf.Error("Failed to apply WireGuard config: ", err.Error())
		} else {
			e.logf.Info("WireGuard configuration applied successfully")

			if err := e.sysConf.ConfigureInterface(e.config.WireGuard.InterfaceName, config.Interface.Addresses, config.Interface.DNS, config.Interface.MTU); err != nil {
				e.logf.Error("Failed to configure network interface: ", err.Error())
			} else {
				e.logf.Info("Network interface configured successfully")
			}

			routes := make([]string, 0)
			for _, peer := range config.Peers {
				routes = append(routes, peer.AllowedIPs...)
			}
			if len(routes) > 0 {
				if err := e.sysConf.AddRoutes(e.config.WireGuard.InterfaceName, routes); err != nil {
					e.logf.Error("Failed to add routes: ", err.Error())
				} else {
					e.logf.Infof("Added %d routes successfully", len(routes))
				}
			}

			hostEntries := make([]system.HostEntry, 0)
			for _, peer := range config.Peers {
				if len(peer.AllowedIPs) > 0 {
					ip := peer.AllowedIPs[0]
					if idx := strings.Index(ip, "/"); idx != -1 {
						ip = ip[:idx]
					}

					if peer.Name != "" {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Name,
						})
					}
					if peer.Hostname != "" && peer.Hostname != peer.Name {
						hostEntries = append(hostEntries, system.HostEntry{
							IP:       ip,
							Hostname: peer.Hostname,
						})
					}
				}
			}

			if len(hostEntries) > 0 {
				if err := e.hostsMgr.UpdateHosts(hostEntries); err != nil {
					e.logf.Error("Failed to update hosts file: ", err.Error())
				} else {
					e.logf.Infof("Updated hosts file with %d entries", len(hostEntries))
				}
			}

			// Start Chat Listener if we have an IP
			if len(config.Interface.Addresses) > 0 {
				ip := config.Interface.Addresses[0]
				if idx := strings.Index(ip, "/"); idx != -1 {
					ip = ip[:idx]
				}
				// Restart chat manager to bind to new IP
				e.chatMgr.Stop()
				e.chatMgr = chat.NewManager()

				// Setup callback to log messages for now (will be SSE later)
				e.chatMgr.OnMessage(func(msg chat.Message) {
					e.logf.Infof("Received chat from %s: %s", msg.From, msg.Content)
					// Check if it's a transfer signaling message
					if strings.Contains(msg.Content, "\"file_name\"") && strings.Contains(msg.Content, "\"file_size\"") {
						e.transferMgr.HandleSignalingMessage(msg.Content, msg.From)
					}
					if e.onChatMessage != nil {
						e.onChatMessage(msg)
					}
				})

				if err := e.chatMgr.Start(ip, 3000); err != nil {
					e.logf.Error("Failed to start chat listener: ", err.Error())
				}

				// Start Transfer Listener
				e.transferMgr.Stop()
				e.transferMgr = transfer.NewManager()
				e.transferMgr.SetCallbacks(func(s transfer.Session) {
					if e.onTransferProgress != nil {
						e.onTransferProgress(s)
					}
				}, func(req transfer.Request, senderID string) {
					if e.onTransferRequest != nil {
						e.onTransferRequest(req, senderID)
					}
				})

				if err := e.transferMgr.Start(ip); err != nil {
					e.logf.Error("Failed to start transfer listener: ", err.Error())
				}
			}
		}
	}
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running": true,
		"version": e.daemonVersion,
		"paused":  e.paused,
	}

	if e.wgClient != nil {
		wgStatus, err := e.wgClient.GetStatus()
		if err == nil {
			status["wg"] = wgStatus
		} else {
			status["wg"] = map[string]interface{}{"active": false, "error": err.Error()}
		}
	} else {
		status["wg"] = map[string]interface{}{"active": false}
	}

	// Add P2P status
	p2pStatus := make(map[string]interface{})
	e.mu.RLock()
	// We iterate over peerMap to get known peers
	for peerID := range e.peerMap {
		p2pStatus[peerID] = e.p2pMgr.GetPeerStatus(peerID)
	}

	// Add networks and role
	status["networks"] = e.networks
	if len(e.networks) > 0 {
		status["role"] = e.networks[0].Role
		status["network_name"] = e.networks[0].Name
	}
	e.mu.RUnlock()
	status["p2p"] = p2pStatus

	return status
}

func (e *Engine) heartbeatLoop() {
	ticker := time.NewTicker(e.config.Daemon.HealthCheckInterval)
	defer ticker.Stop()

	e.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			e.sendHeartbeat()
		case <-e.stopChan:
			e.logf.Info("Heartbeat loop stopped.")
			return
		}
	}
}

func (e *Engine) sendHeartbeat() {
	id := e.idMgr.Get()
	if id.DeviceID == "" { // Token is now in keyring, not in identity struct
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := api.HeartbeatRequest{
		DaemonVer: e.daemonVersion,
		OSVersion: system.GetOSVersion(),
	}

	if err := e.apiClient.SendHeartbeat(ctx, id.DeviceID, req); err != nil { // No authToken needed here
		e.logf.Error("Heartbeat failed: ", err.Error())
	}
}

func (e *Engine) updatePeerEndpoint(peerID, endpoint string) {
	e.logf.Infof("Updating endpoint for peer %s to %s", peerID, endpoint)

	e.mu.RLock()
	peer, ok := e.peerMap[peerID]
	e.mu.RUnlock()

	if !ok {
		e.logf.Errorf("Unknown peer ID %s, cannot update endpoint", peerID)
		return
	}

	peer.Endpoint = endpoint
	e.mu.Lock()
	e.peerMap[peerID] = peer
	e.mu.Unlock()

}

// ManualConnect initiates a P2P connection to a specific peer manually
func (e *Engine) ManualConnect(peerID string) error {
	if !e.config.P2P.Enabled {
		return fmt.Errorf("P2P is disabled in configuration")
	}
	e.logf.Infof("Manual P2P connection requested for peer %s", peerID)

	e.mu.RLock()
	_, ok := e.peerMap[peerID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown peer ID: %s", peerID)
	}

	if e.p2pMgr.IsConnected(peerID) {
		return fmt.Errorf("already connected to peer %s", peerID)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := e.p2pMgr.Connect(ctx, peerID); err != nil {
			e.logf.Errorf("Failed to manually connect to peer %s: %v", peerID, err)
		} else {
			e.logf.Infof("Successfully manually connected to peer %s", peerID)
		}
	}()

	return nil
}

// SendChatMessage sends a chat message to a peer
func (e *Engine) SendChatMessage(peerID, content string) error {
	e.mu.RLock()
	peer, ok := e.peerMap[peerID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown peer ID: %s", peerID)
	}

	if len(peer.AllowedIPs) == 0 {
		return fmt.Errorf("peer %s has no allowed IPs", peerID)
	}

	// Use the first allowed IP (usually the internal VPN IP)
	ip := peer.AllowedIPs[0]
	if idx := strings.Index(ip, "/"); idx != -1 {
		ip = ip[:idx]
	}

	myID := e.idMgr.Get().DeviceID
	return e.chatMgr.SendMessage(ip, 3000, content, myID)
}

// SendFileRequest initiates a file transfer request
func (e *Engine) SendFileRequest(peerID, filePath string) (*transfer.Session, error) {
	session, err := e.transferMgr.CreateSendSession(peerID, filePath)
	if err != nil {
		return nil, err
	}

	// Send request via chat
	req := transfer.Request{
		ID:       session.ID,
		FileName: session.FileName,
		FileSize: session.FileSize,
	}
	reqBytes, _ := json.Marshal(req)
	if err := e.SendChatMessage(peerID, string(reqBytes)); err != nil {
		return nil, err
	}

	return session, nil
}

// AcceptFile accepts a file transfer
func (e *Engine) AcceptFile(transferID, savePath string) error {
	req, senderID, ok := e.transferMgr.GetPendingRequest(transferID)
	if !ok {
		return fmt.Errorf("transfer request not found or expired")
	}

	session, err := e.transferMgr.CreateReceiveSession(req, senderID, savePath)
	if err != nil {
		return err
	}

	e.mu.RLock()
	peer, ok := e.peerMap[senderID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown peer: %s", senderID)
	}
	if len(peer.AllowedIPs) == 0 {
		return fmt.Errorf("no IP for peer: %s", senderID)
	}
	ip := peer.AllowedIPs[0]
	if idx := strings.Index(ip, "/"); idx != -1 {
		ip = ip[:idx]
	}

	return e.transferMgr.StartDownload(session.ID, ip)
}

// =============================================================================
// NETWORK MANAGEMENT METHODS
// =============================================================================

// CreateNetwork creates a new network via the API
func (e *Engine) CreateNetwork(name string) (*api.NetworkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return e.apiClient.CreateNetwork(ctx, name)
}

// JoinNetwork joins an existing network via invite code
func (e *Engine) JoinNetwork(inviteCode string) (*api.NetworkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return e.apiClient.JoinNetwork(ctx, inviteCode)
}

// GetNetworks returns the list of networks the user is part of
func (e *Engine) GetNetworks() ([]api.NetworkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	networks, err := e.apiClient.GetNetworks(ctx)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.networks = networks
	e.mu.Unlock()

	return networks, nil
}

// GetNetwork returns details of a specific network
func (e *Engine) GetNetwork(networkID string) (*api.NetworkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return e.apiClient.GetNetwork(ctx, networkID)
}

// DeleteNetwork deletes a network (owner only)
func (e *Engine) DeleteNetwork(networkID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.apiClient.DeleteNetwork(ctx, networkID); err != nil {
		return err
	}

	// Remove network from local cache
	e.mu.Lock()
	newNetworks := make([]api.NetworkResponse, 0)
	for _, n := range e.networks {
		if n.ID != networkID {
			newNetworks = append(newNetworks, n)
		}
	}
	e.networks = newNetworks
	e.mu.Unlock()

	// Trigger config resync
	go e.syncConfig()

	return nil
}

// GetPeers returns the list of connected peers
func (e *Engine) GetPeers() []api.PeerConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	peers := make([]api.PeerConfig, 0, len(e.peerMap))
	for _, peer := range e.peerMap {
		peers = append(peers, peer)
	}
	return peers
}

// GetPeerByID returns a specific peer by ID
func (e *Engine) GetPeerByID(peerID string) (*api.PeerConfig, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	peer, ok := e.peerMap[peerID]
	if !ok {
		return nil, false
	}
	return &peer, true
}

// LeaveNetwork removes the current device from a network
func (e *Engine) LeaveNetwork(networkID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.apiClient.LeaveNetwork(ctx, networkID); err != nil {
		return err
	}

	// Remove network from local cache
	e.mu.Lock()
	newNetworks := make([]api.NetworkResponse, 0)
	for _, n := range e.networks {
		if n.ID != networkID {
			newNetworks = append(newNetworks, n)
		}
	}
	e.networks = newNetworks
	e.mu.Unlock()

	// Trigger config resync
	go e.syncConfig()

	return nil
}

// GenerateInvite creates an invite token for a network
func (e *Engine) GenerateInvite(networkID string, maxUses int, expiresHours int) (*api.InviteTokenResponse, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return e.apiClient.GenerateInvite(ctx, networkID, maxUses, expiresHours)
}

// KickPeer removes a peer from a network
func (e *Engine) KickPeer(networkID, peerID, reason string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return e.apiClient.KickPeer(ctx, networkID, peerID, reason)
}

// BanPeer bans a peer from a network
func (e *Engine) BanPeer(networkID, peerID, reason string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return e.apiClient.BanPeer(ctx, networkID, peerID, reason)
}

// UnbanPeer unbans a peer from a network
func (e *Engine) UnbanPeer(networkID, peerID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return e.apiClient.UnbanPeer(ctx, networkID, peerID)
}

// =============================================================================
// CHAT MANAGEMENT METHODS
// =============================================================================

// GetChatMessages returns chat history
func (e *Engine) GetChatMessages(networkID string, limit int, beforeID string) []chat.Message {
	return e.chatMgr.GetMessages(networkID, limit, beforeID)
}

// SubscribeChatMessages returns a channel for real-time chat messages
func (e *Engine) SubscribeChatMessages() chan chat.Message {
	return e.chatMgr.Subscribe()
}

// UnsubscribeChatMessages removes a chat subscription
func (e *Engine) UnsubscribeChatMessages(ch chan chat.Message) {
	e.chatMgr.Unsubscribe(ch)
}

// =============================================================================
// TRANSFER MANAGEMENT METHODS
// =============================================================================

// GetTransfers returns all active transfer sessions
func (e *Engine) GetTransfers() []transfer.Session {
	return e.transferMgr.GetSessions()
}

// SubscribeTransfers returns a channel for real-time transfer updates
func (e *Engine) SubscribeTransfers() chan transfer.Session {
	return e.transferMgr.Subscribe()
}

// UnsubscribeTransfers removes a transfer subscription
func (e *Engine) UnsubscribeTransfers(ch chan transfer.Session) {
	e.transferMgr.Unsubscribe(ch)
}

// RejectTransfer rejects an incoming file transfer
func (e *Engine) RejectTransfer(transferID string) error {
	return e.transferMgr.RejectTransfer(transferID)
}

// CancelTransfer cancels an ongoing file transfer
func (e *Engine) CancelTransfer(transferID string) error {
	return e.transferMgr.CancelTransfer(transferID)
}
