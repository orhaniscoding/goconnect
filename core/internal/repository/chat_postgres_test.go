package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresChatRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresChatRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with all fields", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:          "msg-123",
			Scope:       "host",
			TenantID:    "tenant-1",
			UserID:      "user-1",
			Body:        "Hello, World!",
			Attachments: []string{"file1.txt", "file2.png"},
			Redacted:    false,
			DeletedAt:   nil,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		attachments, _ := json.Marshal(msg.Attachments)

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_messages`)).
			WithArgs(msg.ID, msg.Scope, msg.TenantID, msg.UserID, msg.Body,
				attachments, msg.Redacted, msg.DeletedAt, msg.CreatedAt, msg.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, msg)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with auto-generated ID and timestamps", func(t *testing.T) {
		msg := &domain.ChatMessage{
			Scope:    "network:net-1",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Body:     "Test message",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_messages`)).
			WithArgs(sqlmock.AnyArg(), msg.Scope, msg.TenantID, msg.UserID, msg.Body,
				sqlmock.AnyArg(), false, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, msg)
		require.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.False(t, msg.CreatedAt.IsZero())
		assert.False(t, msg.UpdatedAt.IsZero())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:       "msg-456",
			Scope:    "host",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Body:     "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_messages`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		attachments := []string{"file1.txt"}
		attachmentsJSON, _ := json.Marshal(attachments)

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-123", "host", "tenant-1", "user-1", "Hello", attachmentsJSON,
			false, nil, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, scope, tenant_id, user_id, body, attachments`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		msg, err := repo.GetByID(ctx, "msg-123")
		require.NoError(t, err)
		require.NotNil(t, msg)
		assert.Equal(t, "msg-123", msg.ID)
		assert.Equal(t, "host", msg.Scope)
		assert.Equal(t, "tenant-1", msg.TenantID)
		assert.Equal(t, "user-1", msg.UserID)
		assert.Equal(t, "Hello", msg.Body)
		assert.Equal(t, attachments, msg.Attachments)
		assert.False(t, msg.Redacted)
		assert.Nil(t, msg.DeletedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, scope, tenant_id, user_id, body, attachments`)).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		msg, err := repo.GetByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, msg)

		// Check it's a domain error
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, scope, tenant_id, user_id, body, attachments`)).
			WithArgs("msg-123").
			WillReturnError(errors.New("database error"))

		msg, err := repo.GetByID(ctx, "msg-123")
		require.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty attachments", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-124", "host", "tenant-1", "user-1", "No attachments", []byte{},
			false, nil, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, scope, tenant_id, user_id, body, attachments`)).
			WithArgs("msg-124").
			WillReturnRows(rows)

		msg, err := repo.GetByID(ctx, "msg-124")
		require.NoError(t, err)
		require.NotNil(t, msg)
		assert.Empty(t, msg.Attachments)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with no filters", func(t *testing.T) {
		filter := domain.ChatMessageFilter{}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).
			AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now).
			AddRow("msg-2", "host", "tenant-1", "user-2", "Hi", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(51). // Default limit + 1
			WillReturnRows(rows)

		messages, nextCursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with tenant filter", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			TenantID: "tenant-1",
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs("tenant-1", 51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "tenant-1", messages[0].TenantID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with scope filter", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Scope: "network:net-1",
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "network:net-1", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs("network:net-1", 51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "network:net-1", messages[0].Scope)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with user filter", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			UserID: "user-1",
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs("user-1", 51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "user-1", messages[0].UserID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with time range filters", func(t *testing.T) {
		since := now.Add(-24 * time.Hour)
		before := now

		filter := domain.ChatMessageFilter{
			Since:  since,
			Before: before,
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(since, before, 51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with cursor pagination", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Cursor: "msg-prev",
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs("msg-prev", 51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with custom limit", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Limit: 10,
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		})
		for i := 0; i < 11; i++ {
			rows.AddRow("msg-"+string(rune('0'+i)), "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)
		}

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(11). // limit + 1
			WillReturnRows(rows)

		messages, nextCursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 10)
		assert.NotEmpty(t, nextCursor) // Has next page
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Limit: 200, // Should be capped to 100
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(101). // 100 + 1 (capped)
			WillReturnRows(rows)

		_, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		filter := domain.ChatMessageFilter{}

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(51).
			WillReturnError(errors.New("database error"))

		messages, _, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, messages)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row scan error", func(t *testing.T) {
		filter := domain.ChatMessageFilter{}

		// Return rows with wrong number of columns
		rows := sqlmock.NewRows([]string{"id"}).AddRow("msg-1")

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs(51).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, messages)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:          "msg-123",
			Body:        "Updated message",
			Attachments: []string{"new-file.txt"},
			Redacted:    false,
			DeletedAt:   nil,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages`)).
			WithArgs(msg.Body, sqlmock.AnyArg(), msg.Redacted, msg.DeletedAt, sqlmock.AnyArg(), msg.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, msg)
		require.NoError(t, err)
		assert.False(t, msg.UpdatedAt.IsZero())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:   "nonexistent",
			Body: "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages`)).
			WithArgs(msg.Body, sqlmock.AnyArg(), msg.Redacted, msg.DeletedAt, sqlmock.AnyArg(), msg.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, msg)
		require.Error(t, err)

		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		msg := &domain.ChatMessage{
			ID:   "msg-123",
			Body: "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages`)).
			WithArgs(msg.Body, sqlmock.AnyArg(), msg.Redacted, msg.DeletedAt, sqlmock.AnyArg(), msg.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat_messages WHERE id = $1`)).
			WithArgs("msg-123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "msg-123")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat_messages WHERE id = $1`)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "nonexistent")
		require.Error(t, err)

		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat_messages WHERE id = $1`)).
			WithArgs("msg-123").
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, "msg-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages SET deleted_at = $1, updated_at = $2 WHERE id = $3`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "msg-123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SoftDelete(ctx, "msg-123")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages SET deleted_at = $1, updated_at = $2 WHERE id = $3`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SoftDelete(ctx, "nonexistent")
		require.Error(t, err)

		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat_messages SET deleted_at = $1, updated_at = $2 WHERE id = $3`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "msg-123").
			WillReturnError(errors.New("database error"))

		err := repo.SoftDelete(ctx, "msg-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_AddEdit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with all fields", func(t *testing.T) {
		edit := &domain.ChatMessageEdit{
			ID:        "edit-123",
			MessageID: "msg-123",
			PrevBody:  "Original message",
			NewBody:   "Updated message",
			EditorID:  "user-1",
			EditedAt:  now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_message_edits`)).
			WithArgs(edit.ID, edit.MessageID, edit.PrevBody, edit.NewBody, edit.EditorID, edit.EditedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddEdit(ctx, edit)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with auto-generated ID and timestamp", func(t *testing.T) {
		edit := &domain.ChatMessageEdit{
			MessageID: "msg-123",
			PrevBody:  "Original",
			NewBody:   "Updated",
			EditorID:  "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_message_edits`)).
			WithArgs(sqlmock.AnyArg(), edit.MessageID, edit.PrevBody, edit.NewBody, edit.EditorID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddEdit(ctx, edit)
		require.NoError(t, err)
		assert.NotEmpty(t, edit.ID)
		assert.False(t, edit.EditedAt.IsZero())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		edit := &domain.ChatMessageEdit{
			MessageID: "msg-123",
			PrevBody:  "Original",
			NewBody:   "Updated",
			EditorID:  "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_message_edits`)).
			WithArgs(sqlmock.AnyArg(), edit.MessageID, edit.PrevBody, edit.NewBody, edit.EditorID, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.AddEdit(ctx, edit)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_GetEdits(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with multiple edits", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "message_id", "prev_body", "new_body", "editor_id", "edited_at",
		}).
			AddRow("edit-1", "msg-123", "Original", "First edit", "user-1", now.Add(-1*time.Hour)).
			AddRow("edit-2", "msg-123", "First edit", "Second edit", "user-1", now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, message_id, prev_body, new_body, editor_id, edited_at`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		edits, err := repo.GetEdits(ctx, "msg-123")
		require.NoError(t, err)
		assert.Len(t, edits, 2)
		assert.Equal(t, "edit-1", edits[0].ID)
		assert.Equal(t, "edit-2", edits[1].ID)
		assert.Equal(t, "msg-123", edits[0].MessageID)
		assert.Equal(t, "Original", edits[0].PrevBody)
		assert.Equal(t, "First edit", edits[0].NewBody)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no edits", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "message_id", "prev_body", "new_body", "editor_id", "edited_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, message_id, prev_body, new_body, editor_id, edited_at`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		edits, err := repo.GetEdits(ctx, "msg-123")
		require.NoError(t, err)
		assert.Empty(t, edits)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, message_id, prev_body, new_body, editor_id, edited_at`)).
			WithArgs("msg-123").
			WillReturnError(errors.New("database error"))

		edits, err := repo.GetEdits(ctx, "msg-123")
		require.Error(t, err)
		assert.Nil(t, edits)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row scan error", func(t *testing.T) {
		// Return rows with wrong number of columns
		rows := sqlmock.NewRows([]string{"id"}).AddRow("edit-1")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, message_id, prev_body, new_body, editor_id, edited_at`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		edits, err := repo.GetEdits(ctx, "msg-123")
		require.Error(t, err)
		assert.Nil(t, edits)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_CountToday(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(42)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*)`)).
			WillReturnRows(rows)

		count, err := repo.CountToday(ctx)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with zero count", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*)`)).
			WillReturnRows(rows)

		count, err := repo.CountToday(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*)`)).
			WillReturnError(errors.New("database error"))

		count, err := repo.CountToday(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresChatRepository_List_AllFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with all filters combined", func(t *testing.T) {
		since := now.Add(-24 * time.Hour)
		before := now

		filter := domain.ChatMessageFilter{
			TenantID:       "tenant-1",
			Scope:          "host",
			UserID:         "user-1",
			Since:          since,
			Before:         before,
			IncludeDeleted: true,
			Cursor:         "msg-prev",
			Limit:          25,
		}

		rows := sqlmock.NewRows([]string{
			"id", "scope", "tenant_id", "user_id", "body", "attachments",
			"redacted", "deleted_at", "created_at", "updated_at",
		}).AddRow("msg-1", "host", "tenant-1", "user-1", "Hello", []byte("[]"), false, nil, now, now)

		mock.ExpectQuery(`SELECT id, scope, tenant_id, user_id, body, attachments`).
			WithArgs("tenant-1", "host", "user-1", since, before, "msg-prev", 26).
			WillReturnRows(rows)

		messages, _, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
