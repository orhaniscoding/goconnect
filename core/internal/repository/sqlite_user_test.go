package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteUserRepository_CreateAndFetch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "users.db")

	db, err := database.ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	// seed tenant
	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteUserRepository(db)

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
