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

// setupSQLiteGDPRTest creates a test SQLite database with the deletion_requests table
func setupSQLiteGDPRTest(t *testing.T) (*sql.DB, *SQLiteDeletionRequestRepository, func()) {
	dir := t.TempDir()
	db, err := database.ConnectSQLite(filepath.Join(dir, "gdpr.db"))
	require.NoError(t, err)
	require.NoError(t, database.RunSQLiteMigrations(db, filepath.Join("..", "..", "migrations_sqlite")))

	// Create deletion_requests table (not in standard SQLite migrations)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS deletion_requests (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			error TEXT
		)
	`)
	require.NoError(t, err)

	// Seed tenant and user for foreign key references
	_, err = db.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES ('tenant-1','t1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-1','tenant-1','u@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, tenant_id, email, password_hash, locale, is_admin, is_moderator, created_at, updated_at) VALUES ('user-2','tenant-1','u2@example.com','hash','en',0,0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	repo := NewSQLiteDeletionRequestRepository(db)
	return db, repo, func() { db.Close() }
}

func TestSQLiteDeletionRequestRepository_Create(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}

	err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Verify it was created
	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, "del-1", got.ID)
	assert.Equal(t, "user-1", got.UserID)
	assert.Equal(t, domain.DeletionRequestStatusPending, got.Status)
}

func TestSQLiteDeletionRequestRepository_CreateWithCompletedAt(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)
	completedAt := now.Add(time.Hour)

	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusCompleted,
		RequestedAt: now,
		CompletedAt: &completedAt,
	}

	err := repo.Create(ctx, req)
	require.NoError(t, err)

	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.NotNil(t, got.CompletedAt)
}

func TestSQLiteDeletionRequestRepository_CreateWithError(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusFailed,
		RequestedAt: now,
		Error:       "something went wrong",
	}

	err := repo.Create(ctx, req)
	require.NoError(t, err)

	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, "something went wrong", got.Error)
}

func TestSQLiteDeletionRequestRepository_Get(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Test not found
	_, err := repo.Get(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Create a request
	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req))

	// Get it back
	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, "del-1", got.ID)
	assert.Equal(t, "user-1", got.UserID)
	assert.Equal(t, domain.DeletionRequestStatusPending, got.Status)
	assert.Nil(t, got.CompletedAt)
	assert.Empty(t, got.Error)
}

func TestSQLiteDeletionRequestRepository_GetByUserID(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Test user with no requests
	got, err := repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	assert.Nil(t, got)

	// Create multiple requests for same user (older one first)
	req1 := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusCompleted,
		RequestedAt: now.Add(-time.Hour),
	}
	require.NoError(t, repo.Create(ctx, req1))

	req2 := &domain.DeletionRequest{
		ID:          "del-2",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req2))

	// Should return most recent one
	got, err = repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "del-2", got.ID)
}

func TestSQLiteDeletionRequestRepository_GetByUserID_DifferentUsers(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create requests for different users
	req1 := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req1))

	req2 := &domain.DeletionRequest{
		ID:          "del-2",
		UserID:      "user-2",
		Status:      domain.DeletionRequestStatusCompleted,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req2))

	// Get for user-1
	got, err := repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "del-1", got.ID)
	assert.Equal(t, domain.DeletionRequestStatusPending, got.Status)

	// Get for user-2
	got, err = repo.GetByUserID(ctx, "user-2")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "del-2", got.ID)
	assert.Equal(t, domain.DeletionRequestStatusCompleted, got.Status)
}

func TestSQLiteDeletionRequestRepository_ListPending(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Initially empty
	pending, err := repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Empty(t, pending)

	// Create multiple requests with different statuses
	req1 := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req1))

	req2 := &domain.DeletionRequest{
		ID:          "del-2",
		UserID:      "user-2",
		Status:      domain.DeletionRequestStatusCompleted,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req2))

	req3 := &domain.DeletionRequest{
		ID:          "del-3",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now.Add(time.Minute),
	}
	require.NoError(t, repo.Create(ctx, req3))

	// Should only return pending requests
	pending, err = repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 2)

	// Verify all are pending
	for _, p := range pending {
		assert.Equal(t, domain.DeletionRequestStatusPending, p.Status)
	}
}

func TestSQLiteDeletionRequestRepository_ListPending_ExcludesOtherStatuses(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create requests with all different statuses
	statuses := []domain.DeletionRequestStatus{
		domain.DeletionRequestStatusPending,
		domain.DeletionRequestStatusProcessing,
		domain.DeletionRequestStatusCompleted,
		domain.DeletionRequestStatusFailed,
	}

	for i, status := range statuses {
		req := &domain.DeletionRequest{
			ID:          "del-" + string(rune('1'+i)),
			UserID:      "user-1",
			Status:      status,
			RequestedAt: now.Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, repo.Create(ctx, req))
	}

	// Should only return pending requests
	pending, err := repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, domain.DeletionRequestStatusPending, pending[0].Status)
}

func TestSQLiteDeletionRequestRepository_Update(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create a request
	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req))

	// Update status
	req.Status = domain.DeletionRequestStatusProcessing
	err := repo.Update(ctx, req)
	require.NoError(t, err)

	// Verify update
	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, domain.DeletionRequestStatusProcessing, got.Status)
}

func TestSQLiteDeletionRequestRepository_UpdateWithCompletedAt(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create a pending request
	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req))

	// Complete it
	completedAt := now.Add(time.Hour)
	req.Status = domain.DeletionRequestStatusCompleted
	req.CompletedAt = &completedAt
	err := repo.Update(ctx, req)
	require.NoError(t, err)

	// Verify
	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, domain.DeletionRequestStatusCompleted, got.Status)
	require.NotNil(t, got.CompletedAt)
}

func TestSQLiteDeletionRequestRepository_UpdateWithError(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// Create a pending request
	req := &domain.DeletionRequest{
		ID:          "del-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req))

	// Fail it
	req.Status = domain.DeletionRequestStatusFailed
	req.Error = "database connection lost"
	err := repo.Update(ctx, req)
	require.NoError(t, err)

	// Verify
	got, err := repo.Get(ctx, "del-1")
	require.NoError(t, err)
	assert.Equal(t, domain.DeletionRequestStatusFailed, got.Status)
	assert.Equal(t, "database connection lost", got.Error)
}

func TestSQLiteDeletionRequestRepository_FullWorkflow(t *testing.T) {
	_, repo, cleanup := setupSQLiteGDPRTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	// 1. Create pending request
	req := &domain.DeletionRequest{
		ID:          "del-workflow",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: now,
	}
	require.NoError(t, repo.Create(ctx, req))

	// 2. Verify it appears in pending list
	pending, err := repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 1)

	// 3. Update to processing
	req.Status = domain.DeletionRequestStatusProcessing
	require.NoError(t, repo.Update(ctx, req))

	// 4. Verify it no longer appears in pending list
	pending, err = repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Empty(t, pending)

	// 5. Complete the request
	completedAt := now.Add(10 * time.Minute)
	req.Status = domain.DeletionRequestStatusCompleted
	req.CompletedAt = &completedAt
	require.NoError(t, repo.Update(ctx, req))

	// 6. Verify final state
	got, err := repo.Get(ctx, "del-workflow")
	require.NoError(t, err)
	assert.Equal(t, domain.DeletionRequestStatusCompleted, got.Status)
	require.NotNil(t, got.CompletedAt)

	// 7. Verify by user ID
	byUser, err := repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, byUser)
	assert.Equal(t, "del-workflow", byUser.ID)
}
