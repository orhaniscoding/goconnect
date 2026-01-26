package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteTenantChatRepository implements TenantChatRepository using SQLite.
type SQLiteTenantChatRepository struct {
	db *sql.DB
}

func NewSQLiteTenantChatRepository(db *sql.DB) *SQLiteTenantChatRepository {
	return &SQLiteTenantChatRepository{db: db}
}

func (r *SQLiteTenantChatRepository) Create(ctx context.Context, msg *domain.TenantChatMessage) error {
	if msg.ID == "" {
		msg.ID = domain.GenerateChatMessageID()
	}
	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_chat_messages (id, tenant_id, user_id, content, created_at, edited_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, msg.ID, msg.TenantID, msg.UserID, msg.Content, msg.CreatedAt, msg.EditedAt)
	if err != nil {
		return fmt.Errorf("failed to create tenant chat message: %w", err)
	}
	return nil
}

func (r *SQLiteTenantChatRepository) GetByID(ctx context.Context, id string) (*domain.TenantChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, content, created_at, edited_at, deleted_at
		FROM tenant_chat_messages
		WHERE id = ?
	`, id)
	var msg domain.TenantChatMessage
	var editedAt sql.NullTime
	var deletedAt sql.NullTime
	if err := row.Scan(&msg.ID, &msg.TenantID, &msg.UserID, &msg.Content, &msg.CreatedAt, &editedAt, &deletedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
		}
		return nil, fmt.Errorf("failed to get chat message: %w", err)
	}
	if editedAt.Valid {
		msg.EditedAt = &editedAt.Time
	}
	if deletedAt.Valid {
		msg.DeletedAt = &deletedAt.Time
	}
	return &msg, nil
}

func (r *SQLiteTenantChatRepository) Update(ctx context.Context, msg *domain.TenantChatMessage) error {
	now := time.Now()
	msg.EditedAt = &now
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_chat_messages
		SET content = ?, edited_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, msg.Content, msg.EditedAt, msg.ID)
	if err != nil {
		return fmt.Errorf("failed to update chat message: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
	}
	return nil
}

func (r *SQLiteTenantChatRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_chat_messages
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete chat message: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
	}
	return nil
}

func (r *SQLiteTenantChatRepository) ListByTenant(ctx context.Context, tenantID string, beforeID string, limit int) ([]*domain.TenantChatMessage, error) {
	args := []interface{}{tenantID}
	query := `
		SELECT id, tenant_id, user_id, content, created_at, edited_at, deleted_at
		FROM tenant_chat_messages
		WHERE tenant_id = ?
	`
	if beforeID != "" {
		query += " AND created_at < (SELECT created_at FROM tenant_chat_messages WHERE id = ?)"
		args = append(args, beforeID)
	}
	query += " AND deleted_at IS NULL"
	if limit <= 0 {
		limit = 50
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.TenantChatMessage
	for rows.Next() {
		var msg domain.TenantChatMessage
		var editedAt sql.NullTime
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&msg.ID, &msg.TenantID, &msg.UserID, &msg.Content, &msg.CreatedAt, &editedAt, &deletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		if editedAt.Valid {
			msg.EditedAt = &editedAt.Time
		}
		if deletedAt.Valid {
			msg.DeletedAt = &deletedAt.Time
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate chat messages: %w", err)
	}
	return messages, nil
}

func (r *SQLiteTenantChatRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_chat_messages WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant chat messages: %w", err)
	}
	return nil
}

func (r *SQLiteTenantChatRepository) DeleteOlderThan(ctx context.Context, tenantID string, before time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM tenant_chat_messages
		WHERE tenant_id = ? AND created_at < ?
	`, tenantID, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old chat messages: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return int(rows), nil
}

func (r *SQLiteTenantChatRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_chat_messages
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete tenant chat message: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
	}
	return nil
}
