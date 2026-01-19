package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES CHANNEL REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresChannelRepository implements ChannelRepository for PostgreSQL
type PostgresChannelRepository struct {
	db *sql.DB
}

// NewPostgresChannelRepository creates a new PostgresChannelRepository
func NewPostgresChannelRepository(db *sql.DB) *PostgresChannelRepository {
	return &PostgresChannelRepository{db: db}
}

// Create creates a new channel
func (r *PostgresChannelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	query := `
		INSERT INTO channels (id, tenant_id, section_id, network_id, name, description, type, position, bitrate, user_limit, slowmode, nsfw, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	if channel.ID == "" {
		channel.ID = domain.GenerateChannelID()
	}
	channel.CreatedAt = now
	channel.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		channel.ID,
		channel.TenantID,
		channel.SectionID,
		channel.NetworkID,
		channel.Name,
		channel.Description,
		channel.Type,
		channel.Position,
		channel.Bitrate,
		channel.UserLimit,
		channel.Slowmode,
		channel.NSFW,
		channel.CreatedAt,
		channel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	return nil
}

// GetByID retrieves a channel by ID
func (r *PostgresChannelRepository) GetByID(ctx context.Context, id string) (*domain.Channel, error) {
	query := `
		SELECT id, tenant_id, section_id, network_id, name, description, type, position, bitrate, user_limit, slowmode, nsfw, created_at, updated_at, deleted_at
		FROM channels
		WHERE id = $1 AND deleted_at IS NULL
	`

	var channel domain.Channel
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID,
		&channel.TenantID,
		&channel.SectionID,
		&channel.NetworkID,
		&channel.Name,
		&description,
		&channel.Type,
		&channel.Position,
		&channel.Bitrate,
		&channel.UserLimit,
		&channel.Slowmode,
		&channel.NSFW,
		&channel.CreatedAt,
		&channel.UpdatedAt,
		&channel.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	channel.Description = description.String

	return &channel, nil
}

// List retrieves channels with filters
func (r *PostgresChannelRepository) List(ctx context.Context, filter ChannelFilter) ([]domain.Channel, string, error) {
	query := `
		SELECT id, tenant_id, section_id, network_id, name, description, type, position, bitrate, user_limit, slowmode, nsfw, created_at, updated_at, deleted_at
		FROM channels
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argCount := 0

	if filter.TenantID != nil {
		argCount++
		query += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		args = append(args, *filter.TenantID)
	}

	if filter.SectionID != nil {
		argCount++
		query += fmt.Sprintf(" AND section_id = $%d", argCount)
		args = append(args, *filter.SectionID)
	}

	if filter.NetworkID != nil {
		argCount++
		query += fmt.Sprintf(" AND network_id = $%d", argCount)
		args = append(args, *filter.NetworkID)
	}

	if filter.Type != nil {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filter.Type)
	}

	if filter.Cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND id < $%d", argCount)
		args = append(args, filter.Cursor)
	}

	query += " ORDER BY position ASC, id DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	for rows.Next() {
		var channel domain.Channel
		var description sql.NullString

		if err := rows.Scan(
			&channel.ID,
			&channel.TenantID,
			&channel.SectionID,
			&channel.NetworkID,
			&channel.Name,
			&description,
			&channel.Type,
			&channel.Position,
			&channel.Bitrate,
			&channel.UserLimit,
			&channel.Slowmode,
			&channel.NSFW,
			&channel.CreatedAt,
			&channel.UpdatedAt,
			&channel.DeletedAt,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan channel: %w", err)
		}

		channel.Description = description.String
		channels = append(channels, channel)
	}

	var nextCursor string
	if len(channels) > limit {
		nextCursor = channels[limit-1].ID
		channels = channels[:limit]
	}

	return channels, nextCursor, nil
}

// Update updates a channel
func (r *PostgresChannelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	query := `
		UPDATE channels
		SET name = $2, description = $3, position = $4, bitrate = $5, user_limit = $6, slowmode = $7, nsfw = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	channel.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		channel.ID,
		channel.Name,
		channel.Description,
		channel.Position,
		channel.Bitrate,
		channel.UserLimit,
		channel.Slowmode,
		channel.NSFW,
		channel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}

	return nil
}

// Delete soft-deletes a channel
func (r *PostgresChannelRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE channels
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}

	return nil
}

// UpdatePositions updates positions for multiple channels
func (r *PostgresChannelRepository) UpdatePositions(ctx context.Context, parentID string, positions map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE channels
		SET position = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	for channelID, position := range positions {
		_, err := tx.ExecContext(ctx, query, channelID, position, now)
		if err != nil {
			return fmt.Errorf("failed to update position for channel %s: %w", channelID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Ensure interface compliance
var _ ChannelRepository = (*PostgresChannelRepository)(nil)
