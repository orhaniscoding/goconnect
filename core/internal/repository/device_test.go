package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test device
func mkDevice(id, userID, name, pubkey, platform string) *domain.Device {
	now := time.Now()
	return &domain.Device{
		ID:        id,
		UserID:    userID,
		TenantID:  "tenant-1",
		Name:      name,
		Platform:  platform,
		PubKey:    pubkey,
		Active:    false,
		LastSeen:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewInMemoryDeviceRepository(t *testing.T) {
	repo := NewInMemoryDeviceRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.devices)
	assert.NotNil(t, repo.pubkeys)
	assert.Equal(t, 0, len(repo.devices))
	assert.Equal(t, 0, len(repo.pubkeys))
}

func TestDeviceRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device := mkDevice("device-1", "user-1", "Work Laptop", "pubkey123", "linux")

	err := repo.Create(ctx, device)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.devices))
	assert.Equal(t, 1, len(repo.pubkeys))
	assert.Equal(t, "device-1", repo.pubkeys["pubkey123"])
}

func TestDeviceRepository_Create_GeneratesID(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device := mkDevice("", "user-1", "Work Laptop", "pubkey123", "linux")

	err := repo.Create(ctx, device)

	require.NoError(t, err)
	assert.NotEmpty(t, device.ID)
}

func TestDeviceRepository_Create_DuplicatePubKey(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device1 := mkDevice("device-1", "user-1", "Device 1", "pubkey-duplicate", "linux")
	device2 := mkDevice("device-2", "user-2", "Device 2", "pubkey-duplicate", "windows")

	err1 := repo.Create(ctx, device1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, device2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrConflict, domainErr.Code)
	assert.Contains(t, domainErr.Message, "public key already exists")

	// Only first device should exist
	assert.Equal(t, 1, len(repo.devices))
}

func TestDeviceRepository_Create_MultipleDevices(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	devices := []*domain.Device{
		mkDevice("device-1", "user-1", "Laptop", "pubkey1", "linux"),
		mkDevice("device-2", "user-1", "Phone", "pubkey2", "android"),
		mkDevice("device-3", "user-2", "Desktop", "pubkey3", "windows"),
	}

	for _, device := range devices {
		err := repo.Create(ctx, device)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.devices))
	assert.Equal(t, 3, len(repo.pubkeys))
}

func TestDeviceRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	original := mkDevice("device-1", "user-1", "Work Laptop", "pubkey123", "linux")
	repo.Create(ctx, original)

	retrieved, err := repo.GetByID(ctx, "device-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Name, retrieved.Name)
	assert.Equal(t, original.PubKey, retrieved.PubKey)
	assert.Equal(t, original.Platform, retrieved.Platform)
}

func TestDeviceRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	retrieved, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestDeviceRepository_GetByPubKey_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	original := mkDevice("device-1", "user-1", "Work Laptop", "pubkey123", "linux")
	repo.Create(ctx, original)

	retrieved, err := repo.GetByPubKey(ctx, "pubkey123")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "device-1", retrieved.ID)
	assert.Equal(t, "pubkey123", retrieved.PubKey)
}

func TestDeviceRepository_GetByPubKey_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	retrieved, err := repo.GetByPubKey(ctx, "non-existent-pubkey")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestDeviceRepository_List_AllDevices(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	// Create multiple devices
	devices := []*domain.Device{
		mkDevice("device-1", "user-1", "Laptop", "pubkey1", "linux"),
		mkDevice("device-2", "user-1", "Phone", "pubkey2", "android"),
		mkDevice("device-3", "user-2", "Desktop", "pubkey3", "windows"),
	}

	for _, device := range devices {
		repo.Create(ctx, device)
	}

	filter := domain.DeviceFilter{Limit: 50}
	result, cursor, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Empty(t, cursor) // No next page
}

func TestDeviceRepository_List_ByUserID(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	repo.Create(ctx, mkDevice("device-1", "user-1", "Laptop", "pubkey1", "linux"))
	repo.Create(ctx, mkDevice("device-2", "user-1", "Phone", "pubkey2", "android"))
	repo.Create(ctx, mkDevice("device-3", "user-2", "Desktop", "pubkey3", "windows"))

	filter := domain.DeviceFilter{UserID: "user-1", Limit: 50}
	result, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	for _, device := range result {
		assert.Equal(t, "user-1", device.UserID)
	}
}

func TestDeviceRepository_List_ByPlatform(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	repo.Create(ctx, mkDevice("device-1", "user-1", "Laptop1", "pubkey1", "linux"))
	repo.Create(ctx, mkDevice("device-2", "user-1", "Laptop2", "pubkey2", "linux"))
	repo.Create(ctx, mkDevice("device-3", "user-2", "Desktop", "pubkey3", "windows"))

	filter := domain.DeviceFilter{Platform: "linux", Limit: 50}
	result, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	for _, device := range result {
		assert.Equal(t, "linux", device.Platform)
	}
}

func TestDeviceRepository_List_WithPagination(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	// Create 5 devices with unique pubkeys
	for i := 1; i <= 5; i++ {
		device := mkDevice("", "user-1", "Device", "pubkey-"+string(rune('a'+i)), "linux")
		repo.Create(ctx, device)
		time.Sleep(1 * time.Millisecond) // Ensure different CreatedAt times for sorting
	}

	// Get first page with limit 2
	filter := domain.DeviceFilter{Limit: 2}
	page1, cursor1, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, page1, 2)
	assert.NotEmpty(t, cursor1)

	// Get second page
	filter.Cursor = cursor1
	page2, cursor2, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, page2, 2)
	assert.NotEmpty(t, cursor2)

	// Devices should be different
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestDeviceRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device := mkDevice("device-1", "user-1", "Original Name", "pubkey123", "linux")
	repo.Create(ctx, device)

	// Create updated device
	updatedDevice := mkDevice("device-1", "user-1", "Updated Name", "pubkey123", "linux")
	updatedDevice.Active = true

	err := repo.Update(ctx, updatedDevice)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "device-1")
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.True(t, retrieved.Active)
}

func TestDeviceRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device := mkDevice("non-existent", "user-1", "Test", "pubkey123", "linux")

	err := repo.Update(ctx, device)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestDeviceRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	repo.Create(ctx, device)

	err := repo.Delete(ctx, "device-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.devices))
	assert.Equal(t, 0, len(repo.pubkeys))

	_, err = repo.GetByID(ctx, "device-1")
	assert.Error(t, err)

	_, err = repo.GetByPubKey(ctx, "pubkey123")
	assert.Error(t, err)
}

func TestDeviceRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestDeviceRepository_DifferentPlatforms(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	platforms := []string{"linux", "windows", "macos", "android", "ios"}

	for i, platform := range platforms {
		device := mkDevice("", "user-1", "Device", "pubkey"+string(rune(i)), platform)
		err := repo.Create(ctx, device)
		require.NoError(t, err)
	}

	assert.Equal(t, 5, len(repo.devices))
}

func TestDeviceRepository_FullCRUDCycle(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	// Create
	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// Read by ID
	retrieved, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Device", retrieved.Name)

	// Read by PubKey
	retrieved, err = repo.GetByPubKey(ctx, "pubkey123")
	require.NoError(t, err)
	assert.Equal(t, "device-1", retrieved.ID)

	// Update
	updatedDevice := mkDevice("device-1", "user-1", "Updated Device", "pubkey123", "linux")
	err = repo.Update(ctx, updatedDevice)
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(ctx, "device-1")
	assert.Equal(t, "Updated Device", retrieved.Name)

	// Delete
	err = repo.Delete(ctx, "device-1")
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, "device-1")
	assert.Error(t, err)
}

func TestDeviceRepository_UpdateHeartbeat_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// Update heartbeat
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.100")
	require.NoError(t, err)

	// Verify device was updated
	updated, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.Equal(t, "192.168.1.100", updated.IPAddress)
	assert.True(t, updated.Active)
	assert.False(t, updated.LastSeen.IsZero())
}

func TestDeviceRepository_UpdateHeartbeat_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.UpdateHeartbeat(ctx, "non-existent", "192.168.1.100")
	assert.Error(t, err)

	derr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, derr.Code)
}

func TestDeviceRepository_UpdateHeartbeat_MultipleUpdates(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// First heartbeat
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.100")
	require.NoError(t, err)

	device1, _ := repo.GetByID(ctx, "device-1")
	firstLastSeen := device1.LastSeen

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second heartbeat with different IP
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.101")
	require.NoError(t, err)

	device2, _ := repo.GetByID(ctx, "device-1")
	assert.Equal(t, "192.168.1.101", device2.IPAddress)
	assert.True(t, device2.LastSeen.After(firstLastSeen))
	assert.True(t, device2.Active)
}

func TestDeviceRepository_MarkInactive_Success(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// Make device active first
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.100")
	require.NoError(t, err)

	// Mark as inactive
	err = repo.MarkInactive(ctx, "device-1")
	require.NoError(t, err)

	// Verify device is inactive
	updated, err := repo.GetByID(ctx, "device-1")
	require.NoError(t, err)
	assert.False(t, updated.Active)
}

func TestDeviceRepository_MarkInactive_NotFound(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.MarkInactive(ctx, "non-existent")
	assert.Error(t, err)

	derr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, derr.Code)
}

func TestDeviceRepository_HeartbeatCycle(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	device := mkDevice("device-1", "user-1", "Test Device", "pubkey123", "linux")
	err := repo.Create(ctx, device)
	require.NoError(t, err)

	// Initially inactive (no heartbeat yet)
	d, _ := repo.GetByID(ctx, "device-1")
	assert.False(t, d.Active)

	// Send heartbeat -> active
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.100")
	require.NoError(t, err)
	d, _ = repo.GetByID(ctx, "device-1")
	assert.True(t, d.Active)

	// Mark inactive
	err = repo.MarkInactive(ctx, "device-1")
	require.NoError(t, err)
	d, _ = repo.GetByID(ctx, "device-1")
	assert.False(t, d.Active)

	// Send another heartbeat -> active again
	err = repo.UpdateHeartbeat(ctx, "device-1", "192.168.1.101")
	require.NoError(t, err)
	d, _ = repo.GetByID(ctx, "device-1")
	assert.True(t, d.Active)
	assert.Equal(t, "192.168.1.101", d.IPAddress)
}

func TestDeviceRepository_List_Search(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	devices := []*domain.Device{
		mkDevice("d1", "u1", "Alpha Device", "pk1", "linux"),
		mkDevice("d2", "u1", "Beta Device", "pk2", "windows"),
		mkDevice("d3", "u1", "Gamma Device", "pk3", "macos"),
	}
	// Set hostname for one device to test hostname search
	devices[1].HostName = "beta-host"

	for _, d := range devices {
		repo.Create(ctx, d)
	}

	// Test search by name "Alpha"
	results, _, err := repo.List(ctx, domain.DeviceFilter{
		TenantID: "tenant-1",
		Search:   "Alpha",
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Alpha Device", results[0].Name)

	// Test search by hostname "beta-host"
	results, _, err = repo.List(ctx, domain.DeviceFilter{
		TenantID: "tenant-1",
		Search:   "beta-host",
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Beta Device", results[0].Name)

	// Test case insensitive
	results, _, err = repo.List(ctx, domain.DeviceFilter{
		TenantID: "tenant-1",
		Search:   "GAMMA",
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Gamma Device", results[0].Name)

	// Test empty search
	results, _, err = repo.List(ctx, domain.DeviceFilter{
		TenantID: "tenant-1",
		Search:   "",
	})
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestDeviceRepository_GetStaleDevices(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()
	now := time.Now()

	// Device 1: Active and recent (Online)
	d1 := mkDevice("d1", "u1", "D1", "pk1", "linux")
	d1.Active = true
	d1.LastSeen = now
	repo.Create(ctx, d1)

	// Device 2: Active but old (Stale -> Offline)
	d2 := mkDevice("d2", "u1", "D2", "pk2", "linux")
	d2.Active = true
	d2.LastSeen = now.Add(-10 * time.Minute)
	repo.Create(ctx, d2)

	// Device 3: Inactive (Already Offline)
	d3 := mkDevice("d3", "u1", "D3", "pk3", "linux")
	d3.Active = false
	d3.LastSeen = now.Add(-10 * time.Minute)
	repo.Create(ctx, d3)

	// Device 4: Disabled
	d4 := mkDevice("d4", "u1", "D4", "pk4", "linux")
	d4.Active = true
	d4.LastSeen = now.Add(-10 * time.Minute)
	disabledAt := now
	d4.DisabledAt = &disabledAt
	repo.Create(ctx, d4)

	// Check for stale devices (threshold 5 mins)
	stale, err := repo.GetStaleDevices(ctx, 5*time.Minute)
	require.NoError(t, err)
	assert.Len(t, stale, 1)
	assert.Equal(t, "d2", stale[0].ID)
}

func TestDeviceRepository_Count(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	// Initially empty
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Add some devices
	d1 := mkDevice("d1", "u1", "Device 1", "pk1", "linux")
	d2 := mkDevice("d2", "u1", "Device 2", "pk2", "windows")
	d3 := mkDevice("d3", "u2", "Device 3", "pk3", "darwin")

	repo.Create(ctx, d1)
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	repo.Create(ctx, d2)
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	repo.Create(ctx, d3)
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Delete a device
	repo.Delete(ctx, "d1")
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}
