package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebSocket upgrader for test servers
var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{
			URL: "http://localhost:8080",
		},
	}
	
	client := NewClient(cfg)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
	if client.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
	if client.config.Server.URL != "http://localhost:8080" {
		t.Errorf("Expected server URL 'http://localhost:8080', got %s", client.config.Server.URL)
	}
}

func TestGenerateInvite_EmptyNetworkID(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.GenerateInvite(context.Background(), "", 10, 24)
	if err == nil {
		t.Error("Expected error for empty networkID")
	}
	if err.Error() != "networkID is required" {
		t.Errorf("Expected 'networkID is required' error, got: %s", err.Error())
	}
}

// TestRegister_Success tests the Register function which takes an explicit auth token
func TestRegister_Success(t *testing.T) {
	expectedResponse := RegisterDeviceResponse{
		ID: "device-123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/devices" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected auth header 'Bearer test-token', got: %s", r.Header.Get("Authorization"))
		}

		var req RegisterDeviceRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "test-device" {
			t.Errorf("Expected name 'test-device', got %s", req.Name)
		}
		if req.Platform != "windows" {
			t.Errorf("Expected platform 'windows', got %s", req.Platform)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: server.URL},
	}
	client := NewClient(cfg)

	// Register takes an explicit auth token, so it doesn't need keyring
	resp, err := client.Register(context.Background(), "test-token", RegisterDeviceRequest{
		Name:     "test-device",
		Platform: "windows",
		PubKey:   "pubkey123",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.ID != "device-123" {
		t.Errorf("Expected ID 'device-123', got %s", resp.ID)
	}
}

func TestRegister_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
	}))
	defer server.Close()

	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: server.URL},
	}
	client := NewClient(cfg)

	_, err := client.Register(context.Background(), "test-token", RegisterDeviceRequest{
		Name: "test",
	})
	if err == nil {
		t.Error("Expected error for server error")
	}
}

func TestRegister_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: server.URL},
	}
	client := NewClient(cfg)

	_, err := client.Register(context.Background(), "test-token", RegisterDeviceRequest{
		Name: "test",
	})
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

// Test request/response types serialization
func TestCreateInviteRequest_Serialization(t *testing.T) {
	req := CreateInviteRequest{
		ExpiresIn: 86400,
		UsesMax:   5,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded CreateInviteRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ExpiresIn != 86400 {
		t.Errorf("Expected ExpiresIn 86400, got %d", decoded.ExpiresIn)
	}
	if decoded.UsesMax != 5 {
		t.Errorf("Expected UsesMax 5, got %d", decoded.UsesMax)
	}
}

func TestNetworkResponse_Serialization(t *testing.T) {
	resp := NetworkResponse{
		ID:         "net-123",
		Name:       "Test Network",
		InviteCode: "abc123",
		Role:       "owner",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded NetworkResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != "net-123" {
		t.Errorf("Expected ID 'net-123', got %s", decoded.ID)
	}
	if decoded.Role != "owner" {
		t.Errorf("Expected Role 'owner', got %s", decoded.Role)
	}
}

func TestDeviceConfig_Serialization(t *testing.T) {
	cfg := DeviceConfig{
		Interface: InterfaceConfig{
			ListenPort: 51820,
			Addresses:  []string{"10.0.0.1/24"},
			DNS:        []string{"1.1.1.1"},
			MTU:        1420,
		},
		Peers: []PeerConfig{
			{
				ID:         "peer-1",
				PublicKey:  "abc123",
				AllowedIPs: []string{"10.0.0.2/32"},
				Name:       "Device 1",
			},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded DeviceConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Interface.ListenPort != 51820 {
		t.Errorf("Expected ListenPort 51820, got %d", decoded.Interface.ListenPort)
	}
	if len(decoded.Peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(decoded.Peers))
	}
	if decoded.Peers[0].PublicKey != "abc123" {
		t.Errorf("Expected PublicKey 'abc123', got %s", decoded.Peers[0].PublicKey)
	}
}

func TestHeartbeatRequest_Serialization(t *testing.T) {
	req := HeartbeatRequest{
		IPAddress: "192.168.1.1",
		DaemonVer: "1.0.0",
		OSVersion: "10.0",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded HeartbeatRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IPAddress '192.168.1.1', got %s", decoded.IPAddress)
	}
}

func TestRegisterDeviceRequest_Serialization(t *testing.T) {
	req := RegisterDeviceRequest{
		Name:      "my-device",
		Platform:  "linux",
		PubKey:    "pubkey",
		HostName:  "hostname",
		OSVersion: "ubuntu-22.04",
		DaemonVer: "1.0.0",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded RegisterDeviceRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != "my-device" {
		t.Errorf("Expected Name 'my-device', got %s", decoded.Name)
	}
	if decoded.Platform != "linux" {
		t.Errorf("Expected Platform 'linux', got %s", decoded.Platform)
	}
}

// Test WebSocket callback setters
func TestSignalCallbacks(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	// Test OnOffer
	offerCalled := false
	client.OnOffer(func(sourceID, ufrag, pwd string) {
		offerCalled = true
	})
	if client.signalCallbacks.onOffer == nil {
		t.Error("Expected onOffer callback to be set")
	}

	// Test OnAnswer
	answerCalled := false
	client.OnAnswer(func(sourceID, ufrag, pwd string) {
		answerCalled = true
	})
	if client.signalCallbacks.onAnswer == nil {
		t.Error("Expected onAnswer callback to be set")
	}

	// Test OnCandidate
	candidateCalled := false
	client.OnCandidate(func(sourceID, candidate string) {
		candidateCalled = true
	})
	if client.signalCallbacks.onCandidate == nil {
		t.Error("Expected onCandidate callback to be set")
	}

	// Trigger callbacks to verify they work
	client.signalCallbacks.onOffer("src", "ufrag", "pwd")
	client.signalCallbacks.onAnswer("src", "ufrag", "pwd")
	client.signalCallbacks.onCandidate("src", "candidate")

	if !offerCalled {
		t.Error("OnOffer callback was not called")
	}
	if !answerCalled {
		t.Error("OnAnswer callback was not called")
	}
	if !candidateCalled {
		t.Error("OnCandidate callback was not called")
	}
}

// Test CloseWebSocket when not connected
func TestCloseWebSocket_NotConnected(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	// Should not panic when closing without connection
	client.CloseWebSocket()
}

// ==================== Network Management Tests ====================

// Test that CreateNetwork fails when keyring is not initialized
func TestCreateNetwork_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.CreateNetwork(context.Background(), "Test Network")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that JoinNetwork fails when keyring is not initialized
func TestJoinNetwork_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.JoinNetwork(context.Background(), "ABC123")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that GetNetworks fails when keyring is not initialized  
func TestGetNetworks_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.GetNetworks(context.Background())
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that LeaveNetwork fails when keyring is not initialized
func TestLeaveNetwork_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.LeaveNetwork(context.Background(), "net-123")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// ==================== Heartbeat Tests ====================

// Test that SendHeartbeat fails when keyring is not initialized
func TestSendHeartbeat_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	req := HeartbeatRequest{
		IPAddress: "192.168.1.100",
		DaemonVer: "1.0.0",
		OSVersion: "Linux 5.15",
	}
	err := client.SendHeartbeat(context.Background(), "device-123", req)
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// ==================== GetConfig Tests ====================

// Test that GetConfig fails when keyring is not initialized
func TestGetConfig_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.GetConfig(context.Background(), "device-123")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// ==================== Peer Management Tests ====================

// Test that KickPeer fails when keyring is not initialized
func TestKickPeer_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.KickPeer(context.Background(), "net-123", "peer-456", "violation")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that BanPeer fails when keyring is not initialized
func TestBanPeer_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.BanPeer(context.Background(), "net-123", "peer-456", "spam")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that UnbanPeer fails when keyring is not initialized
func TestUnbanPeer_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.UnbanPeer(context.Background(), "net-123", "peer-456")
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// ==================== Signal Methods Tests ====================

func TestSendOffer_NoConnection(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.SendOffer("target-123", "ufrag", "pwd")
	if err == nil {
		t.Error("Expected error when no WebSocket connection")
	}
}

func TestSendAnswer_NoConnection(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.SendAnswer("target-123", "ufrag", "pwd")
	if err == nil {
		t.Error("Expected error when no WebSocket connection")
	}
}

func TestSendCandidate_NoConnection(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.SendCandidate("target-123", "candidate-string")
	if err == nil {
		t.Error("Expected error when no WebSocket connection")
	}
}

// ==================== GenerateInvite Tests ====================

// Test that GenerateInvite fails when keyring is not initialized
func TestGenerateInvite_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	_, err := client.GenerateInvite(context.Background(), "net-123", 10, 24)
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// Test that WebSocket StartWebSocket fails when keyring is not initialized
func TestStartWebSocket_NoKeyring(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	err := client.StartWebSocket(context.Background())
	if err == nil {
		t.Error("Expected error when keyring is not initialized")
	}
	if !strings.Contains(err.Error(), "keyring not initialized") {
		t.Errorf("Expected keyring error, got: %v", err)
	}
}

// ==================== Struct Tests ====================

func TestDeviceConfigStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		cfg := DeviceConfig{
			Interface: InterfaceConfig{
				ListenPort: 51820,
				Addresses:  []string{"10.0.0.1/24"},
				DNS:        []string{"1.1.1.1", "8.8.8.8"},
				MTU:        1420,
			},
			Peers: []PeerConfig{
				{ID: "peer-1", PublicKey: "key1"},
				{ID: "peer-2", PublicKey: "key2"},
			},
		}

		if cfg.Interface.ListenPort != 51820 {
			t.Errorf("Expected ListenPort 51820, got %d", cfg.Interface.ListenPort)
		}
		if len(cfg.Peers) != 2 {
			t.Errorf("Expected 2 peers, got %d", len(cfg.Peers))
		}
	})
}

func TestInterfaceConfigStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		cfg := InterfaceConfig{
			ListenPort: 51820,
			Addresses:  []string{"10.0.0.1/24", "fd00::1/64"},
			DNS:        []string{"1.1.1.1"},
			MTU:        1400,
		}

		if cfg.ListenPort != 51820 {
			t.Errorf("Expected ListenPort 51820, got %d", cfg.ListenPort)
		}
		if len(cfg.Addresses) != 2 {
			t.Errorf("Expected 2 addresses, got %d", len(cfg.Addresses))
		}
		if cfg.MTU != 1400 {
			t.Errorf("Expected MTU 1400, got %d", cfg.MTU)
		}
	})
}

func TestPeerConfigStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		peer := PeerConfig{
			ID:                  "peer-123",
			PublicKey:           "pubkey123",
			Endpoint:            "10.0.0.2:51820",
			AllowedIPs:          []string{"10.0.0.2/32"},
			PresharedKey:        "psk123",
			PersistentKeepalive: 25,
			Name:                "Test Peer",
			Hostname:            "test.local",
		}

		if peer.ID != "peer-123" {
			t.Errorf("Expected ID 'peer-123', got %s", peer.ID)
		}
		if peer.PublicKey != "pubkey123" {
			t.Errorf("Expected PublicKey 'pubkey123', got %s", peer.PublicKey)
		}
		if peer.PersistentKeepalive != 25 {
			t.Errorf("Expected PersistentKeepalive 25, got %d", peer.PersistentKeepalive)
		}
	})
}

func TestSignalCallbacksStruct(t *testing.T) {
	t.Run("All Callback Fields", func(t *testing.T) {
		callbacks := SignalCallbacks{
			onOffer:     func(sourceID, ufrag, pwd string) {},
			onAnswer:    func(sourceID, ufrag, pwd string) {},
			onCandidate: func(sourceID, candidate string) {},
		}

		if callbacks.onOffer == nil {
			t.Error("Expected onOffer callback to be set")
		}
		if callbacks.onAnswer == nil {
			t.Error("Expected onAnswer callback to be set")
		}
		if callbacks.onCandidate == nil {
			t.Error("Expected onCandidate callback to be set")
		}
	})
}

func TestRegisterDeviceRequestStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		req := RegisterDeviceRequest{
			Name:      "test-device",
			Platform:  "linux",
			PubKey:    "pubkey123",
			HostName:  "test-host",
			OSVersion: "Ubuntu 22.04",
			DaemonVer: "1.0.0",
		}

		if req.Name != "test-device" {
			t.Errorf("Expected Name 'test-device', got %s", req.Name)
		}
		if req.Platform != "linux" {
			t.Errorf("Expected Platform 'linux', got %s", req.Platform)
		}
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		req := RegisterDeviceRequest{
			Name:     "test",
			Platform: "linux",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var unmarshaled RegisterDeviceRequest
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if unmarshaled.Name != req.Name {
			t.Errorf("Expected Name %s, got %s", req.Name, unmarshaled.Name)
		}
	})
}

func TestHeartbeatRequestStruct(t *testing.T) {
	t.Run("Has Optional Fields", func(t *testing.T) {
		req := HeartbeatRequest{
			IPAddress: "10.0.0.1",
			DaemonVer: "1.0.0",
			OSVersion: "Linux",
		}

		if req.IPAddress != "10.0.0.1" {
			t.Errorf("Expected IPAddress '10.0.0.1', got %s", req.IPAddress)
		}
	})

	t.Run("Empty Fields Are Omitted", func(t *testing.T) {
		req := HeartbeatRequest{}
		data, _ := json.Marshal(req)
		
		// All fields have omitempty, so empty request should be "{}" or close to it
		if len(data) > 50 {
			t.Errorf("Expected small JSON for empty request, got %s", string(data))
		}
	})
}

func TestNetworkResponseStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		network := NetworkResponse{
			ID:         "net-123",
			Name:       "Test Network",
			InviteCode: "ABC123",
			Role:       "owner",
		}

		if network.ID != "net-123" {
			t.Errorf("Expected ID 'net-123', got %s", network.ID)
		}
		if network.Name != "Test Network" {
			t.Errorf("Expected Name 'Test Network', got %s", network.Name)
		}
		if network.Role != "owner" {
			t.Errorf("Expected Role 'owner', got %s", network.Role)
		}
	})
}

func TestInviteTokenResponseStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		invite := InviteTokenResponse{
			ID:        "invite-123",
			NetworkID: "net-123",
			Token:     "token123",
			InviteURL: "https://example.com/join/token123",
			UsesMax:   10,
			UsesLeft:  7,
			IsActive:  true,
		}

		if invite.ID != "invite-123" {
			t.Errorf("Expected ID 'invite-123', got %s", invite.ID)
		}
		if invite.Token != "token123" {
			t.Errorf("Expected Token 'token123', got %s", invite.Token)
		}
		if !invite.IsActive {
			t.Error("Expected IsActive to be true")
		}
	})
}

// ==================== JSON Marshaling Tests ====================

func TestDeviceConfigJSON(t *testing.T) {
	t.Run("Marshals Correctly", func(t *testing.T) {
		cfg := DeviceConfig{
			Interface: InterfaceConfig{
				ListenPort: 51820,
				Addresses:  []string{"10.0.0.1/24"},
			},
			Peers: []PeerConfig{
				{ID: "peer-1", PublicKey: "key1"},
			},
		}

		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		if !strings.Contains(string(data), "51820") {
			t.Error("Expected JSON to contain listen port")
		}
		if !strings.Contains(string(data), "peer-1") {
			t.Error("Expected JSON to contain peer ID")
		}
	})
}

func TestPeerConfigJSON(t *testing.T) {
	t.Run("Marshals And Unmarshals", func(t *testing.T) {
		original := PeerConfig{
			ID:         "peer-123",
			PublicKey:  "pubkey",
			AllowedIPs: []string{"10.0.0.1/32"},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var unmarshaled PeerConfig
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if unmarshaled.ID != original.ID {
			t.Errorf("Expected ID %s, got %s", original.ID, unmarshaled.ID)
		}
		if len(unmarshaled.AllowedIPs) != 1 {
			t.Errorf("Expected 1 allowed IP, got %d", len(unmarshaled.AllowedIPs))
		}
	})
}

// ==================== Client Fields Tests ====================

func TestClientStruct(t *testing.T) {
	t.Run("Has All Required Fields After Init", func(t *testing.T) {
		cfg := &config.Config{
			Server: struct {
				URL string `yaml:"url"`
			}{URL: "http://localhost"},
		}
		client := NewClient(cfg)

		if client.config == nil {
			t.Error("Expected config to be set")
		}
		if client.httpClient == nil {
			t.Error("Expected httpClient to be set")
		}
		if client.stopChan == nil {
			t.Error("Expected stopChan to be set")
		}
	})
}

// ==================== CloseWebSocket Tests ====================

func TestCloseWebSocket_NoConnection(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	// Should not panic with nil connection
	client.CloseWebSocket()
}

// ==================== Functional Tests ====================

func setupMockClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: server.URL},
	}
	
	// Initialize memory keyring for tests
	kr, err := storage.NewTestKeyring(t.TempDir())
	require.NoError(t, err)
	cfg.Keyring = kr
	// Store a dummy token
	err = kr.StoreAuthToken("valid-token")
	require.NoError(t, err)

	return NewClient(cfg), server
}

func TestGetNetworks_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Errorf("Unexpected Auth header: %s", r.Header.Get("Authorization"))
		}
		
		_ = json.NewEncoder(w).Encode([]NetworkResponse{
			{ID: "net-1", Name: "Network 1", Role: "admin"},
			{ID: "net-2", Name: "Network 2", Role: "member"},
		})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	networks, err := client.GetNetworks(context.Background())
	if err != nil {
		t.Fatalf("GetNetworks failed: %v", err)
	}

	if len(networks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(networks))
	}
	if networks[0].ID != "net-1" || networks[1].ID != "net-2" {
		t.Error("Network IDs mismatch")
	}
}

func TestGetNetwork_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/net-123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Errorf("Unexpected Auth header: %s", r.Header.Get("Authorization"))
		}
		
		json.NewEncoder(w).Encode(NetworkResponse{ID: "net-123", Name: "Test Network", Role: "admin"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	network, err := client.GetNetwork(context.Background(), "net-123")
	if err != nil {
		t.Fatalf("GetNetwork failed: %v", err)
	}

	if network.ID != "net-123" {
		t.Errorf("Expected net-123, got %s", network.ID)
	}
	if network.Name != "Test Network" {
		t.Errorf("Expected Test Network, got %s", network.Name)
	}
}

func TestGetNetwork_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GetNetwork(context.Background(), "net-unknown")
	if err == nil {
		t.Fatal("Expected error for not found network")
	}
	if err.Error() != "network not found" {
		t.Errorf("Expected 'network not found', got: %s", err.Error())
	}
}

func TestDeleteNetwork_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/net-123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Errorf("Unexpected Auth header: %s", r.Header.Get("Authorization"))
		}
		
		w.WriteHeader(http.StatusNoContent)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.DeleteNetwork(context.Background(), "net-123")
	if err != nil {
		t.Fatalf("DeleteNetwork failed: %v", err)
	}
}

func TestDeleteNetwork_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.DeleteNetwork(context.Background(), "net-unknown")
	if err == nil {
		t.Fatal("Expected error for not found network")
	}
	if err.Error() != "network not found" {
		t.Errorf("Expected 'network not found', got: %s", err.Error())
	}
}

func TestDeleteNetwork_Forbidden(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.DeleteNetwork(context.Background(), "net-other")
	if err == nil {
		t.Fatal("Expected error for forbidden")
	}
	if err.Error() != "only the network owner can delete the network" {
		t.Errorf("Expected permission error, got: %s", err.Error())
	}
}

func TestJoinNetwork_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/join" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		
		var req JoinNetworkRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.InviteCode != "ABC12345" {
			t.Errorf("Expected invite code ABC12345, got %s", req.InviteCode)
		}

		_ = json.NewEncoder(w).Encode(NetworkResponse{ID: "net-join", Name: "Joined Net"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	net, err := client.JoinNetwork(context.Background(), "ABC12345")
	if err != nil {
		t.Fatalf("JoinNetwork failed: %v", err)
	}
	if net.ID != "net-join" {
		t.Errorf("Expected network ID net-join, got %s", net.ID)
	}
}

func TestLeaveNetwork_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/net-123/leave" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.LeaveNetwork(context.Background(), "net-123")
	if err != nil {
		t.Fatalf("LeaveNetwork failed: %v", err)
	}
}

func TestGetConfig_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/devices/dev-123/config" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		
		cfg := DeviceConfig{
			Interface: InterfaceConfig{ListenPort: 51820},
		}
		json.NewEncoder(w).Encode(cfg)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	cfg, err := client.GetConfig(context.Background(), "dev-123")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg.Interface.ListenPort != 51820 {
		t.Errorf("Expected port 51820, got %d", cfg.Interface.ListenPort)
	}
}

func TestPeerManagement_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/members/peer-1"):
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE, got %s", r.Method)
			}
			// Kick
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/members/peer-2/ban") && r.Method == "POST":
			// Ban
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/members/peer-3/ban") && r.Method == "DELETE":
			// Unban
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	// Test Kick
	if err := client.KickPeer(context.Background(), "net-1", "peer-1", "reason"); err != nil {
		t.Errorf("KickPeer failed: %v", err)
	}

	// Test Ban
	if err := client.BanPeer(context.Background(), "net-1", "peer-2", "reason"); err != nil {
		t.Errorf("BanPeer failed: %v", err)
	}

	// Test Unban
	if err := client.UnbanPeer(context.Background(), "net-1", "peer-3"); err != nil {
		t.Errorf("UnbanPeer failed: %v", err)
	}
}

func TestGenerateInvite_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/networks/net-1/invites" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req CreateInviteRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		
		// 24 hours * 3600
		if req.ExpiresIn != 86400 {
			t.Errorf("Expected ExpiresIn 86400, got %d", req.ExpiresIn)
		}

		_ = json.NewEncoder(w).Encode(InviteTokenResponse{Token: "inv-token"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	resp, err := client.GenerateInvite(context.Background(), "net-1", 1, 24)
	if err != nil {
		t.Fatalf("GenerateInvite failed: %v", err)
	}
	if resp.Token != "inv-token" {
		t.Errorf("Expected token inv-token, got %s", resp.Token)
	}
}

// ==================== SendHeartbeat Tests ====================

func TestSendHeartbeat_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/devices/device-123/heartbeat", r.URL.Path)
		assert.Equal(t, "Bearer valid-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req HeartbeatRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.100", req.IPAddress)
		assert.Equal(t, "1.0.0", req.DaemonVer)

		w.WriteHeader(http.StatusOK)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.SendHeartbeat(context.Background(), "device-123", HeartbeatRequest{
		IPAddress: "192.168.1.100",
		DaemonVer: "1.0.0",
		OSVersion: "Linux 5.15",
	})
	require.NoError(t, err)
}

func TestSendHeartbeat_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal error"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.SendHeartbeat(context.Background(), "device-123", HeartbeatRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Internal error")
}

func TestSendHeartbeat_ServerErrorNoMessage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.SendHeartbeat(context.Background(), "device-123", HeartbeatRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 400")
}

func TestSendHeartbeat_NetworkError(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost:99999"},
	}
	kr, _ := storage.NewTestKeyring(t.TempDir())
	_ = kr.StoreAuthToken("token")
	cfg.Keyring = kr
	client := NewClient(cfg)

	err := client.SendHeartbeat(context.Background(), "device-123", HeartbeatRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send http request")
}

// ==================== CreateNetwork Tests ====================

func TestCreateNetwork_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/networks", r.URL.Path)
		assert.Equal(t, "Bearer valid-token", r.Header.Get("Authorization"))

		var req CreateNetworkRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "My Network", req.Name)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(NetworkResponse{
			ID:         "net-created",
			Name:       "My Network",
			InviteCode: "ABC123",
			Role:       "owner",
		})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	resp, err := client.CreateNetwork(context.Background(), "My Network")
	require.NoError(t, err)
	assert.Equal(t, "net-created", resp.ID)
	assert.Equal(t, "My Network", resp.Name)
	assert.Equal(t, "owner", resp.Role)
}

func TestCreateNetwork_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Network name already exists"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.CreateNetwork(context.Background(), "Existing Network")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Network name already exists")
}

func TestCreateNetwork_ServerErrorNoMessage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.CreateNetwork(context.Background(), "Test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 403")
}

func TestCreateNetwork_InvalidResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("invalid json"))
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.CreateNetwork(context.Background(), "Test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

// ==================== WebSocket Tests ====================

// setupWebSocketServer creates a test server that upgrades HTTP to WebSocket
func setupWebSocketServer(t *testing.T, handler func(*websocket.Conn)) (*Client, *httptest.Server, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		handler(conn)
	})

	server := httptest.NewServer(mux)

	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: server.URL},
	}
	kr, _ := storage.NewTestKeyring(t.TempDir())
	kr.StoreAuthToken("valid-token")
	cfg.Keyring = kr
	client := NewClient(cfg)

	return client, server, func() {
		client.CloseWebSocket()
		server.Close()
	}
}

func TestStartWebSocket_Success(t *testing.T) {
	connected := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		connected <- true
		// Keep connection alive briefly
		time.Sleep(100 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-connected:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("WebSocket connection timed out")
	}
}

func TestStartWebSocket_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "://invalid-url"},
	}
	kr, _ := storage.NewTestKeyring(t.TempDir())
	kr.StoreAuthToken("token")
	cfg.Keyring = kr
	client := NewClient(cfg)

	err := client.StartWebSocket(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server URL")
}

func TestStartWebSocket_ConnectionRefused(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost:59999"},
	}
	kr, _ := storage.NewTestKeyring(t.TempDir())
	kr.StoreAuthToken("token")
	cfg.Keyring = kr
	client := NewClient(cfg)

	err := client.StartWebSocket(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "websocket dial failed")
}

func TestStartWebSocket_HTTPSUpgrade(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "https://localhost:59999"},
	}
	kr, _ := storage.NewTestKeyring(t.TempDir())
	kr.StoreAuthToken("token")
	cfg.Keyring = kr
	client := NewClient(cfg)

	// This will fail to connect, but we're testing the URL scheme conversion
	err := client.StartWebSocket(context.Background())
	require.Error(t, err)
	// The scheme should be converted to wss for https
}

func TestCloseWebSocket_WithConnection(t *testing.T) {
	closed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Wait for close or message
		_, _, err := conn.ReadMessage()
		if err != nil {
			closed <- true
		}
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	// Give it time to establish
	time.Sleep(50 * time.Millisecond)

	// Now close
	client.CloseWebSocket()

	select {
	case <-closed:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("WebSocket close timed out")
	}

	cleanup()
}

func TestCloseWebSocket_MultipleCalls(t *testing.T) {
	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		time.Sleep(100 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	// Multiple close calls should not panic
	client.CloseWebSocket()
	client.CloseWebSocket()
	client.CloseWebSocket()
}

// ==================== readLoop and handleMessage Tests ====================

func TestReadLoop_ReceivesOffer(t *testing.T) {
	offerReceived := make(chan bool, 1)
	var receivedSourceID, receivedUfrag, receivedPwd string

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send an offer message
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test-ufrag", Pwd: "test-pwd"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "user-123", Signal: signalPayload})
		msg := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg)

		// Keep alive for the client to process
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		receivedSourceID = sourceID
		receivedUfrag = ufrag
		receivedPwd = pwd
		offerReceived <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-offerReceived:
		assert.Equal(t, "user-123", receivedSourceID)
		assert.Equal(t, "test-ufrag", receivedUfrag)
		assert.Equal(t, "test-pwd", receivedPwd)
	case <-time.After(2 * time.Second):
		t.Fatal("Offer callback not called")
	}
}

func TestReadLoop_ReceivesAnswer(t *testing.T) {
	answerReceived := make(chan bool, 1)
	var receivedSourceID, receivedUfrag, receivedPwd string

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "answer-ufrag", Pwd: "answer-pwd"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "peer-456", Signal: signalPayload})
		msg := wsMessage{Type: "call.answer", Data: signalData}
		conn.WriteJSON(msg)
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnAnswer(func(sourceID, ufrag, pwd string) {
		receivedSourceID = sourceID
		receivedUfrag = ufrag
		receivedPwd = pwd
		answerReceived <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-answerReceived:
		assert.Equal(t, "peer-456", receivedSourceID)
		assert.Equal(t, "answer-ufrag", receivedUfrag)
		assert.Equal(t, "answer-pwd", receivedPwd)
	case <-time.After(2 * time.Second):
		t.Fatal("Answer callback not called")
	}
}

func TestReadLoop_ReceivesICECandidate(t *testing.T) {
	candidateReceived := make(chan bool, 1)
	var receivedSourceID, receivedCandidate string

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		signalPayload, _ := json.Marshal(signalPayload{Candidate: "candidate:1234567890"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "peer-789", Signal: signalPayload})
		msg := wsMessage{Type: "call.ice", Data: signalData}
		conn.WriteJSON(msg)
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnCandidate(func(sourceID, candidate string) {
		receivedSourceID = sourceID
		receivedCandidate = candidate
		candidateReceived <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-candidateReceived:
		assert.Equal(t, "peer-789", receivedSourceID)
		assert.Equal(t, "candidate:1234567890", receivedCandidate)
	case <-time.After(2 * time.Second):
		t.Fatal("Candidate callback not called")
	}
}

func TestReadLoop_IgnoresUnknownMessageType(t *testing.T) {
	unknownProcessed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send unknown message type
		msg := wsMessage{Type: "unknown.type", Data: []byte(`{"foo":"bar"}`)}
		conn.WriteJSON(msg)

		// Then send a known type to verify the loop continues
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
		msg2 := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg2)

		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		unknownProcessed <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-unknownProcessed:
		// Unknown message was skipped and processing continued
	case <-time.After(2 * time.Second):
		t.Fatal("Processing did not continue after unknown message")
	}
}

func TestReadLoop_HandlesInvalidJSON(t *testing.T) {
	processed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send invalid JSON
		conn.WriteMessage(websocket.TextMessage, []byte("not json"))

		// Then send valid message
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
		msg := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg)

		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		processed <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-processed:
		// Invalid JSON was handled gracefully
	case <-time.After(2 * time.Second):
		t.Fatal("Processing did not continue after invalid JSON")
	}
}

func TestReadLoop_HandlesInvalidSignalData(t *testing.T) {
	processed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send message with invalid signal data
		msg := wsMessage{Type: "call.offer", Data: []byte("invalid signal")}
		conn.WriteJSON(msg)

		// Then send valid message
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
		msg2 := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg2)

		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		processed <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-processed:
		// Invalid signal data was handled gracefully
	case <-time.After(2 * time.Second):
		t.Fatal("Processing did not continue after invalid signal data")
	}
}

func TestReadLoop_HandlesInvalidSignalPayload(t *testing.T) {
	processed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send message with invalid signal payload
		signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: []byte("invalid payload")})
		msg := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg)

		// Then send valid message
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
		signalData2, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
		msg2 := wsMessage{Type: "call.offer", Data: signalData2}
		conn.WriteJSON(msg2)

		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		processed <- true
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-processed:
		// Invalid payload was handled gracefully
	case <-time.After(2 * time.Second):
		t.Fatal("Processing did not continue after invalid payload")
	}
}

func TestReadLoop_NoCallbackSet(t *testing.T) {
	processed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Send offer without callback set
		signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
		signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
		msg := wsMessage{Type: "call.offer", Data: signalData}
		conn.WriteJSON(msg)

		// Give time to process
		time.Sleep(100 * time.Millisecond)
		processed <- true
	})
	defer cleanup()

	// Don't set any callbacks - should not panic

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	select {
	case <-processed:
		// No panic, message was ignored
	case <-time.After(2 * time.Second):
		t.Fatal("Processing did not complete")
	}
}

func TestReadLoop_ServerCloseConnection(t *testing.T) {
	closed := make(chan bool, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		// Close connection immediately
		conn.Close()
	})

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)

	// The readLoop should exit when connection is closed
	go func() {
		time.Sleep(500 * time.Millisecond)
		closed <- true
	}()

	select {
	case <-closed:
		// readLoop exited gracefully
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop did not exit after connection closed")
	}

	cleanup()
}

// ==================== sendSignal Tests ====================

func TestSendSignal_Success(t *testing.T) {
	messageReceived := make(chan wsMessage, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		var msg wsMessage
		err := conn.ReadJSON(&msg)
		if err == nil {
			messageReceived <- msg
		}
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	err = client.SendOffer("target-peer", "ufrag123", "pwd456")
	require.NoError(t, err)

	select {
	case msg := <-messageReceived:
		assert.Equal(t, "call.offer", msg.Type)
		var signalData callSignalData
		json.Unmarshal(msg.Data, &signalData)
		assert.Equal(t, "target-peer", signalData.TargetID)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received by server")
	}
}

func TestSendAnswer_Success(t *testing.T) {
	messageReceived := make(chan wsMessage, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		var msg wsMessage
		err := conn.ReadJSON(&msg)
		if err == nil {
			messageReceived <- msg
		}
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	err = client.SendAnswer("target-peer", "answer-ufrag", "answer-pwd")
	require.NoError(t, err)

	select {
	case msg := <-messageReceived:
		assert.Equal(t, "call.answer", msg.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received by server")
	}
}

func TestSendCandidate_Success(t *testing.T) {
	messageReceived := make(chan wsMessage, 1)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		var msg wsMessage
		err := conn.ReadJSON(&msg)
		if err == nil {
			messageReceived <- msg
		}
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	err = client.SendCandidate("target-peer", "candidate:1234567890")
	require.NoError(t, err)

	select {
	case msg := <-messageReceived:
		assert.Equal(t, "call.ice", msg.Type)
		var signalData callSignalData
		json.Unmarshal(msg.Data, &signalData)
		var payload signalPayload
		json.Unmarshal(signalData.Signal, &payload)
		assert.Equal(t, "candidate:1234567890", payload.Candidate)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received by server")
	}
}

func TestSendSignal_ConcurrentSends(t *testing.T) {
	messagesReceived := make(chan wsMessage, 10)

	client, _, cleanup := setupWebSocketServer(t, func(conn *websocket.Conn) {
		for i := 0; i < 5; i++ {
			var msg wsMessage
			err := conn.ReadJSON(&msg)
			if err == nil {
				messagesReceived <- msg
			}
		}
		time.Sleep(200 * time.Millisecond)
	})
	defer cleanup()

	err := client.StartWebSocket(context.Background())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	// Send multiple messages concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client.SendOffer("target", "ufrag", "pwd")
		}(i)
	}
	wg.Wait()

	// Wait for all messages
	count := 0
	timeout := time.After(2 * time.Second)
	for count < 5 {
		select {
		case <-messagesReceived:
			count++
		case <-timeout:
			t.Fatalf("Only received %d of 5 messages", count)
		}
	}
	assert.Equal(t, 5, count)
}

// ==================== GetConfig Tests ====================

func TestGetConfig_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Access denied"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GetConfig(context.Background(), "device-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Access denied")
}

func TestGetConfig_InvalidResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GetConfig(context.Background(), "device-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

// ==================== JoinNetwork Tests ====================

func TestJoinNetwork_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid invite code"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.JoinNetwork(context.Background(), "INVALID")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid invite code")
}

func TestJoinNetwork_ServerErrorNoMessage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.JoinNetwork(context.Background(), "INVALID")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 404")
}

// ==================== LeaveNetwork Tests ====================

func TestLeaveNetwork_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.LeaveNetwork(context.Background(), "net-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 403")
}

// ==================== Peer Management Error Tests ====================

func TestKickPeer_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.KickPeer(context.Background(), "net-1", "peer-1", "reason")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 403")
}

func TestBanPeer_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.BanPeer(context.Background(), "net-1", "peer-1", "reason")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 403")
}

func TestUnbanPeer_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	err := client.UnbanPeer(context.Background(), "net-1", "peer-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 404")
}

// ==================== GenerateInvite Error Tests ====================

func TestGenerateInvite_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not authorized"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GenerateInvite(context.Background(), "net-1", 10, 24)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Not authorized")
}

func TestGenerateInvite_ServerErrorNoMessage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GenerateInvite(context.Background(), "net-1", 10, 24)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 500")
}

func TestGenerateInvite_ZeroExpiry(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateInviteRequest
		json.NewDecoder(r.Body).Decode(&req)
		// When expiresHours is 0, ExpiresIn should be 0 (server default)
		assert.Equal(t, 0, req.ExpiresIn)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(InviteTokenResponse{Token: "token"})
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GenerateInvite(context.Background(), "net-1", 10, 0)
	require.NoError(t, err)
}

// ==================== GetNetworks Error Tests ====================

func TestGetNetworks_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GetNetworks(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status: 500")
}

func TestGetNetworks_InvalidResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	})

	client, server := setupMockClient(t, handler)
	defer server.Close()

	_, err := client.GetNetworks(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

// ==================== handleMessage Tests ====================

func TestHandleMessage_DirectCall(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	// Test direct call with invalid JSON
	client.handleMessage([]byte("invalid json"))
	// Should not panic

	// Test with nil callbacks
	signalPayload, _ := json.Marshal(signalPayload{Ufrag: "test", Pwd: "test"})
	signalData, _ := json.Marshal(callSignalData{FromUser: "user", Signal: signalPayload})
	msg, _ := json.Marshal(wsMessage{Type: "call.offer", Data: signalData})
	client.handleMessage(msg)
	// Should not panic when callbacks are nil
}

func TestHandleMessage_AllMessageTypes(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			URL string `yaml:"url"`
		}{URL: "http://localhost"},
	}
	client := NewClient(cfg)

	offerCalled := false
	answerCalled := false
	candidateCalled := false

	client.OnOffer(func(sourceID, ufrag, pwd string) {
		offerCalled = true
	})
	client.OnAnswer(func(sourceID, ufrag, pwd string) {
		answerCalled = true
	})
	client.OnCandidate(func(sourceID, candidate string) {
		candidateCalled = true
	})

	// Test offer
	payload, _ := json.Marshal(signalPayload{Ufrag: "u", Pwd: "p"})
	data, _ := json.Marshal(callSignalData{FromUser: "user", Signal: payload})
	msg, _ := json.Marshal(wsMessage{Type: "call.offer", Data: data})
	client.handleMessage(msg)
	assert.True(t, offerCalled)

	// Test answer
	msg, _ = json.Marshal(wsMessage{Type: "call.answer", Data: data})
	client.handleMessage(msg)
	assert.True(t, answerCalled)

	// Test ICE candidate
	payload, _ = json.Marshal(signalPayload{Candidate: "candidate"})
	data, _ = json.Marshal(callSignalData{FromUser: "user", Signal: payload})
	msg, _ = json.Marshal(wsMessage{Type: "call.ice", Data: data})
	client.handleMessage(msg)
	assert.True(t, candidateCalled)
}
