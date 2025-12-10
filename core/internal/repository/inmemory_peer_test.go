package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryPeerRepository_Create(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "pubkey-1",
		AllowedIPs: []string{"10.0.0.1/32"},
		Active:     true,
	}

	err := repo.Create(ctx, peer)
	require.NoError(t, err)
	assert.NotEmpty(t, peer.ID)
	assert.False(t, peer.CreatedAt.IsZero())
	assert.False(t, peer.UpdatedAt.IsZero())
}

func TestInMemoryPeerRepository_Create_DuplicateDeviceInNetwork(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer1 := &domain.Peer{
		NetworkID: "net-1",
		DeviceID:  "dev-1",
		TenantID:  "tenant-1",
		PublicKey: "pubkey-1",
	}
	require.NoError(t, repo.Create(ctx, peer1))

	peer2 := &domain.Peer{
		NetworkID: "net-1",
		DeviceID:  "dev-1",
		TenantID:  "tenant-1",
		PublicKey: "pubkey-2",
	}

	err := repo.Create(ctx, peer2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device already has a peer")
}

func TestInMemoryPeerRepository_GetByID(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{
		NetworkID: "net-1",
		DeviceID:  "dev-1",
		TenantID:  "tenant-1",
		PublicKey: "pubkey-1",
	}
	require.NoError(t, repo.Create(ctx, peer))

	found, err := repo.GetByID(ctx, peer.ID)
	require.NoError(t, err)
	assert.Equal(t, peer.ID, found.ID)
	assert.Equal(t, peer.NetworkID, found.NetworkID)
}

func TestInMemoryPeerRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "peer not found")
}

func TestInMemoryPeerRepository_GetByNetworkID(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer1 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	peer2 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-2", TenantID: "t1", PublicKey: "pk2"}
	peer3 := &domain.Peer{NetworkID: "net-2", DeviceID: "dev-3", TenantID: "t1", PublicKey: "pk3"}

	require.NoError(t, repo.Create(ctx, peer1))
	require.NoError(t, repo.Create(ctx, peer2))
	require.NoError(t, repo.Create(ctx, peer3))

	peers, err := repo.GetByNetworkID(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestInMemoryPeerRepository_GetByDeviceID(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer1 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	peer2 := &domain.Peer{NetworkID: "net-2", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk2"}

	require.NoError(t, repo.Create(ctx, peer1))
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetByDeviceID(ctx, "dev-1")
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestInMemoryPeerRepository_GetByNetworkAndDevice(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	require.NoError(t, repo.Create(ctx, peer))

	found, err := repo.GetByNetworkAndDevice(ctx, "net-1", "dev-1")
	require.NoError(t, err)
	assert.Equal(t, peer.ID, found.ID)

	_, err = repo.GetByNetworkAndDevice(ctx, "net-1", "dev-nonexistent")
	require.Error(t, err)
}

func TestInMemoryPeerRepository_GetByPublicKey(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "unique-pubkey"}
	require.NoError(t, repo.Create(ctx, peer))

	found, err := repo.GetByPublicKey(ctx, "unique-pubkey")
	require.NoError(t, err)
	assert.Equal(t, peer.ID, found.ID)

	_, err = repo.GetByPublicKey(ctx, "nonexistent")
	require.Error(t, err)
}

func TestInMemoryPeerRepository_GetActivePeers(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer1 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1", Active: true}
	peer2 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-2", TenantID: "t1", PublicKey: "pk2", Active: false}

	require.NoError(t, repo.Create(ctx, peer1))
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetActivePeers(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.True(t, peers[0].Active)
}

func TestInMemoryPeerRepository_GetAllActive(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer1 := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	peer2 := &domain.Peer{NetworkID: "net-2", DeviceID: "dev-2", TenantID: "t1", PublicKey: "pk2"}

	require.NoError(t, repo.Create(ctx, peer1))
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetAllActive(ctx)
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestInMemoryPeerRepository_Update(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	require.NoError(t, repo.Create(ctx, peer))

	peer.AllowedIPs = []string{"10.0.0.100/32"}
	peer.Active = true

	err := repo.Update(ctx, peer)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, peer.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.100/32"}, found.AllowedIPs)
	assert.True(t, found.Active)
}

func TestInMemoryPeerRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{ID: "nonexistent"}
	err := repo.Update(ctx, peer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "peer not found")
}

func TestInMemoryPeerRepository_UpdateStats(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	require.NoError(t, repo.Create(ctx, peer))

	now := time.Now()
	stats := &domain.UpdatePeerStatsRequest{
		Endpoint:      "192.168.1.1:51820",
		LastHandshake: &now,
		RxBytes:       1000,
		TxBytes:       2000,
	}

	err := repo.UpdateStats(ctx, peer.ID, stats)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, peer.ID)
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1:51820", found.Endpoint)
	assert.Equal(t, int64(1000), found.RxBytes)
	assert.Equal(t, int64(2000), found.TxBytes)
	assert.True(t, found.Active) // Recent handshake means active
}

func TestInMemoryPeerRepository_UpdateStats_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	stats := &domain.UpdatePeerStatsRequest{RxBytes: 100}
	err := repo.UpdateStats(ctx, "nonexistent", stats)
	require.Error(t, err)
}

func TestInMemoryPeerRepository_Delete(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	require.NoError(t, repo.Create(ctx, peer))

	err := repo.Delete(ctx, peer.ID)
	require.NoError(t, err)

	// Should not be found after soft delete
	_, err = repo.GetByID(ctx, peer.ID)
	require.Error(t, err)
}

func TestInMemoryPeerRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
}

func TestInMemoryPeerRepository_HardDelete(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	peer := &domain.Peer{NetworkID: "net-1", DeviceID: "dev-1", TenantID: "t1", PublicKey: "pk1"}
	require.NoError(t, repo.Create(ctx, peer))

	err := repo.HardDelete(ctx, peer.ID)
	require.NoError(t, err)

	// Should not be found after hard delete
	_, err = repo.GetByID(ctx, peer.ID)
	require.Error(t, err)
}

func TestInMemoryPeerRepository_HardDelete_NotFound(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	err := repo.HardDelete(ctx, "nonexistent")
	require.Error(t, err)
}

func TestInMemoryPeerRepository_ListByTenant(t *testing.T) {
	repo := NewInMemoryPeerRepository()
	ctx := context.Background()

	// Create peers for different tenants
	for i := 0; i < 5; i++ {
		peer := &domain.Peer{
			NetworkID: "net-1",
			DeviceID:  "dev-" + string(rune('a'+i)),
			TenantID:  "tenant-1",
			PublicKey: "pk-" + string(rune('a'+i)),
		}
		require.NoError(t, repo.Create(ctx, peer))
	}

	peer := &domain.Peer{
		NetworkID: "net-2",
		DeviceID:  "dev-z",
		TenantID:  "tenant-2",
		PublicKey: "pk-z",
	}
	require.NoError(t, repo.Create(ctx, peer))

	// List first page
	peers, err := repo.ListByTenant(ctx, "tenant-1", 3, 0)
	require.NoError(t, err)
	assert.Len(t, peers, 3)

	// List second page
	peers, err = repo.ListByTenant(ctx, "tenant-1", 3, 3)
	require.NoError(t, err)
	assert.Len(t, peers, 2)

	// Out of range
	peers, err = repo.ListByTenant(ctx, "tenant-1", 3, 10)
	require.NoError(t, err)
	assert.Len(t, peers, 0)
}
