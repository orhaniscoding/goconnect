package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresChatRepository implements ChatRepository with PostgreSQL
type PostgresChatRepository struct {
	db *sql.DB
}

// NewPostgresChatRepository creates a new PostgreSQL chat repository
func NewPostgresChatRepository(db *sql.DB) *PostgresChatRepository {
	return &PostgresChatRepository{
		db: db,
	}
}

// Create creates a new chat message
func (r *PostgresChatRepository) Create(ctx context.Context, msg *domain.ChatMessage) error {
	if msg.ID == "" {
		msg.ID = domain.GenerateNetworkID()
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = msg.CreatedAt
	}

	attachments, _ := json.Marshal(msg.Attachments)

	query := `
		INSERT INTO chat_messages (
			id, scope, tenant_id, user_id, body, attachments, 
			redacted, deleted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.Scope, msg.TenantID, msg.UserID, msg.Body,
		attachments, msg.Redacted, msg.DeletedAt,
		msg.CreatedAt, msg.UpdatedAt,
	)

	return err
}

// GetByID retrieves a message by ID
func (r *PostgresChatRepository) GetByID(ctx context.Context, id string) (*domain.ChatMessage, error) {
	query := `
		SELECT id, scope, tenant_id, user_id, body, attachments,
		       redacted, deleted_at, created_at, updated_at
		FROM chat_messages
		WHERE id = $1
	`

	msg := &domain.ChatMessage{}
	var attachmentsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.Scope, &msg.TenantID, &msg.UserID, &msg.Body,
		&attachmentsJSON, &msg.Redacted, &msg.DeletedAt,
		&msg.CreatedAt, &msg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}
	if err != nil {
		return nil, err
	}

	if len(attachmentsJSON) > 0 {
		json.Unmarshal(attachmentsJSON, &msg.Attachments)
	}

	return msg, nil
}

// List retrieves messages matching the filter
func (r *PostgresChatRepository) List(ctx context.Context, filter domain.ChatMessageFilter) ([]*domain.ChatMessage, string, error) {
	// Build query with filters
	query := `
		SELECT id, scope, tenant_id, user_id, body, attachments,
		       redacted, deleted_at, created_at, updated_at
		FROM chat_messages
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.TenantID != "" {
		query += ` AND tenant_id = $` + strconv.Itoa(argPos)
		args = append(args, filter.TenantID)
		argPos++
	}

	if filter.Scope != "" {
		query += ` AND scope = $` + strconv.Itoa(argPos)
		args = append(args, filter.Scope)
		argPos++
	}

	if filter.UserID != "" {
		query += ` AND user_id = $` + strconv.Itoa(argPos)
		args = append(args, filter.UserID)
		argPos++
	}

	if !filter.Since.IsZero() {
		query += ` AND created_at >= $` + strconv.Itoa(argPos)
		args = append(args, filter.Since)
		argPos++
	}

	if !filter.Before.IsZero() {
		query += ` AND created_at <= $` + strconv.Itoa(argPos)
		args = append(args, filter.Before)
		argPos++
	}

	if !filter.IncludeDeleted {
		query += ` AND deleted_at IS NULL`
	}

	// Cursor pagination
	if filter.Cursor != "" {
		query += ` AND created_at < (SELECT created_at FROM chat_messages WHERE id = $` + strconv.Itoa(argPos) + `)`
		args = append(args, filter.Cursor)
		argPos++
	}

	// Order and limit
	query += ` ORDER BY created_at DESC`

	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Fetch one extra to check if there are more results
	query += ` LIMIT $` + strconv.Itoa(argPos)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	messages := make([]*domain.ChatMessage, 0, limit)
	for rows.Next() {
		msg := &domain.ChatMessage{}
		var attachmentsJSON []byte

		err := rows.Scan(
			&msg.ID, &msg.Scope, &msg.TenantID, &msg.UserID, &msg.Body,
			&attachmentsJSON, &msg.Redacted, &msg.DeletedAt,
			&msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}

		if len(attachmentsJSON) > 0 {
			json.Unmarshal(attachmentsJSON, &msg.Attachments)
		}

		messages = append(messages, msg)
	}

	// Check for next cursor
	var nextCursor string
	if len(messages) > limit {
		nextCursor = messages[limit-1].ID
		messages = messages[:limit]
	}

	return messages, nextCursor, nil
}

// Update updates an existing message
func (r *PostgresChatRepository) Update(ctx context.Context, msg *domain.ChatMessage) error {
	msg.UpdatedAt = time.Now()

	attachments, _ := json.Marshal(msg.Attachments)

	query := `
		UPDATE chat_messages
		SET body = $1, attachments = $2, redacted = $3,
		    deleted_at = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		msg.Body, attachments, msg.Redacted,
		msg.DeletedAt, msg.UpdatedAt, msg.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": msg.ID,
		})
	}

	return nil
}

// Delete hard deletes a message
func (r *PostgresChatRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM chat_messages WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}

	return nil
}

// SoftDelete marks a message as deleted
func (r *PostgresChatRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()

	query := `UPDATE chat_messages SET deleted_at = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}

	return nil
}

// AddEdit adds an edit history entry
func (r *PostgresChatRepository) AddEdit(ctx context.Context, edit *domain.ChatMessageEdit) error {
	if edit.ID == "" {
		edit.ID = domain.GenerateNetworkID()
	}

	if edit.EditedAt.IsZero() {
		edit.EditedAt = time.Now()
	}

	query := `
		INSERT INTO chat_message_edits (
			id, message_id, prev_body, new_body, editor_id, edited_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		edit.ID, edit.MessageID, edit.PrevBody, edit.NewBody,
		edit.EditorID, edit.EditedAt,
	)

	return err
}

// GetEdits retrieves edit history for a message
func (r *PostgresChatRepository) GetEdits(ctx context.Context, messageID string) ([]*domain.ChatMessageEdit, error) {
	query := `
		SELECT id, message_id, prev_body, new_body, editor_id, edited_at
		FROM chat_message_edits
		WHERE message_id = $1
		ORDER BY edited_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edits := make([]*domain.ChatMessageEdit, 0)
	for rows.Next() {
		edit := &domain.ChatMessageEdit{}
		err := rows.Scan(
			&edit.ID, &edit.MessageID, &edit.PrevBody, &edit.NewBody,
			&edit.EditorID, &edit.EditedAt,
		)
		if err != nil {
			return nil, err
		}
		edits = append(edits, edit)
	}

	return edits, nil
}

// CountToday returns the number of messages created today
func (r *PostgresChatRepository) CountToday(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM chat_messages 
		WHERE created_at >= CURRENT_DATE 
		AND created_at < CURRENT_DATE + INTERVAL '1 day'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
