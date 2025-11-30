package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresTenantRepository implements TenantRepository using PostgreSQL
type PostgresTenantRepository struct {
	db *sql.DB
}

// NewPostgresTenantRepository creates a new PostgreSQL-backed tenant repository
func NewPostgresTenantRepository(db *sql.DB) *PostgresTenantRepository {
	return &PostgresTenantRepository{db: db}
}

func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	query := `
		SELECT 
			id,
			name,
			description,
			icon_url,
			visibility,
			access_type,
			password_hash,
			max_members,
			owner_id,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1
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
		return nil, fmt.Errorf("%s: tenant not found", domain.ErrNotFound)
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

func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3
	`
	tenant.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		tenant.Name,
		tenant.UpdatedAt,
		tenant.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: tenant not found", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresTenantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenants WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: tenant not found", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresTenantRepository) List(ctx context.Context) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		tenant := &domain.Tenant{}
		var description, iconURL, passwordHash, ownerID sql.NullString
		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.Description = description.String
		tenant.IconURL = iconURL.String
		tenant.PasswordHash = passwordHash.String
		tenant.OwnerID = ownerID.String
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenants, nil
}

func (r *PostgresTenantRepository) ListAll(ctx context.Context, limit, offset int, query string) ([]*domain.Tenant, int, error) {
	// Build query
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if query != "" {
		whereClause = fmt.Sprintf("WHERE name ILIKE $%d", argIdx)
		args = append(args, "%"+query+"%")
		argIdx++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM tenants " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Get data
	listQuery := fmt.Sprintf(`
		SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

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
		err := rows.Scan(
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
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.Description = description.String
		tenant.IconURL = iconURL.String
		tenant.PasswordHash = passwordHash.String
		tenant.OwnerID = ownerID.String
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenants, total, nil
}

func (r *PostgresTenantRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tenants`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	return count, nil
}
