package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/redis/go-redis/v9"
)

// AdminService handles admin-related operations
type AdminService struct {
	userRepo             repository.UserRepository
	adminRepo            repository.AdminRepositoryInterface
	tenantRepo           repository.TenantRepository
	networkRepo          repository.NetworkRepository
	deviceRepo           repository.DeviceRepository
	chatRepo             repository.ChatRepository
	auditor              audit.Auditor
	redisClient          *redis.Client
	getActiveConnections func() int
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo repository.UserRepository,
	adminRepo repository.AdminRepositoryInterface,
	tenantRepo repository.TenantRepository,
	networkRepo repository.NetworkRepository,
	deviceRepo repository.DeviceRepository,
	chatRepo repository.ChatRepository,
	auditor audit.Auditor,
	redisClient *redis.Client,
	getActiveConnections func() int,
) *AdminService {
	return &AdminService{
		userRepo:             userRepo,
		adminRepo:            adminRepo,
		tenantRepo:           tenantRepo,
		networkRepo:          networkRepo,
		deviceRepo:           deviceRepo,
		chatRepo:             chatRepo,
		auditor:              auditor,
		redisClient:          redisClient,
		getActiveConnections: getActiveConnections,
	}
}

// ListUsers retrieves a paginated list of users
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int, query string) ([]*domain.User, int, error) {
	users, total, err := s.userRepo.ListAll(ctx, limit, offset, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}

// ListTenants retrieves a paginated list of tenants
func (s *AdminService) ListTenants(ctx context.Context, limit, offset int, query string) ([]*domain.Tenant, int, error) {
	tenants, total, err := s.tenantRepo.ListAll(ctx, limit, offset, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, total, nil
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
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	tenantCount, err := s.tenantRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tenants: %w", err)
	}

	networkCount, err := s.networkRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count networks: %w", err)
	}

	deviceCount, err := s.deviceRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count devices: %w", err)
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
		return nil, fmt.Errorf("failed to get user for admin toggle: %w", err)
	}

	user.IsAdmin = !user.IsAdmin
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user admin status: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user and all associated data
func (s *AdminService) DeleteUser(ctx context.Context, userID string) error {
	// In a real implementation, we should probably soft-delete or check for active resources
	// For now, we'll just delete the user.
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// DeleteTenant deletes a tenant and all associated data
func (s *AdminService) DeleteTenant(ctx context.Context, tenantID string) error {
	// In a real implementation, we should probably soft-delete or check for active resources
	// For now, we'll just delete the tenant.
	// Note: Foreign key constraints in DB should handle cascading if configured,
	// otherwise we might need to delete users/networks first.
	if err := s.tenantRepo.Delete(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

// ListNetworks retrieves a paginated list of all networks (system-wide)
func (s *AdminService) ListNetworks(ctx context.Context, limit int, cursor, query string) ([]*domain.Network, string, error) {
	nets, nextCursor, err := s.networkRepo.List(ctx, repository.NetworkFilter{
		TenantID: "", // Empty means all tenants
		IsAdmin:  true,
		Limit:    limit,
		Cursor:   cursor,
		Search:   query,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to list all networks: %w", err)
	}
	return nets, nextCursor, nil
}

// ListDevices retrieves a paginated list of all devices (system-wide)
func (s *AdminService) ListDevices(ctx context.Context, limit int, cursor, query string) ([]*domain.Device, string, error) {
	devices, nextCursor, err := s.deviceRepo.List(ctx, domain.DeviceFilter{
		TenantID: "", // Empty means all tenants
		Limit:    limit,
		Cursor:   cursor,
		Search:   query,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to list all devices: %w", err)
	}
	return devices, nextCursor, nil
}

// ====== NEW ADMIN USER MANAGEMENT METHODS ======

// ListAllUsers retrieves all users with filtering and pagination (admin only)
func (s *AdminService) ListAllUsers(ctx context.Context, adminUserID string, filters domain.UserFilters, pagination domain.PaginationParams) ([]*domain.UserListItem, int, error) {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return nil, 0, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return nil, 0, domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Set default pagination
	if pagination.PerPage == 0 {
		pagination.PerPage = 50
	}
	if pagination.PerPage > 100 {
		pagination.PerPage = 100
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	// Fetch users
	users, totalCount, err := s.adminRepo.ListAllUsers(ctx, filters, pagination)
	if err != nil {
		return nil, 0, domain.NewError(domain.ErrInternalServer, "Failed to retrieve users", nil)
	}

	// Log admin action
	if s.auditor != nil {
		s.auditor.Event(ctx, admin.TenantID, "LIST_USERS", adminUserID, "users", map[string]any{
			"filters": fmt.Sprintf("%+v", filters),
		})
	}

	return users, totalCount, nil
}

// GetUserStats retrieves system-wide user statistics (admin only)
func (s *AdminService) GetUserStats(ctx context.Context, adminUserID string) (*domain.SystemStats, error) {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return nil, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return nil, domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Fetch stats
	stats, err := s.adminRepo.GetUserStats(ctx)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to retrieve stats", nil)
	}

	return stats, nil
}

// UpdateUserRole updates a user's admin or moderator status (admin only)
func (s *AdminService) UpdateUserRole(ctx context.Context, adminUserID, targetUserID string, isAdmin, isModerator *bool) error {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Prevent self-demotion
	if adminUserID == targetUserID {
		if isAdmin != nil && !*isAdmin {
			return domain.NewError(domain.ErrInvalidRequest, "Cannot remove your own admin privileges", nil)
		}
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return domain.NewError(domain.ErrNotFound, "User not found", nil)
	}

	// Update role
	err = s.adminRepo.UpdateUserRole(ctx, targetUserID, isAdmin, isModerator)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to update user role", nil)
	}

	// Log admin action
	if s.auditor != nil {
		auditDetails := map[string]any{
			"target_email": targetUser.Email,
		}
		if isAdmin != nil {
			auditDetails["is_admin"] = *isAdmin
		}
		if isModerator != nil {
			auditDetails["is_moderator"] = *isModerator
		}
		s.auditor.Event(ctx, admin.TenantID, "UPDATE_USER_ROLE", adminUserID, "user:"+targetUserID, auditDetails)
	}

	return nil
}

// SuspendUser suspends a user account (admin only)
func (s *AdminService) SuspendUser(ctx context.Context, adminUserID, targetUserID, reason string) error {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Prevent self-suspension
	if adminUserID == targetUserID {
		return domain.NewError(domain.ErrInvalidRequest, "Cannot suspend your own account", nil)
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return domain.NewError(domain.ErrNotFound, "User not found", nil)
	}

	// Cannot suspend another admin (only super admin can do that in future)
	if targetUser.IsAdmin {
		return domain.NewError(domain.ErrForbidden, "Cannot suspend another admin user", nil)
	}

	// Suspend user
	err = s.adminRepo.SuspendUser(ctx, targetUserID, reason, adminUserID)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to suspend user", nil)
	}

	// Blacklist all active sessions (invalidate all tokens immediately)
	if s.redisClient != nil {
		sessionKey := fmt.Sprintf("user_sessions:%s", targetUserID)
		jtis, err := s.redisClient.SMembers(ctx, sessionKey).Result()
		if err == nil && len(jtis) > 0 {
			// Blacklist each JTI (access and refresh tokens)
			for _, jti := range jtis {
				blacklistKey := fmt.Sprintf("blacklist:%s", jti)
				// TTL: 7 days (max refresh token lifetime)
				s.redisClient.Set(ctx, blacklistKey, "suspended", 7*24*time.Hour)
			}
			// Clear user sessions set
			s.redisClient.Del(ctx, sessionKey)
		}
	}

	// Log admin action
	if s.auditor != nil {
		s.auditor.Event(ctx, admin.TenantID, "SUSPEND_USER", adminUserID, "user:"+targetUserID, map[string]any{
			"target_email": targetUser.Email,
			"reason":       reason,
		})
	}

	return nil
}

// UnsuspendUser unsuspends a user account (admin only)
func (s *AdminService) UnsuspendUser(ctx context.Context, adminUserID, targetUserID string) error {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return domain.NewError(domain.ErrNotFound, "User not found", nil)
	}

	// Unsuspend user
	err = s.adminRepo.UnsuspendUser(ctx, targetUserID)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to unsuspend user", nil)
	}

	// Log admin action
	if s.auditor != nil {
		s.auditor.Event(ctx, admin.TenantID, "UNSUSPEND_USER", adminUserID, "user:"+targetUserID, map[string]any{
			"target_email": targetUser.Email,
		})
	}

	return nil
}

// GetUserDetails retrieves full details of a user (admin only)
func (s *AdminService) GetUserDetails(ctx context.Context, adminUserID, targetUserID string) (*domain.User, error) {
	// Verify admin status
	admin, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return nil, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil)
	}

	if !admin.IsAdmin {
		return nil, domain.NewError(domain.ErrForbidden, "Admin access required", nil)
	}

	// Get user details
	user, err := s.adminRepo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, domain.NewError(domain.ErrNotFound, "User not found", nil)
	}

	// Remove sensitive data
	user.PasswordHash = ""
	user.TwoFAKey = ""
	user.RecoveryCodes = nil

	return user, nil
}

// UpdateLastSeen updates the last seen timestamp for a user
func (s *AdminService) UpdateLastSeen(ctx context.Context, userID string) error {
	if err := s.adminRepo.UpdateLastSeen(ctx, userID); err != nil {
		return fmt.Errorf("failed to update user last seen: %w", err)
	}
	return nil
}
