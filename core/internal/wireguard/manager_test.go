package wireguard

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// mockWGClient is a mock implementation of WGClient for testing
type mockWGClient struct {
	closeErr         error
	closeCalled      bool
	deviceResult     *wgtypes.Device
	deviceErr        error
	deviceCalled     bool
	deviceName       string
	configureErr     error
	configureCalled  bool
	configureDevName string
	configureConfig  wgtypes.Config
}

func (m *mockWGClient) Close() error {
	m.closeCalled = true
	return m.closeErr
}

func (m *mockWGClient) Device(name string) (*wgtypes.Device, error) {
	m.deviceCalled = true
	m.deviceName = name
	return m.deviceResult, m.deviceErr
}

func (m *mockWGClient) ConfigureDevice(name string, cfg wgtypes.Config) error {
	m.configureCalled = true
	m.configureDevName = name
	m.configureConfig = cfg
	return m.configureErr
}

// Helper function to generate a valid WireGuard key pair for testing
func generateTestKeyPair(t *testing.T) (wgtypes.Key, wgtypes.Key) {
	t.Helper()
	privKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(t, err)
	return privKey, privKey.PublicKey()
}

func TestNewManager(t *testing.T) {
	// Note: NewManager creates a real wgctrl.Client which requires
	// WireGuard kernel module. These tests verify the function signature
	// and error handling patterns. Integration tests would test actual creation.

	t.Run("returns error when wgctrl client creation fails", func(t *testing.T) {
		// This test documents expected behavior - actual failure would occur
		// in environments without WireGuard support
		// The function is expected to wrap errors with context
		manager, err := NewManager("wg0", "invalid-key", 51820)

		// In most test environments without WireGuard, this will fail
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create wgctrl client")
			assert.Nil(t, manager)
		} else {
			// If it succeeds (WireGuard is available), clean up
			assert.NotNil(t, manager)
			_ = manager.Close()
		}
	})
}

func TestNewManagerWithClient(t *testing.T) {
	// Test the manager creation using dependency injection pattern
	t.Run("creates manager with provided parameters", func(t *testing.T) {
		mockClient := &mockWGClient{}
		privKey, _ := generateTestKeyPair(t)

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		assert.Equal(t, "wg0", manager.interfaceName)
		assert.Equal(t, privKey.String(), manager.privateKey)
		assert.Equal(t, 51820, manager.port)
		assert.Same(t, mockClient, manager.client)
	})
}

func TestManager_Close(t *testing.T) {
	t.Run("closes client successfully", func(t *testing.T) {
		mockClient := &mockWGClient{}
		manager := &Manager{client: mockClient}

		err := manager.Close()

		assert.NoError(t, err)
		assert.True(t, mockClient.closeCalled)
	})

	t.Run("returns error when client close fails", func(t *testing.T) {
		mockClient := &mockWGClient{
			closeErr: errors.New("close failed"),
		}
		manager := &Manager{client: mockClient}

		err := manager.Close()

		assert.Error(t, err)
		assert.Equal(t, "close failed", err.Error())
		assert.True(t, mockClient.closeCalled)
	})
}

func TestManager_SyncPeers(t *testing.T) {
	t.Run("syncs peers successfully", func(t *testing.T) {
		mockClient := &mockWGClient{}
		privKey, _ := generateTestKeyPair(t)
		peerPubKey, _ := generateTestKeyPair(t)

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		_, ipNet, _ := net.ParseCIDR("10.0.0.2/32")
		peers := []wgtypes.PeerConfig{
			{
				PublicKey:  peerPubKey,
				AllowedIPs: []net.IPNet{*ipNet},
			},
		}

		err := manager.SyncPeers(peers)

		require.NoError(t, err)
		assert.True(t, mockClient.configureCalled)
		assert.Equal(t, "wg0", mockClient.configureDevName)
		assert.NotNil(t, mockClient.configureConfig.PrivateKey)
		assert.Equal(t, privKey.String(), mockClient.configureConfig.PrivateKey.String())
		assert.NotNil(t, mockClient.configureConfig.ListenPort)
		assert.Equal(t, 51820, *mockClient.configureConfig.ListenPort)
		assert.True(t, mockClient.configureConfig.ReplacePeers)
		assert.Len(t, mockClient.configureConfig.Peers, 1)
	})

	t.Run("syncs empty peers list", func(t *testing.T) {
		mockClient := &mockWGClient{}
		privKey, _ := generateTestKeyPair(t)

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		err := manager.SyncPeers([]wgtypes.PeerConfig{})

		require.NoError(t, err)
		assert.True(t, mockClient.configureCalled)
		assert.Len(t, mockClient.configureConfig.Peers, 0)
	})

	t.Run("returns error for invalid private key", func(t *testing.T) {
		mockClient := &mockWGClient{}

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    "invalid-key",
			port:          51820,
			client:        mockClient,
		}

		err := manager.SyncPeers([]wgtypes.PeerConfig{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid server private key")
		assert.False(t, mockClient.configureCalled)
	})

	t.Run("returns error when configure device fails", func(t *testing.T) {
		mockClient := &mockWGClient{
			configureErr: errors.New("device not found"),
		}
		privKey, _ := generateTestKeyPair(t)

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		err := manager.SyncPeers([]wgtypes.PeerConfig{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to configure device wg0")
		assert.Contains(t, err.Error(), "device not found")
	})

	t.Run("syncs multiple peers", func(t *testing.T) {
		mockClient := &mockWGClient{}
		privKey, _ := generateTestKeyPair(t)
		peerPubKey1, _ := generateTestKeyPair(t)
		peerPubKey2, _ := generateTestKeyPair(t)

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		_, ipNet1, _ := net.ParseCIDR("10.0.0.2/32")
		_, ipNet2, _ := net.ParseCIDR("10.0.0.3/32")
		peers := []wgtypes.PeerConfig{
			{
				PublicKey:  peerPubKey1,
				AllowedIPs: []net.IPNet{*ipNet1},
			},
			{
				PublicKey:  peerPubKey2,
				AllowedIPs: []net.IPNet{*ipNet2},
			},
		}

		err := manager.SyncPeers(peers)

		require.NoError(t, err)
		assert.Len(t, mockClient.configureConfig.Peers, 2)
	})
}

func TestManager_UpdateMetrics(t *testing.T) {
	t.Run("updates metrics for device with no peers", func(t *testing.T) {
		privKey, pubKey := generateTestKeyPair(t)
		mockClient := &mockWGClient{
			deviceResult: &wgtypes.Device{
				Name:       "wg0",
				Type:       wgtypes.LinuxKernel,
				PrivateKey: privKey,
				PublicKey:  pubKey,
				ListenPort: 51820,
				Peers:      []wgtypes.Peer{},
			},
		}

		manager := &Manager{
			interfaceName: "wg0",
			client:        mockClient,
		}

		err := manager.UpdateMetrics()

		require.NoError(t, err)
		assert.True(t, mockClient.deviceCalled)
		assert.Equal(t, "wg0", mockClient.deviceName)
	})

	t.Run("updates metrics for device with peers", func(t *testing.T) {
		privKey, pubKey := generateTestKeyPair(t)
		peerPubKey1, _ := generateTestKeyPair(t)
		peerPubKey2, _ := generateTestKeyPair(t)

		_, ipNet1, _ := net.ParseCIDR("10.0.0.2/32")
		_, ipNet2, _ := net.ParseCIDR("10.0.0.3/32")

		mockClient := &mockWGClient{
			deviceResult: &wgtypes.Device{
				Name:       "wg0",
				Type:       wgtypes.LinuxKernel,
				PrivateKey: privKey,
				PublicKey:  pubKey,
				ListenPort: 51820,
				Peers: []wgtypes.Peer{
					{
						PublicKey:         peerPubKey1,
						AllowedIPs:        []net.IPNet{*ipNet1},
						ReceiveBytes:      1024,
						TransmitBytes:     2048,
						LastHandshakeTime: time.Now().Add(-30 * time.Second),
					},
					{
						PublicKey:         peerPubKey2,
						AllowedIPs:        []net.IPNet{*ipNet2},
						ReceiveBytes:      4096,
						TransmitBytes:     8192,
						LastHandshakeTime: time.Now().Add(-60 * time.Second),
					},
				},
			},
		}

		manager := &Manager{
			interfaceName: "wg0",
			client:        mockClient,
		}

		err := manager.UpdateMetrics()

		require.NoError(t, err)
		assert.True(t, mockClient.deviceCalled)
	})

	t.Run("handles peer with zero handshake time", func(t *testing.T) {
		privKey, pubKey := generateTestKeyPair(t)
		peerPubKey, _ := generateTestKeyPair(t)

		_, ipNet, _ := net.ParseCIDR("10.0.0.2/32")

		mockClient := &mockWGClient{
			deviceResult: &wgtypes.Device{
				Name:       "wg0",
				Type:       wgtypes.LinuxKernel,
				PrivateKey: privKey,
				PublicKey:  pubKey,
				ListenPort: 51820,
				Peers: []wgtypes.Peer{
					{
						PublicKey:         peerPubKey,
						AllowedIPs:        []net.IPNet{*ipNet},
						ReceiveBytes:      0,
						TransmitBytes:     0,
						LastHandshakeTime: time.Time{}, // Zero value
					},
				},
			},
		}

		manager := &Manager{
			interfaceName: "wg0",
			client:        mockClient,
		}

		err := manager.UpdateMetrics()

		// Should not error even with zero handshake time
		require.NoError(t, err)
		assert.True(t, mockClient.deviceCalled)
	})

	t.Run("returns error when device lookup fails", func(t *testing.T) {
		mockClient := &mockWGClient{
			deviceErr: errors.New("device not found"),
		}

		manager := &Manager{
			interfaceName: "wg0",
			client:        mockClient,
		}

		err := manager.UpdateMetrics()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get device wg0")
		assert.Contains(t, err.Error(), "device not found")
	})

	t.Run("uses correct interface name", func(t *testing.T) {
		privKey, pubKey := generateTestKeyPair(t)
		mockClient := &mockWGClient{
			deviceResult: &wgtypes.Device{
				Name:       "wg1",
				Type:       wgtypes.LinuxKernel,
				PrivateKey: privKey,
				PublicKey:  pubKey,
				ListenPort: 51821,
				Peers:      []wgtypes.Peer{},
			},
		}

		manager := &Manager{
			interfaceName: "wg1",
			client:        mockClient,
		}

		err := manager.UpdateMetrics()

		require.NoError(t, err)
		assert.Equal(t, "wg1", mockClient.deviceName)
	})
}

func TestManager_Integration(t *testing.T) {
	t.Run("full lifecycle with mocked client", func(t *testing.T) {
		privKey, pubKey := generateTestKeyPair(t)
		peerPubKey, _ := generateTestKeyPair(t)

		_, ipNet, _ := net.ParseCIDR("10.0.0.2/32")

		mockClient := &mockWGClient{
			deviceResult: &wgtypes.Device{
				Name:       "wg0",
				Type:       wgtypes.LinuxKernel,
				PrivateKey: privKey,
				PublicKey:  pubKey,
				ListenPort: 51820,
				Peers: []wgtypes.Peer{
					{
						PublicKey:         peerPubKey,
						AllowedIPs:        []net.IPNet{*ipNet},
						ReceiveBytes:      1024,
						TransmitBytes:     2048,
						LastHandshakeTime: time.Now(),
					},
				},
			},
		}

		manager := &Manager{
			interfaceName: "wg0",
			privateKey:    privKey.String(),
			port:          51820,
			client:        mockClient,
		}

		// Sync peers
		peers := []wgtypes.PeerConfig{
			{
				PublicKey:  peerPubKey,
				AllowedIPs: []net.IPNet{*ipNet},
			},
		}
		err := manager.SyncPeers(peers)
		require.NoError(t, err)

		// Update metrics
		err = manager.UpdateMetrics()
		require.NoError(t, err)

		// Close
		err = manager.Close()
		require.NoError(t, err)

		// Verify all operations were called
		assert.True(t, mockClient.configureCalled)
		assert.True(t, mockClient.deviceCalled)
		assert.True(t, mockClient.closeCalled)
	})
}
