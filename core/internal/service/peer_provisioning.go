package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
)

// PeerProvisioningService handles automatic peer creation for devices in networks
type PeerProvisioningService struct {
	peerRepo       repository.PeerRepository
	deviceRepo     repository.DeviceRepository
	networkRepo    repository.NetworkRepository
	membershipRepo repository.MembershipRepository
	ipamRepo       repository.IPAMRepository
}

// NewPeerProvisioningService creates a new peer provisioning service
func NewPeerProvisioningService(
	peerRepo repository.PeerRepository,
	deviceRepo repository.DeviceRepository,
	networkRepo repository.NetworkRepository,
	membershipRepo repository.MembershipRepository,
	ipamRepo repository.IPAMRepository,
) *PeerProvisioningService {
	return &PeerProvisioningService{
		peerRepo:       peerRepo,
		deviceRepo:     deviceRepo,
		networkRepo:    networkRepo,
		membershipRepo: membershipRepo,
		ipamRepo:       ipamRepo,
	}
}

// ProvisionPeersForNewMember creates peers for all user's devices when they join a network
func (s *PeerProvisioningService) ProvisionPeersForNewMember(ctx context.Context, networkID, userID string) error {
	// Get network
	network, err := s.networkRepo.GetByID(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Get all user's devices using List with filter
	devices, _, err := s.deviceRepo.List(ctx, domain.DeviceFilter{
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to list user devices: %w", err)
	}

	if len(devices) == 0 {
		// No devices to provision, not an error
		slog.Info("User has no devices to provision in network", "user_id", userID, "network_id", networkID)
		return nil
	}

	// Create peers for each device
	successCount := 0
	for _, device := range devices {
		if err := s.provisionPeerForDevice(ctx, network, device); err != nil {
			// Log error but continue with other devices
			slog.Error("Failed to provision peer for device", "device_id", device.ID, "network_id", networkID, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("Provisioned peers for user", "success_count", successCount, "total_devices", len(devices), "user_id", userID, "network_id", networkID)
	return nil
}

// ProvisionPeersForNewDevice creates peers for a device in all networks the user is a member of
func (s *PeerProvisioningService) ProvisionPeersForNewDevice(ctx context.Context, device *domain.Device) error {
	// Get all networks in tenant
	networks, _, err := s.networkRepo.List(ctx, repository.NetworkFilter{
		TenantID: device.TenantID,
		Limit:    1000, // Large limit to get all networks
	})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	if len(networks) == 0 {
		slog.Info("No networks found in tenant for device", "device_id", device.ID)
		return nil
	}

	// Check each network for user membership
	approvedNetworks := make([]*domain.Network, 0)
	for _, network := range networks {
		membership, err := s.membershipRepo.Get(ctx, network.ID, device.UserID)
		if err != nil {
			continue // User is not a member
		}
		if membership.Status == domain.StatusApproved {
			approvedNetworks = append(approvedNetworks, network)
		}
	}

	if len(approvedNetworks) == 0 {
		slog.Info("User has no approved memberships for device", "device_id", device.ID, "user_id", device.UserID)
		return nil
	}

	// Create peers for each approved network
	successCount := 0
	for _, network := range approvedNetworks {
		if err := s.provisionPeerForDevice(ctx, network, device); err != nil {
			slog.Error("Failed to provision peer for device", "device_id", device.ID, "network_id", network.ID, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("Provisioned peers for device", "success_count", successCount, "total_networks", len(approvedNetworks), "device_id", device.ID)
	return nil
}

// DeprovisionPeersForRemovedMember removes all peers for a user's devices when they leave a network
func (s *PeerProvisioningService) DeprovisionPeersForRemovedMember(ctx context.Context, networkID, userID string) error {
	// Get all user's devices using List with filter
	devices, _, err := s.deviceRepo.List(ctx, domain.DeviceFilter{
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to list user devices: %w", err)
	}

	successCount := 0
	for _, device := range devices {
		// Get peer for this device-network combination
		peer, err := s.peerRepo.GetByNetworkAndDevice(ctx, networkID, device.ID)
		if err != nil {
			// Peer might not exist, continue
			continue
		}

		// Soft delete the peer
		if err := s.peerRepo.Delete(ctx, peer.ID); err != nil {
			slog.Error("Failed to delete peer", "peer_id", peer.ID, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("Deprovisioned peers for user", "count", successCount, "user_id", userID, "network_id", networkID)
	return nil
}

// provisionPeerForDevice creates a peer for a specific device in a network
func (s *PeerProvisioningService) provisionPeerForDevice(ctx context.Context, network *domain.Network, device *domain.Device) error {
	// Check if peer already exists
	existing, err := s.peerRepo.GetByNetworkAndDevice(ctx, network.ID, device.ID)
	if err == nil && existing != nil {
		// Peer already exists, not an error
		slog.Info("Peer already exists", "device_id", device.ID, "network_id", network.ID)
		return nil
	}

	// Allocate IP for the device in this network
	// Use device ID instead of user ID for unique IP per device
	allocation, err := s.ipamRepo.GetOrAllocate(ctx, network.ID, device.ID, network.CIDR)
	if err != nil {
		return fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Validate device public key
	if err := wireguard.ValidatePublicKey(device.PubKey); err != nil {
		return fmt.Errorf("invalid device public key: %w", err)
	}

	// Create peer
	peer := &domain.Peer{
		NetworkID:           network.ID,
		DeviceID:            device.ID,
		TenantID:            network.TenantID,
		PublicKey:           device.PubKey,
		AllowedIPs:          []string{fmt.Sprintf("%s/32", allocation.IP)}, // Single IP for the device
		PersistentKeepalive: 25,                                            // Default keepalive for NAT traversal
		Active:              false,
		RxBytes:             0,
		TxBytes:             0,
	}

	if err := s.peerRepo.Create(ctx, peer); err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}

	slog.Info("Provisioned peer for device", "peer_id", peer.ID, "device_id", device.ID, "network_id", network.ID, "ip", allocation.IP)
	return nil
}

// GetDevicePeerStatus returns the peer provisioning status for a device
func (s *PeerProvisioningService) GetDevicePeerStatus(ctx context.Context, deviceID string) (map[string]*domain.Peer, error) {
	peers, err := s.peerRepo.GetByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	status := make(map[string]*domain.Peer)
	for _, peer := range peers {
		status[peer.NetworkID] = peer
	}

	return status, nil
}
