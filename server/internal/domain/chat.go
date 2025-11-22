package domain

import (
	"time"
)

// ChatMessage represents a chat message in a scope
type ChatMessage struct {
	ID          string     `json:"id"`
	Scope       string     `json:"scope"`     // "host" or "network:<network_id>"
	TenantID    string     `json:"tenant_id"` // For multi-tenancy
	UserID      string     `json:"user_id"`
	Body        string     `json:"body"`
	Attachments []string   `json:"attachments,omitempty"`
	Redacted    bool       `json:"redacted"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ParentID    string     `json:"parent_id,omitempty"` // For threads
}

// ChatMessageEdit represents an edit history entry
type ChatMessageEdit struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	PrevBody  string    `json:"prev_body"`
	NewBody   string    `json:"new_body"`
	EditorID  string    `json:"editor_id"`
	EditedAt  time.Time `json:"edited_at"`
}

// RedactionMask represents redaction information
type RedactionMask struct {
	Pattern    string    `json:"pattern"`
	RedactedBy string    `json:"redacted_by"`
	RedactedAt time.Time `json:"redacted_at"`
	Reason     string    `json:"reason,omitempty"`
}

// ChatMessageFilter represents filtering options for chat messages
type ChatMessageFilter struct {
	Scope          string    // Filter by scope
	TenantID       string    // Filter by tenant
	UserID         string    // Filter by user
	Since          time.Time // Messages after this time
	Before         time.Time // Messages before this time
	Limit          int       // Max messages to return (default 50, max 100)
	Cursor         string    // Pagination cursor
	IncludeDeleted bool      // Include soft-deleted messages
}

// Validate validates the chat message
func (m *ChatMessage) Validate() error {
	if m.Scope == "" {
		return NewError(ErrValidation, "scope is required", nil)
	}

	if m.UserID == "" {
		return NewError(ErrValidation, "user_id is required", nil)
	}

	if m.Body == "" && !m.Redacted {
		return NewError(ErrValidation, "body is required for non-redacted messages", nil)
	}

	if len(m.Body) > 4096 {
		return NewError(ErrValidation, "body exceeds maximum length of 4096 characters", map[string]string{
			"max_length":    "4096",
			"actual_length": string(rune(len(m.Body))),
		})
	}

	return nil
}

// IsDeleted checks if the message is soft-deleted
func (m *ChatMessage) IsDeleted() bool {
	return m.DeletedAt != nil
}

// CanEdit checks if a user can edit this message
func (m *ChatMessage) CanEdit(userID string, isAdmin bool) bool {
	// User can edit their own message (within time limit handled by service)
	// Admin can edit any message
	return m.UserID == userID || isAdmin
}

// CanDelete checks if a user can delete this message
func (m *ChatMessage) CanDelete(userID string, isAdmin bool, isModerator bool) bool {
	// User can delete their own message
	// Admin/moderator can delete any message
	return m.UserID == userID || isAdmin || isModerator
}

// CanRedact checks if a user can redact this message
func (m *ChatMessage) CanRedact(isAdmin bool, isModerator bool) bool {
	// Only admin/moderator can redact messages
	return isAdmin || isModerator
}

// SoftDelete marks the message as deleted
func (m *ChatMessage) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
}

// Edit updates the message body
func (m *ChatMessage) Edit(newBody string) {
	m.Body = newBody
	m.UpdatedAt = time.Now()
}

// Redact redacts the message
func (m *ChatMessage) Redact() {
	m.Redacted = true
	m.Body = "[REDACTED]"
	m.UpdatedAt = time.Now()
}
