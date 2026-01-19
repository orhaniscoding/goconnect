package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// SECTION REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// SectionRepository defines operations for sections
type SectionRepository interface {
	// Create creates a new section
	Create(ctx context.Context, section *domain.Section) error

	// GetByID retrieves a section by ID
	GetByID(ctx context.Context, id string) (*domain.Section, error)

	// GetByTenantID retrieves all sections for a tenant
	GetByTenantID(ctx context.Context, tenantID string) ([]domain.Section, error)

	// Update updates a section
	Update(ctx context.Context, section *domain.Section) error

	// Delete soft-deletes a section
	Delete(ctx context.Context, id string) error

	// UpdatePositions updates positions for multiple sections
	UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error
}

// ═══════════════════════════════════════════════════════════════════════════
// CHANNEL REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// ChannelFilter defines filters for listing channels
type ChannelFilter struct {
	TenantID  *string
	SectionID *string
	NetworkID *string
	Type      *domain.ChannelType
	Limit     int
	Cursor    string
}

// ChannelRepository defines operations for channels
type ChannelRepository interface {
	// Create creates a new channel
	Create(ctx context.Context, channel *domain.Channel) error

	// GetByID retrieves a channel by ID
	GetByID(ctx context.Context, id string) (*domain.Channel, error)

	// List retrieves channels with filters
	List(ctx context.Context, filter ChannelFilter) ([]domain.Channel, string, error)

	// Update updates a channel
	Update(ctx context.Context, channel *domain.Channel) error

	// Delete soft-deletes a channel
	Delete(ctx context.Context, id string) error

	// UpdatePositions updates positions for multiple channels
	UpdatePositions(ctx context.Context, parentID string, positions map[string]int) error
}

// ═══════════════════════════════════════════════════════════════════════════
// ROLE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// RoleFilter defines filters for listing roles
type RoleFilter struct {
	TenantID string
	Limit    int
	Cursor   string
}

// RoleRepository defines operations for roles
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *domain.Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id string) (*domain.Role, error)

	// GetByTenantID retrieves all roles for a tenant (ordered by position DESC)
	GetByTenantID(ctx context.Context, tenantID string) ([]domain.Role, error)

	// GetDefaultRole retrieves the default role for a tenant
	GetDefaultRole(ctx context.Context, tenantID string) (*domain.Role, error)

	// Update updates a role
	Update(ctx context.Context, role *domain.Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id string) error

	// UpdatePositions updates positions for multiple roles
	UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error

	// GetPermissions retrieves all permissions for a role
	GetPermissions(ctx context.Context, roleID string) ([]domain.RolePermission, error)

	// SetPermissions sets permissions for a role (replaces all)
	SetPermissions(ctx context.Context, roleID string, permissions []domain.RolePermission) error

	// AssignToUser assigns a role to a user
	AssignToUser(ctx context.Context, userRole *domain.UserRole) error

	// RemoveFromUser removes a role from a user
	RemoveFromUser(ctx context.Context, userID, roleID string) error

	// GetUserRoles retrieves all roles for a user in a tenant
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]domain.Role, error)
}

// ═══════════════════════════════════════════════════════════════════════════
// PERMISSION REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PermissionRepository defines operations for permissions
type PermissionRepository interface {
	// GetAllDefinitions retrieves all permission definitions
	GetAllDefinitions(ctx context.Context) ([]domain.PermissionDefinition, error)

	// GetDefinitionsByCategory retrieves permissions by category
	GetDefinitionsByCategory(ctx context.Context, category domain.PermissionCategory) ([]domain.PermissionDefinition, error)

	// GetChannelOverrides retrieves permission overrides for a channel
	GetChannelOverrides(ctx context.Context, channelID string) ([]domain.ChannelPermissionOverride, error)

	// SetChannelOverride sets a permission override for a channel
	SetChannelOverride(ctx context.Context, override *domain.ChannelPermissionOverride) error

	// DeleteChannelOverride deletes a permission override
	DeleteChannelOverride(ctx context.Context, channelID, targetID, permissionID string) error

	// GetEffectivePermissions calculates effective permissions for a user in a channel
	GetEffectivePermissions(ctx context.Context, userID, channelID string) (map[string]bool, error)
}
