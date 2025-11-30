package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// TenantMemberRepository defines the interface for tenant membership operations
type TenantMemberRepository interface {
	// CRUD operations
	Create(ctx context.Context, member *domain.TenantMember) error
	GetByID(ctx context.Context, id string) (*domain.TenantMember, error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.TenantMember, error)
	Update(ctx context.Context, member *domain.TenantMember) error
	Delete(ctx context.Context, id string) error

	// List operations
	ListByTenant(ctx context.Context, tenantID string, role string, limit int, cursor string) ([]*domain.TenantMember, string, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.TenantMember, error)

	// Count operations
	CountByTenant(ctx context.Context, tenantID string) (int, error)

	// Role operations
	UpdateRole(ctx context.Context, id string, role domain.TenantRole) error
	GetUserRole(ctx context.Context, userID, tenantID string) (domain.TenantRole, error)
	HasRole(ctx context.Context, userID, tenantID string, requiredRole domain.TenantRole) (bool, error)

	// Ban operations
	Ban(ctx context.Context, id string, bannedBy string) error
	Unban(ctx context.Context, id string) error
	IsBanned(ctx context.Context, userID, tenantID string) (bool, error)
	ListBannedByTenant(ctx context.Context, tenantID string) ([]*domain.TenantMember, error)

	// Bulk delete - used when deleting tenant
	DeleteAllByTenant(ctx context.Context, tenantID string) error
}

// InMemoryTenantMemberRepository is an in-memory implementation
type InMemoryTenantMemberRepository struct {
	mu      sync.RWMutex
	members map[string]*domain.TenantMember // id -> member
}

// NewInMemoryTenantMemberRepository creates a new in-memory repository
func NewInMemoryTenantMemberRepository() *InMemoryTenantMemberRepository {
	return &InMemoryTenantMemberRepository{
		members: make(map[string]*domain.TenantMember),
	}
}

func (r *InMemoryTenantMemberRepository) Create(ctx context.Context, member *domain.TenantMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate
	for _, m := range r.members {
		if m.TenantID == member.TenantID && m.UserID == member.UserID {
			return domain.NewError(domain.ErrValidation, "User is already a member of this tenant", nil)
		}
	}

	if member.ID == "" {
		member.ID = domain.GenerateTenantMemberID()
	}
	member.JoinedAt = time.Now()
	member.UpdatedAt = time.Now()
	r.members[member.ID] = member
	return nil
}

func (r *InMemoryTenantMemberRepository) GetByID(ctx context.Context, id string) (*domain.TenantMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	member, exists := r.members[id]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	return member, nil
}

func (r *InMemoryTenantMemberRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.TenantMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.members {
		if m.UserID == userID && m.TenantID == tenantID {
			return m, nil
		}
	}
	return nil, domain.NewError(domain.ErrNotFound, "Tenant membership not found", nil)
}

func (r *InMemoryTenantMemberRepository) Update(ctx context.Context, member *domain.TenantMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[member.ID]; !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	member.UpdatedAt = time.Now()
	r.members[member.ID] = member
	return nil
}

func (r *InMemoryTenantMemberRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[id]; !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	delete(r.members, id)
	return nil
}

func (r *InMemoryTenantMemberRepository) ListByTenant(ctx context.Context, tenantID string, role string, limit int, cursor string) ([]*domain.TenantMember, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantMember
	for _, m := range r.members {
		if m.TenantID == tenantID {
			if role == "" || string(m.Role) == role {
				result = append(result, m)
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, "", nil
}

func (r *InMemoryTenantMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.TenantMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantMember
	for _, m := range r.members {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryTenantMemberRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, m := range r.members {
		if m.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryTenantMemberRepository) UpdateRole(ctx context.Context, id string, role domain.TenantRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	member, exists := r.members[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	member.Role = role
	member.UpdatedAt = time.Now()
	return nil
}

func (r *InMemoryTenantMemberRepository) GetUserRole(ctx context.Context, userID, tenantID string) (domain.TenantRole, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.members {
		if m.UserID == userID && m.TenantID == tenantID {
			return m.Role, nil
		}
	}
	return "", domain.NewError(domain.ErrNotFound, "User is not a member of this tenant", nil)
}

func (r *InMemoryTenantMemberRepository) HasRole(ctx context.Context, userID, tenantID string, requiredRole domain.TenantRole) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.members {
		if m.UserID == userID && m.TenantID == tenantID {
			return m.Role.HasPermission(requiredRole), nil
		}
	}
	return false, nil
}

func (r *InMemoryTenantMemberRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, m := range r.members {
		if m.TenantID == tenantID {
			delete(r.members, id)
		}
	}
	return nil
}

func (r *InMemoryTenantMemberRepository) Ban(ctx context.Context, id string, bannedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	member, exists := r.members[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	now := time.Now()
	member.BannedAt = &now
	member.BannedBy = bannedBy
	member.UpdatedAt = now
	return nil
}

func (r *InMemoryTenantMemberRepository) Unban(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	member, exists := r.members[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Tenant member not found", nil)
	}
	member.BannedAt = nil
	member.BannedBy = ""
	member.UpdatedAt = time.Now()
	return nil
}

func (r *InMemoryTenantMemberRepository) IsBanned(ctx context.Context, userID, tenantID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.members {
		if m.UserID == userID && m.TenantID == tenantID {
			return m.IsBanned(), nil
		}
	}
	return false, nil
}

func (r *InMemoryTenantMemberRepository) ListBannedByTenant(ctx context.Context, tenantID string) ([]*domain.TenantMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantMember
	for _, m := range r.members {
		if m.TenantID == tenantID && m.IsBanned() {
			result = append(result, m)
		}
	}
	return result, nil
}
