package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSQLiteIPAMTest creates a test database with necessary seed data
func setupSQLiteIPAMTest(t *testing.T) (*sql.DB, *SQLiteIPAMRepository, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "ipam.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	// Seed tenant
	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	// Seed users
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u1@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-2','tenant-1','u2@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-3','tenant-1','u3@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	// Seed network
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-2','tenant-1','n2','192.168.1.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteIPAMRepository(db)
	return db, repo, func() { db.Close() }
}

func TestSQLiteIPAMRepository_AllocateRelease(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()
	ip, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Contains(t, ip, "10.0.0.")

	err = repo.ReleaseIP(ctx, "net-1", "user-1")
	require.NoError(t, err)
}

func TestSQLiteIPAMRepository_AllocateIP(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate first IP
	ip1, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", ip1)

	// Allocate second IP for different user
	ip2, err := repo.AllocateIP(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", ip2)

	// Allocate third IP
	ip3, err := repo.AllocateIP(ctx, "net-1", "user-3", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.3", ip3)
}

func TestSQLiteIPAMRepository_AllocateIP_InvalidCIDR(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Invalid CIDR
	_, err := repo.AllocateIP(ctx, "net-1", "user-1", "invalid-cidr")
	require.Error(t, err)
}

func TestSQLiteIPAMRepository_AllocateIP_DifferentNetworks(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate in first network
	ip1, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", ip1)

	// Allocate in second network for same user
	ip2, err := repo.AllocateIP(ctx, "net-2", "user-1", "192.168.1.0/24")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.1", ip2)
}

func TestSQLiteIPAMRepository_ReleaseIP(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate an IP
	ip, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", ip)

	// Release it
	err = repo.ReleaseIP(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// List should be empty
	allocs, err := repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Empty(t, allocs)
}

func TestSQLiteIPAMRepository_ReleaseIP_NotFound(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Try to release non-existent allocation
	err := repo.ReleaseIP(ctx, "net-1", "user-1")
	require.Error(t, err)
}

func TestSQLiteIPAMRepository_ListAllocations(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Initially empty
	allocs, err := repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Empty(t, allocs)

	// Allocate some IPs
	_, err = repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	_, err = repo.AllocateIP(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)

	// List allocations
	allocs, err = repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, allocs, 2)

	// Verify allocation data
	ips := make(map[string]string)
	for _, a := range allocs {
		ips[a.UserID] = a.IP
		assert.Equal(t, "net-1", a.NetworkID)
	}
	assert.Equal(t, "10.0.0.1", ips["user-1"])
	assert.Equal(t, "10.0.0.2", ips["user-2"])
}

func TestSQLiteIPAMRepository_ListAllocations_DifferentNetworks(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate in both networks
	_, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	_, err = repo.AllocateIP(ctx, "net-2", "user-2", "192.168.1.0/24")
	require.NoError(t, err)

	// List net-1 allocations
	allocs1, err := repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, allocs1, 1)
	assert.Equal(t, "user-1", allocs1[0].UserID)

	// List net-2 allocations
	allocs2, err := repo.ListAllocations(ctx, "net-2")
	require.NoError(t, err)
	assert.Len(t, allocs2, 1)
	assert.Equal(t, "user-2", allocs2[0].UserID)
}

func TestSQLiteIPAMRepository_GetOrAllocate_ExistingAllocation(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// First allocation
	alloc1, err := repo.GetOrAllocate(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc1.IP)

	// Second call should return same allocation
	alloc2, err := repo.GetOrAllocate(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, alloc1.IP, alloc2.IP)
}

func TestSQLiteIPAMRepository_GetOrAllocate_NewAllocation(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate for first user
	alloc1, err := repo.GetOrAllocate(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc1.IP)

	// Allocate for second user (should get new IP)
	alloc2, err := repo.GetOrAllocate(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", alloc2.IP)
	assert.NotEqual(t, alloc1.IP, alloc2.IP)
}

func TestSQLiteIPAMRepository_List(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate some IPs
	_, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	_, err = repo.AllocateIP(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)

	// List (alias for ListAllocations)
	allocs, err := repo.List(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, allocs, 2)
}

func TestSQLiteIPAMRepository_Release(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate an IP
	_, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)

	// Release (alias method, doesn't error on not found)
	err = repo.Release(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// List should be empty
	allocs, err := repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Empty(t, allocs)
}

func TestSQLiteIPAMRepository_Release_NotFound(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Release non-existent (Release doesn't error, unlike ReleaseIP)
	err := repo.Release(ctx, "net-1", "user-1")
	require.NoError(t, err)
}

func TestSQLiteIPAMRepository_AllocateAfterRelease(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// Allocate IPs for two users
	ip1, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", ip1)

	ip2, err := repo.AllocateIP(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", ip2)

	// Release first user's IP
	err = repo.ReleaseIP(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// Allocate for third user - should get 10.0.0.1 (the released one)
	ip3, err := repo.AllocateIP(ctx, "net-1", "user-3", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", ip3)
}

func TestSQLiteIPAMRepository_NotInitialized(t *testing.T) {
	// Create repository without db
	repo := NewSQLiteIPAM()

	ctx := context.Background()
	_, err := repo.AllocateIP(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestSQLiteIPAMRepository_FullWorkflow(t *testing.T) {
	_, repo, cleanup := setupSQLiteIPAMTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Initially no allocations
	allocs, err := repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Empty(t, allocs)

	// 2. GetOrAllocate for user-1
	alloc1, err := repo.GetOrAllocate(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc1.IP)
	assert.Equal(t, "net-1", alloc1.NetworkID)
	assert.Equal(t, "user-1", alloc1.UserID)

	// 3. GetOrAllocate for same user returns same IP
	alloc1Again, err := repo.GetOrAllocate(ctx, "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, alloc1.IP, alloc1Again.IP)

	// 4. Allocate for user-2
	alloc2, err := repo.GetOrAllocate(ctx, "net-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", alloc2.IP)

	// 5. List all allocations
	allocs, err = repo.List(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, allocs, 2)

	// 6. Release user-1's IP
	err = repo.Release(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// 7. Verify only user-2's allocation remains
	allocs, err = repo.ListAllocations(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, allocs, 1)
	assert.Equal(t, "user-2", allocs[0].UserID)

	// 8. Allocate for user-3 - should reuse released IP
	alloc3, err := repo.AllocateIP(ctx, "net-1", "user-3", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc3)
}
