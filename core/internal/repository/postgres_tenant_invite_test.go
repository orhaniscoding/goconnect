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

func TestNewPostgresTenantInviteRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresTenantInviteRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success with generated ID", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		invite := &domain.TenantInvite{
			TenantID:  "tenant-1",
			Code:      "abc123",
			MaxUses:   10,
			UseCount:  0,
			ExpiresAt: &expiresAt,
			CreatedBy: "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_invites (id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at)`)).
			WithArgs(sqlmock.AnyArg(), "tenant-1", "ABC123", 10, 0, &expiresAt, "user-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, invite)
		require.NoError(t, err)
		assert.NotEmpty(t, invite.ID)
		assert.NotZero(t, invite.CreatedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with provided ID", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		invite := &domain.TenantInvite{
			ID:        "invite-123",
			TenantID:  "tenant-1",
			Code:      "xyz789",
			MaxUses:   5,
			UseCount:  0,
			ExpiresAt: &expiresAt,
			CreatedBy: "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_invites (id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at)`)).
			WithArgs("invite-123", "tenant-1", "XYZ789", 5, 0, &expiresAt, "user-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, invite)
		require.NoError(t, err)
		assert.Equal(t, "invite-123", invite.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate code error", func(t *testing.T) {
		invite := &domain.TenantInvite{
			TenantID:  "tenant-1",
			Code:      "dup123",
			MaxUses:   10,
			CreatedBy: "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_invites`)).
			WillReturnError(errors.New("pq: duplicate key value violates unique constraint"))

		err := repo.Create(ctx, invite)
		require.Error(t, err)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		invite := &domain.TenantInvite{
			TenantID:  "tenant-1",
			Code:      "err123",
			MaxUses:   10,
			CreatedBy: "user-1",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_invites`)).
			WillReturnError(errors.New("connection refused"))

		err := repo.Create(ctx, invite)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create tenant invite")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with all fields", func(t *testing.T) {
		expiresAt := now.Add(24 * time.Hour)
		revokedAt := now.Add(12 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-1", "tenant-1", "ABC123", 10, 5, expiresAt, "user-1", now, revokedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnRows(rows)

		invite, err := repo.GetByID(ctx, "invite-1")
		require.NoError(t, err)
		assert.Equal(t, "invite-1", invite.ID)
		assert.Equal(t, "tenant-1", invite.TenantID)
		assert.Equal(t, "ABC123", invite.Code)
		assert.Equal(t, 10, invite.MaxUses)
		assert.Equal(t, 5, invite.UseCount)
		assert.NotNil(t, invite.ExpiresAt)
		assert.NotNil(t, invite.RevokedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null optional fields", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-2", "tenant-1", "DEF456", 0, 0, nil, "user-1", now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE id = $1`)).
			WithArgs("invite-2").
			WillReturnRows(rows)

		invite, err := repo.GetByID(ctx, "invite-2")
		require.NoError(t, err)
		assert.Equal(t, "invite-2", invite.ID)
		assert.Nil(t, invite.ExpiresAt)
		assert.Nil(t, invite.RevokedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE id = $1`)).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		invite, err := repo.GetByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, invite)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnError(errors.New("connection error"))

		invite, err := repo.GetByID(ctx, "invite-1")
		require.Error(t, err)
		assert.Nil(t, invite)
		assert.Contains(t, err.Error(), "failed to get tenant invite")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_GetByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		expiresAt := now.Add(24 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-1", "tenant-1", "ABC123", 10, 2, expiresAt, "user-1", now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE UPPER(code) = UPPER($1)`)).
			WithArgs("abc123").
			WillReturnRows(rows)

		invite, err := repo.GetByCode(ctx, "abc123")
		require.NoError(t, err)
		assert.Equal(t, "invite-1", invite.ID)
		assert.Equal(t, "ABC123", invite.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with whitespace trimmed", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-1", "tenant-1", "XYZ789", 10, 0, nil, "user-1", now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE UPPER(code) = UPPER($1)`)).
			WithArgs("XYZ789").
			WillReturnRows(rows)

		invite, err := repo.GetByCode(ctx, "  XYZ789  ")
		require.NoError(t, err)
		assert.Equal(t, "invite-1", invite.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE UPPER(code) = UPPER($1)`)).
			WithArgs("INVALID").
			WillReturnError(sql.ErrNoRows)

		invite, err := repo.GetByCode(ctx, "INVALID")
		require.Error(t, err)
		assert.Nil(t, invite)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		assert.Contains(t, domainErr.Message, "Invalid invite code")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE UPPER(code) = UPPER($1)`)).
			WithArgs("ABC123").
			WillReturnError(errors.New("database connection lost"))

		invite, err := repo.GetByCode(ctx, "ABC123")
		require.Error(t, err)
		assert.Nil(t, invite)
		assert.Contains(t, err.Error(), "failed to get invite by code")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "invite-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE id = $1`)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "nonexistent")
		require.Error(t, err)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, "invite-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant invite")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with multiple invites", func(t *testing.T) {
		expiresAt := now.Add(24 * time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).
			AddRow("invite-1", "tenant-1", "ABC123", 10, 5, expiresAt, "user-1", now, nil).
			AddRow("invite-2", "tenant-1", "DEF456", 0, 0, nil, "user-2", now.Add(-time.Hour), nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE tenant_id = $1 ORDER BY created_at DESC`)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		invites, err := repo.ListByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		assert.Len(t, invites, 2)
		assert.Equal(t, "invite-1", invites[0].ID)
		assert.Equal(t, "invite-2", invites[1].ID)
		assert.NotNil(t, invites[0].ExpiresAt)
		assert.Nil(t, invites[1].ExpiresAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE tenant_id = $1 ORDER BY created_at DESC`)).
			WithArgs("empty-tenant").
			WillReturnRows(rows)

		invites, err := repo.ListByTenant(ctx, "empty-tenant")
		require.NoError(t, err)
		assert.Empty(t, invites)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with revoked invite", func(t *testing.T) {
		revokedAt := now.Add(-time.Hour)

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-1", "tenant-1", "ABC123", 10, 5, nil, "user-1", now, revokedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE tenant_id = $1 ORDER BY created_at DESC`)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		invites, err := repo.ListByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		assert.Len(t, invites, 1)
		assert.NotNil(t, invites[0].RevokedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on query", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE tenant_id = $1 ORDER BY created_at DESC`)).
			WithArgs("tenant-1").
			WillReturnError(errors.New("connection timeout"))

		invites, err := repo.ListByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Nil(t, invites)
		assert.Contains(t, err.Error(), "failed to list tenant invites")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "code", "max_uses", "use_count", "expires_at", "created_by", "created_at", "revoked_at",
		}).AddRow("invite-1", "tenant-1", "ABC123", "invalid-int", 0, nil, "user-1", now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at FROM tenant_invites WHERE tenant_id = $1 ORDER BY created_at DESC`)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		invites, err := repo.ListByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Nil(t, invites)
		assert.Contains(t, err.Error(), "failed to scan tenant invite")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_IncrementUseCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET use_count = use_count + 1 WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.IncrementUseCount(ctx, "invite-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET use_count = use_count + 1 WHERE id = $1`)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.IncrementUseCount(ctx, "nonexistent")
		require.Error(t, err)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET use_count = use_count + 1 WHERE id = $1`)).
			WithArgs("invite-1").
			WillReturnError(errors.New("database error"))

		err := repo.IncrementUseCount(ctx, "invite-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to increment use count")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_Revoke(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET revoked_at = $2 WHERE id = $1`)).
			WithArgs("invite-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Revoke(ctx, "invite-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET revoked_at = $2 WHERE id = $1`)).
			WithArgs("nonexistent", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Revoke(ctx, "nonexistent")
		require.Error(t, err)
		var domainErr *domain.Error
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenant_invites SET revoked_at = $2 WHERE id = $1`)).
			WithArgs("invite-1", sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Revoke(ctx, "invite-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke tenant invite")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success with deleted rows", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE (expires_at IS NOT NULL AND expires_at < NOW()) OR revoked_at IS NOT NULL`)).
			WillReturnResult(sqlmock.NewResult(0, 5))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no expired invites", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE (expires_at IS NOT NULL AND expires_at < NOW()) OR revoked_at IS NOT NULL`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE (expires_at IS NOT NULL AND expires_at < NOW()) OR revoked_at IS NOT NULL`)).
			WillReturnError(errors.New("database error"))

		count, err := repo.DeleteExpired(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to delete expired invites")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantInviteRepository_DeleteAllByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantInviteRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnResult(sqlmock.NewResult(0, 3))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no invites to delete", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE tenant_id = $1`)).
			WithArgs("empty-tenant").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteAllByTenant(ctx, "empty-tenant")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_invites WHERE tenant_id = $1`)).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant invites")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
