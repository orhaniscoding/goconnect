package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// TenantRole defines the role of a user within a tenant
type TenantRole string

const (
	TenantRoleOwner     TenantRole = "owner"     // Full control, can delete tenant
	TenantRoleAdmin     TenantRole = "admin"     // Manage users, networks, settings
	TenantRoleModerator TenantRole = "moderator" // Manage chat, announcements
	TenantRoleVIP       TenantRole = "vip"       // Premium access to restricted networks
	TenantRoleMember    TenantRole = "member"    // Basic access
)

// TenantRoleHierarchy returns the priority of a role (higher = more permissions)
func TenantRoleHierarchy(role TenantRole) int {
	switch role {
	case TenantRoleOwner:
		return 100
	case TenantRoleAdmin:
		return 80
	case TenantRoleModerator:
		return 60
	case TenantRoleVIP:
		return 40
	case TenantRoleMember:
		return 20
	default:
		return 0
	}
}

// HasPermission checks if a role has at least the required permission level
func (r TenantRole) HasPermission(required TenantRole) bool {
	return TenantRoleHierarchy(r) >= TenantRoleHierarchy(required)
}

// TenantVisibility defines who can see/find the tenant
type TenantVisibility string

const (
	TenantVisibilityPublic   TenantVisibility = "public"   // Discoverable in search/list
	TenantVisibilityUnlisted TenantVisibility = "unlisted" // Accessible by link only
	TenantVisibilityPrivate  TenantVisibility = "private"  // Invite only
)

// TenantAccessType defines how users can join the tenant
type TenantAccessType string

const (
	TenantAccessOpen       TenantAccessType = "open"        // Anyone can join
	TenantAccessPassword   TenantAccessType = "password"    // Requires password
	TenantAccessInviteOnly TenantAccessType = "invite_only" // Requires invite code
)

// TenantMember represents a user's membership in a tenant (N:N relationship)
type TenantMember struct {
	ID        string     `json:"id" db:"id"`
	TenantID  string     `json:"tenant_id" db:"tenant_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Role      TenantRole `json:"role" db:"role"`
	Nickname  string     `json:"nickname,omitempty" db:"nickname"` // Tenant-specific display name
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	// Enriched fields (not in DB)
	User *User `json:"user,omitempty" db:"-"` // Populated on list queries
}

// TenantInvite represents a tenant-level invitation (Steam-like codes)
type TenantInvite struct {
	ID        string     `json:"id" db:"id"`
	TenantID  string     `json:"tenant_id" db:"tenant_id"`
	Code      string     `json:"code" db:"code"` // Short code like "ABC123XY"
	MaxUses   int        `json:"max_uses" db:"max_uses"`
	UseCount  int        `json:"use_count" db:"use_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedBy string     `json:"created_by" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// TenantAnnouncement represents admin announcements in a tenant
type TenantAnnouncement struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	AuthorID  string    `json:"author_id" db:"author_id"`
	IsPinned  bool      `json:"is_pinned" db:"is_pinned"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// Enriched fields
	Author *User `json:"author,omitempty" db:"-"`
}

// TenantChatMessage represents a message in tenant's general chat
type TenantChatMessage struct {
	ID        string     `json:"id" db:"id"`
	TenantID  string     `json:"tenant_id" db:"tenant_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Content   string     `json:"content" db:"content"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	// Enriched fields
	User *User `json:"user,omitempty" db:"-"`
}

// ---- Extended Tenant struct (replaces basic one in auth.go) ----

// TenantExtended represents a tenant with multi-membership fields
type TenantExtended struct {
	ID           string           `json:"id" db:"id"`
	Name         string           `json:"name" db:"name"`
	Description  string           `json:"description" db:"description"`
	IconURL      string           `json:"icon_url,omitempty" db:"icon_url"`
	Visibility   TenantVisibility `json:"visibility" db:"visibility"`
	AccessType   TenantAccessType `json:"access_type" db:"access_type"`
	PasswordHash string           `json:"-" db:"password_hash"` // For password access
	MaxMembers   int              `json:"max_members" db:"max_members"`
	OwnerID      string           `json:"owner_id" db:"owner_id"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
	// Computed fields
	MemberCount int `json:"member_count" db:"member_count"`
}

// ---- Request/Response Types ----

// CreateTenantRequest for creating a new tenant
type CreateTenantRequest struct {
	Name        string           `json:"name" binding:"required,min=3,max=64"`
	Description string           `json:"description" binding:"max=500"`
	Visibility  TenantVisibility `json:"visibility" binding:"required,oneof=public unlisted private"`
	AccessType  TenantAccessType `json:"access_type" binding:"required,oneof=open password invite_only"`
	Password    string           `json:"password,omitempty" binding:"omitempty,min=4,max=64"`
	MaxMembers  int              `json:"max_members,omitempty" binding:"omitempty,min=0,max=10000"`
}

// UpdateTenantRequest for updating tenant settings
type UpdateTenantRequest struct {
	Name        *string           `json:"name,omitempty" binding:"omitempty,min=3,max=64"`
	Description *string           `json:"description,omitempty" binding:"omitempty,max=500"`
	Visibility  *TenantVisibility `json:"visibility,omitempty" binding:"omitempty,oneof=public unlisted private"`
	AccessType  *TenantAccessType `json:"access_type,omitempty" binding:"omitempty,oneof=open password invite_only"`
	Password    *string           `json:"password,omitempty" binding:"omitempty,min=4,max=64"`
	MaxMembers  *int              `json:"max_members,omitempty" binding:"omitempty,min=0,max=10000"`
}

// JoinTenantRequest for joining a tenant
type JoinTenantRequest struct {
	Password string `json:"password,omitempty"` // For password-protected tenants
}

// JoinByCodeRequest for joining via invite code
type JoinByCodeRequest struct {
	Code string `json:"code" binding:"required,min=6,max=12"`
}

// UpdateMemberRoleRequest for updating a member's role
type UpdateMemberRoleRequest struct {
	Role TenantRole `json:"role" binding:"required,oneof=admin moderator vip member"`
}

// CreateTenantInviteRequest for creating tenant invites
type CreateTenantInviteRequest struct {
	MaxUses   int `json:"max_uses" binding:"omitempty,min=0,max=1000"`
	ExpiresIn int `json:"expires_in" binding:"omitempty,min=3600,max=2592000"` // 1 hour to 30 days in seconds
}

// CreateAnnouncementRequest for creating announcements
type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required,min=3,max=200"`
	Content  string `json:"content" binding:"required,min=1,max=5000"`
	IsPinned bool   `json:"is_pinned"`
}

// UpdateAnnouncementRequest for updating announcements
type UpdateAnnouncementRequest struct {
	Title    *string `json:"title,omitempty" binding:"omitempty,min=3,max=200"`
	Content  *string `json:"content,omitempty" binding:"omitempty,min=1,max=5000"`
	IsPinned *bool   `json:"is_pinned,omitempty"`
}

// SendChatMessageRequest for sending chat messages
type SendChatMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// ListTenantsRequest for listing/searching tenants
type ListTenantsRequest struct {
	Search string `form:"search"`           // Search by name
	Limit  int    `form:"limit,default=20"` // Max 100
	Cursor string `form:"cursor"`           // Pagination cursor
}

// ListMembersRequest for listing tenant members
type ListTenantMembersRequest struct {
	Role   string `form:"role"`             // Filter by role
	Limit  int    `form:"limit,default=50"` // Max 100
	Cursor string `form:"cursor"`
}

// ListAnnouncementsRequest for listing announcements
type ListAnnouncementsRequest struct {
	Pinned *bool  `form:"pinned"`           // Filter pinned only
	Limit  int    `form:"limit,default=20"` // Max 50
	Cursor string `form:"cursor"`
}

// ListChatMessagesRequest for listing chat messages
type ListChatMessagesRequest struct {
	Before string `form:"before"`           // Get messages before this ID
	Limit  int    `form:"limit,default=50"` // Max 100
}

// ---- Helper Functions ----

// GenerateTenantInviteCode generates a short, easy-to-share code
func GenerateTenantInviteCode() (string, error) {
	// Character set: uppercase letters and numbers, excluding confusing ones (0, O, I, 1, L)
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	const codeLength = 8

	code := make([]byte, codeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate invite code: %w", err)
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

// GenerateTenantMemberID generates an ID for tenant members
func GenerateTenantMemberID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("tm_%s", hex.EncodeToString(bytes))
}

// GenerateTenantInviteID generates an ID for tenant invites
func GenerateTenantInviteID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("ti_%s", hex.EncodeToString(bytes))
}

// GenerateAnnouncementID generates an ID for announcements
func GenerateAnnouncementID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("ann_%s", hex.EncodeToString(bytes))
}

// GenerateChatMessageID generates an ID for chat messages
func GenerateChatMessageID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("msg_%s", hex.EncodeToString(bytes))
}

// IsValid checks if the tenant invite is still valid
func (ti *TenantInvite) IsValid() bool {
	if ti.RevokedAt != nil {
		return false
	}
	if ti.ExpiresAt != nil && time.Now().After(*ti.ExpiresAt) {
		return false
	}
	if ti.MaxUses > 0 && ti.UseCount >= ti.MaxUses {
		return false
	}
	return true
}

// CanJoin checks if the tenant can accept new members
func (t *TenantExtended) CanJoin() bool {
	if t.MaxMembers == 0 {
		return true // Unlimited
	}
	return t.MemberCount < t.MaxMembers
}
