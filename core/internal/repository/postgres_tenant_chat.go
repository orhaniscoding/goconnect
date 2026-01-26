package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresTenantChatRepository implements TenantChatRepository using PostgreSQL
type PostgresTenantChatRepository struct {
	db *sql.DB
}

// NewPostgresTenantChatRepository creates a new PostgreSQL-backed repository
func NewPostgresTenantChatRepository(db *sql.DB) *PostgresTenantChatRepository {
	return &PostgresTenantChatRepository{db: db}
}

func (r *PostgresTenantChatRepository) Create(ctx context.Context, message *domain.TenantChatMessage) error {
	if message.ID == "" {
		message.ID = domain.GenerateChatMessageID()
	}

	query := `
		INSERT INTO tenant_chat_messages (id, tenant_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	message.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.TenantID,
		message.UserID,
		message.Content,
		message.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}
	return nil
}

func (r *PostgresTenantChatRepository) GetByID(ctx context.Context, id string) (*domain.TenantChatMessage, error) {
	query := `
		SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at, m.deleted_at,
		       u.email, u.locale
		FROM tenant_chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`
	msg := &domain.TenantChatMessage{
		User: &domain.User{},
	}
	var editedAt, deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.TenantID,
		&msg.UserID,
		&msg.Content,
		&msg.CreatedAt,
		&editedAt,
		&deletedAt,
		&msg.User.Email,
		&msg.User.Locale,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat message: %w", err)
	}
	if editedAt.Valid {
		msg.EditedAt = &editedAt.Time
	}
	if deletedAt.Valid {
		msg.DeletedAt = &deletedAt.Time
	}
	msg.User.ID = msg.UserID
	return msg, nil
}

func (r *PostgresTenantChatRepository) Update(ctx context.Context, message *domain.TenantChatMessage) error {
	query := `
		UPDATE tenant_chat_messages
		SET content = $2, edited_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	now := time.Now()
	message.EditedAt = &now
	result, err := r.db.ExecContext(ctx, query, message.ID, message.Content, message.EditedAt)
	if err != nil {
		return fmt.Errorf("failed to update chat message: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Message not found or deleted", nil)
	}
	return nil
}

func (r *PostgresTenantChatRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE tenant_chat_messages SET deleted_at = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete chat message: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	return nil
}

func (r *PostgresTenantChatRepository) ListByTenant(ctx context.Context, tenantID string, beforeID string, limit int) ([]*domain.TenantChatMessage, error) {
	query := `
		SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at,
		       u.email, u.locale
		FROM tenant_chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1 AND m.deleted_at IS NULL
	`
	args := []interface{}{tenantID}

	if beforeID != "" {
		// Get messages before the specified message
		query += ` AND m.created_at < (
			SELECT created_at FROM tenant_chat_messages WHERE id = $2
		)`
		args = append(args, beforeID)
	}

	query += " ORDER BY m.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.TenantChatMessage
	for rows.Next() {
		msg := &domain.TenantChatMessage{
			User: &domain.User{},
		}
		var editedAt sql.NullTime
		err := rows.Scan(
			&msg.ID,
			&msg.TenantID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
			&editedAt,
			&msg.User.Email,
			&msg.User.Locale,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		if editedAt.Valid {
			msg.EditedAt = &editedAt.Time
		}
		msg.User.ID = msg.UserID
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *PostgresTenantChatRepository) DeleteOlderThan(ctx context.Context, tenantID string, before time.Time) (int, error) {
	query := `DELETE FROM tenant_chat_messages WHERE tenant_id = $1 AND created_at < $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old messages: %w", err)
	}
	rows, err := result.RowsAffected()
	return int(rows), nil
}

func (r *PostgresTenantChatRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenant_chat_messages WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant chat messages: %w", err)
	}
	return nil
}
