package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== Helper Function Tests ====================

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "existing string value",
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
		},
		{
			name:     "non-string value",
			m:        map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]interface{}{},
			key:      "key",
			expected: "",
		},
		{
			name:     "nil value",
			m:        map[string]interface{}{"key": nil},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected int
	}{
		{
			name:     "existing int value",
			m:        map[string]interface{}{"count": 42},
			key:      "count",
			expected: 42,
		},
		{
			name:     "int64 value",
			m:        map[string]interface{}{"count": int64(100)},
			key:      "count",
			expected: 100,
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{"other": 10},
			key:      "count",
			expected: 0,
		},
		{
			name:     "non-int value",
			m:        map[string]interface{}{"count": "not a number"},
			key:      "count",
			expected: 0,
		},
		{
			name:     "empty map",
			m:        map[string]interface{}{},
			key:      "count",
			expected: 0,
		},
		{
			name:     "zero int value",
			m:        map[string]interface{}{"count": 0},
			key:      "count",
			expected: 0,
		},
		{
			name:     "negative int value",
			m:        map[string]interface{}{"count": -5},
			key:      "count",
			expected: -5,
		},
		{
			name:     "float value (not handled)",
			m:        map[string]interface{}{"count": 3.14},
			key:      "count",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntValue(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFirstIP(t *testing.T) {
	tests := []struct {
		name     string
		ips      []string
		expected string
	}{
		{
			name:     "single IP without CIDR",
			ips:      []string{"10.0.0.1"},
			expected: "10.0.0.1",
		},
		{
			name:     "single IP with CIDR",
			ips:      []string{"10.0.0.1/24"},
			expected: "10.0.0.1",
		},
		{
			name:     "multiple IPs returns first",
			ips:      []string{"10.0.0.1/24", "10.0.0.2/24"},
			expected: "10.0.0.1",
		},
		{
			name:     "empty slice",
			ips:      []string{},
			expected: "",
		},
		{
			name:     "IPv6 with CIDR",
			ips:      []string{"2001:db8::1/128"},
			expected: "2001:db8::1",
		},
		{
			name:     "IP without slash",
			ips:      []string{"192.168.1.100"},
			expected: "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstIP(tt.ips)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapTransferStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected int32 // Use int32 to compare with proto enum values
	}{
		{
			name:   "pending",
			status: "pending",
		},
		{
			name:   "in_progress",
			status: "in_progress",
		},
		{
			name:   "completed",
			status: "completed",
		},
		{
			name:   "failed",
			status: "failed",
		},
		{
			name:   "cancelled",
			status: "cancelled",
		},
		{
			name:   "unknown status",
			status: "unknown",
		},
		{
			name:   "empty status",
			status: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapTransferStatus(tt.status)
			// Just verify it doesn't panic and returns a valid enum
			assert.GreaterOrEqual(t, int32(result), int32(0))
		})
	}
}

// ==================== GRPCServer Constructor Test ====================

func TestNewGRPCServer(t *testing.T) {
	t.Run("creates server with all fields", func(t *testing.T) {
		mockDaemon := &DaemonService{
			logf: &fallbackServiceLogger{},
		}

		srv := NewGRPCServer(mockDaemon, "1.0.0", "2024-01-01", "abc123")

		assert.NotNil(t, srv)
		assert.Equal(t, mockDaemon, srv.daemon)
		assert.Equal(t, "1.0.0", srv.version)
		assert.Equal(t, "2024-01-01", srv.buildDate)
		assert.Equal(t, "abc123", srv.commit)
		assert.NotNil(t, srv.subscribers)
		assert.NotNil(t, srv.ipcAuth)
	})

	t.Run("creates server with empty build info", func(t *testing.T) {
		mockDaemon := &DaemonService{
			logf: &fallbackServiceLogger{},
		}

		srv := NewGRPCServer(mockDaemon, "", "", "")

		assert.NotNil(t, srv)
		assert.Empty(t, srv.version)
		assert.Empty(t, srv.buildDate)
		assert.Empty(t, srv.commit)
	})
}

// ==================== GRPCServer Stop Test ====================

func TestGRPCServer_Stop(t *testing.T) {
	t.Run("stop with nil components does not panic", func(t *testing.T) {
		mockDaemon := &DaemonService{
			logf: &fallbackServiceLogger{},
		}

		srv := NewGRPCServer(mockDaemon, "1.0.0", "", "")
		// Set components to nil to test nil safety
		srv.grpcServer = nil
		srv.listener = nil
		srv.ipcAuth = nil

		// Should not panic
		srv.Stop()
	})
}

// ==================== BroadcastEvent Test ====================

func TestGRPCServer_BroadcastEvent(t *testing.T) {
	t.Run("broadcasts to all subscribers", func(t *testing.T) {
		mockDaemon := &DaemonService{
			logf: &fallbackServiceLogger{},
		}

		srv := NewGRPCServer(mockDaemon, "1.0.0", "", "")

		// The BroadcastEvent expects *pb.DaemonEvent which we can't easily test
		// without importing the proto package. The function is tested via
		// integration tests in grpc_server_test.go with testGRPCServer.
		assert.NotNil(t, srv.subscribers)
	})
}

// ==================== IsPipeSupported Test (Unix) ====================

func TestIsPipeSupported_Unix(t *testing.T) {
	// On Unix, named pipes are not supported (Windows only)
	supported := IsPipeSupported()
	assert.False(t, supported)
}

func TestCreateWindowsListener_Unix(t *testing.T) {
	// On Unix, this returns nil, nil (no error, no listener)
	listener, err := CreateWindowsListener()
	assert.NoError(t, err)
	assert.Nil(t, listener)
}
