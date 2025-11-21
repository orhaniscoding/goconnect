package wireguard

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Manager handles WireGuard interface configuration
type Manager struct {
	interfaceName string
	privateKey    string
	port          int
	client        *wgctrl.Client
}

// NewManager creates a new WireGuard manager
func NewManager(iface string, privKey string, port int) (*Manager, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	return &Manager{
		interfaceName: iface,
		privateKey:    privKey,
		port:          port,
		client:        client,
	}, nil
}

// Close closes the wgctrl client
func (m *Manager) Close() error {
	return m.client.Close()
}

// SyncPeers applies the list of peers to the WireGuard interface
func (m *Manager) SyncPeers(peers []wgtypes.PeerConfig) error {
	// Parse private key
	key, err := wgtypes.ParseKey(m.privateKey)
	if err != nil {
		return fmt.Errorf("invalid server private key: %w", err)
	}

	cfg := wgtypes.Config{
		PrivateKey:   &key,
		ListenPort:   &m.port,
		ReplacePeers: true, // Replace existing peers with the new list
		Peers:        peers,
	}

	if err := m.client.ConfigureDevice(m.interfaceName, cfg); err != nil {
		return fmt.Errorf("failed to configure device %s: %w", m.interfaceName, err)
	}

	return nil
}
