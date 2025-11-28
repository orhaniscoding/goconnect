package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteInviteTokenRepository implements InviteTokenRepository using SQLite.
type SQLiteInviteTokenRepository struct {
	db *sql.DB
}

func NewSQLiteInviteTokenRepository(db *sql.DB) *SQLiteInviteTokenRepository {
	return &SQLiteInviteTokenRepository{db: db}
}

func (r *SQLiteInviteTokenRepository) Create(ctx context.Context, token *domain.InviteToken) error {
	if token.ID == "" {
		token.ID = domain.GenerateNetworkID()
	}
	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	usesLeft := token.UsesLeft
	if usesLeft == 0 && token.UsesMax > 0 {
		usesLeft = token.UsesMax
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO invite_tokens (id, network_id, token, max_uses, use_count, expires_at, created_by, created_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)
	`, token.ID, token.NetworkID, token.Token, token.UsesMax, usesLeft, token.ExpiresAt, token.CreatedBy, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invite token: %w", err)
	}
	return nil
}

func (r *SQLiteInviteTokenRepository) GetByToken(ctx context.Context, token string) (*domain.InviteToken, error) {
	return r.get(ctx, `token = ?`, token)
}

func (r *SQLiteInviteTokenRepository) GetByID(ctx context.Context, id string) (*domain.InviteToken, error) {
	return r.get(ctx, `id = ?`, id)
}

func (r *SQLiteInviteTokenRepository) get(ctx context.Context, where string, arg interface{}) (*domain.InviteToken, error) {
	query := fmt.Sprintf(`
		SELECT id, network_id, token, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM invite_tokens
		WHERE %s
	`, where)
	var it domain.InviteToken
	var expiresAt, revokedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&it.ID, &it.NetworkID, &it.Token, &it.UsesMax, &it.UsesLeft, &expiresAt, &it.CreatedBy, &it.CreatedAt, &revokedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Invite token not found", nil)
		}
		return nil, fmt.Errorf("failed to get invite token: %w", err)
	}
	if expiresAt.Valid {
		it.ExpiresAt = expiresAt.Time
	}
	if revokedAt.Valid {
		it.RevokedAt = &revokedAt.Time
	}
	return &it, nil
}

func (r *SQLiteInviteTokenRepository) IncrementUseCount(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE invite_tokens
		SET use_count = CASE 
				WHEN max_uses = 0 THEN use_count
				ELSE CASE WHEN use_count > 0 THEN use_count - 1 ELSE use_count END
			END
		WHERE id = ? AND (max_uses = 0 OR use_count > 0)
	`, id)
	if err != nil {
		return fmt.Errorf("failed to increment use count: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Invite token not found or exhausted", nil)
	}
	return nil
}

func (r *SQLiteInviteTokenRepository) Revoke(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE invite_tokens SET revoked_at = ? WHERE id = ?`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke invite token: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Invite token not found", nil)
	}
	return nil
}

func (r *SQLiteInviteTokenRepository) DeleteByID(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM invite_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete invite token: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Invite token not found", nil)
	}
	return nil
}

func (r *SQLiteInviteTokenRepository) DeleteExpired(ctx context.Context) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM invite_tokens
		WHERE (expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP)
		   OR revoked_at IS NOT NULL
		   OR (max_uses > 0 AND use_count <= 0)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired invite tokens: %w", err)
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func (r *SQLiteInviteTokenRepository) ListByNetwork(ctx context.Context, networkID string) ([]*domain.InviteToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, token, max_uses, use_count, expires_at, created_by, created_at, revoked_at
		FROM invite_tokens
		WHERE network_id = ?
		ORDER BY created_at DESC
	`, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invite tokens: %w", err)
	}
	defer rows.Close()

	var result []*domain.InviteToken
	for rows.Next() {
		it, err := r.scanInviteToken(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	return result, rows.Err()
}

func (r *SQLiteInviteTokenRepository) scanInviteToken(row interface{ Scan(dest ...any) error }) (*domain.InviteToken, error) {
	var it domain.InviteToken
	var expiresAt, revokedAt sql.NullTime
	if err := row.Scan(
		&it.ID, &it.NetworkID, &it.Token, &it.UsesMax, &it.UsesLeft,
		&expiresAt, &it.CreatedBy, &it.CreatedAt, &revokedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Invite token not found", nil)
		}
		return nil, fmt.Errorf("failed to scan invite token: %w", err)
	}
	if expiresAt.Valid {
		it.ExpiresAt = expiresAt.Time
	}
	if revokedAt.Valid {
		it.RevokedAt = &revokedAt.Time
	}
	return &it, nil
}

// UseToken decrements uses_left and returns the token if valid.
func (r *SQLiteInviteTokenRepository) UseToken(ctx context.Context, token string) (*domain.InviteToken, error) {
	it, err := r.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if it.RevokedAt != nil {
		return nil, domain.NewError(domain.ErrInviteTokenRevoked, "Invite token has been revoked", nil)
	}
	if time.Now().After(it.ExpiresAt) || (it.UsesMax > 0 && it.UsesLeft <= 0) {
		return nil, domain.NewError(domain.ErrInviteTokenExpired, "Invite token has expired or reached max uses", nil)
	}
	// decrement uses_left
	if it.UsesMax > 0 && it.UsesLeft > 0 {
		it.UsesLeft--
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE invite_tokens
		SET use_count = ?
		WHERE id = ?
	`, it.UsesLeft, it.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrement uses_left: %w", err)
	}
	return it, nil
}
