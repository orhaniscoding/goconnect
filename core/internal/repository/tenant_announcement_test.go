package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test tenant announcement
func mkTenantAnnouncement(id, tenantID, title, content, authorID string, pinned bool) *domain.TenantAnnouncement {
	return &domain.TenantAnnouncement{
		ID:        id,
		TenantID:  tenantID,
		Title:     title,
		Content:   content,
		AuthorID:  authorID,
		IsPinned:  pinned,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewInMemoryTenantAnnouncementRepository(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.announcements)
	assert.Equal(t, 0, len(repo.announcements))
}

func TestTenantAnnouncementRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("", "tenant-1", "Welcome!", "Welcome to our server.", "user-1", false)

	err := repo.Create(ctx, ann)

	require.NoError(t, err)
	assert.NotEmpty(t, ann.ID)
	assert.Equal(t, 1, len(repo.announcements))
}

func TestTenantAnnouncementRepository_Create_WithExistingID(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-123", "tenant-1", "Title", "Content", "user-1", false)

	err := repo.Create(ctx, ann)

	require.NoError(t, err)
	assert.Equal(t, "ann-123", ann.ID)
}

func TestTenantAnnouncementRepository_Create_Pinned(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("", "tenant-1", "Important!", "This is important.", "user-1", true)

	err := repo.Create(ctx, ann)

	require.NoError(t, err)
	assert.True(t, ann.IsPinned)
}

func TestTenantAnnouncementRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-1", "tenant-1", "Title", "Content", "user-1", false)
	_ = repo.Create(ctx, ann)

	result, err := repo.GetByID(ctx, "ann-1")

	require.NoError(t, err)
	assert.Equal(t, "ann-1", result.ID)
	assert.Equal(t, "Title", result.Title)
	assert.Equal(t, "Content", result.Content)
}

func TestTenantAnnouncementRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantAnnouncementRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-1", "tenant-1", "Old Title", "Old Content", "user-1", false)
	_ = repo.Create(ctx, ann)

	ann.Title = "New Title"
	ann.Content = "New Content"
	err := repo.Update(ctx, ann)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "ann-1")
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "New Content", updated.Content)
}

func TestTenantAnnouncementRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("non-existent", "tenant-1", "Title", "Content", "user-1", false)

	err := repo.Update(ctx, ann)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantAnnouncementRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-1", "tenant-1", "Title", "Content", "user-1", false)
	_ = repo.Create(ctx, ann)

	err := repo.Delete(ctx, "ann-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.announcements))
}

func TestTenantAnnouncementRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantAnnouncementRepository_ListByTenant_All(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-1", "Ann 1", "Content 1", "user-1", false))
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-1", "Ann 2", "Content 2", "user-1", true))
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-2", "Ann 3", "Content 3", "user-1", false))

	results, cursor, err := repo.ListByTenant(ctx, "tenant-1", false, 10, "")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Empty(t, cursor)
}

func TestTenantAnnouncementRepository_ListByTenant_PinnedOnly(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-1", "Regular", "Content 1", "user-1", false))
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-1", "Pinned 1", "Content 2", "user-1", true))
	_ = repo.Create(ctx, mkTenantAnnouncement("", "tenant-1", "Pinned 2", "Content 3", "user-1", true))

	results, _, err := repo.ListByTenant(ctx, "tenant-1", true, 10, "")

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	for _, r := range results {
		assert.True(t, r.IsPinned)
	}
}

func TestTenantAnnouncementRepository_ListByTenant_Pagination(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		ann := mkTenantAnnouncement("", "tenant-1", "Ann "+string(rune('A'+i)), "Content", "user-1", false)
		_ = repo.Create(ctx, ann)
	}

	results, _, err := repo.ListByTenant(ctx, "tenant-1", false, 3, "")

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
	// Note: In-memory implementation doesn't use cursor
}

func TestTenantAnnouncementRepository_ListByTenant_Empty(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	results, cursor, err := repo.ListByTenant(ctx, "empty-tenant", false, 10, "")

	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
	assert.Empty(t, cursor)
}

func TestTenantAnnouncementRepository_SetPinned_Pin(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-1", "tenant-1", "Title", "Content", "user-1", false)
	_ = repo.Create(ctx, ann)

	err := repo.SetPinned(ctx, "ann-1", true)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "ann-1")
	assert.True(t, updated.IsPinned)
}

func TestTenantAnnouncementRepository_SetPinned_Unpin(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()
	ann := mkTenantAnnouncement("ann-1", "tenant-1", "Title", "Content", "user-1", true)
	_ = repo.Create(ctx, ann)

	err := repo.SetPinned(ctx, "ann-1", false)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "ann-1")
	assert.False(t, updated.IsPinned)
}

func TestTenantAnnouncementRepository_SetPinned_NotFound(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	err := repo.SetPinned(ctx, "non-existent", true)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantAnnouncementRepository_FullLifecycle(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	// Create
	ann := mkTenantAnnouncement("", "tenant-1", "Initial Title", "Initial Content", "user-1", false)
	err := repo.Create(ctx, ann)
	require.NoError(t, err)
	annID := ann.ID

	// Update
	ann.Title = "Updated Title"
	_ = repo.Update(ctx, ann)

	// Pin
	_ = repo.SetPinned(ctx, annID, true)

	// Verify
	updated, _ := repo.GetByID(ctx, annID)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.True(t, updated.IsPinned)

	// List
	results, _, _ := repo.ListByTenant(ctx, "tenant-1", true, 10, "")
	assert.Equal(t, 1, len(results))

	// Delete
	err = repo.Delete(ctx, annID)
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, annID)
	require.Error(t, err)
}

func TestTenantAnnouncementRepository_MultipleAnnouncements(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	// Create 10 announcements
	for i := 0; i < 10; i++ {
		ann := mkTenantAnnouncement("", "tenant-1", "Ann "+string(rune('0'+i)), "Content", "user-1", i%2 == 0)
		_ = repo.Create(ctx, ann)
	}

	// Verify total count
	all, _, _ := repo.ListByTenant(ctx, "tenant-1", false, 100, "")
	assert.Equal(t, 10, len(all))

	// Verify pinned count
	pinned, _, _ := repo.ListByTenant(ctx, "tenant-1", true, 100, "")
	assert.Equal(t, 5, len(pinned))
}

func TestTenantAnnouncementRepository_DeleteAllByTenant(t *testing.T) {
	repo := NewInMemoryTenantAnnouncementRepository()
	ctx := context.Background()

	// Create announcements in tenant-1
	ann1 := mkTenantAnnouncement("", "tenant-1", "Ann 1", "Content", "user-1", false)
	ann2 := mkTenantAnnouncement("", "tenant-1", "Ann 2", "Content", "user-1", false)
	ann3 := mkTenantAnnouncement("", "tenant-2", "Ann 3", "Content", "user-1", false)
	_ = repo.Create(ctx, ann1)
	_ = repo.Create(ctx, ann2)
	_ = repo.Create(ctx, ann3)

	err := repo.DeleteAllByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	results1, _, _ := repo.ListByTenant(ctx, "tenant-1", false, 100, "")
	assert.Empty(t, results1)
	results2, _, _ := repo.ListByTenant(ctx, "tenant-2", false, 100, "")
	assert.Len(t, results2, 1)
}
