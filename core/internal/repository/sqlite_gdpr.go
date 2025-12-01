package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteDeletionRequestRepository is a SQLite implementation
type SQLiteDeletionRequestRepository struct {
	db *sql.DB
}

// NewSQLiteDeletionRequestRepository creates a new SQLite repository
func NewSQLiteDeletionRequestRepository(db *sql.DB) *SQLiteDeletionRequestRepository {
	return &SQLiteDeletionRequestRepository{db: db}
}

func (r *SQLiteDeletionRequestRepository) Create(ctx context.Context, req *domain.DeletionRequest) error {
	query := `
		INSERT INTO deletion_requests (id, user_id, status, requested_at, completed_at, error)
		VALUES (?, ?, ?, ?, ?, ?)
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

func (r *SQLiteDeletionRequestRepository) Get(ctx context.Context, id string) (*domain.DeletionRequest, error) {
	query := `
		SELECT id, user_id, status, requested_at, completed_at, error
		FROM deletion_requests
		WHERE id = ?
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

func (r *SQLiteDeletionRequestRepository) GetByUserID(ctx context.Context, userID string) (*domain.DeletionRequest, error) {
	query := `
		SELECT id, user_id, status, requested_at, completed_at, error
		FROM deletion_requests
		WHERE user_id = ?
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

func (r *SQLiteDeletionRequestRepository) ListPending(ctx context.Context) ([]*domain.DeletionRequest, error) {
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

func (r *SQLiteDeletionRequestRepository) Update(ctx context.Context, req *domain.DeletionRequest) error {
	query := `
		UPDATE deletion_requests
		SET status = ?, completed_at = ?, error = ?
		WHERE id = ?
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
