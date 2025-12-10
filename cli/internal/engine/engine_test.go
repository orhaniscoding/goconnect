package engine

import (
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// fallbackLogger implements service.Logger for tests
type fallbackLogger struct {
	*log.Logger
}
func (l *fallbackLogger) Info(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Infof(format string, v ...interface{}) error { l.Printf(format, v...); return nil }
func (l *fallbackLogger) Warning(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Warningf(format string, v ...interface{}) error { l.Printf(format, v...); return nil }
func (l *fallbackLogger) Error(v ...interface{}) error { l.Println(v...); return nil }
func (l *fallbackLogger) Errorf(format string, v ...interface{}) error { l.Printf(format, v...); return nil }

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
			json.NewDecoder(r.Body).Decode(&req)
			if req["name"] == "My Network" {
				json.NewEncoder(w).Encode(api.NetworkResponse{
					ID: "net-123",
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
			json.NewDecoder(r.Body).Decode(&req)
			if req["invite_code"] == "CODE123" {
				json.NewEncoder(w).Encode(api.NetworkResponse{
					ID: "net-456",
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
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
		// for config sync
		if  r.Method == "GET" && len(r.URL.Path) > 0 {
			json.NewEncoder(w).Encode(api.DeviceConfig{})
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
			json.NewDecoder(r.Body).Decode(&req)
			
			if req.UsesMax == 5 {
				json.NewEncoder(w).Encode(api.InviteTokenResponse{
					Token: "INVITE-5",
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
			json.NewEncoder(w).Encode(api.DeviceConfig{
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
			json.NewEncoder(w).Encode([]api.NetworkResponse{
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

	if eng.onTransferProgress != nil { eng.onTransferProgress(transfer.Session{}) }
	if eng.onTransferRequest != nil { eng.onTransferRequest(transfer.Request{}, "p") }
	if eng.onChatMessage != nil { eng.onChatMessage(chat.Message{}) }

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
		ID: peerID, 
		Name: "Transfer Peer", 
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
		"id": "incoming-id",
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
			if err != nil { return }
			// Read/Write dummy
			buf := make([]byte, 1024)
			conn.Read(buf)
			conn.Close()
		}
	}()

	eng, server, _ := setupTestEngine(t, nil)
	defer server.Close()

	eng.peerMap["peer-chat"] = api.PeerConfig{
		ID: "peer-chat", 
		Name: "BiDi", 
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
				Peers: []api.PeerConfig{},
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


