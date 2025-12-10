package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/engine"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/transfer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== NewDaemonService Tests ====================

func TestNewDaemonService(t *testing.T) {
	t.Run("Creates Service With Config", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewDaemonService(cfg, "1.0.0")

		require.NotNil(t, svc)
		assert.Equal(t, cfg, svc.config)
		assert.Equal(t, "1.0.0", svc.daemonVersion)
		assert.NotNil(t, svc.sseClients)
	})

	t.Run("Creates Service With Nil Config", func(t *testing.T) {
		svc := NewDaemonService(nil, "1.0.0")
		require.NotNil(t, svc)
		assert.Nil(t, svc.config)
	})

	t.Run("Initializes SSE Clients Map", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")
		assert.NotNil(t, svc.sseClients)
		assert.Empty(t, svc.sseClients)
	})
}

func TestNewDaemonServiceWithBuildInfo(t *testing.T) {
	t.Run("Creates Service With Build Info", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewDaemonServiceWithBuildInfo(cfg, "1.0.0", "2024-01-01", "abc123")

		require.NotNil(t, svc)
		assert.Equal(t, cfg, svc.config)
		assert.Equal(t, "1.0.0", svc.daemonVersion)
		assert.Equal(t, "2024-01-01", svc.buildDate)
		assert.Equal(t, "abc123", svc.commit)
		assert.NotNil(t, svc.sseClients)
	})

	t.Run("All Build Info Fields Set", func(t *testing.T) {
		svc := NewDaemonServiceWithBuildInfo(&config.Config{}, "v2.0.0-beta", "2024-12-01", "deadbeef")

		assert.Equal(t, "v2.0.0-beta", svc.daemonVersion)
		assert.Equal(t, "2024-12-01", svc.buildDate)
		assert.Equal(t, "deadbeef", svc.commit)
	})
}

// ==================== broadcastSSE Tests ====================

func TestDaemonService_broadcastSSE(t *testing.T) {
	t.Run("Broadcasts To All Clients", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")

		// Add some clients
		client1 := make(chan string, 1)
		client2 := make(chan string, 1)

		svc.sseMu.Lock()
		svc.sseClients[client1] = true
		svc.sseClients[client2] = true
		svc.sseMu.Unlock()

		// Broadcast message
		svc.broadcastSSE("test message")

		// Verify both clients received
		assert.Equal(t, "test message", <-client1)
		assert.Equal(t, "test message", <-client2)
	})

	t.Run("Non Blocking On Slow Client", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")

		// Unbuffered channel simulates slow client
		slowClient := make(chan string)
		svc.sseMu.Lock()
		svc.sseClients[slowClient] = true
		svc.sseMu.Unlock()

		// Should not block
		svc.broadcastSSE("test message")
		// Test passes if we get here without blocking
	})

	t.Run("Empty Clients Does Not Panic", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")

		// No clients registered
		svc.broadcastSSE("test message")
		// Test passes if no panic
	})
}

// ==================== fallbackServiceLogger Tests ====================

func TestFallbackServiceLogger(t *testing.T) {
	t.Run("Info Logs Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Info("test info")
		assert.NoError(t, err)
	})

	t.Run("Infof Logs Formatted Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Infof("test %s %d", "message", 42)
		assert.NoError(t, err)
	})

	t.Run("Warning Logs Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Warning("test warning")
		assert.NoError(t, err)
	})

	t.Run("Warningf Logs Formatted Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Warningf("test %s", "warning")
		assert.NoError(t, err)
	})

	t.Run("Error Logs Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Error("test error")
		assert.NoError(t, err)
	})

	t.Run("Errorf Logs Formatted Message", func(t *testing.T) {
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		fsl := &fallbackServiceLogger{logger}

		err := fsl.Errorf("test %v", "error")
		assert.NoError(t, err)
	})
}

// ==================== program Tests ====================

func TestProgram(t *testing.T) {
	t.Run("Program Wraps DaemonService", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")
		prg := &program{daemon: svc}

		require.NotNil(t, prg)
		assert.Equal(t, svc, prg.daemon)
	})
}

// ==================== DaemonService Struct Tests ====================

func TestDaemonServiceStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		svc := &DaemonService{
			daemonVersion: "1.0.0",
			buildDate:     "2024-01-01",
			commit:        "abc123",
			sseClients:    make(map[chan string]bool),
		}

		assert.Equal(t, "1.0.0", svc.daemonVersion)
		assert.Equal(t, "2024-01-01", svc.buildDate)
		assert.Equal(t, "abc123", svc.commit)
	})

	t.Run("SSE Fields Are Thread Safe", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")

		done := make(chan bool, 10)

		// Concurrent access
		for i := 0; i < 5; i++ {
			go func() {
				client := make(chan string, 1)
				svc.sseMu.Lock()
				svc.sseClients[client] = true
				svc.sseMu.Unlock()
				done <- true
			}()
		}

		// Concurrent broadcast
		for i := 0; i < 5; i++ {
			go func() {
				svc.broadcastSSE("message")
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// ==================== Engine/API Client Fields Tests ====================

func TestDaemonService_NilSafety(t *testing.T) {
	t.Run("Service Can Be Created With Minimal Config", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "test")
		require.NotNil(t, svc)
		assert.Nil(t, svc.engine)
		assert.Nil(t, svc.apiClient)
		assert.Nil(t, svc.idManager)
	})
}

// ==================== Stop Method Safety Tests ====================

func TestDaemonService_Stop(t *testing.T) {
	t.Run("Stop Does Not Panic With Nil Components", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")
		logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
		svc.logf = &fallbackServiceLogger{logger}

		// All components are nil
		svc.cancel = nil
		svc.grpcServer = nil
		svc.engine = nil
		svc.localHTTPServer = nil

		// Should not panic
		err := svc.Stop(nil)
		assert.NoError(t, err)
	})
}

// ==================== HTTP Handler Tests ====================

func setupTestDaemon(t *testing.T) (*DaemonService, *httptest.Server) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")
	cfg.Daemon.LocalPort = 0

	// Initialize Memory Keyring
	kr, err := storage.NewTestKeyring(t.TempDir())
	require.NoError(t, err)
	cfg.Keyring = kr
	// Store dummy token
	err = kr.StoreAuthToken("valid-token")
	require.NoError(t, err)

	// Create mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/networks":
			json.NewEncoder(w).Encode([]api.NetworkResponse{{ID: "net-1", Name: "Test Net"}})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	cfg.Server.URL = apiServer.URL

	svc := NewDaemonService(cfg, "1.0.0")
	// Initialize logger to prevent panics
	svc.logf = &fallbackServiceLogger{log.New(os.Stderr, "[test] ", log.LstdFlags)}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	// Create identity file
	_, err = svc.idManager.LoadOrCreateIdentity()
	require.NoError(t, err)
	// Mark as registered by setting device ID
	err = svc.idManager.Update("test-device-id")
	require.NoError(t, err)

	svc.apiClient = api.NewClient(cfg)
	// Initialize Engine properly
	// We need a do-nothing wgClient? wireguard.NewClient returns error if interface not found?
	// We can pass nil to wgClient if NewEngine allows it? 
	// NewEngine signature: func NewEngine(..., wgClient *wireguard.Client, ...)
	// It stores it.
	
	// Create a real engine instance
	eng, err := engine.NewEngine(cfg, svc.idManager, nil, svc.apiClient.(*api.Client), svc.logf)
	// If NewEngine fails (e.g. system calls), we might fallback to struct. 
	// But NewEngine mainly just inits structs.
	if err == nil {
		svc.engine = eng
	} else {
		// If NewEngine fails (unlikely for basic init), fallback to unsafe empty struct
		// and log warning
		t.Logf("Warning: Failed to create real engine: %v. Using empty struct.", err)
		svc.engine = &engine.Engine{}
	}

	// Initialize handlers
	mux := http.NewServeMux()
	svc.setupLocalhostBridgeHandlers(mux)
	svc.localHTTPServer = &http.Server{Handler: mux}

	return svc, apiServer
}

// Redefine setupTestDaemonWithHandler to allow custom API behavior
func setupTestDaemonWithAPI(t *testing.T, apiHandler http.HandlerFunc) (*DaemonService, *httptest.Server) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")
	cfg.Daemon.LocalPort = 0
	
	// Keyring
	kr, _ := storage.NewTestKeyring(t.TempDir())
	cfg.Keyring = kr
	
	apiServer := httptest.NewServer(apiHandler)
	cfg.Server.URL = apiServer.URL

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{log.New(os.Stderr, "[test] ", log.LstdFlags)}
	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()
	svc.apiClient = api.NewClient(cfg)
	
	eng, _ := engine.NewEngine(cfg, svc.idManager, nil, svc.apiClient.(*api.Client), svc.logf)
	svc.engine = eng

	mux := http.NewServeMux()
	svc.setupLocalhostBridgeHandlers(mux)
	svc.localHTTPServer = &http.Server{Handler: mux}

	return svc, apiServer
}

func TestDaemonService_HTTP_Register_Success(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/devices" && r.Method == "POST" {
			// Return success
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(api.RegisterDeviceResponse{
				ID: "new-device-id",
			})
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()

	// Call /register
	reqBody, _ := json.Marshal(map[string]string{
		"token": "valid-token",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "connected", resp["status"])
	
	// Verify ID updated (requires inspecting idManager, but we can't easily given encapsulation, 
	// unless we trust the result flow).
}

func TestDaemonService_HTTP_Networks_Create(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks" && r.Method == "POST" {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] == "New Net" {
				json.NewEncoder(w).Encode(api.NetworkResponse{ID: "net-2", Name: "New Net"})
				return
			}
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()
	svc.config.Keyring.StoreAuthToken("valid-token")

	reqBody, _ := json.Marshal(map[string]string{
		"name": "New Net",
	})
	req := httptest.NewRequest("POST", "/networks/create", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var net api.NetworkResponse
	json.Unmarshal(w.Body.Bytes(), &net)
	assert.Equal(t, "net-2", net.ID)
	assert.Equal(t, "New Net", net.Name)
}

func TestDaemonService_HTTP_Networks_Join(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/networks/join" && r.Method == "POST" {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["invite_code"] == "INVITE123" {
				json.NewEncoder(w).Encode(api.NetworkResponse{ID: "net-3", Name: "Joined Net"})
				return
			}
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()
	svc.config.Keyring.StoreAuthToken("valid-token")

	reqBody, _ := json.Marshal(map[string]string{
		"invite_code": "INVITE123",
	})
	req := httptest.NewRequest("POST", "/networks/join", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var net api.NetworkResponse
	json.Unmarshal(w.Body.Bytes(), &net)
	assert.Equal(t, "net-3", net.ID)
	assert.Equal(t, "Joined Net", net.Name)
}

func TestDaemonService_HTTP_Connect_Disconnect(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// Use mock engine to avoid real API calls
	mockEng := new(MockEngine)
	mockEng.On("Connect").Return()
	mockEng.On("Disconnect").Return()
	svc.engine = mockEng

	// Test Connect
	req := httptest.NewRequest("POST", "/connect", nil)
	w := httptest.NewRecorder()
	svc.localHTTPServer.Handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test Disconnect
	req = httptest.NewRequest("POST", "/disconnect", nil)
	w = httptest.NewRecorder()
	svc.localHTTPServer.Handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockEng.AssertExpectations(t)
}

func TestDaemonService_HTTP_CORS(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// 1. Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/status", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))

	// 2. Test valid Origin
	req = httptest.NewRequest("GET", "/status", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	w = httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)
	assert.Equal(t, "http://127.0.0.1:8080", w.Header().Get("Access-Control-Allow-Origin"))

	// 3. Test invalid Origin (should still process but maybe not set CORS headers or log warning)
	// The implementation blocks CORS headers but serves if allowed logic permits?
	// Logic: "if allowed { set headers } else { log warning }" then next(w,r)
	// So it proceeds even if origin is invalid for CORS, but browser would block it.
	req = httptest.NewRequest("GET", "/status", nil)
	req.Header.Set("Origin", "http://evil.com")
	w = httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestDaemonService_HTTP_Status(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, true, resp["running"])
	assert.Equal(t, "dev", resp["version"]) // "dev" from engine package default
	
	// Verify device info
	device, ok := resp["device"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, device["public_key"])
	assert.Equal(t, true, device["registered"])
}

func TestDaemonService_HTTP_Config(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// GET Config
	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var cfg config.Config
	err := json.Unmarshal(w.Body.Bytes(), &cfg)
	require.NoError(t, err)
	// Server URL should match our mock server
	assert.Equal(t, server.URL, cfg.Server.URL)

	// POST Config update
	updateBody := map[string]interface{}{
		"p2p_enabled": true,
		"stun_server": "stun:test.com",
	}
	body, _ := json.Marshal(updateBody)
	req = httptest.NewRequest("POST", "/config", bytes.NewReader(body))
	w = httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify config update in memory
	assert.True(t, svc.config.P2P.Enabled)
	assert.Equal(t, "stun:test.com", svc.config.P2P.StunServer)
}

func TestDaemonService_HTTP_Networks(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/networks", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var networks []api.NetworkResponse
	err := json.Unmarshal(w.Body.Bytes(), &networks)
	require.NoError(t, err)

	// Should return the network from our mock API server
	require.Len(t, networks, 1)
	assert.Equal(t, "net-1", networks[0].ID)
	assert.Equal(t, "Test Net", networks[0].Name)
}

func TestDaemonService_AutoConnectLoop(t *testing.T) {
	// Override retry interval for test
	origInterval := autoConnectRetryInterval
	autoConnectRetryInterval = 10 * time.Millisecond
	defer func() { autoConnectRetryInterval = origInterval }()

	// Mock API that returns config
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/config") {
			json.NewEncoder(w).Encode(api.DeviceConfig{
				Interface: api.InterfaceConfig{Addresses: []string{"10.0.0.1/24"}},
				Peers:     []api.PeerConfig{},
			})
			return
		}
		if strings.Contains(r.URL.Path, "/v1/networks") {
			json.NewEncoder(w).Encode([]api.NetworkResponse{})
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()

	// Ensure we have identity and token
	svc.idManager.Update("device-auto")
	svc.config.Keyring.StoreAuthToken("valid-token")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify auto-connect triggers Connect() which sets paused=false.
    // First, verify we are disconnected (paused=true) or force it.
    svc.engine.Disconnect()
    
    // Check initial state
    status := svc.engine.GetStatus()
    if p, ok := status["paused"].(bool); !ok || !p {
        t.Fatal("Expected engine to be paused after Disconnect")
    }

	// Run loop in background
	go svc.autoConnectLoop(ctx)

	// Wait for connection attempt (paused -> false)
	assert.Eventually(t, func() bool {
		status := svc.engine.GetStatus()
		paused, ok := status["paused"].(bool)
		return ok && !paused
	}, 1*time.Second, 50*time.Millisecond, "Should auto-connect and unpause")
}

func TestDaemonService_RunLoop(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// Capture logs
	var logBuf bytes.Buffer
	svc.logf = &fallbackServiceLogger{log.New(&logBuf, "", 0)}

	// Shorten health check interval
	svc.config.Daemon.HealthCheckInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	
	// Run loop
	done := make(chan struct{})
	go func() {
		svc.run(ctx)
		close(done)
	}()

	// Wait for a few ticks
	time.Sleep(50 * time.Millisecond)

	// Stop loop
	cancel()
	
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Run loop did not exit")
	}

	// Verify logs
	logs := logBuf.String()
	assert.Contains(t, logs, "Daemon main loop started")
	assert.Contains(t, logs, "Performing daemon health check/sync")
	assert.Contains(t, logs, "Daemon run loop context cancelled")
}

// ==================== Additional HTTP Handler Tests ====================

func TestDaemonService_HTTP_P2PConnect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		// Create mock engine
		mockEng := new(MockEngine)
		mockEng.On("ManualConnect", "peer-123").Return(nil)
		svc.engine = mockEng

		reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-123"})
		req := httptest.NewRequest("POST", "/p2p/connect", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "initiated", resp["status"])
		assert.Equal(t, "peer-123", resp["peer_id"])
		mockEng.AssertExpectations(t)
	})

	t.Run("Missing Peer ID", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		reqBody, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest("POST", "/p2p/connect", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("GET", "/p2p/connect", nil)
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDaemonService_HTTP_ChatSend(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		mockEng := new(MockEngine)
		mockEng.On("SendChatMessage", "peer-456", "Hello World").Return(nil)
		svc.engine = mockEng

		reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-456", "content": "Hello World"})
		req := httptest.NewRequest("POST", "/chat/send", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "sent", resp["status"])
		mockEng.AssertExpectations(t)
	})

	t.Run("Missing Fields", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		// Missing content
		reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-456"})
		req := httptest.NewRequest("POST", "/chat/send", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("GET", "/chat/send", nil)
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDaemonService_HTTP_FileSend(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		mockEng := new(MockEngine)
		mockSession := &transfer.Session{ID: "transfer-1", Status: transfer.StatusPending}
		mockEng.On("SendFileRequest", "peer-789", "/path/to/file.txt").Return(mockSession, nil)
		svc.engine = mockEng

		reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-789", "file_path": "/path/to/file.txt"})
		req := httptest.NewRequest("POST", "/file/send", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp transfer.Session
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "transfer-1", resp.ID)
		mockEng.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("POST", "/file/send", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDaemonService_HTTP_FileAccept(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		mockEng := new(MockEngine)
		mockEng.On("AcceptFile", "req-123", "/save/path").Return(nil)
		svc.engine = mockEng

		reqBody, _ := json.Marshal(map[string]interface{}{
			"request":   map[string]interface{}{"id": "req-123"},
			"peer_id":   "peer-xyz",
			"save_path": "/save/path",
		})
		req := httptest.NewRequest("POST", "/file/accept", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "accepted", resp["status"])
		mockEng.AssertExpectations(t)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("GET", "/file/accept", nil)
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestDaemonService_HTTP_Networks_GetMethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("POST", "/networks", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDaemonService_HTTP_NetworksCreate_ValidationErrors(t *testing.T) {
	t.Run("Missing Name", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		reqBody, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest("POST", "/networks/create", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("POST", "/networks/create", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDaemonService_HTTP_NetworksJoin_ValidationErrors(t *testing.T) {
	t.Run("Missing Invite Code", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		reqBody, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest("POST", "/networks/join", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("POST", "/networks/join", bytes.NewReader([]byte("not json")))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDaemonService_HTTP_Register_ValidationErrors(t *testing.T) {
	t.Run("Missing Token", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		reqBody, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("bad json")))
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		req := httptest.NewRequest("GET", "/register", nil)
		w := httptest.NewRecorder()

		svc.localHTTPServer.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
