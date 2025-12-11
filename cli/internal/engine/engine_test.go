package engine

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/chat"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/wireguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Mock WireGuard Client ====================

// MockWireGuardClient implements WireGuardClient interface for testing
type MockWireGuardClient struct {
	DownFunc        func() error
	ApplyConfigFunc func(config *api.DeviceConfig, privateKey string) error
	GetStatusFunc   func() (*wireguard.Status, error)
}

func (m *MockWireGuardClient) Down() error {
	if m.DownFunc != nil {
		return m.DownFunc()
	}
	return nil
}

func (m *MockWireGuardClient) ApplyConfig(config *api.DeviceConfig, privateKey string) error {
	if m.ApplyConfigFunc != nil {
		return m.ApplyConfigFunc(config, privateKey)
	}
	return nil
}

func (m *MockWireGuardClient) GetStatus() (*wireguard.Status, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc()
	}
	return &wireguard.Status{Active: true}, nil
}

// NewMockWireGuardClient creates a default mock that succeeds
func NewMockWireGuardClient() *MockWireGuardClient {
	return &MockWireGuardClient{}
}

// ==================== Helper Functions ====================

// setupTestEngine creates an Engine with a mock API server and basic dependencies.
// It returns the engine, the mock server (which must be closed), and the temp dir.
func setupTestEngine(t *testing.T, apiHandler http.HandlerFunc) (*Engine, *httptest.Server, string) {
	tmpDir := t.TempDir()

	// Setup Config
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")
	cfg.Daemon.HealthCheckInterval = 100 * time.Millisecond
	// Setup Keyring
	kr, err := storage.NewTestKeyring(tmpDir)
	require.NoError(t, err)
	cfg.Keyring = kr
	// Store dummy token
	err = kr.StoreAuthToken("valid-token")
	require.NoError(t, err)

	// Setup Mock API
	server := httptest.NewServer(apiHandler)
	cfg.Server.URL = server.URL
	cfg.P2P.Enabled = true // Enable P2P to test manager init

	// Setup Identity
	idMgr := identity.NewManager(cfg.IdentityPath)
	_, err = idMgr.LoadOrCreateIdentity()
	require.NoError(t, err)

	// Setup Logger
	logger := &fallbackLogger{log.New(os.Stderr, "[test_engine] ", log.LstdFlags)}

	// Setup API Client
	apiClient := api.NewClient(cfg)

	// Create Engine (pass nil for wgClient to avoid system deps)
	eng, err := NewEngine(cfg, idMgr, nil, apiClient, logger)
	require.NoError(t, err)

	return eng, server, tmpDir
}

// setupTestEngineWithWg creates an Engine with a mock WireGuard client
func setupTestEngineWithWg(t *testing.T, apiHandler http.HandlerFunc, wgClient WireGuardClient) (*Engine, *httptest.Server, string) {
	tmpDir := t.TempDir()

	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")
	cfg.Daemon.HealthCheckInterval = 100 * time.Millisecond
	cfg.WireGuard.InterfaceName = "wg0"

	kr, err := storage.NewTestKeyring(tmpDir)
	require.NoError(t, err)
	cfg.Keyring = kr
	err = kr.StoreAuthToken("valid-token")
	require.NoError(t, err)

	server := httptest.NewServer(apiHandler)
	cfg.Server.URL = server.URL
	cfg.P2P.Enabled = true

	idMgr := identity.NewManager(cfg.IdentityPath)
	_, err = idMgr.LoadOrCreateIdentity()
	require.NoError(t, err)

	logger := &fallbackLogger{log.New(os.Stderr, "[test_engine] ", log.LstdFlags)}
	apiClient := api.NewClient(cfg)

	eng, err := NewEngine(cfg, idMgr, wgClient, apiClient, logger)
	require.NoError(t, err)

	return eng, server, tmpDir
}

// fallbackLogger implements service.Logger for tests
type fallbackLogger struct {
	*log.Logger
}

func (l *fallbackLogger) Info(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Infof(format string, v ...interface{}) error {
	l.Printf(format, v...)
	return nil
}
func (l *fallbackLogger) Warning(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Warningf(format string, v ...interface{}) error {
	l.Printf(format, v...)
	return nil
}
func (l *fallbackLogger) Error(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Errorf(format string, v ...interface{}) error {
	l.Printf(format, v...)
	return nil
}

// ==================== Initialization Tests ====================

func TestNewEngine(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	assert.NotNil(t, eng)
	assert.NotNil(t, eng.p2pMgr)
	assert.NotNil(t, eng.chatMgr)
	assert.NotNil(t, eng.transferMgr)
	assert.NotNil(t, eng.sysConf)
	assert.NotNil(t, eng.hostsMgr)
	assert.NotNil(t, eng.peerMap)
}

// ==================== Network Management Tests ====================

func TestEngine_CreateNetwork(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks" && r.Method == "POST" {
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["name"] == "My Network" {
				_ = json.NewEncoder(w).Encode(api.NetworkResponse{
					ID:   "net-123",
					Name: "My Network",
					Role: "owner",
				})
				return
			}
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	net, err := eng.CreateNetwork("My Network")
	require.NoError(t, err)
	assert.Equal(t, "net-123", net.ID)
	assert.Equal(t, "My Network", net.Name)
}

func TestEngine_JoinNetwork(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks/join" && r.Method == "POST" {
			var req map[string]string
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["invite_code"] == "CODE123" {
				_ = json.NewEncoder(w).Encode(api.NetworkResponse{
					ID:   "net-456",
					Name: "Joined Network",
					Role: "member",
				})
				return
			}
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	net, err := eng.JoinNetwork("CODE123")
	require.NoError(t, err)
	assert.Equal(t, "net-456", net.ID)
	assert.Equal(t, "Joined Network", net.Name)
}

func TestEngine_LeaveNetwork(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks/net-123/leave" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// for syncConfig triggered after leave
		if r.URL.Path == "/v1/networks" {
			_ = json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
		// for config sync
		if r.Method == "GET" && len(r.URL.Path) > 0 {
			_ = json.NewEncoder(w).Encode(api.DeviceConfig{})
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Pre-populate a network
	eng.networks = []api.NetworkResponse{{ID: "net-123", Name: "To Leave"}}

	err := eng.LeaveNetwork("net-123")
	require.NoError(t, err)

	// Verify local cache updated
	eng.mu.Lock()
	defer eng.mu.Unlock()
	assert.Len(t, eng.networks, 0, "Network should be removed from local cache")
}

func TestEngine_GetNetworks(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks" && r.Method == "GET" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{
				{ID: "net-1", Name: "Network 1"},
				{ID: "net-2", Name: "Network 2"},
			})
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	nets, err := eng.GetNetworks()
	require.NoError(t, err)
	assert.Len(t, nets, 2)
	assert.Equal(t, "Network 1", nets[0].Name)

	// Verify cache updated
	eng.mu.Lock()
	defer eng.mu.Unlock()
	assert.Len(t, eng.networks, 2)
}

// ==================== Peer Management Tests ====================

func TestEngine_GetPeers_And_GetPeerByID(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Manually inject peers into map (as syncConfig does)
	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Name: "Alice"}
	eng.peerMap["peer-2"] = api.PeerConfig{ID: "peer-2", Name: "Bob"}

	// Test GetPeers
	peers := eng.GetPeers()
	assert.Len(t, peers, 2)

	// Test GetPeerByID
	p, found := eng.GetPeerByID("peer-1")
	assert.True(t, found)
	assert.Equal(t, "Alice", p.Name)

	_, found = eng.GetPeerByID("peer-99")
	assert.False(t, found)
}

func TestEngine_KickPeer(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/networks/net-1/members/peer-1"
		if r.URL.Path == expectedPath && r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.KickPeer("net-1", "peer-1", "violation")
	require.NoError(t, err)
}

func TestEngine_BanPeer(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/networks/net-1/members/peer-1/ban"
		if r.URL.Path == expectedPath && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.BanPeer("net-1", "peer-1", "spam")
	require.NoError(t, err)
}

func TestEngine_UnbanPeer(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/networks/net-1/members/peer-1/ban"
		if r.URL.Path == expectedPath && r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.UnbanPeer("net-1", "peer-1")
	require.NoError(t, err)
}

// ==================== Invite Management Tests ====================

func TestEngine_GenerateInvite(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/networks/net-1/invites"
		if r.URL.Path == expectedPath && r.Method == "POST" {
			var req api.CreateInviteRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.UsesMax == 5 {
				_ = json.NewEncoder(w).Encode(api.InviteTokenResponse{
					Token:   "INVITE-5",
					UsesMax: 5,
				})
				return
			}
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	invite, err := eng.GenerateInvite("net-1", 5, 24)
	require.NoError(t, err)
	assert.Equal(t, "INVITE-5", invite.Token)
}

// ==================== Config Sync Tests ====================

func TestEngine_SyncConfig(t *testing.T) {
	configCalled := false
	networksCalled := false

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		// Mock Config Response
		if strings.Contains(r.URL.Path, "/config") {
			configCalled = true
			_ = json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.2/24"},
				},
				Peers: []api.PeerConfig{
					{ID: "peer-A", Name: "Peer A", AllowedIPs: []string{"10.0.0.3/32"}},
				},
			})
			return
		}
		// Mock Networks Response
		if r.URL.Path == "/v1/networks" {
			networksCalled = true
			_ = json.NewEncoder(w).Encode([]api.NetworkResponse{
				{ID: "net-1", Name: "Synced Net"},
			})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Mocking DeviceID to allow sync to proceed
	eng.idMgr.Update("device-test-id")

	// Manually trigger syncConfig (private method, but exposed via configLoop or just call it if we use reflection or if we are in same package)
	// Since we are in 'package engine', we can access private methods!
	eng.syncConfig()

	assert.True(t, configCalled, "Should have requested config")
	assert.True(t, networksCalled, "Should have requested networks")

	// Verify State Updates
	eng.mu.Lock()
	defer eng.mu.Unlock()

	assert.Len(t, eng.networks, 1)
	assert.Equal(t, "Synced Net", eng.networks[0].Name)

	assert.Len(t, eng.peerMap, 1)
	assert.Equal(t, "Peer A", eng.peerMap["peer-A"].Name)
}

// ==================== Existing basic tests preserved but refactored ====================

func TestEngine_ChatMethods(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Initially empty
	messages := eng.GetChatMessages("", 10, "")
	assert.Empty(t, messages)

	// Subscribe
	ch := eng.SubscribeChatMessages()
	assert.NotNil(t, ch)
	eng.UnsubscribeChatMessages(ch)
}

func TestEngine_TransferMethods(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Initially empty
	sessions := eng.GetTransfers()
	assert.Empty(t, sessions)

	// Subscribe
	ch := eng.SubscribeTransfers()
	assert.NotNil(t, ch)
	eng.UnsubscribeTransfers(ch)

	// Errors
	err := eng.RejectTransfer("non-existent")
	assert.Error(t, err)

	err = eng.CancelTransfer("non-existent")
	assert.Error(t, err)
}

func TestEngine_SetCallbacks(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	progCalled := false
	reqCalled := false
	chatCalled := false

	eng.SetTransferCallbacks(func(s transfer.Session) { progCalled = true }, func(r transfer.Request, p string) { reqCalled = true })
	eng.SetOnChatMessage(func(m chat.Message) { chatCalled = true })

	if eng.onTransferProgress != nil {
		eng.onTransferProgress(transfer.Session{})
	}
	if eng.onTransferRequest != nil {
		eng.onTransferRequest(transfer.Request{}, "p")
	}
	if eng.onChatMessage != nil {
		eng.onChatMessage(chat.Message{})
	}

	assert.True(t, progCalled)
	assert.True(t, reqCalled)
	assert.True(t, chatCalled)
}

func TestEngine_GetStatus_Complete(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {}
	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Inject state
	eng.daemonVersion = "1.2.3"
	eng.networks = []api.NetworkResponse{{ID: "n1", Name: "N1", Role: "admin"}}

	status := eng.GetStatus()
	assert.Equal(t, "1.2.3", status["version"])
	assert.Equal(t, true, status["running"])
	assert.Equal(t, "N1", status["network_name"])
	assert.Equal(t, "admin", status["role"])

	// WG inactive
	wg := status["wg"].(map[string]interface{})
	assert.Equal(t, false, wg["active"])
}

// ==================== Heartbeat Tests ====================

func TestEngine_Heartbeat(t *testing.T) {
	heartbeatCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/heartbeat") && r.Method == "POST" {
			heartbeatCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Inject identity
	eng.idMgr.Update("device-1")
	// Ensure token for API client (setupTestEngine does this)

	eng.sendHeartbeat()
	assert.True(t, heartbeatCalled)
}

// ==================== Logic Validation Tests ====================

func TestEngine_SendChatMessage_Validation(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Case 1: Unknown Peer
	err := eng.SendChatMessage("unknown-peer", "hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peer ID")

	// Case 2: Peer without AllowedIPs
	eng.peerMap["peer-no-ip"] = api.PeerConfig{ID: "peer-no-ip", Name: "NoIP"}
	err = eng.SendChatMessage("peer-no-ip", "hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no allowed IPs")
}

func TestEngine_ManualConnect_Validation(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Case 1: P2P Disabled
	eng.config.P2P.Enabled = false
	err := eng.ManualConnect("peer-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "P2P is disabled")

	// Case 2: Unknown Peer
	eng.config.P2P.Enabled = true
	err = eng.ManualConnect("unknown-peer")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peer ID")

	// Case 3: Already connected (mocking P2P manager state is hard without interface,
	// checking if we can rely on p2pMgr internal state or just skip this for now)
}

// ==================== Complex Flow Tests ====================

func TestEngine_FileTransfer_Flow(t *testing.T) {
	// Setup dummy peer listener to accept chat connection
	// Use port 3000 as that's what Engine hardcodes for now
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		t.Skipf("Cannot bind to port 3000 for test: %v", err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close() // Accept and close immediately is fine for "sent" check
		}
	}()

	eng, server, tmpDir := setupTestEngine(t, nil)
	defer server.Close()

	// 1. Setup Peer
	peerID := "peer-transfer"
	eng.peerMap[peerID] = api.PeerConfig{
		ID:         peerID,
		Name:       "Transfer Peer",
		AllowedIPs: []string{"127.0.0.1/32"},
	}

	// 2. Create Temp File to Send
	filePath := filepath.Join(tmpDir, "testfile.txt")
	err = os.WriteFile(filePath, []byte("hello world"), 0644)
	require.NoError(t, err)

	// 3. SendFileRequest
	session, err := eng.SendFileRequest(peerID, filePath)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "testfile.txt", session.FileName)
	assert.NotEmpty(t, session.ID)

	// 4. Verify Session in Manager
	sessions := eng.GetTransfers()
	assert.Len(t, sessions, 1)
	assert.Equal(t, session.ID, sessions[0].ID)

	// ==================================================
	// Part 2: AcceptFile
	// Simulate receiving a request from the peer
	incomingReq := map[string]interface{}{
		"id":        "incoming-id",
		"file_name": "received.txt",
		"file_size": 123,
	}
	reqBytes, _ := json.Marshal(incomingReq)

	// Manually inject into transfer manager (simulating chat message receipt)
	eng.transferMgr.HandleSignalingMessage(string(reqBytes), peerID)

	// Accept it
	savePath := filepath.Join(tmpDir, "received.txt")
	// Note: AcceptFile triggers StartDownload which tries to connect to the peer's HTTP server
	// The peer (listener above) is just a raw TCP listener, not a transfer server.
	// So StartDownload might fail or hang.
	// transfer.Manager.StartDownload usually connects to `http://ip:8080/download/...`
	// failing is fine, we just want to verify AcceptFile logic doesn't error out on "request not found".
	// AcceptFile returns error if connection fails.

	// StartDownload might be async or successful in initiation.
	// Since we returned nil, let's assume valid initiation is enough for this test.
	err = eng.AcceptFile("incoming-id", savePath)
	assert.NoError(t, err)
}

func TestEngine_SendChatMessage_Success(t *testing.T) {
	// Setup dummy peer listener
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		t.Skipf("Cannot bind to port 3000: %v", err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			// Read/Write dummy
			buf := make([]byte, 1024)
			conn.Read(buf)
			conn.Close()
		}
	}()

	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.peerMap["peer-chat"] = api.PeerConfig{
		ID:         "peer-chat",
		Name:       "BiDi",
		AllowedIPs: []string{"127.0.0.1/32"},
	}

	err = eng.SendChatMessage("peer-chat", "hello world")
	assert.NoError(t, err)
}

func TestEngine_UpdatePeerEndpoint(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Endpoint: "old:123"}

	eng.updatePeerEndpoint("peer-1", "new:456")

	p, _ := eng.GetPeerByID("peer-1")
	assert.Equal(t, "new:456", p.Endpoint)

	// Unknown peer - should log error but not panic
	eng.updatePeerEndpoint("unknown", "addr")
}

// ==================== Engine Lifecycle Tests ====================

func TestEngine_StartStop(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		// Default handler - just respond OK
		w.WriteHeader(http.StatusOK)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Start should not panic
	assert.NotPanics(t, func() {
		eng.Start()
	})

	// Engine should have started successfully
	assert.NotNil(t, eng.stopChan)

	// Stop should not panic
	assert.NotPanics(t, func() {
		eng.Stop()
	})
}

func TestEngine_ConnectDisconnect(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers:     []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// NewEngine initializes with paused = true by default
	// Connect should set paused = false
	eng.Connect()
	assert.False(t, eng.paused)

	// Disconnect should set paused = true
	eng.Disconnect()
	assert.True(t, eng.paused)
}

func TestEngine_GetStatus_WhenPaused(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Ensure engine is paused
	eng.paused = true

	status := eng.GetStatus()

	assert.True(t, status["paused"].(bool))
	assert.True(t, status["running"].(bool))
}

func TestEngine_GetStatus_WithPeers(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Add some peers to peerMap
	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Name: "Test Peer"}
	eng.peerMap["peer-2"] = api.PeerConfig{ID: "peer-2", Name: "Another Peer"}

	status := eng.GetStatus()

	// Verify p2p status contains the peers
	p2pStatus, ok := status["p2p"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, p2pStatus, 2)
}

// ==================== SyncConfig Detailed Tests ====================

func TestEngine_SyncConfig_NoDeviceID(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		// Should NOT be called since device is not registered
		t.Error("API should not be called when device is not registered")
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Clear device ID to simulate unregistered device
	eng.idMgr.Update("")

	// Should return early without calling API
	eng.syncConfig()
}

func TestEngine_SyncConfig_ConfigFetchError(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-test-id")

	// Should log error but not panic
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
}

func TestEngine_SyncConfig_NetworkFetchError(t *testing.T) {
	configCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			configCalled = true
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers:     []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			http.Error(w, "network error", http.StatusInternalServerError)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-test-id")

	// Should continue without panicking even when network fetch fails
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
	assert.True(t, configCalled)
}

func TestEngine_SyncConfig_WithP2PConnections(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "peer-a", Name: "Peer A", AllowedIPs: []string{"10.0.0.2/32"}},
					{ID: "peer-z", Name: "Peer Z", AllowedIPs: []string{"10.0.0.3/32"}},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Set device ID lower than "peer-z" but higher than "peer-a"
	// to test deterministic initiator selection
	eng.idMgr.Update("peer-m")
	eng.config.P2P.Enabled = true

	// Call syncConfig
	eng.syncConfig()

	// Verify peers were added to peerMap
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.peerMap, 2)
	assert.Equal(t, "Peer A", eng.peerMap["peer-a"].Name)
	assert.Equal(t, "Peer Z", eng.peerMap["peer-z"].Name)
}

func TestEngine_SyncConfig_P2PDisabled(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "peer-a", Name: "Peer A"},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-test-id")
	eng.config.P2P.Enabled = false

	// Should not panic when P2P is disabled
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
}

func TestEngine_SyncConfig_EmptyPeerID(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "", Name: "Empty ID Peer"}, // Should be skipped
					{ID: "peer-1", Name: "Valid Peer"},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-test-id")
	eng.syncConfig()

	// Only valid peer should be in the map
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.peerMap, 1)
	_, exists := eng.peerMap[""]
	assert.False(t, exists, "Empty ID peer should not be in map")
}

// ==================== ManualConnect Tests ====================

func TestEngine_ManualConnect_Success(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.config.P2P.Enabled = true
	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Name: "Test Peer"}

	// ManualConnect starts a goroutine, so we just verify no immediate error
	err := eng.ManualConnect("peer-1")
	assert.NoError(t, err)
}

func TestEngine_ManualConnect_AlreadyConnected(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.config.P2P.Enabled = true
	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Name: "Test Peer"}

	// First connect is fine
	err := eng.ManualConnect("peer-1")
	assert.NoError(t, err)

	// Give the goroutine time to potentially mark as connected
	// The p2pMgr mock state depends on implementation
}

// ==================== Disconnect Tests ====================

func TestEngine_Disconnect_WithoutWgClient(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// wgClient is nil
	eng.paused = false

	// Should not panic
	assert.NotPanics(t, func() {
		eng.Disconnect()
	})
	assert.True(t, eng.paused)
}

func TestEngine_Disconnect_WithoutChatManager(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.chatMgr = nil
	eng.paused = false

	// Should not panic even with nil chatMgr
	assert.NotPanics(t, func() {
		eng.Disconnect()
	})
	assert.True(t, eng.paused)
}

// ==================== Start/Stop Tests ====================

func TestEngine_Start_WithP2PDisabled(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.config.P2P.Enabled = false

	assert.NotPanics(t, func() {
		eng.Start()
	})

	// Cleanup
	eng.Stop()
}

func TestEngine_Stop_WithNilWgClient(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.wgClient = nil

	// Should not panic
	assert.NotPanics(t, func() {
		eng.Stop()
	})
}

func TestEngine_Stop_WithNilChatManager(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.chatMgr = nil

	// Should not panic
	assert.NotPanics(t, func() {
		eng.Stop()
	})
}

// ==================== AcceptFile Tests ====================

func TestEngine_AcceptFile_NotFound(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	err := eng.AcceptFile("non-existent-id", "/tmp/save.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEngine_AcceptFile_UnknownPeer(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Add a pending request
	req := transfer.Request{
		ID:       "req-1",
		FileName: "test.txt",
		FileSize: 100,
	}
	reqBytes, _ := json.Marshal(req)
	eng.transferMgr.HandleSignalingMessage(string(reqBytes), "unknown-peer")

	err := eng.AcceptFile("req-1", "/tmp/save.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peer")
}

func TestEngine_AcceptFile_NoAllowedIPs(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Add peer without AllowedIPs
	eng.peerMap["sender-peer"] = api.PeerConfig{
		ID:         "sender-peer",
		Name:       "Sender",
		AllowedIPs: []string{},
	}

	// Add a pending request
	req := transfer.Request{
		ID:       "req-2",
		FileName: "test.txt",
		FileSize: 100,
	}
	reqBytes, _ := json.Marshal(req)
	eng.transferMgr.HandleSignalingMessage(string(reqBytes), "sender-peer")

	err := eng.AcceptFile("req-2", "/tmp/save.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no IP for peer")
}

// ==================== SendFileRequest Tests ====================

func TestEngine_SendFileRequest_FileNotFound(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.peerMap["peer-1"] = api.PeerConfig{
		ID:         "peer-1",
		Name:       "Test",
		AllowedIPs: []string{"10.0.0.1/32"},
	}

	_, err := eng.SendFileRequest("peer-1", "/non/existent/file.txt")
	assert.Error(t, err)
}

func TestEngine_SendFileRequest_UnknownPeer(t *testing.T) {
	eng, server, tmpDir := setupTestEngine(t, nil)
	defer server.Close()

	// Create a test file
	filePath := filepath.Join(tmpDir, "send.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Try to send to unknown peer
	_, err = eng.SendFileRequest("unknown-peer", filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown peer ID")
}

// ==================== GetStatus Tests ====================

func TestEngine_GetStatus_NoNetworks(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.networks = []api.NetworkResponse{}

	status := eng.GetStatus()
	assert.NotContains(t, status, "role")
	assert.NotContains(t, status, "network_name")
}

func TestEngine_GetStatus_MultipleNetworks(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.networks = []api.NetworkResponse{
		{ID: "n1", Name: "Primary", Role: "owner"},
		{ID: "n2", Name: "Secondary", Role: "member"},
	}

	status := eng.GetStatus()
	// First network is used
	assert.Equal(t, "owner", status["role"])
	assert.Equal(t, "Primary", status["network_name"])
}

// ==================== Heartbeat Tests ====================

func TestEngine_SendHeartbeat_NoDeviceID(t *testing.T) {
	heartbeatCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/heartbeat") {
			heartbeatCalled = true
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Clear device ID
	eng.idMgr.Update("")

	eng.sendHeartbeat()
	assert.False(t, heartbeatCalled, "Heartbeat should not be sent when device is not registered")
}

func TestEngine_SendHeartbeat_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/heartbeat") {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-1")

	// Should not panic on error
	assert.NotPanics(t, func() {
		eng.sendHeartbeat()
	})
}

// ==================== GetNetworks Tests ====================

func TestEngine_GetNetworks_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks" {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	networks, err := eng.GetNetworks()
	assert.Error(t, err)
	assert.Nil(t, networks)
}

// ==================== LeaveNetwork Tests ====================

func TestEngine_LeaveNetwork_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/leave") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.LeaveNetwork("net-123")
	assert.Error(t, err)
}

// ==================== ConfigLoop Tests ====================

func TestEngine_ConfigLoop_WhenPaused(t *testing.T) {
	syncCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			syncCalled = true
		}
		w.WriteHeader(http.StatusOK)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.paused = true
	eng.config.Daemon.HealthCheckInterval = 50 * time.Millisecond

	// Start config loop in goroutine
	go eng.configLoop()

	// Wait a bit for tick
	time.Sleep(150 * time.Millisecond)

	// Stop the engine
	close(eng.stopChan)

	// When paused, syncConfig should not be called from the ticker
	// (Initial call is skipped because paused is true at start)
	assert.False(t, syncCalled)
}

func TestEngine_ConfigLoop_StopSignal(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.config.Daemon.HealthCheckInterval = 100 * time.Millisecond

	// Start config loop
	done := make(chan bool)
	go func() {
		eng.configLoop()
		done <- true
	}()

	// Signal stop immediately
	close(eng.stopChan)

	// Wait for configLoop to exit
	select {
	case <-done:
		// Success - loop exited
	case <-time.After(1 * time.Second):
		t.Fatal("configLoop did not exit after stop signal")
	}
}

// ==================== HeartbeatLoop Tests ====================

func TestEngine_HeartbeatLoop_StopSignal(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.config.Daemon.HealthCheckInterval = 100 * time.Millisecond

	// Start heartbeat loop
	done := make(chan bool)
	go func() {
		eng.heartbeatLoop()
		done <- true
	}()

	// Signal stop
	close(eng.stopChan)

	// Wait for heartbeatLoop to exit
	select {
	case <-done:
		// Success - loop exited
	case <-time.After(1 * time.Second):
		t.Fatal("heartbeatLoop did not exit after stop signal")
	}
}

// ==================== Callback Tests ====================

func TestEngine_OnChatMessage_WithTransferSignaling(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	chatReceived := false
	eng.SetOnChatMessage(func(msg chat.Message) {
		chatReceived = true
	})

	// Simulate receiving a message with transfer signaling content
	transferMsg := `{"id":"tf-1","file_name":"test.txt","file_size":100}`

	// Manually trigger the callback that would happen in syncConfig
	if eng.onChatMessage != nil {
		eng.onChatMessage(chat.Message{
			From:    "peer-1",
			Content: transferMsg,
		})
	}

	assert.True(t, chatReceived)
}

func TestEngine_TransferCallbacks_NilSafe(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Don't set any callbacks - they should be nil
	eng.onTransferProgress = nil
	eng.onTransferRequest = nil
	eng.onChatMessage = nil

	// These operations should not panic even with nil callbacks
	assert.NotPanics(t, func() {
		if eng.onTransferProgress != nil {
			eng.onTransferProgress(transfer.Session{})
		}
		if eng.onTransferRequest != nil {
			eng.onTransferRequest(transfer.Request{}, "peer")
		}
		if eng.onChatMessage != nil {
			eng.onChatMessage(chat.Message{})
		}
	})
}

// ==================== Additional Coverage Tests ====================

func TestEngine_Start_WithP2PEnabled(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.config.P2P.Enabled = true

	// Should setup P2P callbacks and start
	assert.NotPanics(t, func() {
		eng.Start()
	})

	// Give it a moment to start goroutines
	time.Sleep(50 * time.Millisecond)

	// Clean up
	eng.Stop()
}

func TestEngine_ConfigLoop_InitialSyncWhenNotPaused(t *testing.T) {
	syncCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			syncCalled = true
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers:     []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-id")
	eng.paused = false
	eng.config.Daemon.HealthCheckInterval = 50 * time.Millisecond

	// Recreate stopChan since it might be closed
	eng.stopChan = make(chan struct{})

	// Start config loop in goroutine
	go eng.configLoop()

	// Wait for initial sync
	time.Sleep(100 * time.Millisecond)

	// Stop the engine
	close(eng.stopChan)

	// When not paused, initial syncConfig should be called
	assert.True(t, syncCalled)
}

func TestEngine_SyncConfig_WithPeersButLowerDeviceID(t *testing.T) {
	// Test the case where device ID is lower than peer ID (will not initiate)
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "aaa", Name: "Peer AAA", AllowedIPs: []string{"10.0.0.2/32"}},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Set device ID higher than peer ID - should NOT initiate P2P
	eng.idMgr.Update("zzz")
	eng.config.P2P.Enabled = true

	eng.syncConfig()

	// Peer should be in map
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	_, exists := eng.peerMap["aaa"]
	assert.True(t, exists)
}

func TestEngine_SyncConfig_WithPeersButHigherDeviceID(t *testing.T) {
	// Test the case where device ID is higher than peer ID (will initiate)
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "zzz", Name: "Peer ZZZ", AllowedIPs: []string{"10.0.0.2/32"}},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Set device ID lower than peer ID - should initiate P2P
	eng.idMgr.Update("aaa")
	eng.config.P2P.Enabled = true

	eng.syncConfig()

	// Peer should be in map
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	_, exists := eng.peerMap["zzz"]
	assert.True(t, exists)
}

func TestEngine_SyncConfig_P2PInitiationWithMultiplePeers(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers: []api.PeerConfig{
					{ID: "aaa", Name: "Lower", AllowedIPs: []string{"10.0.0.2/32"}},
					{ID: "zzz", Name: "Higher", AllowedIPs: []string{"10.0.0.3/32"}},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Set device ID in the middle
	eng.idMgr.Update("mmm")
	eng.config.P2P.Enabled = true

	eng.syncConfig()

	// Both peers should be in map
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.peerMap, 2)
}

func TestEngine_Version(t *testing.T) {
	// Test that Version variable exists and engine uses it
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	// Default is "dev"
	status := eng.GetStatus()
	assert.NotEmpty(t, status["version"])
}

func TestEngine_GetPeers_Empty(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	peers := eng.GetPeers()
	assert.Empty(t, peers)
	assert.NotNil(t, peers) // Should be empty slice, not nil
}

func TestEngine_GetPeerByID_Multiple(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.peerMap["peer-1"] = api.PeerConfig{ID: "peer-1", Name: "One"}
	eng.peerMap["peer-2"] = api.PeerConfig{ID: "peer-2", Name: "Two"}
	eng.peerMap["peer-3"] = api.PeerConfig{ID: "peer-3", Name: "Three"}

	// Get each peer
	p1, ok := eng.GetPeerByID("peer-1")
	assert.True(t, ok)
	assert.Equal(t, "One", p1.Name)

	p2, ok := eng.GetPeerByID("peer-2")
	assert.True(t, ok)
	assert.Equal(t, "Two", p2.Name)

	p3, ok := eng.GetPeerByID("peer-3")
	assert.True(t, ok)
	assert.Equal(t, "Three", p3.Name)
}

func TestEngine_AcceptFile_WithValidRequest(t *testing.T) {
	eng, server, tmpDir := setupTestEngine(t, nil)
	defer server.Close()

	// Add peer with valid IP
	eng.peerMap["sender"] = api.PeerConfig{
		ID:         "sender",
		Name:       "Sender",
		AllowedIPs: []string{"10.0.0.5/32"},
	}

	// Add a pending request via signaling
	req := transfer.Request{
		ID:       "valid-req",
		FileName: "file.txt",
		FileSize: 200,
	}
	reqBytes, _ := json.Marshal(req)
	eng.transferMgr.HandleSignalingMessage(string(reqBytes), "sender")

	savePath := filepath.Join(tmpDir, "saved.txt")

	// AcceptFile should succeed in creating the session
	// but StartDownload will fail since no actual server is running
	err := eng.AcceptFile("valid-req", savePath)
	// Error is expected because there's no transfer server
	// But the function should reach the StartDownload call
	// We're testing that the validation passes
	if err != nil {
		// Should fail at download, not at validation
		assert.NotContains(t, err.Error(), "not found")
		assert.NotContains(t, err.Error(), "unknown peer")
		assert.NotContains(t, err.Error(), "no IP")
	}
}

func TestEngine_SendFileRequest_PeerWithNoIP(t *testing.T) {
	eng, server, tmpDir := setupTestEngine(t, nil)
	defer server.Close()

	// Create a test file
	filePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	require.NoError(t, err)

	// Add peer without IPs
	eng.peerMap["no-ip-peer"] = api.PeerConfig{
		ID:         "no-ip-peer",
		Name:       "No IP",
		AllowedIPs: []string{},
	}

	_, err = eng.SendFileRequest("no-ip-peer", filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no allowed IPs")
}

func TestEngine_CreateNetwork_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks" && r.Method == "POST" {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	net, err := eng.CreateNetwork("Duplicate Name")
	assert.Error(t, err)
	assert.Nil(t, net)
}

func TestEngine_JoinNetwork_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks/join" {
			http.Error(w, "invalid code", http.StatusBadRequest)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	net, err := eng.JoinNetwork("INVALID")
	assert.Error(t, err)
	assert.Nil(t, net)
}

func TestEngine_GenerateInvite_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/invites") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	invite, err := eng.GenerateInvite("net-1", 5, 24)
	assert.Error(t, err)
	assert.Nil(t, invite)
}

func TestEngine_KickPeer_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/members/") && r.Method == "DELETE" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.KickPeer("net-1", "peer-1", "reason")
	assert.Error(t, err)
}

func TestEngine_BanPeer_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/ban") && r.Method == "POST" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.BanPeer("net-1", "peer-1", "spam")
	assert.Error(t, err)
}

func TestEngine_UnbanPeer_Error(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/ban") && r.Method == "DELETE" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	err := eng.UnbanPeer("net-1", "peer-1")
	assert.Error(t, err)
}

func TestEngine_LeaveNetwork_Success(t *testing.T) {
	leaveCalled := false
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/leave") && r.Method == "POST" {
			leaveCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		// Handle subsequent syncConfig calls
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	// Add network to leave
	eng.networks = []api.NetworkResponse{
		{ID: "net-to-leave", Name: "Leaving"},
		{ID: "net-keep", Name: "Keep"},
	}

	err := eng.LeaveNetwork("net-to-leave")
	assert.NoError(t, err)
	assert.True(t, leaveCalled)

	// Verify network was removed from cache
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.networks, 1)
	assert.Equal(t, "net-keep", eng.networks[0].ID)
}

func TestEngine_GetStatus_AllFields(t *testing.T) {
	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.daemonVersion = "v1.0.0"
	eng.paused = false
	eng.networks = []api.NetworkResponse{
		{ID: "n1", Name: "Network One", Role: "owner"},
	}
	eng.peerMap["p1"] = api.PeerConfig{ID: "p1", Name: "Peer One"}

	status := eng.GetStatus()

	// Check all expected fields
	assert.Equal(t, "v1.0.0", status["version"])
	assert.Equal(t, true, status["running"])
	assert.Equal(t, false, status["paused"])
	assert.Equal(t, "owner", status["role"])
	assert.Equal(t, "Network One", status["network_name"])
	assert.NotNil(t, status["wg"])
	assert.NotNil(t, status["p2p"])
	assert.NotNil(t, status["networks"])
}

func TestEngine_HeartbeatLoop_MultipleIterations(t *testing.T) {
	heartbeatCount := 0
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/heartbeat") {
			heartbeatCount++
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	eng, server, _ := setupTestEngine(t, apiHandler)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.config.Daemon.HealthCheckInterval = 30 * time.Millisecond

	// Recreate stopChan
	eng.stopChan = make(chan struct{})

	// Start heartbeat loop
	go eng.heartbeatLoop()

	// Wait for a few iterations
	time.Sleep(100 * time.Millisecond)

	// Stop the engine
	close(eng.stopChan)

	// Should have at least 2 heartbeats (initial + at least one tick)
	assert.GreaterOrEqual(t, heartbeatCount, 2)
}

// ==================== WireGuard Client Tests ====================

func TestEngine_Stop_WithWgClient(t *testing.T) {
	downCalled := false
	mockWg := &MockWireGuardClient{
		DownFunc: func() error {
			downCalled = true
			return nil
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	eng.Stop()
	assert.True(t, downCalled, "wgClient.Down should be called on Stop")
}

func TestEngine_Stop_WithWgClientError(t *testing.T) {
	mockWg := &MockWireGuardClient{
		DownFunc: func() error {
			return errors.New("wg down error")
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	// Should not panic even with error
	assert.NotPanics(t, func() {
		eng.Stop()
	})
}

func TestEngine_Disconnect_WithWgClient(t *testing.T) {
	downCalled := false
	mockWg := &MockWireGuardClient{
		DownFunc: func() error {
			downCalled = true
			return nil
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	eng.paused = false
	eng.Disconnect()

	assert.True(t, downCalled, "wgClient.Down should be called on Disconnect")
	assert.True(t, eng.paused)
}

func TestEngine_Disconnect_WithWgClientError(t *testing.T) {
	mockWg := &MockWireGuardClient{
		DownFunc: func() error {
			return errors.New("wg down error")
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	eng.paused = false
	// Should not panic and should set paused
	assert.NotPanics(t, func() {
		eng.Disconnect()
	})
	assert.True(t, eng.paused)
}

func TestEngine_GetStatus_WithWgClient(t *testing.T) {
	mockWg := &MockWireGuardClient{
		GetStatusFunc: func() (*wireguard.Status, error) {
			return &wireguard.Status{
				Active:     true,
				ListenPort: 51820,
				Peers:      2,
			}, nil
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	status := eng.GetStatus()
	wgStatus := status["wg"].(*wireguard.Status)
	assert.True(t, wgStatus.Active)
	assert.Equal(t, 51820, wgStatus.ListenPort)
}

func TestEngine_GetStatus_WithWgClientError(t *testing.T) {
	mockWg := &MockWireGuardClient{
		GetStatusFunc: func() (*wireguard.Status, error) {
			return nil, errors.New("wg status error")
		},
	}

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	status := eng.GetStatus()
	wgStatus := status["wg"].(map[string]interface{})
	assert.False(t, wgStatus["active"].(bool))
	assert.Equal(t, "wg status error", wgStatus["error"])
}

func TestEngine_SyncConfig_WithWgClient_Success(t *testing.T) {
	applyConfigCalled := false
	mockWg := &MockWireGuardClient{
		ApplyConfigFunc: func(config *api.DeviceConfig, privateKey string) error {
			applyConfigCalled = true
			return nil
		},
	}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.1/24"},
					DNS:       []string{"8.8.8.8"},
					MTU:       1420,
				},
				Peers: []api.PeerConfig{
					{ID: "peer-1", Name: "Peer 1", AllowedIPs: []string{"10.0.0.2/32"}},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.syncConfig()

	assert.True(t, applyConfigCalled, "wgClient.ApplyConfig should be called")
}

func TestEngine_SyncConfig_WithWgClient_ApplyConfigError(t *testing.T) {
	mockWg := &MockWireGuardClient{
		ApplyConfigFunc: func(config *api.DeviceConfig, privateKey string) error {
			return errors.New("apply config error")
		},
	}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers:     []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")

	// Should not panic on error
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
}

func TestEngine_SyncConfig_WithWgClient_FullConfig(t *testing.T) {
	applyConfigCalled := false
	mockWg := &MockWireGuardClient{
		ApplyConfigFunc: func(config *api.DeviceConfig, privateKey string) error {
			applyConfigCalled = true
			return nil
		},
	}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.1/24"},
					DNS:       []string{"8.8.8.8"},
					MTU:       1420,
				},
				Peers: []api.PeerConfig{
					{
						ID:         "peer-1",
						Name:       "Peer One",
						Hostname:   "peer1.local",
						AllowedIPs: []string{"10.0.0.2/32", "10.0.1.0/24"},
					},
					{
						ID:         "peer-2",
						Name:       "Peer Two",
						AllowedIPs: []string{"10.0.0.3/32"},
					},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.config.P2P.Enabled = false // Disable P2P to simplify test

	eng.syncConfig()

	assert.True(t, applyConfigCalled)

	// Verify peers added
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.peerMap, 2)
}

func TestEngine_SyncConfig_WithPeerHostnames(t *testing.T) {
	mockWg := &MockWireGuardClient{}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.1/24"},
				},
				Peers: []api.PeerConfig{
					{
						ID:         "peer-1",
						Name:       "myserver",
						Hostname:   "myserver.local",
						AllowedIPs: []string{"10.0.0.2/32"},
					},
					{
						ID:         "peer-2",
						Name:       "samename",
						Hostname:   "samename", // Same as Name, should only add once
						AllowedIPs: []string{"10.0.0.3/32"},
					},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.syncConfig()

	// Verify peers are added
	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Len(t, eng.peerMap, 2)
}

func TestEngine_SyncConfig_WithRoutes(t *testing.T) {
	mockWg := &MockWireGuardClient{}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.1/24"},
				},
				Peers: []api.PeerConfig{
					{
						ID:         "peer-1",
						AllowedIPs: []string{"10.0.0.2/32", "192.168.1.0/24"},
					},
				},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")

	// Should not panic
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
}

func TestEngine_SyncConfig_NoPeers(t *testing.T) {
	mockWg := &MockWireGuardClient{}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{"10.0.0.1/24"},
				},
				Peers: []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.syncConfig()

	eng.mu.RLock()
	defer eng.mu.RUnlock()
	assert.Empty(t, eng.peerMap)
}

func TestEngine_SyncConfig_NoAddresses(t *testing.T) {
	mockWg := &MockWireGuardClient{}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{
					Addresses: []string{}, // No addresses
				},
				Peers: []api.PeerConfig{},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")

	// Should not start chat listener with no addresses
	assert.NotPanics(t, func() {
		eng.syncConfig()
	})
}

func TestEngine_NewEngineWithWgClient(t *testing.T) {
	mockWg := NewMockWireGuardClient()

	eng, server, _ := setupTestEngineWithWg(t, nil, mockWg)
	defer server.Close()

	assert.NotNil(t, eng)
	assert.Equal(t, mockWg, eng.wgClient)
}

func TestEngine_Connect_TriggersSync(t *testing.T) {
	syncCalled := false
	mockWg := &MockWireGuardClient{
		ApplyConfigFunc: func(config *api.DeviceConfig, privateKey string) error {
			syncCalled = true
			return nil
		},
	}

	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
			})
			return
		}
		if r.URL.Path == "/v1/networks" {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
	}

	eng, server, _ := setupTestEngineWithWg(t, apiHandler, mockWg)
	defer server.Close()

	eng.idMgr.Update("device-1")
	eng.paused = true

	eng.Connect()

	// Wait for async syncConfig
	time.Sleep(150 * time.Millisecond)

	assert.False(t, eng.paused)
	assert.True(t, syncCalled)
}
