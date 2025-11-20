package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresMembershipRepository implements MembershipRepository using PostgreSQL
type PostgresMembershipRepository struct {
	db *sql.DB
}

// NewPostgresMembershipRepository creates a new PostgreSQL-backed membership repository
func NewPostgresMembershipRepository(db *sql.DB) *PostgresMembershipRepository {
	return &PostgresMembershipRepository{db: db}
}

func (r *PostgresMembershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	query := `
		INSERT INTO memberships (id, network_id, user_id, role, status, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		membership.ID,
		membership.NetworkID,
		membership.UserID,
		membership.Role,
		membership.Status,
		membership.JoinedAt,
		membership.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}
	return nil
}

func (r *PostgresMembershipRepository) GetByID(ctx context.Context, id string) (*domain.Membership, error) {
	query := `
		SELECT id, network_id, user_id, role, status, joined_at, updated_at
		FROM memberships
		WHERE id = $1
	`
	membership := &domain.Membership{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&membership.ID,
		&membership.NetworkID,
		&membership.UserID,
		&membership.Role,
		&membership.Status,
		&membership.JoinedAt,
		&membership.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: membership not found", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get membership by ID: %w", err)
	}
	return membership, nil
}

func (r *PostgresMembershipRepository) GetByNetworkAndUser(ctx context.Context, networkID, userID string) (*domain.Membership, error) {
	query := `
		SELECT id, network_id, user_id, role, status, joined_at, updated_at
		FROM memberships
		WHERE network_id = $1 AND user_id = $2
	`
	membership := &domain.Membership{}
	err := r.db.QueryRowContext(ctx, query, networkID, userID).Scan(
		&membership.ID,
		&membership.NetworkID,
		&membership.UserID,
		&membership.Role,
		&membership.Status,
		&membership.JoinedAt,
		&membership.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: membership not found", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}
	return membership, nil
}

func (r *PostgresMembershipRepository) ListByNetwork(ctx context.Context, networkID string) ([]*domain.Membership, error) {
	query := `
		SELECT id, network_id, user_id, role, status, joined_at, updated_at
		FROM memberships
		WHERE network_id = $1
		ORDER BY joined_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		membership := &domain.Membership{}
		err := rows.Scan(
			&membership.ID,
			&membership.NetworkID,
			&membership.UserID,
			&membership.Role,
			&membership.Status,
			&membership.JoinedAt,
			&membership.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, membership)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate memberships: %w", err)
	}

	return memberships, nil
}

func (r *PostgresMembershipRepository) Update(ctx context.Context, membership *domain.Membership) error {
	query := `
		UPDATE memberships
		SET role = $1, status = $2, updated_at = $3
		WHERE id = $4
	`
	membership.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		membership.Role,
		membership.Status,
		membership.UpdatedAt,
		membership.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: membership not found", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresMembershipRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM memberships WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: membership not found", domain.ErrNotFound)
	}
	return nil
}

// Interface implementation methods (matching MembershipRepository interface)

func (r *PostgresMembershipRepository) Get(ctx context.Context, networkID, userID string) (*domain.Membership, error) {
	return r.GetByNetworkAndUser(ctx, networkID, userID)
}

func (r *PostgresMembershipRepository) UpsertApproved(ctx context.Context, networkID, userID string, role domain.MembershipRole, joinedAt time.Time) (*domain.Membership, error) {
	// Check if membership exists
	existing, err := r.GetByNetworkAndUser(ctx, networkID, userID)
	if err == nil {
		// Update existing membership
		existing.Role = role
		existing.Status = domain.StatusApproved
		existing.JoinedAt = &joinedAt
		existing.UpdatedAt = time.Now()
		if err := r.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new membership
	membership := &domain.Membership{
		ID:        domain.GenerateNetworkID(),
		NetworkID: networkID,
		UserID:    userID,
		Role:      role,
		Status:    domain.StatusApproved,
		JoinedAt:  &joinedAt,
		UpdatedAt: time.Now(),
	}

	if err := r.Create(ctx, membership); err != nil {
		return nil, err
	}

	return membership, nil
}

func (r *PostgresMembershipRepository) SetStatus(ctx context.Context, networkID, userID string, status domain.MembershipStatus) error {
	query := `
		UPDATE memberships
		SET status = $1, updated_at = $2
		WHERE network_id = $3 AND user_id = $4
	`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to set membership status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Membership not found", nil)
	}

	return nil
}

func (r *PostgresMembershipRepository) List(ctx context.Context, networkID string, status string, limit int, cursor string) ([]*domain.Membership, string, error) {
	query := `
		SELECT id, network_id, user_id, role, status, joined_at, updated_at
		FROM memberships
		WHERE network_id = $1
	`
	args := []interface{}{networkID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if cursor != "" {
		query += fmt.Sprintf(" AND id > $%d", argIdx)
		args = append(args, cursor)
		argIdx++
	}

	query += " ORDER BY id ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit+1) // Fetch one extra to determine if there's a next page
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		membership := &domain.Membership{}
		var joinedAt sql.NullTime
		err := rows.Scan(
			&membership.ID,
			&membership.NetworkID,
			&membership.UserID,
			&membership.Role,
			&membership.Status,
			&joinedAt,
			&membership.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan membership: %w", err)
		}
		if joinedAt.Valid {
			membership.JoinedAt = &joinedAt.Time
		}
		memberships = append(memberships, membership)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate memberships: %w", err)
	}

	// Determine next cursor
	var nextCursor string
	if limit > 0 && len(memberships) > limit {
		nextCursor = memberships[limit-1].ID
		memberships = memberships[:limit]
	}

	return memberships, nextCursor, nil
}

func (r *PostgresMembershipRepository) Remove(ctx context.Context, networkID, userID string) error {
	query := `DELETE FROM memberships WHERE network_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove membership: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Membership not found", nil)
	}

	return nil
}
