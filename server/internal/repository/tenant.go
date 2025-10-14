package repository

import (
	"context"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// TenantRepository defines the interface for tenant storage operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id string) error
}

// InMemoryTenantRepository is an in-memory implementation of TenantRepository
type InMemoryTenantRepository struct {
	mu      sync.RWMutex
	tenants map[string]*domain.Tenant // id -> tenant
}

// NewInMemoryTenantRepository creates a new in-memory tenant repository
func NewInMemoryTenantRepository() *InMemoryTenantRepository {
	return &InMemoryTenantRepository{
		tenants: make(map[string]*domain.Tenant),
	}
}

// Create adds a new tenant to the repository
func (r *InMemoryTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tenants[tenant.ID] = tenant
	return nil
}

// GetByID retrieves a tenant by ID
func (r *InMemoryTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenant, exists := r.tenants[id]
	if !exists {
		return nil, domain.NewError(domain.ErrTenantNotFound, "Tenant not found", map[string]string{"tenant_id": id})
	}

	return tenant, nil
}

// Update updates an existing tenant
func (r *InMemoryTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tenants[tenant.ID]; !exists {
		return domain.NewError(domain.ErrTenantNotFound, "Tenant not found", map[string]string{"tenant_id": tenant.ID})
	}

	r.tenants[tenant.ID] = tenant
	return nil
}

// Delete removes a tenant from the repository
func (r *InMemoryTenantRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tenants[id]; !exists {
		return domain.NewError(domain.ErrTenantNotFound, "Tenant not found", map[string]string{"tenant_id": id})
	}

	delete(r.tenants, id)
	return nil
}
