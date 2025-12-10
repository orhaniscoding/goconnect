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

func setupTenantTestDB(t *testing.T) (*sql.DB, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))
	return db, func() { db.Close() }
}

func TestSQLiteTenantRepository_Create(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	tenant := &domain.Tenant{
		ID:          "tenant-1",
		Name:        "Test Tenant",
		Description: "A test tenant",
		OwnerID:     "owner-1",
		MaxMembers:  100,
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	// Duplicate should fail
	err = repo.Create(ctx, tenant)
	require.Error(t, err)
}

func TestSQLiteTenantRepository_GetByID(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Create tenant
	tenant := &domain.Tenant{
		ID:          "tenant-1",
		Name:        "Test Tenant",
		Description: "A test tenant",
		OwnerID:     "owner-1",
		MaxMembers:  100,
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Get existing
	got, err := repo.GetByID(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Tenant", got.Name)
	assert.Equal(t, domain.TenantVisibilityPublic, got.Visibility)

	// Get non-existent
	_, err = repo.GetByID(ctx, "nonexistent")
	require.Error(t, err)
}

func TestSQLiteTenantRepository_Update(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Create tenant
	tenant := &domain.Tenant{
		ID:          "tenant-1",
		Name:        "Original Name",
		Description: "Original Description",
		OwnerID:     "owner-1",
		MaxMembers:  100,
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Update
	tenant.Name = "Updated Name"
	tenant.Description = "Updated Description"
	tenant.MaxMembers = 200
	tenant.Visibility = domain.TenantVisibilityPrivate
	tenant.UpdatedAt = time.Now()

	err := repo.Update(ctx, tenant)
	require.NoError(t, err)

	// Verify
	got, err := repo.GetByID(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", got.Name)
	assert.Equal(t, "Updated Description", got.Description)
	assert.Equal(t, 200, got.MaxMembers)
	assert.Equal(t, domain.TenantVisibilityPrivate, got.Visibility)
}

func TestSQLiteTenantRepository_Delete(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Create tenant
	tenant := &domain.Tenant{
		ID:          "tenant-1",
		Name:        "Test Tenant",
		OwnerID:     "owner-1",
		MaxMembers:  100,
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Delete
	err := repo.Delete(ctx, "tenant-1")
	require.NoError(t, err)

	// Should not exist
	_, err = repo.GetByID(ctx, "tenant-1")
	require.Error(t, err)

	// Delete non-existent - should error (tenant not found)
	err = repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
}

func TestSQLiteTenantRepository_List(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Create multiple tenants
	for i := 0; i < 5; i++ {
		tenant := &domain.Tenant{
			ID:          "tenant-" + string(rune('a'+i)),
			Name:        "Tenant " + string(rune('A'+i)),
			OwnerID:     "owner-1",
			MaxMembers:  100,
			Visibility:  domain.TenantVisibilityPublic,
			AccessType:  domain.TenantAccessOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, tenant))
	}

	// List all
	tenants, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, tenants, 5)
}

func TestSQLiteTenantRepository_ListAll(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Create multiple tenants
	for i := 0; i < 5; i++ {
		tenant := &domain.Tenant{
			ID:          "tenant-" + string(rune('a'+i)),
			Name:        "Tenant " + string(rune('A'+i)),
			OwnerID:     "owner-1",
			MaxMembers:  100,
			Visibility:  domain.TenantVisibilityPublic,
			AccessType:  domain.TenantAccessOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, tenant))
	}

	// List all with limit (offset pagination)
	tenants, total, err := repo.ListAll(ctx, 3, 0, "")
	require.NoError(t, err)
	assert.Len(t, tenants, 3)
	assert.Equal(t, 5, total)

	// List next page
	tenants2, total2, err := repo.ListAll(ctx, 3, 3, "")
	require.NoError(t, err)
	assert.Len(t, tenants2, 2)
	assert.Equal(t, 5, total2)

	// List with search query
	tenantsFiltered, _, err := repo.ListAll(ctx, 10, 0, "Tenant A")
	require.NoError(t, err)
	assert.Len(t, tenantsFiltered, 1)
}

func TestSQLiteTenantRepository_Count(t *testing.T) {
	db, cleanup := setupTenantTestDB(t)
	defer cleanup()

	repo := NewSQLiteTenantRepository(db)
	ctx := context.Background()

	// Initial count should be 0
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create tenants
	for i := 0; i < 3; i++ {
		tenant := &domain.Tenant{
			ID:          "tenant-" + string(rune('a'+i)),
			Name:        "Tenant " + string(rune('A'+i)),
			OwnerID:     "owner-1",
			MaxMembers:  100,
			Visibility:  domain.TenantVisibilityPublic,
			AccessType:  domain.TenantAccessOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, tenant))
	}

	// Count should be 3
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}
