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

// setupMembershipTestDB creates a test database with required tenant, user, and network
func setupMembershipTestDB(t *testing.T) (*sql.DB, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	ctx := context.Background()

	// Create tenant FIRST (users have FK to tenants)
	tenantRepo := NewSQLiteTenantRepository(db)
	tenant := &domain.Tenant{
		ID:         "tenant-1",
		Name:       "Test Tenant",
		OwnerID:    "user-1", // Will be created next
		MaxMembers: 100,
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant), "failed to create tenant")

	// Now create users with TenantID set
	userRepo := NewSQLiteUserRepository(db)
	user := &domain.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		PasswordHash: "hashedpass",
		Locale:       "en",
	}
	require.NoError(t, userRepo.Create(ctx, user), "failed to create user-1")

	// Create another user
	user2 := &domain.User{
		ID:           "user-2",
		TenantID:     "tenant-1",
		Email:        "test2@example.com",
		PasswordHash: "hashedpass",
		Locale:       "en",
	}
	require.NoError(t, userRepo.Create(ctx, user2), "failed to create user-2")

	// Create more users for list tests
	for i := 0; i < 5; i++ {
		u := &domain.User{
			ID:           "user-" + string(rune('a'+i)),
			TenantID:     "tenant-1",
			Email:        "user" + string(rune('a'+i)) + "@example.com",
			PasswordHash: "hashedpass",
			Locale:       "en",
		}
		_ = userRepo.Create(ctx, u) // Ignore errors for duplicates
	}

	// Create network
	networkRepo := NewSQLiteNetworkRepository(db)
	network := &domain.Network{
		ID:         "net-1",
		TenantID:   "tenant-1",
		Name:       "Test Network",
		CIDR:       "10.0.0.0/24",
		Visibility: domain.NetworkVisibilityPrivate,
		JoinPolicy: domain.JoinPolicyApproval,
		CreatedBy:  "user-1",
	}
	require.NoError(t, networkRepo.Create(ctx, network))

	return db, func() { db.Close() }
}

func TestSQLiteMembershipRepository_UpsertApproved(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteMembershipRepository(db)
	ctx := context.Background()

	// Insert new membership
	joinedAt := time.Now()
	m, err := repo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleMember, joinedAt)
	require.NoError(t, err)
	assert.NotEmpty(t, m.ID)
	assert.Equal(t, "net-1", m.NetworkID)
	assert.Equal(t, "user-1", m.UserID)
	assert.Equal(t, domain.RoleMember, m.Role)
	assert.Equal(t, domain.StatusApproved, m.Status)

	// Update existing membership
	m2, err := repo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleAdmin, joinedAt)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleAdmin, m2.Role)
}

func TestSQLiteMembershipRepository_Get(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteMembershipRepository(db)
	ctx := context.Background()

	// Create membership
	_, err := repo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleMember, time.Now())
	require.NoError(t, err)

	// Get existing
	m, err := repo.Get(ctx, "net-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "net-1", m.NetworkID)
	assert.Equal(t, "user-1", m.UserID)

	// Get non-existent
	_, err = repo.Get(ctx, "net-1", "nonexistent")
	require.Error(t, err)
}

func TestSQLiteMembershipRepository_SetStatus(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteMembershipRepository(db)
	ctx := context.Background()

	// Create membership
	_, err := repo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleMember, time.Now())
	require.NoError(t, err)

	// Set status
	err = repo.SetStatus(ctx, "net-1", "user-1", domain.StatusBanned)
	require.NoError(t, err)

	// Verify
	m, err := repo.Get(ctx, "net-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusBanned, m.Status)
}

func TestSQLiteMembershipRepository_List(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteMembershipRepository(db)
	ctx := context.Background()

	// Create multiple memberships
	for i := 0; i < 5; i++ {
		userID := "user-" + string(rune('a'+i))
		_, err := repo.UpsertApproved(ctx, "net-1", userID, domain.RoleMember, time.Now())
		require.NoError(t, err)
	}

	// List all
	list, next, err := repo.List(ctx, "net-1", "", 10, "")
	require.NoError(t, err)
	assert.Len(t, list, 5)
	assert.Empty(t, next)

	// List with limit (pagination)
	list, next, err = repo.List(ctx, "net-1", "", 2, "")
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NotEmpty(t, next)

	// List with status filter
	list, _, err = repo.List(ctx, "net-1", string(domain.StatusApproved), 10, "")
	require.NoError(t, err)
	assert.Len(t, list, 5)
}

func TestSQLiteMembershipRepository_Remove(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteMembershipRepository(db)
	ctx := context.Background()

	// Create membership
	_, err := repo.UpsertApproved(ctx, "net-1", "user-1", domain.RoleMember, time.Now())
	require.NoError(t, err)

	// Remove
	err = repo.Remove(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// Should not be found
	_, err = repo.Get(ctx, "net-1", "user-1")
	require.Error(t, err)

	// Remove non-existent
	err = repo.Remove(ctx, "net-1", "nonexistent")
	require.Error(t, err)
}

func TestSQLiteJoinRequestRepository_CreatePending(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteJoinRequestRepository(db)
	ctx := context.Background()

	// Create pending
	jr, err := repo.CreatePending(ctx, "net-1", "user-1")
	require.NoError(t, err)
	assert.NotEmpty(t, jr.ID)
	assert.Equal(t, "net-1", jr.NetworkID)
	assert.Equal(t, "user-1", jr.UserID)
	assert.Equal(t, "pending", jr.Status)

	// Duplicate should error
	_, err = repo.CreatePending(ctx, "net-1", "user-1")
	require.Error(t, err)
}

func TestSQLiteJoinRequestRepository_GetPending(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteJoinRequestRepository(db)
	ctx := context.Background()

	// Create pending
	_, err := repo.CreatePending(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// Get pending
	jr, err := repo.GetPending(ctx, "net-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "pending", jr.Status)

	// Get non-existent
	_, err = repo.GetPending(ctx, "net-1", "nonexistent")
	require.Error(t, err)
}

func TestSQLiteJoinRequestRepository_Decide(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteJoinRequestRepository(db)
	ctx := context.Background()

	// Create pending
	jr, err := repo.CreatePending(ctx, "net-1", "user-1")
	require.NoError(t, err)

	// Approve
	err = repo.Decide(ctx, jr.ID, true)
	require.NoError(t, err)

	// Should no longer be pending
	_, err = repo.GetPending(ctx, "net-1", "user-1")
	require.Error(t, err)

	// Create another and deny
	jr2, err := repo.CreatePending(ctx, "net-1", "user-2")
	require.NoError(t, err)

	err = repo.Decide(ctx, jr2.ID, false)
	require.NoError(t, err)
}

func TestSQLiteJoinRequestRepository_ListPending(t *testing.T) {
	db, cleanup := setupMembershipTestDB(t)
	defer cleanup()

	repo := NewSQLiteJoinRequestRepository(db)
	ctx := context.Background()

	// Create multiple pending requests
	for i := 0; i < 3; i++ {
		userID := "user-" + string(rune('a'+i))
		_, err := repo.CreatePending(ctx, "net-1", userID)
		require.NoError(t, err)
	}

	// List pending
	list, err := repo.ListPending(ctx, "net-1")
	require.NoError(t, err)
	assert.Len(t, list, 3)
}
