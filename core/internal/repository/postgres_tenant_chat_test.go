package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresTenantChatRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresTenantChatRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()

	t.Run("success with provided ID", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			ID:       "msg-123",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Content:  "Hello, World!",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_chat_messages (id, tenant_id, user_id, content, created_at)`)).
			WithArgs(
				message.ID,
				message.TenantID,
				message.UserID,
				message.Content,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, message)
		require.NoError(t, err)
		assert.NotZero(t, message.CreatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without provided ID", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			TenantID: "tenant-1",
			UserID:   "user-1",
			Content:  "Hello, World!",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_chat_messages (id, tenant_id, user_id, content, created_at)`)).
			WithArgs(
				sqlmock.AnyArg(),
				message.TenantID,
				message.UserID,
				message.Content,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, message)
		require.NoError(t, err)
		assert.NotEmpty(t, message.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			ID:       "msg-123",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Content:  "Hello, World!",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_chat_messages`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, message)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chat message")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without edited_at", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at", "deleted_at",
			"email", "locale",
		}).AddRow(
			"msg-123", "tenant-1", "user-1", "Hello, World!", now, nil, nil,
			"user@example.com", "en",
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at, m.deleted_at`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		message, err := repo.GetByID(ctx, "msg-123")
		require.NoError(t, err)
		assert.Equal(t, "msg-123", message.ID)
		assert.Equal(t, "tenant-1", message.TenantID)
		assert.Equal(t, "user-1", message.UserID)
		assert.Equal(t, "Hello, World!", message.Content)
		assert.Nil(t, message.EditedAt)
		assert.Nil(t, message.DeletedAt)
		assert.NotNil(t, message.User)
		assert.Equal(t, "user@example.com", message.User.Email)
		assert.Equal(t, "en", message.User.Locale)
		assert.Equal(t, "user-1", message.User.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with edited_at", func(t *testing.T) {
		editedAt := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at", "deleted_at",
			"email", "locale",
		}).AddRow(
			"msg-123", "tenant-1", "user-1", "Edited message", now, editedAt, nil,
			"user@example.com", "en",
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at, m.deleted_at`)).
			WithArgs("msg-123").
			WillReturnRows(rows)

		message, err := repo.GetByID(ctx, "msg-123")
		require.NoError(t, err)
		require.NotNil(t, message.EditedAt)
		assert.Equal(t, editedAt.Unix(), message.EditedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at, m.deleted_at`)).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		message, err := repo.GetByID(ctx, "non-existent")
		require.Error(t, err)
		assert.Nil(t, message)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at, m.deleted_at`)).
			WithArgs("msg-123").
			WillReturnError(errors.New("database error"))

		message, err := repo.GetByID(ctx, "msg-123")
		require.Error(t, err)
		assert.Nil(t, message)
		assert.Contains(t, err.Error(), "failed to get chat message")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			ID:      "msg-123",
			Content: "Updated content",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages`)).
			WithArgs(
				message.ID,
				message.Content,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, message)
		require.NoError(t, err)
		assert.NotNil(t, message.EditedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or deleted", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			ID:      "non-existent",
			Content: "Updated content",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages`)).
			WithArgs(
				message.ID,
				message.Content,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, message)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		message := &domain.TenantChatMessage{
			ID:      "msg-123",
			Content: "Updated content",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, message)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update chat message")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages SET deleted_at = $2 WHERE id = $1`)).
			WithArgs("msg-123", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SoftDelete(ctx, "msg-123")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages SET deleted_at = $2 WHERE id = $1`)).
			WithArgs("non-existent", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SoftDelete(ctx, "non-existent")
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_chat_messages SET deleted_at = $2 WHERE id = $1`)).
			WithArgs("msg-123", sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.SoftDelete(ctx, "msg-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete chat message")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with no filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at",
			"email", "locale",
		}).
			AddRow("msg-1", "tenant-1", "user-1", "Hello", now, nil, "user1@example.com", "en").
			AddRow("msg-2", "tenant-1", "user-2", "World", now, nil, "user2@example.com", "fr")

		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		messages, err := repo.ListByTenant(ctx, "tenant-1", "", 0)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "msg-1", messages[0].ID)
		assert.Equal(t, "msg-2", messages[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with beforeID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at",
			"email", "locale",
		}).
			AddRow("msg-1", "tenant-1", "user-1", "Earlier message", now.Add(-time.Hour), nil, "user1@example.com", "en")

		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1", "msg-2").
			WillReturnRows(rows)

		messages, err := repo.ListByTenant(ctx, "tenant-1", "msg-2", 0)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, "msg-1", messages[0].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with limit", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at",
			"email", "locale",
		}).
			AddRow("msg-1", "tenant-1", "user-1", "Hello", now, nil, "user1@example.com", "en").
			AddRow("msg-2", "tenant-1", "user-2", "World", now, nil, "user2@example.com", "fr")

		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		messages, err := repo.ListByTenant(ctx, "tenant-1", "", 10)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with edited_at", func(t *testing.T) {
		editedAt := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "content", "created_at", "edited_at",
			"email", "locale",
		}).
			AddRow("msg-1", "tenant-1", "user-1", "Edited", now, editedAt, "user1@example.com", "en")

		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		messages, err := repo.ListByTenant(ctx, "tenant-1", "", 0)
		require.NoError(t, err)
		require.Len(t, messages, 1)
		require.NotNil(t, messages[0].EditedAt)
		assert.Equal(t, editedAt.Unix(), messages[0].EditedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		messages, err := repo.ListByTenant(ctx, "tenant-1", "", 0)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Contains(t, err.Error(), "failed to list chat messages")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id",
		}).
			AddRow("msg-1", "tenant-1")

		mock.ExpectQuery(`SELECT m.id, m.tenant_id, m.user_id, m.content, m.created_at, m.edited_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		messages, err := repo.ListByTenant(ctx, "tenant-1", "", 0)
		require.Error(t, err)
		assert.Nil(t, messages)
		assert.Contains(t, err.Error(), "failed to scan chat message")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_DeleteOlderThan(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()
	before := time.Now().Add(-24 * time.Hour)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1 AND created_at < $2`)).
			WithArgs("tenant-1", before).
			WillReturnResult(sqlmock.NewResult(0, 10))

		count, err := repo.DeleteOlderThan(ctx, "tenant-1", before)
		require.NoError(t, err)
		assert.Equal(t, 10, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success no rows deleted", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1 AND created_at < $2`)).
			WithArgs("tenant-1", before).
			WillReturnResult(sqlmock.NewResult(0, 0))

		count, err := repo.DeleteOlderThan(ctx, "tenant-1", before)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1 AND created_at < $2`)).
			WithArgs("tenant-1", before).
			WillReturnError(errors.New("database error"))

		count, err := repo.DeleteOlderThan(ctx, "tenant-1", before)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to delete old messages")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantChatRepository_DeleteAllByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantChatRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnResult(sqlmock.NewResult(0, 25))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success no rows affected", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1`)).
			WithArgs("tenant-empty").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteAllByTenant(ctx, "tenant-empty")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_chat_messages WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant chat messages")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
