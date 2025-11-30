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

func TestSQLiteTenantMemberRepository_CreateGet(t *testing.T) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "tenant_members.db"))
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, join_policy, created_by, created_at, updated_at) VALUES ('net-1','tenant-1','n1','10.0.0.0/24','public','open','user-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteTenantMemberRepository(db)
	member := &domain.TenantMember{
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Role:      domain.TenantRoleMember,
		Nickname:  "nick",
		JoinedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(context.Background(), member))

	got, err := repo.GetByUserAndTenant(context.Background(), "user-1", "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, member.Role, got.Role)
	assert.Equal(t, "nick", got.Nickname)
}
