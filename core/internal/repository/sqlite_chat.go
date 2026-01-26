package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteChatRepository implements ChatRepository for network chat using SQLite.
type SQLiteChatRepository struct {
	db *sql.DB
}

func NewSQLiteChatRepository(db *sql.DB) *SQLiteChatRepository {
	return &SQLiteChatRepository{db: db}
}

func (r *SQLiteChatRepository) Create(ctx context.Context, msg *domain.ChatMessage) error {
	if msg.ID == "" {
		msg.ID = domain.GenerateChatMessageID()
	}
	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_messages (id, scope, tenant_id, user_id, body, attachments, redacted, deleted_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, '[]', 0, NULL, ?, ?)
	`, msg.ID, msg.Scope, msg.TenantID, msg.UserID, msg.Body, msg.CreatedAt, msg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}
	return nil
}

func (r *SQLiteChatRepository) GetByID(ctx context.Context, id string) (*domain.ChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, scope, tenant_id, user_id, body, redacted, deleted_at, created_at, updated_at
		FROM chat_messages
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	return scanChatMessage(row)
}

func (r *SQLiteChatRepository) List(ctx context.Context, filter domain.ChatMessageFilter) ([]*domain.ChatMessage, string, error) {
	args := []interface{}{}
	query := `
		SELECT id, scope, tenant_id, user_id, body, redacted, deleted_at, created_at, updated_at
		FROM chat_messages
		WHERE 1=1
	`
	if filter.Scope != "" {
		query += " AND scope = ?"
		args = append(args, filter.Scope)
	}
	if filter.TenantID != "" {
		query += " AND tenant_id = ?"
		args = append(args, filter.TenantID)
	}
	if filter.UserID != "" {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if !filter.Since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Since)
	}
	if !filter.Before.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.Before)
	}
	if !filter.IncludeDeleted {
		query += " AND deleted_at IS NULL"
	}
	if filter.Cursor != "" {
		query += " AND created_at < (SELECT created_at FROM chat_messages WHERE id = ?)"
		args = append(args, filter.Cursor)
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list chat messages: %w", err)
	}
	defer rows.Close()

	var msgs []*domain.ChatMessage
	for rows.Next() {
		msg, err := scanChatMessage(rows)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate chat messages: %w", err)
	}
	next := ""
	if len(msgs) > limit {
		next = msgs[limit].ID
		msgs = msgs[:limit]
	}
	return msgs, next, nil
}

func (r *SQLiteChatRepository) Update(ctx context.Context, msg *domain.ChatMessage) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE chat_messages
		SET body = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, msg.Body, time.Now(), msg.ID)
	if err != nil {
		return fmt.Errorf("failed to update chat message: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
	}
	return nil
}

func (r *SQLiteChatRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE chat_messages
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

func scanChatMessage(row interface{ Scan(dest ...any) error }) (*domain.ChatMessage, error) {
	var msg domain.ChatMessage
	var deletedAt sql.NullTime
	if err := row.Scan(
		&msg.ID,
		&msg.Scope,
		&msg.TenantID,
		&msg.UserID,
		&msg.Body,
		&msg.Redacted,
		&deletedAt,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
		}
		return nil, fmt.Errorf("failed to scan chat message: %w", err)
	}
	if deletedAt.Valid {
		msg.DeletedAt = &deletedAt.Time
	}
	return &msg, nil
}

func (r *SQLiteChatRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE chat_messages
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete chat message: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Chat message not found", nil)
	}
	return nil
}

func (r *SQLiteChatRepository) AddEdit(ctx context.Context, edit *domain.ChatMessageEdit) error {
	if edit.ID == "" {
		edit.ID = domain.GenerateNetworkID()
	}
	if edit.EditedAt.IsZero() {
		edit.EditedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_message_edits (id, message_id, prev_body, new_body, editor_id, edited_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, edit.ID, edit.MessageID, edit.PrevBody, edit.NewBody, edit.EditorID, edit.EditedAt)
	if err != nil {
		return fmt.Errorf("failed to add chat edit: %w", err)
	}
	return nil
}

func (r *SQLiteChatRepository) GetEdits(ctx context.Context, messageID string) ([]*domain.ChatMessageEdit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, message_id, prev_body, new_body, editor_id, edited_at
		FROM chat_message_edits
		WHERE message_id = ?
		ORDER BY edited_at ASC
	`, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat edits: %w", err)
	}
	defer rows.Close()

	var edits []*domain.ChatMessageEdit
	for rows.Next() {
		var e domain.ChatMessageEdit
		if err := rows.Scan(&e.ID, &e.MessageID, &e.PrevBody, &e.NewBody, &e.EditorID, &e.EditedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat edit: %w", err)
		}
		edits = append(edits, &e)
	}
	return edits, rows.Err()
}

func (r *SQLiteChatRepository) CountToday(ctx context.Context) (int, error) {
	var count int
	today := time.Now().Format("2006-01-02")
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM chat_messages
		WHERE deleted_at IS NULL AND date(created_at) = ?
	`, today).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count chat messages: %w", err)
	}
	return count, nil
}
