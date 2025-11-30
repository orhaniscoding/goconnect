package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteUserRepository implements UserRepository using SQLite (modernc.org/sqlite driver).
// recovery_codes are stored as a JSON-encoded TEXT array for portability.
type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Create(ctx context.Context, user *domain.User) error {
	recoveryCodesJSON, err := marshalRecoveryCodes(user.RecoveryCodes)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
			two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.Locale,
		user.IsAdmin,
		user.IsModerator,
		nullIfEmpty(user.TwoFAKey),
		user.TwoFAEnabled,
		recoveryCodesJSON,
		nullIfEmpty(user.AuthProvider),
		nullIfEmpty(user.ExternalID),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrEmailAlreadyExists, "Email already registered", map[string]string{"email": user.Email})
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *SQLiteUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
		       two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id,
		       created_at, updated_at
		FROM users
		WHERE id = ?
	`
	return r.scanUser(ctx, query, id)
}

func (r *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
		       two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id,
		       created_at, updated_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`
	return r.scanUser(ctx, query, email)
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *domain.User) error {
	recoveryCodesJSON, err := marshalRecoveryCodes(user.RecoveryCodes)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET email = ?, password_hash = ?, locale = ?, is_admin = ?, is_moderator = ?,
		    two_fa_key = ?, two_fa_enabled = ?, recovery_codes = ?, auth_provider = ?, external_id = ?,
		    updated_at = ?
		WHERE id = ?
	`
	user.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Locale,
		user.IsAdmin,
		user.IsModerator,
		nullIfEmpty(user.TwoFAKey),
		user.TwoFAEnabled,
		recoveryCodesJSON,
		nullIfEmpty(user.AuthProvider),
		nullIfEmpty(user.ExternalID),
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": user.ID})
	}
	return nil
}

func (r *SQLiteUserRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: user not found", domain.ErrNotFound)
	}
	return nil
}

func (r *SQLiteUserRepository) ListAll(ctx context.Context, limit, offset int, queryStr string) ([]*domain.User, int, error) {
	where := ""
	args := []interface{}{}
	if strings.TrimSpace(queryStr) != "" {
		where = "WHERE LOWER(email) LIKE ?"
		args = append(args, "%"+strings.ToLower(queryStr)+"%")
	}

	countQuery := "SELECT COUNT(*) FROM users " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	listQuery := `
		SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator,
		       two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id,
		       created_at, updated_at
		FROM users
	` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user, err := scanUserRow(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

func (r *SQLiteUserRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

func (r *SQLiteUserRepository) scanUser(ctx context.Context, query string, arg interface{}) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, query, arg)
	return scanUserRow(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUserRow(row scanner) (*domain.User, error) {
	user := &domain.User{}
	var twoFAKey, authProvider, externalID sql.NullString
	var recoveryCodesRaw sql.NullString
	err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Locale,
		&user.IsAdmin,
		&user.IsModerator,
		&twoFAKey,
		&user.TwoFAEnabled,
		&recoveryCodesRaw,
		&authProvider,
		&externalID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}
	user.TwoFAKey = twoFAKey.String
	user.AuthProvider = authProvider.String
	user.ExternalID = externalID.String
	if recoveryCodesRaw.Valid && recoveryCodesRaw.String != "" {
		var codes []string
		if err := json.Unmarshal([]byte(recoveryCodesRaw.String), &codes); err == nil {
			user.RecoveryCodes = codes
		}
	}
	return user, nil
}

func marshalRecoveryCodes(codes []string) (sql.NullString, error) {
	if len(codes) == 0 {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(codes)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("encode recovery_codes: %w", err)
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func nullIfEmpty(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
