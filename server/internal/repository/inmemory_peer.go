package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

type InMemoryPeerRepository struct {
	mu    sync.RWMutex
	peers map[string]*domain.Peer
}

func NewInMemoryPeerRepository() *InMemoryPeerRepository {
	return &InMemoryPeerRepository{
		peers: make(map[string]*domain.Peer),
	}
}

func (r *InMemoryPeerRepository) Create(ctx context.Context, peer *domain.Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if peer.ID == "" {
		peer.ID = uuid.New().String()
	}
	peer.CreatedAt = time.Now()
	peer.UpdatedAt = time.Now()

	// Check for duplicate device in network (unique constraint: network + device)
	for _, p := range r.peers {
		if p.NetworkID == peer.NetworkID && p.DeviceID == peer.DeviceID && p.DisabledAt == nil {
			return domain.NewError(domain.ErrConflict, "device already has a peer in this network", nil)
		}
	}

	r.peers[peer.ID] = peer
	return nil
}

func (r *InMemoryPeerRepository) GetByID(ctx context.Context, peerID string) (*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, ok := r.peers[peerID]
	if !ok || peer.DisabledAt != nil {
		return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	return peer, nil
}

func (r *InMemoryPeerRepository) GetByNetworkID(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var peers []*domain.Peer
	for _, peer := range r.peers {
		if peer.NetworkID == networkID && peer.DisabledAt == nil {
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

func (r *InMemoryPeerRepository) GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var peers []*domain.Peer
	for _, peer := range r.peers {
		if peer.DeviceID == deviceID && peer.DisabledAt == nil {
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

func (r *InMemoryPeerRepository) GetByNetworkAndDevice(ctx context.Context, networkID, deviceID string) (*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, peer := range r.peers {
		if peer.NetworkID == networkID && peer.DeviceID == deviceID && peer.DisabledAt == nil {
			return peer, nil
		}
	}

	return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
}

func (r *InMemoryPeerRepository) GetByPublicKey(ctx context.Context, publicKey string) (*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, peer := range r.peers {
		if peer.PublicKey == publicKey && peer.DisabledAt == nil {
			return peer, nil
		}
	}

	return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
}

func (r *InMemoryPeerRepository) GetActivePeers(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var peers []*domain.Peer
	for _, peer := range r.peers {
		if peer.NetworkID == networkID && peer.Active && peer.DisabledAt == nil {
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

func (r *InMemoryPeerRepository) Update(ctx context.Context, peer *domain.Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.peers[peer.ID]
	if !ok || existing.DisabledAt != nil {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	peer.UpdatedAt = time.Now()
	peer.CreatedAt = existing.CreatedAt // Preserve creation time
	r.peers[peer.ID] = peer

	return nil
}

func (r *InMemoryPeerRepository) UpdateStats(ctx context.Context, peerID string, stats *domain.UpdatePeerStatsRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, ok := r.peers[peerID]
	if !ok || peer.DisabledAt != nil {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	if stats.Endpoint != "" {
		peer.Endpoint = stats.Endpoint
	}
	if stats.LastHandshake != nil {
		peer.LastHandshake = stats.LastHandshake
	}
	peer.RxBytes = stats.RxBytes
	peer.TxBytes = stats.TxBytes
	peer.UpdatedAt = time.Now()

	// Update active status based on handshake
	if stats.LastHandshake != nil {
		peer.Active = time.Since(*stats.LastHandshake) < 3*time.Minute
	}

	return nil
}

func (r *InMemoryPeerRepository) Delete(ctx context.Context, peerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, ok := r.peers[peerID]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	now := time.Now()
	peer.DisabledAt = &now
	peer.Active = false
	peer.UpdatedAt = now

	return nil
}

func (r *InMemoryPeerRepository) HardDelete(ctx context.Context, peerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.peers[peerID]; !ok {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	delete(r.peers, peerID)
	return nil
}

func (r *InMemoryPeerRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var peers []*domain.Peer
	for _, peer := range r.peers {
		if peer.TenantID == tenantID && peer.DisabledAt == nil {
			peers = append(peers, peer)
		}
	}

	// Apply pagination
	if offset >= len(peers) {
		return []*domain.Peer{}, nil
	}

	end := offset + limit
	if end > len(peers) {
		end = len(peers)
	}

	return peers[offset:end], nil
}
