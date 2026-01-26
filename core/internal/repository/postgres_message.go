package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES MESSAGE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresMessageRepository implements MessageRepository for PostgreSQL
type PostgresMessageRepository struct {
	db *sql.DB
}

// NewPostgresMessageRepository creates a new PostgresMessageRepository
func NewPostgresMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

// Create creates a new message
func (r *PostgresMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (id, channel_id, author_id, content, reply_to_id, thread_id, attachments, embeds, mentions, mention_roles, mention_everyone, pinned, encrypted, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	now := time.Now()
	if message.ID == "" {
		message.ID = domain.GenerateMessageID()
	}
	message.CreatedAt = now

	attachmentsJSON, err := json.Marshal(message.Attachments)
	if err != nil {
		return fmt.Errorf("failed to marshal attachments: %w", err)
	}
	embedsJSON, err := json.Marshal(message.Embeds)
	if err != nil {
		return fmt.Errorf("failed to marshal embeds: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		message.ID,
		message.ChannelID,
		message.AuthorID,
		message.Content,
		message.ReplyToID,
		message.ThreadID,
		attachmentsJSON,
		embedsJSON,
		pq.Array(message.Mentions),
		pq.Array(message.MentionRoles),
		message.MentionEveryone,
		message.Pinned,
		message.Encrypted,
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by ID
func (r *PostgresMessageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, channel_id, author_id, content, reply_to_id, thread_id, attachments, embeds, mentions, mention_roles, mention_everyone, pinned, edited_at, encrypted, created_at, deleted_at
		FROM messages
		WHERE id = $1 AND deleted_at IS NULL
	`

	var message domain.Message
	var attachmentsJSON, embedsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ChannelID,
		&message.AuthorID,
		&message.Content,
		&message.ReplyToID,
		&message.ThreadID,
		&attachmentsJSON,
		&embedsJSON,
		pq.Array(&message.Mentions),
		pq.Array(&message.MentionRoles),
		&message.MentionEveryone,
		&message.Pinned,
		&message.EditedAt,
		&message.Encrypted,
		&message.CreatedAt,
		&message.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	json.Unmarshal(attachmentsJSON, &message.Attachments)
	json.Unmarshal(embedsJSON, &message.Embeds)

	return &message, nil
}

// List retrieves messages with filters
func (r *PostgresMessageRepository) List(ctx context.Context, filter MessageFilter) ([]domain.Message, error) {
	query := `
		SELECT id, channel_id, author_id, content, reply_to_id, thread_id, attachments, embeds, mentions, mention_roles, mention_everyone, pinned, edited_at, encrypted, created_at, deleted_at
		FROM messages
		WHERE channel_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{filter.ChannelID}
	argCount := 1

	if filter.Before != "" {
		argCount++
		query += fmt.Sprintf(" AND id < $%d", argCount)
		args = append(args, filter.Before)
	}

	if filter.After != "" {
		argCount++
		query += fmt.Sprintf(" AND id > $%d", argCount)
		args = append(args, filter.After)
	}

	if filter.AuthorID != nil {
		argCount++
		query += fmt.Sprintf(" AND author_id = $%d", argCount)
		args = append(args, *filter.AuthorID)
	}

	if filter.Pinned != nil {
		argCount++
		query += fmt.Sprintf(" AND pinned = $%d", argCount)
		args = append(args, *filter.Pinned)
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var message domain.Message
		var attachmentsJSON, embedsJSON []byte

		if err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.AuthorID,
			&message.Content,
			&message.ReplyToID,
			&message.ThreadID,
			&attachmentsJSON,
			&embedsJSON,
			pq.Array(&message.Mentions),
			pq.Array(&message.MentionRoles),
			&message.MentionEveryone,
			&message.Pinned,
			&message.EditedAt,
			&message.Encrypted,
			&message.CreatedAt,
			&message.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		json.Unmarshal(attachmentsJSON, &message.Attachments)
		json.Unmarshal(embedsJSON, &message.Embeds)
		messages = append(messages, message)
	}

	return messages, nil
}

// Update updates a message (for edits)
func (r *PostgresMessageRepository) Update(ctx context.Context, message *domain.Message) error {
	query := `
		UPDATE messages
		SET content = $2, edited_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	message.EditedAt = &now

	result, err := r.db.ExecContext(ctx, query, message.ID, message.Content, message.EditedAt)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// Delete soft-deletes a message
func (r *PostgresMessageRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE messages
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// SetPinned pins/unpins a message
func (r *PostgresMessageRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	query := `
		UPDATE messages
		SET pinned = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, pinned)
	if err != nil {
		return fmt.Errorf("failed to set pinned: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// GetPinned retrieves all pinned messages in a channel
func (r *PostgresMessageRepository) GetPinned(ctx context.Context, channelID string) ([]domain.Message, error) {
	filter := MessageFilter{
		ChannelID: channelID,
		Pinned:    boolPtr(true),
		Limit:     50,
	}
	return r.List(ctx, filter)
}

// Search searches messages (full-text)
func (r *PostgresMessageRepository) Search(ctx context.Context, channelID, query string, limit int) ([]domain.Message, error) {
	sqlQuery := `
		SELECT id, channel_id, author_id, content, reply_to_id, thread_id, attachments, embeds, mentions, mention_roles, mention_everyone, pinned, edited_at, encrypted, created_at, deleted_at
		FROM messages
		WHERE channel_id = $1 AND deleted_at IS NULL AND encrypted = FALSE
		AND to_tsvector('english', content) @@ plainto_tsquery('english', $2)
		ORDER BY created_at DESC
		LIMIT $3
	`

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, channelID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var message domain.Message
		var attachmentsJSON, embedsJSON []byte

		if err := rows.Scan(
			&message.ID,
			&message.ChannelID,
			&message.AuthorID,
			&message.Content,
			&message.ReplyToID,
			&message.ThreadID,
			&attachmentsJSON,
			&embedsJSON,
			pq.Array(&message.Mentions),
			pq.Array(&message.MentionRoles),
			&message.MentionEveryone,
			&message.Pinned,
			&message.EditedAt,
			&message.Encrypted,
			&message.CreatedAt,
			&message.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		json.Unmarshal(attachmentsJSON, &message.Attachments)
		json.Unmarshal(embedsJSON, &message.Embeds)
		messages = append(messages, message)
	}

	return messages, nil
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

// Ensure interface compliance
var _ MessageRepository = (*PostgresMessageRepository)(nil)
