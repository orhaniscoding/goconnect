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

func TestNewPostgresInviteTokenRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresInviteTokenRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success", func(t *testing.T) {
		token := &domain.InviteToken{
			ID:        "invite-123",
			NetworkID: "network-456",
			TenantID:  "tenant-789",
			Token:     "abc123def456",
			CreatedBy: "user-001",
			ExpiresAt: expiresAt,
			UsesMax:   10,
			UsesLeft:  10,
			CreatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO invite_tokens (id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`)).
			WithArgs(token.ID, token.NetworkID, token.TenantID, token.Token, token.CreatedBy, token.ExpiresAt, token.UsesMax, token.UsesLeft, token.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, token)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		token := &domain.InviteToken{
			ID:        "invite-error",
			NetworkID: "network-456",
			TenantID:  "tenant-789",
			Token:     "error-token",
			CreatedBy: "user-001",
			ExpiresAt: expiresAt,
			UsesMax:   5,
			UsesLeft:  5,
			CreatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO invite_tokens (id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`)).
			WithArgs(token.ID, token.NetworkID, token.TenantID, token.Token, token.CreatedBy, token.ExpiresAt, token.UsesMax, token.UsesLeft, token.CreatedAt).
			WillReturnError(errors.New("duplicate key violation"))

		err := repo.Create(ctx, token)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create invite token")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success without revoked_at", func(t *testing.T) {
		id := "invite-123"

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow(id, "network-456", "tenant-789", "abc123", "user-001", expiresAt, 10, 8, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE id = $1`)).
			WithArgs(id).
			WillReturnRows(rows)

		token, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, id, token.ID)
		assert.Equal(t, "network-456", token.NetworkID)
		assert.Equal(t, "tenant-789", token.TenantID)
		assert.Equal(t, "abc123", token.Token)
		assert.Equal(t, 10, token.UsesMax)
		assert.Equal(t, 8, token.UsesLeft)
		assert.Nil(t, token.RevokedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with revoked_at", func(t *testing.T) {
		id := "invite-revoked"
		revokedAt := now.Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow(id, "network-456", "tenant-789", "revoked123", "user-001", expiresAt, 10, 5, now, revokedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE id = $1`)).
			WithArgs(id).
			WillReturnRows(rows)

		token, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.NotNil(t, token.RevokedAt)
		assert.Equal(t, revokedAt.Unix(), token.RevokedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		id := "nonexistent"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE id = $1`)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		token, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		id := "error-id"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE id = $1`)).
			WithArgs(id).
			WillReturnError(errors.New("connection lost"))

		token, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to get invite token")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_GetByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success", func(t *testing.T) {
		tokenStr := "abc123def456"

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-123", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 8, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		token, err := repo.GetByToken(ctx, tokenStr)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, tokenStr, token.Token)
		assert.Equal(t, "invite-123", token.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		tokenStr := "nonexistent-token"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1`)).
			WithArgs(tokenStr).
			WillReturnError(sql.ErrNoRows)

		token, err := repo.GetByToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		tokenStr := "error-token"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1`)).
			WithArgs(tokenStr).
			WillReturnError(errors.New("database timeout"))

		token, err := repo.GetByToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to get invite token")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_ListByNetwork(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success with multiple tokens", func(t *testing.T) {
		networkID := "network-456"

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-1", networkID, "tenant-789", "token1", "user-001", expiresAt, 10, 8, now, nil).
			AddRow("invite-2", networkID, "tenant-789", "token2", "user-002", expiresAt, 5, 3, now.Add(-1*time.Hour), nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE network_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		tokens, err := repo.ListByNetwork(ctx, networkID)
		require.NoError(t, err)
		assert.Len(t, tokens, 2)
		assert.Equal(t, "invite-1", tokens[0].ID)
		assert.Equal(t, "invite-2", tokens[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty result", func(t *testing.T) {
		networkID := "empty-network"

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE network_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		tokens, err := repo.ListByNetwork(ctx, networkID)
		require.NoError(t, err)
		assert.Empty(t, tokens)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		networkID := "error-network"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE network_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`)).
			WithArgs(networkID).
			WillReturnError(errors.New("query failed"))

		tokens, err := repo.ListByNetwork(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "failed to list invite tokens")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		networkID := "scan-error-network"

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "network_id"}).
			AddRow("invite-1", networkID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE network_id = $1 AND revoked_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		tokens, err := repo.ListByNetwork(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "failed to scan invite token")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_UseToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success with limited uses", func(t *testing.T) {
		tokenStr := "valid-token"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-123", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 5, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET uses_left = $1 WHERE id = $2`)).
			WithArgs(4, "invite-123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		token, err := repo.UseToken(ctx, tokenStr)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, 4, token.UsesLeft)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with unlimited uses", func(t *testing.T) {
		tokenStr := "unlimited-token"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-unlimited", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 0, 0, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		// No UPDATE expected since UsesMax is 0 (unlimited)

		mock.ExpectCommit()

		token, err := repo.UseToken(ctx, tokenStr)
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, 0, token.UsesMax)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("token not found", func(t *testing.T) {
		tokenStr := "nonexistent-token"

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("token revoked", func(t *testing.T) {
		tokenStr := "revoked-token"
		revokedAt := now.Add(-1 * time.Hour)

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-revoked", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 5, now, revokedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectRollback()

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenRevoked, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("token expired", func(t *testing.T) {
		tokenStr := "expired-token"
		expiredAt := now.Add(-1 * time.Hour)

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-expired", "network-456", "tenant-789", tokenStr, "user-001", expiredAt, 10, 5, now.Add(-2*time.Hour), nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectRollback()

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenExpired, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no uses left", func(t *testing.T) {
		tokenStr := "exhausted-token"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-exhausted", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 0, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectRollback()

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenExpired, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		tokenStr := "tx-error-token"

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to start transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		tokenStr := "update-error-token"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-update-err", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 5, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET uses_left = $1 WHERE id = $2`)).
			WithArgs(4, "invite-update-err").
			WillReturnError(errors.New("update failed"))

		mock.ExpectRollback()

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to update uses_left")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		tokenStr := "commit-error-token"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "network_id", "tenant_id", "token", "created_by", "expires_at", "uses_max", "uses_left", "created_at", "revoked_at"}).
			AddRow("invite-commit-err", "network-456", "tenant-789", tokenStr, "user-001", expiresAt, 10, 5, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at FROM invite_tokens WHERE token = $1 FOR UPDATE`)).
			WithArgs(tokenStr).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET uses_left = $1 WHERE id = $2`)).
			WithArgs(4, "invite-commit-err").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		token, err := repo.UseToken(ctx, tokenStr)
		require.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_Revoke(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := "invite-to-revoke"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`)).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Revoke(ctx, id)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or already revoked", func(t *testing.T) {
		id := "nonexistent-or-revoked"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`)).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Revoke(ctx, id)
		require.Error(t, err)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInviteTokenNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		id := "error-revoke-id"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`)).
			WithArgs(id).
			WillReturnError(errors.New("revoke failed"))

		err := repo.Revoke(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke invite token")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		id := "rows-affected-error"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE invite_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`)).
			WithArgs(id).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Revoke(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresInviteTokenRepository_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresInviteTokenRepository(db)
	ctx := context.Background()

	t.Run("success with deleted rows", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM invite_tokens WHERE expires_at < NOW()`)).
			WillReturnResult(sqlmock.NewResult(0, 5))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no deleted rows", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM invite_tokens WHERE expires_at < NOW()`)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM invite_tokens WHERE expires_at < NOW()`)).
			WillReturnError(errors.New("delete failed"))

		count, err := repo.DeleteExpired(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to delete expired tokens")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM invite_tokens WHERE expires_at < NOW()`)).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		count, err := repo.DeleteExpired(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
