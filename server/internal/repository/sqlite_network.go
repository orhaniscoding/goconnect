package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteNetworkRepository implements NetworkRepository for SQLite.
type SQLiteNetworkRepository struct {
	db *sql.DB
}

func NewSQLiteNetworkRepository(db *sql.DB) *SQLiteNetworkRepository {
	return &SQLiteNetworkRepository{db: db}
}

func (r *SQLiteNetworkRepository) Create(ctx context.Context, network *domain.Network) error {
	query := `
		INSERT INTO networks (
			id, tenant_id, name, cidr, visibility, join_policy, dns, mtu, split_tunnel,
			created_by, description, required_role, is_hidden, created_at, updated_at, moderation_redacted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', 'member', 0, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		network.ID,
		network.TenantID,
		network.Name,
		network.CIDR,
		network.Visibility,
		network.JoinPolicy,
		nullableString(network.DNS),
		nullableInt(network.MTU),
		nullableBool(network.SplitTunnel),
		network.CreatedBy,
		network.CreatedAt,
		network.UpdatedAt,
		network.ModerationRedacted,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrInvalidRequest,
				fmt.Sprintf("Network with name '%s' already exists", network.Name),
				map[string]string{"field": "name"})
		}
		return fmt.Errorf("failed to create network: %w", err)
	}
	return nil
}

func (r *SQLiteNetworkRepository) GetByID(ctx context.Context, id string) (*domain.Network, error) {
	query := `
		SELECT id, tenant_id, name, cidr, visibility, join_policy, dns, mtu, split_tunnel,
		       created_by, created_at, updated_at, deleted_at, moderation_redacted
		FROM networks
		WHERE id = ? AND deleted_at IS NULL
	`
	var n domain.Network
	var dns sql.NullString
	var mtu sql.NullInt64
	var split sql.NullBool
	var deletedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID,
		&n.TenantID,
		&n.Name,
		&n.CIDR,
		&n.Visibility,
		&n.JoinPolicy,
		&dns,
		&mtu,
		&split,
		&n.CreatedBy,
		&n.CreatedAt,
		&n.UpdatedAt,
		&deletedAt,
		&n.ModerationRedacted,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
		}
		return nil, fmt.Errorf("failed to get network: %w", err)
	}
	if dns.Valid {
		n.DNS = &dns.String
	}
	if mtu.Valid {
		val := int(mtu.Int64)
		n.MTU = &val
	}
	if split.Valid {
		val := split.Bool
		n.SplitTunnel = &val
	}
	if deletedAt.Valid {
		n.SoftDeletedAt = &deletedAt.Time
	}
	return &n, nil
}

func (r *SQLiteNetworkRepository) List(ctx context.Context, filter NetworkFilter) ([]*domain.Network, string, error) {
	args := []interface{}{}
	query := `
		SELECT id, tenant_id, name, cidr, visibility, join_policy, dns, mtu, split_tunnel,
		       created_by, created_at, updated_at, deleted_at, moderation_redacted
		FROM networks
		WHERE deleted_at IS NULL
	`
	if filter.TenantID != "" {
		query += " AND tenant_id = ?"
		args = append(args, filter.TenantID)
	}
	if filter.Visibility == "public" {
		query += " AND visibility = ?"
		args = append(args, domain.NetworkVisibilityPublic)
	} else if filter.Visibility == "mine" && filter.UserID != "" {
		query += " AND created_by = ?"
		args = append(args, filter.UserID)
	}
	if filter.Search != "" {
		query += " AND (LOWER(name) LIKE ? OR LOWER(cidr) LIKE ?)"
		term := "%" + strings.ToLower(filter.Search) + "%"
		args = append(args, term, term)
	}
	if filter.Cursor != "" {
		// use created_at of cursor ID to paginate
		query += " AND created_at < (SELECT created_at FROM networks WHERE id = ?)"
		args = append(args, filter.Cursor)
	}
	query += " ORDER BY created_at DESC"
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query += " LIMIT ?"
	args = append(args, limit+1) // fetch one extra for cursor

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list networks: %w", err)
	}
	defer rows.Close()

	var result []*domain.Network
	for rows.Next() {
		var n domain.Network
		var dns sql.NullString
		var mtu sql.NullInt64
		var split sql.NullBool
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&n.ID,
			&n.TenantID,
			&n.Name,
			&n.CIDR,
			&n.Visibility,
			&n.JoinPolicy,
			&dns,
			&mtu,
			&split,
			&n.CreatedBy,
			&n.CreatedAt,
			&n.UpdatedAt,
			&deletedAt,
			&n.ModerationRedacted,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan network: %w", err)
		}
		if dns.Valid {
			n.DNS = &dns.String
		}
		if mtu.Valid {
			val := int(mtu.Int64)
			n.MTU = &val
		}
		if split.Valid {
			val := split.Bool
			n.SplitTunnel = &val
		}
		if deletedAt.Valid {
			n.SoftDeletedAt = &deletedAt.Time
		}
		result = append(result, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate networks: %w", err)
	}

	next := ""
	if len(result) > limit {
		next = result[limit].ID
		result = result[:limit]
	}
	return result, next, nil
}

func (r *SQLiteNetworkRepository) CheckCIDROverlap(ctx context.Context, cidr string, excludeID string, tenantID string) (bool, error) {
	query := `
		SELECT cidr FROM networks
		WHERE cidr = ? AND deleted_at IS NULL
	`
	args := []interface{}{cidr}
	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}
	if tenantID != "" {
		query += " AND tenant_id = ?"
		args = append(args, tenantID)
	}
	var existing string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&existing)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check cidr overlap: %w", err)
	}
	return true, nil
}

func (r *SQLiteNetworkRepository) Update(ctx context.Context, id string, mutate func(n *domain.Network) error) (*domain.Network, error) {
	// load existing
	n, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := mutate(n); err != nil {
		return nil, err
	}
	n.UpdatedAt = time.Now()

	query := `
		UPDATE networks
		SET name = ?, cidr = ?, visibility = ?, join_policy = ?, dns = ?, mtu = ?, split_tunnel = ?, updated_at = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query,
		n.Name,
		n.CIDR,
		n.Visibility,
		n.JoinPolicy,
		nullableString(n.DNS),
		nullableInt(n.MTU),
		nullableBool(n.SplitTunnel),
		n.UpdatedAt,
		n.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update network: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}
	return n, nil
}

func (r *SQLiteNetworkRepository) SoftDelete(ctx context.Context, id string, at time.Time) error {
	res, err := r.db.ExecContext(ctx, `UPDATE networks SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`, at, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete network: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Network not found or already deleted", nil)
	}
	return nil
}

func (r *SQLiteNetworkRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM networks WHERE deleted_at IS NULL`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count networks: %w", err)
	}
	return count, nil
}

func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullableInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func nullableBool(b *bool) interface{} {
	if b == nil {
		return nil
	}
	return *b
}
