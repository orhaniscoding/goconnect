package service

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestPublicKey generates a valid WireGuard public key for testing
func generateTestPublicKey(t *testing.T) string {
	keyPair, err := wireguard.GenerateKeyPair()
	require.NoError(t, err)
	return keyPair.PublicKey
}

func setupPeerProvisioningTest(t *testing.T) (*PeerProvisioningService, context.Context) {
	peerRepo := repository.NewInMemoryPeerRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	ipamRepo := repository.NewInMemoryIPAM()

	service := NewPeerProvisioningService(peerRepo, deviceRepo, networkRepo, membershipRepo, ipamRepo)
	ctx := context.Background()

	return service, ctx
}

func TestProvisionPeersForNewMember(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	// Create network
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	// Create user devices
	device1 := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Laptop",
		PubKey:   generateTestPublicKey(t),
		Platform: "linux",
	}
	device2 := &domain.Device{
		ID:       "device-2",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Phone",
		PubKey:   generateTestPublicKey(t),
		Platform: "android",
	}
	require.NoError(t, service.deviceRepo.Create(ctx, device1))
	require.NoError(t, service.deviceRepo.Create(ctx, device2))

	// Provision peers for new member
	err := service.ProvisionPeersForNewMember(ctx, network.ID, "user-1")
	require.NoError(t, err)

	// Verify peers were created
	peers, err := service.peerRepo.GetByNetworkID(ctx, network.ID)
	require.NoError(t, err)
	assert.Len(t, peers, 2)

	// Verify peer details
	peer1, err := service.peerRepo.GetByNetworkAndDevice(ctx, network.ID, device1.ID)
	require.NoError(t, err)
	assert.Equal(t, device1.PubKey, peer1.PublicKey)
	assert.NotEmpty(t, peer1.AllowedIPs)
	assert.Equal(t, 25, peer1.PersistentKeepalive)

	peer2, err := service.peerRepo.GetByNetworkAndDevice(ctx, network.ID, device2.ID)
	require.NoError(t, err)
	assert.Equal(t, device2.PubKey, peer2.PublicKey)
}

func TestProvisionPeersForNewMember_NoDevices(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	// User has no devices - should not error
	err := service.ProvisionPeersForNewMember(ctx, network.ID, "user-1")
	assert.NoError(t, err)

	// Verify no peers were created
	peers, err := service.peerRepo.GetByNetworkID(ctx, network.ID)
	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestProvisionPeersForNewDevice(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	// Create networks
	network1 := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Network 1",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	network2 := &domain.Network{
		ID:         "net-2",
		TenantID:   "tenant-1",
		Name:       "Network 2",
		CIDR:       "10.1.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network1))
	require.NoError(t, service.networkRepo.Create(ctx, network2))

	// Create memberships
	now := time.Now()
	_, err := service.membershipRepo.UpsertApproved(ctx, network1.ID, "user-1", domain.RoleMember, now)
	require.NoError(t, err)
	_, err = service.membershipRepo.UpsertApproved(ctx, network2.ID, "user-1", domain.RoleMember, now)
	require.NoError(t, err)

	// Create new device
	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "New Phone",
		PubKey:   generateTestPublicKey(t),
		Platform: "ios",
	}
	require.NoError(t, service.deviceRepo.Create(ctx, device))

	// Provision peers for new device
	err = service.ProvisionPeersForNewDevice(ctx, device)
	require.NoError(t, err)

	// Verify peers were created in both networks
	peer1, err := service.peerRepo.GetByNetworkAndDevice(ctx, network1.ID, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.PubKey, peer1.PublicKey)

	peer2, err := service.peerRepo.GetByNetworkAndDevice(ctx, network2.ID, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.PubKey, peer2.PublicKey)
}

func TestProvisionPeersForNewDevice_NoMemberships(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Device",
		PubKey:   generateTestPublicKey(t),
		Platform: "windows",
	}

	// User has no memberships - should not error
	err := service.ProvisionPeersForNewDevice(ctx, device)
	assert.NoError(t, err)

	// Verify no peers were created
	peers, err := service.peerRepo.GetByDeviceID(ctx, device.ID)
	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestDeprovisionPeersForRemovedMember(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	// Setup network and devices
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Device",
		PubKey:   generateTestPublicKey(t),
		Platform: "linux",
	}
	require.NoError(t, service.deviceRepo.Create(ctx, device))

	// Create peer
	peer := &domain.Peer{
		ID:                  "peer-1",
		NetworkID:           network.ID,
		DeviceID:            device.ID,
		TenantID:            "tenant-1",
		PublicKey:           device.PubKey,
		AllowedIPs:          []string{"10.0.0.5/32"},
		PersistentKeepalive: 25,
	}
	require.NoError(t, service.peerRepo.Create(ctx, peer))

	// Deprovision
	err := service.DeprovisionPeersForRemovedMember(ctx, network.ID, "user-1")
	require.NoError(t, err)

	// Verify peer was soft-deleted
	_, err = service.peerRepo.GetByID(ctx, peer.ID)
	assert.Error(t, err) // Should return not found error
}

func TestProvisionPeerForDevice_DuplicatePeer(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	// Setup
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Device",
		PubKey:   generateTestPublicKey(t),
		Platform: "macos",
	}
	require.NoError(t, service.deviceRepo.Create(ctx, device))

	// First provision - should succeed
	err := service.ProvisionPeersForNewMember(ctx, network.ID, "user-1")
	require.NoError(t, err)

	// Second provision - should be idempotent (not error)
	err = service.ProvisionPeersForNewMember(ctx, network.ID, "user-1")
	assert.NoError(t, err)

	// Verify only one peer exists
	peers, err := service.peerRepo.GetByNetworkID(ctx, network.ID)
	require.NoError(t, err)
	assert.Len(t, peers, 1)
}

func TestGetDevicePeerStatus(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	// Setup networks
	network1 := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Network 1",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	network2 := &domain.Network{
		ID:         "net-2",
		TenantID:   "tenant-1",
		Name:       "Network 2",
		CIDR:       "10.1.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network1))
	require.NoError(t, service.networkRepo.Create(ctx, network2))

	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Device",
		PubKey:   generateTestPublicKey(t),
		Platform: "android",
	}

	// Create peers in both networks
	peer1 := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  network1.ID,
		DeviceID:   device.ID,
		TenantID:   "tenant-1",
		PublicKey:  device.PubKey,
		AllowedIPs: []string{"10.0.0.5/32"},
	}
	peer2 := &domain.Peer{
		ID:         "peer-2",
		NetworkID:  network2.ID,
		DeviceID:   device.ID,
		TenantID:   "tenant-1",
		PublicKey:  device.PubKey,
		AllowedIPs: []string{"10.1.0.5/32"},
	}
	require.NoError(t, service.peerRepo.Create(ctx, peer1))
	require.NoError(t, service.peerRepo.Create(ctx, peer2))

	// Get peer status
	status, err := service.GetDevicePeerStatus(ctx, device.ID)
	require.NoError(t, err)
	assert.Len(t, status, 2)
	assert.Contains(t, status, network1.ID)
	assert.Contains(t, status, network2.ID)
}

func TestProvisionPeerForDevice_IPAllocation(t *testing.T) {
	service, ctx := setupPeerProvisioningTest(t)

	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
	}
	require.NoError(t, service.networkRepo.Create(ctx, network))

	// Create multiple devices with unique keys
	for i := 1; i <= 5; i++ {
		device := &domain.Device{
			ID:       domain.GenerateNetworkID(),
			UserID:   "user-1",
			TenantID: "tenant-1",
			Name:     "Device",
			PubKey:   generateTestPublicKey(t),
			Platform: "linux",
		}
		require.NoError(t, service.deviceRepo.Create(ctx, device))
	}

	// Provision peers
	err := service.ProvisionPeersForNewMember(ctx, network.ID, "user-1")
	require.NoError(t, err)

	// Verify all peers have unique IPs
	peers, err := service.peerRepo.GetByNetworkID(ctx, network.ID)
	require.NoError(t, err)

	ipSet := make(map[string]bool)
	for _, peer := range peers {
		require.NotEmpty(t, peer.AllowedIPs)
		ip := peer.AllowedIPs[0]
		assert.False(t, ipSet[ip], "Duplicate IP allocation: %s", ip)
		ipSet[ip] = true
	}
}
