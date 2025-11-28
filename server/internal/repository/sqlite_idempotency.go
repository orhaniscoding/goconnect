package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SQLiteIdempotencyRepository implements IdempotencyRepository using SQLite.
type SQLiteIdempotencyRepository struct {
	db *sql.DB
}

func NewSQLiteIdempotencyRepository(db *sql.DB) *SQLiteIdempotencyRepository {
	return &SQLiteIdempotencyRepository{db: db}
}

func (r *SQLiteIdempotencyRepository) Save(ctx context.Context, key, responseBody string, status int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO idempotency_keys (key, response_body, response_status, created_at)
		VALUES (?, ?, ?, ?)
	`, key, responseBody, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save idempotency key: %w", err)
	}
	return nil
}

func (r *SQLiteIdempotencyRepository) Get(ctx context.Context, key string) (string, int, bool, error) {
	var body string
	var status int
	err := r.db.QueryRowContext(ctx, `
		SELECT response_body, response_status FROM idempotency_keys WHERE key = ?
	`, key).Scan(&body, &status)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	}
	if err != nil {
		return "", 0, false, fmt.Errorf("failed to get idempotency key: %w", err)
	}
	return body, status, true, nil
}

// Cleanup removes entries older than 24h (matching in-memory behavior).
func (r *SQLiteIdempotencyRepository) Cleanup(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM idempotency_keys WHERE created_at < datetime('now', '-24 hours')
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup idempotency keys: %w", err)
	}
	return nil
}
