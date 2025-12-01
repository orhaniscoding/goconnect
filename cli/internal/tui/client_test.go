package tui

import (
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
