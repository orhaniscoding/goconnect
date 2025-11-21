package service

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// AdminService handles admin-related operations
type AdminService struct {
	userRepo             repository.UserRepository
	tenantRepo           repository.TenantRepository
	networkRepo          repository.NetworkRepository
	deviceRepo           repository.DeviceRepository
	chatRepo             repository.ChatRepository
	getActiveConnections func() int
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	networkRepo repository.NetworkRepository,
	deviceRepo repository.DeviceRepository,
	chatRepo repository.ChatRepository,
	getActiveConnections func() int,
) *AdminService {
	return &AdminService{
		userRepo:             userRepo,
		tenantRepo:           tenantRepo,
		networkRepo:          networkRepo,
		deviceRepo:           deviceRepo,
		chatRepo:             chatRepo,
		getActiveConnections: getActiveConnections,
	}
}

// ListUsers retrieves a paginated list of users
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int, query string) ([]*domain.User, int, error) {
	return s.userRepo.ListAll(ctx, limit, offset, query)
}

// ListTenants retrieves a paginated list of tenants
func (s *AdminService) ListTenants(ctx context.Context, limit, offset int, query string) ([]*domain.Tenant, int, error) {
	return s.tenantRepo.ListAll(ctx, limit, offset, query)
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers        int `json:"total_users"`
	TotalTenants      int `json:"total_tenants"`
	TotalNetworks     int `json:"total_networks"`
	TotalDevices      int `json:"total_devices"`
	ActiveConnections int `json:"active_connections"`
	MessagesToday     int `json:"messages_today"`
}

// GetSystemStats retrieves system statistics
func (s *AdminService) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	tenantCount, err := s.tenantRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	networkCount, err := s.networkRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	deviceCount, err := s.deviceRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	activeConnections := 0
	if s.getActiveConnections != nil {
		activeConnections = s.getActiveConnections()
	}

	messagesToday, err := s.chatRepo.CountToday(ctx)
	if err != nil {
		// Log error but don't fail the whole request
		messagesToday = 0
	}

	return &SystemStats{
		TotalUsers:        userCount,
		TotalTenants:      tenantCount,
		TotalNetworks:     networkCount,
		TotalDevices:      deviceCount,
		ActiveConnections: activeConnections,
		MessagesToday:     messagesToday,
	}, nil
}

// ToggleUserAdmin toggles the admin status of a user
func (s *AdminService) ToggleUserAdmin(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsAdmin = !user.IsAdmin
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user and all associated data
func (s *AdminService) DeleteUser(ctx context.Context, userID string) error {
	// In a real implementation, we should probably soft-delete or check for active resources
	// For now, we'll just delete the user.
	return s.userRepo.Delete(ctx, userID)
}

// DeleteTenant deletes a tenant and all associated data
func (s *AdminService) DeleteTenant(ctx context.Context, tenantID string) error {
	// In a real implementation, we should probably soft-delete or check for active resources
	// For now, we'll just delete the tenant.
	// Note: Foreign key constraints in DB should handle cascading if configured,
	// otherwise we might need to delete users/networks first.
	return s.tenantRepo.Delete(ctx, tenantID)
}

// ListNetworks retrieves a paginated list of all networks (system-wide)
func (s *AdminService) ListNetworks(ctx context.Context, limit int, cursor, query string) ([]*domain.Network, string, error) {
	return s.networkRepo.List(ctx, repository.NetworkFilter{
		TenantID: "", // Empty means all tenants
		IsAdmin:  true,
		Limit:    limit,
		Cursor:   cursor,
		Search:   query,
	})
}

// ListDevices retrieves a paginated list of all devices (system-wide)
func (s *AdminService) ListDevices(ctx context.Context, limit int, cursor, query string) ([]*domain.Device, string, error) {
	return s.deviceRepo.List(ctx, domain.DeviceFilter{
		TenantID: "", // Empty means all tenants
		Limit:    limit,
		Cursor:   cursor,
		Search:   query,
	})
}
