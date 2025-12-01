package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/config"
)

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
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "test-device" {
			t.Errorf("Expected name 'test-device', got %s", req.Name)
		}
		if req.Platform != "windows" {
			t.Errorf("Expected platform 'windows', got %s", req.Platform)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedResponse)
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
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
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
		w.Write([]byte("invalid json"))
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

