package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteTenantInviteRepository_CreateGet(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "tenant_invites.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteTenantInviteRepository(db)
	invite := &domain.TenantInvite{
		TenantID:  "tenant-1",
		Code:      "ABC123",
		MaxUses:   5,
		UseCount:  0,
		ExpiresAt: nil,
		CreatedBy: "user-1",
	}

	require.NoError(t, repo.Create(context.Background(), invite))

	got, err := repo.GetByCode(context.Background(), "ABC123")
	require.NoError(t, err)
	assert.Equal(t, invite.TenantID, got.TenantID)
	assert.Equal(t, "ABC123", got.Code)

	require.NoError(t, repo.IncrementUseCount(context.Background(), got.ID))
	got2, err := repo.GetByID(context.Background(), got.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got2.UseCount)
	require.NoError(t, repo.Revoke(context.Background(), got.ID))
	got3, err := repo.GetByID(context.Background(), got.ID)
	require.NoError(t, err)
	assert.NotNil(t, got3.RevokedAt)
	_, err = repo.DeleteExpired(context.Background())
	require.NoError(t, err)
}
