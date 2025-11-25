package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test peer
func mkPeer(id, networkID, deviceID, tenantID, pubkey string) *domain.Peer {
	now := time.Now()
	return &domain.Peer{
		ID:                  id,
		NetworkID:           networkID,
		DeviceID:            deviceID,
		TenantID:            tenantID,
		PublicKey:           pubkey,
		PresharedKey:        "",
		Endpoint:            "",
		AllowedIPs:          []string{"10.0.0.1/32"},
		PersistentKeepalive: 25,
		Active:              false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// ==================== Constructor Tests ====================

func TestNewInMemoryPeerRepository(t *testing.T) {
	repo := NewInMemoryPeerRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.peers)
	assert.Equal(t, 0, len(repo.peers))
}

// ==================== Create Tests ====================

func TestPeerRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123456789012345678901234567890123456")

	err := repo.Create(ctx, peer)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.peers))
	assert.NotEmpty(t, peer.CreatedAt)
	assert.NotEmpty(t, peer.UpdatedAt)
}

func TestPeerRepository_Create_GeneratesID(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("", "network-1", "device-1", "tenant-1", "pubkey123456789012345678901234567890123456")

	err := repo.Create(ctx, peer)

	require.NoError(t, err)
	assert.NotEmpty(t, peer.ID)
}

func TestPeerRepository_Create_DuplicateDeviceInNetwork(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer1 := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey1")
	peer2 := mkPeer("peer-2", "network-1", "device-1", "tenant-1", "pubkey2")

	err1 := repo.Create(ctx, peer1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, peer2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrConflict, domainErr.Code)
	assert.Contains(t, domainErr.Message, "device already has a peer in this network")
}

func TestPeerRepository_Create_SameDeviceDifferentNetworks(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer1 := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey1")
	peer2 := mkPeer("peer-2", "network-2", "device-1", "tenant-1", "pubkey2")

	err1 := repo.Create(ctx, peer1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, peer2)
	require.NoError(t, err2)

	assert.Equal(t, 2, len(repo.peers))
}

func TestPeerRepository_Create_MultiplePeers(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peers := []*domain.Peer{
		mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey1"),
		mkPeer("peer-2", "network-1", "device-2", "tenant-1", "pubkey2"),
		mkPeer("peer-3", "network-2", "device-1", "tenant-1", "pubkey3"),
		mkPeer("peer-4", "network-2", "device-3", "tenant-2", "pubkey4"),
	}

	for _, peer := range peers {
		err := repo.Create(ctx, peer)
		require.NoError(t, err)
	}

	assert.Equal(t, 4, len(repo.peers))
}

// ==================== GetByID Tests ====================

func TestPeerRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	original := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, original)

	retrieved, err := repo.GetByID(ctx, "peer-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.NetworkID, retrieved.NetworkID)
	assert.Equal(t, original.DeviceID, retrieved.DeviceID)
	assert.Equal(t, original.PublicKey, retrieved.PublicKey)
}

func TestPeerRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	retrieved, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_GetByID_DisabledPeer(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	// Should not find disabled peer
	retrieved, err := repo.GetByID(ctx, "peer-1")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

// ==================== GetByNetworkID Tests ====================

func TestPeerRepository_GetByNetworkID_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-1", "device-2", "tenant-1", "pk2"))
	repo.Create(ctx, mkPeer("peer-3", "network-2", "device-3", "tenant-1", "pk3"))

	peers, err := repo.GetByNetworkID(ctx, "network-1")

	require.NoError(t, err)
	assert.Len(t, peers, 2)
	for _, peer := range peers {
		assert.Equal(t, "network-1", peer.NetworkID)
	}
}

func TestPeerRepository_GetByNetworkID_Empty(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peers, err := repo.GetByNetworkID(ctx, "non-existent-network")

	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestPeerRepository_GetByNetworkID_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-1", "device-2", "tenant-1", "pk2"))

	// Disable one peer
	repo.Delete(ctx, "peer-1")

	peers, err := repo.GetByNetworkID(ctx, "network-1")

	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-2", peers[0].ID)
}

// ==================== GetByDeviceID Tests ====================

func TestPeerRepository_GetByDeviceID_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-1", "tenant-1", "pk2"))
	repo.Create(ctx, mkPeer("peer-3", "network-1", "device-2", "tenant-1", "pk3"))

	peers, err := repo.GetByDeviceID(ctx, "device-1")

	require.NoError(t, err)
	assert.Len(t, peers, 2)
	for _, peer := range peers {
		assert.Equal(t, "device-1", peer.DeviceID)
	}
}

func TestPeerRepository_GetByDeviceID_Empty(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peers, err := repo.GetByDeviceID(ctx, "non-existent-device")

	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestPeerRepository_GetByDeviceID_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-1", "tenant-1", "pk2"))

	// Disable one peer
	repo.Delete(ctx, "peer-1")

	peers, err := repo.GetByDeviceID(ctx, "device-1")

	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-2", peers[0].ID)
}

// ==================== GetByNetworkAndDevice Tests ====================

func TestPeerRepository_GetByNetworkAndDevice_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-1", "device-2", "tenant-1", "pk2"))
	repo.Create(ctx, mkPeer("peer-3", "network-2", "device-1", "tenant-1", "pk3"))

	peer, err := repo.GetByNetworkAndDevice(ctx, "network-1", "device-1")

	require.NoError(t, err)
	require.NotNil(t, peer)
	assert.Equal(t, "peer-1", peer.ID)
	assert.Equal(t, "network-1", peer.NetworkID)
	assert.Equal(t, "device-1", peer.DeviceID)
}

func TestPeerRepository_GetByNetworkAndDevice_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer, err := repo.GetByNetworkAndDevice(ctx, "network-1", "device-1")

	require.Error(t, err)
	assert.Nil(t, peer)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_GetByNetworkAndDevice_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	peer, err := repo.GetByNetworkAndDevice(ctx, "network-1", "device-1")

	require.Error(t, err)
	assert.Nil(t, peer)
}

// ==================== GetByPublicKey Tests ====================

func TestPeerRepository_GetByPublicKey_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "unique-pubkey-123"))

	peer, err := repo.GetByPublicKey(ctx, "unique-pubkey-123")

	require.NoError(t, err)
	require.NotNil(t, peer)
	assert.Equal(t, "peer-1", peer.ID)
	assert.Equal(t, "unique-pubkey-123", peer.PublicKey)
}

func TestPeerRepository_GetByPublicKey_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer, err := repo.GetByPublicKey(ctx, "non-existent-pubkey")

	require.Error(t, err)
	assert.Nil(t, peer)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_GetByPublicKey_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "unique-pubkey-123"))

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	peer, err := repo.GetByPublicKey(ctx, "unique-pubkey-123")

	require.Error(t, err)
	assert.Nil(t, peer)
}

// ==================== GetActivePeers Tests ====================

func TestPeerRepository_GetActivePeers_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create active peer
	activePeer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1")
	activePeer.Active = true
	repo.Create(ctx, activePeer)

	// Create inactive peer
	inactivePeer := mkPeer("peer-2", "network-1", "device-2", "tenant-1", "pk2")
	inactivePeer.Active = false
	repo.Create(ctx, inactivePeer)

	// Create active peer in different network
	otherNetworkPeer := mkPeer("peer-3", "network-2", "device-3", "tenant-1", "pk3")
	otherNetworkPeer.Active = true
	repo.Create(ctx, otherNetworkPeer)

	peers, err := repo.GetActivePeers(ctx, "network-1")

	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-1", peers[0].ID)
	assert.True(t, peers[0].Active)
}

func TestPeerRepository_GetActivePeers_Empty(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create only inactive peer
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1")
	peer.Active = false
	repo.Create(ctx, peer)

	peers, err := repo.GetActivePeers(ctx, "network-1")

	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestPeerRepository_GetActivePeers_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create active peer
	activePeer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1")
	activePeer.Active = true
	repo.Create(ctx, activePeer)

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	peers, err := repo.GetActivePeers(ctx, "network-1")

	require.NoError(t, err)
	assert.Empty(t, peers)
}

// ==================== GetAllActive Tests ====================

func TestPeerRepository_GetAllActive_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-2", "tenant-2", "pk2"))
	repo.Create(ctx, mkPeer("peer-3", "network-3", "device-3", "tenant-1", "pk3"))

	peers, err := repo.GetAllActive(ctx)

	require.NoError(t, err)
	assert.Len(t, peers, 3)
}

func TestPeerRepository_GetAllActive_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-2", "tenant-2", "pk2"))

	// Disable one peer
	repo.Delete(ctx, "peer-1")

	peers, err := repo.GetAllActive(ctx)

	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-2", peers[0].ID)
}

// ==================== Update Tests ====================

func TestPeerRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	original := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, original)
	originalCreatedAt := repo.peers["peer-1"].CreatedAt

	// Update peer
	updatedPeer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	updatedPeer.Endpoint = "192.168.1.100:51820"
	updatedPeer.Active = true
	updatedPeer.AllowedIPs = []string{"10.0.0.1/32", "10.0.0.2/32"}

	time.Sleep(10 * time.Millisecond)
	err := repo.Update(ctx, updatedPeer)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "peer-1")
	assert.Equal(t, "192.168.1.100:51820", retrieved.Endpoint)
	assert.True(t, retrieved.Active)
	assert.Len(t, retrieved.AllowedIPs, 2)
	assert.Equal(t, originalCreatedAt, retrieved.CreatedAt) // CreatedAt preserved
	assert.True(t, retrieved.UpdatedAt.After(originalCreatedAt))
}

func TestPeerRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("non-existent", "network-1", "device-1", "tenant-1", "pubkey123")

	err := repo.Update(ctx, peer)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_Update_DisabledPeer(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	// Try to update disabled peer
	updatedPeer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	err := repo.Update(ctx, updatedPeer)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

// ==================== UpdateStats Tests ====================

func TestPeerRepository_UpdateStats_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	now := time.Now()
	stats := &domain.UpdatePeerStatsRequest{
		Endpoint:      "192.168.1.100:51820",
		LastHandshake: &now,
		RxBytes:       1024,
		TxBytes:       2048,
	}

	err := repo.UpdateStats(ctx, "peer-1", stats)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "peer-1")
	assert.Equal(t, "192.168.1.100:51820", retrieved.Endpoint)
	assert.NotNil(t, retrieved.LastHandshake)
	assert.Equal(t, int64(1024), retrieved.RxBytes)
	assert.Equal(t, int64(2048), retrieved.TxBytes)
	assert.True(t, retrieved.Active) // Active because handshake is recent
}

func TestPeerRepository_UpdateStats_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	stats := &domain.UpdatePeerStatsRequest{
		RxBytes: 1024,
		TxBytes: 2048,
	}

	err := repo.UpdateStats(ctx, "non-existent", stats)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_UpdateStats_DisabledPeer(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	// Disable the peer
	repo.Delete(ctx, "peer-1")

	stats := &domain.UpdatePeerStatsRequest{
		RxBytes: 1024,
		TxBytes: 2048,
	}

	err := repo.UpdateStats(ctx, "peer-1", stats)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_UpdateStats_SetsActiveBasedOnHandshake(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	// Recent handshake -> Active = true
	recentTime := time.Now()
	stats := &domain.UpdatePeerStatsRequest{
		LastHandshake: &recentTime,
	}
	err := repo.UpdateStats(ctx, "peer-1", stats)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "peer-1")
	assert.True(t, retrieved.Active)

	// Old handshake -> Active = false
	oldTime := time.Now().Add(-10 * time.Minute)
	stats = &domain.UpdatePeerStatsRequest{
		LastHandshake: &oldTime,
	}
	err = repo.UpdateStats(ctx, "peer-1", stats)
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(ctx, "peer-1")
	assert.False(t, retrieved.Active)
}

func TestPeerRepository_UpdateStats_PartialUpdate(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	peer.Endpoint = "192.168.1.1:51820"
	repo.Create(ctx, peer)

	// Update only traffic, not endpoint
	stats := &domain.UpdatePeerStatsRequest{
		Endpoint: "", // Empty endpoint should not overwrite
		RxBytes:  5000,
		TxBytes:  3000,
	}

	err := repo.UpdateStats(ctx, "peer-1", stats)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "peer-1")
	assert.Equal(t, "192.168.1.1:51820", retrieved.Endpoint) // Preserved
	assert.Equal(t, int64(5000), retrieved.RxBytes)
	assert.Equal(t, int64(3000), retrieved.TxBytes)
}

// ==================== Delete Tests (Soft Delete) ====================

func TestPeerRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	peer.Active = true
	repo.Create(ctx, peer)

	err := repo.Delete(ctx, "peer-1")

	require.NoError(t, err)

	// Peer should still exist in storage but be disabled
	assert.Equal(t, 1, len(repo.peers))
	assert.NotNil(t, repo.peers["peer-1"].DisabledAt)
	assert.False(t, repo.peers["peer-1"].Active)

	// Should not be retrievable by normal methods
	_, err = repo.GetByID(ctx, "peer-1")
	assert.Error(t, err)
}

func TestPeerRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

// ==================== HardDelete Tests ====================

func TestPeerRepository_HardDelete_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	err := repo.HardDelete(ctx, "peer-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.peers)) // Actually removed
}

func TestPeerRepository_HardDelete_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	err := repo.HardDelete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestPeerRepository_HardDelete_AfterSoftDelete(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	// Soft delete first
	err := repo.Delete(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.peers)) // Still in storage

	// Hard delete
	err = repo.HardDelete(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.peers)) // Now removed
}

// ==================== ListByTenant Tests ====================

func TestPeerRepository_ListByTenant_Success(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-2", "tenant-1", "pk2"))
	repo.Create(ctx, mkPeer("peer-3", "network-1", "device-3", "tenant-2", "pk3"))

	peers, err := repo.ListByTenant(ctx, "tenant-1", 50, 0)

	require.NoError(t, err)
	assert.Len(t, peers, 2)
	for _, peer := range peers {
		assert.Equal(t, "tenant-1", peer.TenantID)
	}
}

func TestPeerRepository_ListByTenant_Empty(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peers, err := repo.ListByTenant(ctx, "non-existent-tenant", 50, 0)

	require.NoError(t, err)
	assert.Empty(t, peers)
}

func TestPeerRepository_ListByTenant_ExcludesDisabled(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	repo.Create(ctx, mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pk1"))
	repo.Create(ctx, mkPeer("peer-2", "network-2", "device-2", "tenant-1", "pk2"))

	// Disable one peer
	repo.Delete(ctx, "peer-1")

	peers, err := repo.ListByTenant(ctx, "tenant-1", 50, 0)

	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-2", peers[0].ID)
}

func TestPeerRepository_ListByTenant_Pagination(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create 5 peers for tenant-1
	for i := 1; i <= 5; i++ {
		repo.Create(ctx, mkPeer("", "network-1", "device-"+string(rune('0'+i)), "tenant-1", "pk"+string(rune('0'+i))))
	}

	// Get first 2
	page1, err := repo.ListByTenant(ctx, "tenant-1", 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get next 2
	page2, err := repo.ListByTenant(ctx, "tenant-1", 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	// Get last 1
	page3, err := repo.ListByTenant(ctx, "tenant-1", 2, 4)
	require.NoError(t, err)
	assert.Len(t, page3, 1)

	// Offset beyond data
	page4, err := repo.ListByTenant(ctx, "tenant-1", 2, 10)
	require.NoError(t, err)
	assert.Empty(t, page4)
}

// ==================== Full Lifecycle Tests ====================

func TestPeerRepository_FullCRUDCycle(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	err := repo.Create(ctx, peer)
	require.NoError(t, err)

	// Read
	retrieved, err := repo.GetByID(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, "peer-1", retrieved.ID)

	// Update
	retrieved.Endpoint = "192.168.1.100:51820"
	err = repo.Update(ctx, retrieved)
	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "peer-1")
	assert.Equal(t, "192.168.1.100:51820", updated.Endpoint)

	// Soft Delete
	err = repo.Delete(ctx, "peer-1")
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, "peer-1")
	assert.Error(t, err)

	// Hard Delete
	err = repo.HardDelete(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.peers))
}

func TestPeerRepository_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create initial peer
	peer := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey123")
	repo.Create(ctx, peer)

	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			repo.GetByID(ctx, "peer-1")
			repo.GetByNetworkID(ctx, "network-1")
			repo.GetAllActive(ctx)
		}
		done <- true
	}()

	// Concurrent stats updates
	go func() {
		for i := 0; i < 100; i++ {
			stats := &domain.UpdatePeerStatsRequest{
				RxBytes: int64(i * 100),
				TxBytes: int64(i * 50),
			}
			repo.UpdateStats(ctx, "peer-1", stats)
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic
	retrieved, err := repo.GetByID(ctx, "peer-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestPeerRepository_SoftDeleteAllowsRejoining(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create peer
	peer1 := mkPeer("peer-1", "network-1", "device-1", "tenant-1", "pubkey1")
	err := repo.Create(ctx, peer1)
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(ctx, "peer-1")
	require.NoError(t, err)

	// Device can rejoin the same network with a new peer
	peer2 := mkPeer("peer-2", "network-1", "device-1", "tenant-1", "pubkey2")
	err = repo.Create(ctx, peer2)
	require.NoError(t, err) // Should succeed because peer-1 is disabled

	// Verify new peer exists
	retrieved, err := repo.GetByNetworkAndDevice(ctx, "network-1", "device-1")
	require.NoError(t, err)
	assert.Equal(t, "peer-2", retrieved.ID)
}
