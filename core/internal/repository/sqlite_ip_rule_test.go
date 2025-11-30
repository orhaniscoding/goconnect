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

func TestSQLiteIPRuleRepository_CreateGet(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "ip_rules.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteIPRuleRepository(db)
	rule := &domain.IPRule{
		TenantID:    "tenant-1",
		CIDR:        "10.0.0.0/24",
		Type:        domain.IPRuleTypeAllow,
		Description: "test",
		CreatedBy:   "user-1",
	}

	require.NoError(t, repo.Create(context.Background(), rule))

	got, err := repo.GetByID(context.Background(), rule.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.IPRuleTypeAllow, got.Type)
}
