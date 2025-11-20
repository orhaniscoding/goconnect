package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresJoinRequestRepository implements JoinRequestRepository using PostgreSQL
type PostgresJoinRequestRepository struct {
	db *sql.DB
}

// NewPostgresJoinRequestRepository creates a new PostgreSQL-backed join request repository
func NewPostgresJoinRequestRepository(db *sql.DB) *PostgresJoinRequestRepository {
	return &PostgresJoinRequestRepository{db: db}
}

func (r *PostgresJoinRequestRepository) CreatePending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	// Check if request already exists
	existingQuery := `
		SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by
		FROM join_requests
		WHERE network_id = $1 AND user_id = $2
	`
	jr := &domain.JoinRequest{}
	var reviewedAt sql.NullTime
	var reviewedBy sql.NullString

	err := r.db.QueryRowContext(ctx, existingQuery, networkID, userID).Scan(
		&jr.ID,
		&jr.NetworkID,
		&jr.UserID,
		&jr.Status,
		&jr.CreatedAt,
		&reviewedAt,
		&reviewedBy,
	)

	if err == nil {
		// Request already exists
		if reviewedAt.Valid {
			jr.DecidedAt = &reviewedAt.Time
		}
		// Note: ReviewedBy is not in domain model, so we ignore it
		_ = reviewedBy
		return jr, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing join request: %w", err)
	}

	// Create new pending request
	jr = &domain.JoinRequest{
		ID:        domain.GenerateNetworkID(), // Using same ID generator
		NetworkID: networkID,
		UserID:    userID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO join_requests (id, network_id, user_id, status, requested_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.db.ExecContext(ctx, query,
		jr.ID,
		jr.NetworkID,
		jr.UserID,
		jr.Status,
		jr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}

	return jr, nil
}

func (r *PostgresJoinRequestRepository) GetPending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	query := `
		SELECT id, network_id, user_id, status, requested_at, reviewed_at, reviewed_by
		FROM join_requests
		WHERE network_id = $1 AND user_id = $2 AND status = 'pending'
	`
	jr := &domain.JoinRequest{}
	var reviewedAt sql.NullTime
	var reviewedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, networkID, userID).Scan(
		&jr.ID,
		&jr.NetworkID,
		&jr.UserID,
		&jr.Status,
		&jr.CreatedAt,
		&reviewedAt,
		&reviewedBy,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Join request not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get join request: %w", err)
	}

	if reviewedAt.Valid {
		jr.DecidedAt = &reviewedAt.Time
	}
	// Note: ReviewedBy is not in domain model, so we ignore it
	_ = reviewedBy

	return jr, nil
}

func (r *PostgresJoinRequestRepository) Decide(ctx context.Context, id string, approve bool) error {
	status := "denied"
	if approve {
		status = "approved"
	}

	query := `
		UPDATE join_requests
		SET status = $1, reviewed_at = $2
		WHERE id = $3 AND status = 'pending'
	`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to decide join request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.NewError(domain.ErrNotFound, "Join request not found or already reviewed", nil)
	}

	return nil
}
