package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteMembershipRepository implements MembershipRepository for SQLite.
type SQLiteMembershipRepository struct {
	db *sql.DB
}

// SQLiteJoinRequestRepository implements JoinRequestRepository for SQLite.
type SQLiteJoinRequestRepository struct {
	db *sql.DB
}

func NewSQLiteMembershipRepository(db *sql.DB) *SQLiteMembershipRepository {
	return &SQLiteMembershipRepository{db: db}
}

func NewSQLiteJoinRequestRepository(db *sql.DB) *SQLiteJoinRequestRepository {
	return &SQLiteJoinRequestRepository{db: db}
}

func (r *SQLiteMembershipRepository) UpsertApproved(ctx context.Context, networkID, userID string, role domain.MembershipRole, joinedAt time.Time) (*domain.Membership, error) {
	// Try update first
	queryUpdate := `
		UPDATE memberships
		SET role = ?, status = ?, joined_at = ?, updated_at = ?
		WHERE network_id = ? AND user_id = ?
	`
	now := time.Now()
	res, err := r.db.ExecContext(ctx, queryUpdate, role, domain.StatusApproved, joinedAt, now, networkID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update membership: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		// insert
		id := domain.GenerateNetworkID()
		insert := `
			INSERT INTO memberships (id, network_id, user_id, role, status, joined_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = r.db.ExecContext(ctx, insert, id, networkID, userID, role, domain.StatusApproved, joinedAt, now)
		if err != nil {
			return nil, fmt.Errorf("failed to insert membership: %w", err)
		}
		return &domain.Membership{
			ID:        id,
			NetworkID: networkID,
			UserID:    userID,
			Role:      role,
			Status:    domain.StatusApproved,
			JoinedAt:  &joinedAt,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}
	// fetch existing id for completeness
	var existingID string
	_ = r.db.QueryRowContext(ctx, `SELECT id FROM memberships WHERE network_id = ? AND user_id = ?`, networkID, userID).Scan(&existingID)
	return &domain.Membership{
		ID:        existingID,
		NetworkID: networkID,
		UserID:    userID,
		Role:      role,
		Status:    domain.StatusApproved,
		JoinedAt:  &joinedAt,
		UpdatedAt: now,
	}, nil
}

func (r *SQLiteMembershipRepository) Get(ctx context.Context, networkID, userID string) (*domain.Membership, error) {
	query := `
		SELECT id, role, status, joined_at, updated_at
		FROM memberships
		WHERE network_id = ? AND user_id = ?
	`
	var m domain.Membership
	var joined sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, networkID, userID).Scan(
		&m.ID,
		&m.Role,
		&m.Status,
		&joined,
		&m.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Membership not found", nil)
		}
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}
	m.NetworkID = networkID
	m.UserID = userID
	if joined.Valid {
		m.JoinedAt = &joined.Time
	}
	return &m, nil
}

func (r *SQLiteMembershipRepository) SetStatus(ctx context.Context, networkID, userID string, status domain.MembershipStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE memberships SET status = ?, updated_at = ? WHERE network_id = ? AND user_id = ?
	`, status, time.Now(), networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to set membership status: %w", err)
	}
	return nil
}

func (r *SQLiteMembershipRepository) List(ctx context.Context, networkID string, status string, limit int, cursor string) ([]*domain.Membership, string, error) {
	args := []interface{}{networkID}
	query := `
		SELECT id, network_id, user_id, role, status, joined_at, updated_at
		FROM memberships
		WHERE network_id = ?
	`
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if cursor != "" {
		query += " AND id > ?" // simplistic cursor by ID ordering
		args = append(args, cursor)
	}
	query += " ORDER BY id ASC"
	if limit <= 0 {
		limit = 20
	}
	query += " LIMIT ?"
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var out []*domain.Membership
	for rows.Next() {
		var m domain.Membership
		var joined sql.NullTime
		if err := rows.Scan(
			&m.ID,
			&m.NetworkID,
			&m.UserID,
			&m.Role,
			&m.Status,
			&joined,
			&m.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan membership: %w", err)
		}
		if joined.Valid {
			m.JoinedAt = &joined.Time
		}
		out = append(out, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate memberships: %w", err)
	}
	next := ""
	if len(out) > limit {
		next = out[limit].ID
		out = out[:limit]
	}
	return out, next, nil
}

func (r *SQLiteMembershipRepository) Remove(ctx context.Context, networkID, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM memberships WHERE network_id = ? AND user_id = ?`, networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove membership: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Membership not found", nil)
	}
	return nil
}

// Join requests

func (r *SQLiteJoinRequestRepository) CreatePending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	// Check existing pending
	if _, err := r.GetPending(ctx, networkID, userID); err == nil {
		return nil, domain.NewError(domain.ErrAlreadyRequested, "Join request already pending", nil)
	}
	id := domain.GenerateNetworkID()
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO join_requests (id, network_id, user_id, status, requested_at)
		VALUES (?, ?, ?, 'pending', ?)
	`, id, networkID, userID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create join request: %w", err)
	}
	return &domain.JoinRequest{
		ID:        id,
		NetworkID: networkID,
		UserID:    userID,
		Status:    "pending",
		CreatedAt: now,
	}, nil
}

func (r *SQLiteJoinRequestRepository) GetPending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	var jr domain.JoinRequest
	var decided sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status, requested_at, reviewed_at
		FROM join_requests
		WHERE network_id = ? AND user_id = ? AND status = 'pending'
		LIMIT 1
	`, networkID, userID).Scan(&jr.ID, &jr.Status, &jr.CreatedAt, &decided)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "No pending join request", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pending join request: %w", err)
	}
	jr.NetworkID = networkID
	jr.UserID = userID
	if decided.Valid {
		jr.DecidedAt = &decided.Time
	}
	return &jr, nil
}

func (r *SQLiteJoinRequestRepository) Decide(ctx context.Context, id string, approve bool) error {
	status := "denied"
	if approve {
		status = "approved"
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE join_requests
		SET status = ?, reviewed_at = ?
		WHERE id = ? AND status = 'pending'
	`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to decide join request: %w", err)
	}
	return nil
}

func (r *SQLiteJoinRequestRepository) ListPending(ctx context.Context, networkID string) ([]*domain.JoinRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, requested_at
		FROM join_requests
		WHERE network_id = ? AND status = 'pending'
		ORDER BY requested_at ASC
	`, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending join requests: %w", err)
	}
	defer rows.Close()

	var out []*domain.JoinRequest
	for rows.Next() {
		var jr domain.JoinRequest
		if err := rows.Scan(&jr.ID, &jr.UserID, &jr.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan join request: %w", err)
		}
		jr.NetworkID = networkID
		jr.Status = "pending"
		out = append(out, &jr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate join requests: %w", err)
	}
	return out, nil
}
