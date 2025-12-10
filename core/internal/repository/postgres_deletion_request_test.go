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

func TestNewPostgresDeletionRequestRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresDeletionRequestRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		req := &domain.DeletionRequest{
			ID:          "req-123",
			UserID:      "user-456",
			Status:      domain.DeletionRequestStatusPending,
			RequestedAt: now,
			CompletedAt: nil,
			Error:       "",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO deletion_requests (id, user_id, status, requested_at, completed_at, error) VALUES ($1, $2, $3, $4, $5, $6)`)).
			WithArgs(req.ID, req.UserID, req.Status, req.RequestedAt, req.CompletedAt, req.Error).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, req)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with completed_at and error", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		req := &domain.DeletionRequest{
			ID:          "req-789",
			UserID:      "user-101",
			Status:      domain.DeletionRequestStatusFailed,
			RequestedAt: now,
			CompletedAt: &completedAt,
			Error:       "some error occurred",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO deletion_requests (id, user_id, status, requested_at, completed_at, error) VALUES ($1, $2, $3, $4, $5, $6)`)).
			WithArgs(req.ID, req.UserID, req.Status, req.RequestedAt, req.CompletedAt, req.Error).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, req)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		req := &domain.DeletionRequest{
			ID:          "req-error",
			UserID:      "user-error",
			Status:      domain.DeletionRequestStatusPending,
			RequestedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO deletion_requests (id, user_id, status, requested_at, completed_at, error) VALUES ($1, $2, $3, $4, $5, $6)`)).
			WithArgs(req.ID, req.UserID, req.Status, req.RequestedAt, req.CompletedAt, req.Error).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create deletion request")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeletionRequestRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without completed_at", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-123", "user-456", domain.DeletionRequestStatusPending, now, nil, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE id = $1`)).
			WithArgs("req-123").
			WillReturnRows(rows)

		req, err := repo.Get(ctx, "req-123")
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "req-123", req.ID)
		assert.Equal(t, "user-456", req.UserID)
		assert.Equal(t, domain.DeletionRequestStatusPending, req.Status)
		assert.Equal(t, now.Unix(), req.RequestedAt.Unix())
		assert.Nil(t, req.CompletedAt)
		assert.Empty(t, req.Error)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with completed_at and error", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-789", "user-101", domain.DeletionRequestStatusFailed, now, completedAt, "some error")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE id = $1`)).
			WithArgs("req-789").
			WillReturnRows(rows)

		req, err := repo.Get(ctx, "req-789")
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "req-789", req.ID)
		assert.Equal(t, "user-101", req.UserID)
		assert.Equal(t, domain.DeletionRequestStatusFailed, req.Status)
		require.NotNil(t, req.CompletedAt)
		assert.Equal(t, completedAt.Unix(), req.CompletedAt.Unix())
		assert.Equal(t, "some error", req.Error)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE id = $1`)).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		req, err := repo.Get(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request not found")
		assert.Nil(t, req)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE id = $1`)).
			WithArgs("req-error").
			WillReturnError(errors.New("database error"))

		req, err := repo.Get(ctx, "req-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get deletion request")
		assert.Nil(t, req)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeletionRequestRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-123", "user-456", domain.DeletionRequestStatusPending, now, nil, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE user_id = $1 ORDER BY requested_at DESC LIMIT 1`)).
			WithArgs("user-456").
			WillReturnRows(rows)

		req, err := repo.GetByUserID(ctx, "user-456")
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "req-123", req.ID)
		assert.Equal(t, "user-456", req.UserID)
		assert.Equal(t, domain.DeletionRequestStatusPending, req.Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with completed_at and error", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-789", "user-101", domain.DeletionRequestStatusCompleted, now, completedAt, "")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE user_id = $1 ORDER BY requested_at DESC LIMIT 1`)).
			WithArgs("user-101").
			WillReturnRows(rows)

		req, err := repo.GetByUserID(ctx, "user-101")
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "req-789", req.ID)
		assert.Equal(t, domain.DeletionRequestStatusCompleted, req.Status)
		require.NotNil(t, req.CompletedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found returns nil without error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE user_id = $1 ORDER BY requested_at DESC LIMIT 1`)).
			WithArgs("non-existent-user").
			WillReturnError(sql.ErrNoRows)

		req, err := repo.GetByUserID(ctx, "non-existent-user")
		require.NoError(t, err)
		assert.Nil(t, req)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE user_id = $1 ORDER BY requested_at DESC LIMIT 1`)).
			WithArgs("user-error").
			WillReturnError(errors.New("database error"))

		req, err := repo.GetByUserID(ctx, "user-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get deletion request")
		assert.Nil(t, req)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeletionRequestRepository_ListPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with multiple results", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-1", "user-1", domain.DeletionRequestStatusPending, now, nil, nil).
			AddRow("req-2", "user-2", domain.DeletionRequestStatusPending, now.Add(-time.Hour), nil, nil).
			AddRow("req-3", "user-3", domain.DeletionRequestStatusPending, now.Add(-2*time.Hour), nil, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE status = 'pending'`)).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx)
		require.NoError(t, err)
		require.Len(t, requests, 3)
		assert.Equal(t, "req-1", requests[0].ID)
		assert.Equal(t, "req-2", requests[1].ID)
		assert.Equal(t, "req-3", requests[2].ID)
		for _, req := range requests {
			assert.Equal(t, domain.DeletionRequestStatusPending, req.Status)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty results", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE status = 'pending'`)).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx)
		require.NoError(t, err)
		assert.Empty(t, requests)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with completed_at and error fields", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "requested_at", "completed_at", "error"}).
			AddRow("req-1", "user-1", domain.DeletionRequestStatusPending, now, completedAt, "partial error")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE status = 'pending'`)).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx)
		require.NoError(t, err)
		require.Len(t, requests, 1)
		require.NotNil(t, requests[0].CompletedAt)
		assert.Equal(t, "partial error", requests[0].Error)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE status = 'pending'`)).
			WillReturnError(errors.New("database error"))

		requests, err := repo.ListPending(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list pending requests")
		assert.Nil(t, requests)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow("req-1", "user-1")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, status, requested_at, completed_at, error FROM deletion_requests WHERE status = 'pending'`)).
			WillReturnRows(rows)

		requests, err := repo.ListPending(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan request")
		assert.Nil(t, requests)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeletionRequestRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeletionRequestRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success update to completed", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		req := &domain.DeletionRequest{
			ID:          "req-123",
			UserID:      "user-456",
			Status:      domain.DeletionRequestStatusCompleted,
			RequestedAt: now,
			CompletedAt: &completedAt,
			Error:       "",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE deletion_requests SET status = $1, completed_at = $2, error = $3 WHERE id = $4`)).
			WithArgs(req.Status, req.CompletedAt, req.Error, req.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, req)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success update to failed with error message", func(t *testing.T) {
		completedAt := now.Add(time.Hour)
		req := &domain.DeletionRequest{
			ID:          "req-789",
			UserID:      "user-101",
			Status:      domain.DeletionRequestStatusFailed,
			RequestedAt: now,
			CompletedAt: &completedAt,
			Error:       "deletion failed due to database constraint",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE deletion_requests SET status = $1, completed_at = $2, error = $3 WHERE id = $4`)).
			WithArgs(req.Status, req.CompletedAt, req.Error, req.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, req)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success update to processing", func(t *testing.T) {
		req := &domain.DeletionRequest{
			ID:          "req-111",
			UserID:      "user-222",
			Status:      domain.DeletionRequestStatusProcessing,
			RequestedAt: now,
			CompletedAt: nil,
			Error:       "",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE deletion_requests SET status = $1, completed_at = $2, error = $3 WHERE id = $4`)).
			WithArgs(req.Status, req.CompletedAt, req.Error, req.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, req)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		req := &domain.DeletionRequest{
			ID:          "req-error",
			UserID:      "user-error",
			Status:      domain.DeletionRequestStatusCompleted,
			RequestedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE deletion_requests SET status = $1, completed_at = $2, error = $3 WHERE id = $4`)).
			WithArgs(req.Status, req.CompletedAt, req.Error, req.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update deletion request")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
