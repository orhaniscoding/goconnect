package service

import (
	"context"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// PERMISSION HELPER FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════
// Utility functions for common permission patterns and checks

// PermissionHelper provides utility functions for permission management
type PermissionHelper struct {
	permissionResolver *PermissionResolver
	roleRepo           repository.RoleRepository
	channelRepo        repository.ChannelRepository
}

// NewPermissionHelper creates a new permission helper instance
func NewPermissionHelper(
	permissionResolver *PermissionResolver,
	roleRepo repository.RoleRepository,
	channelRepo repository.ChannelRepository,
) *PermissionHelper {
	return &PermissionHelper{
		permissionResolver: permissionResolver,
		roleRepo:           roleRepo,
		channelRepo:        channelRepo,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// QUICK CHECK FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

// CanManageChannel checks if user can manage a specific channel
func (h *PermissionHelper) CanManageChannel(
	ctx context.Context,
	userID, tenantID, channelID string,
) (bool, error) {
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		channelID,
		domain.PermissionManageChannels,
	)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CanSendMessages checks if user can send messages in a channel
func (h *PermissionHelper) CanSendMessages(
	ctx context.Context,
	userID, tenantID, channelID string,
) (bool, error) {
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		channelID,
		domain.PermissionSendMessages,
	)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// CanDeleteMessage checks if user can delete a specific message
// Returns true if user has MANAGE_MESSAGES or is the message author
func (h *PermissionHelper) CanDeleteMessage(
	ctx context.Context,
	userID, messageAuthorID, tenantID, channelID string,
) (bool, string, error) {
	// Check if user is the message author
	if userID == messageAuthorID {
		return true, "message owner", nil
	}

	// Check if user has MANAGE_MESSAGES permission
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		channelID,
		domain.PermissionManageMessages,
	)
	if err != nil {
		return false, "", err
	}

	return result.Allowed, result.Reason, nil
}

// CanKickMember checks if user can kick members from tenant
func (h *PermissionHelper) CanKickMember(
	ctx context.Context,
	userID, tenantID string,
) (bool, error) {
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"", // No channel context
		domain.PermissionKickMembers,
	)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// IsAdministrator checks if user has ADMINISTRATOR permission in tenant
func (h *PermissionHelper) IsAdministrator(
	ctx context.Context,
	userID, tenantID string,
) (bool, error) {
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"", // No channel context
		domain.PermissionAdministrator,
	)
	if err != nil {
		return false, err
	}
	return result.Allowed, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// BATCH PERMISSION CHECKS
// ══════════════════════════════════════════════════════════════════════════════

// GetUserChannelPermissions retrieves all permissions for a user in a channel
// Returns a map of permission -> allowed status
func (h *PermissionHelper) GetUserChannelPermissions(
	ctx context.Context,
	userID, tenantID, channelID string,
) (map[string]bool, error) {
	allPermissions := []string{
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
	}

	results, err := h.permissionResolver.CheckMultiplePermissions(
		ctx,
		userID,
		tenantID,
		channelID,
		allPermissions,
	)
	if err != nil {
		return nil, err
	}

	// Convert to simple map
	permMap := make(map[string]bool)
	for perm, result := range results {
		permMap[perm] = result.Allowed
	}

	return permMap, nil
}

// GetUserVoicePermissions retrieves all voice permissions for a user in a channel
func (h *PermissionHelper) GetUserVoicePermissions(
	ctx context.Context,
	userID, tenantID, channelID string,
) (map[string]bool, error) {
	voicePermissions := []string{
		domain.PermissionConnect,
		domain.PermissionSpeak,
		domain.PermissionVideo,
		domain.PermissionMuteMembers,
		domain.PermissionDeafenMembers,
		domain.PermissionMoveMembers,
	}

	results, err := h.permissionResolver.CheckMultiplePermissions(
		ctx,
		userID,
		tenantID,
		channelID,
		voicePermissions,
	)
	if err != nil {
		return nil, err
	}

	// Convert to simple map
	permMap := make(map[string]bool)
	for perm, result := range results {
		permMap[perm] = result.Allowed
	}

	return permMap, nil
}

// GetUserServerPermissions retrieves all server-level permissions for a user
func (h *PermissionHelper) GetUserServerPermissions(
	ctx context.Context,
	userID, tenantID string,
) (map[string]bool, error) {
	serverPermissions := []string{
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
	}

	results, err := h.permissionResolver.CheckMultiplePermissions(
		ctx,
		userID,
		tenantID,
		"", // No channel context
		serverPermissions,
	)
	if err != nil {
		return nil, err
	}

	// Convert to simple map
	permMap := make(map[string]bool)
	for perm, result := range results {
		permMap[perm] = result.Allowed
	}

	return permMap, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// ROLE MANAGEMENT HELPERS
// ══════════════════════════════════════════════════════════════════════════════

// CanManageRole checks if user can manage a specific role
// Rules:
// - ADMINISTRATOR can manage all roles
// - Users can only manage roles with lower position than their highest role
func (h *PermissionHelper) CanManageRole(
	ctx context.Context,
	userID, tenantID, targetRoleID string,
) (bool, string, error) {
	// Check if user is administrator
	isAdmin, err := h.IsAdministrator(ctx, userID, tenantID)
	if err != nil {
		return false, "", err
	}
	if isAdmin {
		return true, "administrator bypass", nil
	}

	// Check if user has MANAGE_ROLES permission
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"",
		domain.PermissionManageRoles,
	)
	if err != nil {
		return false, "", err
	}

	if !result.Allowed {
		return false, "no MANAGE_ROLES permission", nil
	}

	// Get user's roles
	userRoles, err := h.roleRepo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(userRoles) == 0 {
		return false, "user has no roles", nil
	}

	// Get target role
	targetRole, err := h.roleRepo.GetByID(ctx, targetRoleID)
	if err != nil {
		return false, "", fmt.Errorf("failed to get target role: %w", err)
	}

	// Find user's highest role position
	highestPosition := -1
	for _, role := range userRoles {
		if role.Position > highestPosition {
			highestPosition = role.Position
		}
	}

	// User can only manage roles with lower position
	if highestPosition > targetRole.Position {
		return true, "higher role position", nil
	}

	return false, "target role position is equal or higher", nil
}

// CanAssignRole checks if user can assign a role to another user
func (h *PermissionHelper) CanAssignRole(
	ctx context.Context,
	userID, tenantID, roleID string,
) (bool, string, error) {
	return h.CanManageRole(ctx, userID, tenantID, roleID)
}

// GetHighestRole returns the user's highest positioned role in a tenant
func (h *PermissionHelper) GetHighestRole(
	ctx context.Context,
	userID, tenantID string,
) (*domain.Role, error) {
	roles, err := h.roleRepo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("user has no roles in tenant")
	}

	highestRole := &roles[0]
	for i := range roles {
		if roles[i].Position > highestRole.Position {
			highestRole = &roles[i]
		}
	}

	return highestRole, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// PERMISSION DEBUGGING HELPERS
// ══════════════════════════════════════════════════════════════════════════════

// ExplainPermission provides detailed explanation of why a permission was granted/denied
// Useful for debugging and audit logs
func (h *PermissionHelper) ExplainPermission(
	ctx context.Context,
	userID, tenantID, channelID, permission string,
) (string, error) {
	result, err := h.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		channelID,
		permission,
	)
	if err != nil {
		return "", err
	}

	explanation := fmt.Sprintf("Permission: %s\n", permission)
	explanation += fmt.Sprintf("User: %s\n", userID)
	explanation += fmt.Sprintf("Tenant: %s\n", tenantID)
	if channelID != "" {
		explanation += fmt.Sprintf("Channel: %s\n", channelID)
	}
	explanation += fmt.Sprintf("Result: %v\n", result.Allowed)
	explanation += fmt.Sprintf("Base Allowed: %v\n", result.BaseAllowed)
	explanation += fmt.Sprintf("Override Type: %s\n", result.OverrideType)
	explanation += fmt.Sprintf("Reason: %s\n", result.Reason)

	return explanation, nil
}

// ListUserPermissions returns a formatted list of all user permissions
func (h *PermissionHelper) ListUserPermissions(
	ctx context.Context,
	userID, tenantID, channelID string,
) (map[string]interface{}, error) {
	// Get all permission categories
	serverPerms, err := h.GetUserServerPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"user_id":            userID,
		"tenant_id":          tenantID,
		"server_permissions": serverPerms,
	}

	// If channel specified, get channel permissions
	if channelID != "" {
		channelPerms, err := h.GetUserChannelPermissions(ctx, userID, tenantID, channelID)
		if err != nil {
			return nil, err
		}

		voicePerms, err := h.GetUserVoicePermissions(ctx, userID, tenantID, channelID)
		if err != nil {
			return nil, err
		}

		result["channel_id"] = channelID
		result["channel_permissions"] = channelPerms
		result["voice_permissions"] = voicePerms
	}

	return result, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// VALIDATION HELPERS
// ══════════════════════════════════════════════════════════════════════════════

// ValidatePermissionList checks if all permissions in a list are valid
func (h *PermissionHelper) ValidatePermissionList(permissions []string) (bool, []string) {
	invalid := []string{}

	for _, perm := range permissions {
		if !h.permissionResolver.IsValidPermission(perm) {
			invalid = append(invalid, perm)
		}
	}

	return len(invalid) == 0, invalid
}

// GetAllPermissions returns all available permissions grouped by category
func (h *PermissionHelper) GetAllPermissions() map[string][]string {
	return map[string][]string{
		"administrator": {
			domain.PermissionAdministrator,
		},
		"server": {
			domain.PermissionManageServer,
			domain.PermissionManageRoles,
			domain.PermissionManageChannels,
			domain.PermissionManageSections,
			domain.PermissionKickMembers,
			domain.PermissionBanMembers,
			domain.PermissionInviteMembers,
			domain.PermissionChangeNickname,
			domain.PermissionManageNicknames,
		},
		"channel": {
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
		},
		"voice": {
			domain.PermissionConnect,
			domain.PermissionSpeak,
			domain.PermissionVideo,
			domain.PermissionMuteMembers,
			domain.PermissionDeafenMembers,
			domain.PermissionMoveMembers,
		},
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CONVENIENCE FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

// RequirePermissions checks multiple permissions and returns error if any denied
// Useful for handler functions that need to enforce multiple permissions
func (h *PermissionHelper) RequirePermissions(
	ctx context.Context,
	userID, tenantID, channelID string,
	permissions ...string,
) error {
	results, err := h.permissionResolver.CheckMultiplePermissions(
		ctx,
		userID,
		tenantID,
		channelID,
		permissions,
	)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	denied := []string{}
	for perm, result := range results {
		if !result.Allowed {
			denied = append(denied, perm)
		}
	}

	if len(denied) > 0 {
		return fmt.Errorf("missing permissions: %v", denied)
	}

	return nil
}

// HasAnyPermission checks if user has at least one of the specified permissions
func (h *PermissionHelper) HasAnyPermission(
	ctx context.Context,
	userID, tenantID, channelID string,
	permissions ...string,
) (bool, error) {
	results, err := h.permissionResolver.CheckMultiplePermissions(
		ctx,
		userID,
		tenantID,
		channelID,
		permissions,
	)
	if err != nil {
		return false, err
	}

	for _, result := range results {
		if result.Allowed {
			return true, nil
		}
	}

	return false, nil
}
