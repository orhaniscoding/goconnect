package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// SERVER DISCOVERY REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// DiscoveryFilter defines filters for searching discoverable servers
type DiscoveryFilter struct {
	Query    string
	Category *domain.DiscoveryCategory
	Tags     []string
	Featured *bool
	Verified *bool
	Sort     string // member_count, online_count, created_at
	Limit    int
	Cursor   string
}

// DiscoveryRepository defines operations for server discovery
type DiscoveryRepository interface {
	// Upsert creates or updates discovery settings for a tenant
	Upsert(ctx context.Context, discovery *domain.ServerDiscovery) error

	// GetByTenantID retrieves discovery settings for a tenant
	GetByTenantID(ctx context.Context, tenantID string) (*domain.ServerDiscovery, error)

	// Search searches discoverable servers
	Search(ctx context.Context, filter DiscoveryFilter) ([]domain.ServerDiscovery, string, error)

	// GetFeatured retrieves featured servers
	GetFeatured(ctx context.Context, limit int) ([]domain.ServerDiscovery, error)

	// UpdateStats updates cached member/online counts
	UpdateStats(ctx context.Context, tenantID string, memberCount, onlineCount int) error

	// SetFeatured sets the featured status for a server
	SetFeatured(ctx context.Context, tenantID string, featured bool) error

	// SetVerified sets the verified status for a server
	SetVerified(ctx context.Context, tenantID string, verified bool) error

	// Delete removes discovery settings (makes server non-discoverable)
	Delete(ctx context.Context, tenantID string) error
}

// ═══════════════════════════════════════════════════════════════════════════
// VANITY URL REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// VanityURLRepository defines operations for vanity URLs
type VanityURLRepository interface {
	// Set sets a vanity URL for a tenant
	Set(ctx context.Context, vanity *domain.ServerVanityURL) error

	// GetByTenantID retrieves the vanity URL for a tenant
	GetByTenantID(ctx context.Context, tenantID string) (*domain.ServerVanityURL, error)

	// GetByCode retrieves the tenant for a vanity code
	GetByCode(ctx context.Context, code string) (*domain.ServerVanityURL, error)

	// Delete removes a vanity URL
	Delete(ctx context.Context, tenantID string) error

	// IsAvailable checks if a vanity code is available
	IsAvailable(ctx context.Context, code string) (bool, error)
}

// ═══════════════════════════════════════════════════════════════════════════
// COMBINED REPOSITORY INTERFACE
// ═══════════════════════════════════════════════════════════════════════════

// Repositories aggregates all repository interfaces
type Repositories struct {
	Section     SectionRepository
	Channel     ChannelRepository
	Role        RoleRepository
	Permission  PermissionRepository
	Message     MessageRepository
	Reaction    ReactionRepository
	VoiceState  VoiceStateRepository
	Presence    PresenceRepository
	Friendship  FriendshipRepository
	DMChannel   DMChannelRepository
	DMMessage   DMMessageRepository
	Discovery   DiscoveryRepository
	VanityURL   VanityURLRepository
}
