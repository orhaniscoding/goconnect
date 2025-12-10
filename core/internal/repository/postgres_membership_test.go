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

func TestNewPostgresMembershipRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresMembershipRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	membership := &domain.Membership{
		ID:        "membership-1",
		NetworkID: "network-1",
		UserID:    "user-1",
		Role:      domain.RoleMember,
		Status:    domain.StatusApproved,
		JoinedAt:  &now,
		UpdatedAt: now,
	}

	t.Run("success", func(t *testing.T) {
		query := `INSERT INTO memberships (id, network_id, user_id, role, status, joined_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				membership.ID,
				membership.NetworkID,
				membership.UserID,
				membership.Role,
				membership.Status,
				membership.JoinedAt,
				membership.UpdatedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, membership)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `INSERT INTO memberships (id, network_id, user_id, role, status, joined_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				membership.ID,
				membership.NetworkID,
				membership.UserID,
				membership.Role,
				membership.Status,
				membership.JoinedAt,
				membership.UpdatedAt,
			).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, membership)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE id = $1`

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("membership-1").
			WillReturnRows(rows)

		membership, err := repo.GetByID(ctx, "membership-1")
		require.NoError(t, err)
		assert.Equal(t, "membership-1", membership.ID)
		assert.Equal(t, "network-1", membership.NetworkID)
		assert.Equal(t, "user-1", membership.UserID)
		assert.Equal(t, domain.RoleMember, membership.Role)
		assert.Equal(t, domain.StatusApproved, membership.Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		membership, err := repo.GetByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, membership)
		assert.Contains(t, err.Error(), "membership not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("membership-1").
			WillReturnError(errors.New("database error"))

		membership, err := repo.GetByID(ctx, "membership-1")
		require.Error(t, err)
		assert.Nil(t, membership)
		assert.Contains(t, err.Error(), "failed to get membership by ID")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_GetByNetworkAndUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND user_id = $2`

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleAdmin, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnRows(rows)

		membership, err := repo.GetByNetworkAndUser(ctx, "network-1", "user-1")
		require.NoError(t, err)
		assert.Equal(t, "membership-1", membership.ID)
		assert.Equal(t, domain.RoleAdmin, membership.Role)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "nonexistent").
			WillReturnError(sql.ErrNoRows)

		membership, err := repo.GetByNetworkAndUser(ctx, "network-1", "nonexistent")
		require.Error(t, err)
		assert.Nil(t, membership)
		assert.Contains(t, err.Error(), "membership not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnError(errors.New("database error"))

		membership, err := repo.GetByNetworkAndUser(ctx, "network-1", "user-1")
		require.Error(t, err)
		assert.Nil(t, membership)
		assert.Contains(t, err.Error(), "failed to get membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND user_id = $2`

	t.Run("success - delegates to GetByNetworkAndUser", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnRows(rows)

		membership, err := repo.Get(ctx, "network-1", "user-1")
		require.NoError(t, err)
		assert.Equal(t, "membership-1", membership.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_ListByNetwork(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY joined_at DESC`

	t.Run("success with results", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).
			AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now).
			AddRow("membership-2", "network-1", "user-2", domain.RoleAdmin, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		memberships, err := repo.ListByNetwork(ctx, "network-1")
		require.NoError(t, err)
		assert.Len(t, memberships, 2)
		assert.Equal(t, "membership-1", memberships[0].ID)
		assert.Equal(t, "membership-2", memberships[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no results", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		memberships, err := repo.ListByNetwork(ctx, "network-1")
		require.NoError(t, err)
		assert.Empty(t, memberships)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnError(errors.New("database error"))

		memberships, err := repo.ListByNetwork(ctx, "network-1")
		require.Error(t, err)
		assert.Nil(t, memberships)
		assert.Contains(t, err.Error(), "failed to list memberships")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, "invalid-time", now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		memberships, err := repo.ListByNetwork(ctx, "network-1")
		require.Error(t, err)
		assert.Nil(t, memberships)
		assert.Contains(t, err.Error(), "failed to scan membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	membership := &domain.Membership{
		ID:        "membership-1",
		NetworkID: "network-1",
		UserID:    "user-1",
		Role:      domain.RoleAdmin,
		Status:    domain.StatusApproved,
		JoinedAt:  &now,
		UpdatedAt: now,
	}

	query := `UPDATE memberships SET role = $1, status = $2, updated_at = $3 WHERE id = $4`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(membership.Role, membership.Status, sqlmock.AnyArg(), membership.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, membership)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(membership.Role, membership.Status, sqlmock.AnyArg(), membership.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, membership)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "membership not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(membership.Role, membership.Status, sqlmock.AnyArg(), membership.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, membership)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(membership.Role, membership.Status, sqlmock.AnyArg(), membership.ID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Update(ctx, membership)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()

	query := `DELETE FROM memberships WHERE id = $1`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("membership-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "membership-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "membership not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("membership-1").
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, "membership-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("membership-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Delete(ctx, "membership-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_SetStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()

	query := `UPDATE memberships SET status = $1, updated_at = $2 WHERE network_id = $3 AND user_id = $4`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(domain.StatusBanned, sqlmock.AnyArg(), "network-1", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetStatus(ctx, "network-1", "user-1", domain.StatusBanned)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(domain.StatusApproved, sqlmock.AnyArg(), "network-1", "nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SetStatus(ctx, "network-1", "nonexistent", domain.StatusApproved)
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(domain.StatusApproved, sqlmock.AnyArg(), "network-1", "user-1").
			WillReturnError(errors.New("database error"))

		err := repo.SetStatus(ctx, "network-1", "user-1", domain.StatusApproved)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set membership status")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(domain.StatusApproved, sqlmock.AnyArg(), "network-1", "user-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.SetStatus(ctx, "network-1", "user-1", domain.StatusApproved)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with no filters", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC LIMIT $2`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).
			AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now).
			AddRow("membership-2", "network-1", "user-2", domain.RoleAdmin, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", 11). // limit + 1 for next page detection
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 10, "")
		require.NoError(t, err)
		assert.Len(t, memberships, 2)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with status filter", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND status = $2 ORDER BY id ASC LIMIT $3`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusPending, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "pending", 11).
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "pending", 10, "")
		require.NoError(t, err)
		assert.Len(t, memberships, 1)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with cursor", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND id > $2 ORDER BY id ASC LIMIT $3`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-3", "network-1", "user-3", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "membership-2", 11).
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 10, "membership-2")
		require.NoError(t, err)
		assert.Len(t, memberships, 1)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with status and cursor", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND status = $2 AND id > $3 ORDER BY id ASC LIMIT $4`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-3", "network-1", "user-3", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", "approved", "membership-2", 11).
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "approved", 10, "membership-2")
		require.NoError(t, err)
		assert.Len(t, memberships, 1)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with next cursor", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC LIMIT $2`

		// Return 3 rows when limit is 2 (limit+1 = 3)
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).
			AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now).
			AddRow("membership-2", "network-1", "user-2", domain.RoleMember, domain.StatusApproved, now, now).
			AddRow("membership-3", "network-1", "user-3", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", 3). // limit + 1
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 2, "")
		require.NoError(t, err)
		assert.Len(t, memberships, 2)
		assert.Equal(t, "membership-2", nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null joined_at", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC LIMIT $2`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusPending, nil, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", 11).
			WillReturnRows(rows)

		memberships, _, err := repo.List(ctx, "network-1", "", 10, "")
		require.NoError(t, err)
		assert.Len(t, memberships, 1)
		assert.Nil(t, memberships[0].JoinedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC LIMIT $2`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", 11).
			WillReturnError(errors.New("database error"))

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 10, "")
		require.Error(t, err)
		assert.Nil(t, memberships)
		assert.Empty(t, nextCursor)
		assert.Contains(t, err.Error(), "failed to list memberships")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC LIMIT $2`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, "invalid-time", now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1", 11).
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 10, "")
		require.Error(t, err)
		assert.Nil(t, memberships)
		assert.Empty(t, nextCursor)
		assert.Contains(t, err.Error(), "failed to scan membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no limit", func(t *testing.T) {
		query := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 ORDER BY id ASC`

		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).
			AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusApproved, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		memberships, nextCursor, err := repo.List(ctx, "network-1", "", 0, "")
		require.NoError(t, err)
		assert.Len(t, memberships, 1)
		assert.Empty(t, nextCursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_Remove(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()

	query := `DELETE FROM memberships WHERE network_id = $1 AND user_id = $2`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Remove(ctx, "network-1", "user-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Remove(ctx, "network-1", "nonexistent")
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnError(errors.New("database error"))

		err := repo.Remove(ctx, "network-1", "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove membership")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Remove(ctx, "network-1", "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresMembershipRepository_UpsertApproved(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMembershipRepository(db)
	ctx := context.Background()
	now := time.Now()

	getQuery := `SELECT id, network_id, user_id, role, status, joined_at, updated_at FROM memberships WHERE network_id = $1 AND user_id = $2`
	updateQuery := `UPDATE memberships SET role = $1, status = $2, updated_at = $3 WHERE id = $4`
	insertQuery := `INSERT INTO memberships (id, network_id, user_id, role, status, joined_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	t.Run("update existing membership", func(t *testing.T) {
		// First, the GetByNetworkAndUser call finds existing membership
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusPending, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(getQuery)).
			WithArgs("network-1", "user-1").
			WillReturnRows(rows)

		// Then, the Update call
		mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
			WithArgs(domain.RoleAdmin, domain.StatusApproved, sqlmock.AnyArg(), "membership-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		membership, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleAdmin, now)
		require.NoError(t, err)
		assert.Equal(t, "membership-1", membership.ID)
		assert.Equal(t, domain.RoleAdmin, membership.Role)
		assert.Equal(t, domain.StatusApproved, membership.Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create new membership", func(t *testing.T) {
		// First, the GetByNetworkAndUser call returns not found
		mock.ExpectQuery(regexp.QuoteMeta(getQuery)).
			WithArgs("network-1", "user-2").
			WillReturnError(sql.ErrNoRows)

		// Then, the Create call
		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-2", domain.RoleMember, domain.StatusApproved, &now, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		membership, err := repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, now)
		require.NoError(t, err)
		assert.Equal(t, "network-1", membership.NetworkID)
		assert.Equal(t, "user-2", membership.UserID)
		assert.Equal(t, domain.RoleMember, membership.Role)
		assert.Equal(t, domain.StatusApproved, membership.Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update fails", func(t *testing.T) {
		// First, the GetByNetworkAndUser call finds existing membership
		rows := sqlmock.NewRows([]string{
			"id", "network_id", "user_id", "role", "status", "joined_at", "updated_at",
		}).AddRow("membership-1", "network-1", "user-1", domain.RoleMember, domain.StatusPending, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(getQuery)).
			WithArgs("network-1", "user-1").
			WillReturnRows(rows)

		// Then, the Update call fails
		mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
			WithArgs(domain.RoleAdmin, domain.StatusApproved, sqlmock.AnyArg(), "membership-1").
			WillReturnError(errors.New("update error"))

		membership, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleAdmin, now)
		require.Error(t, err)
		assert.Nil(t, membership)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create fails", func(t *testing.T) {
		// First, the GetByNetworkAndUser call returns not found
		mock.ExpectQuery(regexp.QuoteMeta(getQuery)).
			WithArgs("network-1", "user-2").
			WillReturnError(sql.ErrNoRows)

		// Then, the Create call fails
		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-2", domain.RoleMember, domain.StatusApproved, &now, sqlmock.AnyArg()).
			WillReturnError(errors.New("create error"))

		membership, err := repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, now)
		require.Error(t, err)
		assert.Nil(t, membership)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
