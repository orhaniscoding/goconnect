package service

import (
	"context"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// DeviceService handles device operations
type DeviceService struct {
	deviceRepo       repository.DeviceRepository
	userRepo         repository.UserRepository
	peerProvisioning *PeerProvisioningService
	auditor          Auditor
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo repository.DeviceRepository, userRepo repository.UserRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		userRepo:   userRepo,
		auditor:    noopAuditor,
	}
}

// SetPeerProvisioning sets the peer provisioning service
func (s *DeviceService) SetPeerProvisioning(pp *PeerProvisioningService) {
	s.peerProvisioning = pp
}

// SetAuditor sets the auditor for the service
func (s *DeviceService) SetAuditor(auditor Auditor) {
	if auditor != nil {
		s.auditor = auditor
	}
}

// RegisterDevice registers a new device for a user
func (s *DeviceService) RegisterDevice(ctx context.Context, userID, tenantID string, req *domain.RegisterDeviceRequest) (*domain.Device, error) {
	// Validate user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{
			"user_id": userID,
		})
	}

	// Ensure tenant matches
	if user.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrForbidden, "Tenant mismatch", map[string]string{
			"user_tenant":    user.TenantID,
			"request_tenant": tenantID,
		})
	}

	// Check if pubkey already exists
	existing, err := s.deviceRepo.GetByPubKey(ctx, req.PubKey)
	if err == nil && existing != nil {
		return nil, domain.NewError(domain.ErrConflict, "A device with this public key already exists", map[string]string{
			"existing_device_id": existing.ID,
			"existing_user_id":   existing.UserID,
		})
	}

	// Create device
	device := &domain.Device{
		ID:        domain.GenerateNetworkID(),
		UserID:    userID,
		TenantID:  tenantID,
		Name:      req.Name,
		Platform:  req.Platform,
		PubKey:    req.PubKey,
		HostName:  req.HostName,
		OSVersion: req.OSVersion,
		DaemonVer: req.DaemonVer,
		Active:    false, // Not active until first heartbeat
	}

	// Validate
	if err := device.Validate(); err != nil {
		return nil, err
	}

	// Save
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "DEVICE_REGISTERED", userID, device.ID, map[string]any{
		"device_name": device.Name,
		"platform":    device.Platform,
		"tenant_id":   tenantID,
	})

	// Automatically provision peers for this device in all user's networks
	if s.peerProvisioning != nil {
		if err := s.peerProvisioning.ProvisionPeersForNewDevice(ctx, device); err != nil {
			// Log error but don't fail device registration
			s.auditor.Event(ctx, tenantID, "PEER_PROVISION_FAILED", userID, device.ID, map[string]any{
				"error": err.Error(),
			})
		}
	}

	return device, nil
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(ctx context.Context, deviceID, userID, tenantID string, isAdmin bool) (*domain.Device, error) {
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: user can only view their own devices unless admin
	if !isAdmin && device.UserID != userID {
		return nil, domain.NewError(domain.ErrForbidden, "You can only view your own devices", nil)
	}

	return device, nil
}

// ListDevices retrieves devices matching the filter
func (s *DeviceService) ListDevices(ctx context.Context, userID, tenantID string, isAdmin bool, filter domain.DeviceFilter) ([]*domain.Device, string, error) {
	// Non-admins can only list their own devices
	if !isAdmin {
		filter.UserID = userID
	}
	// Always filter by tenant
	filter.TenantID = tenantID

	return s.deviceRepo.List(ctx, filter)
}

// UpdateDevice updates device information
func (s *DeviceService) UpdateDevice(ctx context.Context, deviceID, userID, tenantID string, isAdmin bool, req *domain.UpdateDeviceRequest) (*domain.Device, error) {
	// Get device
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: user can only update their own devices unless admin
	if !isAdmin && device.UserID != userID {
		return nil, domain.NewError(domain.ErrForbidden, "You can only update your own devices", nil)
	}

	// Apply updates
	if req.Name != nil {
		device.Name = *req.Name
	}

	if req.PubKey != nil {
		// Check if new pubkey already exists
		existing, err := s.deviceRepo.GetByPubKey(ctx, *req.PubKey)
		if err == nil && existing != nil && existing.ID != deviceID {
			return nil, domain.NewError(domain.ErrConflict, "A device with this public key already exists", map[string]string{
				"existing_device_id": existing.ID,
			})
		}
		device.PubKey = *req.PubKey
	}

	if req.HostName != nil {
		device.HostName = *req.HostName
	}

	if req.OSVersion != nil {
		device.OSVersion = *req.OSVersion
	}

	if req.DaemonVer != nil {
		device.DaemonVer = *req.DaemonVer
	}

	// Validate
	if err := device.Validate(); err != nil {
		return nil, err
	}

	// Update
	if err := s.deviceRepo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "DEVICE_UPDATED", userID, deviceID, map[string]any{
		"updated_fields": getUpdatedFields(req),
	})

	return device, nil
}

// DeleteDevice deletes a device
func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID, userID, tenantID string, isAdmin bool) error {
	// Get device
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: user can only delete their own devices unless admin
	if !isAdmin && device.UserID != userID {
		return domain.NewError(domain.ErrForbidden, "You can only delete your own devices", nil)
	}

	// Delete
	if err := s.deviceRepo.Delete(ctx, deviceID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "DEVICE_DELETED", userID, deviceID, map[string]any{
		"device_name": device.Name,
	})

	return nil
}

// Heartbeat processes a heartbeat from a device
func (s *DeviceService) Heartbeat(ctx context.Context, deviceID, userID, tenantID string, req *domain.DeviceHeartbeatRequest) error {
	// Get device
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: only device owner can send heartbeat
	if device.UserID != userID {
		return domain.NewError(domain.ErrForbidden, "You can only send heartbeat for your own devices", nil)
	}

	// Check if device is disabled
	if device.IsDisabled() {
		return domain.NewError(domain.ErrForbidden, "Device is disabled", nil)
	}

	// Update heartbeat
	if err := s.deviceRepo.UpdateHeartbeat(ctx, deviceID, req.IPAddress); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// Update daemon version and OS version if provided
	if req.DaemonVer != "" || req.OSVersion != "" {
		device, _ := s.deviceRepo.GetByID(ctx, deviceID)
		if req.DaemonVer != "" {
			device.DaemonVer = req.DaemonVer
		}
		if req.OSVersion != "" {
			device.OSVersion = req.OSVersion
		}
		s.deviceRepo.Update(ctx, device)
	}

	return nil
}

// DisableDevice disables a device (soft disable)
func (s *DeviceService) DisableDevice(ctx context.Context, deviceID, userID, tenantID string, isAdmin bool) error {
	// Get device
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: user can only disable their own devices unless admin
	if !isAdmin && device.UserID != userID {
		return domain.NewError(domain.ErrForbidden, "You can only disable your own devices", nil)
	}

	// Disable
	device.Disable()
	if err := s.deviceRepo.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to disable device: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "DEVICE_DISABLED", userID, deviceID, nil)

	return nil
}

// EnableDevice re-enables a disabled device
func (s *DeviceService) EnableDevice(ctx context.Context, deviceID, userID, tenantID string, isAdmin bool) error {
	// Get device
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}

	// Ensure device belongs to the tenant
	if device.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Device not found", nil)
	}

	// Authorization: user can only enable their own devices unless admin
	if !isAdmin && device.UserID != userID {
		return domain.NewError(domain.ErrForbidden, "You can only enable your own devices", nil)
	}

	// Enable
	device.Enable()
	if err := s.deviceRepo.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to enable device: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "DEVICE_ENABLED", userID, deviceID, nil)

	return nil
}

// Helper function to track which fields were updated
func getUpdatedFields(req *domain.UpdateDeviceRequest) []string {
	fields := []string{}
	if req.Name != nil {
		fields = append(fields, "name")
	}
	if req.PubKey != nil {
		fields = append(fields, "pubkey")
	}
	if req.HostName != nil {
		fields = append(fields, "hostname")
	}
	if req.OSVersion != nil {
		fields = append(fields, "os_version")
	}
	if req.DaemonVer != nil {
		fields = append(fields, "daemon_ver")
	}
	return fields
}
