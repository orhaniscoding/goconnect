package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresIPRuleRepository implements IPRuleRepository using PostgreSQL
type PostgresIPRuleRepository struct {
	db *sql.DB
}

// NewPostgresIPRuleRepository creates a new PostgreSQL-backed IP rule repository
func NewPostgresIPRuleRepository(db *sql.DB) *PostgresIPRuleRepository {
	return &PostgresIPRuleRepository{db: db}
}

// Create stores a new IP rule
func (r *PostgresIPRuleRepository) Create(ctx context.Context, rule *domain.IPRule) error {
	query := `
		INSERT INTO ip_rules (id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.TenantID,
		rule.Type,
		rule.CIDR,
		rule.Description,
		rule.CreatedBy,
		rule.CreatedAt,
		rule.UpdatedAt,
		rule.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create IP rule: %w", err)
	}
	return nil
}

// GetByID retrieves an IP rule by its ID
func (r *PostgresIPRuleRepository) GetByID(ctx context.Context, id string) (*domain.IPRule, error) {
	query := `
		SELECT id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at
		FROM ip_rules
		WHERE id = $1
	`
	rule := &domain.IPRule{}
	var expiresAt sql.NullTime
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.TenantID,
		&rule.Type,
		&rule.CIDR,
		&description,
		&rule.CreatedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&expiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get IP rule: %w", err)
	}
	if expiresAt.Valid {
		rule.ExpiresAt = &expiresAt.Time
	}
	rule.Description = description.String
	return rule, nil
}

// ListByTenant lists all active IP rules for a tenant
func (r *PostgresIPRuleRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.IPRule, error) {
	query := `
		SELECT id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at
		FROM ip_rules
		WHERE tenant_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list IP rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.IPRule
	for rows.Next() {
		rule := &domain.IPRule{}
		var expiresAt sql.NullTime
		var description sql.NullString
		err := rows.Scan(
			&rule.ID,
			&rule.TenantID,
			&rule.Type,
			&rule.CIDR,
			&description,
			&rule.CreatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&expiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan IP rule: %w", err)
		}
		if expiresAt.Valid {
			rule.ExpiresAt = &expiresAt.Time
		}
		rule.Description = description.String
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// Delete removes an IP rule
func (r *PostgresIPRuleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM ip_rules WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete IP rule: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}
	return nil
}

// DeleteExpired removes all expired IP rules
func (r *PostgresIPRuleRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `DELETE FROM ip_rules WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired IP rules: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return int(rows), nil
}

// Update updates an existing IP rule
func (r *PostgresIPRuleRepository) Update(ctx context.Context, rule *domain.IPRule) error {
	rule.UpdatedAt = time.Now()
	query := `
		UPDATE ip_rules 
		SET type = $1, cidr = $2, description = $3, updated_at = $4, expires_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		rule.Type,
		rule.CIDR,
		rule.Description,
		rule.UpdatedAt,
		rule.ExpiresAt,
		rule.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update IP rule: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}
	return nil
}
