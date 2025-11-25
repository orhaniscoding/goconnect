package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// TenantInviteRepository defines the interface for tenant invite operations
type TenantInviteRepository interface {
	// CRUD operations
	Create(ctx context.Context, invite *domain.TenantInvite) error
	GetByID(ctx context.Context, id string) (*domain.TenantInvite, error)
	GetByCode(ctx context.Context, code string) (*domain.TenantInvite, error)
	Delete(ctx context.Context, id string) error

	// List operations
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.TenantInvite, error)

	// Usage operations
	IncrementUseCount(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error

	// Cleanup
	DeleteExpired(ctx context.Context) (int, error)
}

// InMemoryTenantInviteRepository is an in-memory implementation
type InMemoryTenantInviteRepository struct {
	mu      sync.RWMutex
	invites map[string]*domain.TenantInvite // id -> invite
}

// NewInMemoryTenantInviteRepository creates a new in-memory repository
func NewInMemoryTenantInviteRepository() *InMemoryTenantInviteRepository {
	return &InMemoryTenantInviteRepository{
		invites: make(map[string]*domain.TenantInvite),
	}
}

func (r *InMemoryTenantInviteRepository) Create(ctx context.Context, invite *domain.TenantInvite) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate code
	for _, i := range r.invites {
		if strings.EqualFold(i.Code, invite.Code) {
			return domain.NewError(domain.ErrValidation, "Invite code already exists", nil)
		}
	}

	if invite.ID == "" {
		invite.ID = domain.GenerateTenantInviteID()
	}
	invite.CreatedAt = time.Now()
	r.invites[invite.ID] = invite
	return nil
}

func (r *InMemoryTenantInviteRepository) GetByID(ctx context.Context, id string) (*domain.TenantInvite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invite, exists := r.invites[id]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	return invite, nil
}

func (r *InMemoryTenantInviteRepository) GetByCode(ctx context.Context, code string) (*domain.TenantInvite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	for _, i := range r.invites {
		if strings.ToUpper(i.Code) == normalizedCode {
			return i, nil
		}
	}
	return nil, domain.NewError(domain.ErrNotFound, "Invalid invite code", nil)
}

func (r *InMemoryTenantInviteRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.invites[id]; !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	delete(r.invites, id)
	return nil
}

func (r *InMemoryTenantInviteRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.TenantInvite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantInvite
	for _, i := range r.invites {
		if i.TenantID == tenantID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (r *InMemoryTenantInviteRepository) IncrementUseCount(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, exists := r.invites[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	invite.UseCount++
	return nil
}

func (r *InMemoryTenantInviteRepository) Revoke(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, exists := r.invites[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant invite not found", nil)
	}
	now := time.Now()
	invite.RevokedAt = &now
	return nil
}

func (r *InMemoryTenantInviteRepository) DeleteExpired(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	now := time.Now()
	for id, i := range r.invites {
		if i.ExpiresAt != nil && now.After(*i.ExpiresAt) {
			delete(r.invites, id)
			deleted++
		}
	}
	return deleted, nil
}
