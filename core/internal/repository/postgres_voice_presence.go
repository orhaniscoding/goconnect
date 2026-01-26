package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES VOICE STATE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresVoiceStateRepository implements VoiceStateRepository for PostgreSQL
type PostgresVoiceStateRepository struct {
	db *sql.DB
}

// NewPostgresVoiceStateRepository creates a new PostgresVoiceStateRepository
func NewPostgresVoiceStateRepository(db *sql.DB) *PostgresVoiceStateRepository {
	return &PostgresVoiceStateRepository{db: db}
}

// Join adds a user to a voice channel
func (r *PostgresVoiceStateRepository) Join(ctx context.Context, state *domain.VoiceState) error {
	query := `
		INSERT INTO voice_states (user_id, channel_id, self_mute, self_deaf, server_mute, server_deaf, connected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			channel_id = EXCLUDED.channel_id,
			self_mute = EXCLUDED.self_mute,
			self_deaf = EXCLUDED.self_deaf,
			connected_at = EXCLUDED.connected_at
	`

	if state.ConnectedAt.IsZero() {
		state.ConnectedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		state.UserID,
		state.ChannelID,
		state.SelfMute,
		state.SelfDeaf,
		state.ServerMute,
		state.ServerDeaf,
		state.ConnectedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	return nil
}

// Leave removes a user from their voice channel
func (r *PostgresVoiceStateRepository) Leave(ctx context.Context, userID string) error {
	query := `DELETE FROM voice_states WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to leave voice channel: %w", err)
	}

	return nil
}

// Update updates a user's voice state
func (r *PostgresVoiceStateRepository) Update(ctx context.Context, state *domain.VoiceState) error {
	query := `
		UPDATE voice_states
		SET self_mute = $2, self_deaf = $3, server_mute = $4, server_deaf = $5
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		state.UserID,
		state.SelfMute,
		state.SelfDeaf,
		state.ServerMute,
		state.ServerDeaf,
	)

	if err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("voice state not found")
	}

	return nil
}

// GetByUser retrieves a user's current voice state
func (r *PostgresVoiceStateRepository) GetByUser(ctx context.Context, userID string) (*domain.VoiceState, error) {
	query := `
		SELECT user_id, channel_id, self_mute, self_deaf, server_mute, server_deaf, connected_at
		FROM voice_states
		WHERE user_id = $1
	`

	var state domain.VoiceState
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&state.UserID,
		&state.ChannelID,
		&state.SelfMute,
		&state.SelfDeaf,
		&state.ServerMute,
		&state.ServerDeaf,
		&state.ConnectedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get voice state: %w", err)
	}

	return &state, nil
}

// GetByChannel retrieves all users in a voice channel
func (r *PostgresVoiceStateRepository) GetByChannel(ctx context.Context, channelID string) ([]domain.VoiceState, error) {
	query := `
		SELECT user_id, channel_id, self_mute, self_deaf, server_mute, server_deaf, connected_at
		FROM voice_states
		WHERE channel_id = $1
		ORDER BY connected_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get voice states: %w", err)
	}
	defer rows.Close()

	var states []domain.VoiceState
	for rows.Next() {
		var state domain.VoiceState
		if err := rows.Scan(
			&state.UserID,
			&state.ChannelID,
			&state.SelfMute,
			&state.SelfDeaf,
			&state.ServerMute,
			&state.ServerDeaf,
			&state.ConnectedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan voice state: %w", err)
		}
		states = append(states, state)
	}

	return states, nil
}

// MoveUser moves a user to a different channel
func (r *PostgresVoiceStateRepository) MoveUser(ctx context.Context, userID, newChannelID string) error {
	query := `
		UPDATE voice_states
		SET channel_id = $2
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, newChannelID)
	if err != nil {
		return fmt.Errorf("failed to move user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not in voice channel")
	}

	return nil
}

// Ensure interface compliance
var _ VoiceStateRepository = (*PostgresVoiceStateRepository)(nil)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES PRESENCE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresPresenceRepository implements PresenceRepository for PostgreSQL
type PostgresPresenceRepository struct {
	db *sql.DB
}

// NewPostgresPresenceRepository creates a new PostgresPresenceRepository
func NewPostgresPresenceRepository(db *sql.DB) *PostgresPresenceRepository {
	return &PostgresPresenceRepository{db: db}
}

// Upsert creates or updates a user's presence
func (r *PostgresPresenceRepository) Upsert(ctx context.Context, presence *domain.UserPresence) error {
	query := `
		INSERT INTO user_presence (user_id, status, custom_status, activity_type, activity_name, last_seen, desktop_status, mobile_status, web_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			status = EXCLUDED.status,
			custom_status = EXCLUDED.custom_status,
			activity_type = EXCLUDED.activity_type,
			activity_name = EXCLUDED.activity_name,
			last_seen = EXCLUDED.last_seen,
			desktop_status = EXCLUDED.desktop_status,
			mobile_status = EXCLUDED.mobile_status,
			web_status = EXCLUDED.web_status
	`

	if presence.LastSeen.IsZero() {
		presence.LastSeen = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		presence.UserID,
		presence.Status,
		presence.CustomStatus,
		presence.ActivityType,
		presence.ActivityName,
		presence.LastSeen,
		presence.DesktopStatus,
		presence.MobileStatus,
		presence.WebStatus,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert presence: %w", err)
	}

	return nil
}

// GetByUser retrieves a user's presence
func (r *PostgresPresenceRepository) GetByUser(ctx context.Context, userID string) (*domain.UserPresence, error) {
	query := `
		SELECT user_id, status, custom_status, activity_type, activity_name, last_seen, desktop_status, mobile_status, web_status
		FROM user_presence
		WHERE user_id = $1
	`

	var presence domain.UserPresence
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&presence.UserID,
		&presence.Status,
		&presence.CustomStatus,
		&presence.ActivityType,
		&presence.ActivityName,
		&presence.LastSeen,
		&presence.DesktopStatus,
		&presence.MobileStatus,
		&presence.WebStatus,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}

	return &presence, nil
}

// GetByUsers retrieves presence for multiple users
func (r *PostgresPresenceRepository) GetByUsers(ctx context.Context, userIDs []string) ([]domain.UserPresence, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT user_id, status, custom_status, activity_type, activity_name, last_seen, desktop_status, mobile_status, web_status
		FROM user_presence
		WHERE user_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get presences: %w", err)
	}
	defer rows.Close()

	var presences []domain.UserPresence
	for rows.Next() {
		var presence domain.UserPresence
		if err := rows.Scan(
			&presence.UserID,
			&presence.Status,
			&presence.CustomStatus,
			&presence.ActivityType,
			&presence.ActivityName,
			&presence.LastSeen,
			&presence.DesktopStatus,
			&presence.MobileStatus,
			&presence.WebStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan presence: %w", err)
		}
		presences = append(presences, presence)
	}

	return presences, nil
}

// GetOnlineInTenant retrieves online users in a tenant
func (r *PostgresPresenceRepository) GetOnlineInTenant(ctx context.Context, tenantID string, limit int) ([]domain.UserPresence, error) {
	query := `
		SELECT p.user_id, p.status, p.custom_status, p.activity_type, p.activity_name, p.last_seen, p.desktop_status, p.mobile_status, p.web_status
		FROM user_presence p
		INNER JOIN tenant_members tm ON p.user_id = tm.user_id
		WHERE tm.tenant_id = $1 AND p.status != 'offline'
		ORDER BY p.last_seen DESC
		LIMIT $2
	`

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}
	defer rows.Close()

	var presences []domain.UserPresence
	for rows.Next() {
		var presence domain.UserPresence
		if err := rows.Scan(
			&presence.UserID,
			&presence.Status,
			&presence.CustomStatus,
			&presence.ActivityType,
			&presence.ActivityName,
			&presence.LastSeen,
			&presence.DesktopStatus,
			&presence.MobileStatus,
			&presence.WebStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan presence: %w", err)
		}
		presences = append(presences, presence)
	}

	return presences, nil
}

// UpdateLastSeen updates a user's last seen timestamp
func (r *PostgresPresenceRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	query := `
		UPDATE user_presence
		SET last_seen = $2
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

// SetOffline sets a user's status to offline
func (r *PostgresPresenceRepository) SetOffline(ctx context.Context, userID string) error {
	query := `
		UPDATE user_presence
		SET status = 'offline', desktop_status = NULL, mobile_status = NULL, web_status = NULL
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to set offline: %w", err)
	}

	return nil
}

// CleanupStale marks stale presences as offline
func (r *PostgresPresenceRepository) CleanupStale(ctx context.Context, olderThanMinutes int) (int, error) {
	query := `
		UPDATE user_presence
		SET status = 'offline', desktop_status = NULL, mobile_status = NULL, web_status = NULL
		WHERE status != 'offline' AND last_seen < NOW() - INTERVAL '%d minutes'
	`

	result, err := r.db.ExecContext(ctx, fmt.Sprintf(query, olderThanMinutes))
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup stale presences: %w", err)
	}

	count, err := result.RowsAffected()
	return int(count), nil
}

// Ensure interface compliance
var _ PresenceRepository = (*PostgresPresenceRepository)(nil)
