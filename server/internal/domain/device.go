package domain

import (
	"strings"
	"time"
)

// Device represents a physical or virtual device
type Device struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	TenantID   string     `json:"tenant_id"`
	Name       string     `json:"name"`       // User-friendly name (e.g., "Work Laptop")
	Platform   string     `json:"platform"`   // "windows", "macos", "linux", "android", "ios"
	PubKey     string     `json:"pubkey"`     // WireGuard public key
	LastSeen   time.Time  `json:"last_seen"`  // Last heartbeat from daemon
	Active     bool       `json:"active"`     // Currently connected/active
	IPAddress  string     `json:"ip_address"` // Last known IP address
	DaemonVer  string     `json:"daemon_ver"` // Client daemon version
	OSVersion  string     `json:"os_version"` // Operating system version
	HostName   string     `json:"hostname"`   // Device hostname
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DisabledAt *time.Time `json:"disabled_at,omitempty"` // Soft disable
}

// RegisterDeviceRequest is the request to register a new device
type RegisterDeviceRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Platform  string `json:"platform" binding:"required,oneof=windows macos linux android ios"`
	PubKey    string `json:"pubkey" binding:"required,min=44,max=44"` // Base64 WireGuard key (32 bytes = 44 chars)
	HostName  string `json:"hostname,omitempty"`
	OSVersion string `json:"os_version,omitempty"`
	DaemonVer string `json:"daemon_ver,omitempty"`
}

// UpdateDeviceRequest is the request to update device information
type UpdateDeviceRequest struct {
	Name      *string `json:"name,omitempty"`
	PubKey    *string `json:"pubkey,omitempty"`
	HostName  *string `json:"hostname,omitempty"`
	OSVersion *string `json:"os_version,omitempty"`
	DaemonVer *string `json:"daemon_ver,omitempty"`
}

// DeviceHeartbeatRequest is the heartbeat from daemon
type DeviceHeartbeatRequest struct {
	IPAddress string `json:"ip_address,omitempty"`
	DaemonVer string `json:"daemon_ver,omitempty"`
	OSVersion string `json:"os_version,omitempty"`
}

// DeviceFilter represents filtering options for device queries
type DeviceFilter struct {
	UserID   string // Filter by user
	TenantID string // Filter by tenant
	Platform string // Filter by platform
	Active   *bool  // Filter by active status
	Limit    int    // Max results (default 50, max 100)
	Cursor   string // Pagination cursor
}

// Validate validates the device
func (d *Device) Validate() error {
	if d.UserID == "" {
		return NewError(ErrValidation, "user_id is required", nil)
	}

	if d.TenantID == "" {
		return NewError(ErrValidation, "tenant_id is required", nil)
	}

	if d.Name == "" {
		return NewError(ErrValidation, "name is required", nil)
	}

	if len(d.Name) > 100 {
		return NewError(ErrValidation, "name exceeds maximum length of 100 characters", nil)
	}

	if d.Platform == "" {
		return NewError(ErrValidation, "platform is required", nil)
	}

	validPlatforms := map[string]bool{
		"windows": true,
		"macos":   true,
		"linux":   true,
		"android": true,
		"ios":     true,
	}

	if !validPlatforms[strings.ToLower(d.Platform)] {
		return NewError(ErrValidation, "invalid platform", map[string]string{
			"platform":        d.Platform,
			"valid_platforms": "windows, macos, linux, android, ios",
		})
	}

	if d.PubKey == "" {
		return NewError(ErrValidation, "pubkey is required", nil)
	}

	// WireGuard public key is 32 bytes, base64 encoded = 44 characters
	if len(d.PubKey) != 44 {
		return NewError(ErrValidation, "invalid pubkey format (must be 44 characters)", map[string]string{
			"expected_length": "44",
			"actual_length":   string(rune(len(d.PubKey))),
		})
	}

	return nil
}

// IsDisabled checks if the device is disabled
func (d *Device) IsDisabled() bool {
	return d.DisabledAt != nil
}

// Disable marks the device as disabled
func (d *Device) Disable() {
	now := time.Now()
	d.DisabledAt = &now
	d.UpdatedAt = now
	d.Active = false
}

// Enable re-enables a disabled device
func (d *Device) Enable() {
	d.DisabledAt = nil
	d.UpdatedAt = time.Now()
}

// UpdateHeartbeat updates last seen timestamp and marks as active
func (d *Device) UpdateHeartbeat(ipAddress string) {
	d.LastSeen = time.Now()
	d.Active = true
	if ipAddress != "" {
		d.IPAddress = ipAddress
	}
	d.UpdatedAt = time.Now()
}

// MarkInactive marks the device as inactive (e.g., after timeout)
func (d *Device) MarkInactive() {
	d.Active = false
	d.UpdatedAt = time.Now()
}
