package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES DISCOVERY REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresDiscoveryRepository implements DiscoveryRepository for PostgreSQL
type PostgresDiscoveryRepository struct {
	db *sql.DB
}

// NewPostgresDiscoveryRepository creates a new PostgresDiscoveryRepository
func NewPostgresDiscoveryRepository(db *sql.DB) *PostgresDiscoveryRepository {
	return &PostgresDiscoveryRepository{db: db}
}

// Upsert creates or updates discovery settings for a tenant
func (r *PostgresDiscoveryRepository) Upsert(ctx context.Context, discovery *domain.ServerDiscovery) error {
	query := `
		INSERT INTO server_discovery (tenant_id, enabled, category, tags, short_description, member_count, online_count, featured, verified, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			short_description = EXCLUDED.short_description,
			updated_at = EXCLUDED.updated_at
	`

	discovery.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		discovery.TenantID,
		discovery.Enabled,
		discovery.Category,
		pq.Array(discovery.Tags),
		discovery.ShortDescription,
		discovery.MemberCount,
		discovery.OnlineCount,
		discovery.Featured,
		discovery.Verified,
		discovery.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert discovery: %w", err)
	}

	return nil
}

// GetByTenantID retrieves discovery settings for a tenant
func (r *PostgresDiscoveryRepository) GetByTenantID(ctx context.Context, tenantID string) (*domain.ServerDiscovery, error) {
	query := `
		SELECT tenant_id, enabled, category, tags, short_description, member_count, online_count, featured, verified, updated_at
		FROM server_discovery
		WHERE tenant_id = $1
	`

	var discovery domain.ServerDiscovery
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&discovery.TenantID,
		&discovery.Enabled,
		&discovery.Category,
		pq.Array(&discovery.Tags),
		&discovery.ShortDescription,
		&discovery.MemberCount,
		&discovery.OnlineCount,
		&discovery.Featured,
		&discovery.Verified,
		&discovery.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery: %w", err)
	}

	return &discovery, nil
}

// Search searches discoverable servers
func (r *PostgresDiscoveryRepository) Search(ctx context.Context, filter DiscoveryFilter) ([]domain.ServerDiscovery, string, error) {
	query := `
		SELECT d.tenant_id, d.enabled, d.category, d.tags, d.short_description, d.member_count, d.online_count, d.featured, d.verified, d.updated_at
		FROM server_discovery d
		INNER JOIN tenants t ON d.tenant_id = t.id
		WHERE d.enabled = TRUE
	`

	args := []interface{}{}
	argCount := 0

	if filter.Query != "" {
		argCount++
		query += fmt.Sprintf(" AND (t.name ILIKE $%d OR d.short_description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Query+"%")
	}

	if filter.Category != nil {
		argCount++
		query += fmt.Sprintf(" AND d.category = $%d", argCount)
		args = append(args, *filter.Category)
	}

	if len(filter.Tags) > 0 {
		argCount++
		query += fmt.Sprintf(" AND d.tags && $%d", argCount)
		args = append(args, pq.Array(filter.Tags))
	}

	if filter.Featured != nil {
		argCount++
		query += fmt.Sprintf(" AND d.featured = $%d", argCount)
		args = append(args, *filter.Featured)
	}

	if filter.Verified != nil {
		argCount++
		query += fmt.Sprintf(" AND d.verified = $%d", argCount)
		args = append(args, *filter.Verified)
	}

	if filter.Cursor != "" {
		argCount++
		query += fmt.Sprintf(" AND d.tenant_id < $%d", argCount)
		args = append(args, filter.Cursor)
	}

	// Sorting
	switch filter.Sort {
	case "online_count":
		query += " ORDER BY d.online_count DESC"
	case "created_at":
		query += " ORDER BY d.updated_at DESC"
	default:
		query += " ORDER BY d.member_count DESC"
	}
	query += ", d.tenant_id DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to search discoveries: %w", err)
	}
	defer rows.Close()

	var discoveries []domain.ServerDiscovery
	for rows.Next() {
		var discovery domain.ServerDiscovery
		if err := rows.Scan(
			&discovery.TenantID,
			&discovery.Enabled,
			&discovery.Category,
			pq.Array(&discovery.Tags),
			&discovery.ShortDescription,
			&discovery.MemberCount,
			&discovery.OnlineCount,
			&discovery.Featured,
			&discovery.Verified,
			&discovery.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan discovery: %w", err)
		}
		discoveries = append(discoveries, discovery)
	}

	var nextCursor string
	if len(discoveries) > limit {
		nextCursor = discoveries[limit-1].TenantID
		discoveries = discoveries[:limit]
	}

	return discoveries, nextCursor, nil
}

// GetFeatured retrieves featured servers
func (r *PostgresDiscoveryRepository) GetFeatured(ctx context.Context, limit int) ([]domain.ServerDiscovery, error) {
	filter := DiscoveryFilter{
		Featured: boolPtr(true),
		Limit:    limit,
	}
	discoveries, _, err := r.Search(ctx, filter)
	return discoveries, err
}

// UpdateStats updates cached member/online counts
func (r *PostgresDiscoveryRepository) UpdateStats(ctx context.Context, tenantID string, memberCount, onlineCount int) error {
	query := `
		UPDATE server_discovery
		SET member_count = $2, online_count = $3, updated_at = $4
		WHERE tenant_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, tenantID, memberCount, onlineCount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	return nil
}

// SetFeatured sets the featured status for a server
func (r *PostgresDiscoveryRepository) SetFeatured(ctx context.Context, tenantID string, featured bool) error {
	query := `UPDATE server_discovery SET featured = $2, updated_at = $3 WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID, featured, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set featured: %w", err)
	}

	return nil
}

// SetVerified sets the verified status for a server
func (r *PostgresDiscoveryRepository) SetVerified(ctx context.Context, tenantID string, verified bool) error {
	query := `UPDATE server_discovery SET verified = $2, updated_at = $3 WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID, verified, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set verified: %w", err)
	}

	return nil
}

// Delete removes discovery settings
func (r *PostgresDiscoveryRepository) Delete(ctx context.Context, tenantID string) error {
	query := `DELETE FROM server_discovery WHERE tenant_id = $1`

	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete discovery: %w", err)
	}

	return nil
}

// Ensure interface compliance
var _ DiscoveryRepository = (*PostgresDiscoveryRepository)(nil)
