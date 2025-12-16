package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupIsolatedDaemon() (*DaemonService, *MockEngine, *MockAPIClient, *http.ServeMux) {
	cfg := &config.Config{
		Settings: struct {
			AutoConnect          bool   `yaml:"auto_connect"`
			NotificationsEnabled bool   `yaml:"notifications_enabled"`
			DownloadPath         string `yaml:"download_path"`
			LogLevel             string `yaml:"log_level"`
		}{
			DownloadPath: "/tmp",
		},
	}

	// Initialize test keyring
	keyringDir, _ := os.MkdirTemp("", "goconnect-keyring")
	testKeyring, _ := storage.NewTestKeyring(keyringDir)
	cfg.Keyring = testKeyring

	svc := NewDaemonService(cfg, "test-version")
	svc.logf = &fallbackServiceLogger{}

	// Mock ID Manager?
	// DaemonService.idManager is private.
	// We might need to initialize it or mock the check in handlers.
	// Handlers use s.idManager.Get().

	mockEngine := new(MockEngine)
	mockAPI := new(MockAPIClient)

	svc.engine = mockEngine
	svc.apiClient = mockAPI

	// We need to initialize idManager for handlers to work
	tmpFile, _ := os.CreateTemp("", "identity.json")
	tmpFile.Close()
	svc.idManager = identity.NewManager(tmpFile.Name())
	svc.idManager.LoadOrCreateIdentity()

	mux := http.NewServeMux()
	svc.setupLocalhostBridgeHandlers(mux)

	return svc, mockEngine, mockAPI, mux
}

func TestDaemonService_Isolated_Register(t *testing.T) {
	svc, _, mockAPI, mux := setupIsolatedDaemon()

	mockResp := &api.RegisterDeviceResponse{
		ID: "dev123",
	}
	// Note: In strict mocking, context might need exact matching, or use mock.Anything
	mockAPI.On("Register", mock.Anything, "valid-token", mock.Anything).Return(mockResp, nil)

	// Mock engine connect since register calls it
	// svc.engine is accessed in register handler
	svc.engine.(*MockEngine).On("Connect").Return()

	body := map[string]string{"token": "valid-token"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAPI.AssertExpectations(t)
	svc.engine.(*MockEngine).AssertExpectations(t)
}

func TestDaemonService_Isolated_CreateNetwork(t *testing.T) {
	_, _, mockAPI, mux := setupIsolatedDaemon()

	mockNet := &api.NetworkResponse{
		ID:   "net1",
		Name: "test-net",
	}
	mockAPI.On("CreateNetwork", mock.Anything, "test-net", mock.Anything).Return(mockNet, nil)

	body := map[string]string{"name": "test-net"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/networks/create", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp api.NetworkResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "net1", resp.ID)
	mockAPI.AssertExpectations(t)
}

func TestDaemonService_Isolated_ManualConnect(t *testing.T) {
	_, mockEngine, _, mux := setupIsolatedDaemon()

	mockEngine.On("ManualConnect", "peer-123").Return(nil)

	body := map[string]string{"peer_id": "peer-123"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/p2p/connect", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEngine.AssertExpectations(t)
}

func TestDaemonService_Isolated_SendChat(t *testing.T) {
	_, mockEngine, _, mux := setupIsolatedDaemon()

	mockEngine.On("SendChatMessage", "peer-1", "hello").Return(nil)

	body := map[string]string{"peer_id": "peer-1", "content": "hello"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/chat/send", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEngine.AssertExpectations(t)
}

func TestDaemonService_Isolated_Status(t *testing.T) {
	_, mockEngine, _, mux := setupIsolatedDaemon()

	mockStatus := map[string]interface{}{
		"connected": true,
		"uptime":    100,
	}
	mockEngine.On("GetStatus").Return(mockStatus)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, true, resp["connected"])
	mockEngine.AssertExpectations(t)
}

func TestDaemonService_Isolated_Networks_List(t *testing.T) {
	_, _, mockAPI, mux := setupIsolatedDaemon()

	mockNets := []api.NetworkResponse{
		{ID: "n1", Name: "Network 1"},
	}
	mockAPI.On("GetNetworks", mock.Anything).Return(mockNets, nil)

	req := httptest.NewRequest("GET", "/networks", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []api.NetworkResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "n1", resp[0].ID)
}

func TestDaemonService_Isolated_Networks_Join(t *testing.T) {
	_, _, mockAPI, mux := setupIsolatedDaemon()

	mockNet := &api.NetworkResponse{ID: "n2", Name: "Joined Net"}
	mockAPI.On("JoinNetwork", mock.Anything, "invite-code").Return(mockNet, nil)

	body := map[string]string{"invite_code": "invite-code"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/networks/join", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.NetworkResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "n2", resp.ID)
}
