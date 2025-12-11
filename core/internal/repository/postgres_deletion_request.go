package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresDeletionRequestRepository is a PostgreSQL implementation
type PostgresDeletionRequestRepository struct {
	db *sql.DB
}

// NewPostgresDeletionRequestRepository creates a new PostgreSQL repository
func NewPostgresDeletionRequestRepository(db *sql.DB) *PostgresDeletionRequestRepository {
	return &PostgresDeletionRequestRepository{db: db}
}

func (r *PostgresDeletionRequestRepository) Create(ctx context.Context, req *domain.DeletionRequest) error {
	query := `
		INSERT INTO deletion_requests (id, user_id, status, requested_at, completed_at, error)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.UserID,
		req.Status,
		req.RequestedAt,
		req.CompletedAt,
		req.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to create deletion request: %w", err)
	}
	return nil
}

func (r *PostgresDeletionRequestRepository) Get(ctx context.Context, id string) (*domain.DeletionRequest, error) {
	query := `
		SELECT id, user_id, status, requested_at, completed_at, error
		FROM deletion_requests
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var req domain.DeletionRequest
	var completedAt sql.NullTime
	var errStr sql.NullString

	if err := row.Scan(
		&req.ID,
		&req.UserID,
		&req.Status,
		&req.RequestedAt,
		&completedAt,
		&errStr,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("request not found")
		}
		return nil, fmt.Errorf("failed to get deletion request: %w", err)
	}

	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}
	if errStr.Valid {
		req.Error = errStr.String
	}

	return &req, nil
}

func (r *PostgresDeletionRequestRepository) GetByUserID(ctx context.Context, userID string) (*domain.DeletionRequest, error) {
	query := `
		SELECT id, user_id, status, requested_at, completed_at, error
		FROM deletion_requests
		WHERE user_id = $1
		ORDER BY requested_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, userID)

	var req domain.DeletionRequest
	var completedAt sql.NullTime
	var errStr sql.NullString

	if err := row.Scan(
		&req.ID,
		&req.UserID,
		&req.Status,
		&req.RequestedAt,
		&completedAt,
		&errStr,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deletion request: %w", err)
	}

	if completedAt.Valid {
		req.CompletedAt = &completedAt.Time
	}
	if errStr.Valid {
		req.Error = errStr.String
	}

	return &req, nil
}

func (r *PostgresDeletionRequestRepository) ListPending(ctx context.Context) ([]*domain.DeletionRequest, error) {
	query := `
		SELECT id, user_id, status, requested_at, completed_at, error
		FROM deletion_requests
		WHERE status = 'pending'
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.DeletionRequest
	for rows.Next() {
		var req domain.DeletionRequest
		var completedAt sql.NullTime
		var errStr sql.NullString

		if err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Status,
			&req.RequestedAt,
			&completedAt,
			&errStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		if completedAt.Valid {
			req.CompletedAt = &completedAt.Time
		}
		if errStr.Valid {
			req.Error = errStr.String
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

func (r *PostgresDeletionRequestRepository) Update(ctx context.Context, req *domain.DeletionRequest) error {
	query := `
		UPDATE deletion_requests
		SET status = $1, completed_at = $2, error = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		req.Status,
		req.CompletedAt,
		req.Error,
		req.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deletion request: %w", err)
	}
	return nil
}
