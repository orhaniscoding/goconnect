package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// ROLE (Custom roles with permissions)
// ═══════════════════════════════════════════════════════════════════════════

// Role represents a custom role within a tenant
type Role struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Color       *string   `json:"color,omitempty" db:"color"`
	Icon        *string   `json:"icon,omitempty" db:"icon"`
	Position    int       `json:"position" db:"position"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	IsAdmin     bool      `json:"is_admin" db:"is_admin"`
	Mentionable bool      `json:"mentionable" db:"mentionable"`
	Hoist       bool      `json:"hoist" db:"hoist"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Enriched fields
	Permissions []RolePermission `json:"permissions,omitempty" db:"-"`
	MemberCount int              `json:"member_count,omitempty" db:"-"`
}

// HasHigherPosition checks if this role is higher than another
func (r *Role) HasHigherPosition(other *Role) bool {
	return r.Position > other.Position
}

// CanManage checks if this role can manage another role
func (r *Role) CanManage(other *Role) bool {
	if r.IsAdmin {
		return true
	}
	return r.Position > other.Position
}

// CreateRoleRequest for creating a new role
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Color       *string `json:"color,omitempty" binding:"omitempty,hexcolor"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=255"`
	Mentionable bool    `json:"mentionable,omitempty"`
	Hoist       bool    `json:"hoist,omitempty"`
}

// UpdateRoleRequest for updating a role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Color       *string `json:"color,omitempty" binding:"omitempty,hexcolor"`
	Icon        *string `json:"icon,omitempty" binding:"omitempty,max=255"`
	Position    *int    `json:"position,omitempty" binding:"omitempty,min=0"`
	Mentionable *bool   `json:"mentionable,omitempty"`
	Hoist       *bool   `json:"hoist,omitempty"`
}

// GenerateRoleID generates a new role ID
func GenerateRoleID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("role_%s", hex.EncodeToString(bytes))
}

// ValidateRoleName validates and sanitizes a role name
func ValidateRoleName(name string) (string, error) {
	name = strings.TrimSpace(name)

	spaceRegex := regexp.MustCompile(`\s+`)
	name = spaceRegex.ReplaceAllString(name, " ")

	if len(name) < 1 {
		return "", NewError(ErrValidation, "role name is required", map[string]any{
			"field": "name",
			"min":   1,
		})
	}

	if len(name) > 100 {
		return "", NewError(ErrValidation, "role name must be at most 100 characters", map[string]any{
			"field": "name",
			"max":   100,
		})
	}

	return name, nil
}

// ValidateHexColor validates a hex color string
func ValidateHexColor(color string) error {
	if color == "" {
		return nil
	}

	matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
	if !matched {
		return NewError(ErrValidation, "invalid hex color format (expected #RRGGBB)", map[string]any{
			"field": "color",
		})
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// PERMISSION (Granular permission system)
// ═══════════════════════════════════════════════════════════════════════════

// PermissionCategory groups permissions
type PermissionCategory string

const (
	PermissionCategoryServer  PermissionCategory = "server"
	PermissionCategoryNetwork PermissionCategory = "network"
	PermissionCategoryChannel PermissionCategory = "channel"
	PermissionCategoryMember  PermissionCategory = "member"
	PermissionCategoryVoice   PermissionCategory = "voice"
)

// PermissionDefinition represents a permission that can be granted
type PermissionDefinition struct {
	ID           string             `json:"id" db:"id"`
	Category     PermissionCategory `json:"category" db:"category"`
	Name         string             `json:"name" db:"name"`
	Description  string             `json:"description,omitempty" db:"description"`
	DefaultValue bool               `json:"default_value" db:"default_value"`
}

// RolePermission represents a permission granted/denied to a role
type RolePermission struct {
	RoleID       string `json:"role_id" db:"role_id"`
	PermissionID string `json:"permission_id" db:"permission_id"`
	Permission   string `json:"permission" db:"-"` // Enriched: permission code
	Allowed      bool   `json:"allowed" db:"allowed"`
}

// ChannelPermissionOverride represents a per-channel permission override
type ChannelPermissionOverride struct {
	ID           string  `json:"id" db:"id"`
	ChannelID    string  `json:"channel_id" db:"channel_id"`
	RoleID       *string `json:"role_id,omitempty" db:"role_id"`
	UserID       *string `json:"user_id,omitempty" db:"user_id"`
	PermissionID string  `json:"permission_id" db:"permission_id"`
	Permission   string  `json:"permission" db:"-"` // Enriched: permission code
	Allowed      *bool   `json:"allowed" db:"allowed"` // nil = inherit, true = allow, false = deny

	// Helper fields for service layer (populated at runtime)
	AllowPermissions []string `json:"-" db:"-"` // Populated by service layer
	DenyPermissions  []string `json:"-" db:"-"` // Populated by service layer
}

// IsRoleOverride checks if this is a role-based override
func (o *ChannelPermissionOverride) IsRoleOverride() bool {
	return o.RoleID != nil
}

// IsUserOverride checks if this is a user-based override
func (o *ChannelPermissionOverride) IsUserOverride() bool {
	return o.UserID != nil
}

// UserRole represents a role assigned to a user
type UserRole struct {
	UserID     string    `json:"user_id" db:"user_id"`
	RoleID     string    `json:"role_id" db:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy *string   `json:"assigned_by,omitempty" db:"assigned_by"`

	// Enriched fields
	Role *Role `json:"role,omitempty" db:"-"`
}

// ═══════════════════════════════════════════════════════════════════════════
// PERMISSION CONSTANTS (All available permissions)
// ═══════════════════════════════════════════════════════════════════════════

const (
	// ADMINISTRATOR - Grants all permissions (bypass)
	PermissionAdministrator = "administrator"

	// Server permissions
	PermissionManageServer    = "server.manage"
	PermissionManageRoles     = "server.manage_roles"
	PermissionManageChannels  = "server.manage_channels"
	PermissionManageSections  = "server.manage_sections"
	PermissionKickMembers     = "server.kick_members"
	PermissionBanMembers      = "server.ban_members"
	PermissionInviteMembers   = "server.create_invite"
	PermissionChangeNickname  = "server.change_nickname"
	PermissionManageNicknames = "server.manage_nicknames"

	// Legacy aliases (backward compatibility)
	PermServerManage         = PermissionManageServer
	PermServerDelete         = "server.delete"
	PermServerViewAuditLog   = "server.view_audit_log"
	PermServerManageRoles    = PermissionManageRoles
	PermServerManageChannels = PermissionManageChannels
	PermServerManageSections = PermissionManageSections
	PermServerKickMembers    = PermissionKickMembers
	PermServerBanMembers     = PermissionBanMembers
	PermServerCreateInvite   = PermissionInviteMembers

	// Network permissions
	PermNetworkCreate      = "network.create"
	PermNetworkManage      = "network.manage"
	PermNetworkDelete      = "network.delete"
	PermNetworkConnect     = "network.connect"
	PermNetworkKickPeers   = "network.kick_peers"
	PermNetworkBanPeers    = "network.ban_peers"
	PermNetworkApproveJoin = "network.approve_join"

	// Channel permissions
	PermissionViewChannels           = "channel.view"
	PermissionSendMessages           = "channel.send_messages"
	PermissionSendMessagesInThreads  = "channel.send_messages_in_threads"
	PermissionCreateThreads          = "channel.create_threads"
	PermissionEmbedLinks             = "channel.embed_links"
	PermissionAttachFiles            = "channel.attach_files"
	PermissionAddReactions           = "channel.add_reactions"
	PermissionUseExternalEmojis      = "channel.use_external_emojis"
	PermissionMentionEveryone        = "channel.mention_everyone"
	PermissionManageMessages         = "channel.manage_messages"
	PermissionManageThreads          = "channel.manage_threads"
	PermissionReadMessageHistory     = "channel.read_message_history"

	// Legacy aliases
	PermChannelView            = PermissionViewChannels
	PermChannelSendMessages    = PermissionSendMessages
	PermChannelEmbedLinks      = PermissionEmbedLinks
	PermChannelAttachFiles     = PermissionAttachFiles
	PermChannelAddReactions    = PermissionAddReactions
	PermChannelMentionEveryone = PermissionMentionEveryone
	PermChannelManageMessages  = PermissionManageMessages
	PermChannelManageThreads   = PermissionManageThreads

	// Voice permissions
	PermissionConnect       = "voice.connect"
	PermissionSpeak         = "voice.speak"
	PermissionVideo         = "voice.video"
	PermissionMuteMembers   = "voice.mute_members"
	PermissionDeafenMembers = "voice.deafen_members"
	PermissionMoveMembers   = "voice.move_members"

	// Legacy aliases
	PermVoiceConnect         = PermissionConnect
	PermVoiceSpeak           = PermissionSpeak
	PermVoiceMuteMembers     = PermissionMuteMembers
	PermVoiceDeafenMembers   = PermissionDeafenMembers
	PermVoiceMoveMembers     = PermissionMoveMembers
	PermVoicePrioritySpeaker = "voice.priority_speaker"
)

// UpdateRolePermissionsRequest for bulk updating role permissions
type UpdateRolePermissionsRequest struct {
	Permissions map[string]bool `json:"permissions" binding:"required"`
}

// SetChannelOverrideRequest for setting a channel permission override
type SetChannelOverrideRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
	Allowed      *bool  `json:"allowed"` // nil to remove override (inherit)
}

// ListRolesRequest for listing roles
type ListRolesRequest struct {
	Limit  int    `form:"limit,default=50"`
	Cursor string `form:"cursor"`
}
