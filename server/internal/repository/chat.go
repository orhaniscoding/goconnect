package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ChatRepository defines the interface for chat message storage
type ChatRepository interface {
	Create(ctx context.Context, msg *domain.ChatMessage) error
	GetByID(ctx context.Context, id string) (*domain.ChatMessage, error)
	List(ctx context.Context, filter domain.ChatMessageFilter) ([]*domain.ChatMessage, string, error)
	Update(ctx context.Context, msg *domain.ChatMessage) error
	Delete(ctx context.Context, id string) error // Hard delete
	SoftDelete(ctx context.Context, id string) error

	// Edit history
	AddEdit(ctx context.Context, edit *domain.ChatMessageEdit) error
	GetEdits(ctx context.Context, messageID string) ([]*domain.ChatMessageEdit, error)
}

// InMemoryChatRepository implements ChatRepository with in-memory storage
type InMemoryChatRepository struct {
	mu       sync.RWMutex
	messages map[string]*domain.ChatMessage
	edits    map[string][]*domain.ChatMessageEdit // messageID -> edits
}

// NewInMemoryChatRepository creates a new in-memory chat repository
func NewInMemoryChatRepository() *InMemoryChatRepository {
	return &InMemoryChatRepository{
		messages: make(map[string]*domain.ChatMessage),
		edits:    make(map[string][]*domain.ChatMessageEdit),
	}
}

// Create creates a new chat message
func (r *InMemoryChatRepository) Create(ctx context.Context, msg *domain.ChatMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg.ID == "" {
		msg.ID = domain.GenerateNetworkID() // Reuse ULID generator
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = msg.CreatedAt
	}

	r.messages[msg.ID] = msg
	return nil
}

// GetByID retrieves a message by ID
func (r *InMemoryChatRepository) GetByID(ctx context.Context, id string) (*domain.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, ok := r.messages[id]
	if !ok {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}

	return msg, nil
}

// List retrieves messages matching the filter
func (r *InMemoryChatRepository) List(ctx context.Context, filter domain.ChatMessageFilter) ([]*domain.ChatMessage, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*domain.ChatMessage

	// Filter messages
	for _, msg := range r.messages {
		// Tenant filter
		if filter.TenantID != "" && msg.TenantID != filter.TenantID {
			continue
		}

		// Scope filter
		if filter.Scope != "" && msg.Scope != filter.Scope {
			continue
		}

		// User filter
		if filter.UserID != "" && msg.UserID != filter.UserID {
			continue
		}

		// Time filters
		if !filter.Since.IsZero() && msg.CreatedAt.Before(filter.Since) {
			continue
		}
		if !filter.Before.IsZero() && msg.CreatedAt.After(filter.Before) {
			continue
		}

		// Deleted filter
		if !filter.IncludeDeleted && msg.IsDeleted() {
			continue
		}

		matches = append(matches, msg)
	}

	// Sort by CreatedAt DESC (newest first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].CreatedAt.Before(matches[j].CreatedAt) {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Apply pagination
	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Cursor pagination
	startIdx := 0
	if filter.Cursor != "" {
		for i, msg := range matches {
			if msg.ID == filter.Cursor {
				startIdx = i + 1
				break
			}
		}
	}

	endIdx := startIdx + limit
	if endIdx > len(matches) {
		endIdx = len(matches)
	}

	result := matches[startIdx:endIdx]

	var nextCursor string
	if endIdx < len(matches) {
		nextCursor = matches[endIdx-1].ID
	}

	return result, nextCursor, nil
}

// Update updates an existing message
func (r *InMemoryChatRepository) Update(ctx context.Context, msg *domain.ChatMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[msg.ID]; !ok {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": msg.ID,
		})
	}

	msg.UpdatedAt = time.Now()
	r.messages[msg.ID] = msg
	return nil
}

// Delete hard deletes a message
func (r *InMemoryChatRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[id]; !ok {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}

	delete(r.messages, id)
	delete(r.edits, id) // Also delete edit history
	return nil
}

// SoftDelete marks a message as deleted
func (r *InMemoryChatRepository) SoftDelete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, ok := r.messages[id]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Message not found", map[string]string{
			"message_id": id,
		})
	}

	msg.SoftDelete()
	return nil
}

// AddEdit adds an edit history entry
func (r *InMemoryChatRepository) AddEdit(ctx context.Context, edit *domain.ChatMessageEdit) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if edit.ID == "" {
		edit.ID = domain.GenerateNetworkID()
	}

	if edit.EditedAt.IsZero() {
		edit.EditedAt = time.Now()
	}

	r.edits[edit.MessageID] = append(r.edits[edit.MessageID], edit)
	return nil
}

// GetEdits retrieves edit history for a message
func (r *InMemoryChatRepository) GetEdits(ctx context.Context, messageID string) ([]*domain.ChatMessageEdit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	edits := r.edits[messageID]
	if edits == nil {
		return []*domain.ChatMessageEdit{}, nil
	}

	return edits, nil
}
