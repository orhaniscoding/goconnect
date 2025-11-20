package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// DeviceRepository defines the interface for device storage
type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetByID(ctx context.Context, id string) (*domain.Device, error)
	GetByPubKey(ctx context.Context, pubkey string) (*domain.Device, error)
	List(ctx context.Context, filter domain.DeviceFilter) ([]*domain.Device, string, error)
	Update(ctx context.Context, device *domain.Device) error
	Delete(ctx context.Context, id string) error

	// Heartbeat operations
	UpdateHeartbeat(ctx context.Context, id string, ipAddress string) error
	MarkInactive(ctx context.Context, id string) error
}

// InMemoryDeviceRepository implements DeviceRepository with in-memory storage
type InMemoryDeviceRepository struct {
	mu      sync.RWMutex
	devices map[string]*domain.Device // id -> device
	pubkeys map[string]string         // pubkey -> device_id
}

// NewInMemoryDeviceRepository creates a new in-memory device repository
func NewInMemoryDeviceRepository() *InMemoryDeviceRepository {
	return &InMemoryDeviceRepository{
		devices: make(map[string]*domain.Device),
		pubkeys: make(map[string]string),
	}
}

// Create creates a new device
func (r *InMemoryDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate pubkey
	if existingID, exists := r.pubkeys[device.PubKey]; exists {
		return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
			"pubkey":             device.PubKey,
			"existing_device_id": existingID,
		})
	}

	if device.ID == "" {
		device.ID = domain.GenerateNetworkID() // Reuse ULID generator
	}

	if device.CreatedAt.IsZero() {
		device.CreatedAt = time.Now()
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = device.CreatedAt
	}

	// Store device
	r.devices[device.ID] = device
	r.pubkeys[device.PubKey] = device.ID

	return nil
}

// GetByID retrieves a device by ID
func (r *InMemoryDeviceRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, ok := r.devices[id]
	if !ok {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	return device, nil
}

// GetByPubKey retrieves a device by public key
func (r *InMemoryDeviceRepository) GetByPubKey(ctx context.Context, pubkey string) (*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deviceID, ok := r.pubkeys[pubkey]
	if !ok {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"pubkey": pubkey,
		})
	}

	device := r.devices[deviceID]
	return device, nil
}

// List retrieves devices matching the filter
func (r *InMemoryDeviceRepository) List(ctx context.Context, filter domain.DeviceFilter) ([]*domain.Device, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*domain.Device

	// Filter devices
	for _, device := range r.devices {
		// User filter
		if filter.UserID != "" && device.UserID != filter.UserID {
			continue
		}

		// Tenant filter
		if filter.TenantID != "" && device.TenantID != filter.TenantID {
			continue
		}

		// Platform filter
		if filter.Platform != "" && device.Platform != filter.Platform {
			continue
		}

		// Active filter
		if filter.Active != nil && device.Active != *filter.Active {
			continue
		}

		matches = append(matches, device)
	}

	// Sort by CreatedAt DESC (newest first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].CreatedAt.Before(matches[j].CreatedAt) {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Apply pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Cursor pagination
	startIdx := 0
	if filter.Cursor != "" {
		for i, device := range matches {
			if device.ID == filter.Cursor {
				startIdx = i + 1
				break
			}
		}
	}

	endIdx := startIdx + limit
	if endIdx > len(matches) {
		endIdx = len(matches)
	}

	result := matches[startIdx:endIdx]

	var nextCursor string
	if endIdx < len(matches) {
		nextCursor = matches[endIdx-1].ID
	}

	return result, nextCursor, nil
}

// Update updates an existing device
func (r *InMemoryDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.devices[device.ID]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": device.ID,
		})
	}

	// If pubkey changed, update index
	if existing.PubKey != device.PubKey {
		// Check for duplicate new pubkey
		if existingID, exists := r.pubkeys[device.PubKey]; exists && existingID != device.ID {
			return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
				"pubkey":             device.PubKey,
				"existing_device_id": existingID,
			})
		}

		// Remove old pubkey mapping
		delete(r.pubkeys, existing.PubKey)
		// Add new pubkey mapping
		r.pubkeys[device.PubKey] = device.ID
	}

	device.UpdatedAt = time.Now()
	r.devices[device.ID] = device
	return nil
}

// Delete deletes a device
func (r *InMemoryDeviceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	device, ok := r.devices[id]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	delete(r.devices, id)
	delete(r.pubkeys, device.PubKey)
	return nil
}

// UpdateHeartbeat updates the last seen timestamp and marks device as active
func (r *InMemoryDeviceRepository) UpdateHeartbeat(ctx context.Context, id string, ipAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	device, ok := r.devices[id]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	device.UpdateHeartbeat(ipAddress)
	return nil
}

// MarkInactive marks a device as inactive
func (r *InMemoryDeviceRepository) MarkInactive(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	device, ok := r.devices[id]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	device.MarkInactive()
	return nil
}
