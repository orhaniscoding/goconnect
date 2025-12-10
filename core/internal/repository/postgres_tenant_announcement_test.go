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

func TestNewPostgresTenantAnnouncementRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresTenantAnnouncementRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()

	t.Run("success with provided ID", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			ID:       "ann-123",
			TenantID: "tenant-1",
			Title:    "Test Announcement",
			Content:  "This is a test announcement",
			AuthorID: "user-1",
			IsPinned: false,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_announcements (id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at)`)).
			WithArgs(
				announcement.ID,
				announcement.TenantID,
				announcement.Title,
				announcement.Content,
				announcement.AuthorID,
				announcement.IsPinned,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, announcement)
		require.NoError(t, err)
		assert.NotZero(t, announcement.CreatedAt)
		assert.NotZero(t, announcement.UpdatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without provided ID", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			TenantID: "tenant-1",
			Title:    "Test Announcement",
			Content:  "This is a test announcement",
			AuthorID: "user-1",
			IsPinned: true,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_announcements (id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at)`)).
			WithArgs(
				sqlmock.AnyArg(),
				announcement.TenantID,
				announcement.Title,
				announcement.Content,
				announcement.AuthorID,
				announcement.IsPinned,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, announcement)
		require.NoError(t, err)
		assert.NotEmpty(t, announcement.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			ID:       "ann-123",
			TenantID: "tenant-1",
			Title:    "Test Announcement",
			Content:  "This is a test announcement",
			AuthorID: "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_announcements`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, announcement)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create announcement")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "content", "author_id", "is_pinned", "created_at", "updated_at",
			"email", "locale",
		}).AddRow(
			"ann-123", "tenant-1", "Test Title", "Test Content", "user-1", true, now, now,
			"author@example.com", "en",
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`)).
			WithArgs("ann-123").
			WillReturnRows(rows)

		announcement, err := repo.GetByID(ctx, "ann-123")
		require.NoError(t, err)
		assert.Equal(t, "ann-123", announcement.ID)
		assert.Equal(t, "tenant-1", announcement.TenantID)
		assert.Equal(t, "Test Title", announcement.Title)
		assert.Equal(t, "Test Content", announcement.Content)
		assert.Equal(t, "user-1", announcement.AuthorID)
		assert.True(t, announcement.IsPinned)
		assert.NotNil(t, announcement.Author)
		assert.Equal(t, "author@example.com", announcement.Author.Email)
		assert.Equal(t, "en", announcement.Author.Locale)
		assert.Equal(t, "user-1", announcement.Author.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`)).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		announcement, err := repo.GetByID(ctx, "non-existent")
		require.Error(t, err)
		assert.Nil(t, announcement)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`)).
			WithArgs("ann-123").
			WillReturnError(errors.New("database error"))

		announcement, err := repo.GetByID(ctx, "ann-123")
		require.Error(t, err)
		assert.Nil(t, announcement)
		assert.Contains(t, err.Error(), "failed to get announcement")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			ID:       "ann-123",
			Title:    "Updated Title",
			Content:  "Updated Content",
			IsPinned: true,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements`)).
			WithArgs(
				announcement.ID,
				announcement.Title,
				announcement.Content,
				announcement.IsPinned,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, announcement)
		require.NoError(t, err)
		assert.NotZero(t, announcement.UpdatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			ID:       "non-existent",
			Title:    "Updated Title",
			Content:  "Updated Content",
			IsPinned: false,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements`)).
			WithArgs(
				announcement.ID,
				announcement.Title,
				announcement.Content,
				announcement.IsPinned,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, announcement)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		announcement := &domain.TenantAnnouncement{
			ID:       "ann-123",
			Title:    "Updated Title",
			Content:  "Updated Content",
			IsPinned: false,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, announcement)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update announcement")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE id = $1`)).
			WithArgs("ann-123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "ann-123")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE id = $1`)).
			WithArgs("non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "non-existent")
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE id = $1`)).
			WithArgs("ann-123").
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, "ann-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete announcement")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with no filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "content", "author_id", "is_pinned", "created_at", "updated_at",
			"email", "locale",
		}).
			AddRow("ann-1", "tenant-1", "Title 1", "Content 1", "user-1", true, now, now, "user1@example.com", "en").
			AddRow("ann-2", "tenant-1", "Title 2", "Content 2", "user-2", false, now, now, "user2@example.com", "fr")

		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		announcements, nextCursor, err := repo.ListByTenant(ctx, "tenant-1", false, 0, "")
		require.NoError(t, err)
		assert.Len(t, announcements, 2)
		assert.Empty(t, nextCursor)
		assert.Equal(t, "ann-1", announcements[0].ID)
		assert.Equal(t, "ann-2", announcements[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with pinned only filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "content", "author_id", "is_pinned", "created_at", "updated_at",
			"email", "locale",
		}).
			AddRow("ann-1", "tenant-1", "Pinned Title", "Content", "user-1", true, now, now, "user1@example.com", "en")

		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		announcements, _, err := repo.ListByTenant(ctx, "tenant-1", true, 0, "")
		require.NoError(t, err)
		assert.Len(t, announcements, 1)
		assert.True(t, announcements[0].IsPinned)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with cursor", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "content", "author_id", "is_pinned", "created_at", "updated_at",
			"email", "locale",
		}).
			AddRow("ann-2", "tenant-1", "Title 2", "Content 2", "user-1", false, now, now, "user1@example.com", "en")

		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1", "ann-1").
			WillReturnRows(rows)

		announcements, _, err := repo.ListByTenant(ctx, "tenant-1", false, 0, "ann-1")
		require.NoError(t, err)
		assert.Len(t, announcements, 1)
		assert.Equal(t, "ann-2", announcements[0].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with limit and next cursor", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "content", "author_id", "is_pinned", "created_at", "updated_at",
			"email", "locale",
		}).
			AddRow("ann-1", "tenant-1", "Title 1", "Content 1", "user-1", true, now, now, "user1@example.com", "en").
			AddRow("ann-2", "tenant-1", "Title 2", "Content 2", "user-2", false, now, now, "user2@example.com", "fr").
			AddRow("ann-3", "tenant-1", "Title 3", "Content 3", "user-3", false, now, now, "user3@example.com", "de")

		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		announcements, nextCursor, err := repo.ListByTenant(ctx, "tenant-1", false, 2, "")
		require.NoError(t, err)
		assert.Len(t, announcements, 2)
		assert.Equal(t, "ann-2", nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		announcements, nextCursor, err := repo.ListByTenant(ctx, "tenant-1", false, 0, "")
		require.Error(t, err)
		assert.Nil(t, announcements)
		assert.Empty(t, nextCursor)
		assert.Contains(t, err.Error(), "failed to list announcements")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id",
		}).
			AddRow("ann-1", "tenant-1")

		mock.ExpectQuery(`SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		announcements, nextCursor, err := repo.ListByTenant(ctx, "tenant-1", false, 0, "")
		require.Error(t, err)
		assert.Nil(t, announcements)
		assert.Empty(t, nextCursor)
		assert.Contains(t, err.Error(), "failed to scan announcement")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_SetPinned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()

	t.Run("success pin", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements SET is_pinned = $2, updated_at = $3 WHERE id = $1`)).
			WithArgs("ann-123", true, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetPinned(ctx, "ann-123", true)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success unpin", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements SET is_pinned = $2, updated_at = $3 WHERE id = $1`)).
			WithArgs("ann-123", false, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetPinned(ctx, "ann-123", false)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements SET is_pinned = $2, updated_at = $3 WHERE id = $1`)).
			WithArgs("non-existent", true, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SetPinned(ctx, "non-existent", true)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_announcements SET is_pinned = $2, updated_at = $3 WHERE id = $1`)).
			WithArgs("ann-123", true, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.SetPinned(ctx, "ann-123", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update pinned status")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantAnnouncementRepository_DeleteAllByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantAnnouncementRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnResult(sqlmock.NewResult(0, 5))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success no rows affected", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE tenant_id = $1`)).
			WithArgs("tenant-empty").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteAllByTenant(ctx, "tenant-empty")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_announcements WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant announcements")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
