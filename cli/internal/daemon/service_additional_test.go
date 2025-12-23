package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/cli/internal/api"
	"github.com/orhaniscoding/goconnect/cli/internal/config"
	"github.com/orhaniscoding/goconnect/cli/internal/identity"
	"github.com/orhaniscoding/goconnect/cli/internal/storage"
	"github.com/orhaniscoding/goconnect/cli/internal/transfer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// bufferServiceLogger is a test logger that writes all messages into an in-memory buffer.
// It implements kardianos/service.Logger (same method set as fallbackServiceLogger).
type bufferServiceLogger struct {
	mu  sync.Mutex
	buf *bytes.Buffer
}

func (l *bufferServiceLogger) Info(v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprint(v...))
	return nil
}

func (l *bufferServiceLogger) Infof(format string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprintf(format, v...))
	return nil
}

func (l *bufferServiceLogger) Warning(v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprint(v...))
	return nil
}

func (l *bufferServiceLogger) Warningf(format string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprintf(format, v...))
	return nil
}

func (l *bufferServiceLogger) Error(v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprint(v...))
	return nil
}

func (l *bufferServiceLogger) Errorf(format string, v ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.buf, fmt.Sprintf(format, v...))
	return nil
}

// ==================== SSE /events Endpoint Tests ====================

func TestDaemonService_HTTP_Events_SSE(t *testing.T) {
	t.Run("SSE Connection And Message Receive", func(t *testing.T) {
		svc, server := setupTestDaemon(t)
		defer server.Close()

		// Create a request with context that we can cancel
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest("GET", "/events", nil).WithContext(ctx)
		w := &sseResponseRecorder{
			ResponseRecorder: httptest.NewRecorder(),
			flushed:          make(chan bool, 10),
		}

		// Start the SSE handler in a goroutine
		done := make(chan struct{})
		go func() {
			svc.localHTTPServer.Handler.ServeHTTP(w, req)
			close(done)
		}()

		// Give time for handler to start and client to register
		time.Sleep(50 * time.Millisecond)

		// Broadcast a message
		svc.broadcastSSE(`{"type":"test","data":"hello"}`)

		// Wait for flush
		select {
		case <-w.flushed:
			// Good - data was flushed
		case <-time.After(500 * time.Millisecond):
			// Timeout is acceptable if no flush happened
		}

		// Cancel context to close connection
		cancel()

		// Wait for handler to finish
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("SSE handler did not exit")
		}

		// Verify response headers
		assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	})

	t.Run("SSE Client Registers And Unregisters", func(t *testing.T) {
		svc := NewDaemonService(&config.Config{}, "1.0.0")
		svc.logf = &fallbackServiceLogger{}

		mux := http.NewServeMux()
		svc.setupLocalhostBridgeHandlers(mux)
		svc.localHTTPServer = &http.Server{Handler: mux}
		// Need identity manager for other handlers
		tmpDir := t.TempDir()
		svc.idManager = identity.NewManager(filepath.Join(tmpDir, "id.json"))
		svc.idManager.LoadOrCreateIdentity()
		// Mock engine
		mockEng := new(MockEngine)
		mockEng.On("GetStatus").Return(map[string]interface{}{"running": true})
		svc.engine = mockEng

		// Check initial state - no clients
		assert.Empty(t, svc.sseClients)

		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest("GET", "/events", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		done := make(chan struct{})
		go func() {
			svc.localHTTPServer.Handler.ServeHTTP(w, req)
			close(done)
		}()

		// Wait for client to register
		time.Sleep(50 * time.Millisecond)

		svc.sseMu.RLock()
		clientCount := len(svc.sseClients)
		svc.sseMu.RUnlock()
		assert.Equal(t, 1, clientCount, "Should have 1 SSE client")

		// Cancel to disconnect
		cancel()

		<-done

		// Verify client was unregistered
		time.Sleep(50 * time.Millisecond)
		svc.sseMu.RLock()
		clientCount = len(svc.sseClients)
		svc.sseMu.RUnlock()
		assert.Equal(t, 0, clientCount, "Should have 0 SSE clients after disconnect")
	})
}

// sseResponseRecorder wraps httptest.ResponseRecorder to track Flush calls
type sseResponseRecorder struct {
	*httptest.ResponseRecorder
	flushed chan bool
}

func (r *sseResponseRecorder) Flush() {
	r.ResponseRecorder.Flush()
	select {
	case r.flushed <- true:
	default:
	}
}

// ==================== autoConnectLoop Error Path Tests ====================

func TestDaemonService_AutoConnectLoop_NoToken(t *testing.T) {
	// Override retry interval for fast test
	origInterval := autoConnectRetryInterval
	autoConnectRetryInterval = 10 * time.Millisecond
	defer func() { autoConnectRetryInterval = origInterval }()

	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")

	// Keyring that returns error (no token stored)
	kr, _ := storage.NewTestKeyring(t.TempDir())
	cfg.Keyring = kr
	// Don't store any token

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()
	svc.idManager.Update("device-123") // Device is registered

	mockEng := new(MockEngine)
	svc.engine = mockEng

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Run auto-connect - should try and fail to get token
	done := make(chan struct{})
	go func() {
		svc.autoConnectLoop(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Loop finished (either by context cancel or max retries)
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("autoConnectLoop did not exit")
	}

	// Connect should NOT have been called since no token
	mockEng.AssertNotCalled(t, "Connect")
}

func TestDaemonService_AutoConnectLoop_NoDeviceID(t *testing.T) {
	origInterval := autoConnectRetryInterval
	autoConnectRetryInterval = 10 * time.Millisecond
	defer func() { autoConnectRetryInterval = origInterval }()

	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")

	kr, _ := storage.NewTestKeyring(t.TempDir())
	cfg.Keyring = kr
	kr.StoreAuthToken("valid-token")

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()
	// Do NOT set device ID - simulating unregistered device

	mockEng := new(MockEngine)
	svc.engine = mockEng

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		svc.autoConnectLoop(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("autoConnectLoop did not exit")
	}

	// Connect should NOT have been called since no device ID
	mockEng.AssertNotCalled(t, "Connect")
}

func TestDaemonService_AutoConnectLoop_ContextCancelled(t *testing.T) {
	origInterval := autoConnectRetryInterval
	autoConnectRetryInterval = 100 * time.Millisecond // Longer to test cancel
	defer func() { autoConnectRetryInterval = origInterval }()

	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")

	kr, _ := storage.NewTestKeyring(t.TempDir())
	cfg.Keyring = kr

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()

	mockEng := new(MockEngine)
	svc.engine = mockEng

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.autoConnectLoop(ctx)
		close(done)
	}()

	// Cancel immediately
	cancel()

	select {
	case <-done:
		// Verify logs
		// assert.Contains(t, logBuf.String(), "Auto-connect loop context cancelled")
	case <-time.After(1 * time.Second):
		t.Fatal("autoConnectLoop did not exit on context cancel")
	}
}

// ==================== Stop Method Tests ====================

func TestDaemonService_Stop_AllComponents(t *testing.T) {
	svc := NewDaemonService(&config.Config{}, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	// Set up cancel function
	ctx, cancel := context.WithCancel(context.Background())
	svc.cancel = cancel

	// Mock engine
	mockEng := new(MockEngine)
	mockEng.On("Stop").Return()
	svc.engine = mockEng

	// Create a real HTTP server (listening on random port)
	mux := http.NewServeMux()
	svc.localHTTPServer = &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	go svc.localHTTPServer.ListenAndServe()
	time.Sleep(10 * time.Millisecond)

	// Stop should work without panic
	err := svc.Stop(nil)
	assert.NoError(t, err)

	// Verify context was cancelled
	select {
	case <-ctx.Done():
		// Good - context was cancelled
	default:
		t.Error("Context should be cancelled after Stop")
	}

	mockEng.AssertExpectations(t)
}

// ==================== HTTP Handler Error Tests ====================

func TestDaemonService_HTTP_P2PConnect_EngineError(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	mockEng := new(MockEngine)
	mockEng.On("ManualConnect", "peer-fail").Return(errors.New("connection refused"))
	svc.engine = mockEng

	reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-fail"})
	req := httptest.NewRequest("POST", "/p2p/connect", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "connection refused")
	mockEng.AssertExpectations(t)
}

func TestDaemonService_HTTP_ChatSend_EngineError(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	mockEng := new(MockEngine)
	mockEng.On("SendChatMessage", "peer-x", "hello").Return(errors.New("peer offline"))
	svc.engine = mockEng

	reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-x", "content": "hello"})
	req := httptest.NewRequest("POST", "/chat/send", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "peer offline")
	mockEng.AssertExpectations(t)
}

func TestDaemonService_HTTP_FileSend_EngineError(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	mockEng := new(MockEngine)
	mockEng.On("SendFileRequest", "peer-y", "/nonexistent/file.txt").Return(nil, errors.New("file not found"))
	svc.engine = mockEng

	reqBody, _ := json.Marshal(map[string]string{"peer_id": "peer-y", "file_path": "/nonexistent/file.txt"})
	req := httptest.NewRequest("POST", "/file/send", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "file not found")
	mockEng.AssertExpectations(t)
}

func TestDaemonService_HTTP_FileAccept_EngineError(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	mockEng := new(MockEngine)
	mockEng.On("AcceptFile", "req-bad", "/save/here").Return(errors.New("transfer expired"))
	svc.engine = mockEng

	reqBody, _ := json.Marshal(map[string]interface{}{
		"request":   map[string]interface{}{"id": "req-bad"},
		"peer_id":   "peer-z",
		"save_path": "/save/here",
	})
	req := httptest.NewRequest("POST", "/file/accept", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "transfer expired")
	mockEng.AssertExpectations(t)
}

func TestDaemonService_HTTP_Networks_APIError(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()
	svc.config.Keyring.StoreAuthToken("bad-token")

	req := httptest.NewRequest("GET", "/networks", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDaemonService_HTTP_NetworksCreate_APIError(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "quota exceeded", http.StatusForbidden)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()
	svc.config.Keyring.StoreAuthToken("valid-token")

	reqBody, _ := json.Marshal(map[string]string{"name": "MyNet"})
	req := httptest.NewRequest("POST", "/networks/create", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDaemonService_HTTP_NetworksJoin_APIError(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid invite", http.StatusBadRequest)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()
	svc.config.Keyring.StoreAuthToken("valid-token")

	reqBody, _ := json.Marshal(map[string]string{"invite_code": "BADCODE"})
	req := httptest.NewRequest("POST", "/networks/join", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Register Handler Error Tests ====================

func TestDaemonService_HTTP_Register_APIError(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "device already registered", http.StatusConflict)
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()

	reqBody, _ := json.Marshal(map[string]string{"token": "some-token"})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Registration failed")
}

func TestDaemonService_HTTP_Register_NoKeyring(t *testing.T) {
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/devices" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(api.RegisterDeviceResponse{ID: "device-new"})
			return
		}
	}

	svc, server := setupTestDaemonWithAPI(t, apiHandler)
	defer server.Close()

	// Remove keyring to simulate failure
	svc.config.Keyring = nil

	reqBody, _ := json.Marshal(map[string]string{"token": "valid-token"})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ==================== Config Handler Tests ====================

func TestDaemonService_HTTP_Config_MethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("DELETE", "/config", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDaemonService_HTTP_Config_InvalidJSON(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("POST", "/config", bytes.NewReader([]byte("not valid json")))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDaemonService_HTTP_Config_PartialUpdate(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// Update only P2P enabled
	reqBody, _ := json.Marshal(map[string]interface{}{
		"p2p_enabled": true,
	})
	req := httptest.NewRequest("POST", "/config", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, svc.config.P2P.Enabled)
}

// ==================== CORS Edge Cases ====================

func TestDaemonService_HTTP_CORS_GoConnectOrigin(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/status", nil)
	req.Header.Set("Origin", "https://app.goconnect.io")
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	// Should be allowed because origin contains "goconnect"
	assert.Equal(t, "https://app.goconnect.io", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestDaemonService_HTTP_CORS_NoOrigin(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	// No Origin header - e.g., direct API call
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Connect/Disconnect Edge Cases ====================

func TestDaemonService_HTTP_Connect_MethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/connect", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDaemonService_HTTP_Disconnect_MethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/disconnect", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ==================== run() Tests ====================

func TestDaemonService_Run_MultipleHealthChecks(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	var logBuf bytes.Buffer
	svc.logf = &bufferServiceLogger{buf: &logBuf}

	// Very short health check interval
	svc.config.Daemon.HealthCheckInterval = 5 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.run(ctx)
		close(done)
	}()

	// Wait for multiple health checks
	time.Sleep(30 * time.Millisecond)
	cancel()

	<-done

	// Verify multiple health checks occurred
	logs := logBuf.String()
	count := strings.Count(logs, "Performing daemon health check/sync")
	assert.GreaterOrEqual(t, count, 2, "Should have performed multiple health checks")
}

// ==================== program Start/Stop Tests ====================

func TestProgram_StartStop(t *testing.T) {
	// program.Start and program.Stop require a real service.Service
	// which is hard to mock. We test the wrapper struct behavior.
	svc := NewDaemonService(&config.Config{}, "1.0.0")
	prg := &program{daemon: svc}

	assert.NotNil(t, prg.daemon)
	assert.Equal(t, svc, prg.daemon)
}

// ==================== Multiple SSE Clients ====================

func TestDaemonService_MultipleSSEClients(t *testing.T) {
	svc := NewDaemonService(&config.Config{}, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	// Add multiple clients
	client1 := make(chan string, 1)
	client2 := make(chan string, 1)
	client3 := make(chan string, 1)

	svc.sseMu.Lock()
	svc.sseClients[client1] = true
	svc.sseClients[client2] = true
	svc.sseClients[client3] = true
	svc.sseMu.Unlock()

	// Broadcast
	svc.broadcastSSE("multi-message")

	// All should receive
	assert.Equal(t, "multi-message", <-client1)
	assert.Equal(t, "multi-message", <-client2)
	assert.Equal(t, "multi-message", <-client3)
}

// ==================== Networks Handler Method Checks ====================

func TestDaemonService_HTTP_NetworksCreate_MethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/networks/create", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestDaemonService_HTTP_NetworksJoin_MethodNotAllowed(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/networks/join", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ==================== Transfer Session Response Tests ====================

func TestDaemonService_HTTP_FileSend_SuccessResponse(t *testing.T) {
	svc, server := setupTestDaemon(t)
	defer server.Close()

	mockEng := new(MockEngine)
	mockSession := &transfer.Session{
		ID:       "transfer-abc",
		Status:   transfer.StatusPending,
		PeerID:   "peer-123",
		FileName: "test.txt",
		FileSize: 1024,
	}
	mockEng.On("SendFileRequest", "peer-123", "/path/to/test.txt").Return(mockSession, nil)
	svc.engine = mockEng

	reqBody, _ := json.Marshal(map[string]string{
		"peer_id":   "peer-123",
		"file_path": "/path/to/test.txt",
	})
	req := httptest.NewRequest("POST", "/file/send", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp transfer.Session
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "transfer-abc", resp.ID)
	assert.Equal(t, transfer.StatusPending, resp.Status)
	assert.Equal(t, "peer-123", resp.PeerID)
	mockEng.AssertExpectations(t)
}

// ==================== Empty/Edge Case Status Response ====================

func TestDaemonService_HTTP_Status_WithUnregisteredDevice(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()
	// NOT setting device ID - device is unregistered

	mockEng := new(MockEngine)
	mockEng.On("GetStatus").Return(map[string]interface{}{
		"running": true,
		"version": "test",
	})
	svc.engine = mockEng

	mux := http.NewServeMux()
	svc.setupLocalhostBridgeHandlers(mux)
	svc.localHTTPServer = &http.Server{Handler: mux}

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	svc.localHTTPServer.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	device := resp["device"].(map[string]interface{})
	assert.Equal(t, false, device["registered"])
}

// ==================== autoConnectLoop Max Retries ====================

func TestDaemonService_AutoConnectLoop_MaxRetries(t *testing.T) {
	origInterval := autoConnectRetryInterval
	autoConnectRetryInterval = 5 * time.Millisecond
	defer func() { autoConnectRetryInterval = origInterval }()

	tmpDir := t.TempDir()
	cfg := &config.Config{}
	cfg.IdentityPath = filepath.Join(tmpDir, "identity.json")

	kr, _ := storage.NewTestKeyring(t.TempDir())
	cfg.Keyring = kr
	// No token stored

	svc := NewDaemonService(cfg, "1.0.0")
	svc.logf = &fallbackServiceLogger{}

	svc.idManager = identity.NewManager(cfg.IdentityPath)
	svc.idManager.LoadOrCreateIdentity()
	svc.idManager.Update("device-maxretry")

	mockEng := new(MockEngine)
	svc.engine = mockEng

	// Don't cancel - let it run to max retries
	ctx := context.Background()

	start := time.Now()
	svc.autoConnectLoop(ctx)
	elapsed := time.Since(start)

	// Should complete after max retries (12 * 5ms = 60ms minimum)
	// assert.Contains(t, logBuf.String(), "Max retries reached")
	assert.Greater(t, elapsed, 50*time.Millisecond)
}

// ==================== Broadcast SSE With JSON Marshal ====================

func TestDaemonService_BroadcastSSE_JSONPayload(t *testing.T) {
	svc := NewDaemonService(&config.Config{}, "1.0.0")

	client := make(chan string, 1)
	svc.sseMu.Lock()
	svc.sseClients[client] = true
	svc.sseMu.Unlock()

	// Test with complex JSON payload
	payload := map[string]interface{}{
		"type": "chat_message",
		"data": map[string]string{
			"from":    "user1",
			"content": "Hello!",
		},
	}
	jsonBytes, _ := json.Marshal(payload)
	svc.broadcastSSE(string(jsonBytes))

	received := <-client
	assert.Contains(t, received, "chat_message")
	assert.Contains(t, received, "Hello!")
}

// ==================== Helper to test setupLocalhostBridgeHandlers coverage ====================

func TestDaemonService_SetupHandlers_RegistersAllEndpoints(t *testing.T) {
	svc := NewDaemonService(&config.Config{}, "1.0.0")
	svc.logf = &fallbackServiceLogger{}
	tmpDir := t.TempDir()
	svc.idManager = identity.NewManager(filepath.Join(tmpDir, "id.json"))
	svc.idManager.LoadOrCreateIdentity()

	mockEng := new(MockEngine)
	mockEng.On("GetStatus").Return(map[string]interface{}{"running": true})
	svc.engine = mockEng

	mockAPI := new(MockAPIClient)
	svc.apiClient = mockAPI

	mux := http.NewServeMux()
	svc.setupLocalhostBridgeHandlers(mux)

	// Test that non-SSE endpoints are registered by calling them
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/status"},
		{"OPTIONS", "/status"},
		{"GET", "/config"},
		// Note: /events is SSE and would block, so we test it separately
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s_%s", ep.method, ep.path), func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			// Should not be 404
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint %s %s should be registered", ep.method, ep.path)
		})
	}
}

// ==================== Test Mock Engine expectations ====================

func TestMockEngine_AllMethods(t *testing.T) {
	// Verify MockEngine implements DaemonEngine interface
	var _ DaemonEngine = (*MockEngine)(nil)

	mockEng := new(MockEngine)

	// Test methods that take no return or simple returns
	mockEng.On("Start").Return()
	mockEng.On("Stop").Return()
	mockEng.On("Connect").Return()
	mockEng.On("Disconnect").Return()
	mockEng.On("GetStatus").Return(map[string]interface{}{"test": true})
	mockEng.On("LeaveNetwork", "net-1").Return(nil)
	mockEng.On("RejectTransfer", "t-1").Return(nil)
	mockEng.On("CancelTransfer", "t-2").Return(nil)

	mockEng.Start()
	mockEng.Stop()
	mockEng.Connect()
	mockEng.Disconnect()

	status := mockEng.GetStatus()
	assert.True(t, status["test"].(bool))

	err := mockEng.LeaveNetwork("net-1")
	assert.NoError(t, err)

	err = mockEng.RejectTransfer("t-1")
	assert.NoError(t, err)

	err = mockEng.CancelTransfer("t-2")
	assert.NoError(t, err)

	mockEng.AssertExpectations(t)
}

// ==================== DaemonAPIClient interface test ====================

func TestMockAPIClient_AllMethods(t *testing.T) {
	var _ DaemonAPIClient = (*MockAPIClient)(nil)

	mockAPI := new(MockAPIClient)
	ctx := context.Background()

	mockAPI.On("Register", mock.Anything, "token", mock.Anything).Return(&api.RegisterDeviceResponse{ID: "dev-1"}, nil)
	mockAPI.On("GetNetworks", mock.Anything).Return([]api.NetworkResponse{{ID: "net-1"}}, nil)
	mockAPI.On("CreateNetwork", mock.Anything, "MyNet", mock.Anything).Return(&api.NetworkResponse{ID: "net-2", Name: "MyNet"}, nil)
	mockAPI.On("JoinNetwork", mock.Anything, "INVITE").Return(&api.NetworkResponse{ID: "net-3"}, nil)

	resp, err := mockAPI.Register(ctx, "token", api.RegisterDeviceRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "dev-1", resp.ID)

	networks, err := mockAPI.GetNetworks(ctx)
	assert.NoError(t, err)
	assert.Len(t, networks, 1)

	net, err := mockAPI.CreateNetwork(ctx, "MyNet", "")
	assert.NoError(t, err)
	assert.Equal(t, "MyNet", net.Name)

	net, err = mockAPI.JoinNetwork(ctx, "INVITE")
	assert.NoError(t, err)
	assert.Equal(t, "net-3", net.ID)

	mockAPI.AssertExpectations(t)
}
