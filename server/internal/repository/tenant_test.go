package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test tenant
func mkTenant(id, name, ownerID string) *domain.Tenant {
	now := time.Now()
	return &domain.Tenant{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewInMemoryTenantRepository(t *testing.T) {
	repo := NewInMemoryTenantRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.tenants)
	assert.Equal(t, 0, len(repo.tenants))
}

func TestTenantRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("tenant-1", "Test Company", "owner-1")

	err := repo.Create(context.Background(), tenant)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.tenants))
}

func TestTenantRepository_Create_MultipleTenants(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenants := []*domain.Tenant{
		mkTenant("tenant-1", "Company A", "owner-1"),
		mkTenant("tenant-2", "Company B", "owner-2"),
		mkTenant("tenant-3", "Company C", "owner-3"),
	}

	for _, tenant := range tenants {
		err := repo.Create(context.Background(), tenant)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.tenants))
}

func TestTenantRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	original := mkTenant("tenant-1", "Test Company", "owner-1")
	repo.Create(context.Background(), original)

	retrieved, err := repo.GetByID(context.Background(), "tenant-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Name, retrieved.Name)
	assert.Equal(t, original.OwnerID, retrieved.OwnerID)
}

func TestTenantRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryTenantRepository()

	retrieved, err := repo.GetByID(context.Background(), "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrTenantNotFound, domainErr.Code)
	assert.Contains(t, domainErr.Message, "Tenant not found")
}

func TestTenantRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("tenant-1", "Original Name", "owner-1")
	repo.Create(context.Background(), tenant)

	// Create new tenant object for update
	updatedTenant := mkTenant("tenant-1", "Updated Name", "owner-2")

	err := repo.Update(context.Background(), updatedTenant)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(context.Background(), "tenant-1")
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "owner-2", retrieved.OwnerID)
}

func TestTenantRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("non-existent", "Test", "owner-1")

	err := repo.Update(context.Background(), tenant)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrTenantNotFound, domainErr.Code)
}

func TestTenantRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("tenant-1", "Test Company", "owner-1")
	repo.Create(context.Background(), tenant)

	err := repo.Delete(context.Background(), "tenant-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.tenants))

	_, err = repo.GetByID(context.Background(), "tenant-1")
	assert.Error(t, err)
}

func TestTenantRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryTenantRepository()

	err := repo.Delete(context.Background(), "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrTenantNotFound, domainErr.Code)
}

func TestTenantRepository_FullCRUDCycle(t *testing.T) {
	repo := NewInMemoryTenantRepository()

	// Create
	tenant := mkTenant("tenant-1", "Test Company", "owner-1")
	err := repo.Create(context.Background(), tenant)
	require.NoError(t, err)

	// Read
	retrieved, err := repo.GetByID(context.Background(), "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Company", retrieved.Name)

	// Update
	tenant.Name = "Updated Company"
	err = repo.Update(context.Background(), tenant)
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(context.Background(), "tenant-1")
	assert.Equal(t, "Updated Company", retrieved.Name)

	// Delete
	err = repo.Delete(context.Background(), "tenant-1")
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), "tenant-1")
	assert.Error(t, err)
}

func TestTenantRepository_ConcurrentReadsSafe(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("tenant-1", "Test Company", "owner-1")
	repo.Create(context.Background(), tenant)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := repo.GetByID(context.Background(), "tenant-1")
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTenantRepository_TimestampsPreserved(t *testing.T) {
	repo := NewInMemoryTenantRepository()
	tenant := mkTenant("tenant-1", "Test Company", "owner-1")

	createdAt := tenant.CreatedAt
	updatedAt := tenant.UpdatedAt

	repo.Create(context.Background(), tenant)

	retrieved, _ := repo.GetByID(context.Background(), "tenant-1")
	assert.Equal(t, createdAt.Unix(), retrieved.CreatedAt.Unix())
	assert.Equal(t, updatedAt.Unix(), retrieved.UpdatedAt.Unix())
}
