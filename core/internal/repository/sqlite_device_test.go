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

func setupSQLiteDeviceTest(t *testing.T) (*SQLiteDeviceRepository, func()) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "devices.db")
	db, err := database.ConnectSQLite(dbPath)
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	// Seed tenant and user
	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, created_at, updated_at) VALUES ('user-1', 'tenant-1', 'test@example.com', 'hash', 'en', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	return NewSQLiteDeviceRepository(db), func() { db.Close() }
}

func TestSQLiteDeviceRepository_Create(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create device
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// Verify creation
	got, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Device", got.Name)
	assert.Equal(t, "linux", got.Platform)
	assert.True(t, got.Active)
}

func TestSQLiteDeviceRepository_Create_DuplicatePubKey(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device1 := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device 1",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device1))

	// Try to create with same pubkey
	device2 := &domain.Device{
		ID:        "device-2",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device 2",
		Platform:  "linux",
		PubKey:    "test-pubkey-1", // Same pubkey
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, device2)
	assert.Error(t, err)
}

func TestSQLiteDeviceRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	_, err := repo.GetByID(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSQLiteDeviceRepository_GetByPubKey(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "unique-pubkey-123",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device))

	got, err := repo.GetByPubKey(ctx, "unique-pubkey-123")
	require.NoError(t, err)
	assert.Equal(t, "device-1", got.ID)
}

func TestSQLiteDeviceRepository_List(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create devices
	for i := 1; i <= 5; i++ {
		active := i%2 == 0
		device := &domain.Device{
			ID:        stringIDDevice("device", i),
			UserID:    "user-1",
			TenantID:  "tenant-1",
			Name:      stringIDDevice("Device", i),
			Platform:  "linux",
			PubKey:    stringIDDevice("pubkey", i),
			Active:    active,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, device))
	}

	// List all
	devices, _, err := repo.List(ctx, domain.DeviceFilter{UserID: "user-1", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, devices, 5)

	// List active only
	activeTrue := true
	devices, _, err = repo.List(ctx, domain.DeviceFilter{UserID: "user-1", Active: &activeTrue, Limit: 10})
	require.NoError(t, err)
	assert.Len(t, devices, 2)

	// List with limit
	devices, _, err = repo.List(ctx, domain.DeviceFilter{UserID: "user-1", Limit: 2})
	require.NoError(t, err)
	assert.Len(t, devices, 2)
}

func TestSQLiteDeviceRepository_Update(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device))

	// Update device
	device.Name = "Updated Device"
	device.Active = false
	device.IPAddress = "10.0.0.1"
	device.DaemonVer = "1.2.3"

	require.NoError(t, repo.Update(ctx, device))

	// Verify update
	got, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Device", got.Name)
	assert.False(t, got.Active)
	assert.Equal(t, "10.0.0.1", got.IPAddress)
	assert.Equal(t, "1.2.3", got.DaemonVer)
}

func TestSQLiteDeviceRepository_Delete(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device))

	// Delete device
	require.NoError(t, repo.Delete(ctx, "device-1"))

	// Verify deletion
	_, err := repo.GetByID(ctx, "device-1")
	assert.Error(t, err)
}

func TestSQLiteDeviceRepository_UpdateHeartbeat(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device))

	// Update heartbeat with IP address
	require.NoError(t, repo.UpdateHeartbeat(ctx, "device-1", "10.0.0.1"))

	// Verify heartbeat update
	got, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.True(t, got.Active)
	assert.Equal(t, "10.0.0.1", got.IPAddress)
}

func TestSQLiteDeviceRepository_MarkInactive(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Test Device",
		Platform:  "linux",
		PubKey:    "test-pubkey-1",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, device))

	// Mark inactive
	require.NoError(t, repo.MarkInactive(ctx, "device-1"))

	// Verify
	got, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.False(t, got.Active)
}

func TestSQLiteDeviceRepository_Count(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	// Empty count
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create devices
	for i := 1; i <= 3; i++ {
		device := &domain.Device{
			ID:        stringIDDevice("device", i),
			UserID:    "user-1",
			TenantID:  "tenant-1",
			Name:      stringIDDevice("Device", i),
			Platform:  "linux",
			PubKey:    stringIDDevice("pubkey", i),
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, device))
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSQLiteDeviceRepository_GetStaleDevices(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create stale device (old last_seen)
	staleDevice := &domain.Device{
		ID:        "device-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Stale Device",
		Platform:  "linux",
		PubKey:    "pubkey-1",
		Active:    true,
		LastSeen:  time.Now().Add(-2 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, staleDevice))

	// Create active device
	activeDevice := &domain.Device{
		ID:        "device-2",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Name:      "Active Device",
		Platform:  "linux",
		PubKey:    "pubkey-2",
		Active:    true,
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, activeDevice))

	// Get stale devices (older than 1 hour)
	threshold := 1 * time.Hour
	staleDevices, err := repo.GetStaleDevices(ctx, threshold)
	require.NoError(t, err)
	assert.Len(t, staleDevices, 1)
	assert.Equal(t, "device-1", staleDevices[0].ID)
}

func stringIDDevice(prefix string, i int) string {
	return prefix + "-" + string(rune('0'+i))
}

// TestSQLiteDeviceRepository_List_Filters tests various filter combinations
func TestSQLiteDeviceRepository_List_Filters(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create devices with different attributes
	devices := []*domain.Device{
		{ID: "d1", UserID: "user-1", TenantID: "tenant-1", Name: "Device 1", Platform: "linux", PubKey: "pk1", Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "d2", UserID: "user-1", TenantID: "tenant-1", Name: "Device 2", Platform: "windows", PubKey: "pk2", Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "d3", UserID: "user-2", TenantID: "tenant-1", Name: "Device 3", Platform: "linux", PubKey: "pk3", Active: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "d4", UserID: "user-2", TenantID: "tenant-2", Name: "Device 4", Platform: "darwin", PubKey: "pk4", Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, d := range devices {
		require.NoError(t, repo.Create(ctx, d))
	}

	t.Run("Filter by TenantID", func(t *testing.T) {
		result, _, err := repo.List(ctx, domain.DeviceFilter{TenantID: "tenant-1", Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Filter by UserID", func(t *testing.T) {
		result, _, err := repo.List(ctx, domain.DeviceFilter{UserID: "user-2", Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Filter by Active", func(t *testing.T) {
		active := true
		result, _, err := repo.List(ctx, domain.DeviceFilter{Active: &active, Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Filter by Platform", func(t *testing.T) {
		result, _, err := repo.List(ctx, domain.DeviceFilter{Platform: "linux", Limit: 10})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Pagination with Cursor", func(t *testing.T) {
		result1, cursor, err := repo.List(ctx, domain.DeviceFilter{TenantID: "tenant-1", Limit: 2})
		require.NoError(t, err)
		assert.Len(t, result1, 2)
		assert.NotEmpty(t, cursor)

		result2, _, err := repo.List(ctx, domain.DeviceFilter{TenantID: "tenant-1", Limit: 2, Cursor: cursor})
		require.NoError(t, err)
		// Should have remaining items (1 or more based on sort order)
		assert.True(t, len(result2) >= 1)
	})
}

// TestSQLiteDeviceRepository_Update_NotFound tests updating non-existent device
func TestSQLiteDeviceRepository_Update_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	device := &domain.Device{
		ID:   "non-existent",
		Name: "Updated Name",
	}
	err := repo.Update(ctx, device)
	require.Error(t, err)
}

// TestSQLiteDeviceRepository_Delete_NotFound tests deleting non-existent device
func TestSQLiteDeviceRepository_Delete_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")
	require.Error(t, err)
}

// TestSQLiteDeviceRepository_UpdateHeartbeat_NotFound tests heartbeat update for non-existent device
func TestSQLiteDeviceRepository_UpdateHeartbeat_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	err := repo.UpdateHeartbeat(ctx, "non-existent", "192.168.1.1")
	require.Error(t, err)
}

// TestSQLiteDeviceRepository_MarkInactive_NotFound tests marking non-existent device inactive
func TestSQLiteDeviceRepository_MarkInactive_NotFound(t *testing.T) {
	repo, cleanup := setupSQLiteDeviceTest(t)
	defer cleanup()

	ctx := context.Background()

	err := repo.MarkInactive(ctx, "non-existent")
	require.Error(t, err)
}

