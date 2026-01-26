package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteIPRuleRepository implements IPRuleRepository using SQLite.
type SQLiteIPRuleRepository struct {
	db *sql.DB
}

func NewSQLiteIPRuleRepository(db *sql.DB) *SQLiteIPRuleRepository {
	return &SQLiteIPRuleRepository{db: db}
}

func (r *SQLiteIPRuleRepository) Create(ctx context.Context, rule *domain.IPRule) error {
	if rule.ID == "" {
		rule.ID = domain.GenerateNetworkID()
	}
	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ip_rules (id, tenant_id, cidr, action, description, created_by, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, rule.ID, rule.TenantID, rule.CIDR, string(rule.Type), rule.Description, rule.CreatedBy, rule.ExpiresAt, rule.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create ip rule: %w", err)
	}
	return nil
}

func (r *SQLiteIPRuleRepository) GetByID(ctx context.Context, id string) (*domain.IPRule, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, cidr, action, description, created_by, expires_at, created_at
		FROM ip_rules
		WHERE id = ?
	`, id)
	return scanIPRule(row)
}

func (r *SQLiteIPRuleRepository) List(ctx context.Context, tenantID string) ([]*domain.IPRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, cidr, action, description, created_by, expires_at, created_at
		FROM ip_rules
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ip rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.IPRule
	for rows.Next() {
		rule, err := scanIPRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *SQLiteIPRuleRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.IPRule, error) {
	return r.List(ctx, tenantID)
}

func (r *SQLiteIPRuleRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM ip_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete ip rule: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}
	return nil
}

func (r *SQLiteIPRuleRepository) CheckIP(ctx context.Context, tenantID, ip string) (*domain.IPRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, cidr, action, description, created_by, expires_at, created_at
		FROM ip_rules
		WHERE tenant_id = ?
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check ip rules: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		rule, err := scanIPRule(rows)
		if err != nil {
			return nil, err
		}
		if rule.MatchesIP(ip) {
			return rule, nil
		}
	}
	return nil, domain.NewError(domain.ErrNotFound, "No matching IP rule", nil)
}

func scanIPRule(row interface{ Scan(dest ...any) error }) (*domain.IPRule, error) {
	var rule domain.IPRule
	var expiresAt sql.NullTime
	if err := row.Scan(
		&rule.ID, &rule.TenantID, &rule.CIDR, &rule.Type, &rule.Description,
		&rule.CreatedBy, &expiresAt, &rule.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
		}
		return nil, fmt.Errorf("failed to scan ip rule: %w", err)
	}
	if expiresAt.Valid {
		rule.ExpiresAt = &expiresAt.Time
	}
	return &rule, nil
}

// DeleteExpired removes expired rules.
func (r *SQLiteIPRuleRepository) DeleteExpired(ctx context.Context) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM ip_rules
		WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired ip rules: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return int(rows), nil
}
