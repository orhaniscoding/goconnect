package tui

import (
	"context"
	"testing"

	pb "github.com/orhaniscoding/goconnect/client-daemon/internal/proto"
	"github.com/stretchr/testify/assert"
)

// ==================== NewUnifiedClient Tests ====================

func TestNewUnifiedClient(t *testing.T) {
	t.Run("Creates Client With HTTP Fallback", func(t *testing.T) {
		// gRPC is not available in test environment, will fall back to HTTP
		client := NewUnifiedClient()
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		// gRPC might or might not be available depending on daemon status
	})
}

func TestNewUnifiedClientWithMode(t *testing.T) {
	t.Run("Creates HTTP Mode Client", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false) // prefer HTTP
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		assert.False(t, client.useGRPC)
	})

	t.Run("Creates GRPC Mode Client Falls Back", func(t *testing.T) {
		// gRPC not available in test, will fail to connect
		client := NewUnifiedClientWithMode(true) // prefer gRPC
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		// useGRPC will be false since daemon is not running
		// This tests the fallback behavior
	})
}

// ==================== UnifiedClient.IsUsingGRPC Tests ====================

func TestUnifiedClient_IsUsingGRPC_Extended(t *testing.T) {
	t.Run("Returns False For HTTP Client", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		assert.False(t, client.IsUsingGRPC())
	})

	t.Run("Is Thread Safe", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				_ = client.IsUsingGRPC()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// ==================== UnifiedClient.SwitchToHTTP Tests ====================

func TestUnifiedClient_SwitchToHTTP(t *testing.T) {
	t.Run("Switches To HTTP Mode", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		// Even if we were using gRPC, this should switch to HTTP
		client.SwitchToHTTP()
		assert.False(t, client.IsUsingGRPC())
		assert.Nil(t, client.grpcClient)
	})

	t.Run("Safe To Call Multiple Times", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		client.SwitchToHTTP()
		client.SwitchToHTTP()
		assert.False(t, client.IsUsingGRPC())
	})
}

// ==================== UnifiedClient.SwitchToGRPC Tests ====================

func TestUnifiedClient_SwitchToGRPC(t *testing.T) {
	t.Run("Returns Error When Daemon Not Running", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.SwitchToGRPC()
		// Should error because daemon is not running
		assert.Error(t, err)
	})
}

// ==================== UnifiedClient.Close Tests ====================

func TestUnifiedClient_Close(t *testing.T) {
	t.Run("Close HTTP Client Does Not Error", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.Close()
		assert.NoError(t, err)
	})

	t.Run("Safe To Close Multiple Times", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.Close()
		assert.NoError(t, err)
		err = client.Close()
		assert.NoError(t, err)
	})
}

// ==================== UnifiedClient.CheckDaemonStatus Tests ====================

func TestUnifiedClient_CheckDaemonStatus(t *testing.T) {
	t.Run("Returns Status Using HTTP When Not Using GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		// Daemon is not running in tests, so should return false
		status := client.CheckDaemonStatus()
		assert.False(t, status)
	})
}

// ==================== UnifiedClient GRPC-Only Methods Tests ====================

func TestUnifiedClient_GRPCOnlyMethods(t *testing.T) {
	t.Run("LeaveNetwork Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.LeaveNetwork("network-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("GetPeers Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		peers, err := client.GetPeers()
		assert.Error(t, err)
		assert.Nil(t, peers)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("SendChatMessage Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.SendChatMessage("network-1", "hello")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("SendFile Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		id, err := client.SendFile("peer-1", "/tmp/test.txt")
		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("GetSettings Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		settings, err := client.GetSettings()
		assert.Error(t, err)
		assert.Nil(t, settings)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("UpdateSettings Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		err := client.UpdateSettings(&pb.Settings{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gRPC connection")
	})

	t.Run("Subscribe Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		stream, err := client.Subscribe(context.Background(), []pb.EventType{})
		assert.Error(t, err)
		assert.Nil(t, stream)
		assert.Contains(t, err.Error(), "gRPC")
	})

	t.Run("GetVersion Errors Without GRPC", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		version, err := client.GetVersion()
		assert.Error(t, err)
		assert.Nil(t, version)
		assert.Contains(t, err.Error(), "gRPC")
	})
}

// ==================== UnifiedClient HTTP Methods Tests ====================

func TestUnifiedClient_HTTPMethods(t *testing.T) {
	t.Run("GetStatus Returns Error When Daemon Not Running", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		status, err := client.GetStatus()
		// Daemon not running, should error
		assert.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("CreateNetwork Returns Error When Daemon Not Running", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		network, err := client.CreateNetwork("test-network")
		assert.Error(t, err)
		assert.Nil(t, network)
	})

	t.Run("JoinNetwork Returns Error When Daemon Not Running", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		network, err := client.JoinNetwork("invite-code")
		assert.Error(t, err)
		assert.Nil(t, network)
	})

	t.Run("GetNetworks Returns Error When Daemon Not Running", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		networks, err := client.GetNetworks()
		assert.Error(t, err)
		assert.Nil(t, networks)
	})
}

// ==================== UnifiedClient Struct Tests ====================

func TestUnifiedClientStruct(t *testing.T) {
	t.Run("Has All Required Fields", func(t *testing.T) {
		client := &UnifiedClient{
			httpClient: NewClient(),
			useGRPC:    false,
		}

		assert.NotNil(t, client.httpClient)
		assert.False(t, client.useGRPC)
		assert.Nil(t, client.grpcClient)
	})
}

// ==================== DaemonClient Interface Tests ====================

func TestDaemonClientInterface(t *testing.T) {
	t.Run("UnifiedClient Implements DaemonClient", func(t *testing.T) {
		var _ DaemonClient = (*UnifiedClient)(nil)
		// If this compiles, UnifiedClient implements DaemonClient
	})
}

// ==================== Concurrent Access Tests ====================

func TestUnifiedClient_ConcurrentAccess(t *testing.T) {
	t.Run("Safe Concurrent Method Calls", func(t *testing.T) {
		client := NewUnifiedClientWithMode(false)
		done := make(chan bool, 30)

		// Concurrent IsUsingGRPC
		for i := 0; i < 10; i++ {
			go func() {
				_ = client.IsUsingGRPC()
				done <- true
			}()
		}

		// Concurrent CheckDaemonStatus
		for i := 0; i < 10; i++ {
			go func() {
				_ = client.CheckDaemonStatus()
				done <- true
			}()
		}

		// Concurrent SwitchToHTTP
		for i := 0; i < 10; i++ {
			go func() {
				client.SwitchToHTTP()
				done <- true
			}()
		}

		for i := 0; i < 30; i++ {
			<-done
		}
	})
}
