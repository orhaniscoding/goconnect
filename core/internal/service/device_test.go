package service

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceService_RegisterDevice(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:           "user-123",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "dummy",
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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	// Create users
	user1 := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "user1@example.com",
		PasswordHash: "dummy",
	}
	user2 := &domain.User{
		ID:           "user-2",
		TenantID:     "tenant-1",
		Email:        "user2@example.com",
		PasswordHash: "dummy",
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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	user := &domain.User{
		ID:           "user-123",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "dummy",
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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	user := &domain.User{
		ID:           "user-123",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "dummy",
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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	user := &domain.User{
		ID:           "user-123",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "dummy",
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
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	ctx := context.Background()

	user := &domain.User{
		ID:           "user-123",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "dummy",
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

// MockNotifier for testing
type MockNotifier struct {
	onlineCalls  []string
	offlineCalls []string
}

func (m *MockNotifier) DeviceOnline(deviceID, userID string) {
	m.onlineCalls = append(m.onlineCalls, deviceID)
}

func (m *MockNotifier) DeviceOffline(deviceID, userID string) {
	m.offlineCalls = append(m.offlineCalls, deviceID)
}

func TestDeviceService_OfflineDetection(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	mockNotifier := &MockNotifier{}
	service.SetNotifier(mockNotifier)

	ctx := context.Background()

	// Create user
	user := &domain.User{ID: "u1", TenantID: "t1"}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create devices
	// D1: Online and recent
	d1, err := service.RegisterDevice(ctx, "u1", "t1", &domain.RegisterDeviceRequest{
		Name: "D1", Platform: "linux", PubKey: "cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
	})
	require.NoError(t, err)
	d1.Active = true
	d1.LastSeen = time.Now()
	deviceRepo.Update(ctx, d1)

	// D2: Online but stale (should be detected)
	d2, err := service.RegisterDevice(ctx, "u1", "t1", &domain.RegisterDeviceRequest{
		Name: "D2", Platform: "linux", PubKey: "dOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ=",
	})
	require.NoError(t, err)
	d2.Active = true
	d2.LastSeen = time.Now().Add(-10 * time.Minute)
	deviceRepo.Update(ctx, d2)

	// Run detection
	service.detectOfflineDevices(ctx, 5*time.Minute)

	// Verify D2 is now inactive
	updatedD2, _ := deviceRepo.GetByID(ctx, d2.ID)
	assert.False(t, updatedD2.Active)

	// Verify D1 is still active
	updatedD1, _ := deviceRepo.GetByID(ctx, d1.ID)
	assert.True(t, updatedD1.Active)

	// Verify notification
	assert.Contains(t, mockNotifier.offlineCalls, d2.ID)
	assert.NotContains(t, mockNotifier.offlineCalls, d1.ID)
}

// ==================== SetAuditor Tests ====================

func TestDeviceService_SetAuditor(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}

	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	t.Run("Set Auditor With Nil", func(t *testing.T) {
		// Setting nil auditor should be a no-op (not panic)
		service.SetAuditor(nil)
	})

	t.Run("Set Auditor With Valid Auditor", func(t *testing.T) {
		mockAuditor := &mockAuditor{}
		service.SetAuditor(mockAuditor)
		// Should not panic and auditor should be set
	})
}

// ==================== SetPeerProvisioning Tests ====================

func TestDeviceService_SetPeerProvisioning(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}

	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	t.Run("Set PeerProvisioning With Nil", func(t *testing.T) {
		// Setting nil should work
		service.SetPeerProvisioning(nil)
	})

	t.Run("Set PeerProvisioning With Valid Service", func(t *testing.T) {
		membershipRepo := repository.NewInMemoryMembershipRepository()
		ipamRepo := repository.NewInMemoryIPAM()
		peerProvService := NewPeerProvisioningService(peerRepo, deviceRepo, networkRepo, membershipRepo, ipamRepo)
		service.SetPeerProvisioning(peerProvService)
		// Should not panic
	})
}

// Mock auditor for testing
type mockAuditor struct {
	calls []string
}

func (m *mockAuditor) Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
	m.calls = append(m.calls, action)
}

// ==================== GetDeviceConfig Tests ====================

func TestDeviceService_GetDeviceConfig(t *testing.T) {
	ctx := context.Background()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{
		DNS:            "10.0.0.1",
		ServerPubKey:   "testserverpubkey123456789012345678901234=",
		ServerEndpoint: "vpn.example.com:51820",
		Keepalive:      25,
	}

	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	// Create test user and device
	username := "testuser"
	user := &domain.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		Username: &username,
	}
	userRepo.Create(ctx, user)

	device := &domain.Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Test Device",
		PubKey:   "testpubkey123456789012345678901234567890=",
	}
	deviceRepo.Create(ctx, device)

	t.Run("Device Not Found", func(t *testing.T) {
		_, err := service.GetDeviceConfig(ctx, "non-existent", "user-1")
		assert.Error(t, err)
	})

	t.Run("Device Ownership Mismatch", func(t *testing.T) {
		_, err := service.GetDeviceConfig(ctx, "device-1", "other-user")
		assert.Error(t, err)
	})

	t.Run("No Peers Found", func(t *testing.T) {
		_, err := service.GetDeviceConfig(ctx, "device-1", "user-1")
		assert.Error(t, err)
	})

	t.Run("Success With Peers", func(t *testing.T) {
		// Create network
		dns := "8.8.8.8"
		mtu := 1400
		network := &domain.Network{
			ID:       "network-1",
			TenantID: "tenant-1",
			Name:     "Test Network",
			CIDR:     "10.1.0.0/24",
			DNS:      &dns,
			MTU:      &mtu,
		}
		networkRepo.Create(ctx, network)

		// Create peer for the device
		peer := &domain.Peer{
			ID:        "peer-1",
			DeviceID:  "device-1",
			NetworkID: "network-1",
			PublicKey: "testpubkey123456789012345678901234567890=",
			AllowedIPs: []string{"10.1.0.2/32"},
		}
		peerRepo.Create(ctx, peer)

		config, err := service.GetDeviceConfig(ctx, "device-1", "user-1")
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, []string{"10.1.0.2/32"}, config.Interface.Addresses)
		assert.Equal(t, 1400, config.Interface.MTU)
		assert.Contains(t, config.Interface.DNS, "8.8.8.8")
		assert.NotEmpty(t, config.Peers)
	})
}

// ==================== DetectOfflineDevices Tests ====================

func TestDeviceService_DetectOfflineDevices(t *testing.T) {
	ctx := context.Background()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}

	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	// Create test user
	username := "testuser"
	user := &domain.User{
		ID:       "user-1",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		Username: &username,
	}
	userRepo.Create(ctx, user)

	// Create a device with old last seen time
	oldTime := time.Now().Add(-1 * time.Hour)
	device := &domain.Device{
		ID:       "stale-device",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Stale Device",
		PubKey:   "stalepubkey12345678901234567890123456789=",
		LastSeen: oldTime,
		Active:   true,
	}
	deviceRepo.Create(ctx, device)

	// Set up mock notifier
	mockNotify := &MockNotifier{}
	service.SetNotifier(mockNotify)

	// Call detectOfflineDevices (private method, but we can test via StartOfflineDetection behavior)
	// Since detectOfflineDevices is private, we test via StartOfflineDetection with short context
	shortCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	go service.StartOfflineDetection(shortCtx, 50*time.Millisecond, 30*time.Minute)

	// Wait for context to cancel
	<-shortCtx.Done()

	// The device should have been detected (if GetStaleDevices is implemented)
	// This tests that StartOfflineDetection doesn't panic
}

// ==================== GetUpdatedFields Tests ====================

func TestGetUpdatedFields(t *testing.T) {
	t.Run("No Fields Updated", func(t *testing.T) {
		req := &domain.UpdateDeviceRequest{}
		fields := getUpdatedFields(req)
		assert.Empty(t, fields)
	})

	t.Run("All Fields Updated", func(t *testing.T) {
		name := "New Name"
		pubkey := "newpubkey1234567890123456789012345678901="
		hostname := "newhostname"
		osVersion := "1.0.0"
		daemonVer := "2.0.0"
		req := &domain.UpdateDeviceRequest{
			Name:      &name,
			PubKey:    &pubkey,
			HostName:  &hostname,
			OSVersion: &osVersion,
			DaemonVer: &daemonVer,
		}
		fields := getUpdatedFields(req)
		assert.Len(t, fields, 5)
		assert.Contains(t, fields, "name")
		assert.Contains(t, fields, "pubkey")
		assert.Contains(t, fields, "hostname")
		assert.Contains(t, fields, "os_version")
		assert.Contains(t, fields, "daemon_ver")
	})

	t.Run("Some Fields Updated", func(t *testing.T) {
		name := "New Name"
		req := &domain.UpdateDeviceRequest{
			Name: &name,
		}
		fields := getUpdatedFields(req)
		assert.Len(t, fields, 1)
		assert.Contains(t, fields, "name")
	})
}

// ==================== DeviceOnline/DeviceOffline Notifier Tests ====================

func TestDeviceService_HeartbeatTriggersDeviceOnline(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	mockNotifier := &MockNotifier{}
	service.SetNotifier(mockNotifier)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		ID:           "user-notify",
		TenantID:     "tenant-1",
		Email:        "notify@example.com",
		PasswordHash: "dummy",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Register device
	device, err := service.RegisterDevice(ctx, "user-notify", "tenant-1", &domain.RegisterDeviceRequest{
		Name:     "Notify Test Device",
		Platform: "linux",
		PubKey:   "NotifyKeyABCDEFGHIJKLMNOPQRSTUVWXYZ12345678=",
	})
	require.NoError(t, err)

	t.Run("Heartbeat triggers DeviceOnline notification", func(t *testing.T) {
		// Clear previous calls
		mockNotifier.onlineCalls = nil

		req := &domain.DeviceHeartbeatRequest{
			IPAddress: "192.168.1.100",
		}

		err := service.Heartbeat(ctx, device.ID, "user-notify", "tenant-1", req)
		require.NoError(t, err)

		// Verify DeviceOnline was called
		assert.Contains(t, mockNotifier.onlineCalls, device.ID)
	})

	t.Run("Heartbeat from offline device triggers DeviceOnline", func(t *testing.T) {
		// Set device as offline (no recent LastSeen)
		device.LastSeen = time.Time{} // Zero time means never seen
		deviceRepo.Update(ctx, device)

		// Clear previous calls
		mockNotifier.onlineCalls = nil

		req := &domain.DeviceHeartbeatRequest{
			IPAddress: "192.168.1.101",
		}

		err := service.Heartbeat(ctx, device.ID, "user-notify", "tenant-1", req)
		require.NoError(t, err)

		// Verify DeviceOnline was called
		assert.Contains(t, mockNotifier.onlineCalls, device.ID)
	})
}

func TestDeviceService_DetectOfflineDevicesTriggersDeviceOffline(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	mockNotifier := &MockNotifier{}
	service.SetNotifier(mockNotifier)

	ctx := context.Background()

	// Create user
	user := &domain.User{ID: "user-offline", TenantID: "tenant-1"}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create a device with stale last seen time
	device, err := service.RegisterDevice(ctx, "user-offline", "tenant-1", &domain.RegisterDeviceRequest{
		Name:     "Stale Device",
		Platform: "linux",
		PubKey:   "StaleKeyXABCDEFGHIJKLMNOPQRSTUVWXYZ12345678=",
	})
	require.NoError(t, err)

	// Mark as active but stale
	device.Active = true
	device.LastSeen = time.Now().Add(-10 * time.Minute) // Stale
	deviceRepo.Update(ctx, device)

	t.Run("Detect offline triggers DeviceOffline notification", func(t *testing.T) {
		// Clear previous calls
		mockNotifier.offlineCalls = nil

		// Run offline detection with 5 minute threshold
		service.detectOfflineDevices(ctx, 5*time.Minute)

		// Verify DeviceOffline was called for stale device
		assert.Contains(t, mockNotifier.offlineCalls, device.ID)
	})
}

func TestNoopDeviceNotifier(t *testing.T) {
	// Test that noopDeviceNotifier doesn't panic
	notifier := noopDeviceNotifier{}

	t.Run("DeviceOnline does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			notifier.DeviceOnline("device-id", "user-id")
		})
	})

	t.Run("DeviceOffline does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			notifier.DeviceOffline("device-id", "user-id")
		})
	})
}

// ==================== SetNotifier Tests ====================

func TestDeviceService_SetNotifier(t *testing.T) {
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	wgConfig := config.WireGuardConfig{}
	service := NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	t.Run("Set Notifier With Nil", func(t *testing.T) {
		// Setting nil notifier should be a no-op (not panic)
		service.SetNotifier(nil)
	})

	t.Run("Set Notifier With Valid Notifier", func(t *testing.T) {
		mockNotifier := &MockNotifier{}
		service.SetNotifier(mockNotifier)
		// Should not panic and notifier should be set
	})
}
