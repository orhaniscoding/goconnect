package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// FRIENDSHIP REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// FriendshipFilter defines filters for listing friendships
type FriendshipFilter struct {
	UserID string
	Status *domain.FriendshipStatus
	Limit  int
	Cursor string
}

// FriendshipRepository defines operations for friendships
type FriendshipRepository interface {
	// Create creates a new friendship (pending status)
	Create(ctx context.Context, friendship *domain.Friendship) error

	// GetByID retrieves a friendship by ID
	GetByID(ctx context.Context, id string) (*domain.Friendship, error)

	// GetByUsers retrieves friendship between two users (if exists)
	GetByUsers(ctx context.Context, userID, friendID string) (*domain.Friendship, error)

	// List retrieves friendships with filters
	List(ctx context.Context, filter FriendshipFilter) ([]domain.Friendship, string, error)

	// Accept accepts a pending friendship
	Accept(ctx context.Context, id string) error

	// Block blocks a user (creates or updates friendship to blocked)
	Block(ctx context.Context, userID, blockedID string) error

	// Delete removes a friendship
	Delete(ctx context.Context, id string) error

	// IsBlocked checks if userA has blocked userB
	IsBlocked(ctx context.Context, userA, userB string) (bool, error)

	// AreFriends checks if two users are friends
	AreFriends(ctx context.Context, userA, userB string) (bool, error)
}

// ═══════════════════════════════════════════════════════════════════════════
// DM CHANNEL REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// DMChannelFilter defines filters for listing DM channels
type DMChannelFilter struct {
	UserID string
	Type   *domain.DMChannelType
	Limit  int
	Cursor string
}

// DMChannelRepository defines operations for DM channels
type DMChannelRepository interface {
	// Create creates a new DM channel
	Create(ctx context.Context, channel *domain.DMChannel) error

	// GetByID retrieves a DM channel by ID
	GetByID(ctx context.Context, id string) (*domain.DMChannel, error)

	// GetOrCreateDM gets or creates a 1:1 DM channel between two users
	GetOrCreateDM(ctx context.Context, userA, userB string) (*domain.DMChannel, error)

	// List retrieves DM channels for a user
	List(ctx context.Context, filter DMChannelFilter) ([]domain.DMChannel, string, error)

	// AddMember adds a member to a group DM
	AddMember(ctx context.Context, channelID, userID string) error

	// RemoveMember removes a member from a group DM
	RemoveMember(ctx context.Context, channelID, userID string) error

	// GetMembers retrieves members of a DM channel
	GetMembers(ctx context.Context, channelID string) ([]domain.DMChannelMember, error)

	// IsMember checks if a user is a member of a DM channel
	IsMember(ctx context.Context, channelID, userID string) (bool, error)

	// SetMuted sets mute status for a user in a DM channel
	SetMuted(ctx context.Context, channelID, userID string, muted bool) error
}

// ═══════════════════════════════════════════════════════════════════════════
// DM MESSAGE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// DMMessageFilter defines filters for listing DM messages
type DMMessageFilter struct {
	ChannelID string
	Before    string
	After     string
	Limit     int
}

// DMMessageRepository defines operations for DM messages
type DMMessageRepository interface {
	// Create creates a new DM message
	Create(ctx context.Context, message *domain.DMMessage) error

	// GetByID retrieves a DM message by ID
	GetByID(ctx context.Context, id string) (*domain.DMMessage, error)

	// List retrieves DM messages with filters
	List(ctx context.Context, filter DMMessageFilter) ([]domain.DMMessage, error)

	// Update updates a DM message (for edits)
	Update(ctx context.Context, message *domain.DMMessage) error

	// Delete soft-deletes a DM message
	Delete(ctx context.Context, id string) error

	// GetLatest retrieves the latest message in a DM channel
	GetLatest(ctx context.Context, channelID string) (*domain.DMMessage, error)
}
