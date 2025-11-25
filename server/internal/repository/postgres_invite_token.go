package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresInviteTokenRepository implements InviteTokenRepository using PostgreSQL
type PostgresInviteTokenRepository struct {
	db *sql.DB
}

// NewPostgresInviteTokenRepository creates a new PostgreSQL-backed invite token repository
func NewPostgresInviteTokenRepository(db *sql.DB) *PostgresInviteTokenRepository {
	return &PostgresInviteTokenRepository{db: db}
}

// Create stores a new invite token
func (r *PostgresInviteTokenRepository) Create(ctx context.Context, token *domain.InviteToken) error {
	query := `
		INSERT INTO invite_tokens (id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.NetworkID,
		token.TenantID,
		token.Token,
		token.CreatedBy,
		token.ExpiresAt,
		token.UsesMax,
		token.UsesLeft,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create invite token: %w", err)
	}
	return nil
}

// GetByID retrieves an invite token by its ID
func (r *PostgresInviteTokenRepository) GetByID(ctx context.Context, id string) (*domain.InviteToken, error) {
	query := `
		SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at
		FROM invite_tokens
		WHERE id = $1
	`
	token := &domain.InviteToken{}
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID,
		&token.NetworkID,
		&token.TenantID,
		&token.Token,
		&token.CreatedBy,
		&token.ExpiresAt,
		&token.UsesMax,
		&token.UsesLeft,
		&token.CreatedAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite token: %w", err)
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	return token, nil
}

// GetByToken retrieves an invite token by the token string
func (r *PostgresInviteTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.InviteToken, error) {
	query := `
		SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at
		FROM invite_tokens
		WHERE token = $1
	`
	token := &domain.InviteToken{}
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.NetworkID,
		&token.TenantID,
		&token.Token,
		&token.CreatedBy,
		&token.ExpiresAt,
		&token.UsesMax,
		&token.UsesLeft,
		&token.CreatedAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite token: %w", err)
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	return token, nil
}

// ListByNetwork lists all active invite tokens for a network
func (r *PostgresInviteTokenRepository) ListByNetwork(ctx context.Context, networkID string) ([]*domain.InviteToken, error) {
	query := `
		SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at
		FROM invite_tokens
		WHERE network_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invite tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.InviteToken
	for rows.Next() {
		token := &domain.InviteToken{}
		var revokedAt sql.NullTime
		err := rows.Scan(
			&token.ID,
			&token.NetworkID,
			&token.TenantID,
			&token.Token,
			&token.CreatedBy,
			&token.ExpiresAt,
			&token.UsesMax,
			&token.UsesLeft,
			&token.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invite token: %w", err)
		}
		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

// UseToken decrements the uses_left counter and returns the token
func (r *PostgresInviteTokenRepository) UseToken(ctx context.Context, tokenStr string) (*domain.InviteToken, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get token with lock
	query := `
		SELECT id, network_id, tenant_id, token, created_by, expires_at, uses_max, uses_left, created_at, revoked_at
		FROM invite_tokens
		WHERE token = $1
		FOR UPDATE
	`
	token := &domain.InviteToken{}
	var revokedAt sql.NullTime
	err = tx.QueryRowContext(ctx, query, tokenStr).Scan(
		&token.ID,
		&token.NetworkID,
		&token.TenantID,
		&token.Token,
		&token.CreatedBy,
		&token.ExpiresAt,
		&token.UsesMax,
		&token.UsesLeft,
		&token.CreatedAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite token: %w", err)
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	// Check validity
	if !token.IsValid() {
		if token.RevokedAt != nil {
			return nil, domain.NewError(domain.ErrInviteTokenRevoked, "Invite token has been revoked", nil)
		}
		return nil, domain.NewError(domain.ErrInviteTokenExpired, "Invite token has expired or reached max uses", nil)
	}

	// Decrement uses_left if limited
	if token.UsesMax > 0 {
		token.UsesLeft--
		updateQuery := `UPDATE invite_tokens SET uses_left = $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, updateQuery, token.UsesLeft, token.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update uses_left: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return token, nil
}

// Revoke marks an invite token as revoked
func (r *PostgresInviteTokenRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE invite_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke invite token: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found or already revoked", nil)
	}
	return nil
}

// DeleteExpired removes all expired tokens
func (r *PostgresInviteTokenRepository) DeleteExpired(ctx context.Context) (int, error) {
	query := `DELETE FROM invite_tokens WHERE expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return int(rows), nil
}
