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
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	return s.userRepo.ListAll(ctx, limit, offset)
}

// ListTenants retrieves a paginated list of tenants
func (s *AdminService) ListTenants(ctx context.Context, limit, offset int) ([]*domain.Tenant, int, error) {
	return s.tenantRepo.ListAll(ctx, limit, offset)
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers        int `json:"total_users"`
	TotalTenants      int `json:"total_tenants"`
	TotalNetworks     int `json:"total_networks"`
	TotalDevices      int `json:"total_devices"`
	ActiveConnections int `json:"active_connections"` // Placeholder
	MessagesToday     int `json:"messages_today"`     // Placeholder
}

// GetSystemStats retrieves system statistics
func (s *AdminService) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	// We can optimize this by adding Count() methods to repositories
	// For now, we'll use ListAll with limit 1 to get the total count

	_, userCount, err := s.userRepo.ListAll(ctx, 1, 0)
	if err != nil {
		return nil, err
	}

	_, tenantCount, err := s.tenantRepo.ListAll(ctx, 1, 0)
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
