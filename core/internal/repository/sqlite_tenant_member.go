package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteTenantMemberRepository implements TenantMemberRepository using SQLite.
type SQLiteTenantMemberRepository struct {
	db *sql.DB
}

func NewSQLiteTenantMemberRepository(db *sql.DB) *SQLiteTenantMemberRepository {
	return &SQLiteTenantMemberRepository{db: db}
}

func (r *SQLiteTenantMemberRepository) Create(ctx context.Context, member *domain.TenantMember) error {
	if member.ID == "" {
		member.ID = domain.GenerateTenantMemberID()
	}
	now := time.Now()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	member.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL, NULL)
	`, member.ID, member.TenantID, member.UserID, member.Role, member.Nickname, member.JoinedAt, member.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tenant member: %w", err)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) GetByID(ctx context.Context, id string) (*domain.TenantMember, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by
		FROM tenant_members
		WHERE id = ?
	`, id)
	return scanTenantMember(row)
}

func (r *SQLiteTenantMemberRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.TenantMember, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by
		FROM tenant_members
		WHERE user_id = ? AND tenant_id = ?
	`, userID, tenantID)
	return scanTenantMember(row)
}

func (r *SQLiteTenantMemberRepository) Update(ctx context.Context, member *domain.TenantMember) error {
	member.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_members
		SET role = ?, nickname = ?, updated_at = ?
		WHERE id = ?
	`, member.Role, member.Nickname, member.UpdatedAt, member.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant member: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tenant_members WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant member: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) ListByTenant(ctx context.Context, tenantID string, role string, limit int, cursor string) ([]*domain.TenantMember, string, error) {
	args := []interface{}{tenantID}
	query := `
		SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by
		FROM tenant_members
		WHERE tenant_id = ?
	`
	if role != "" {
		query += " AND role = ?"
		args = append(args, role)
	}
	if cursor != "" {
		query += " AND id > ?"
		args = append(args, cursor)
	}
	if limit <= 0 {
		limit = 50
	}
	query += " ORDER BY joined_at DESC, id DESC LIMIT ?"
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tenant members: %w", err)
	}
	defer rows.Close()

	var out []*domain.TenantMember
	for rows.Next() {
		member, err := scanTenantMember(rows)
		if err != nil {
			return nil, "", err
		}
		out = append(out, member)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate tenant members: %w", err)
	}
	next := ""
	if len(out) > limit {
		next = out[limit].ID
		out = out[:limit]
	}
	return out, next, nil
}

func (r *SQLiteTenantMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.TenantMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by
		FROM tenant_members
		WHERE user_id = ?
		ORDER BY joined_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	var out []*domain.TenantMember
	for rows.Next() {
		member, err := scanTenantMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, member)
	}
	return out, rows.Err()
}

func (r *SQLiteTenantMemberRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenant_members WHERE tenant_id = ?`, tenantID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count tenant members: %w", err)
	}
	return count, nil
}

func (r *SQLiteTenantMemberRepository) UpdateRole(ctx context.Context, id string, role domain.TenantRole) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tenant_members SET role = ?, updated_at = ? WHERE id = ?`, role, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) GetUserRole(ctx context.Context, userID, tenantID string) (domain.TenantRole, error) {
	var role domain.TenantRole
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM tenant_members WHERE user_id = ? AND tenant_id = ?
	`, userID, tenantID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", domain.NewError(domain.ErrNotFound, "User is not a member of this tenant", nil)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return role, nil
}

func (r *SQLiteTenantMemberRepository) HasRole(ctx context.Context, userID, tenantID string, requiredRole domain.TenantRole) (bool, error) {
	role, err := r.GetUserRole(ctx, userID, tenantID)
	if err != nil {
		if d, ok := err.(*domain.Error); ok && d.Code == domain.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return role.HasPermission(requiredRole), nil
}

func (r *SQLiteTenantMemberRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_members WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant members: %w", err)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) Ban(ctx context.Context, id string, bannedBy string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_members
		SET banned_at = ?, banned_by = ?, updated_at = ?
		WHERE id = ?
	`, now, bannedBy, now, id)
	if err != nil {
		return fmt.Errorf("failed to ban member: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) Unban(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_members
		SET banned_at = NULL, banned_by = '', updated_at = ?
		WHERE id = ?
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to unban member: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return nil
}

func (r *SQLiteTenantMemberRepository) IsBanned(ctx context.Context, userID, tenantID string) (bool, error) {
	var banned sql.NullBool
	err := r.db.QueryRowContext(ctx, `
		SELECT banned_at IS NOT NULL FROM tenant_members WHERE user_id = ? AND tenant_id = ?
	`, userID, tenantID).Scan(&banned)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check ban status: %w", err)
	}
	return banned.Valid && banned.Bool, nil
}

func (r *SQLiteTenantMemberRepository) ListBannedByTenant(ctx context.Context, tenantID string) ([]*domain.TenantMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, role, nickname, joined_at, updated_at, banned_at, banned_by
		FROM tenant_members
		WHERE tenant_id = ? AND banned_at IS NOT NULL
		ORDER BY banned_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list banned members: %w", err)
	}
	defer rows.Close()

	var out []*domain.TenantMember
	for rows.Next() {
		member, err := scanTenantMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, member)
	}
	return out, rows.Err()
}

func scanTenantMember(row interface {
	Scan(dest ...any) error
}) (*domain.TenantMember, error) {
	var m domain.TenantMember
	var nickname sql.NullString
	var bannedAt sql.NullTime
	var bannedBy sql.NullString
	if err := row.Scan(
		&m.ID,
		&m.TenantID,
		&m.UserID,
		&m.Role,
		&nickname,
		&m.JoinedAt,
		&m.UpdatedAt,
		&bannedAt,
		&bannedBy,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
		}
		return nil, fmt.Errorf("failed to scan tenant member: %w", err)
	}
	m.Nickname = nickname.String
	if bannedAt.Valid {
		m.BannedAt = &bannedAt.Time
	}
	m.BannedBy = bannedBy.String
	return &m, nil
}
