package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// AdminRepository handles database operations for admin functionality
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ListAllUsers retrieves all users with filtering and pagination
func (r *AdminRepository) ListAllUsers(ctx context.Context, filters domain.UserFilters, pagination domain.PaginationParams) ([]*domain.UserListItem, int, error) {
	// Build query with filters
	query := `SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen 
			  FROM users WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.Role == "admin" {
		query += fmt.Sprintf(" AND is_admin = $%d", argCount)
		countQuery += fmt.Sprintf(" AND is_admin = $%d", argCount)
		args = append(args, true)
		argCount++
	} else if filters.Role == "moderator" {
		query += fmt.Sprintf(" AND is_moderator = $%d", argCount)
		countQuery += fmt.Sprintf(" AND is_moderator = $%d", argCount)
		args = append(args, true)
		argCount++
	} else if filters.Role == "user" {
		query += fmt.Sprintf(" AND is_admin = $%d AND is_moderator = $%d", argCount, argCount+1)
		countQuery += fmt.Sprintf(" AND is_admin = $%d AND is_moderator = $%d", argCount, argCount+1)
		args = append(args, false, false)
		argCount += 2
	}

	if filters.Status == "active" {
		query += fmt.Sprintf(" AND suspended = $%d", argCount)
		countQuery += fmt.Sprintf(" AND suspended = $%d", argCount)
		args = append(args, false)
		argCount++
	} else if filters.Status == "suspended" {
		query += fmt.Sprintf(" AND suspended = $%d", argCount)
		countQuery += fmt.Sprintf(" AND suspended = $%d", argCount)
		args = append(args, true)
		argCount++
	}

	if filters.TenantID != "" {
		query += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		args = append(args, filters.TenantID)
		argCount++
	}

	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query += fmt.Sprintf(" AND (LOWER(email) LIKE $%d OR LOWER(username) LIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(email) LIKE $%d OR LOWER(username) LIKE $%d)", argCount, argCount)
		args = append(args, searchPattern)
		argCount++
	}

	// Get total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Add pagination and ordering
	query += " ORDER BY created_at DESC"
	if pagination.PerPage > 0 {
		offset := (pagination.Page - 1) * pagination.PerPage
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, pagination.PerPage, offset)
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*domain.UserListItem{}
	for rows.Next() {
		var user domain.UserListItem
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.TenantID,
			&user.IsAdmin,
			&user.IsModerator,
			&user.Suspended,
			&user.CreatedAt,
			&user.LastSeen,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}

// GetUserStats retrieves system-wide user statistics
func (r *AdminRepository) GetUserStats(ctx context.Context) (*domain.SystemStats, error) {
	stats := &domain.SystemStats{}

	// Count users
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Count admin users
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_admin = true").Scan(&stats.AdminUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	// Count moderator users
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_moderator = true").Scan(&stats.ModeratorUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count moderator users: %w", err)
	}

	// Count suspended users
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE suspended = true").Scan(&stats.SuspendedUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count suspended users: %w", err)
	}

	// Count tenants (if table exists)
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tenants").Scan(&stats.TotalTenants)

	// Count networks
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM networks").Scan(&stats.TotalNetworks)

	// Count devices
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&stats.TotalDevices)

	// Count active peers (last_seen within 5 minutes)
	_ = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM peers WHERE last_seen > NOW() - INTERVAL '5 minutes'").Scan(&stats.ActivePeers)

	return stats, nil
}

// UpdateUserRole updates a user's admin or moderator status
func (r *AdminRepository) UpdateUserRole(ctx context.Context, userID string, isAdmin, isModerator *bool) error {
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if isAdmin != nil {
		updates = append(updates, fmt.Sprintf("is_admin = $%d", argCount))
		args = append(args, *isAdmin)
		argCount++
	}

	if isModerator != nil {
		updates = append(updates, fmt.Sprintf("is_moderator = $%d", argCount))
		args = append(args, *isModerator)
		argCount++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no role updates provided")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, "NOW()")
	argCount++

	// Add userID as last parameter
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), argCount)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// SuspendUser suspends a user account
func (r *AdminRepository) SuspendUser(ctx context.Context, userID, reason, suspendedBy string) error {
	query := `UPDATE users 
			  SET suspended = true, 
			      suspended_at = NOW(), 
			      suspended_reason = $1, 
			      suspended_by = $2,
			      updated_at = NOW()
			  WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, reason, suspendedBy, userID)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UnsuspendUser unsuspends a user account
func (r *AdminRepository) UnsuspendUser(ctx context.Context, userID string) error {
	query := `UPDATE users 
			  SET suspended = false, 
			      suspended_at = NULL, 
			      suspended_reason = NULL, 
			      suspended_by = NULL,
			      updated_at = NOW()
			  WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unsuspend user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetUserByID retrieves a single user by ID (for detail view)
func (r *AdminRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT id, tenant_id, email, username, full_name, bio, avatar_url, locale, 
			         is_admin, is_moderator, two_fa_enabled, auth_provider, external_id,
			         suspended, suspended_at, suspended_reason, suspended_by,
			         created_at, updated_at, last_seen
			  FROM users WHERE id = $1`

	var user domain.User
	var suspendedAt sql.NullTime
	var suspendedReason, suspendedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.Bio,
		&user.AvatarURL,
		&user.Locale,
		&user.IsAdmin,
		&user.IsModerator,
		&user.TwoFAEnabled,
		&user.AuthProvider,
		&user.ExternalID,
		&suspendedAt,
		&suspendedReason,
		&suspendedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateLastSeen updates the last_seen timestamp for a user
func (r *AdminRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_seen = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}

	return nil
}
