package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLiteNetworkTest(t *testing.T) (*SQLiteNetworkRepository, func()) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "networks.db")
	db, err := database.ConnectSQLite(dbPath)
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	return NewSQLiteNetworkRepository(db), func() { db.Close() }
}

func TestSQLiteNetworkRepository_CreateAndGet(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), n))

	got, err := repo.GetByID(context.Background(), "net-1")
	require.NoError(t, err)
	assert.Equal(t, n.Name, got.Name)
	assert.Equal(t, n.CIDR, got.CIDR)
	assert.Equal(t, domain.NetworkVisibilityPublic, got.Visibility)
}

func TestSQLiteNetworkRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	_, err := repo.GetByID(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSQLiteNetworkRepository_List(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple networks
	for i := 1; i <= 5; i++ {
		n := &domain.Network{
			ID:         stringID("net", i),
			TenantID:   "tenant-1",
			Name:       stringID("Network", i),
			CIDR:       stringID("10.0.", i) + ".0/24",
			Visibility: domain.NetworkVisibilityPublic,
			JoinPolicy: domain.JoinPolicyOpen,
			CreatedBy:  "user-1",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
			UpdatedAt:  time.Now(),
		}
		require.NoError(t, repo.Create(ctx, n))
	}

	// Test basic list
	networks, _, err := repo.List(ctx, NetworkFilter{TenantID: "tenant-1", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, networks, 5)

	// Test with limit
	networks, _, err = repo.List(ctx, NetworkFilter{TenantID: "tenant-1", Limit: 2})
	require.NoError(t, err)
	assert.Len(t, networks, 2)

	// Test visibility filter
	networks, _, err = repo.List(ctx, NetworkFilter{TenantID: "tenant-1", Visibility: "public", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, networks, 5)
}

func TestSQLiteNetworkRepository_Update(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	ctx := context.Background()

	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(ctx, n))

	// Update using mutator function
	dns := "8.8.8.8"
	mtu := 1420
	split := true

	got, err := repo.Update(ctx, "net-1", func(n *domain.Network) error {
		n.Name = "UpdatedNet"
		n.Visibility = domain.NetworkVisibilityPrivate
		n.DNS = &dns
		n.MTU = &mtu
		n.SplitTunnel = &split
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "UpdatedNet", got.Name)
	assert.Equal(t, domain.NetworkVisibilityPrivate, got.Visibility)
	assert.Equal(t, "8.8.8.8", *got.DNS)
	assert.Equal(t, 1420, *got.MTU)
	assert.True(t, *got.SplitTunnel)
}

func TestSQLiteNetworkRepository_SoftDelete(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	ctx := context.Background()

	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(ctx, n))

	// Soft delete
	require.NoError(t, repo.SoftDelete(ctx, "net-1", time.Now()))

	// Should not appear in list
	networks, _, err := repo.List(ctx, NetworkFilter{TenantID: "tenant-1", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, networks, 0)
}

func TestSQLiteNetworkRepository_CheckCIDROverlap(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	ctx := context.Background()

	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(ctx, n))

	// Check overlap - same CIDR exists
	overlap, err := repo.CheckCIDROverlap(ctx, "10.0.0.0/24", "", "tenant-1")
	require.NoError(t, err)
	assert.True(t, overlap)

	// Check no overlap - different CIDR
	overlap, err = repo.CheckCIDROverlap(ctx, "192.168.0.0/24", "", "tenant-1")
	require.NoError(t, err)
	assert.False(t, overlap)

	// Check self-exclude
	overlap, err = repo.CheckCIDROverlap(ctx, "10.0.0.0/24", "net-1", "tenant-1")
	require.NoError(t, err)
	assert.False(t, overlap)
}

func TestSQLiteNetworkRepository_Count(t *testing.T) {
	repo, cleanup := setupSQLiteNetworkTest(t)
	defer cleanup()

	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(ctx, n))

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func stringID(prefix string, i int) string {
	return prefix + "-" + string(rune('0'+i))
}
