package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES FRIENDSHIP REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresFriendshipRepository implements FriendshipRepository for PostgreSQL
type PostgresFriendshipRepository struct {
	db *sql.DB
}

// NewPostgresFriendshipRepository creates a new PostgresFriendshipRepository
func NewPostgresFriendshipRepository(db *sql.DB) *PostgresFriendshipRepository {
	return &PostgresFriendshipRepository{db: db}
}

// Create creates a new friendship (pending status)
func (r *PostgresFriendshipRepository) Create(ctx context.Context, friendship *domain.Friendship) error {
	query := `
		INSERT INTO friendships (id, user_id, friend_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	if friendship.ID == "" {
		friendship.ID = domain.GenerateFriendshipID()
	}
	friendship.CreatedAt = now
	friendship.Status = domain.FriendshipStatusPending

	_, err := r.db.ExecContext(ctx, query,
		friendship.ID,
		friendship.UserID,
		friendship.FriendID,
		friendship.Status,
		friendship.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create friendship: %w", err)
	}

	return nil
}

// GetByID retrieves a friendship by ID
func (r *PostgresFriendshipRepository) GetByID(ctx context.Context, id string) (*domain.Friendship, error) {
	query := `
		SELECT id, user_id, friend_id, status, created_at, accepted_at
		FROM friendships
		WHERE id = $1
	`

	var friendship domain.Friendship
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&friendship.ID,
		&friendship.UserID,
		&friendship.FriendID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.AcceptedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get friendship: %w", err)
	}

	return &friendship, nil
}

// GetByUsers retrieves friendship between two users (if exists)
func (r *PostgresFriendshipRepository) GetByUsers(ctx context.Context, userID, friendID string) (*domain.Friendship, error) {
	query := `
		SELECT id, user_id, friend_id, status, created_at, accepted_at
		FROM friendships
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`

	var friendship domain.Friendship
	err := r.db.QueryRowContext(ctx, query, userID, friendID).Scan(
		&friendship.ID,
		&friendship.UserID,
		&friendship.FriendID,
		&friendship.Status,
		&friendship.CreatedAt,
		&friendship.AcceptedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get friendship: %w", err)
	}

	return &friendship, nil
}

// List retrieves friendships with filters
func (r *PostgresFriendshipRepository) List(ctx context.Context, filter FriendshipFilter) ([]domain.Friendship, string, error) {
	query := `
		SELECT id, user_id, friend_id, status, created_at, accepted_at
		FROM friendships
		WHERE (user_id = $1 OR friend_id = $1)
	`

	args := []interface{}{filter.UserID}
	argCount := 1

	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	if filter.Cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND id < $%d", argCount)
		args = append(args, filter.Cursor)
	}

	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list friendships: %w", err)
	}
	defer rows.Close()

	var friendships []domain.Friendship
	for rows.Next() {
		var friendship domain.Friendship
		if err := rows.Scan(
			&friendship.ID,
			&friendship.UserID,
			&friendship.FriendID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.AcceptedAt,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan friendship: %w", err)
		}
		friendships = append(friendships, friendship)
	}

	var nextCursor string
	if len(friendships) > limit {
		nextCursor = friendships[limit-1].ID
		friendships = friendships[:limit]
	}

	return friendships, nextCursor, nil
}

// Accept accepts a pending friendship
func (r *PostgresFriendshipRepository) Accept(ctx context.Context, id string) error {
	query := `
		UPDATE friendships
		SET status = $2, accepted_at = $3
		WHERE id = $1 AND status = 'pending'
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, domain.FriendshipStatusAccepted, now)
	if err != nil {
		return fmt.Errorf("failed to accept friendship: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("friendship not found or already accepted")
	}

	return nil
}

// Block blocks a user (creates or updates friendship to blocked)
func (r *PostgresFriendshipRepository) Block(ctx context.Context, userID, blockedID string) error {
	query := `
		INSERT INTO friendships (id, user_id, friend_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, friend_id) DO UPDATE SET status = 'blocked'
	`

	_, err := r.db.ExecContext(ctx, query,
		domain.GenerateFriendshipID(),
		userID,
		blockedID,
		domain.FriendshipStatusBlocked,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	return nil
}

// Delete removes a friendship
func (r *PostgresFriendshipRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM friendships WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete friendship: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("friendship not found")
	}

	return nil
}

// IsBlocked checks if userA has blocked userB
func (r *PostgresFriendshipRepository) IsBlocked(ctx context.Context, userA, userB string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM friendships
			WHERE user_id = $1 AND friend_id = $2 AND status = 'blocked'
		)
	`

	var blocked bool
	err := r.db.QueryRowContext(ctx, query, userA, userB).Scan(&blocked)
	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}

	return blocked, nil
}

// AreFriends checks if two users are friends
func (r *PostgresFriendshipRepository) AreFriends(ctx context.Context, userA, userB string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM friendships
			WHERE ((user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1))
			AND status = 'accepted'
		)
	`

	var friends bool
	err := r.db.QueryRowContext(ctx, query, userA, userB).Scan(&friends)
	if err != nil {
		return false, fmt.Errorf("failed to check friendship: %w", err)
	}

	return friends, nil
}

// Ensure interface compliance
var _ FriendshipRepository = (*PostgresFriendshipRepository)(nil)
