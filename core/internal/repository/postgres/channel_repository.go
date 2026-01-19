package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// CHANNEL REPOSITORY POSTGRES IMPLEMENTATION
// ══════════════════════════════════════════════════════════════════════════════

type ChannelRepositoryPostgres struct {
	db *sql.DB
}

// NewChannelRepository creates a new PostgreSQL channel repository
func NewChannelRepository(db *sql.DB) repository.ChannelRepository {
	return &ChannelRepositoryPostgres{db: db}
}

// ══════════════════════════════════════════════════════════════════════════════
// CREATE
// ══════════════════════════════════════════════════════════════════════════════

func (r *ChannelRepositoryPostgres) Create(ctx context.Context, channel *domain.Channel) error {
	query := `
		INSERT INTO channels (
			id, tenant_id, section_id, network_id, name, description, type,
			position, bitrate, user_limit, slowmode, nsfw,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	now := time.Now()
	channel.CreatedAt = now
	channel.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
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

// ══════════════════════════════════════════════════════════════════════════════
// READ
// ══════════════════════════════════════════════════════════════════════════════

func (r *ChannelRepositoryPostgres) GetByID(ctx context.Context, id string) (*domain.Channel, error) {
	query := `
		SELECT
			id, tenant_id, section_id, network_id, name, description, type,
			position, bitrate, user_limit, slowmode, nsfw,
			created_at, updated_at, deleted_at
		FROM channels
		WHERE id = $1 AND deleted_at IS NULL
	`

	var channel domain.Channel
	row := r.db.QueryRowContext(ctx, query, id)
	err := database.ScanRow(row, &channel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "channel not found", map[string]any{
				"channel_id": id,
			})
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &channel, nil
}

func (r *ChannelRepositoryPostgres) List(ctx context.Context, filter repository.ChannelFilter) ([]domain.Channel, string, error) {
	// Build dynamic query based on filters
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition: only non-deleted channels
	conditions = append(conditions, "deleted_at IS NULL")

	// Filter by parent (TenantID, SectionID, or NetworkID)
	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, *filter.TenantID)
		argIndex++
	}

	if filter.SectionID != nil {
		conditions = append(conditions, fmt.Sprintf("section_id = $%d", argIndex))
		args = append(args, *filter.SectionID)
		argIndex++
	}

	if filter.NetworkID != nil {
		conditions = append(conditions, fmt.Sprintf("network_id = $%d", argIndex))
		args = append(args, *filter.NetworkID)
		argIndex++
	}

	// Filter by type
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	// Cursor-based pagination
	if filter.Cursor != "" {
		conditions = append(conditions, fmt.Sprintf("id > $%d", argIndex))
		args = append(args, filter.Cursor)
		argIndex++
	}

	// Build final query
	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
		SELECT
			id, tenant_id, section_id, network_id, name, description, type,
			position, bitrate, user_limit, slowmode, nsfw,
			created_at, updated_at, deleted_at
		FROM channels
		WHERE %s
		ORDER BY position ASC, created_at DESC
		LIMIT $%d
	`, whereClause, argIndex)

	// Default limit
	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	args = append(args, limit+1) // Fetch one extra to determine if there's a next page

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	if err := database.ScanRows(rows, &channels); err != nil {
		return nil, "", fmt.Errorf("failed to scan channels: %w", err)
	}

	// Return empty slice instead of nil
	if channels == nil {
		channels = []domain.Channel{}
	}

	// Determine next cursor
	var nextCursor string
	if len(channels) > limit {
		nextCursor = channels[limit-1].ID
		channels = channels[:limit]
	}

	return channels, nextCursor, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// UPDATE
// ══════════════════════════════════════════════════════════════════════════════

func (r *ChannelRepositoryPostgres) Update(ctx context.Context, channel *domain.Channel) error {
	query := `
		UPDATE channels
		SET
			name = $2,
			description = $3,
			position = $4,
			bitrate = $5,
			user_limit = $6,
			slowmode = $7,
			nsfw = $8,
			updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	channel.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewError(domain.ErrNotFound, "channel not found or already deleted", map[string]any{
			"channel_id": channel.ID,
		})
	}

	return nil
}

func (r *ChannelRepositoryPostgres) UpdatePositions(ctx context.Context, parentID string, positions map[string]int) error {
	// Start transaction for atomic batch update
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Determine parent column based on ID prefix (or accept generic parent_id)
	// For simplicity, we'll update channels matching any of the parent IDs
	query := `
		UPDATE channels
		SET position = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
		AND (tenant_id = $4 OR section_id = $4 OR network_id = $4)
	`

	now := time.Now()
	for channelID, position := range positions {
		result, err := tx.ExecContext(ctx, query, channelID, position, now, parentID)
		if err != nil {
			return fmt.Errorf("failed to update channel position: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return domain.NewError(domain.ErrNotFound, "channel not found in parent", map[string]any{
				"channel_id": channelID,
				"parent_id":  parentID,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// DELETE (Soft Delete)
// ══════════════════════════════════════════════════════════════════════════════

func (r *ChannelRepositoryPostgres) Delete(ctx context.Context, id string) error {
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewError(domain.ErrNotFound, "channel not found or already deleted", map[string]any{
			"channel_id": id,
		})
	}

	return nil
}
