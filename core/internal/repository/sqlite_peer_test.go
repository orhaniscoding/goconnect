package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLitePeerTest(t *testing.T) (*SQLitePeerRepository, *sql.DB, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "peers.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO devices (id, user_id, tenant_id, name, platform, pubkey, created_at, updated_at) VALUES ('dev-1','user-1','tenant-1','device','linux','PUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEY12',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO devices (id, user_id, tenant_id, name, platform, pubkey, created_at, updated_at) VALUES ('dev-2','user-1','tenant-1','device2','linux','PUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEY21',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-2','tenant-1','n2','192.168.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	return NewSQLitePeerRepository(db), db, func() { db.Close() }
}

func TestSQLitePeerRepository_CreateGet(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	peer := &domain.Peer{
		ID:                  "peer-1",
		NetworkID:           "net-1",
		DeviceID:            "dev-1",
		TenantID:            "tenant-1",
		PublicKey:           "PUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEYPUBKEY12",
		PresharedKey:        "",
		Endpoint:            "127.0.0.1:51820",
		AllowedIPs:          []string{"10.0.0.2/32"},
		PersistentKeepalive: 25,
		Active:              true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	require.NoError(t, repo.Create(context.Background(), peer))

	got, err := repo.GetByID(context.Background(), "peer-1")
	require.NoError(t, err)
	assert.Equal(t, peer.PublicKey, got.PublicKey)
	assert.True(t, got.Active)
}

func TestSQLitePeerRepository_GetByNetworkID(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create peers for net-1
	peer1 := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer1))

	peer2 := &domain.Peer{
		ID:         "peer-2",
		NetworkID:  "net-1",
		DeviceID:   "dev-2",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY2222222222222222222222222222222222222",
		Endpoint:   "127.0.0.1:51821",
		AllowedIPs: []string{"10.0.0.3/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer2))

	// Create peer for net-2
	peer3 := &domain.Peer{
		ID:         "peer-3",
		NetworkID:  "net-2",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY3333333333333333333333333333333333333",
		Endpoint:   "127.0.0.1:51822",
		AllowedIPs: []string{"192.168.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer3))

	// Get by network ID
	peers, err := repo.GetByNetworkID(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestSQLitePeerRepository_GetByDeviceID(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer1 := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer1))

	peer2 := &domain.Peer{
		ID:         "peer-2",
		NetworkID:  "net-2",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY2222222222222222222222222222222222222",
		Endpoint:   "127.0.0.1:51821",
		AllowedIPs: []string{"192.168.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetByDeviceID(ctx, "dev-1")
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestSQLitePeerRepository_GetByNetworkAndDevice(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	got, err := repo.GetByNetworkAndDevice(ctx, "net-1", "dev-1")
	require.NoError(t, err)
	assert.Equal(t, "peer-1", got.ID)

	// Not found case
	_, err = repo.GetByNetworkAndDevice(ctx, "net-1", "dev-2")
	assert.Error(t, err)
}

func TestSQLitePeerRepository_Update(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	// Update peer
	peer.Endpoint = "192.168.1.1:51820"
	peer.AllowedIPs = []string{"10.0.0.2/32", "10.0.0.0/24"}
	peer.Active = false

	require.NoError(t, repo.Update(ctx, peer))

	got, err := repo.GetByID(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1:51820", got.Endpoint)
	assert.False(t, got.Active)
}

func TestSQLitePeerRepository_UpdateStats(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	// Update stats
	now := time.Now()
	stats := &domain.UpdatePeerStatsRequest{
		Endpoint:      "192.168.1.1:51820",
		LastHandshake: &now,
		RxBytes:       1000,
		TxBytes:       2000,
	}
	require.NoError(t, repo.UpdateStats(ctx, "peer-1", stats))

	got, err := repo.GetByID(ctx, "peer-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), got.RxBytes)
	assert.Equal(t, int64(2000), got.TxBytes)
}

func TestSQLitePeerRepository_Delete(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	// Delete (soft delete)
	require.NoError(t, repo.Delete(ctx, "peer-1"))

	// Should not find
	_, err := repo.GetByID(ctx, "peer-1")
	assert.Error(t, err)
}

func TestSQLitePeerRepository_GetActivePeers(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create active peer
	peer1 := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer1))

	// Create inactive peer
	peer2 := &domain.Peer{
		ID:         "peer-2",
		NetworkID:  "net-1",
		DeviceID:   "dev-2",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY2222222222222222222222222222222222222",
		Endpoint:   "127.0.0.1:51821",
		AllowedIPs: []string{"10.0.0.3/32"},
		Active:     false,
	}
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetActivePeers(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer-1", peers[0].ID)
}

func TestSQLitePeerRepository_GetAllActive(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer1 := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer1))

	peer2 := &domain.Peer{
		ID:         "peer-2",
		NetworkID:  "net-2",
		DeviceID:   "dev-2",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY2222222222222222222222222222222222222",
		Endpoint:   "127.0.0.1:51821",
		AllowedIPs: []string{"192.168.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer2))

	peers, err := repo.GetAllActive(ctx)
	require.NoError(t, err)
	assert.Len(t, peers, 2)
}

func TestSQLitePeerRepository_ListByTenant(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-1",
		NetworkID:  "net-1",
		DeviceID:   "dev-1",
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY1111111111111111111111111111111111111",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	peers, err := repo.ListByTenant(ctx, "tenant-1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, peers, 1)
}

func TestSQLitePeerRepository_GetByPublicKey(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	pubKey := "PUBKEY_UNIQUE_123456789012345678901234567890"
	peer := &domain.Peer{
		ID:         "peer-pk-1",
		NetworkID:  "net-1", // Use existing network from setup
		DeviceID:   "dev-1", // Use existing device from setup
		TenantID:   "tenant-1",
		PublicKey:  pubKey,
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	// Get by public key
	found, err := repo.GetByPublicKey(ctx, pubKey)
	require.NoError(t, err)
	assert.Equal(t, "peer-pk-1", found.ID)
	assert.Equal(t, pubKey, found.PublicKey)

	// Get non-existent
	_, err = repo.GetByPublicKey(ctx, "nonexistent_key")
	require.Error(t, err)
}

func TestSQLitePeerRepository_HardDelete(t *testing.T) {
	repo, _, cleanup := setupSQLitePeerTest(t)
	defer cleanup()

	ctx := context.Background()

	peer := &domain.Peer{
		ID:         "peer-hd-1",
		NetworkID:  "net-2", // Use existing network from setup
		DeviceID:   "dev-2", // Use existing device from setup
		TenantID:   "tenant-1",
		PublicKey:  "PUBKEY_HD_1234567890123456789012345678901234",
		Endpoint:   "127.0.0.1:51820",
		AllowedIPs: []string{"10.0.0.2/32"},
		Active:     true,
	}
	require.NoError(t, repo.Create(ctx, peer))

	// Verify it exists
	_, err := repo.GetByID(ctx, "peer-hd-1")
	require.NoError(t, err)

	// Hard delete
	err = repo.HardDelete(ctx, "peer-hd-1")
	require.NoError(t, err)

	// Should not exist anymore (even with disabled_at check)
	_, err = repo.GetByID(ctx, "peer-hd-1")
	require.Error(t, err)

	// Hard delete non-existent should error
	err = repo.HardDelete(ctx, "nonexistent")
	require.Error(t, err)
}
