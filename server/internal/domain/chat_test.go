package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatMessage_Validate_Success(t *testing.T) {
	msg := &ChatMessage{
		Scope:  "network:123",
		UserID: "user-456",
		Body:   "Hello, world!",
	}

	err := msg.Validate()
	assert.NoError(t, err)
}

func TestChatMessage_Validate_MissingScope(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-456",
		Body:   "Hello",
	}

	err := msg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scope is required")
}

func TestChatMessage_Validate_MissingUserID(t *testing.T) {
	msg := &ChatMessage{
		Scope: "network:123",
		Body:  "Hello",
	}

	err := msg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestChatMessage_Validate_MissingBody(t *testing.T) {
	msg := &ChatMessage{
		Scope:  "network:123",
		UserID: "user-456",
		Body:   "",
	}

	err := msg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body is required")
}

func TestChatMessage_Validate_RedactedWithoutBody(t *testing.T) {
	msg := &ChatMessage{
		Scope:    "network:123",
		UserID:   "user-456",
		Body:     "",
		Redacted: true,
	}

	// Redacted messages don't need body
	err := msg.Validate()
	assert.NoError(t, err)
}

func TestChatMessage_Validate_BodyTooLong(t *testing.T) {
	longBody := make([]byte, 4097)
	for i := range longBody {
		longBody[i] = 'a'
	}

	msg := &ChatMessage{
		Scope:  "network:123",
		UserID: "user-456",
		Body:   string(longBody),
	}

	err := msg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body exceeds maximum length")
}

func TestChatMessage_Validate_BodyMaxLength(t *testing.T) {
	// Exactly 4096 chars should be valid
	maxBody := make([]byte, 4096)
	for i := range maxBody {
		maxBody[i] = 'a'
	}

	msg := &ChatMessage{
		Scope:  "network:123",
		UserID: "user-456",
		Body:   string(maxBody),
	}

	err := msg.Validate()
	assert.NoError(t, err)
}

func TestChatMessage_IsDeleted(t *testing.T) {
	// Not deleted
	msg := &ChatMessage{
		DeletedAt: nil,
	}
	assert.False(t, msg.IsDeleted())

	// Deleted
	now := time.Now()
	msg.DeletedAt = &now
	assert.True(t, msg.IsDeleted())
}

func TestChatMessage_CanEdit_Owner(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Owner can edit
	assert.True(t, msg.CanEdit("user-123", false))

	// Other user cannot edit
	assert.False(t, msg.CanEdit("user-456", false))
}

func TestChatMessage_CanEdit_Admin(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Admin can edit any message
	assert.True(t, msg.CanEdit("user-456", true))
	assert.True(t, msg.CanEdit("user-123", true))
}

func TestChatMessage_CanDelete_Owner(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Owner can delete
	assert.True(t, msg.CanDelete("user-123", false, false))

	// Other user cannot delete
	assert.False(t, msg.CanDelete("user-456", false, false))
}

func TestChatMessage_CanDelete_Admin(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Admin can delete any message
	assert.True(t, msg.CanDelete("user-456", true, false))
	assert.True(t, msg.CanDelete("admin-1", true, false))
}

func TestChatMessage_CanDelete_Moderator(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Moderator can delete any message
	assert.True(t, msg.CanDelete("mod-1", false, true))
	assert.True(t, msg.CanDelete("user-456", false, true))
}

func TestChatMessage_CanRedact_RegularUser(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Regular user cannot redact
	assert.False(t, msg.CanRedact(false, false))
}

func TestChatMessage_CanRedact_Admin(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Admin can redact
	assert.True(t, msg.CanRedact(true, false))
}

func TestChatMessage_CanRedact_Moderator(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-123",
	}

	// Moderator can redact
	assert.True(t, msg.CanRedact(false, true))
}

func TestChatMessage_SoftDelete(t *testing.T) {
	msg := &ChatMessage{
		ID:        "msg-1",
		DeletedAt: nil,
	}

	msg.SoftDelete()

	assert.NotNil(t, msg.DeletedAt)
	assert.WithinDuration(t, time.Now(), *msg.DeletedAt, 1*time.Second)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, 1*time.Second)
}

func TestChatMessage_Edit(t *testing.T) {
	msg := &ChatMessage{
		ID:   "msg-1",
		Body: "Original message",
	}

	msg.Edit("Updated message")

	assert.Equal(t, "Updated message", msg.Body)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, 1*time.Second)
}

func TestChatMessage_Redact(t *testing.T) {
	msg := &ChatMessage{
		ID:       "msg-1",
		Body:     "Inappropriate content",
		Redacted: false,
	}

	msg.Redact()

	assert.True(t, msg.Redacted)
	assert.Equal(t, "[REDACTED]", msg.Body)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, 1*time.Second)
}

func TestChatMessage_WithAttachments(t *testing.T) {
	msg := &ChatMessage{
		Scope:       "network:123",
		UserID:      "user-456",
		Body:        "Check these files",
		Attachments: []string{"file1.pdf", "image.png"},
	}

	err := msg.Validate()
	assert.NoError(t, err)
	assert.Len(t, msg.Attachments, 2)
	assert.Contains(t, msg.Attachments, "file1.pdf")
	assert.Contains(t, msg.Attachments, "image.png")
}

func TestChatMessageEdit(t *testing.T) {
	edit := &ChatMessageEdit{
		ID:        "edit-1",
		MessageID: "msg-1",
		PrevBody:  "Old text",
		NewBody:   "New text",
		EditorID:  "user-123",
		EditedAt:  time.Now(),
	}

	assert.Equal(t, "edit-1", edit.ID)
	assert.Equal(t, "msg-1", edit.MessageID)
	assert.Equal(t, "Old text", edit.PrevBody)
	assert.Equal(t, "New text", edit.NewBody)
	assert.Equal(t, "user-123", edit.EditorID)
}

func TestRedactionMask(t *testing.T) {
	mask := &RedactionMask{
		Pattern:    "***",
		RedactedBy: "mod-1",
		RedactedAt: time.Now(),
		Reason:     "Spam",
	}

	assert.Equal(t, "***", mask.Pattern)
	assert.Equal(t, "mod-1", mask.RedactedBy)
	assert.Equal(t, "Spam", mask.Reason)
}

func TestChatMessageFilter(t *testing.T) {
	since := time.Now().Add(-1 * time.Hour)
	before := time.Now()

	filter := &ChatMessageFilter{
		Scope:          "network:123",
		UserID:         "user-456",
		Since:          since,
		Before:         before,
		Limit:          50,
		Cursor:         "cursor-abc",
		IncludeDeleted: true,
	}

	assert.Equal(t, "network:123", filter.Scope)
	assert.Equal(t, "user-456", filter.UserID)
	assert.Equal(t, since, filter.Since)
	assert.Equal(t, before, filter.Before)
	assert.Equal(t, 50, filter.Limit)
	assert.Equal(t, "cursor-abc", filter.Cursor)
	assert.True(t, filter.IncludeDeleted)
}

func TestChatMessage_FullLifecycle(t *testing.T) {
	// Create message
	msg := &ChatMessage{
		ID:       "msg-1",
		Scope:    "network:123",
		TenantID: "tenant-1",
		UserID:   "user-1",
		Body:     "Original message",
	}

	// Validate
	err := msg.Validate()
	require.NoError(t, err)

	// Owner can edit
	assert.True(t, msg.CanEdit("user-1", false))

	// Edit message
	msg.Edit("Updated message")
	assert.Equal(t, "Updated message", msg.Body)

	// Not deleted yet
	assert.False(t, msg.IsDeleted())

	// Soft delete
	msg.SoftDelete()
	assert.True(t, msg.IsDeleted())
}

func TestChatMessage_ModeratorWorkflow(t *testing.T) {
	// User creates message
	msg := &ChatMessage{
		ID:       "msg-1",
		Scope:    "network:123",
		UserID:   "user-1",
		Body:     "Inappropriate content here",
		Redacted: false,
	}

	// Regular user cannot redact
	assert.False(t, msg.CanRedact(false, false))

	// Moderator can redact
	assert.True(t, msg.CanRedact(false, true))

	// Redact the message
	msg.Redact()
	assert.True(t, msg.Redacted)
	assert.Equal(t, "[REDACTED]", msg.Body)

	// Moderator can delete
	assert.True(t, msg.CanDelete("mod-1", false, true))
}

func TestChatMessage_Permissions(t *testing.T) {
	msg := &ChatMessage{
		UserID: "user-1",
	}

	testCases := []struct {
		name        string
		userID      string
		isAdmin     bool
		isModerator bool
		canEdit     bool
		canDelete   bool
		canRedact   bool
	}{
		{
			name:        "owner - regular user",
			userID:      "user-1",
			isAdmin:     false,
			isModerator: false,
			canEdit:     true,
			canDelete:   true,
			canRedact:   false,
		},
		{
			name:        "other user - no permissions",
			userID:      "user-2",
			isAdmin:     false,
			isModerator: false,
			canEdit:     false,
			canDelete:   false,
			canRedact:   false,
		},
		{
			name:        "admin - all permissions",
			userID:      "admin-1",
			isAdmin:     true,
			isModerator: false,
			canEdit:     true,
			canDelete:   true,
			canRedact:   true,
		},
		{
			name:        "moderator - delete and redact",
			userID:      "mod-1",
			isAdmin:     false,
			isModerator: true,
			canEdit:     false,
			canDelete:   true,
			canRedact:   true,
		},
		{
			name:        "admin+moderator - all permissions",
			userID:      "super-1",
			isAdmin:     true,
			isModerator: true,
			canEdit:     true,
			canDelete:   true,
			canRedact:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.canEdit, msg.CanEdit(tc.userID, tc.isAdmin))
			assert.Equal(t, tc.canDelete, msg.CanDelete(tc.userID, tc.isAdmin, tc.isModerator))
			assert.Equal(t, tc.canRedact, msg.CanRedact(tc.isAdmin, tc.isModerator))
		})
	}
}
