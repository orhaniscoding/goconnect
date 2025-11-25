package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// TenantChatRepository defines the interface for tenant chat messages
type TenantChatRepository interface {
	// CRUD operations
	Create(ctx context.Context, message *domain.TenantChatMessage) error
	GetByID(ctx context.Context, id string) (*domain.TenantChatMessage, error)
	Update(ctx context.Context, message *domain.TenantChatMessage) error
	SoftDelete(ctx context.Context, id string) error

	// List operations - returns messages before the cursor, newest first
	ListByTenant(ctx context.Context, tenantID string, beforeID string, limit int) ([]*domain.TenantChatMessage, error)

	// Cleanup - delete old messages (retention policy)
	DeleteOlderThan(ctx context.Context, tenantID string, before time.Time) (int, error)
}

// InMemoryTenantChatRepository is an in-memory implementation
type InMemoryTenantChatRepository struct {
	mu       sync.RWMutex
	messages map[string]*domain.TenantChatMessage // id -> message
}

// NewInMemoryTenantChatRepository creates a new in-memory repository
func NewInMemoryTenantChatRepository() *InMemoryTenantChatRepository {
	return &InMemoryTenantChatRepository{
		messages: make(map[string]*domain.TenantChatMessage),
	}
}

func (r *InMemoryTenantChatRepository) Create(ctx context.Context, message *domain.TenantChatMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if message.ID == "" {
		message.ID = domain.GenerateChatMessageID()
	}
	message.CreatedAt = time.Now()
	r.messages[message.ID] = message
	return nil
}

func (r *InMemoryTenantChatRepository) GetByID(ctx context.Context, id string) (*domain.TenantChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, exists := r.messages[id]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	if msg.DeletedAt != nil {
		return nil, domain.NewError(domain.ErrNotFound, "Message has been deleted", nil)
	}
	return msg, nil
}

func (r *InMemoryTenantChatRepository) Update(ctx context.Context, message *domain.TenantChatMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.messages[message.ID]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	if existing.DeletedAt != nil {
		return domain.NewError(domain.ErrNotFound, "Message has been deleted", nil)
	}
	now := time.Now()
	message.EditedAt = &now
	r.messages[message.ID] = message
	return nil
}

func (r *InMemoryTenantChatRepository) SoftDelete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, exists := r.messages[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Message not found", nil)
	}
	now := time.Now()
	msg.DeletedAt = &now
	return nil
}

func (r *InMemoryTenantChatRepository) ListByTenant(ctx context.Context, tenantID string, beforeID string, limit int) ([]*domain.TenantChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantChatMessage
	var beforeTime time.Time

	// Find the "before" message's timestamp
	if beforeID != "" {
		if beforeMsg, exists := r.messages[beforeID]; exists {
			beforeTime = beforeMsg.CreatedAt
		}
	}

	for _, m := range r.messages {
		if m.TenantID == tenantID && m.DeletedAt == nil {
			if beforeID == "" || m.CreatedAt.Before(beforeTime) {
				result = append(result, m)
			}
		}
	}

	// Sort by created_at desc (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryTenantChatRepository) DeleteOlderThan(ctx context.Context, tenantID string, before time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for id, m := range r.messages {
		if m.TenantID == tenantID && m.CreatedAt.Before(before) {
			delete(r.messages, id)
			deleted++
		}
	}
	return deleted, nil
}
