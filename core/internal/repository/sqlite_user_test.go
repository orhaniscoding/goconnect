package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLiteUserTest(t *testing.T) (*SQLiteUserRepository, func()) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "users.db")

	db, err := database.ConnectSQLite(dbPath)
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	// seed tenant
	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	return NewSQLiteUserRepository(db), func() { db.Close() }
}

func TestSQLiteUserRepository_CreateAndFetch(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	u := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		IsAdmin:      true,
		IsModerator:  false,
		TwoFAKey:     "secret",
		TwoFAEnabled: true,
		RecoveryCodes: []string{
			"code1", "code2",
		},
		AuthProvider: "local",
		ExternalID:   "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)
	assert.Equal(t, u.Email, got.Email)
	assert.True(t, got.TwoFAEnabled)
	assert.Equal(t, "secret", got.TwoFAKey)
	assert.Len(t, got.RecoveryCodes, 2)
}

func TestSQLiteUserRepository_GetByID(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Test not found
	_, err := repo.GetByID(ctx, "nonexistent")
	assert.Error(t, err)

	// Create user
	u := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	// Get by ID
	got, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", got.Email)
}

func TestSQLiteUserRepository_Update(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	u := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	// Update user
	u.Email = "updated@example.com"
	u.IsAdmin = true
	u.TwoFAKey = "newsecret"
	u.TwoFAEnabled = true
	u.RecoveryCodes = []string{"new1", "new2", "new3"}

	require.NoError(t, repo.Update(ctx, u))

	// Verify update
	got, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", got.Email)
	assert.True(t, got.IsAdmin)
	assert.Equal(t, "newsecret", got.TwoFAKey)
	assert.True(t, got.TwoFAEnabled)
	assert.Len(t, got.RecoveryCodes, 3)

	// Test update nonexistent user
	nonexistent := &domain.User{ID: "nonexistent", TenantID: "tenant-1", Email: "x@example.com"}
	err = repo.Update(ctx, nonexistent)
	assert.Error(t, err)
}

func TestSQLiteUserRepository_Delete(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	u := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	// Delete user
	require.NoError(t, repo.Delete(ctx, "user-1"))

	// Verify deletion
	_, err := repo.GetByID(ctx, "user-1")
	assert.Error(t, err)

	// Test delete nonexistent
	err = repo.Delete(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestSQLiteUserRepository_ListAll(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple users
	for i := 1; i <= 5; i++ {
		u := &domain.User{
			ID:           fmt.Sprintf("user-%d", i),
			TenantID:     "tenant-1",
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hash",
			Locale:       "en",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, u))
	}

	// Test list all
	users, total, err := repo.ListAll(ctx, 10, 0, "")
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, users, 5)

	// Test with limit
	users, total, err = repo.ListAll(ctx, 2, 0, "")
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, users, 2)

	// Test with offset
	users, _, err = repo.ListAll(ctx, 10, 3, "")
	require.NoError(t, err)
	assert.Len(t, users, 2)

	// Test with search query
	users, total, err = repo.ListAll(ctx, 10, 0, "user1")
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, users, 1)
}

func TestSQLiteUserRepository_Count(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Empty count
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create users
	for i := 1; i <= 3; i++ {
		u := &domain.User{
			ID:           fmt.Sprintf("user-%d", i),
			TenantID:     "tenant-1",
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hash",
			Locale:       "en",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		require.NoError(t, repo.Create(ctx, u))
	}

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestSQLiteUserRepository_DuplicateEmail(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	u1 := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u1))

	// Try to create with same email
	u2 := &domain.User{
		ID:           "user-2",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Locale:       "en",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, u2)
	assert.Error(t, err)
}

func TestSQLiteUserRepository_EmptyRecoveryCodes(t *testing.T) {
	repo, cleanup := setupSQLiteUserTest(t)
	defer cleanup()

	ctx := context.Background()

	u := &domain.User{
		ID:            "user-1",
		TenantID:      "tenant-1",
		Email:         "test@example.com",
		PasswordHash:  "hash",
		Locale:        "en",
		RecoveryCodes: nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Empty(t, got.RecoveryCodes)
}
