package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresTenantMemberRepository implements TenantMemberRepository using PostgreSQL
type PostgresTenantMemberRepository struct {
	db *sql.DB
}

// NewPostgresTenantMemberRepository creates a new PostgreSQL-backed repository
func NewPostgresTenantMemberRepository(db *sql.DB) *PostgresTenantMemberRepository {
	return &PostgresTenantMemberRepository{db: db}
}

func (r *PostgresTenantMemberRepository) Create(ctx context.Context, member *domain.TenantMember) error {
	if member.ID == "" {
		member.ID = domain.GenerateTenantMemberID()
	}

	query := `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	member.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		member.ID,
		member.TenantID,
		member.UserID,
		member.Role,
		nullString(member.Nickname),
		member.JoinedAt,
		member.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(err) {
			return domain.NewError(domain.ErrValidation, "User is already a member of this tenant", nil)
		}
		return fmt.Errorf("failed to create tenant member: %w", err)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) GetByID(ctx context.Context, id string) (*domain.TenantMember, error) {
	query := `
		SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE id = $1
	`
	member := &domain.TenantMember{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&member.ID,
		&member.TenantID,
		&member.UserID,
		&member.Role,
		&member.Nickname,
		&member.JoinedAt,
		&member.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant member: %w", err)
	}
	return member, nil
}

func (r *PostgresTenantMemberRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.TenantMember, error) {
	query := `
		SELECT id, tenant_id, user_id, role, COALESCE(nickname, ''), joined_at, updated_at
		FROM tenant_members
		WHERE user_id = $1 AND tenant_id = $2
	`
	member := &domain.TenantMember{}
	err := r.db.QueryRowContext(ctx, query, userID, tenantID).Scan(
		&member.ID,
		&member.TenantID,
		&member.UserID,
		&member.Role,
		&member.Nickname,
		&member.JoinedAt,
		&member.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Tenant membership not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant membership: %w", err)
	}
	return member, nil
}

func (r *PostgresTenantMemberRepository) Update(ctx context.Context, member *domain.TenantMember) error {
	query := `
		UPDATE tenant_members
		SET role = $2, nickname = $3, updated_at = $4
		WHERE id = $1
	`
	member.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		member.ID,
		member.Role,
		nullString(member.Nickname),
		member.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenant_members WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) ListByTenant(ctx context.Context, tenantID string, role string, limit int, cursor string) ([]*domain.TenantMember, string, error) {
	// Build query with optional filters
	query := `
		SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at,
		       u.email, u.locale
		FROM tenant_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if role != "" {
		query += fmt.Sprintf(" AND tm.role = $%d", argIdx)
		args = append(args, role)
		argIdx++
	}

	if cursor != "" {
		query += fmt.Sprintf(" AND tm.id > $%d", argIdx)
		args = append(args, cursor)
		// argIdx++ not needed as no more parameters after this
	}

	query += " ORDER BY tm.joined_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit+1) // +1 to check for more
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tenant members: %w", err)
	}
	defer rows.Close()

	var members []*domain.TenantMember
	for rows.Next() {
		member := &domain.TenantMember{
			User: &domain.User{},
		}
		err := rows.Scan(
			&member.ID,
			&member.TenantID,
			&member.UserID,
			&member.Role,
			&member.Nickname,
			&member.JoinedAt,
			&member.UpdatedAt,
			&member.User.Email,
			&member.User.Locale,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan tenant member: %w", err)
		}
		member.User.ID = member.UserID
		members = append(members, member)
	}

	// Determine next cursor
	nextCursor := ""
	if limit > 0 && len(members) > limit {
		nextCursor = members[limit-1].ID
		members = members[:limit]
	}

	return members, nextCursor, nil
}

func (r *PostgresTenantMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.TenantMember, error) {
	query := `
		SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, COALESCE(tm.nickname, ''), tm.joined_at, tm.updated_at
		FROM tenant_members tm
		WHERE tm.user_id = $1
		ORDER BY tm.joined_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user's tenant memberships: %w", err)
	}
	defer rows.Close()

	var members []*domain.TenantMember
	for rows.Next() {
		member := &domain.TenantMember{}
		err := rows.Scan(
			&member.ID,
			&member.TenantID,
			&member.UserID,
			&member.Role,
			&member.Nickname,
			&member.JoinedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		members = append(members, member)
	}
	return members, nil
}

func (r *PostgresTenantMemberRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM tenant_members WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant members: %w", err)
	}
	return count, nil
}

func (r *PostgresTenantMemberRepository) UpdateRole(ctx context.Context, id string, role domain.TenantRole) error {
	query := `UPDATE tenant_members SET role = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, role, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) GetUserRole(ctx context.Context, userID, tenantID string) (domain.TenantRole, error) {
	query := `SELECT role FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`
	var role domain.TenantRole
	err := r.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", domain.NewError(domain.ErrNotFound, "User is not a member of this tenant", nil)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return role, nil
}

func (r *PostgresTenantMemberRepository) HasRole(ctx context.Context, userID, tenantID string, requiredRole domain.TenantRole) (bool, error) {
	role, err := r.GetUserRole(ctx, userID, tenantID)
	if err != nil {
		// If not found, user doesn't have the role
		if domainErr, ok := err.(*domain.Error); ok && domainErr.Code == domain.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return role.HasPermission(requiredRole), nil
}

func (r *PostgresTenantMemberRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenant_members WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant members: %w", err)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) Ban(ctx context.Context, id string, bannedBy string) error {
	query := `UPDATE tenant_members SET banned_at = $2, banned_by = $3, updated_at = $2 WHERE id = $1`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now, bannedBy)
	if err != nil {
		return fmt.Errorf("failed to ban member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) Unban(ctx context.Context, id string) error {
	query := `UPDATE tenant_members SET banned_at = NULL, banned_by = '', updated_at = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to unban member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *PostgresTenantMemberRepository) IsBanned(ctx context.Context, userID, tenantID string) (bool, error) {
	query := `SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = $1 AND tenant_id = $2`
	var isBanned bool
	err := r.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&isBanned)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check ban status: %w", err)
	}
	return isBanned, nil
}

func (r *PostgresTenantMemberRepository) ListBannedByTenant(ctx context.Context, tenantID string) ([]*domain.TenantMember, error) {
	query := `SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by 
              FROM tenant_members WHERE tenant_id = $1 AND banned_at IS NOT NULL ORDER BY banned_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list banned members: %w", err)
	}
	defer rows.Close()

	var members []*domain.TenantMember
	for rows.Next() {
		m := &domain.TenantMember{}
		var nickname sql.NullString
		var bannedAt sql.NullTime
		var bannedBy sql.NullString
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.Role, &nickname, &m.JoinedAt, &m.UpdatedAt, &bannedAt, &bannedBy); err != nil {
			return nil, fmt.Errorf("failed to scan banned member: %w", err)
		}
		if nickname.Valid {
			m.Nickname = nickname.String
		}
		if bannedAt.Valid {
			m.BannedAt = &bannedAt.Time
		}
		if bannedBy.Valid {
			m.BannedBy = bannedBy.String
		}
		members = append(members, m)
	}
	return members, nil
}

// Helper function for nullable strings
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// isDuplicateKeyError checks if the error is a unique constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 23505 is unique_violation
	return contains(err.Error(), "23505") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
