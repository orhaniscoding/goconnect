package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresIdempotencyRepository implements IdempotencyRepository using PostgreSQL
type PostgresIdempotencyRepository struct {
	db *sql.DB
}

// NewPostgresIdempotencyRepository creates a new PostgreSQL-backed idempotency repository
func NewPostgresIdempotencyRepository(db *sql.DB) *PostgresIdempotencyRepository {
	return &PostgresIdempotencyRepository{db: db}
}

func (r *PostgresIdempotencyRepository) Get(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	query := `
		SELECT key, response_body, response_status, created_at
		FROM idempotency_keys
		WHERE key = $1 AND created_at > $2
	`
	// 24-hour TTL
	cutoff := time.Now().Add(-24 * time.Hour)

	record := &domain.IdempotencyRecord{}
	var responseBody string
	var responseStatus int

	err := r.db.QueryRowContext(ctx, query, key, cutoff).Scan(
		&record.Key,
		&responseBody,
		&responseStatus,
		&record.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Idempotency key not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency record: %w", err)
	}

	// Store response as bytes
	record.Response = []byte(responseBody)
	// Note: response_status from DB is not used in domain model currently
	_ = responseStatus

	return record, nil
}

func (r *PostgresIdempotencyRepository) Set(ctx context.Context, record *domain.IdempotencyRecord) error {
	query := `
		INSERT INTO idempotency_keys (key, response_body, response_status, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE
		SET response_body = EXCLUDED.response_body,
		    response_status = EXCLUDED.response_status,
		    created_at = EXCLUDED.created_at
	`
	// Convert response bytes to string for storage
	responseBody := string(record.Response)
	// Use a default status code (200) since domain model doesn't track it
	responseStatus := 200

	_, err := r.db.ExecContext(ctx, query,
		record.Key,
		responseBody,
		responseStatus,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to set idempotency record: %w", err)
	}

	return nil
}

func (r *PostgresIdempotencyRepository) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM idempotency_keys WHERE key = $1`
	_, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete idempotency record: %w", err)
	}
	return nil
}

func (r *PostgresIdempotencyRepository) Cleanup(ctx context.Context) error {
	// Delete records older than 24 hours
	query := `DELETE FROM idempotency_keys WHERE created_at < $1`
	cutoff := time.Now().Add(-24 * time.Hour)

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup idempotency records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected > 0 {
		// Log cleanup (could add metrics here)
		_ = rowsAffected
	}

	return nil
}
