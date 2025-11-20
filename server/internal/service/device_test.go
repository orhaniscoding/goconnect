package service

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceService_RegisterDevice(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("Success - register new device", func(t *testing.T) {
		req := &domain.RegisterDeviceRequest{
			Name:      "Work Laptop",
			Platform:  "windows",
			PubKey:    "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
			HostName:  "LAPTOP-123",
			OSVersion: "Windows 11",
			DaemonVer: "v1.0.0",
		}

		device, err := service.RegisterDevice(ctx, "user-123", "tenant-1", req)

		require.NoError(t, err)
		assert.NotEmpty(t, device.ID)
		assert.Equal(t, "user-123", device.UserID)
		assert.Equal(t, "tenant-1", device.TenantID)
		assert.Equal(t, "Work Laptop", device.Name)
		assert.Equal(t, "windows", device.Platform)
		assert.False(t, device.Active) // Not active until first heartbeat
	})

	t.Run("Forbidden - user not found", func(t *testing.T) {
		req := &domain.RegisterDeviceRequest{
			Name:     "Mobile Phone",
			Platform: "android",
			PubKey:   "aNothErKeY7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA=",
		}

		_, err := service.RegisterDevice(ctx, "non-existent", "tenant-1", req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrUserNotFound, domainErr.Code)
	})

	t.Run("Forbidden - tenant mismatch", func(t *testing.T) {
		req := &domain.RegisterDeviceRequest{
			Name:     "Tablet",
			Platform: "ios",
			PubKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		}

		_, err := service.RegisterDevice(ctx, "user-123", "wrong-tenant", req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
		assert.Contains(t, err.Error(), "Tenant mismatch")
	})

	t.Run("Conflict - duplicate public key", func(t *testing.T) {
		req1 := &domain.RegisterDeviceRequest{
			Name:     "Device 1",
			Platform: "linux",
			PubKey:   "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=",
		}

		_, err := service.RegisterDevice(ctx, "user-123", "tenant-1", req1)
		require.NoError(t, err)

		// Try to register another device with same pubkey
		req2 := &domain.RegisterDeviceRequest{
			Name:     "Device 2",
			Platform: "macos",
			PubKey:   "DuplicateKey7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pWX=",
		}

		_, err = service.RegisterDevice(ctx, "user-123", "tenant-1", req2)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrConflict, domainErr.Code)
	})

	t.Run("Validation - invalid platform", func(t *testing.T) {
		req := &domain.RegisterDeviceRequest{
			Name:     "Invalid Device",
			Platform: "playstation", // Invalid
			PubKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		}

		_, err := service.RegisterDevice(ctx, "user-123", "tenant-1", req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})
}

func TestDeviceService_GetDevice(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Register device
	req := &domain.RegisterDeviceRequest{
		Name:     "Test Device",
		Platform: "windows",
		PubKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	}
	device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", req)

	t.Run("Success - owner gets device", func(t *testing.T) {
		retrieved, err := service.GetDevice(ctx, device.ID, "user-123", "tenant-1", false)

		require.NoError(t, err)
		assert.Equal(t, device.ID, retrieved.ID)
		assert.Equal(t, "Test Device", retrieved.Name)
	})

	t.Run("Success - admin gets any device", func(t *testing.T) {
		retrieved, err := service.GetDevice(ctx, device.ID, "admin-456", "tenant-1", true)

		require.NoError(t, err)
		assert.Equal(t, device.ID, retrieved.ID)
	})

	t.Run("Forbidden - non-owner cannot access", func(t *testing.T) {
		_, err := service.GetDevice(ctx, device.ID, "other-user", "tenant-1", false)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("Not found - device doesn't exist", func(t *testing.T) {
		_, err := service.GetDevice(ctx, "non-existent-device", "user-123", "tenant-1", false)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	})
}

func TestDeviceService_ListDevices(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	// Create users
	user1 := &domain.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "user1@example.com",
		PassHash: "dummy",
	}
	user2 := &domain.User{
		ID:       "user-2",
		TenantID: "tenant-1",
		Email:    "user2@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user1))
	require.NoError(t, userRepo.Create(ctx, user2))

	// Register devices for user1
	service.RegisterDevice(ctx, "user-1", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "User1 Device1", Platform: "windows", PubKey: "U1D1KeyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	})
	service.RegisterDevice(ctx, "user-1", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "User1 Device2", Platform: "android", PubKey: "U1D2KeyBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
	})

	// Register device for user2
	service.RegisterDevice(ctx, "user-2", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "User2 Device1", Platform: "ios", PubKey: "U2D1KeyCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=",
	})

	t.Run("Success - user lists own devices", func(t *testing.T) {
		filter := domain.DeviceFilter{Limit: 50}
		devices, cursor, err := service.ListDevices(ctx, "user-1", "tenant-1", false, filter)

		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Empty(t, cursor)
	})

	t.Run("Success - admin lists all devices", func(t *testing.T) {
		filter := domain.DeviceFilter{Limit: 50}
		devices, _, err := service.ListDevices(ctx, "admin-user", "tenant-1", true, filter)

		require.NoError(t, err)
		assert.Len(t, devices, 3) // All devices
	})

	t.Run("Success - filter by platform", func(t *testing.T) {
		filter := domain.DeviceFilter{Platform: "windows", Limit: 50}
		devices, _, err := service.ListDevices(ctx, "user-1", "tenant-1", false, filter)

		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, "User1 Device1", devices[0].Name)
	})
}

func TestDeviceService_UpdateDevice(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Register device
	device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "Original Name", Platform: "windows", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	})

	t.Run("Success - owner updates device", func(t *testing.T) {
		newName := "Updated Name"
		req := &domain.UpdateDeviceRequest{Name: &newName}

		updated, err := service.UpdateDevice(ctx, device.ID, "user-123", "tenant-1", false, req)

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("Success - admin updates any device", func(t *testing.T) {
		newHostName := "ADMIN-CHANGED"
		req := &domain.UpdateDeviceRequest{HostName: &newHostName}

		updated, err := service.UpdateDevice(ctx, device.ID, "admin-456", "tenant-1", true, req)

		require.NoError(t, err)
		assert.Equal(t, "ADMIN-CHANGED", updated.HostName)
	})

	t.Run("Forbidden - non-owner cannot update", func(t *testing.T) {
		newName := "Hacked"
		req := &domain.UpdateDeviceRequest{Name: &newName}

		_, err := service.UpdateDevice(ctx, device.ID, "other-user", "tenant-1", false, req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("Conflict - duplicate pubkey on update", func(t *testing.T) {
		// First device is already registered (device variable from parent scope)
		// Register a second device with a different unique 44-char pubkey
		device2, err := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
			Name:     "Device2",
			Platform: "linux",
			PubKey:   "bOtHerKeyXYZ123456789ABCDEFGHIJKLMNOPQRSTUV=", // Exactly 44 chars
		})
		require.NoError(t, err)
		require.NotNil(t, device2)

		// Try to update first device to use second device's pubkey (should fail with conflict)
		existingPubKey := device2.PubKey
		req := &domain.UpdateDeviceRequest{PubKey: &existingPubKey}

		_, err = service.UpdateDevice(ctx, device.ID, "user-123", "tenant-1", false, req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrConflict, domainErr.Code)
	})
}

func TestDeviceService_DeleteDevice(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("Success - owner deletes device", func(t *testing.T) {
		device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
			Name: "To Delete", Platform: "windows", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		})

		err := service.DeleteDevice(ctx, device.ID, "user-123", "tenant-1", false)

		require.NoError(t, err)

		// Verify deleted
		_, err = service.GetDevice(ctx, device.ID, "user-123", "tenant-1", false)
		require.Error(t, err)
	})

	t.Run("Success - admin deletes any device", func(t *testing.T) {
		device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
			Name: "Admin Delete", Platform: "linux", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		})

		err := service.DeleteDevice(ctx, device.ID, "admin-456", "tenant-1", true)

		require.NoError(t, err)
	})

	t.Run("Forbidden - non-owner cannot delete", func(t *testing.T) {
		device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
			Name: "Protected", Platform: "macos", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		})

		err := service.DeleteDevice(ctx, device.ID, "other-user", "tenant-1", false)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})
}

func TestDeviceService_Heartbeat(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Register device
	device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "Heartbeat Test", Platform: "windows", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	})

	t.Run("Success - owner sends heartbeat", func(t *testing.T) {
		req := &domain.DeviceHeartbeatRequest{
			IPAddress: "192.168.1.100",
			DaemonVer: "v1.2.0",
		}

		err := service.Heartbeat(ctx, device.ID, "user-123", "tenant-1", req)

		require.NoError(t, err)

		// Verify device is now active
		updated, _ := service.GetDevice(ctx, device.ID, "user-123", "tenant-1", false)
		assert.True(t, updated.Active)
		assert.Equal(t, "192.168.1.100", updated.IPAddress)
		assert.Equal(t, "v1.2.0", updated.DaemonVer)
	})

	t.Run("Forbidden - non-owner cannot send heartbeat", func(t *testing.T) {
		req := &domain.DeviceHeartbeatRequest{IPAddress: "192.168.1.200"}

		err := service.Heartbeat(ctx, device.ID, "other-user", "tenant-1", req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("Forbidden - disabled device cannot heartbeat", func(t *testing.T) {
		// Disable device
		service.DisableDevice(ctx, device.ID, "user-123", "tenant-1", false)

		req := &domain.DeviceHeartbeatRequest{IPAddress: "192.168.1.100"}
		err := service.Heartbeat(ctx, device.ID, "user-123", "tenant-1", req)

		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
		assert.Contains(t, err.Error(), "disabled")
	})
}

func TestDeviceService_DisableEnable(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewDeviceService(deviceRepo, userRepo)

	ctx := context.Background()

	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PassHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Register device
	device, _ := service.RegisterDevice(ctx, "user-123", "tenant-1", &domain.RegisterDeviceRequest{
		Name: "Disable Test", Platform: "android", PubKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
	})

	t.Run("Success - disable then enable cycle", func(t *testing.T) {
		// Disable
		err := service.DisableDevice(ctx, device.ID, "user-123", "tenant-1", false)
		require.NoError(t, err)

		disabled, _ := service.GetDevice(ctx, device.ID, "user-123", "tenant-1", false)
		assert.True(t, disabled.IsDisabled())
		assert.False(t, disabled.Active)

		// Enable
		err = service.EnableDevice(ctx, device.ID, "user-123", "tenant-1", false)
		require.NoError(t, err)

		enabled, _ := service.GetDevice(ctx, device.ID, "user-123", "tenant-1", false)
		assert.False(t, enabled.IsDisabled())
	})

	t.Run("Forbidden - non-owner cannot disable/enable", func(t *testing.T) {
		err := service.DisableDevice(ctx, device.ID, "other-user", "tenant-1", false)
		require.Error(t, err)

		err = service.EnableDevice(ctx, device.ID, "other-user", "tenant-1", false)
		require.Error(t, err)
	})
}
