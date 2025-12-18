package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatService_SendMessage(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
	}
	userRepo.Create(ctx, user)

	t.Run("Success", func(t *testing.T) {
		msg, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Hello world!", nil, "")

		require.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, "user-123", msg.UserID)
		assert.Equal(t, "tenant-1", msg.TenantID)
		assert.Equal(t, "host", msg.Scope)
		assert.Equal(t, "Hello world!", msg.Body)
		assert.False(t, msg.Redacted)
		assert.Nil(t, msg.DeletedAt)
	})

	t.Run("User not found", func(t *testing.T) {
		_, err := service.SendMessage(ctx, "non-existent", "tenant-1", "host", "Hello", nil, "")

		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrUserNotFound, domainErr.Code)
	})

	t.Run("Validation - empty scope", func(t *testing.T) {
		_, err := service.SendMessage(ctx, "user-123", "tenant-1", "", "Hello", nil, "")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})

	t.Run("Validation - empty body", func(t *testing.T) {
		_, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "", nil, "")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})

	t.Run("Validation - body too long", func(t *testing.T) {
		longBody := string(make([]byte, 4097))
		_, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", longBody, nil, "")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})

	t.Run("Success - reply to message (threading)", func(t *testing.T) {
		// Create parent message
		parent, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Parent message", nil, "")
		require.NoError(t, err)

		// Create reply
		reply, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Reply message", nil, parent.ID)
		require.NoError(t, err)

		assert.Equal(t, parent.ID, reply.ParentID)

		// Verify retrieval
		retrieved, err := service.GetMessage(ctx, reply.ID, "tenant-1")
		require.NoError(t, err)
		assert.Equal(t, parent.ID, retrieved.ParentID)
	})
}

func TestChatService_GetMessage(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{ID: "user-123", TenantID: "tenant-1", Email: "test@example.com"}
	userRepo.Create(ctx, user)

	// Send a message
	service.SendMessage(ctx, "user-123", "tenant-1", "host", "Initial message", nil, "")

	t.Run("Success - get message", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to get", nil, "")

		retrieved, err := service.GetMessage(ctx, msg.ID, "tenant-1")

		require.NoError(t, err)
		assert.Equal(t, msg.ID, retrieved.ID)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := service.GetMessage(ctx, "non-existent-id", "tenant-1")

		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
	})
}

func TestChatService_EditMessage(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{ID: "user-123", TenantID: "tenant-1", Email: "test@example.com"}
	userRepo.Create(ctx, user)

	t.Run("Success - owner edits", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")

		edited, err := service.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "Edited text", false)

		require.NoError(t, err)
		assert.Equal(t, "Edited text", edited.Body)
		assert.NotNil(t, edited.UpdatedAt)

		// Verify history
		edits, _ := service.GetEditHistory(ctx, msg.ID, "tenant-1")
		assert.Len(t, edits, 1)
		assert.Equal(t, "Original text", edits[0].PrevBody)
	})

	t.Run("Success - admin edits", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")

		edited, err := service.EditMessage(ctx, msg.ID, "admin-456", "tenant-1", "Admin edit", true)

		require.NoError(t, err)
		assert.Equal(t, "Admin edit", edited.Body)
	})

	t.Run("Forbidden - not owner", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")

		_, err := service.EditMessage(ctx, msg.ID, "other-user", "tenant-1", "Hack", false)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("Validation - empty body", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")

		_, err := service.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "", false)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})

	t.Run("Validation - deleted message", func(t *testing.T) {
		oldMsg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")
		service.DeleteMessage(ctx, oldMsg.ID, "user-123", "tenant-1", "soft", false, false)

		_, err := service.EditMessage(ctx, oldMsg.ID, "user-123", "tenant-1", "Too late", false)

		require.Error(t, err)
	})
}

func TestChatService_DeleteMessage(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{ID: "user-123", TenantID: "tenant-1", Email: "test@example.com"}
	userRepo.Create(ctx, user)

	t.Run("Success - soft delete by owner", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "user-123", "tenant-1", "soft", false, false)

		require.NoError(t, err)

		// Verify soft deleted
		retrieved, _ := chatRepo.GetByID(ctx, msg.ID)
		assert.NotNil(t, retrieved.DeletedAt)
	})

	t.Run("Success - hard delete by owner", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "user-123", "tenant-1", "hard", false, false)

		require.NoError(t, err)

		// Verify hard deleted
		_, err = chatRepo.GetByID(ctx, msg.ID)
		require.Error(t, err)
	})

	t.Run("Success - admin deletes", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "admin-456", "tenant-1", "soft", true, false)

		require.NoError(t, err)
	})

	t.Run("Success - moderator deletes", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "mod-789", "tenant-1", "soft", false, true)

		require.NoError(t, err)
	})

	t.Run("Forbidden - not owner", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "other-user", "tenant-1", "soft", false, false)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})

	t.Run("Validation - invalid mode", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to delete", nil, "")

		err := service.DeleteMessage(ctx, msg.ID, "user-123", "tenant-1", "invalid", false, false)

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrValidation, domainErr.Code)
	})
}

func TestChatService_RedactMessage(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{ID: "user-123", TenantID: "tenant-1", Email: "test@example.com"}
	userRepo.Create(ctx, user)

	t.Run("Success - admin redacts", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to redact", nil, "")

		redacted, err := service.RedactMessage(ctx, msg.ID, "admin-456", "tenant-1", true, false, "Violates policy")

		require.NoError(t, err)
		assert.Equal(t, "[REDACTED]", redacted.Body)
		assert.True(t, redacted.Redacted)
	})

	t.Run("Success - moderator redacts", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to redact", nil, "")

		redacted, err := service.RedactMessage(ctx, msg.ID, "mod-789", "tenant-1", false, true, "Spam")

		require.NoError(t, err)
		assert.Equal(t, "[REDACTED]", redacted.Body)
		assert.True(t, redacted.Redacted)
	})

	t.Run("Forbidden - regular user", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message to redact", nil, "")

		_, err := service.RedactMessage(ctx, msg.ID, "other-user", "tenant-1", false, false, "No reason")

		require.Error(t, err)
		var domainErr *domain.Error; ok := errors.As(err, &domainErr)
		require.True(t, ok)
		assert.Equal(t, domain.ErrForbidden, domainErr.Code)
	})
}

func TestChatService_ListMessages(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	ctx := context.Background()

	// Create test user
	user := &domain.User{ID: "user-123", TenantID: "tenant-1", Email: "test@example.com"}
	userRepo.Create(ctx, user)

	// Create test messages
	service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message 1", nil, "")
	time.Sleep(10 * time.Millisecond)
	service.SendMessage(ctx, "user-123", "tenant-1", "host", "Message 2", nil, "")
	time.Sleep(10 * time.Millisecond)
	service.SendMessage(ctx, "user-123", "tenant-1", "network:123", "Message 3", nil, "")

	t.Run("Filter by scope", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Scope:    "host",
			TenantID: "tenant-1",
			Limit:    50,
		}

		messages, cursor, err := service.ListMessages(ctx, filter, "tenant-1")

		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Empty(t, cursor)
		assert.Equal(t, "Message 2", messages[0].Body) // Newest first
		assert.Equal(t, "Message 1", messages[1].Body)
	})

	t.Run("Pagination", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Scope:    "host",
			TenantID: "tenant-1",
			Limit:    1,
		}

		messages, cursor, err := service.ListMessages(ctx, filter, "tenant-1")

		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.NotEmpty(t, cursor)
		assert.Equal(t, "Message 2", messages[0].Body)

		// Get next page
		filter.Cursor = cursor
		messages, cursor, err = service.ListMessages(ctx, filter, "tenant-1")

		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Empty(t, cursor) // No more pages
		assert.Equal(t, "Message 1", messages[0].Body)
	})

	t.Run("Exclude deleted by default", func(t *testing.T) {
		msg, _ := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Deleted message", nil, "")
		service.DeleteMessage(ctx, msg.ID, "user-123", "tenant-1", "soft", false, false)

		filter := domain.ChatMessageFilter{
			Scope:    "host",
			TenantID: "tenant-1",
			Limit:    50,
		}

		messages, _, err := service.ListMessages(ctx, filter, "tenant-1")

		require.NoError(t, err)
		assert.Len(t, messages, 2) // Doesn't include deleted message
	})

	t.Run("Include deleted", func(t *testing.T) {
		filter := domain.ChatMessageFilter{
			Scope:          "host",
			TenantID:       "tenant-1",
			Limit:          50,
			IncludeDeleted: true,
		}

		messages, _, err := service.ListMessages(ctx, filter, "tenant-1")

		require.NoError(t, err)
		assert.Len(t, messages, 3) // Includes deleted message
	})
}

// ==================== SetAuditor Tests ====================

func TestChatService_SetAuditor(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)

	t.Run("Set Auditor With Nil", func(t *testing.T) {
		// Setting nil auditor should be a no-op (not panic)
		service.SetAuditor(nil)
	})

	t.Run("Set Auditor With Valid Auditor", func(t *testing.T) {
		mockAuditor := auditorFunc(func(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {})
		service.SetAuditor(mockAuditor)
		// Should not panic
	})
}

// ==================== GetEditHistory Edge Case Tests ====================

func TestChatService_GetEditHistory_MessageNotFound(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)
	ctx := context.Background()

	_, err := service.GetEditHistory(ctx, "non-existent-msg", "tenant-1")
	require.Error(t, err)
}

func TestChatService_GetEditHistory_WrongTenant(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)
	ctx := context.Background()

	// Create user first
	userRepo.Create(ctx, &domain.User{ID: "user-123", Email: "user@test.com", TenantID: "tenant-1"})

	// Create a message
	msg, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Test message", nil, "")
	require.NoError(t, err)

	// Try to get edit history with wrong tenant
	_, err = service.GetEditHistory(ctx, msg.ID, "wrong-tenant")
	require.Error(t, err)
	var domainErr *domain.Error
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestChatService_GetEditHistory_Success(t *testing.T) {
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()
	service := NewChatService(chatRepo, userRepo)
	ctx := context.Background()

	// Create user first
	userRepo.Create(ctx, &domain.User{ID: "user-123", Email: "user@test.com", TenantID: "tenant-1"})

	// Create a message and edit it
	msg, err := service.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil, "")
	require.NoError(t, err)

	_, err = service.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "Edit 1", false)
	require.NoError(t, err)

	_, err = service.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "Edit 2", false)
	require.NoError(t, err)

	// Get edit history
	edits, err := service.GetEditHistory(ctx, msg.ID, "tenant-1")
	require.NoError(t, err)
	assert.Len(t, edits, 2)
}
