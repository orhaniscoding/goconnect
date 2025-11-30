package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
)

// PeerService handles peer management business logic
type PeerService struct {
	peerRepo    repository.PeerRepository
	deviceRepo  repository.DeviceRepository
	networkRepo repository.NetworkRepository
}

// NewPeerService creates a new peer service
func NewPeerService(
	peerRepo repository.PeerRepository,
	deviceRepo repository.DeviceRepository,
	networkRepo repository.NetworkRepository,
) *PeerService {
	return &PeerService{
		peerRepo:    peerRepo,
		deviceRepo:  deviceRepo,
		networkRepo: networkRepo,
	}
}

// CreatePeer creates a new peer for a device in a network
func (s *PeerService) CreatePeer(ctx context.Context, req *domain.CreatePeerRequest) (*domain.Peer, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate public key format
	if err := wireguard.ValidatePublicKey(req.PublicKey); err != nil {
		return nil, domain.NewError(domain.ErrValidation, "invalid public key", map[string]string{
			"field": "public_key",
			"error": err.Error(),
		})
	}

	// Verify network exists
	network, err := s.networkRepo.GetByID(ctx, req.NetworkID)
	if err != nil {
		return nil, err
	}

	// Verify device exists
	device, err := s.deviceRepo.GetByID(ctx, req.DeviceID)
	if err != nil {
		return nil, err
	}

	// Check tenant consistency
	if network.TenantID != device.TenantID {
		return nil, domain.NewError(domain.ErrForbidden, "device and network belong to different tenants", nil)
	}

	// Validate preshared key if provided
	if req.PresharedKey != "" {
		// Preshared key should also be 44 characters (32 bytes base64)
		if len(req.PresharedKey) != 44 {
			return nil, domain.NewError(domain.ErrValidation, "invalid preshared key length", map[string]string{
				"field": "preshared_key",
			})
		}
	}

	// Create peer
	peer := &domain.Peer{
		NetworkID:           req.NetworkID,
		DeviceID:            req.DeviceID,
		TenantID:            network.TenantID,
		PublicKey:           req.PublicKey,
		PresharedKey:        req.PresharedKey,
		AllowedIPs:          req.AllowedIPs,
		PersistentKeepalive: req.PersistentKeepalive,
		Active:              false, // Will become active after first handshake
		RxBytes:             0,
		TxBytes:             0,
	}

	if err := s.peerRepo.Create(ctx, peer); err != nil {
		return nil, err
	}

	return peer, nil
}

// GetPeer retrieves a peer by ID
func (s *PeerService) GetPeer(ctx context.Context, peerID string) (*domain.Peer, error) {
	return s.peerRepo.GetByID(ctx, peerID)
}

// GetPeersByNetwork retrieves all peers in a network
func (s *PeerService) GetPeersByNetwork(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	// Verify network exists
	if _, err := s.networkRepo.GetByID(ctx, networkID); err != nil {
		return nil, err
	}

	return s.peerRepo.GetByNetworkID(ctx, networkID)
}

// GetPeersByDevice retrieves all peers for a device
func (s *PeerService) GetPeersByDevice(ctx context.Context, deviceID string) ([]*domain.Peer, error) {
	// Verify device exists
	if _, err := s.deviceRepo.GetByID(ctx, deviceID); err != nil {
		return nil, err
	}

	return s.peerRepo.GetByDeviceID(ctx, deviceID)
}

// GetPeerByNetworkAndDevice retrieves a peer for a specific device in a network
func (s *PeerService) GetPeerByNetworkAndDevice(ctx context.Context, networkID, deviceID string) (*domain.Peer, error) {
	// Verify network exists
	if _, err := s.networkRepo.GetByID(ctx, networkID); err != nil {
		return nil, err
	}

	// Verify device exists
	if _, err := s.deviceRepo.GetByID(ctx, deviceID); err != nil {
		return nil, err
	}

	return s.peerRepo.GetByNetworkAndDevice(ctx, networkID, deviceID)
}

// GetActivePeers retrieves all active peers in a network
func (s *PeerService) GetActivePeers(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	// Verify network exists
	if _, err := s.networkRepo.GetByID(ctx, networkID); err != nil {
		return nil, err
	}

	return s.peerRepo.GetActivePeers(ctx, networkID)
}

// GetActivePeersConfig retrieves active peers with device metadata for configuration
func (s *PeerService) GetActivePeersConfig(ctx context.Context, networkID string) ([]domain.PeerConfig, error) {
	peers, err := s.GetActivePeers(ctx, networkID)
	if err != nil {
		return nil, err
	}

	configs := make([]domain.PeerConfig, 0, len(peers))
	for _, peer := range peers {
		device, err := s.deviceRepo.GetByID(ctx, peer.DeviceID)
		if err != nil {
			// Log error but continue
			continue
		}

		configs = append(configs, domain.PeerConfig{
			PublicKey:           peer.PublicKey,
			Endpoint:            peer.Endpoint,
			AllowedIPs:          peer.AllowedIPs,
			PresharedKey:        peer.PresharedKey,
			PersistentKeepalive: peer.PersistentKeepalive,
			Name:                device.Name,
			Hostname:            device.HostName,
		})
	}
	return configs, nil
}

// UpdatePeer updates a peer
func (s *PeerService) UpdatePeer(ctx context.Context, peerID string, req *domain.UpdatePeerRequest) (*domain.Peer, error) {
	// Get existing peer
	peer, err := s.peerRepo.GetByID(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Endpoint != nil {
		peer.Endpoint = *req.Endpoint
	}
	if req.AllowedIPs != nil {
		peer.AllowedIPs = *req.AllowedIPs
	}
	if req.PresharedKey != nil {
		if *req.PresharedKey != "" && len(*req.PresharedKey) != 44 {
			return nil, domain.NewError(domain.ErrValidation, "invalid preshared key length", map[string]string{
				"field": "preshared_key",
			})
		}
		peer.PresharedKey = *req.PresharedKey
	}
	if req.PersistentKeepalive != nil {
		if *req.PersistentKeepalive < 0 || *req.PersistentKeepalive > 65535 {
			return nil, domain.NewError(domain.ErrValidation, "persistent_keepalive must be between 0 and 65535", map[string]string{
				"field": "persistent_keepalive",
			})
		}
		peer.PersistentKeepalive = *req.PersistentKeepalive
	}

	if err := s.peerRepo.Update(ctx, peer); err != nil {
		return nil, err
	}

	return peer, nil
}

// UpdatePeerStats updates peer statistics from WireGuard
func (s *PeerService) UpdatePeerStats(ctx context.Context, peerID string, stats *domain.UpdatePeerStatsRequest) error {
	return s.peerRepo.UpdateStats(ctx, peerID, stats)
}

// DeletePeer soft-deletes a peer
func (s *PeerService) DeletePeer(ctx context.Context, peerID string) error {
	return s.peerRepo.Delete(ctx, peerID)
}

// GetPeerStats retrieves real-time statistics for a peer
func (s *PeerService) GetPeerStats(ctx context.Context, peerID string) (*domain.PeerStats, error) {
	peer, err := s.peerRepo.GetByID(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Calculate latency (placeholder - would be implemented with actual ping)
	latency := 0
	if peer.LastHandshake != nil {
		// Simple heuristic: if handshake is recent, assume low latency
		if time.Since(*peer.LastHandshake) < 30*time.Second {
			latency = 50 // milliseconds
		}
	}

	return &domain.PeerStats{
		PeerID:        peer.ID,
		Endpoint:      peer.Endpoint,
		LastHandshake: peer.LastHandshake,
		RxBytes:       peer.RxBytes,
		TxBytes:       peer.TxBytes,
		Active:        peer.Active,
		Latency:       latency,
	}, nil
}

// GetNetworkPeerStats retrieves statistics for all peers in a network
func (s *PeerService) GetNetworkPeerStats(ctx context.Context, networkID string) ([]*domain.PeerStats, error) {
	peers, err := s.peerRepo.GetByNetworkID(ctx, networkID)
	if err != nil {
		return nil, err
	}

	stats := make([]*domain.PeerStats, 0, len(peers))
	for _, peer := range peers {
		peerStats, err := s.GetPeerStats(ctx, peer.ID)
		if err != nil {
			// Log error but continue with other peers
			continue
		}
		stats = append(stats, peerStats)
	}

	return stats, nil
}

// RotatePeerKeys rotates the WireGuard keys for a peer
func (s *PeerService) RotatePeerKeys(ctx context.Context, peerID string) (*domain.Peer, error) {
	peer, err := s.peerRepo.GetByID(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Generate new key pair
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, fmt.Sprintf("failed to generate keys: %v", err), nil)
	}

	// Update peer with new public key
	peer.PublicKey = keyPair.PublicKey

	// Optionally regenerate preshared key
	if peer.PresharedKey != "" {
		psk, err := wireguard.GeneratePresharedKey()
		if err != nil {
			return nil, domain.NewError(domain.ErrInternalServer, fmt.Sprintf("failed to generate preshared key: %v", err), nil)
		}
		peer.PresharedKey = psk
	}

	if err := s.peerRepo.Update(ctx, peer); err != nil {
		return nil, err
	}

	return peer, nil
}
