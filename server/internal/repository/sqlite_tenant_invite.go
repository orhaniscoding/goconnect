package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteTenantInviteRepository implements TenantInviteRepository using SQLite.
type SQLiteTenantInviteRepository struct {
	db *sql.DB
}

func NewSQLiteTenantInviteRepository(db *sql.DB) *SQLiteTenantInviteRepository {
	return &SQLiteTenantInviteRepository{db: db}
}

func (r *SQLiteTenantInviteRepository) Create(ctx context.Context, invite *domain.TenantInvite) error {
	if invite.ID == "" {
		invite.ID = domain.GenerateTenantInviteID()
	}
	invite.CreatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_invites (id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, invite.ID, invite.TenantID, strings.ToUpper(invite.Code), invite.MaxUses, invite.UseCount, invite.ExpiresAt, invite.CreatedBy, invite.CreatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrValidation, "Invite code already exists", nil)
		}
		return fmt.Errorf("failed to create tenant invite: %w", err)
	}
	return nil
}

func (r *SQLiteTenantInviteRepository) GetByID(ctx context.Context, id string) (*domain.TenantInvite, error) {
	return r.getBy(ctx, "id", id)
}

func (r *SQLiteTenantInviteRepository) GetByCode(ctx context.Context, code string) (*domain.TenantInvite, error) {
	return r.getBy(ctx, "code", strings.ToUpper(strings.TrimSpace(code)))
}

func (r *SQLiteTenantInviteRepository) getBy(ctx context.Context, field string, value interface{}) (*domain.TenantInvite, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM tenant_invites
		WHERE %s = ?
	`, field)

	var invite domain.TenantInvite
	var expiresAt, revokedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, value).Scan(
		&invite.ID,
		&invite.TenantID,
		&invite.Code,
		&invite.MaxUses,
		&invite.UseCount,
		&expiresAt,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&revokedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
		}
		return nil, fmt.Errorf("failed to get tenant invite: %w", err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	if revokedAt.Valid {
		invite.RevokedAt = &revokedAt.Time
	}
	return &invite, nil
}

func (r *SQLiteTenantInviteRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tenant_invites WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete invite: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *SQLiteTenantInviteRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.TenantInvite, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM tenant_invites
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant invites: %w", err)
	}
	defer rows.Close()

	var invites []*domain.TenantInvite
	for rows.Next() {
		invite, err := scanTenantInvite(rows)
		if err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (r *SQLiteTenantInviteRepository) IncrementUseCount(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tenant_invites SET use_count = use_count + 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to increment use count: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *SQLiteTenantInviteRepository) Revoke(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tenant_invites SET revoked_at = ? WHERE id = ?`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke invite: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *SQLiteTenantInviteRepository) DeleteExpired(ctx context.Context) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tenant_invites
		WHERE (expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP)
		   OR revoked_at IS NOT NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired invites: %w", err)
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func (r *SQLiteTenantInviteRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_invites WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant invites: %w", err)
	}
	return nil
}

func scanTenantInvite(row interface{ Scan(dest ...any) error }) (*domain.TenantInvite, error) {
	var invite domain.TenantInvite
	var expiresAt, revokedAt sql.NullTime
	if err := row.Scan(
		&invite.ID,
		&invite.TenantID,
		&invite.Code,
		&invite.MaxUses,
		&invite.UseCount,
		&expiresAt,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&revokedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
		}
		return nil, fmt.Errorf("failed to scan tenant invite: %w", err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	if revokedAt.Valid {
		invite.RevokedAt = &revokedAt.Time
	}
	return &invite, nil
}
