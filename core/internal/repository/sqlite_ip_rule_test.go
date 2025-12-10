package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIPRuleTestDB(t *testing.T) (*sql.DB, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','test@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	return db, func() { db.Close() }
}

func TestSQLiteIPRuleRepository_CreateGet(t *testing.T) {
	db, cleanup := setupIPRuleTestDB(t)
	defer cleanup()

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

	// Get non-existent
	_, err = repo.GetByID(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestSQLiteIPRuleRepository_List(t *testing.T) {
	db, cleanup := setupIPRuleTestDB(t)
	defer cleanup()

	repo := NewSQLiteIPRuleRepository(db)
	ctx := context.Background()

	// Create multiple rules
	for i := 0; i < 3; i++ {
		rule := &domain.IPRule{
			ID:          "rule-" + string(rune('a'+i)),
			TenantID:    "tenant-1",
			CIDR:        "10.0." + string(rune('0'+i)) + ".0/24",
			Type:        domain.IPRuleTypeAllow,
			Description: "Rule " + string(rune('A'+i)),
			CreatedBy:   "user-1",
			CreatedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, rule))
	}

	// List all
	rules, err := repo.List(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Len(t, rules, 3)

	// ListByTenant should return same results
	rules2, err := repo.ListByTenant(ctx, "tenant-1")
	require.NoError(t, err)
	assert.Len(t, rules2, 3)
}

func TestSQLiteIPRuleRepository_Delete(t *testing.T) {
	db, cleanup := setupIPRuleTestDB(t)
	defer cleanup()

	repo := NewSQLiteIPRuleRepository(db)
	ctx := context.Background()

	// Create rule
	rule := &domain.IPRule{
		ID:          "rule-1",
		TenantID:    "tenant-1",
		CIDR:        "192.168.1.0/24",
		Type:        domain.IPRuleTypeAllow,
		Description: "Allow local network",
		CreatedBy:   "user-1",
		CreatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, rule))

	// Delete
	err := repo.Delete(ctx, "rule-1")
	require.NoError(t, err)

	// Should not exist
	_, err = repo.GetByID(ctx, "rule-1")
	require.Error(t, err)

	// Delete non-existent
	err = repo.Delete(ctx, "nonexistent")
	require.Error(t, err)
}

func TestSQLiteIPRuleRepository_CheckIP(t *testing.T) {
	db, cleanup := setupIPRuleTestDB(t)
	defer cleanup()

	repo := NewSQLiteIPRuleRepository(db)
	ctx := context.Background()

	// Create rule
	rule := &domain.IPRule{
		ID:          "rule-1",
		TenantID:    "tenant-1",
		CIDR:        "192.168.1.0/24",
		Type:        domain.IPRuleTypeAllow,
		Description: "Allow local network",
		CreatedBy:   "user-1",
		CreatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(ctx, rule))

	// Check matching IP
	matchedRule, err := repo.CheckIP(ctx, "tenant-1", "192.168.1.50")
	require.NoError(t, err)
	assert.Equal(t, "rule-1", matchedRule.ID)

	// Check non-matching IP
	_, err = repo.CheckIP(ctx, "tenant-1", "10.0.0.1")
	require.Error(t, err)
}

func TestSQLiteIPRuleRepository_DeleteExpired(t *testing.T) {
	db, cleanup := setupIPRuleTestDB(t)
	defer cleanup()

	repo := NewSQLiteIPRuleRepository(db)
	ctx := context.Background()

	// Create expired rule
	pastTime := time.Now().Add(-24 * time.Hour)
	expiredRule := &domain.IPRule{
		ID:          "rule-expired",
		TenantID:    "tenant-1",
		CIDR:        "10.0.0.0/8",
		Type:        domain.IPRuleTypeDeny,
		Description: "Expired rule",
		CreatedBy:   "user-1",
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		ExpiresAt:   &pastTime,
	}
	require.NoError(t, repo.Create(ctx, expiredRule))

	// Create non-expired rule
	futureTime := time.Now().Add(24 * time.Hour)
	activeRule := &domain.IPRule{
		ID:          "rule-active",
		TenantID:    "tenant-1",
		CIDR:        "172.16.0.0/12",
		Type:        domain.IPRuleTypeAllow,
		Description: "Active rule",
		CreatedBy:   "user-1",
		CreatedAt:   time.Now(),
		ExpiresAt:   &futureTime,
	}
	require.NoError(t, repo.Create(ctx, activeRule))

	// Delete expired
	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Expired rule should be gone
	_, err = repo.GetByID(ctx, "rule-expired")
	require.Error(t, err)

	// Active rule should still exist
	_, err = repo.GetByID(ctx, "rule-active")
	require.NoError(t, err)
}
