package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	now := time.Now()

	user := &domain.User{
		ID:            "user-1",
		TenantID:      "tenant-1",
		Email:         "test@example.com",
		PasswordHash:  "hashed",
		Locale:        "en-US",
		IsAdmin:       false,
		IsModerator:   false,
		TwoFAKey:      "",
		TwoFAEnabled:  false,
		RecoveryCodes: []string{"code1"},
		AuthProvider:  "local",
		ExternalID:    "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	query := `INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(
			user.ID, user.TenantID, user.Email, user.PasswordHash, user.Locale,
			user.IsAdmin, user.IsModerator, sql.NullString{Valid: false}, user.TwoFAEnabled,
			"{\"code1\"}", sql.NullString{String: "local", Valid: true}, sql.NullString{Valid: false},
			user.CreatedAt, user.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator, two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at FROM users WHERE id = $1`

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "email", "password_hash", "locale", "is_admin", "is_moderator",
		"two_fa_key", "two_fa_enabled", "recovery_codes", "auth_provider", "external_id", "created_at", "updated_at",
	}).AddRow(
		"user-1", "tenant-1", "test@example.com", "hashed", "en-US", false, false,
		nil, false, "{code1}", "local", nil, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("user-1").
		WillReturnRows(rows)

	user, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "local", user.AuthProvider)
	// Note: pq array parsing might need integration test, but for unit test string representation is often returned by mock.
	// Actually sqlmock driver doesn't fully simulate pq.Array scanning logic perfectly if rely on pq.Array wrapper.
	// But since we are mocking the *DB*, we mock what Scan receives.
	
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT id, tenant_id, email, password_hash, locale, is_admin, is_moderator, two_fa_key, two_fa_enabled, recovery_codes, auth_provider, external_id, created_at, updated_at FROM users WHERE email = $1 ORDER BY created_at DESC LIMIT 1`

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "email", "password_hash", "locale", "is_admin", "is_moderator",
		"two_fa_key", "two_fa_enabled", "recovery_codes", "auth_provider", "external_id", "created_at", "updated_at",
	}).AddRow(
		"user-1", "tenant-1", "test@example.com", "hashed", "en-US", false, false,
		nil, false, "{code1}", "local", nil, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	now := time.Now()

	user := &domain.User{
		ID:            "user-1",
		Email:         "updated@example.com",
		PasswordHash:  "newhash",
		Locale:        "fr-FR",
		IsAdmin:       true,
		IsModerator:   true,
		TwoFAKey:      "secret",
		TwoFAEnabled:  true,
		RecoveryCodes: []string{},
		AuthProvider:  "local",
		UpdatedAt:     now,
	}

	query := `UPDATE users SET email = $1, password_hash = $2, locale = $3, is_admin = $4, is_moderator = $5, two_fa_key = $6, two_fa_enabled = $7, recovery_codes = $8, auth_provider = $9, external_id = $10, updated_at = $11 WHERE id = $12`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(
			user.Email, user.PasswordHash, user.Locale, user.IsAdmin, user.IsModerator,
			sql.NullString{String: "secret", Valid: true}, user.TwoFAEnabled, "{}",
			sql.NullString{String: "local", Valid: true}, sql.NullString{Valid: false},
			sqlmock.AnyArg(), // UpdatedAt changes internally
			user.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(ctx, user)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	query := `DELETE FROM users WHERE id = $1`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, "user-1")
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	now := time.Now()

	countQuery := `SELECT COUNT(*) FROM users`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	listQuery := `SELECT id, tenant_id, email, password_hash, locale, is_admin, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "email", "password_hash", "locale", "is_admin", "created_at", "updated_at",
	}).AddRow(
		"user-1", "tenant-1", "test@example.com", "hashed", "en-US", false, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).
		WithArgs(10, 0).
		WillReturnRows(rows)

	users, total, err := repo.ListAll(ctx, 10, 0, "")
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, users, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_ListAll_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()
	
	// Expect WHERE clause due to search
	countQuery := `SELECT COUNT(*) FROM users WHERE email ILIKE $1`
	mock.ExpectQuery(regexp.QuoteMeta(countQuery)).
		WithArgs("%test%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	listQuery := `SELECT id, tenant_id, email, password_hash, locale, is_admin, created_at, updated_at FROM users WHERE email ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).
		WithArgs("%test%", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "email", "password_hash", "locale", "is_admin", "created_at", "updated_at",
		}))

	users, total, err := repo.ListAll(ctx, 10, 0, "test")
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, users, 0)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	query := `SELECT COUNT(*) FROM users`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 42, count)

	require.NoError(t, mock.ExpectationsWereMet())
}
