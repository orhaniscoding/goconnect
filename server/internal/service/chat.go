package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ChatService handles chat message operations
type ChatService struct {
	chatRepo repository.ChatRepository
	userRepo repository.UserRepository
	auditor  Auditor
}

// NewChatService creates a new chat service
func NewChatService(chatRepo repository.ChatRepository, userRepo repository.UserRepository) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
		auditor:  noopAuditor,
	}
}

// SetAuditor sets the auditor for the service
func (s *ChatService) SetAuditor(auditor Auditor) {
	if auditor != nil {
		s.auditor = auditor
	}
}

// SendMessage creates a new chat message
func (s *ChatService) SendMessage(ctx context.Context, userID, tenantID, scope, body string, attachments []string) (*domain.ChatMessage, error) {
	// Validate user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{
			"user_id": userID,
		})
	}

	// Create message
	msg := &domain.ChatMessage{
		ID:          domain.GenerateNetworkID(),
		Scope:       scope,
		TenantID:    tenantID,
		UserID:      userID,
		Body:        body,
		Attachments: attachments,
		Redacted:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Save
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "CHAT_MESSAGE_SENT", userID, msg.ID, map[string]any{
		"scope":     scope,
		"tenant_id": tenantID,
	})

	return msg, nil
}

// GetMessage retrieves a message by ID
func (s *ChatService) GetMessage(ctx context.Context, messageID, tenantID string) (*domain.ChatMessage, error) {
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	return msg, nil
}

// ListMessages retrieves messages matching the filter
func (s *ChatService) ListMessages(ctx context.Context, filter domain.ChatMessageFilter, tenantID string) ([]*domain.ChatMessage, string, error) {
	filter.TenantID = tenantID
	return s.chatRepo.List(ctx, filter)
}

// EditMessage edits an existing message
func (s *ChatService) EditMessage(ctx context.Context, messageID, userID, tenantID, newBody string, isAdmin bool) (*domain.ChatMessage, error) {
	// Get message
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Ensure tenant match
	if msg.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}

	// Check if deleted
	if msg.IsDeleted() {
		return nil, domain.NewError(domain.ErrForbidden, "Cannot edit deleted message", nil)
	}

	if newBody == "" {
		return nil, domain.NewError(domain.ErrValidation, "Message body cannot be empty", nil)
	}

	// Check permissions
	if !msg.CanEdit(userID, isAdmin) {
		return nil, domain.NewError(domain.ErrForbidden, "You cannot edit this message", nil)
	}

	// Check edit time limit (15 minutes for non-admins)
	if !isAdmin && time.Since(msg.CreatedAt) > 15*time.Minute {
		return nil, domain.NewError(domain.ErrForbidden, "Edit time limit exceeded (15 minutes)", nil)
	}

	// Save edit history
	edit := &domain.ChatMessageEdit{
		MessageID: messageID,
		PrevBody:  msg.Body,
		NewBody:   newBody,
		EditorID:  userID,
		EditedAt:  time.Now(),
	}
	if err := s.chatRepo.AddEdit(ctx, edit); err != nil {
		return nil, fmt.Errorf("failed to save edit history: %w", err)
	}

	// Update message
	msg.Edit(newBody)
	if err := s.chatRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "CHAT_MESSAGE_EDITED", userID, messageID, map[string]any{
		"prev_body": edit.PrevBody,
		"new_body":  newBody,
	})

	return msg, nil
}

// DeleteMessage deletes a message (soft or hard)
func (s *ChatService) DeleteMessage(ctx context.Context, messageID, userID, tenantID string, mode string, isAdmin, isModerator bool) error {
	// Get message
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Ensure tenant match
	if msg.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}

	// Check permissions
	if !msg.CanDelete(userID, isAdmin, isModerator) {
		return domain.NewError(domain.ErrForbidden, "You cannot delete this message", nil)
	}

	// Validate mode
	if mode != "soft" && mode != "hard" {
		return domain.NewError(domain.ErrValidation, "Invalid delete mode", map[string]string{
			"mode":        mode,
			"valid_modes": "soft,hard",
		})
	}

	// Delete
	if mode == "soft" {
		if err := s.chatRepo.SoftDelete(ctx, messageID); err != nil {
			return fmt.Errorf("failed to soft delete message: %w", err)
		}
	} else {
		if err := s.chatRepo.Delete(ctx, messageID); err != nil {
			return fmt.Errorf("failed to hard delete message: %w", err)
		}
	}

	// Audit
	s.auditor.Event(ctx, tenantID, fmt.Sprintf("CHAT_MESSAGE_%s_DELETED", mode), userID, messageID, map[string]any{
		"mode":  mode,
		"scope": msg.Scope,
	})

	return nil
}

// RedactMessage redacts a message (moderator/admin only)
func (s *ChatService) RedactMessage(ctx context.Context, messageID, moderatorID, tenantID string, isAdmin, isModerator bool, reason string) (*domain.ChatMessage, error) {
	// Get message
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Ensure tenant match
	if msg.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}

	// Check permissions
	if !msg.CanRedact(isAdmin, isModerator) {
		return nil, domain.NewError(domain.ErrForbidden, "You cannot redact messages", nil)
	}

	// Save original body
	originalBody := msg.Body

	// Redact
	msg.Redact()
	if err := s.chatRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to redact message: %w", err)
	}

	// Audit
	s.auditor.Event(ctx, tenantID, "CHAT_MESSAGE_REDACTED", moderatorID, messageID, map[string]any{
		"original_body": originalBody,
		"reason":        reason,
		"scope":         msg.Scope,
	})

	return msg, nil
}

// GetEditHistory retrieves edit history for a message
func (s *ChatService) GetEditHistory(ctx context.Context, messageID, tenantID string) ([]*domain.ChatMessageEdit, error) {
	// Get message to check tenant
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg.TenantID != tenantID {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}

	return s.chatRepo.GetEdits(ctx, messageID)
}
