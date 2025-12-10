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

func TestNewPostgresJoinRequestRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresJoinRequestRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresJoinRequestRepository_CreatePending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresJoinRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success creates new pending request", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"

		// First check for existing - not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2`)).
			WithArgs(networkID, userID).
			WillReturnError(sql.ErrNoRows)

		// Then insert new request
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO join_requests (id, network_id, user_id, status, requested_at) VALUES ($1, $2, $3, $4, $5)`)).
			WithArgs(sqlmock.AnyArg(), networkID, userID, "pending", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		jr, err := repo.CreatePending(ctx, networkID, userID)
		require.NoError(t, err)
		assert.NotNil(t, jr)
		assert.Equal(t, networkID, jr.NetworkID)
		assert.Equal(t, userID, jr.UserID)
		assert.Equal(t, "pending", jr.Status)
		assert.NotEmpty(t, jr.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns existing request without reviewed_at", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"
		existingID := "jr-existing"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at", "reviewed_by"}).
			AddRow(existingID, networkID, userID, "pending", now, nil, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2`)).
			WithArgs(networkID, userID).
			WillReturnRows(rows)

		jr, err := repo.CreatePending(ctx, networkID, userID)
		require.NoError(t, err)
		assert.NotNil(t, jr)
		assert.Equal(t, existingID, jr.ID)
		assert.Equal(t, "pending", jr.Status)
		assert.Nil(t, jr.DecidedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns existing request with reviewed_at", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"
		existingID := "jr-reviewed"
		reviewedAt := now.Add(-1 * time.Hour)
		reviewedBy := "admin-001"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at", "reviewed_by"}).
			AddRow(existingID, networkID, userID, "approved", now, reviewedAt, reviewedBy)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2`)).
			WithArgs(networkID, userID).
			WillReturnRows(rows)

		jr, err := repo.CreatePending(ctx, networkID, userID)
		require.NoError(t, err)
		assert.NotNil(t, jr)
		assert.Equal(t, existingID, jr.ID)
		assert.Equal(t, "approved", jr.Status)
		assert.NotNil(t, jr.DecidedAt)
		assert.Equal(t, reviewedAt.Unix(), jr.DecidedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("check existing query error", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2`)).
			WithArgs(networkID, userID).
			WillReturnError(errors.New("database connection error"))

		jr, err := repo.CreatePending(ctx, networkID, userID)
		require.Error(t, err)
		assert.Nil(t, jr)
		assert.Contains(t, err.Error(), "failed to check existing join request")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert error", func(t *testing.T) {
		networkID := "net-error"
		userID := "user-error"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2`)).
			WithArgs(networkID, userID).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO join_requests (id, network_id, user_id, status, requested_at) VALUES ($1, $2, $3, $4, $5)`)).
			WithArgs(sqlmock.AnyArg(), networkID, userID, "pending", sqlmock.AnyArg()).
			WillReturnError(errors.New("insert failed"))

		jr, err := repo.CreatePending(ctx, networkID, userID)
		require.Error(t, err)
		assert.Nil(t, jr)
		assert.Contains(t, err.Error(), "failed to create join request")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresJoinRequestRepository_GetPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresJoinRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without reviewed_at", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at", "reviewed_by"}).
			AddRow("jr-123", networkID, userID, "pending", now, nil, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2 AND status = 'pending'`)).
			WithArgs(networkID, userID).
			WillReturnRows(rows)

		jr, err := repo.GetPending(ctx, networkID, userID)
		require.NoError(t, err)
		assert.NotNil(t, jr)
		assert.Equal(t, "jr-123", jr.ID)
		assert.Equal(t, networkID, jr.NetworkID)
		assert.Equal(t, userID, jr.UserID)
		assert.Equal(t, "pending", jr.Status)
		assert.Nil(t, jr.DecidedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with reviewed_at", func(t *testing.T) {
		networkID := "net-123"
		userID := "user-456"
		reviewedAt := now.Add(-1 * time.Hour)
		reviewedBy := "admin-001"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at", "reviewed_by"}).
			AddRow("jr-reviewed", networkID, userID, "pending", now, reviewedAt, reviewedBy)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2 AND status = 'pending'`)).
			WithArgs(networkID, userID).
			WillReturnRows(rows)

		jr, err := repo.GetPending(ctx, networkID, userID)
		require.NoError(t, err)
		assert.NotNil(t, jr)
		assert.NotNil(t, jr.DecidedAt)
		assert.Equal(t, reviewedAt.Unix(), jr.DecidedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		networkID := "net-nonexistent"
		userID := "user-nonexistent"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2 AND status = 'pending'`)).
			WithArgs(networkID, userID).
			WillReturnError(sql.ErrNoRows)

		jr, err := repo.GetPending(ctx, networkID, userID)
		require.Error(t, err)
		assert.Nil(t, jr)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		networkID := "net-error"
		userID := "user-error"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by FROM join_requests WHERE network_id = $1 AND user_id = $2 AND status = 'pending'`)).
			WithArgs(networkID, userID).
			WillReturnError(errors.New("connection error"))

		jr, err := repo.GetPending(ctx, networkID, userID)
		require.Error(t, err)
		assert.Nil(t, jr)
		assert.Contains(t, err.Error(), "failed to get join request")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresJoinRequestRepository_Decide(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresJoinRequestRepository(db)
	ctx := context.Background()

	t.Run("success approve", func(t *testing.T) {
		id := "jr-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE join_requests SET status = $1, reviewed_at = $2 WHERE id = $3 AND status = 'pending'`)).
			WithArgs("approved", sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Decide(ctx, id, true)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success deny", func(t *testing.T) {
		id := "jr-456"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE join_requests SET status = $1, reviewed_at = $2 WHERE id = $3 AND status = 'pending'`)).
			WithArgs("denied", sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Decide(ctx, id, false)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or already reviewed", func(t *testing.T) {
		id := "jr-nonexistent"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE join_requests SET status = $1, reviewed_at = $2 WHERE id = $3 AND status = 'pending'`)).
			WithArgs("approved", sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Decide(ctx, id, true)
		require.Error(t, err)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		id := "jr-error"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE join_requests SET status = $1, reviewed_at = $2 WHERE id = $3 AND status = 'pending'`)).
			WithArgs("approved", sqlmock.AnyArg(), id).
			WillReturnError(errors.New("connection error"))

		err := repo.Decide(ctx, id, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decide join request")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		id := "jr-rows-err"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE join_requests SET status = $1, reviewed_at = $2 WHERE id = $3 AND status = 'pending'`)).
			WithArgs("denied", sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Decide(ctx, id, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresJoinRequestRepository_ListPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresJoinRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with multiple results", func(t *testing.T) {
		networkID := "net-123"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at"}).
			AddRow("jr-1", networkID, "user-001", "pending", now, nil).
			AddRow("jr-2", networkID, "user-002", "pending", now.Add(-1*time.Hour), nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx, networkID)
		require.NoError(t, err)
		assert.Len(t, requests, 2)
		assert.Equal(t, "jr-1", requests[0].ID)
		assert.Equal(t, "jr-2", requests[1].ID)
		assert.Equal(t, "pending", requests[0].Status)
		assert.Equal(t, "pending", requests[1].Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with reviewed_at populated", func(t *testing.T) {
		networkID := "net-123"
		reviewedAt := now.Add(-30 * time.Minute)

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at"}).
			AddRow("jr-1", networkID, "user-001", "pending", now, reviewedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx, networkID)
		require.NoError(t, err)
		assert.Len(t, requests, 1)
		assert.NotNil(t, requests[0].DecidedAt)
		assert.Equal(t, reviewedAt.Unix(), requests[0].DecidedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success empty result", func(t *testing.T) {
		networkID := "net-empty"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx, networkID)
		require.NoError(t, err)
		assert.Empty(t, requests)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		networkID := "net-error"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnError(errors.New("database error"))

		requests, err := repo.ListPending(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, requests)
		assert.Contains(t, err.Error(), "failed to list pending join requests")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		networkID := "net-scan-err"

		// Return wrong number of columns
		rows := sqlmock.NewRows([]string{"id", "network_id"}).
			AddRow("jr-1", networkID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, requests)
		assert.Contains(t, err.Error(), "failed to scan join request")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		networkID := "net-rows-err"

		rows := sqlmock.NewRows([]string{"id", "network_id", "user_id", "status", "requested_at", "reviewed_at"}).
			AddRow("jr-1", networkID, "user-001", "pending", now, nil).
			RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, network_id, user_id, status, requested_at, reviewed_at FROM join_requests WHERE network_id = $1 AND status = 'pending' ORDER BY requested_at ASC`)).
			WithArgs(networkID).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx, networkID)
		require.Error(t, err)
		assert.Nil(t, requests)
		assert.Contains(t, err.Error(), "error iterating join requests")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
