package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevice_Validate_Success(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     "My Laptop",
		Platform: "windows",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR", // 44 chars
	}

	err := device.Validate()
	assert.NoError(t, err)
}

func TestDevice_Validate_MissingUserID(t *testing.T) {
	device := &Device{
		TenantID: "tenant-456",
		Name:     "My Laptop",
		Platform: "windows",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestDevice_Validate_MissingTenantID(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		Name:     "My Laptop",
		Platform: "windows",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant_id is required")
}

func TestDevice_Validate_MissingName(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Platform: "windows",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestDevice_Validate_NameTooLong(t *testing.T) {
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}

	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     string(longName),
		Platform: "windows",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name exceeds maximum length")
}

func TestDevice_Validate_MissingPlatform(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     "My Laptop",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "platform is required")
}

func TestDevice_Validate_InvalidPlatform(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     "My Device",
		Platform: "invalid",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid platform")
}

func TestDevice_Validate_AllValidPlatforms(t *testing.T) {
	validPlatforms := []string{"windows", "macos", "linux", "android", "ios"}

	for _, platform := range validPlatforms {
		t.Run(platform, func(t *testing.T) {
			device := &Device{
				UserID:   "user-123",
				TenantID: "tenant-456",
				Name:     "Test Device",
				Platform: platform,
				PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
			}

			err := device.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestDevice_Validate_PlatformCaseInsensitive(t *testing.T) {
	// Platform validation should be case-insensitive
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     "Test Device",
		Platform: "Windows", // Uppercase
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
	}

	err := device.Validate()
	assert.NoError(t, err)
}

func TestDevice_Validate_MissingPubKey(t *testing.T) {
	device := &Device{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Name:     "My Laptop",
		Platform: "windows",
	}

	err := device.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pubkey is required")
}

func TestDevice_Validate_InvalidPubKeyLength(t *testing.T) {
	testCases := []struct {
		name   string
		pubKey string
	}{
		{"too short", "short"},
		{"too long", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR4567890"},
		{"43 chars", "abcdefghijklmnopqrstuvwxyz0123456789012="},
		{"45 chars", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR=="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			device := &Device{
				UserID:   "user-123",
				TenantID: "tenant-456",
				Name:     "My Laptop",
				Platform: "windows",
				PubKey:   tc.pubKey,
			}

			err := device.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid pubkey format")
		})
	}
}

func TestDevice_IsDisabled(t *testing.T) {
	// Not disabled
	device := &Device{
		DisabledAt: nil,
	}
	assert.False(t, device.IsDisabled())

	// Disabled
	now := time.Now()
	device.DisabledAt = &now
	assert.True(t, device.IsDisabled())
}

func TestDevice_Disable(t *testing.T) {
	device := &Device{
		Active:     true,
		DisabledAt: nil,
	}

	device.Disable()

	assert.NotNil(t, device.DisabledAt)
	assert.False(t, device.Active)
	assert.WithinDuration(t, time.Now(), device.UpdatedAt, 1*time.Second)
}

func TestDevice_Enable(t *testing.T) {
	now := time.Now()
	device := &Device{
		DisabledAt: &now,
	}

	device.Enable()

	assert.Nil(t, device.DisabledAt)
	assert.WithinDuration(t, time.Now(), device.UpdatedAt, 1*time.Second)
}

func TestDevice_UpdateHeartbeat(t *testing.T) {
	device := &Device{
		Active:    false,
		IPAddress: "1.2.3.4",
	}

	device.UpdateHeartbeat("5.6.7.8")

	assert.True(t, device.Active)
	assert.Equal(t, "5.6.7.8", device.IPAddress)
	assert.WithinDuration(t, time.Now(), device.LastSeen, 1*time.Second)
	assert.WithinDuration(t, time.Now(), device.UpdatedAt, 1*time.Second)
}

func TestDevice_UpdateHeartbeat_EmptyIP(t *testing.T) {
	device := &Device{
		Active:    false,
		IPAddress: "1.2.3.4",
	}

	device.UpdateHeartbeat("")

	assert.True(t, device.Active)
	assert.Equal(t, "1.2.3.4", device.IPAddress) // Should not change
	assert.WithinDuration(t, time.Now(), device.LastSeen, 1*time.Second)
}

func TestDevice_MarkInactive(t *testing.T) {
	device := &Device{
		Active: true,
	}

	device.MarkInactive()

	assert.False(t, device.Active)
	assert.WithinDuration(t, time.Now(), device.UpdatedAt, 1*time.Second)
}

func TestDeviceFilter_Defaults(t *testing.T) {
	filter := &DeviceFilter{
		UserID:   "user-123",
		TenantID: "tenant-456",
	}

	assert.Equal(t, "user-123", filter.UserID)
	assert.Equal(t, "tenant-456", filter.TenantID)
	assert.Equal(t, "", filter.Platform)
	assert.Nil(t, filter.Active)
}

func TestDeviceFilter_WithActive(t *testing.T) {
	activeTrue := true
	activeFalse := false

	filter1 := &DeviceFilter{
		Active: &activeTrue,
	}
	assert.True(t, *filter1.Active)

	filter2 := &DeviceFilter{
		Active: &activeFalse,
	}
	assert.False(t, *filter2.Active)
}

func TestRegisterDeviceRequest(t *testing.T) {
	req := &RegisterDeviceRequest{
		Name:      "My Device",
		Platform:  "windows",
		PubKey:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
		HostName:  "my-laptop",
		OSVersion: "Windows 11",
		DaemonVer: "1.0.0",
	}

	assert.Equal(t, "My Device", req.Name)
	assert.Equal(t, "windows", req.Platform)
	assert.Equal(t, 44, len(req.PubKey))
	assert.Equal(t, "my-laptop", req.HostName)
}

func TestUpdateDeviceRequest(t *testing.T) {
	newName := "Updated Name"
	newPubKey := "newkeyABCDEFGHIJKLMNOPQRSTUVWXYZ01234567"

	req := &UpdateDeviceRequest{
		Name:   &newName,
		PubKey: &newPubKey,
	}

	assert.NotNil(t, req.Name)
	assert.Equal(t, "Updated Name", *req.Name)
	assert.NotNil(t, req.PubKey)
	assert.Equal(t, newPubKey, *req.PubKey)
}

func TestDeviceHeartbeatRequest(t *testing.T) {
	req := &DeviceHeartbeatRequest{
		IPAddress: "10.0.0.5",
		DaemonVer: "1.2.3",
		OSVersion: "Ubuntu 22.04",
	}

	assert.Equal(t, "10.0.0.5", req.IPAddress)
	assert.Equal(t, "1.2.3", req.DaemonVer)
	assert.Equal(t, "Ubuntu 22.04", req.OSVersion)
}

func TestDevice_FullLifecycle(t *testing.T) {
	// Create device
	device := &Device{
		ID:       "device-1",
		UserID:   "user-1",
		TenantID: "tenant-1",
		Name:     "Test Device",
		Platform: "linux",
		PubKey:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQR",
		Active:   false,
	}

	// Validate
	err := device.Validate()
	require.NoError(t, err)

	// Heartbeat - becomes active
	device.UpdateHeartbeat("10.0.0.1")
	assert.True(t, device.Active)
	assert.Equal(t, "10.0.0.1", device.IPAddress)

	// Disable device
	device.Disable()
	assert.True(t, device.IsDisabled())
	assert.False(t, device.Active)

	// Re-enable device
	device.Enable()
	assert.False(t, device.IsDisabled())

	// Mark inactive
	device.MarkInactive()
	assert.False(t, device.Active)
}
