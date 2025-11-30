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

func TestSQLiteInviteTokenRepository_CreateGet(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "invite_tokens.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteInviteTokenRepository(db)
	it := &domain.InviteToken{
		NetworkID: "net-1",
		Token:     "TOK123",
		UsesMax:   2,
		UsesLeft:  2,
		CreatedBy: "user-1",
	}

	require.NoError(t, repo.Create(context.Background(), it))

	got, err := repo.GetByToken(context.Background(), "TOK123")
	require.NoError(t, err)
	assert.Equal(t, it.NetworkID, got.NetworkID)

	require.NoError(t, repo.IncrementUseCount(context.Background(), got.ID))
	got2, err := repo.GetByID(context.Background(), got.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got2.UsesLeft)

	_, err = repo.DeleteExpired(context.Background())
	assert.NoError(t, err)
}
