package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteTenantRepository implements TenantRepository using SQLite.
type SQLiteTenantRepository struct {
	db *sql.DB
}

func NewSQLiteTenantRepository(db *sql.DB) *SQLiteTenantRepository {
	return &SQLiteTenantRepository{db: db}
}

func (r *SQLiteTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (
			id, name, description, icon_url, visibility, access_type, password_hash,
			max_members, owner_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Description,
		tenant.IconURL,
		tenant.Visibility,
		tenant.AccessType,
		tenant.PasswordHash,
		tenant.MaxMembers,
		nullIfEmpty(tenant.OwnerID),
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

func (r *SQLiteTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		WHERE id = ?
	`
	tenant := &domain.Tenant{}
	var description, iconURL, passwordHash, ownerID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&description,
		&iconURL,
		&tenant.Visibility,
		&tenant.AccessType,
		&passwordHash,
		&tenant.MaxMembers,
		&ownerID,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrTenantNotFound, "Tenant not found", map[string]string{"tenant_id": id})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}
	tenant.Description = description.String
	tenant.IconURL = iconURL.String
	tenant.PasswordHash = passwordHash.String
	tenant.OwnerID = ownerID.String
	return tenant, nil
}

func (r *SQLiteTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE tenants
		SET name = ?, description = ?, icon_url = ?, visibility = ?, access_type = ?, password_hash = ?, max_members = ?, owner_id = ?, updated_at = ?
		WHERE id = ?
	`
	tenant.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query,
		tenant.Name,
		tenant.Description,
		tenant.IconURL,
		tenant.Visibility,
		tenant.AccessType,
		tenant.PasswordHash,
		tenant.MaxMembers,
		nullIfEmpty(tenant.OwnerID),
		tenant.UpdatedAt,
		tenant.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrTenantNotFound, "Tenant not found", map[string]string{"tenant_id": tenant.ID})
	}
	return nil
}

func (r *SQLiteTenantRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tenants WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: tenant not found", domain.ErrNotFound)
	}
	return nil
}

func (r *SQLiteTenantRepository) List(ctx context.Context) ([]*domain.Tenant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant := &domain.Tenant{}
		var description, iconURL, passwordHash, ownerID sql.NullString
		if err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&description,
			&iconURL,
			&tenant.Visibility,
			&tenant.AccessType,
			&passwordHash,
			&tenant.MaxMembers,
			&ownerID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.Description = description.String
		tenant.IconURL = iconURL.String
		tenant.PasswordHash = passwordHash.String
		tenant.OwnerID = ownerID.String
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tenants: %w", err)
	}
	return tenants, nil
}

func (r *SQLiteTenantRepository) ListAll(ctx context.Context, limit, offset int, queryStr string) ([]*domain.Tenant, int, error) {
	where := ""
	args := []interface{}{}
	if strings.TrimSpace(queryStr) != "" {
		where = "WHERE LOWER(name) LIKE ?"
		args = append(args, "%"+strings.ToLower(queryStr)+"%")
	}

	countQuery := "SELECT COUNT(*) FROM tenants " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	listQuery := `
		SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
	` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant := &domain.Tenant{}
		var description, iconURL, passwordHash, ownerID sql.NullString
		if err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&description,
			&iconURL,
			&tenant.Visibility,
			&tenant.AccessType,
			&passwordHash,
			&tenant.MaxMembers,
			&ownerID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.Description = description.String
		tenant.IconURL = iconURL.String
		tenant.PasswordHash = passwordHash.String
		tenant.OwnerID = ownerID.String
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tenants: %w", err)
	}
	return tenants, total, nil
}

func (r *SQLiteTenantRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	return count, nil
}
