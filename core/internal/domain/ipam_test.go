package domain

import (
	"testing"
)

// Test NextIP with valid CIDR and various offsets
func TestNextIP_Success(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		offset   uint32
		expected string
	}{
		{
			name:     "First usable IP in /24",
			cidr:     "10.0.0.0/24",
			offset:   1,
			expected: "10.0.0.1",
		},
		{
			name:     "Second IP in /24",
			cidr:     "10.0.0.0/24",
			offset:   2,
			expected: "10.0.0.2",
		},
		{
			name:     "Last usable IP in /24",
			cidr:     "10.0.0.0/24",
			offset:   254,
			expected: "10.0.0.254",
		},
		{
			name:     "First IP in /16",
			cidr:     "192.168.0.0/16",
			offset:   1,
			expected: "192.168.0.1",
		},
		{
			name:     "Mid-range IP in /16",
			cidr:     "192.168.0.0/16",
			offset:   256,
			expected: "192.168.1.0",
		},
		{
			name:     "First IP in /30 (2 usable hosts)",
			cidr:     "10.0.0.0/30",
			offset:   1,
			expected: "10.0.0.1",
		},
		{
			name:     "Second IP in /30",
			cidr:     "10.0.0.0/30",
			offset:   2,
			expected: "10.0.0.2",
		},
		{
			name:     "Non-zero base network",
			cidr:     "172.16.10.0/24",
			offset:   1,
			expected: "172.16.10.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := NextIP(tt.cidr, tt.offset)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if ip != tt.expected {
				t.Errorf("expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

// Test NextIP with invalid CIDRs
func TestNextIP_InvalidCIDR(t *testing.T) {
	tests := []struct {
		name string
		cidr string
	}{
		{
			name: "Invalid CIDR format",
			cidr: "invalid-cidr",
		},
		{
			name: "Missing mask",
			cidr: "10.0.0.0",
		},
		{
			name: "Invalid IP",
			cidr: "999.999.999.999/24",
		},
		{
			name: "Empty string",
			cidr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NextIP(tt.cidr, 1)
			if err == nil {
				t.Error("expected error for invalid CIDR, got nil")
			}
		})
	}
}

// Test NextIP with offset edge cases
func TestNextIP_OffsetEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		cidr          string
		offset        uint32
		expectEmpty   bool
		expectError   bool
		errorContains string
	}{
		{
			name:        "Offset 0 (network address)",
			cidr:        "10.0.0.0/24",
			offset:      0,
			expectEmpty: true,
		},
		{
			name:        "Offset beyond range in /24",
			cidr:        "10.0.0.0/24",
			offset:      255, // Would be broadcast
			expectEmpty: true,
		},
		{
			name:        "Offset way beyond range",
			cidr:        "10.0.0.0/24",
			offset:      1000,
			expectEmpty: true,
		},
		{
			name:        "Offset beyond range in /30",
			cidr:        "10.0.0.0/30",
			offset:      3, // Only 2 usable hosts (offset 1, 2)
			expectEmpty: true,
		},
		{
			name:        "/32 has no usable hosts",
			cidr:        "10.0.0.1/32",
			offset:      1,
			expectEmpty: true,
		},
		{
			name:        "/31 has no usable hosts",
			cidr:        "10.0.0.0/31",
			offset:      1,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := NextIP(tt.cidr, tt.offset)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tt.expectEmpty && ip != "" {
				t.Errorf("expected empty IP, got: %s", ip)
			}
		})
	}
}

// Test NextIP excludes network and broadcast addresses
func TestNextIP_ExcludesNetworkAndBroadcast(t *testing.T) {
	// /24 network: 10.0.0.0 (network), 10.0.0.1-254 (usable), 10.0.0.255 (broadcast)
	cidr := "10.0.0.0/24"

	// Network address (offset 0) should be excluded
	ip, err := NextIP(cidr, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "" {
		t.Errorf("offset 0 should return empty (network address), got: %s", ip)
	}

	// Broadcast address should be excluded (offset 255 in /24)
	ip, err = NextIP(cidr, 255)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "" {
		t.Errorf("offset 255 should return empty (broadcast), got: %s", ip)
	}

	// Last usable IP (offset 254) should work
	ip, err = NextIP(cidr, 254)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.0.0.254" {
		t.Errorf("expected 10.0.0.254, got: %s", ip)
	}
}

// Test NextIP with various subnet masks
func TestNextIP_VariousSubnetMasks(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		offset      uint32
		expected    string
		usableHosts uint32 // Expected number of usable hosts
	}{
		{
			name:        "/8 network (16M hosts)",
			cidr:        "10.0.0.0/8",
			offset:      1,
			expected:    "10.0.0.1",
			usableHosts: 16777214, // 2^24 - 2
		},
		{
			name:        "/16 network (64K hosts)",
			cidr:        "172.16.0.0/16",
			offset:      1,
			expected:    "172.16.0.1",
			usableHosts: 65534, // 2^16 - 2
		},
		{
			name:        "/24 network (254 hosts)",
			cidr:        "192.168.1.0/24",
			offset:      1,
			expected:    "192.168.1.1",
			usableHosts: 254, // 2^8 - 2
		},
		{
			name:        "/29 network (6 hosts)",
			cidr:        "10.0.0.0/29",
			offset:      1,
			expected:    "10.0.0.1",
			usableHosts: 6, // 2^3 - 2
		},
		{
			name:        "/30 network (2 hosts)",
			cidr:        "10.0.0.0/30",
			offset:      1,
			expected:    "10.0.0.1",
			usableHosts: 2, // 2^2 - 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test first usable IP
			ip, err := NextIP(tt.cidr, tt.offset)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if ip != tt.expected {
				t.Errorf("expected IP %s, got %s", tt.expected, ip)
			}

			// Test last usable IP
			lastIP, err := NextIP(tt.cidr, tt.usableHosts)
			if err != nil {
				t.Fatalf("expected no error for last usable IP, got: %v", err)
			}
			if lastIP == "" {
				t.Error("expected valid IP for last usable host, got empty")
			}

			// Test one beyond last (should be empty - broadcast)
			beyondLast, err := NextIP(tt.cidr, tt.usableHosts+1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if beyondLast != "" {
				t.Errorf("expected empty for offset beyond range, got: %s", beyondLast)
			}
		})
	}
}

// Test IPv6 handling (currently not supported)
func TestNextIP_IPv6NotSupported(t *testing.T) {
	cidr := "2001:db8::/64"
	_, err := NextIP(cidr, 1)
	if err == nil {
		t.Error("expected error for IPv6 CIDR")
	}
	if !contains(err.Error(), "IPv4") {
		t.Errorf("expected error mentioning IPv4, got: %v", err)
	}
}

// Test IP calculation correctness across byte boundaries
func TestNextIP_ByteBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		offset   uint32
		expected string
	}{
		{
			name:     "Cross into next octet",
			cidr:     "10.0.0.0/24",
			offset:   255 - 1, // 10.0.0.254 (last valid in /24)
			expected: "10.0.0.254",
		},
		{
			name:     "First IP after 255",
			cidr:     "10.0.0.0/16",
			offset:   256,
			expected: "10.0.1.0",
		},
		{
			name:     "Multiple byte boundaries",
			cidr:     "10.0.0.0/16",
			offset:   257,
			expected: "10.0.1.1",
		},
		{
			name:     "Large offset",
			cidr:     "10.0.0.0/16",
			offset:   65534, // Last usable in /16
			expected: "10.0.255.254",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := NextIP(tt.cidr, tt.offset)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if ip != tt.expected {
				t.Errorf("expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
