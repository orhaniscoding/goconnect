package domain

import (
	"time"
)

// Peer represents a WireGuard peer in a network
// A peer is created when a device joins a network
type Peer struct {
	ID                  string     `json:"id"`
	NetworkID           string     `json:"network_id"`
	DeviceID            string     `json:"device_id"`
	TenantID            string     `json:"tenant_id"`
	PublicKey           string     `json:"public_key"`               // WireGuard public key
	PresharedKey        string     `json:"preshared_key"`            // Optional preshared key for additional security
	Endpoint            string     `json:"endpoint"`                 // Last known endpoint (IP:Port)
	AllowedIPs          []string   `json:"allowed_ips"`              // CIDR blocks this peer can route
	PersistentKeepalive int        `json:"persistent_keepalive"`     // Keepalive interval in seconds (0 = disabled)
	LastHandshake       *time.Time `json:"last_handshake,omitempty"` // Last successful WireGuard handshake
	RxBytes             int64      `json:"rx_bytes"`                 // Received bytes
	TxBytes             int64      `json:"tx_bytes"`                 // Transmitted bytes
	Active              bool       `json:"active"`                   // Whether peer is currently active
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DisabledAt          *time.Time `json:"disabled_at,omitempty"` // Soft disable
}

// PeerStats represents real-time statistics for a peer
type PeerStats struct {
	PeerID        string     `json:"peer_id"`
	Endpoint      string     `json:"endpoint"`
	LastHandshake *time.Time `json:"last_handshake,omitempty"`
	RxBytes       int64      `json:"rx_bytes"`
	TxBytes       int64      `json:"tx_bytes"`
	Active        bool       `json:"active"`
	Latency       int        `json:"latency_ms"` // Round-trip time in milliseconds
}

// CreatePeerRequest is the request to create a new peer
type CreatePeerRequest struct {
	NetworkID           string   `json:"network_id" binding:"required"`
	DeviceID            string   `json:"device_id" binding:"required"`
	PublicKey           string   `json:"public_key" binding:"required,len=44"` // Base64 WireGuard key
	PresharedKey        string   `json:"preshared_key,omitempty"`
	AllowedIPs          []string `json:"allowed_ips" binding:"required,min=1"`
	PersistentKeepalive int      `json:"persistent_keepalive,omitempty"`
}

// UpdatePeerRequest is the request to update peer information
type UpdatePeerRequest struct {
	Endpoint            *string   `json:"endpoint,omitempty"`
	AllowedIPs          *[]string `json:"allowed_ips,omitempty"`
	PresharedKey        *string   `json:"preshared_key,omitempty"`
	PersistentKeepalive *int      `json:"persistent_keepalive,omitempty"`
}

// UpdatePeerStatsRequest is the request to update peer statistics
type UpdatePeerStatsRequest struct {
	Endpoint      string     `json:"endpoint,omitempty"`
	LastHandshake *time.Time `json:"last_handshake,omitempty"`
	RxBytes       int64      `json:"rx_bytes"`
	TxBytes       int64      `json:"tx_bytes"`
}

// Validate validates peer creation request
func (r *CreatePeerRequest) Validate() error {
	if r.NetworkID == "" {
		return NewError(ErrValidation, "network_id is required", map[string]string{"field": "network_id"})
	}
	if r.DeviceID == "" {
		return NewError(ErrValidation, "device_id is required", map[string]string{"field": "device_id"})
	}
	if len(r.PublicKey) != 44 {
		return NewError(ErrValidation, "public_key must be 44 characters (base64 encoded 32 bytes)", map[string]string{"field": "public_key"})
	}
	if len(r.AllowedIPs) == 0 {
		return NewError(ErrValidation, "at least one allowed IP is required", map[string]string{"field": "allowed_ips"})
	}
	if r.PersistentKeepalive < 0 || r.PersistentKeepalive > 65535 {
		return NewError(ErrValidation, "persistent_keepalive must be between 0 and 65535", map[string]string{"field": "persistent_keepalive"})
	}
	return nil
}

// IsHandshakeRecent checks if the last handshake was within the specified duration
func (p *Peer) IsHandshakeRecent(within time.Duration) bool {
	if p.LastHandshake == nil {
		return false
	}
	return time.Since(*p.LastHandshake) < within
}

// IsActive checks if peer is considered active based on handshake time
func (p *Peer) IsActive(timeout time.Duration) bool {
	return p.Active && p.IsHandshakeRecent(timeout)
}

// GetTrafficTotal returns total traffic (rx + tx) in bytes
func (p *Peer) GetTrafficTotal() int64 {
	return p.RxBytes + p.TxBytes
}
