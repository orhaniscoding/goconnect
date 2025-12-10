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

func TestNewPostgresIdempotencyRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIdempotencyRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresIdempotencyRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIdempotencyRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		key := "test-key-123"
		responseBody := `{"status": "ok"}`
		responseStatus := 200

		rows := sqlmock.NewRows([]string{"key", "response_body", "response_status", "created_at"}).
			AddRow(key, responseBody, responseStatus, now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT key, response_body, response_status, created_at FROM idempotency_keys WHERE key = $1 AND created_at > $2`)).
			WithArgs(key, sqlmock.AnyArg()).
			WillReturnRows(rows)

		record, err := repo.Get(ctx, key)
		require.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, key, record.Key)
		assert.Equal(t, []byte(responseBody), record.Response)
		assert.Equal(t, now.Unix(), record.CreatedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		key := "nonexistent-key"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT key, response_body, response_status, created_at FROM idempotency_keys WHERE key = $1 AND created_at > $2`)).
			WithArgs(key, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		record, err := repo.Get(ctx, key)
		require.Error(t, err)
		assert.Nil(t, record)

		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		key := "test-key"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT key, response_body, response_status, created_at FROM idempotency_keys WHERE key = $1 AND created_at > $2`)).
			WithArgs(key, sqlmock.AnyArg()).
			WillReturnError(errors.New("database connection error"))

		record, err := repo.Get(ctx, key)
		require.Error(t, err)
		assert.Nil(t, record)
		assert.Contains(t, err.Error(), "failed to get idempotency record")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIdempotencyRepository_Set(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIdempotencyRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		record := &domain.IdempotencyRecord{
			Key:       "test-key-456",
			Response:  []byte(`{"result": "success"}`),
			CreatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO idempotency_keys (key, response_body, response_status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (key) DO UPDATE SET response_body = EXCLUDED.response_body, response_status = EXCLUDED.response_status, created_at = EXCLUDED.created_at`)).
			WithArgs(record.Key, string(record.Response), 200, record.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Set(ctx, record)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		record := &domain.IdempotencyRecord{
			Key:       "test-key-789",
			Response:  []byte(`{"result": "error"}`),
			CreatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO idempotency_keys (key, response_body, response_status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (key) DO UPDATE SET response_body = EXCLUDED.response_body, response_status = EXCLUDED.response_status, created_at = EXCLUDED.created_at`)).
			WithArgs(record.Key, string(record.Response), 200, record.CreatedAt).
			WillReturnError(errors.New("insert failed"))

		err := repo.Set(ctx, record)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set idempotency record")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("upsert existing key", func(t *testing.T) {
		record := &domain.IdempotencyRecord{
			Key:       "existing-key",
			Response:  []byte(`{"updated": true}`),
			CreatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO idempotency_keys (key, response_body, response_status, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (key) DO UPDATE SET response_body = EXCLUDED.response_body, response_status = EXCLUDED.response_status, created_at = EXCLUDED.created_at`)).
			WithArgs(record.Key, string(record.Response), 200, record.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Set(ctx, record)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIdempotencyRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIdempotencyRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		key := "key-to-delete"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE key = $1`)).
			WithArgs(key).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, key)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("key not found still succeeds", func(t *testing.T) {
		key := "nonexistent-key"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE key = $1`)).
			WithArgs(key).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, key)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		key := "error-key"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE key = $1`)).
			WithArgs(key).
			WillReturnError(errors.New("delete failed"))

		err := repo.Delete(ctx, key)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete idempotency record")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIdempotencyRepository_Cleanup(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIdempotencyRepository(db)
	ctx := context.Background()

	t.Run("success with rows deleted", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE created_at < $1`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 5))

		err := repo.Cleanup(ctx)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no rows deleted", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE created_at < $1`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Cleanup(ctx)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM idempotency_keys WHERE created_at < $1`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("cleanup failed"))

		err := repo.Cleanup(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to cleanup idempotency records")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
