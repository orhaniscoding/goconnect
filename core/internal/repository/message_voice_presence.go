package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// MESSAGE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// MessageFilter defines filters for listing messages
type MessageFilter struct {
	ChannelID string
	Before    string // Get messages before this ID
	After     string // Get messages after this ID
	AuthorID  *string
	Pinned    *bool
	Limit     int
}

// MessageRepository defines operations for messages
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *domain.Message) error

	// GetByID retrieves a message by ID
	GetByID(ctx context.Context, id string) (*domain.Message, error)

	// List retrieves messages with filters
	List(ctx context.Context, filter MessageFilter) ([]domain.Message, error)

	// Update updates a message (for edits)
	Update(ctx context.Context, message *domain.Message) error

	// Delete soft-deletes a message
	Delete(ctx context.Context, id string) error

	// Pin pins/unpins a message
	SetPinned(ctx context.Context, id string, pinned bool) error

	// GetPinned retrieves all pinned messages in a channel
	GetPinned(ctx context.Context, channelID string) ([]domain.Message, error)

	// Search searches messages (full-text)
	Search(ctx context.Context, channelID, query string, limit int) ([]domain.Message, error)
}

// ═══════════════════════════════════════════════════════════════════════════
// REACTION REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// ReactionRepository defines operations for reactions
type ReactionRepository interface {
	// Add adds a reaction to a message
	Add(ctx context.Context, reaction *domain.Reaction) error

	// Remove removes a reaction from a message
	Remove(ctx context.Context, messageID, userID, emoji string) error

	// GetByMessage retrieves all reactions for a message
	GetByMessage(ctx context.Context, messageID string) ([]domain.Reaction, error)

	// GetSummary retrieves aggregated reaction counts for a message
	GetSummary(ctx context.Context, messageID string) ([]domain.ReactionSummary, error)

	// GetUsers retrieves users who reacted with a specific emoji
	GetUsers(ctx context.Context, messageID, emoji string, limit int) ([]domain.User, error)
}

// ═══════════════════════════════════════════════════════════════════════════
// VOICE STATE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// VoiceStateRepository defines operations for voice states
type VoiceStateRepository interface {
	// Join adds a user to a voice channel
	Join(ctx context.Context, state *domain.VoiceState) error

	// Leave removes a user from their voice channel
	Leave(ctx context.Context, userID string) error

	// Update updates a user's voice state
	Update(ctx context.Context, state *domain.VoiceState) error

	// GetByUser retrieves a user's current voice state
	GetByUser(ctx context.Context, userID string) (*domain.VoiceState, error)

	// GetByChannel retrieves all users in a voice channel
	GetByChannel(ctx context.Context, channelID string) ([]domain.VoiceState, error)

	// MoveUser moves a user to a different channel
	MoveUser(ctx context.Context, userID, newChannelID string) error
}

// ═══════════════════════════════════════════════════════════════════════════
// PRESENCE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PresenceRepository defines operations for user presence
type PresenceRepository interface {
	// Upsert creates or updates a user's presence
	Upsert(ctx context.Context, presence *domain.UserPresence) error

	// GetByUser retrieves a user's presence
	GetByUser(ctx context.Context, userID string) (*domain.UserPresence, error)

	// GetByUsers retrieves presence for multiple users
	GetByUsers(ctx context.Context, userIDs []string) ([]domain.UserPresence, error)

	// GetOnlineInTenant retrieves online users in a tenant
	GetOnlineInTenant(ctx context.Context, tenantID string, limit int) ([]domain.UserPresence, error)

	// UpdateLastSeen updates a user's last seen timestamp
	UpdateLastSeen(ctx context.Context, userID string) error

	// SetOffline sets a user's status to offline
	SetOffline(ctx context.Context, userID string) error

	// CleanupStale marks stale presences as offline
	CleanupStale(ctx context.Context, olderThan int) (int, error)
}
