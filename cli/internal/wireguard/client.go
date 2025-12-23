package wireguard

import (
	"fmt"
	"net"
	"time"

	"github.com/orhaniscoding/goconnect/cli/internal/api"
	"github.com/orhaniscoding/goconnect/cli/internal/logger"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WireGuardController defines the interface for interacting with WireGuard
type WireGuardController interface {
	Close() error
	ConfigureDevice(device string, cfg wgtypes.Config) error
	Device(device string) (*wgtypes.Device, error)
}

// Client handles WireGuard interface configuration
type Client struct {
	interfaceName string
	wgClient      WireGuardController
}

// NewClient creates a new WireGuard client
func NewClient(interfaceName string) (*Client, error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	return &Client{
		interfaceName: interfaceName,
		wgClient:      wgClient,
	}, nil
}

// Close closes the wgctrl client
func (c *Client) Close() error {
	return c.wgClient.Close()
}

// ApplyConfig applies the configuration to the WireGuard interface
func (c *Client) ApplyConfig(config *api.DeviceConfig, privateKey string) error {
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
			logger.Warn("Skipping peer with invalid public key", "key", p.PublicKey, "error", err)
			continue
		}

		allowedIPs := make([]net.IPNet, 0, len(p.AllowedIPs))
		for _, ipStr := range p.AllowedIPs {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				logger.Warn("Skipping invalid allowed IP", "ip", ipStr, "peer", p.PublicKey, "error", err)
				continue
			}
			allowedIPs = append(allowedIPs, *ipNet)
		}

		var endpoint *net.UDPAddr
		if p.Endpoint != "" {
			addr, err := net.ResolveUDPAddr("udp", p.Endpoint)
			if err != nil {
				logger.Warn("Skipping invalid endpoint", "endpoint", p.Endpoint, "peer", p.PublicKey, "error", err)
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

	return c.wgClient.ConfigureDevice(c.interfaceName, cfg)
}

// UpdatePeerEndpoint updates the endpoint for a specific peer
func (c *Client) UpdatePeerEndpoint(publicKey string, endpoint string) error {
	pubKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	addr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey:         pubKey,
		UpdateOnly:        true,
		Endpoint:          addr,
		ReplaceAllowedIPs: false,
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	return c.wgClient.ConfigureDevice(c.interfaceName, cfg)
}

// Down brings down the WireGuard interface by removing all peers and resetting interface config
func (c *Client) Down() error {
	cfg := wgtypes.Config{
		Peers:        []wgtypes.PeerConfig{}, // Remove all peers
		ReplacePeers: true,
		// Do not set PrivateKey or ListenPort as we only want to remove peers, not reconfigure the interface completely
	}
	return c.wgClient.ConfigureDevice(c.interfaceName, cfg)
}

// Status represents the WireGuard interface status
type Status struct {
	Active        bool      `json:"active"`
	PublicKey     string    `json:"public_key"`
	ListenPort    int       `json:"listen_port"`
	Peers         int       `json:"peers"`
	TotalRx       int64     `json:"total_rx"`
	TotalTx       int64     `json:"total_tx"`
	LastHandshake time.Time `json:"last_handshake"`
}

// GetStatus retrieves the current interface status
func (c *Client) GetStatus() (*Status, error) {
	dev, err := c.wgClient.Device(c.interfaceName)
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
