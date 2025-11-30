package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteIPAMRepository_AllocateRelease(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "ipam.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteIPAMRepository(db)
	ip, err := repo.AllocateIP(context.Background(), "net-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Contains(t, ip, "10.0.0.")

	err = repo.ReleaseIP(context.Background(), "net-1", "user-1")
	require.NoError(t, err)
}
