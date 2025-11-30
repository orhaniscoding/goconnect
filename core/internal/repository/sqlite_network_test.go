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

func TestSQLiteNetworkRepository_CreateAndGet(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "networks.db")
	db, err := database.ConnectSQLite(dbPath)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteNetworkRepository(db)
	n := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "TestNet",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), n))

	got, err := repo.GetByID(context.Background(), "net-1")
	require.NoError(t, err)
	assert.Equal(t, n.Name, got.Name)
	assert.Equal(t, n.CIDR, got.CIDR)
	assert.Equal(t, domain.NetworkVisibilityPublic, got.Visibility)
}
