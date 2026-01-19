package service_test

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ══════════════════════════════════════════════════════════════════════════════
// MOCK REPOSITORIES
// ══════════════════════════════════════════════════════════════════════════════

type MockRoleRepository struct {
	userRoles       []domain.Role
	rolePermissions []domain.RolePermission
	role            *domain.Role
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	return m.userRoles, nil
}

func (m *MockRoleRepository) GetPermissions(ctx context.Context, roleID string) ([]domain.RolePermission, error) {
	return m.rolePermissions, nil
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	return m.role, nil
}

// Implement other required methods (stubs)
func (m *MockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return nil
}

func (m *MockRoleRepository) GetByTenantID(ctx context.Context, tenantID string) ([]domain.Role, error) {
	return nil, nil
}

func (m *MockRoleRepository) GetDefaultRole(ctx context.Context, tenantID string) (*domain.Role, error) {
	return nil, nil
}

func (m *MockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockRoleRepository) UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error {
	return nil
}

func (m *MockRoleRepository) SetPermissions(ctx context.Context, roleID string, permissions []domain.RolePermission) error {
	return nil
}

func (m *MockRoleRepository) AssignToUser(ctx context.Context, userRole *domain.UserRole) error {
	return nil
}

func (m *MockRoleRepository) RemoveFromUser(ctx context.Context, userID, roleID string) error {
	return nil
}

type MockPermissionRepository struct {
	channelOverrides []domain.ChannelPermissionOverride
}

func (m *MockPermissionRepository) GetChannelOverrides(ctx context.Context, channelID string) ([]domain.ChannelPermissionOverride, error) {
	return m.channelOverrides, nil
}

// Implement other required methods (stubs)
func (m *MockPermissionRepository) GetAllDefinitions(ctx context.Context) ([]domain.PermissionDefinition, error) {
	return nil, nil
}

func (m *MockPermissionRepository) GetDefinitionsByCategory(ctx context.Context, category domain.PermissionCategory) ([]domain.PermissionDefinition, error) {
	return nil, nil
}

func (m *MockPermissionRepository) GetChannelOverridesByRole(ctx context.Context, channelID, roleID string) ([]domain.ChannelPermissionOverride, error) {
	return nil, nil
}

func (m *MockPermissionRepository) GetChannelOverridesByUser(ctx context.Context, channelID, userID string) ([]domain.ChannelPermissionOverride, error) {
	return nil, nil
}

func (m *MockPermissionRepository) SetChannelOverride(ctx context.Context, override *domain.ChannelPermissionOverride) error {
	return nil
}

func (m *MockPermissionRepository) DeleteChannelOverride(ctx context.Context, channelID, targetID, permissionID string) error {
	return nil
}

func (m *MockPermissionRepository) GetEffectivePermissions(ctx context.Context, userID, channelID string) (map[string]bool, error) {
	return nil, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// PERMISSION RESOLVER TESTS
// ══════════════════════════════════════════════════════════════════════════════

func TestPermissionResolver_CheckPermission_BasePermission(t *testing.T) {
	// Setup
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{
				ID:       "role_1",
				TenantID: "tenant_1",
				Name:     "Member",
				Position: 1,
			},
		},
		rolePermissions: []domain.RolePermission{
			{
				RoleID:     "role_1",
				Permission: domain.PermissionViewChannels,
				Allowed:    true,
			},
			{
				RoleID:     "role_1",
				Permission: domain.PermissionSendMessages,
				Allowed:    true,
			},
		},
	}

	permRepo := &MockPermissionRepository{
		channelOverrides: []domain.ChannelPermissionOverride{},
	}

	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: User has VIEW_CHANNELS permission
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "", domain.PermissionViewChannels)

	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.True(t, result.BaseAllowed)
	assert.Equal(t, "none", result.OverrideType)
	assert.Contains(t, result.Reason, "granted by role permissions")
}

func TestPermissionResolver_CheckPermission_NoPermission(t *testing.T) {
	// Setup
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{
				ID:       "role_1",
				TenantID: "tenant_1",
				Name:     "Member",
				Position: 1,
			},
		},
		rolePermissions: []domain.RolePermission{
			{
				RoleID:     "role_1",
				Permission: domain.PermissionViewChannels,
				Allowed:    true,
			},
		},
	}

	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: User does NOT have MANAGE_MESSAGES permission
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "", domain.PermissionManageMessages)

	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.False(t, result.BaseAllowed)
	assert.Equal(t, "none", result.OverrideType)
	assert.Contains(t, result.Reason, "not granted")
}

func TestPermissionResolver_CheckPermission_AdministratorBypass(t *testing.T) {
	// Setup: User has ADMINISTRATOR permission
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{
				ID:       "role_admin",
				TenantID: "tenant_1",
				Name:     "Administrator",
				Position: 100,
			},
		},
		rolePermissions: []domain.RolePermission{
			{
				RoleID:     "role_admin",
				Permission: domain.PermissionAdministrator,
				Allowed:    true,
			},
		},
	}

	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Admin should have ALL permissions (even ones not explicitly granted)
	result, err := resolver.CheckPermission(ctx, "user_admin", "tenant_1", "", domain.PermissionManageMessages)

	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.True(t, result.BaseAllowed)
	assert.Contains(t, result.Reason, "ADMINISTRATOR")
}

func TestPermissionResolver_CheckPermission_ChannelOverride_Deny(t *testing.T) {
	// Setup: User has base permission, but channel override DENIES it
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{
				ID:       "role_1",
				TenantID: "tenant_1",
				Name:     "Member",
				Position: 1,
			},
		},
		rolePermissions: []domain.RolePermission{
			{
				RoleID:     "role_1",
				Permission: domain.PermissionSendMessages,
				Allowed:    true, // Base permission granted
			},
		},
	}

	roleID := "role_1"
	allowedFalse := false

	permRepo := &MockPermissionRepository{
		channelOverrides: []domain.ChannelPermissionOverride{
			{
				ID:         "override_1",
				ChannelID:  "channel_1",
				RoleID:     &roleID,
				Permission: domain.PermissionSendMessages,
				Allowed:    &allowedFalse, // DENY override
			},
		},
	}

	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Channel override should DENY permission
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "channel_1", domain.PermissionSendMessages)

	require.NoError(t, err)
	assert.False(t, result.Allowed)      // Override denies
	assert.True(t, result.BaseAllowed)   // But base allows
	assert.Equal(t, "deny", result.OverrideType)
	assert.Contains(t, result.Reason, "denied by channel override")
}

func TestPermissionResolver_CheckPermission_ChannelOverride_Allow(t *testing.T) {
	// Setup: User does NOT have base permission, but channel override ALLOWS it
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{
				ID:       "role_1",
				TenantID: "tenant_1",
				Name:     "Member",
				Position: 1,
			},
		},
		rolePermissions: []domain.RolePermission{
			// No MANAGE_MESSAGES permission
		},
	}

	roleID := "role_1"
	allowedTrue := true

	permRepo := &MockPermissionRepository{
		channelOverrides: []domain.ChannelPermissionOverride{
			{
				ID:         "override_1",
				ChannelID:  "channel_1",
				RoleID:     &roleID,
				Permission: domain.PermissionManageMessages,
				Allowed:    &allowedTrue, // ALLOW override
			},
		},
	}

	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Channel override should ALLOW permission
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "channel_1", domain.PermissionManageMessages)

	require.NoError(t, err)
	assert.True(t, result.Allowed)       // Override allows
	assert.False(t, result.BaseAllowed)  // But base denies
	assert.Equal(t, "allow", result.OverrideType)
	assert.Contains(t, result.Reason, "allowed by channel override")
}

func TestPermissionResolver_CheckPermission_DenyOverridesAllow(t *testing.T) {
	// Setup: Multiple overrides - DENY should take precedence over ALLOW
	roleID1 := "role_1"
	roleID2 := "role_2"
	allowedTrue := true
	allowedFalse := false

	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{ID: "role_1", TenantID: "tenant_1", Position: 1},
			{ID: "role_2", TenantID: "tenant_1", Position: 2},
		},
		rolePermissions: []domain.RolePermission{},
	}

	permRepo := &MockPermissionRepository{
		channelOverrides: []domain.ChannelPermissionOverride{
			{
				ID:         "override_allow",
				ChannelID:  "channel_1",
				RoleID:     &roleID1,
				Permission: domain.PermissionSendMessages,
				Allowed:    &allowedTrue, // ALLOW
			},
			{
				ID:         "override_deny",
				ChannelID:  "channel_1",
				RoleID:     &roleID2,
				Permission: domain.PermissionSendMessages,
				Allowed:    &allowedFalse, // DENY
			},
		},
	}

	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: DENY should override ALLOW
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "channel_1", domain.PermissionSendMessages)

	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, "deny", result.OverrideType)
	assert.Contains(t, result.Reason, "denied")
}

func TestPermissionResolver_CheckMultiplePermissions(t *testing.T) {
	// Setup
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{ID: "role_1", TenantID: "tenant_1", Position: 1},
		},
		rolePermissions: []domain.RolePermission{
			{RoleID: "role_1", Permission: domain.PermissionViewChannels, Allowed: true},
			{RoleID: "role_1", Permission: domain.PermissionSendMessages, Allowed: true},
			// No MANAGE_MESSAGES
		},
	}

	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Check multiple permissions at once
	results, err := resolver.CheckMultiplePermissions(ctx, "user_1", "tenant_1", "", []string{
		domain.PermissionViewChannels,
		domain.PermissionSendMessages,
		domain.PermissionManageMessages,
	})

	require.NoError(t, err)
	assert.Len(t, results, 3)

	assert.True(t, results[domain.PermissionViewChannels].Allowed)
	assert.True(t, results[domain.PermissionSendMessages].Allowed)
	assert.False(t, results[domain.PermissionManageMessages].Allowed)
}

func TestPermissionResolver_IsValidPermission(t *testing.T) {
	roleRepo := &MockRoleRepository{}
	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)

	// Test valid permissions
	assert.True(t, resolver.IsValidPermission(domain.PermissionAdministrator))
	assert.True(t, resolver.IsValidPermission(domain.PermissionViewChannels))
	assert.True(t, resolver.IsValidPermission(domain.PermissionSendMessages))
	assert.True(t, resolver.IsValidPermission(domain.PermissionConnect))

	// Test invalid permissions
	assert.False(t, resolver.IsValidPermission("invalid.permission"))
	assert.False(t, resolver.IsValidPermission(""))
	assert.False(t, resolver.IsValidPermission("random_string"))
}

func TestPermissionResolver_NoRoles(t *testing.T) {
	// Setup: User has NO roles
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{}, // Empty
	}

	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Should deny all permissions
	result, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "", domain.PermissionViewChannels)

	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "no roles")
}

func TestPermissionResolver_MultipleRoles_AccumulatePermissions(t *testing.T) {
	// Setup: User has multiple roles with different permissions
	roleRepo := &MockRoleRepository{
		userRoles: []domain.Role{
			{ID: "role_1", TenantID: "tenant_1", Position: 1},
			{ID: "role_2", TenantID: "tenant_1", Position: 2},
		},
		rolePermissions: []domain.RolePermission{
			// First call for role_1
			{RoleID: "role_1", Permission: domain.PermissionViewChannels, Allowed: true},
			// Second call for role_2
			{RoleID: "role_2", Permission: domain.PermissionSendMessages, Allowed: true},
		},
	}

	permRepo := &MockPermissionRepository{}
	resolver := service.NewPermissionResolver(roleRepo, permRepo)
	ctx := context.Background()

	// Test: Should accumulate permissions from all roles (bitwise OR)
	result1, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "", domain.PermissionViewChannels)
	require.NoError(t, err)
	assert.True(t, result1.Allowed)

	result2, err := resolver.CheckPermission(ctx, "user_1", "tenant_1", "", domain.PermissionSendMessages)
	require.NoError(t, err)
	assert.True(t, result2.Allowed)
}
