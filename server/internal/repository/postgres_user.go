package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, 
			two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.Locale,
		user.IsAdmin,
		user.IsModerator,
		sql.NullString{String: user.TwoFAKey, Valid: user.TwoFAKey != ""},
		user.TwoFAEnabled,
		pq.Array(user.RecoveryCodes),
		sql.NullString{String: user.AuthProvider, Valid: user.AuthProvider != ""},
		sql.NullString{String: user.ExternalID, Valid: user.ExternalID != ""},
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return domain.NewError(domain.ErrEmailAlreadyExists, "Email already registered", map[string]string{"email": user.Email})
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
			two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}
	var twoFAKey, authProvider, externalID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Locale,
		&user.IsAdmin,
		&user.IsModerator,
		&twoFAKey,
		&user.TwoFAEnabled,
		pq.Array(&user.RecoveryCodes),
		&authProvider,
		&externalID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": id})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	user.TwoFAKey = twoFAKey.String
	user.AuthProvider = authProvider.String
	user.ExternalID = externalID.String
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
			two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at
		FROM users
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	user := &domain.User{}
	var twoFAKey, authProvider, externalID sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Locale,
		&user.IsAdmin,
		&user.IsModerator,
		&twoFAKey,
		&user.TwoFAEnabled,
		pq.Array(&user.RecoveryCodes),
		&authProvider,
		&externalID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"email": email})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	user.TwoFAKey = twoFAKey.String
	user.AuthProvider = authProvider.String
	user.ExternalID = externalID.String
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, locale = $3, is_admin = $4, is_moderator = $5,
			two_fa_key = $6, two_fa_enabled = $7, recovery_codes = $8, 
			auth_provider = $9, external_id = $10, updated_at = $11
		WHERE id = $12
	`
	user.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Locale,
		user.IsAdmin,
		user.IsModerator,
		sql.NullString{String: user.TwoFAKey, Valid: user.TwoFAKey != ""},
		user.TwoFAEnabled,
		pq.Array(user.RecoveryCodes),
		sql.NullString{String: user.AuthProvider, Valid: user.AuthProvider != ""},
		sql.NullString{String: user.ExternalID, Valid: user.ExternalID != ""},
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": user.ID})
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: user not found", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresUserRepository) List(ctx context.Context, tenantID string) ([]*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Email,
			&user.PasswordHash,
			&user.Locale,
			&user.IsAdmin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) ListAll(ctx context.Context, limit, offset int, queryStr string) ([]*domain.User, int, error) {
	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if queryStr != "" {
		whereClause = "WHERE email ILIKE $" + strconv.Itoa(argIdx)
		args = append(args, "%"+queryStr+"%")
		argIdx++
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	listQuery := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, created_at, updated_at
		FROM users
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Email,
			&user.PasswordHash,
			&user.Locale,
			&user.IsAdmin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

func (r *PostgresUserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
