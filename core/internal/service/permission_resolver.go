package service

import (
	"context"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// PERMISSION RESOLVER SERVICE
// ══════════════════════════════════════════════════════════════════════════════
// Discord-style permission system:
// 1. Base permissions from roles (accumulated via bitwise OR)
// 2. Channel-specific overrides (allow/deny)
// 3. Role hierarchy (higher position = override lower roles)
// 4. Admin bypass (ADMINISTRATOR permission grants all permissions)

type PermissionResolver struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

func NewPermissionResolver(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *PermissionResolver {
	return &PermissionResolver{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// PermissionResult represents the final permission calculation
type PermissionResult struct {
	Allowed      bool
	Reason       string // For debugging/audit
	BaseAllowed  bool   // Permission from base roles
	OverrideType string // "allow", "deny", or "none"
}

// ══════════════════════════════════════════════════════════════════════════════
// CORE PERMISSION CHECKING
// ══════════════════════════════════════════════════════════════════════════════

// CheckPermission calculates if a user has a specific permission in a channel
// Algorithm:
// 1. Get all user roles in tenant
// 2. Check for ADMINISTRATOR permission (bypass all checks)
// 3. Calculate base permissions (bitwise OR of all role permissions)
// 4. Apply channel overrides (deny takes precedence over allow)
func (pr *PermissionResolver) CheckPermission(
	ctx context.Context,
	userID, tenantID, channelID, permission string,
) (*PermissionResult, error) {
	result := &PermissionResult{
		Allowed:      false,
		Reason:       "no permission",
		BaseAllowed:  false,
		OverrideType: "none",
	}

	// Step 1: Get user's roles in tenant
	userRoles, err := pr.roleRepo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(userRoles) == 0 {
		result.Reason = "user has no roles in tenant"
		return result, nil
	}

	// Step 2: Check for ADMINISTRATOR permission (bypass)
	// This is checked BEFORE channel overrides (admin always wins)
	hasAdmin, err := pr.hasAdministratorPermission(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to check administrator permission: %w", err)
	}

	if hasAdmin {
		result.Allowed = true
		result.BaseAllowed = true
		result.Reason = "user has ADMINISTRATOR permission"
		return result, nil
	}

	// Step 3: Calculate base permissions from roles
	basePermissions, err := pr.getBasePermissions(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get base permissions: %w", err)
	}

	result.BaseAllowed = pr.hasPermissionInSet(permission, basePermissions)

	// Step 4: If no channel specified, return base permission result
	if channelID == "" {
		result.Allowed = result.BaseAllowed
		if result.BaseAllowed {
			result.Reason = "granted by role permissions"
		} else {
			result.Reason = "not granted by role permissions"
		}
		return result, nil
	}

	// Step 5: Apply channel-specific overrides
	// Channel overrides can ALLOW or DENY specific permissions
	// DENY overrides take precedence over ALLOW
	overrideResult, err := pr.applyChannelOverrides(
		ctx,
		userID,
		userRoles,
		channelID,
		permission,
		result.BaseAllowed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to apply channel overrides: %w", err)
	}

	result.Allowed = overrideResult.Allowed
	result.OverrideType = overrideResult.OverrideType
	result.Reason = overrideResult.Reason

	return result, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// BATCH PERMISSION CHECKING
// ══════════════════════════════════════════════════════════════════════════════

// CheckMultiplePermissions checks multiple permissions at once (optimization)
func (pr *PermissionResolver) CheckMultiplePermissions(
	ctx context.Context,
	userID, tenantID, channelID string,
	permissions []string,
) (map[string]*PermissionResult, error) {
	results := make(map[string]*PermissionResult)

	// Get user roles once (optimization)
	userRoles, err := pr.roleRepo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check admin permission once
	hasAdmin, err := pr.hasAdministratorPermission(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to check administrator permission: %w", err)
	}

	// If admin, grant all permissions
	if hasAdmin {
		for _, perm := range permissions {
			results[perm] = &PermissionResult{
				Allowed:     true,
				BaseAllowed: true,
				Reason:      "user has ADMINISTRATOR permission",
			}
		}
		return results, nil
	}

	// Get base permissions once
	basePermissions, err := pr.getBasePermissions(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get base permissions: %w", err)
	}

	// Get channel overrides once (if channel specified)
	var overrides []domain.ChannelPermissionOverride
	if channelID != "" {
		overrides, err = pr.permissionRepo.GetChannelOverrides(ctx, channelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get channel overrides: %w", err)
		}
	}

	// Check each permission
	for _, perm := range permissions {
		result := &PermissionResult{
			BaseAllowed:  pr.hasPermissionInSet(perm, basePermissions),
			OverrideType: "none",
		}

		if channelID == "" {
			result.Allowed = result.BaseAllowed
			if result.BaseAllowed {
				result.Reason = "granted by role permissions"
			} else {
				result.Reason = "not granted by role permissions"
			}
		} else {
			overrideResult := pr.applyOverridesToPermission(
				userID,
				userRoles,
				overrides,
				perm,
				result.BaseAllowed,
			)
			result.Allowed = overrideResult.Allowed
			result.OverrideType = overrideResult.OverrideType
			result.Reason = overrideResult.Reason
		}

		results[perm] = result
	}

	return results, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

// hasAdministratorPermission checks if user has ADMINISTRATOR permission
// ADMINISTRATOR grants ALL permissions (bypass)
func (pr *PermissionResolver) hasAdministratorPermission(
	ctx context.Context,
	userRoles []domain.Role,
) (bool, error) {
	for _, role := range userRoles {
		perms, err := pr.roleRepo.GetPermissions(ctx, role.ID)
		if err != nil {
			return false, fmt.Errorf("failed to get role permissions: %w", err)
		}

		for _, p := range perms {
			if p.Permission == domain.PermissionAdministrator {
				return true, nil
			}
		}
	}

	return false, nil
}

// getBasePermissions accumulates all permissions from user's roles
// Uses bitwise OR logic (any role granting permission = granted)
func (pr *PermissionResolver) getBasePermissions(
	ctx context.Context,
	userRoles []domain.Role,
) ([]string, error) {
	permissionSet := make(map[string]bool)

	for _, role := range userRoles {
		perms, err := pr.roleRepo.GetPermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", role.ID, err)
		}

		for _, p := range perms {
			// Only add allowed permissions
			if p.Allowed {
				permissionSet[p.Permission] = true
			}
		}
	}

	// Convert set to slice
	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// hasPermissionInSet checks if a permission exists in permission list
func (pr *PermissionResolver) hasPermissionInSet(permission string, permissions []string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// applyChannelOverrides applies channel-specific permission overrides
// Priority: DENY > ALLOW > BASE
func (pr *PermissionResolver) applyChannelOverrides(
	ctx context.Context,
	userID string,
	userRoles []domain.Role,
	channelID, permission string,
	baseAllowed bool,
) (*PermissionResult, error) {
	overrides, err := pr.permissionRepo.GetChannelOverrides(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel overrides: %w", err)
	}

	return pr.applyOverridesToPermission(userID, userRoles, overrides, permission, baseAllowed), nil
}

// applyOverridesToPermission applies overrides to a single permission
// Each override contains a single permission with an Allowed flag (*bool)
// - Allowed = true → ALLOW override
// - Allowed = false → DENY override
// - Allowed = nil → INHERIT (no override)
func (pr *PermissionResolver) applyOverridesToPermission(
	userID string,
	userRoles []domain.Role,
	overrides []domain.ChannelPermissionOverride,
	permission string,
	baseAllowed bool,
) *PermissionResult {
	result := &PermissionResult{
		Allowed:      baseAllowed,
		BaseAllowed:  baseAllowed,
		OverrideType: "none",
		Reason:       "base permission applied",
	}

	// Create role ID set for quick lookup
	roleIDs := make(map[string]bool)
	for _, role := range userRoles {
		roleIDs[role.ID] = true
	}

	// Track highest priority override for this permission
	hasAllowOverride := false
	hasDenyOverride := false

	for _, override := range overrides {
		// Skip if override doesn't match the requested permission
		if override.Permission != permission {
			continue
		}

		// User-specific overrides take HIGHEST priority
		// Check if this is a user-specific override for this user
		if override.UserID != nil {
			if *override.UserID == userID {
				// Found a user-specific override - apply immediately
				// User-specific overrides are final and bypass role overrides
				if override.Allowed != nil {
					if *override.Allowed {
						result.Allowed = true
						result.OverrideType = "allow"
						result.Reason = "allowed by user-specific channel override"
					} else {
						result.Allowed = false
						result.OverrideType = "deny"
						result.Reason = "denied by user-specific channel override"
					}
					return result // User-specific override is final
				}
				// If Allowed is nil (inherit), continue to check role overrides
			}
			// User-specific override for different user - skip
			continue
		}

		// Skip if override doesn't apply to user's roles
		if override.RoleID != nil && !roleIDs[*override.RoleID] {
			continue
		}

		// Skip inherit overrides (Allowed = nil)
		if override.Allowed == nil {
			continue
		}

		// Track allow/deny overrides
		if *override.Allowed {
			// ALLOW override
			if !hasAllowOverride || override.RoleID != nil {
				// Prefer role-specific overrides over generic
				hasAllowOverride = true
			}
		} else {
			// DENY override
			if !hasDenyOverride || override.RoleID != nil {
				// Prefer role-specific overrides over generic
				hasDenyOverride = true
			}
		}
	}

	// Apply override priority: DENY > ALLOW > BASE
	if hasDenyOverride {
		result.Allowed = false
		result.OverrideType = "deny"
		result.Reason = "denied by channel override"
	} else if hasAllowOverride {
		result.Allowed = true
		result.OverrideType = "allow"
		result.Reason = "allowed by channel override"
	} else if baseAllowed {
		result.Reason = "granted by role permissions"
	} else {
		result.Reason = "not granted by role permissions"
	}

	return result
}

// ══════════════════════════════════════════════════════════════════════════════
// PERMISSION VALIDATION
// ══════════════════════════════════════════════════════════════════════════════

// IsValidPermission checks if a permission code is valid
func (pr *PermissionResolver) IsValidPermission(permission string) bool {
	validPermissions := []string{
		domain.PermissionAdministrator,
		domain.PermissionManageServer,
		domain.PermissionManageRoles,
		domain.PermissionManageChannels,
		domain.PermissionManageSections,
		domain.PermissionKickMembers,
		domain.PermissionBanMembers,
		domain.PermissionInviteMembers,
		domain.PermissionChangeNickname,
		domain.PermissionManageNicknames,
		domain.PermissionViewChannels,
		domain.PermissionSendMessages,
		domain.PermissionSendMessagesInThreads,
		domain.PermissionCreateThreads,
		domain.PermissionEmbedLinks,
		domain.PermissionAttachFiles,
		domain.PermissionAddReactions,
		domain.PermissionUseExternalEmojis,
		domain.PermissionMentionEveryone,
		domain.PermissionManageMessages,
		domain.PermissionManageThreads,
		domain.PermissionReadMessageHistory,
		domain.PermissionConnect,
		domain.PermissionSpeak,
		domain.PermissionVideo,
		domain.PermissionMuteMembers,
		domain.PermissionDeafenMembers,
		domain.PermissionMoveMembers,
	}

	for _, valid := range validPermissions {
		if valid == permission {
			return true
		}
	}

	return false
}
