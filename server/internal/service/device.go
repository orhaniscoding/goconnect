package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// DeviceService handles device operations
type DeviceService struct {
	deviceRepo       repository.DeviceRepository
	userRepo         repository.UserRepository
	peerRepo         repository.PeerRepository
	networkRepo      repository.NetworkRepository
	peerProvisioning *PeerProvisioningService
	auditor          Auditor
	wgConfig         config.WireGuardConfig
	notifier         DeviceNotifier
}

// DeviceNotifier defines interface for device events
type DeviceNotifier interface {
	DeviceOnline(deviceID, userID string)
	DeviceOffline(deviceID, userID string)
}

type noopDeviceNotifier struct{}

func (n noopDeviceNotifier) DeviceOnline(deviceID, userID string)  {}
func (n noopDeviceNotifier) DeviceOffline(deviceID, userID string) {}

// NewDeviceService creates a new device service
func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	userRepo repository.UserRepository,
	peerRepo repository.PeerRepository,
	networkRepo repository.NetworkRepository,
	wgConfig config.WireGuardConfig,
) *DeviceService {
	return &DeviceService{
		deviceRepo:  deviceRepo,
		userRepo:    userRepo,
		peerRepo:    peerRepo,
		networkRepo: networkRepo,
		auditor:     noopAuditor,
		wgConfig:    wgConfig,
		notifier:    noopDeviceNotifier{},
	}
}

// SetNotifier sets the device notifier
func (s *DeviceService) SetNotifier(n DeviceNotifier) {
	if n != nil {
		s.notifier = n
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

	// Check if device was offline
	wasOffline := device.LastSeen.IsZero() || time.Since(device.LastSeen) > 2*time.Minute

	// Update heartbeat
	if err := s.deviceRepo.UpdateHeartbeat(ctx, deviceID, req.IPAddress); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	if wasOffline {
		s.notifier.DeviceOnline(deviceID, userID)
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

	// Notify that device is online
	s.notifier.DeviceOnline(deviceID, userID)

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
	device.UpdatedAt = time.Now()

	return s.deviceRepo.Update(ctx, device)
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

// GetDeviceConfig retrieves the WireGuard configuration for a device
func (s *DeviceService) GetDeviceConfig(ctx context.Context, deviceID, userID string) (*domain.DeviceConfig, error) {
	// Verify device ownership
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if device.UserID != userID {
		return nil, domain.NewError(domain.ErrForbidden, "Device does not belong to user", nil)
	}

	// Get all peers for this device (across all networks)
	peers, err := s.peerRepo.GetByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, domain.NewError(domain.ErrNotFound, "No network configuration found for this device", nil)
	}

	// For now, we assume the device is connected to one primary network or we aggregate.
	// The client daemon expects a single interface config.
	// If a device is in multiple networks, WireGuard usually handles this by having multiple AllowedIPs on the interface
	// and multiple peers (one per network gateway, or just one gateway if they share the same server).
	// Since we have a single server endpoint, we can aggregate all networks into one config.

	// Collect all assigned IPs
	var addresses []string
	var dnsServers []string
	var mtu int = 1420 // Default

	// We need to fetch networks to get DNS and MTU settings
	// We'll use the settings from the first network found, or merge them.
	// Merging DNS is tricky, usually we just take the first one or append.

	// The server is the only peer for now (Hub-and-Spoke)
	// But we need to calculate the AllowedIPs for the server peer.
	// The server should route traffic for all networks the device is in.
	var serverAllowedIPs []string

	for _, peer := range peers {
		if len(peer.AllowedIPs) > 0 {
			addresses = append(addresses, peer.AllowedIPs...)
		}

		network, err := s.networkRepo.GetByID(ctx, peer.NetworkID)
		if err != nil {
			continue
		}

		if network.DNS != nil && *network.DNS != "" {
			dnsServers = append(dnsServers, *network.DNS)
		}
		if network.MTU != nil && *network.MTU > 0 {
			mtu = *network.MTU
		}

		// Add network CIDR to server's AllowedIPs so client routes traffic for this network to server
		serverAllowedIPs = append(serverAllowedIPs, network.CIDR)
	}

	// Default DNS if none provided
	if len(dnsServers) == 0 {
		dnsServers = []string{s.wgConfig.DNS}
	}

	// Construct the config
	config := &domain.DeviceConfig{
		Interface: domain.InterfaceConfig{
			ListenPort: 51820, // Client listen port (can be random, but let's set default)
			Addresses:  addresses,
			DNS:        dnsServers,
			MTU:        mtu,
		},
		Peers: []domain.PeerConfig{
			{
				PublicKey:           s.wgConfig.ServerPubKey,
				Endpoint:            s.wgConfig.ServerEndpoint,
				AllowedIPs:          serverAllowedIPs,
				PersistentKeepalive: s.wgConfig.Keepalive,
				Name:                "GoConnect Server",
				Hostname:            "vpn.goconnect",
			},
		},
	}

	return config, nil
}

// StartOfflineDetection starts a background worker to detect offline devices
func (s *DeviceService) StartOfflineDetection(ctx context.Context, checkInterval, offlineThreshold time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.detectOfflineDevices(ctx, offlineThreshold)
		}
	}
}

func (s *DeviceService) detectOfflineDevices(ctx context.Context, threshold time.Duration) {
	devices, err := s.deviceRepo.GetStaleDevices(ctx, threshold)
	if err != nil {
		fmt.Printf("Error detecting offline devices: %v\n", err)
		return
	}

	for _, device := range devices {
		// Mark as inactive
		if err := s.deviceRepo.MarkInactive(ctx, device.ID); err != nil {
			fmt.Printf("Error marking device %s as inactive: %v\n", device.ID, err)
			continue
		}

		// Notify
		s.notifier.DeviceOffline(device.ID, device.UserID)
	}
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
