package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PeerRepository defines the interface for peer persistence
type PeerRepository interface {
	// Create creates a new peer
	Create(ctx context.Context, peer *domain.Peer) error

	// GetByID retrieves a peer by ID
	GetByID(ctx context.Context, peerID string) (*domain.Peer, error)

	// GetByNetworkID retrieves all peers in a network
	GetByNetworkID(ctx context.Context, networkID string) ([]*domain.Peer, error)

	// GetByDeviceID retrieves all peers for a device
	GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.Peer, error)

	// GetByNetworkAndDevice retrieves a peer by network and device
	GetByNetworkAndDevice(ctx context.Context, networkID, deviceID string) (*domain.Peer, error)

	// GetByPublicKey retrieves a peer by public key
	GetByPublicKey(ctx context.Context, publicKey string) (*domain.Peer, error)

	// GetActivePeers retrieves all active peers in a network
	GetActivePeers(ctx context.Context, networkID string) ([]*domain.Peer, error)

	// Update updates a peer
	Update(ctx context.Context, peer *domain.Peer) error

	// UpdateStats updates peer statistics
	UpdateStats(ctx context.Context, peerID string, stats *domain.UpdatePeerStatsRequest) error

	// Delete soft-deletes a peer
	Delete(ctx context.Context, peerID string) error

	// HardDelete permanently deletes a peer
	HardDelete(ctx context.Context, peerID string) error

	// ListByTenant retrieves all peers for a tenant
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Peer, error)
}
