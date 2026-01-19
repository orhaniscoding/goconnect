package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES ROLE REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresRoleRepository implements RoleRepository for PostgreSQL
type PostgresRoleRepository struct {
	db *sql.DB
}

// NewPostgresRoleRepository creates a new PostgresRoleRepository
func NewPostgresRoleRepository(db *sql.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

// Create creates a new role
func (r *PostgresRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, tenant_id, name, color, icon, position, is_default, is_admin, mentionable, hoist, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	if role.ID == "" {
		role.ID = domain.GenerateRoleID()
	}
	role.CreatedAt = now
	role.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		role.ID,
		role.TenantID,
		role.Name,
		role.Color,
		role.Icon,
		role.Position,
		role.IsDefault,
		role.IsAdmin,
		role.Mentionable,
		role.Hoist,
		role.CreatedAt,
		role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *PostgresRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	query := `
		SELECT id, tenant_id, name, color, icon, position, is_default, is_admin, mentionable, hoist, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Color,
		&role.Icon,
		&role.Position,
		&role.IsDefault,
		&role.IsAdmin,
		&role.Mentionable,
		&role.Hoist,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetByTenantID retrieves all roles for a tenant (ordered by position DESC)
func (r *PostgresRoleRepository) GetByTenantID(ctx context.Context, tenantID string) ([]domain.Role, error) {
	query := `
		SELECT id, tenant_id, name, color, icon, position, is_default, is_admin, mentionable, hoist, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY position DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Color,
			&role.Icon,
			&role.Position,
			&role.IsDefault,
			&role.IsAdmin,
			&role.Mentionable,
			&role.Hoist,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetDefaultRole retrieves the default role for a tenant
func (r *PostgresRoleRepository) GetDefaultRole(ctx context.Context, tenantID string) (*domain.Role, error) {
	query := `
		SELECT id, tenant_id, name, color, icon, position, is_default, is_admin, mentionable, hoist, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1 AND is_default = TRUE
		LIMIT 1
	`

	var role domain.Role
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.Color,
		&role.Icon,
		&role.Position,
		&role.IsDefault,
		&role.IsAdmin,
		&role.Mentionable,
		&role.Hoist,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	return &role, nil
}

// Update updates a role
func (r *PostgresRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	query := `
		UPDATE roles
		SET name = $2, color = $3, icon = $4, position = $5, mentionable = $6, hoist = $7, updated_at = $8
		WHERE id = $1
	`

	role.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		role.ID,
		role.Name,
		role.Color,
		role.Icon,
		role.Position,
		role.Mentionable,
		role.Hoist,
		role.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// Delete deletes a role
func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// UpdatePositions updates positions for multiple roles
func (r *PostgresRoleRepository) UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE roles
		SET position = $2, updated_at = $3
		WHERE id = $1 AND tenant_id = $4
	`

	now := time.Now()
	for roleID, position := range positions {
		_, err := tx.ExecContext(ctx, query, roleID, position, now, tenantID)
		if err != nil {
			return fmt.Errorf("failed to update position for role %s: %w", roleID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPermissions retrieves all permissions for a role
func (r *PostgresRoleRepository) GetPermissions(ctx context.Context, roleID string) ([]domain.RolePermission, error) {
	query := `
		SELECT role_id, permission_id, allowed
		FROM role_permissions
		WHERE role_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []domain.RolePermission
	for rows.Next() {
		var perm domain.RolePermission
		if err := rows.Scan(&perm.RoleID, &perm.PermissionID, &perm.Allowed); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// SetPermissions sets permissions for a role (replaces all)
func (r *PostgresRoleRepository) SetPermissions(ctx context.Context, roleID string, permissions []domain.RolePermission) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing permissions
	_, err = tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Insert new permissions
	query := `INSERT INTO role_permissions (role_id, permission_id, allowed) VALUES ($1, $2, $3)`
	for _, perm := range permissions {
		_, err := tx.ExecContext(ctx, query, roleID, perm.PermissionID, perm.Allowed)
		if err != nil {
			return fmt.Errorf("failed to insert permission %s: %w", perm.PermissionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AssignToUser assigns a role to a user
func (r *PostgresRoleRepository) AssignToUser(ctx context.Context, userRole *domain.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	if userRole.AssignedAt.IsZero() {
		userRole.AssignedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		userRole.UserID,
		userRole.RoleID,
		userRole.AssignedAt,
		userRole.AssignedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

// RemoveFromUser removes a role from a user
func (r *PostgresRoleRepository) RemoveFromUser(ctx context.Context, userID, roleID string) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles for a user in a tenant
func (r *PostgresRoleRepository) GetUserRoles(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.name, r.color, r.icon, r.position, r.is_default, r.is_admin, r.mentionable, r.hoist, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.tenant_id = $2
		ORDER BY r.position DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(
			&role.ID,
			&role.TenantID,
			&role.Name,
			&role.Color,
			&role.Icon,
			&role.Position,
			&role.IsDefault,
			&role.IsAdmin,
			&role.Mentionable,
			&role.Hoist,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Ensure interface compliance
var _ RoleRepository = (*PostgresRoleRepository)(nil)
