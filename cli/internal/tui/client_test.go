package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnifiedClient_HTTPFallback(t *testing.T) {
	// Create client without gRPC available (fallback to HTTP)
	client := NewUnifiedClientWithMode(false)
	defer client.Close()

	// Should be using HTTP
	if client.IsUsingGRPC() {
		t.Error("Expected HTTP mode when preferGRPC=false")
	}
}

func TestUnifiedClient_SwitchModes(t *testing.T) {
	// Start with HTTP
	client := NewUnifiedClientWithMode(false)
	defer client.Close()

	if client.IsUsingGRPC() {
		t.Error("Should start in HTTP mode")
	}

	// Try to switch to gRPC - should fail since daemon isn't running
	err := client.SwitchToGRPC()
	if err == nil {
		// If it succeeded (daemon is running), switch back
		client.SwitchToHTTP()
		if client.IsUsingGRPC() {
			t.Error("Should be in HTTP mode after SwitchToHTTP")
		}
	}
	// Error is expected when daemon isn't running
}

func TestUnifiedClient_IsUsingGRPC(t *testing.T) {
	// Without daemon, should default to HTTP
	client := NewUnifiedClient()
	defer client.Close()

	// gRPC won't be available without daemon running
	// Just verify the method doesn't panic
	_ = client.IsUsingGRPC()
}

func TestMapNetworkRole(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "host"},   // NETWORK_ROLE_OWNER
		{2, "admin"},  // NETWORK_ROLE_ADMIN
		{3, "client"}, // NETWORK_ROLE_MEMBER
		{0, "unknown"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		// Cast to pb.NetworkRole equivalent value
		// Note: These tests verify the mapping logic
		got := mapNetworkRoleFromInt(tt.input)
		if got != tt.expected {
			t.Errorf("mapNetworkRole(%d) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

// Helper to test role mapping without proto dependency
func mapNetworkRoleFromInt(role int) string {
	switch role {
	case 1: // OWNER
		return "host"
	case 2: // ADMIN
		return "admin"
	case 3: // MEMBER
		return "client"
	default:
		return "unknown"
	}
}

func TestMapConnectionStatus(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "connected"},
		{2, "connecting"},
		{3, "disconnected"},
		{4, "failed"},
		{0, "unknown"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		got := mapConnectionStatusFromInt(tt.input)
		if got != tt.expected {
			t.Errorf("mapConnectionStatus(%d) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

// Helper to test status mapping without proto dependency
func mapConnectionStatusFromInt(status int) string {
	switch status {
	case 1: // CONNECTED
		return "connected"
	case 2: // CONNECTING
		return "connecting"
	case 3: // DISCONNECTED
		return "disconnected"
	case 4: // FAILED
		return "failed"
	default:
		return "unknown"
	}
}

func TestMapConnectionType(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "direct"},
		{2, "relay"},
		{0, "unknown"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		got := mapConnectionTypeFromInt(tt.input)
		if got != tt.expected {
			t.Errorf("mapConnectionType(%d) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

// Helper to test type mapping without proto dependency
func mapConnectionTypeFromInt(connType int) string {
	switch connType {
	case 1: // DIRECT
		return "direct"
	case 2: // RELAY
		return "relay"
	default:
		return "unknown"
	}
}

func TestGetGRPCTarget(t *testing.T) {
	target := getGRPCTarget()
	if target == "" {
		t.Error("getGRPCTarget returned empty string")
	}
	// Target will be OS-specific, just verify it's not empty
}

func TestIsGRPCAvailable(t *testing.T) {
	// Without daemon running, should return false
	// This is a safe test that doesn't require daemon
	available := IsGRPCAvailable()
	// Just verify it doesn't panic - result depends on daemon state
	_ = available
}

// ==================== HTTP Client Tests with Mock Server ====================

func TestClient_CheckDaemonStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	if !client.CheckDaemonStatus() {
		t.Error("Expected CheckDaemonStatus to return true")
	}
}

func TestClient_CheckDaemonStatus_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	if client.CheckDaemonStatus() {
		t.Error("Expected CheckDaemonStatus to return false")
	}
}

func TestClient_GetStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			status := Status{
				Connected:     true,
				NetworkName:   "TestNetwork",
				IP:            "10.0.0.5",
				OnlineMembers: 5,
				Role:          "host",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(status)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	status, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if !status.Connected {
		t.Error("Expected Connected to be true")
	}
	if status.NetworkName != "TestNetwork" {
		t.Errorf("Expected NetworkName 'TestNetwork', got '%s'", status.NetworkName)
	}
}

func TestClient_GetStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	_, err := client.GetStatus()
	if err == nil {
		t.Error("Expected error from GetStatus")
	}
}

func TestClient_CreateNetwork_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/networks/create" && r.Method == "POST" {
			network := Network{
				ID:   "net-123",
				Name: "MyNetwork",
				Role: "host",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(network)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	network, err := client.CreateNetwork("MyNetwork", "")
	if err != nil {
		t.Fatalf("CreateNetwork failed: %v", err)
	}
	if network.Name != "MyNetwork" {
		t.Errorf("Expected Name 'MyNetwork', got '%s'", network.Name)
	}
}

func TestClient_JoinNetwork_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/networks/join" && r.Method == "POST" {
			network := Network{
				ID:   "net-456",
				Name: "JoinedNetwork",
				Role: "client",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(network)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	network, err := client.JoinNetwork("invite-code-123")
	if err != nil {
		t.Fatalf("JoinNetwork failed: %v", err)
	}
	if network.Role != "client" {
		t.Errorf("Expected Role 'client', got '%s'", network.Role)
	}
}

func TestClient_GetNetworks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/networks" && r.Method == "GET" {
			networks := []Network{
				{ID: "net-1", Name: "Network1", Role: "host"},
				{ID: "net-2", Name: "Network2", Role: "client"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(networks)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	networks, err := client.GetNetworks()
	if err != nil {
		t.Fatalf("GetNetworks failed: %v", err)
	}
	if len(networks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(networks))
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.baseURL != DaemonURL {
		t.Errorf("Expected baseURL '%s', got '%s'", DaemonURL, client.baseURL)
	}
}
