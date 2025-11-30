package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test chat message
func mkChatMessage(id, scope, userID, body string) *domain.ChatMessage {
	now := time.Now()
	return &domain.ChatMessage{
		ID:        id,
		Scope:     scope,
		TenantID:  "tenant-1",
		UserID:    userID,
		Body:      body,
		Redacted:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Helper function to create a test chat message edit
func mkChatMessageEdit(id, messageID, prevBody, newBody, editorID string) *domain.ChatMessageEdit {
	return &domain.ChatMessageEdit{
		ID:        id,
		MessageID: messageID,
		PrevBody:  prevBody,
		NewBody:   newBody,
		EditorID:  editorID,
		EditedAt:  time.Now(),
	}
}

func TestNewInMemoryChatRepository(t *testing.T) {
	repo := NewInMemoryChatRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.messages)
	assert.NotNil(t, repo.edits)
	assert.Equal(t, 0, len(repo.messages))
	assert.Equal(t, 0, len(repo.edits))
}

func TestChatRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("msg-1", "host", "user-1", "Hello World")

	err := repo.Create(ctx, msg)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.messages))
}

func TestChatRepository_Create_GeneratesID(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("", "host", "user-1", "Hello")

	err := repo.Create(ctx, msg)

	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
}

func TestChatRepository_Create_MultipleMessages(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	messages := []*domain.ChatMessage{
		mkChatMessage("msg-1", "host", "user-1", "Message 1"),
		mkChatMessage("msg-2", "network:net-1", "user-2", "Message 2"),
		mkChatMessage("msg-3", "host", "user-1", "Message 3"),
	}

	for _, msg := range messages {
		err := repo.Create(ctx, msg)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.messages))
}

func TestChatRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	original := mkChatMessage("msg-1", "host", "user-1", "Test message")
	repo.Create(ctx, original)

	retrieved, err := repo.GetByID(ctx, "msg-1")

	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Body, retrieved.Body)
	assert.Equal(t, original.UserID, retrieved.UserID)
	assert.Equal(t, original.Scope, retrieved.Scope)
}

func TestChatRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	retrieved, err := repo.GetByID(ctx, "non-existent")

	require.Error(t, err)
	assert.Nil(t, retrieved)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestChatRepository_List_AllMessages(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	// Create multiple messages
	for i := 1; i <= 3; i++ {
		msg := mkChatMessage("", "host", "user-1", "Message")
		repo.Create(ctx, msg)
	}

	filter := domain.ChatMessageFilter{Limit: 50}
	result, cursor, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Empty(t, cursor)
}

func TestChatRepository_List_ByScope(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	repo.Create(ctx, mkChatMessage("msg-1", "host", "user-1", "Host message"))
	repo.Create(ctx, mkChatMessage("msg-2", "network:net-1", "user-2", "Network message"))
	repo.Create(ctx, mkChatMessage("msg-3", "host", "user-3", "Another host"))

	filter := domain.ChatMessageFilter{Scope: "host", Limit: 50}
	result, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	for _, msg := range result {
		assert.Equal(t, "host", msg.Scope)
	}
}

func TestChatRepository_List_ByUserID(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	repo.Create(ctx, mkChatMessage("msg-1", "host", "user-1", "Message 1"))
	repo.Create(ctx, mkChatMessage("msg-2", "host", "user-1", "Message 2"))
	repo.Create(ctx, mkChatMessage("msg-3", "host", "user-2", "Message 3"))

	filter := domain.ChatMessageFilter{UserID: "user-1", Limit: 50}
	result, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	for _, msg := range result {
		assert.Equal(t, "user-1", msg.UserID)
	}
}

func TestChatRepository_Update_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("msg-1", "host", "user-1", "Original message")
	repo.Create(ctx, msg)

	// Create updated message
	updatedMsg := mkChatMessage("msg-1", "host", "user-1", "Updated message")
	updatedMsg.Redacted = true

	err := repo.Update(ctx, updatedMsg)

	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, "msg-1")
	assert.Equal(t, "Updated message", retrieved.Body)
	assert.True(t, retrieved.Redacted)
}

func TestChatRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("non-existent", "host", "user-1", "Test")

	err := repo.Update(ctx, msg)

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestChatRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("msg-1", "host", "user-1", "Test message")
	repo.Create(ctx, msg)

	err := repo.Delete(ctx, "msg-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.messages))

	_, err = repo.GetByID(ctx, "msg-1")
	assert.Error(t, err)
}

func TestChatRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestChatRepository_SoftDelete_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()
	msg := mkChatMessage("msg-1", "host", "user-1", "Test message")
	repo.Create(ctx, msg)

	err := repo.SoftDelete(ctx, "msg-1")

	require.NoError(t, err)

	// Message should still exist
	retrieved, err := repo.GetByID(ctx, "msg-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.DeletedAt)
}

func TestChatRepository_SoftDelete_NotFound(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	err := repo.SoftDelete(ctx, "non-existent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestChatRepository_AddEdit_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	edit := mkChatMessageEdit("edit-1", "msg-1", "Old body", "New body", "user-1")

	err := repo.AddEdit(ctx, edit)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.edits))
	assert.Equal(t, 1, len(repo.edits["msg-1"]))
}

func TestChatRepository_AddEdit_MultipleEdits(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	edits := []*domain.ChatMessageEdit{
		mkChatMessageEdit("edit-1", "msg-1", "V1", "V2", "user-1"),
		mkChatMessageEdit("edit-2", "msg-1", "V2", "V3", "user-1"),
		mkChatMessageEdit("edit-3", "msg-1", "V3", "V4", "user-1"),
	}

	for _, edit := range edits {
		err := repo.AddEdit(ctx, edit)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.edits["msg-1"]))
}

func TestChatRepository_GetEdits_Success(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	edit1 := mkChatMessageEdit("edit-1", "msg-1", "V1", "V2", "user-1")
	edit2 := mkChatMessageEdit("edit-2", "msg-1", "V2", "V3", "user-1")

	repo.AddEdit(ctx, edit1)
	repo.AddEdit(ctx, edit2)

	edits, err := repo.GetEdits(ctx, "msg-1")

	require.NoError(t, err)
	assert.Len(t, edits, 2)
	assert.Equal(t, "V1", edits[0].PrevBody)
	assert.Equal(t, "V2", edits[0].NewBody)
}

func TestChatRepository_GetEdits_NoEdits(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	edits, err := repo.GetEdits(ctx, "non-existent")

	require.NoError(t, err)
	assert.Empty(t, edits)
}

func TestChatRepository_RedactedMessage(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	msg := mkChatMessage("msg-1", "host", "user-1", "[REDACTED]")
	msg.Redacted = true

	repo.Create(ctx, msg)

	retrieved, err := repo.GetByID(ctx, "msg-1")
	require.NoError(t, err)
	assert.True(t, retrieved.Redacted)
	assert.Equal(t, "[REDACTED]", retrieved.Body)
}

func TestChatRepository_MessageWithAttachments(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	msg := mkChatMessage("msg-1", "host", "user-1", "Message with files")
	msg.Attachments = []string{"file1.pdf", "image.png"}

	repo.Create(ctx, msg)

	retrieved, err := repo.GetByID(ctx, "msg-1")
	require.NoError(t, err)
	assert.Len(t, retrieved.Attachments, 2)
	assert.Contains(t, retrieved.Attachments, "file1.pdf")
}

func TestChatRepository_FullCRUDCycle(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	// Create
	msg := mkChatMessage("msg-1", "host", "user-1", "Original message")
	err := repo.Create(ctx, msg)
	require.NoError(t, err)

	// Read
	retrieved, err := repo.GetByID(ctx, "msg-1")
	require.NoError(t, err)
	assert.Equal(t, "Original message", retrieved.Body)

	// Update
	updatedMsg := mkChatMessage("msg-1", "host", "user-1", "Updated message")
	err = repo.Update(ctx, updatedMsg)
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(ctx, "msg-1")
	assert.Equal(t, "Updated message", retrieved.Body)

	// Soft Delete
	err = repo.SoftDelete(ctx, "msg-1")
	require.NoError(t, err)

	retrieved, _ = repo.GetByID(ctx, "msg-1")
	assert.NotNil(t, retrieved.DeletedAt)

	// Hard Delete
	err = repo.Delete(ctx, "msg-1")
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, "msg-1")
	assert.Error(t, err)
}

func TestChatRepository_EditHistoryTracking(t *testing.T) {
	repo := NewInMemoryChatRepository()
	ctx := context.Background()

	// Create message
	msg := mkChatMessage("msg-1", "host", "user-1", "Version 1")
	repo.Create(ctx, msg)

	// First edit
	edit1 := mkChatMessageEdit("edit-1", "msg-1", "Version 1", "Version 2", "user-1")
	repo.AddEdit(ctx, edit1)

	// Second edit
	edit2 := mkChatMessageEdit("edit-2", "msg-1", "Version 2", "Version 3", "user-1")
	repo.AddEdit(ctx, edit2)

	// Retrieve edit history
	edits, err := repo.GetEdits(ctx, "msg-1")
	require.NoError(t, err)
	assert.Len(t, edits, 2)

	// Verify chronological order
	assert.Equal(t, "Version 1", edits[0].PrevBody)
	assert.Equal(t, "Version 2", edits[0].NewBody)
	assert.Equal(t, "Version 2", edits[1].PrevBody)
	assert.Equal(t, "Version 3", edits[1].NewBody)
}
