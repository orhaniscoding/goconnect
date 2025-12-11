package deeplink

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		h := NewHandler()
		require.NotNil(t, h)
		assert.Equal(t, 5*time.Second, h.connectTimeout)
		assert.Equal(t, 30*time.Second, h.requestTimeout)
		// grpcTarget should be set from platform-specific default
		assert.NotEmpty(t, h.grpcTarget)
	})

	t.Run("custom grpc target", func(t *testing.T) {
		h := NewHandler(WithGRPCTarget("custom:12345"))
		assert.Equal(t, "custom:12345", h.grpcTarget)
	})

	t.Run("custom connect timeout", func(t *testing.T) {
		h := NewHandler(WithConnectTimeout(10 * time.Second))
		assert.Equal(t, 10*time.Second, h.connectTimeout)
	})

	t.Run("custom request timeout", func(t *testing.T) {
		h := NewHandler(WithRequestTimeout(60 * time.Second))
		assert.Equal(t, 60*time.Second, h.requestTimeout)
	})

	t.Run("multiple options", func(t *testing.T) {
		h := NewHandler(
			WithGRPCTarget("localhost:9999"),
			WithConnectTimeout(3*time.Second),
			WithRequestTimeout(15*time.Second),
		)
		assert.Equal(t, "localhost:9999", h.grpcTarget)
		assert.Equal(t, 3*time.Second, h.connectTimeout)
		assert.Equal(t, 15*time.Second, h.requestTimeout)
	})
}

func TestHandler_Handle_NilDeepLink(t *testing.T) {
	h := NewHandler()
	result, err := h.Handle(nil)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deep link is nil")
}

func TestHandler_Handle_UnsupportedAction(t *testing.T) {
	h := NewHandler()
	dl := &DeepLink{
		Action: ActionUnknown,
		Target: "something",
	}
	result, err := h.Handle(dl)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported action")
}

func TestHandler_HandleLogin(t *testing.T) {
	h := NewHandler()
	dl := &DeepLink{
		Action: ActionLogin,
		Params: map[string]string{
			"token":  "test-token",
			"server": "https://api.test.com",
		},
	}

	result, err := h.Handle(dl)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, "Login parameters parsed", result.Message)
	assert.Equal(t, "test-token", result.Data["token"])
	assert.Equal(t, "https://api.test.com", result.Data["server"])
}

func TestHandler_Connect_FailsWhenDaemonNotRunning(t *testing.T) {
	// Test that handler returns error when daemon is not reachable
	h := NewHandler(
		WithGRPCTarget("127.0.0.1:19999"), // Non-existent port
		WithConnectTimeout(100*time.Millisecond),
	)

	dl := &DeepLink{
		Action: ActionJoin,
		Target: "test-invite-code",
	}

	result, err := h.Handle(dl)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to connect to daemon")
}

func TestHandler_HandleNetwork_FailsWhenDaemonNotRunning(t *testing.T) {
	h := NewHandler(
		WithGRPCTarget("127.0.0.1:19998"),
		WithConnectTimeout(100*time.Millisecond),
	)

	dl := &DeepLink{
		Action: ActionNetwork,
		Target: "test-network-id",
	}

	result, err := h.Handle(dl)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to connect to daemon")
}

func TestHandler_HandleConnect_FailsWhenDaemonNotRunning(t *testing.T) {
	h := NewHandler(
		WithGRPCTarget("127.0.0.1:19997"),
		WithConnectTimeout(100*time.Millisecond),
	)

	dl := &DeepLink{
		Action: ActionConnect,
		Target: "test-peer-id",
	}

	result, err := h.Handle(dl)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to connect to daemon")
}

func TestResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &Result{
			Success: true,
			Message: "Operation successful",
			Data: map[string]interface{}{
				"key": "value",
			},
		}
		assert.True(t, result.Success)
		assert.Equal(t, "Operation successful", result.Message)
		assert.Equal(t, "value", result.Data["key"])
	})

	t.Run("failure result", func(t *testing.T) {
		result := &Result{
			Success: false,
			Message: "Something went wrong",
			Data:    nil,
		}
		assert.False(t, result.Success)
		assert.Equal(t, "Something went wrong", result.Message)
		assert.Nil(t, result.Data)
	})
}

func TestGetDefaultGRPCTarget(t *testing.T) {
	// The default should be platform-specific
	target := getDefaultGRPCTarget()
	assert.NotEmpty(t, target)
	// On Unix-like systems, it should be a Unix socket
	// On Windows, it should be a TCP address
	// Just verify it's set to something reasonable
	assert.True(t,
		target == "unix:///tmp/goconnect-daemon.sock" || // Unix
			target == "127.0.0.1:34101", // Windows
		"unexpected default target: %s", target)
}
