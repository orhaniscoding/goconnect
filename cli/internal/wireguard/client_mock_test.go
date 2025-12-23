package wireguard

import (
	"fmt"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// MockWireGuardController is a mock for the WireGuardController interface
type MockWireGuardController struct {
	mock.Mock
}

func (m *MockWireGuardController) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWireGuardController) ConfigureDevice(device string, cfg wgtypes.Config) error {
	args := m.Called(device, cfg)
	return args.Error(0)
}

func (m *MockWireGuardController) Device(device string) (*wgtypes.Device, error) {
	args := m.Called(device)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wgtypes.Device), args.Error(1)
}

func TestApplyConfig(t *testing.T) {
	mockWg := new(MockWireGuardController)
	client := &Client{
		interfaceName: "wg0",
		wgClient:      mockWg,
	}

	privKey, _ := wgtypes.GeneratePrivateKey()
	pubKey, _ := wgtypes.GenerateKey()

	deviceConfig := &api.DeviceConfig{
		Interface: api.InterfaceConfig{
			ListenPort: 51820,
		},
		Peers: []api.PeerConfig{
			{
				PublicKey:           pubKey.String(),
				Endpoint:            "10.0.0.1:51820",
				AllowedIPs:          []string{"10.0.0.2/32"},
				PersistentKeepalive: 25,
			},
		},
	}

	// Verify that ConfigureDevice is called with correct parameters
	mockWg.On("ConfigureDevice", "wg0", mock.MatchedBy(func(cfg wgtypes.Config) bool {
		if cfg.PrivateKey == nil || cfg.PrivateKey.String() != privKey.String() {
			return false
		}
		if cfg.ListenPort == nil || *cfg.ListenPort != 51820 {
			return false
		}
		if len(cfg.Peers) != 1 {
			return false
		}
		peer := cfg.Peers[0]
		if peer.PublicKey.String() != pubKey.String() {
			return false
		}
		if peer.Endpoint.String() != "10.0.0.1:51820" {
			return false
		}
		return true
	})).Return(nil)

	err := client.ApplyConfig(deviceConfig, privKey.String())
	assert.NoError(t, err)
	mockWg.AssertExpectations(t)
}

func TestUpdatePeerEndpoint(t *testing.T) {
	mockWg := new(MockWireGuardController)
	client := &Client{
		interfaceName: "wg0",
		wgClient:      mockWg,
	}

	pubKey, _ := wgtypes.GenerateKey()
	endpoint := "192.168.1.1:51820"

	mockWg.On("ConfigureDevice", "wg0", mock.MatchedBy(func(cfg wgtypes.Config) bool {
		if len(cfg.Peers) != 1 {
			return false
		}
		peer := cfg.Peers[0]
		if peer.PublicKey.String() != pubKey.String() {
			return false
		}
		if peer.Endpoint.String() != endpoint {
			return false
		}
		return peer.UpdateOnly // Should be update only
	})).Return(nil)

	err := client.UpdatePeerEndpoint(pubKey.String(), endpoint)
	assert.NoError(t, err)
	mockWg.AssertExpectations(t)
}

func TestDown(t *testing.T) {
	mockWg := new(MockWireGuardController)
	client := &Client{
		interfaceName: "wg0",
		wgClient:      mockWg,
	}

	mockWg.On("ConfigureDevice", "wg0", mock.MatchedBy(func(cfg wgtypes.Config) bool {
		return len(cfg.Peers) == 0 && cfg.ReplacePeers == true
	})).Return(nil)

	err := client.Down()
	assert.NoError(t, err)
	mockWg.AssertExpectations(t)
}

func TestGetStatus(t *testing.T) {
	mockWg := new(MockWireGuardController)
	client := &Client{
		interfaceName: "wg0",
		wgClient:      mockWg,
	}

	pubKey, _ := wgtypes.GenerateKey()
	peerKey, _ := wgtypes.GenerateKey()

	mockDevice := &wgtypes.Device{
		Name:       "wg0",
		PublicKey:  pubKey,
		ListenPort: 51820,
		Peers: []wgtypes.Peer{
			{
				PublicKey:         peerKey,
				ReceiveBytes:      1000,
				TransmitBytes:     2000,
				LastHandshakeTime: time.Now().Add(-1 * time.Minute),
			},
		},
	}

	mockWg.On("Device", "wg0").Return(mockDevice, nil)

	status, err := client.GetStatus()
	assert.NoError(t, err)
	assert.True(t, status.Active)
	assert.Equal(t, pubKey.String(), status.PublicKey)
	assert.Equal(t, 51820, status.ListenPort)
	assert.Equal(t, 1, status.Peers)
	assert.Equal(t, int64(1000), status.TotalRx)
	assert.Equal(t, int64(2000), status.TotalTx)

	mockWg.AssertExpectations(t)
}

func TestGetStatus_InterfaceNotFound(t *testing.T) {
	mockWg := new(MockWireGuardController)
	client := &Client{
		interfaceName: "wg0",
		wgClient:      mockWg,
	}

	mockWg.On("Device", "wg0").Return(nil, fmt.Errorf("no such device"))

	status, err := client.GetStatus()
	assert.NoError(t, err) // Should not return error, just Active=false
	assert.False(t, status.Active)

	mockWg.AssertExpectations(t)
}
