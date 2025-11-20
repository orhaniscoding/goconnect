package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresNetworkRepository implements NetworkRepository using PostgreSQL
type PostgresNetworkRepository struct {
	db *sql.DB
}

// NewPostgresNetworkRepository creates a new PostgreSQL-backed network repository
func NewPostgresNetworkRepository(db *sql.DB) *PostgresNetworkRepository {
	return &PostgresNetworkRepository{db: db}
}

func (r *PostgresNetworkRepository) Create(ctx context.Context, network *domain.Network) error {
	query := `
		INSERT INTO networks (id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		network.ID,
		network.TenantID,
		network.Name,
		network.CIDR,
		network.Visibility,
		network.CreatedBy,
		network.CreatedAt,
		network.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return domain.NewError(domain.ErrInvalidRequest,
				fmt.Sprintf("Network with name '%s' already exists", network.Name),
				map[string]string{"field": "name"})
		}
		return fmt.Errorf("failed to create network: %w", err)
	}
	return nil
}

func (r *PostgresNetworkRepository) GetByID(ctx context.Context, id string) (*domain.Network, error) {
	query := `
		SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at
		FROM networks
		WHERE id = $1 AND deleted_at IS NULL
	`
	network := &domain.Network{}
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&network.ID,
		&network.TenantID,
		&network.Name,
		&network.CIDR,
		&network.Visibility,
		&network.CreatedBy,
		&network.CreatedAt,
		&network.UpdatedAt,
		&deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: network not found", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get network by ID: %w", err)
	}

	if deletedAt.Valid {
		network.SoftDeletedAt = &deletedAt.Time
	}

	return network, nil
}

func (r *PostgresNetworkRepository) List(ctx context.Context, filter NetworkFilter) ([]*domain.Network, string, error) {
	query := `
		SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at
		FROM networks
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argIndex := 1

	// Apply visibility filter
	if filter.Visibility == "public" {
		query += fmt.Sprintf(" AND visibility = $%d", argIndex)
		args = append(args, domain.NetworkVisibilityPublic)
		argIndex++
	} else if filter.Visibility == "mine" && filter.UserID != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}
	// "all" visibility requires admin (no filter needed)

	// Apply search filter
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR cidr ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	// Apply cursor pagination
	if filter.Cursor != "" {
		query += fmt.Sprintf(" AND created_at < (SELECT created_at FROM networks WHERE id = $%d)", argIndex)
		args = append(args, filter.Cursor)
		argIndex++
	}

	// Order and limit
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit+1) // +1 to check if there's more
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list networks: %w", err)
	}
	defer rows.Close()

	var networks []*domain.Network
	for rows.Next() {
		network := &domain.Network{}
		var deletedAt sql.NullTime

		err := rows.Scan(
			&network.ID,
			&network.TenantID,
			&network.Name,
			&network.CIDR,
			&network.Visibility,
			&network.CreatedBy,
			&network.CreatedAt,
			&network.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan network: %w", err)
		}

		if deletedAt.Valid {
			network.SoftDeletedAt = &deletedAt.Time
		}

		networks = append(networks, network)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate networks: %w", err)
	}

	// Pagination: check if there's more
	nextCursor := ""
	if filter.Limit > 0 && len(networks) > filter.Limit {
		networks = networks[:filter.Limit]
		nextCursor = networks[len(networks)-1].ID
	}

	return networks, nextCursor, nil
}

func (r *PostgresNetworkRepository) CheckCIDROverlap(ctx context.Context, cidr string, excludeID string, tenantID string) (bool, error) {
	// PostgreSQL has inet type but for simplicity we'll do basic string check
	// In production, consider using host(network) for proper overlap detection
	query := `
		SELECT EXISTS(
			SELECT 1 FROM networks
			WHERE cidr = $1 AND id != $2 AND tenant_id = $3 AND deleted_at IS NULL
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, cidr, excludeID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check CIDR overlap: %w", err)
	}
	return exists, nil
}

func (r *PostgresNetworkRepository) Update(ctx context.Context, id string, mutate func(n *domain.Network) error) (*domain.Network, error) {
	// Start transaction for read-modify-write
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Read current network
	query := `
		SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at
		FROM networks
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`
	network := &domain.Network{}
	var deletedAt sql.NullTime

	err = tx.QueryRowContext(ctx, query, id).Scan(
		&network.ID,
		&network.TenantID,
		&network.Name,
		&network.CIDR,
		&network.Visibility,
		&network.CreatedBy,
		&network.CreatedAt,
		&network.UpdatedAt,
		&deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: network not found", domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get network for update: %w", err)
	}

	if deletedAt.Valid {
		network.SoftDeletedAt = &deletedAt.Time
	}

	// Apply mutation
	if err := mutate(network); err != nil {
		return nil, err
	}

	// Write updated network
	updateQuery := `
		UPDATE networks
		SET name = $1, cidr = $2, visibility = $3, updated_at = $4
		WHERE id = $5
	`
	network.UpdatedAt = time.Now()
	_, err = tx.ExecContext(ctx, updateQuery,
		network.Name,
		network.CIDR,
		network.Visibility,
		network.UpdatedAt,
		network.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update network: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return network, nil
}

func (r *PostgresNetworkRepository) SoftDelete(ctx context.Context, id string, at time.Time) error {
	query := `
		UPDATE networks
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, at, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete network: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: network not found or already deleted", domain.ErrNotFound)
	}
	return nil
}
