package wireguard

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Manager handles WireGuard interface configuration
type Manager struct {
	interfaceName string
	client        *wgctrl.Client
}

// NewManager creates a new WireGuard manager
func NewManager(interfaceName string) (*Manager, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	return &Manager{
		interfaceName: interfaceName,
		client:        client,
	}, nil
}

// Close closes the wgctrl client
func (m *Manager) Close() error {
	return m.client.Close()
}

// ApplyConfig applies the configuration to the WireGuard interface
func (m *Manager) ApplyConfig(config *api.DeviceConfig, privateKey string) error {
	// Parse private key
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// Prepare peers
	peers := make([]wgtypes.PeerConfig, 0, len(config.Peers))
	for _, p := range config.Peers {
		pubKey, err := wgtypes.ParseKey(p.PublicKey)
		if err != nil {
			log.Printf("Skipping peer with invalid public key %s: %v", p.PublicKey, err)
			continue
		}

		allowedIPs := make([]net.IPNet, 0, len(p.AllowedIPs))
		for _, ipStr := range p.AllowedIPs {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				log.Printf("Skipping invalid allowed IP %s for peer %s: %v", ipStr, p.PublicKey, err)
				continue
			}
			allowedIPs = append(allowedIPs, *ipNet)
		}

		var endpoint *net.UDPAddr
		if p.Endpoint != "" {
			addr, err := net.ResolveUDPAddr("udp", p.Endpoint)
			if err != nil {
				log.Printf("Skipping invalid endpoint %s for peer %s: %v", p.Endpoint, p.PublicKey, err)
			} else {
				endpoint = addr
			}
		}

		var psk *wgtypes.Key
		if p.PresharedKey != "" {
			k, err := wgtypes.ParseKey(p.PresharedKey)
			if err == nil {
				psk = &k
			}
		}

		ka := time.Duration(p.PersistentKeepalive) * time.Second

		peers = append(peers, wgtypes.PeerConfig{
			PublicKey:                   pubKey,
			Remove:                      false,
			UpdateOnly:                  false, // Overwrite
			PresharedKey:                psk,
			Endpoint:                    endpoint,
			PersistentKeepaliveInterval: &ka,
			AllowedIPs:                  allowedIPs,
			ReplaceAllowedIPs:           true,
		})
	}

	// Configure device
	cfg := wgtypes.Config{
		PrivateKey:   &key,
		ListenPort:   &config.Interface.ListenPort,
		Peers:        peers,
		ReplacePeers: true,
	}

	return m.client.ConfigureDevice(m.interfaceName, cfg)
}

// Status represents the WireGuard interface status
type Status struct {
	Active      bool      `json:"active"`
	PublicKey   string    `json:"public_key"`
	ListenPort  int       `json:"listen_port"`
	Peers       int       `json:"peers"`
	TotalRx     int64     `json:"total_rx"`
	TotalTx     int64     `json:"total_tx"`
	LastHandshake time.Time `json:"last_handshake"`
}

// GetStatus retrieves the current interface status
func (m *Manager) GetStatus() (*Status, error) {
	dev, err := m.client.Device(m.interfaceName)
	if err != nil {
		return &Status{Active: false}, nil // Interface likely doesn't exist
	}

	var totalRx, totalTx int64
	var lastHandshake time.Time

	for _, peer := range dev.Peers {
		totalRx += peer.ReceiveBytes
		totalTx += peer.TransmitBytes
		if peer.LastHandshakeTime.After(lastHandshake) {
			lastHandshake = peer.LastHandshakeTime
		}
	}

	return &Status{
		Active:        true,
		PublicKey:     dev.PublicKey.String(),
		ListenPort:    dev.ListenPort,
		Peers:         len(dev.Peers),
		TotalRx:       totalRx,
		TotalTx:       totalTx,
		LastHandshake: lastHandshake,
	}, nil
}
