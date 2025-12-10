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

func TestNewPostgresTenantMemberRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresTenantMemberRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success with existing ID", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:       "member-1",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Role:     domain.TenantRoleMember,
			Nickname: "testuser",
		}

		query := `INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.TenantID,
				member.UserID,
				member.Role,
				sql.NullString{String: "testuser", Valid: true},
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, member)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without ID - generates ID", func(t *testing.T) {
		member := &domain.TenantMember{
			TenantID: "tenant-1",
			UserID:   "user-2",
			Role:     domain.TenantRoleMember,
		}

		query := `INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				sqlmock.AnyArg(), // Generated ID
				member.TenantID,
				member.UserID,
				member.Role,
				sql.NullString{Valid: false}, // Empty nickname
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, member)
		require.NoError(t, err)
		assert.NotEmpty(t, member.ID) // ID should be generated
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate key error", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:       "member-dup",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Role:     domain.TenantRoleMember,
		}

		query := `INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.TenantID,
				member.UserID,
				member.Role,
				sql.NullString{Valid: false},
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("duplicate key value violates unique constraint 23505"))

		err := repo.Create(ctx, member)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:       "member-err",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Role:     domain.TenantRoleMember,
		}

		query := `INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.TenantID,
				member.UserID,
				member.Role,
				sql.NullString{Valid: false},
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("connection error"))

		err := repo.Create(ctx, member)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create tenant member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE id = $1`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at",
		}).AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "testuser", now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("member-1").
			WillReturnRows(rows)

		member, err := repo.GetByID(ctx, "member-1")
		require.NoError(t, err)
		assert.Equal(t, "member-1", member.ID)
		assert.Equal(t, "tenant-1", member.TenantID)
		assert.Equal(t, "user-1", member.UserID)
		assert.Equal(t, domain.TenantRoleMember, member.Role)
		assert.Equal(t, "testuser", member.Nickname)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE id = $1`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		member, err := repo.GetByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, member)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE id = $1`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("member-err").
			WillReturnError(errors.New("connection error"))

		member, err := repo.GetByID(ctx, "member-err")
		require.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "failed to get tenant member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_GetByUserAndTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE user_id = $1 AND tenant_id = $2`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at",
		}).AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleAdmin, "admin", now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(rows)

		member, err := repo.GetByUserAndTenant(ctx, "user-1", "tenant-1")
		require.NoError(t, err)
		assert.Equal(t, "member-1", member.ID)
		assert.Equal(t, domain.TenantRoleAdmin, member.Role)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-nonexistent", "tenant-1").
			WillReturnError(sql.ErrNoRows)

		member, err := repo.GetByUserAndTenant(ctx, "user-nonexistent", "tenant-1")
		require.Error(t, err)
		assert.Nil(t, member)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnError(errors.New("connection error"))

		member, err := repo.GetByUserAndTenant(ctx, "user-1", "tenant-1")
		require.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "failed to get tenant membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:       "member-1",
			Role:     domain.TenantRoleAdmin,
			Nickname: "newname",
		}

		query := `UPDATE tenant_members
		SET role = $2, nickname = $3, updated_at = $4
		WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.Role,
				sql.NullString{String: "newname", Valid: true},
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, member)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:   "nonexistent",
			Role: domain.TenantRoleMember,
		}

		query := `UPDATE tenant_members
		SET role = $2, nickname = $3, updated_at = $4
		WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.Role,
				sql.NullString{Valid: false},
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, member)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		member := &domain.TenantMember{
			ID:   "member-err",
			Role: domain.TenantRoleMember,
		}

		query := `UPDATE tenant_members
		SET role = $2, nickname = $3, updated_at = $4
		WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				member.ID,
				member.Role,
				sql.NullString{Valid: false},
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("connection error"))

		err := repo.Update(ctx, member)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update tenant member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "member-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "nonexistent")
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-err").
			WillReturnError(errors.New("connection error"))

		err := repo.Delete(ctx, "member-err")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without filters", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "email", "locale",
		}).
			AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "nick1", now, now, "user1@example.com", "en-US").
			AddRow("member-2", "tenant-1", "user-2", domain.TenantRoleAdmin, "nick2", now, now, "user2@example.com", "fr-FR")

		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		members, cursor, err := repo.ListByTenant(ctx, "tenant-1", "", 0, "")
		require.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Empty(t, cursor)
		assert.Equal(t, "member-1", members[0].ID)
		assert.Equal(t, "user1@example.com", members[0].User.Email)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with role filter", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "email", "locale",
		}).AddRow("member-2", "tenant-1", "user-2", domain.TenantRoleAdmin, "admin", now, now, "admin@example.com", "en-US")

		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1", "admin").
			WillReturnRows(rows)

		members, _, err := repo.ListByTenant(ctx, "tenant-1", "admin", 0, "")
		require.NoError(t, err)
		assert.Len(t, members, 1)
		assert.Equal(t, domain.TenantRoleAdmin, members[0].Role)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with cursor", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "email", "locale",
		}).AddRow("member-3", "tenant-1", "user-3", domain.TenantRoleMember, "nick3", now, now, "user3@example.com", "en-US")

		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1", "member-2").
			WillReturnRows(rows)

		members, _, err := repo.ListByTenant(ctx, "tenant-1", "", 0, "member-2")
		require.NoError(t, err)
		assert.Len(t, members, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with limit and next cursor", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "email", "locale",
		}).
			AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "nick1", now, now, "user1@example.com", "en-US").
			AddRow("member-2", "tenant-1", "user-2", domain.TenantRoleMember, "nick2", now, now, "user2@example.com", "en-US").
			AddRow("member-3", "tenant-1", "user-3", domain.TenantRoleMember, "nick3", now, now, "user3@example.com", "en-US") // Extra for cursor detection

		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		members, cursor, err := repo.ListByTenant(ctx, "tenant-1", "", 2, "")
		require.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "member-2", cursor) // Next cursor should be the last returned item
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1").
			WillReturnError(errors.New("connection error"))

		members, cursor, err := repo.ListByTenant(ctx, "tenant-1", "", 0, "")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "failed to list tenant members")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Return wrong number of columns
		rows := sqlmock.NewRows([]string{"id", "tenant_id"}).
			AddRow("member-1", "tenant-1")

		mock.ExpectQuery(`SELECT tm.id, tm.tenant_id, tm.user_id, tm.role`).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		members, cursor, err := repo.ListByTenant(ctx, "tenant-1", "", 0, "")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "failed to scan tenant member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_ListByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		query := `SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at
		FROM tenant_members tm
		WHERE tm.user_id = $1
		ORDER BY tm.joined_at DESC`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at",
		}).
			AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "nick1", now, now).
			AddRow("member-2", "tenant-2", "user-1", domain.TenantRoleAdmin, "nick2", now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnRows(rows)

		members, err := repo.ListByUser(ctx, "user-1")
		require.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "member-1", members[0].ID)
		assert.Equal(t, "tenant-1", members[0].TenantID)
		assert.Equal(t, "tenant-2", members[1].TenantID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		query := `SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at
		FROM tenant_members tm
		WHERE tm.user_id = $1
		ORDER BY tm.joined_at DESC`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-no-memberships").
			WillReturnRows(rows)

		members, err := repo.ListByUser(ctx, "user-no-memberships")
		require.NoError(t, err)
		assert.Len(t, members, 0)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at
		FROM tenant_members tm
		WHERE tm.user_id = $1
		ORDER BY tm.joined_at DESC`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnError(errors.New("connection error"))

		members, err := repo.ListByUser(ctx, "user-1")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "failed to list user's tenant memberships")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		query := `SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at
		FROM tenant_members tm
		WHERE tm.user_id = $1
		ORDER BY tm.joined_at DESC`

		// Return wrong number of columns
		rows := sqlmock.NewRows([]string{"id"}).AddRow("member-1")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnRows(rows)

		members, err := repo.ListByUser(ctx, "user-1")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "failed to scan membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_CountByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `SELECT COUNT(*) FROM tenant_members WHERE tenant_id = $1`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

		count, err := repo.CountByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT COUNT(*) FROM tenant_members WHERE tenant_id = $1`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-err").
			WillReturnError(errors.New("connection error"))

		count, err := repo.CountByTenant(ctx, "tenant-err")
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to count tenant members")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_UpdateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `UPDATE tenant_members SET role = $2, updated_at = $3 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-1", domain.TenantRoleAdmin, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateRole(ctx, "member-1", domain.TenantRoleAdmin)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `UPDATE tenant_members SET role = $2, updated_at = $3 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent", domain.TenantRoleMember, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateRole(ctx, "nonexistent", domain.TenantRoleMember)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `UPDATE tenant_members SET role = $2, updated_at = $3 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-err", domain.TenantRoleMember, sqlmock.AnyArg()).
			WillReturnError(errors.New("connection error"))

		err := repo.UpdateRole(ctx, "member-err", domain.TenantRoleMember)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update role")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_GetUserRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(domain.TenantRoleAdmin))

		role, err := repo.GetUserRole(ctx, "user-1", "tenant-1")
		require.NoError(t, err)
		assert.Equal(t, domain.TenantRoleAdmin, role)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-nonexistent", "tenant-1").
			WillReturnError(sql.ErrNoRows)

		role, err := repo.GetUserRole(ctx, "user-nonexistent", "tenant-1")
		require.Error(t, err)
		assert.Empty(t, role)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-err").
			WillReturnError(errors.New("connection error"))

		role, err := repo.GetUserRole(ctx, "user-1", "tenant-err")
		require.Error(t, err)
		assert.Empty(t, role)
		assert.Contains(t, err.Error(), "failed to get user role")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_HasRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("has required role", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(domain.TenantRoleAdmin))

		hasRole, err := repo.HasRole(ctx, "user-1", "tenant-1", domain.TenantRoleMember)
		require.NoError(t, err)
		assert.True(t, hasRole) // Admin has Member permissions
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("does not have required role", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow(domain.TenantRoleMember))

		hasRole, err := repo.HasRole(ctx, "user-1", "tenant-1", domain.TenantRoleAdmin)
		require.NoError(t, err)
		assert.False(t, hasRole) // Member does not have Admin permissions
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not a member - returns false", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-nonmember", "tenant-1").
			WillReturnError(sql.ErrNoRows)

		hasRole, err := repo.HasRole(ctx, "user-nonmember", "tenant-1", domain.TenantRoleMember)
		require.NoError(t, err)
		assert.False(t, hasRole)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-err").
			WillReturnError(errors.New("connection error"))

		hasRole, err := repo.HasRole(ctx, "user-1", "tenant-err", domain.TenantRoleMember)
		require.Error(t, err)
		assert.False(t, hasRole)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_DeleteAllByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE tenant_id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnResult(sqlmock.NewResult(0, 5))

		err := repo.DeleteAllByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no rows deleted", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE tenant_id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("empty-tenant").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteAllByTenant(ctx, "empty-tenant")
		require.NoError(t, err) // Should not error even if no rows deleted
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `DELETE FROM tenant_members WHERE tenant_id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("tenant-err").
			WillReturnError(errors.New("connection error"))

		err := repo.DeleteAllByTenant(ctx, "tenant-err")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant members")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_Ban(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = $2, banned_by = $3, updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-1", sqlmock.AnyArg(), "admin-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Ban(ctx, "member-1", "admin-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = $2, banned_by = $3, updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent", sqlmock.AnyArg(), "admin-1").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Ban(ctx, "nonexistent", "admin-1")
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = $2, banned_by = $3, updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-err", sqlmock.AnyArg(), "admin-1").
			WillReturnError(errors.New("connection error"))

		err := repo.Ban(ctx, "member-err", "admin-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ban member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_Unban(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = NULL, banned_by = '', updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Unban(ctx, "member-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = NULL, banned_by = '', updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Unban(ctx, "nonexistent")
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `UPDATE tenant_members SET banned_at = NULL, banned_by = '', updated_at = $2 WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("member-err", sqlmock.AnyArg()).
			WillReturnError(errors.New("connection error"))

		err := repo.Unban(ctx, "member-err")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unban member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_IsBanned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()

	t.Run("user is banned", func(t *testing.T) {
		query := `SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"is_banned"}).AddRow(true))

		isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-1")
		require.NoError(t, err)
		assert.True(t, isBanned)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user is not banned", func(t *testing.T) {
		query := `SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-1").
			WillReturnRows(sqlmock.NewRows([]string{"is_banned"}).AddRow(false))

		isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-1")
		require.NoError(t, err)
		assert.False(t, isBanned)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not a member - returns false", func(t *testing.T) {
		query := `SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-nonmember", "tenant-1").
			WillReturnError(sql.ErrNoRows)

		isBanned, err := repo.IsBanned(ctx, "user-nonmember", "tenant-1")
		require.NoError(t, err)
		assert.False(t, isBanned)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1", "tenant-err").
			WillReturnError(errors.New("connection error"))

		isBanned, err := repo.IsBanned(ctx, "user-1", "tenant-err")
		require.Error(t, err)
		assert.False(t, isBanned)
		assert.Contains(t, err.Error(), "failed to check ban status")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantMemberRepository_ListBannedByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantMemberRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by 
              FROM tenant_members WHERE tenant_id = $1 AND banned_at IS NOT NULL ORDER BY banned_at DESC`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "banned_at", "banned_by",
		}).
			AddRow("member-1", "tenant-1", "user-1", domain.TenantRoleMember, "nick1", now, now, now, "admin-1").
			AddRow("member-2", "tenant-1", "user-2", domain.TenantRoleMember, nil, now, now, now, "admin-2")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		members, err := repo.ListBannedByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "member-1", members[0].ID)
		assert.Equal(t, "nick1", members[0].Nickname)
		assert.NotNil(t, members[0].BannedAt)
		assert.Equal(t, "admin-1", members[0].BannedBy)
		assert.Empty(t, members[1].Nickname) // nil nickname
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by 
              FROM tenant_members WHERE tenant_id = $1 AND banned_at IS NOT NULL ORDER BY banned_at DESC`

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "user_id", "role", "nickname", "joined_at", "updated_at", "banned_at", "banned_by",
		})

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-no-bans").
			WillReturnRows(rows)

		members, err := repo.ListBannedByTenant(ctx, "tenant-no-bans")
		require.NoError(t, err)
		assert.Len(t, members, 0)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by 
              FROM tenant_members WHERE tenant_id = $1 AND banned_at IS NOT NULL ORDER BY banned_at DESC`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-err").
			WillReturnError(errors.New("connection error"))

		members, err := repo.ListBannedByTenant(ctx, "tenant-err")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "failed to list banned members")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		query := `SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by 
              FROM tenant_members WHERE tenant_id = $1 AND banned_at IS NOT NULL ORDER BY banned_at DESC`

		// Return wrong number of columns
		rows := sqlmock.NewRows([]string{"id"}).AddRow("member-1")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		members, err := repo.ListBannedByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "failed to scan banned member")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNullString(t *testing.T) {
	t.Run("empty string returns invalid NullString", func(t *testing.T) {
		result := nullString("")
		assert.False(t, result.Valid)
		assert.Empty(t, result.String)
	})

	t.Run("non-empty string returns valid NullString", func(t *testing.T) {
		result := nullString("test")
		assert.True(t, result.Valid)
		assert.Equal(t, "test", result.String)
	})
}

func TestIsDuplicateKeyError(t *testing.T) {
	t.Run("nil error returns false", func(t *testing.T) {
		result := isDuplicateKeyError(nil)
		assert.False(t, result)
	})

	t.Run("error with 23505 returns true", func(t *testing.T) {
		err := errors.New("pq: duplicate key value violates unique constraint (23505)")
		result := isDuplicateKeyError(err)
		assert.True(t, result)
	})

	t.Run("error with duplicate key returns true", func(t *testing.T) {
		err := errors.New("duplicate key violation")
		result := isDuplicateKeyError(err)
		assert.True(t, result)
	})

	t.Run("other error returns false", func(t *testing.T) {
		err := errors.New("connection refused")
		result := isDuplicateKeyError(err)
		assert.False(t, result)
	})
}

func TestContains(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		result := contains("hello", "hello")
		assert.True(t, result)
	})

	t.Run("substring at start", func(t *testing.T) {
		result := contains("hello world", "hello")
		assert.True(t, result)
	})

	t.Run("substring in middle", func(t *testing.T) {
		result := contains("hello world today", "world")
		assert.True(t, result)
	})

	t.Run("substring at end", func(t *testing.T) {
		result := contains("hello world", "world")
		assert.True(t, result)
	})

	t.Run("substring not found", func(t *testing.T) {
		result := contains("hello world", "xyz")
		assert.False(t, result)
	})

	t.Run("empty substring", func(t *testing.T) {
		result := contains("hello", "")
		assert.True(t, result)
	})

	t.Run("substring longer than string", func(t *testing.T) {
		result := contains("hi", "hello")
		assert.False(t, result)
	})

	t.Run("empty string and empty substring", func(t *testing.T) {
		result := contains("", "")
		assert.True(t, result)
	})
}
