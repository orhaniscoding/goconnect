package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresTenantInviteRepository implements TenantInviteRepository using PostgreSQL
type PostgresTenantInviteRepository struct {
	db *sql.DB
}

// NewPostgresTenantInviteRepository creates a new PostgreSQL-backed repository
func NewPostgresTenantInviteRepository(db *sql.DB) *PostgresTenantInviteRepository {
	return &PostgresTenantInviteRepository{db: db}
}

func (r *PostgresTenantInviteRepository) Create(ctx context.Context, invite *domain.TenantInvite) error {
	if invite.ID == "" {
		invite.ID = domain.GenerateTenantInviteID()
	}

	query := `
		INSERT INTO tenant_invites (id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	invite.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		invite.ID,
		invite.TenantID,
		strings.ToUpper(invite.Code),
		invite.MaxUses,
		invite.UseCount,
		invite.ExpiresAt,
		invite.CreatedBy,
		invite.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.NewError(domain.ErrValidation, "Invite code already exists", nil)
		}
		return fmt.Errorf("failed to create tenant invite: %w", err)
	}
	return nil
}

func (r *PostgresTenantInviteRepository) GetByID(ctx context.Context, id string) (*domain.TenantInvite, error) {
	query := `
		SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM tenant_invites
		WHERE id = $1
	`
	invite := &domain.TenantInvite{}
	var expiresAt, revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invite.ID,
		&invite.TenantID,
		&invite.Code,
		&invite.MaxUses,
		&invite.UseCount,
		&expiresAt,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant invite: %w", err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	if revokedAt.Valid {
		invite.RevokedAt = &revokedAt.Time
	}
	return invite, nil
}

func (r *PostgresTenantInviteRepository) GetByCode(ctx context.Context, code string) (*domain.TenantInvite, error) {
	query := `
		SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM tenant_invites
		WHERE UPPER(code) = UPPER($1)
	`
	invite := &domain.TenantInvite{}
	var expiresAt, revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(code)).Scan(
		&invite.ID,
		&invite.TenantID,
		&invite.Code,
		&invite.MaxUses,
		&invite.UseCount,
		&expiresAt,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Invalid invite code", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite by code: %w", err)
	}
	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}
	if revokedAt.Valid {
		invite.RevokedAt = &revokedAt.Time
	}
	return invite, nil
}

func (r *PostgresTenantInviteRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenant_invites WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant invite: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *PostgresTenantInviteRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.TenantInvite, error) {
	query := `
		SELECT id, tenant_id, code, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM tenant_invites
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant invites: %w", err)
	}
	defer rows.Close()

	var invites []*domain.TenantInvite
	for rows.Next() {
		invite := &domain.TenantInvite{}
		var expiresAt, revokedAt sql.NullTime
		err := rows.Scan(
			&invite.ID,
			&invite.TenantID,
			&invite.Code,
			&invite.MaxUses,
			&invite.UseCount,
			&expiresAt,
			&invite.CreatedBy,
			&invite.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant invite: %w", err)
		}
		if expiresAt.Valid {
			invite.ExpiresAt = &expiresAt.Time
		}
		if revokedAt.Valid {
			invite.RevokedAt = &revokedAt.Time
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

func (r *PostgresTenantInviteRepository) IncrementUseCount(ctx context.Context, id string) error {
	query := `UPDATE tenant_invites SET use_count = use_count + 1 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment use count: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *PostgresTenantInviteRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE tenant_invites SET revoked_at = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to revoke tenant invite: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return nil
}

func (r *PostgresTenantInviteRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `
		DELETE FROM tenant_invites
		WHERE (expires_at IS NOT NULL AND expires_at < NOW())
		   OR revoked_at IS NOT NULL
	`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired invites: %w", err)
	}
	rows, err := result.RowsAffected()
	return int(rows), nil
}

func (r *PostgresTenantInviteRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenant_invites WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant invites: %w", err)
	}
	return nil
}
