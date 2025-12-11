package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test tenant chat message
func mkTenantChatMessage(id, tenantID, userID, content string) *domain.TenantChatMessage {
	return &domain.TenantChatMessage{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
		EditedAt:  nil,
		DeletedAt: nil,
	}
}

func TestNewInMemoryTenantChatRepository(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.messages)
	assert.Equal(t, 0, len(repo.messages))
}

func TestTenantChatRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("", "tenant-1", "user-1", "Hello world!")

	err := repo.Create(ctx, msg)

	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, 1, len(repo.messages))
}

func TestTenantChatRepository_Create_WithExistingID(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-123", "tenant-1", "user-1", "Hello!")

	err := repo.Create(ctx, msg)

	require.NoError(t, err)
	assert.Equal(t, "msg-123", msg.ID)
}

func TestTenantChatRepository_Create_LongMessage(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	longContent := string(make([]byte, 2000)) // 2KB message
	msg := mkTenantChatMessage("", "tenant-1", "user-1", longContent)

	err := repo.Create(ctx, msg)

	require.NoError(t, err)
	assert.Equal(t, len(longContent), len(msg.Content))
}

func TestTenantChatRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Test message")
	_ = repo.Create(ctx, msg)

	result, err := repo.GetByID(ctx, "msg-1")

	require.NoError(t, err)
	assert.Equal(t, "msg-1", result.ID)
	assert.Equal(t, "tenant-1", result.TenantID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "Test message", result.Content)
}

func TestTenantChatRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantChatRepository_GetByID_Deleted(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Deleted message")
	_ = repo.Create(ctx, msg)
	_ = repo.SoftDelete(ctx, "msg-1")

	result, err := repo.GetByID(ctx, "msg-1")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantChatRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Original content")
	_ = repo.Create(ctx, msg)

	msg.Content = "Edited content"
	err := repo.Update(ctx, msg)

	require.NoError(t, err)

	updated, _ := repo.GetByID(ctx, "msg-1")
	assert.Equal(t, "Edited content", updated.Content)
	assert.NotNil(t, updated.EditedAt)
}

func TestTenantChatRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("non-existent", "tenant-1", "user-1", "Content")

	err := repo.Update(ctx, msg)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantChatRepository_Update_Deleted(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Content")
	_ = repo.Create(ctx, msg)
	_ = repo.SoftDelete(ctx, "msg-1")

	msg.Content = "New content"
	err := repo.Update(ctx, msg)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantChatRepository_SoftDelete_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "To be deleted")
	_ = repo.Create(ctx, msg)

	err := repo.SoftDelete(ctx, "msg-1")

	require.NoError(t, err)

	// Message should still exist in storage but with DeletedAt set
	assert.Equal(t, 1, len(repo.messages))
	assert.NotNil(t, repo.messages["msg-1"].DeletedAt)
}

func TestTenantChatRepository_SoftDelete_NotFound(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	err := repo.SoftDelete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestTenantChatRepository_SoftDelete_AlreadyDeleted(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()
	msg := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Content")
	_ = repo.Create(ctx, msg)
	_ = repo.SoftDelete(ctx, "msg-1")

	// Note: In-memory implementation allows re-soft-delete (updates DeletedAt)
	err := repo.SoftDelete(ctx, "msg-1")

	require.NoError(t, err)
}

func TestTenantChatRepository_ListByTenant_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create messages with slight time delay to ensure order
	for i := 0; i < 3; i++ {
		msg := mkTenantChatMessage("", "tenant-1", "user-1", "Message "+string(rune('A'+i)))
		_ = repo.Create(ctx, msg)
	}

	results, err := repo.ListByTenant(ctx, "tenant-1", "", 10)

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestTenantChatRepository_ListByTenant_Pagination(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create 5 messages
	for i := 0; i < 5; i++ {
		msg := mkTenantChatMessage("", "tenant-1", "user-1", "Message "+string(rune('A'+i)))
		_ = repo.Create(ctx, msg)
	}

	// Get first 3
	results, err := repo.ListByTenant(ctx, "tenant-1", "", 3)

	require.NoError(t, err)
	assert.Equal(t, 3, len(results))
}

func TestTenantChatRepository_ListByTenant_ExcludesDeleted(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	msg1 := mkTenantChatMessage("msg-1", "tenant-1", "user-1", "Visible")
	_ = repo.Create(ctx, msg1)

	msg2 := mkTenantChatMessage("msg-2", "tenant-1", "user-1", "Deleted")
	_ = repo.Create(ctx, msg2)
	_ = repo.SoftDelete(ctx, "msg-2")

	msg3 := mkTenantChatMessage("msg-3", "tenant-1", "user-1", "Also visible")
	_ = repo.Create(ctx, msg3)

	results, err := repo.ListByTenant(ctx, "tenant-1", "", 10)

	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestTenantChatRepository_ListByTenant_Empty(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	results, err := repo.ListByTenant(ctx, "empty-tenant", "", 10)

	require.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestTenantChatRepository_ListByTenant_DifferentTenants(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "Tenant 1 message"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-2", "user-2", "Tenant 2 message"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-3", "Another tenant 1"))

	results1, _ := repo.ListByTenant(ctx, "tenant-1", "", 10)
	results2, _ := repo.ListByTenant(ctx, "tenant-2", "", 10)

	assert.Equal(t, 2, len(results1))
	assert.Equal(t, 1, len(results2))
}

func TestTenantChatRepository_DeleteOlderThan_Success(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create old message
	oldMsg := mkTenantChatMessage("old-msg", "tenant-1", "user-1", "Old message")
	oldMsg.CreatedAt = time.Now().Add(-48 * time.Hour) // 2 days ago
	repo.messages[oldMsg.ID] = oldMsg

	// Create recent message
	newMsg := mkTenantChatMessage("new-msg", "tenant-1", "user-1", "New message")
	_ = repo.Create(ctx, newMsg)

	// Delete messages older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	count, err := repo.DeleteOlderThan(ctx, "tenant-1", cutoff)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(repo.messages))

	// Verify correct message remains
	_, err = repo.GetByID(ctx, "new-msg")
	require.NoError(t, err)
}

func TestTenantChatRepository_DeleteOlderThan_NoOldMessages(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "Recent 1"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "Recent 2"))

	cutoff := time.Now().Add(-24 * time.Hour)
	count, err := repo.DeleteOlderThan(ctx, "tenant-1", cutoff)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 2, len(repo.messages))
}

func TestTenantChatRepository_DeleteOlderThan_OnlyAffectsTenant(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create old message in tenant-1
	oldMsg1 := mkTenantChatMessage("old-1", "tenant-1", "user-1", "Old")
	oldMsg1.CreatedAt = time.Now().Add(-48 * time.Hour)
	repo.messages[oldMsg1.ID] = oldMsg1

	// Create old message in tenant-2
	oldMsg2 := mkTenantChatMessage("old-2", "tenant-2", "user-2", "Old")
	oldMsg2.CreatedAt = time.Now().Add(-48 * time.Hour)
	repo.messages[oldMsg2.ID] = oldMsg2

	// Delete only tenant-1 old messages
	cutoff := time.Now().Add(-24 * time.Hour)
	count, err := repo.DeleteOlderThan(ctx, "tenant-1", cutoff)

	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify tenant-2 message still exists
	_, exists := repo.messages["old-2"]
	assert.True(t, exists)
}

func TestTenantChatRepository_FullLifecycle(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create
	msg := mkTenantChatMessage("", "tenant-1", "user-1", "Initial message")
	err := repo.Create(ctx, msg)
	require.NoError(t, err)
	msgID := msg.ID

	// Update (edit)
	msg.Content = "Edited message"
	_ = repo.Update(ctx, msg)

	// Verify
	updated, _ := repo.GetByID(ctx, msgID)
	assert.Equal(t, "Edited message", updated.Content)
	assert.NotNil(t, updated.EditedAt)

	// List
	results, _ := repo.ListByTenant(ctx, "tenant-1", "", 10)
	assert.Equal(t, 1, len(results))

	// Soft delete
	_ = repo.SoftDelete(ctx, msgID)

	// Verify not accessible
	_, err = repo.GetByID(ctx, msgID)
	require.Error(t, err)

	// List should be empty
	results, _ = repo.ListByTenant(ctx, "tenant-1", "", 10)
	assert.Equal(t, 0, len(results))
}

func TestTenantChatRepository_MultipleSenders(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Simulate conversation
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "Hello!"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-2", "Hi there!"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "How are you?"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-3", "Good thanks!"))

	results, err := repo.ListByTenant(ctx, "tenant-1", "", 10)

	require.NoError(t, err)
	assert.Equal(t, 4, len(results))

	// Verify different users
	users := make(map[string]int)
	for _, r := range results {
		users[r.UserID]++
	}
	assert.Equal(t, 2, users["user-1"])
	assert.Equal(t, 1, users["user-2"])
	assert.Equal(t, 1, users["user-3"])
}

func TestTenantChatRepository_DeleteAllByTenant(t *testing.T) {
	repo := NewInMemoryTenantChatRepository()
	ctx := context.Background()

	// Create messages in tenant-1
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-1", "Msg 1"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-1", "user-2", "Msg 2"))
	_ = repo.Create(ctx, mkTenantChatMessage("", "tenant-2", "user-3", "Msg 3"))

	err := repo.DeleteAllByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	results1, _ := repo.ListByTenant(ctx, "tenant-1", "", 100)
	assert.Empty(t, results1)
	results2, _ := repo.ListByTenant(ctx, "tenant-2", "", 100)
	assert.Len(t, results2, 1)
}
