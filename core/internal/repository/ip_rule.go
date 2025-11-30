package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// IPRuleRepository defines the interface for IP rule operations
type IPRuleRepository interface {
	Create(ctx context.Context, rule *domain.IPRule) error
	GetByID(ctx context.Context, id string) (*domain.IPRule, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.IPRule, error)
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int, error)
}

// InMemoryIPRuleRepository is an in-memory implementation
type InMemoryIPRuleRepository struct {
	mu       sync.RWMutex
	byID     map[string]*domain.IPRule
	byTenant map[string][]*domain.IPRule
}

// NewInMemoryIPRuleRepository creates a new in-memory IP rule repository
func NewInMemoryIPRuleRepository() *InMemoryIPRuleRepository {
	return &InMemoryIPRuleRepository{
		byID:     make(map[string]*domain.IPRule),
		byTenant: make(map[string][]*domain.IPRule),
	}
}

// Create stores a new IP rule
func (r *InMemoryIPRuleRepository) Create(ctx context.Context, rule *domain.IPRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byID[rule.ID]; exists {
		return domain.NewError(domain.ErrConflict, "IP rule ID already exists", nil)
	}

	r.byID[rule.ID] = rule
	r.byTenant[rule.TenantID] = append(r.byTenant[rule.TenantID], rule)

	return nil
}

// GetByID retrieves an IP rule by its ID
func (r *InMemoryIPRuleRepository) GetByID(ctx context.Context, id string) (*domain.IPRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, exists := r.byID[id]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}
	return rule, nil
}

// ListByTenant lists all IP rules for a tenant
func (r *InMemoryIPRuleRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.IPRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := r.byTenant[tenantID]
	result := make([]*domain.IPRule, 0, len(rules))
	now := time.Now()

	for _, rule := range rules {
		// Filter out expired rules
		if rule.ExpiresAt == nil || now.Before(*rule.ExpiresAt) {
			result = append(result, rule)
		}
	}

	return result, nil
}

// Delete removes an IP rule
func (r *InMemoryIPRuleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule, exists := r.byID[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "IP rule not found", nil)
	}

	delete(r.byID, id)

	// Remove from tenant list
	tenantRules := r.byTenant[rule.TenantID]
	for i, tr := range tenantRules {
		if tr.ID == id {
			r.byTenant[rule.TenantID] = append(tenantRules[:i], tenantRules[i+1:]...)
			break
		}
	}

	return nil
}

// DeleteExpired removes all expired IP rules
func (r *InMemoryIPRuleRepository) DeleteExpired(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	now := time.Now()

	for id, rule := range r.byID {
		if rule.ExpiresAt != nil && now.After(*rule.ExpiresAt) {
			delete(r.byID, id)
			count++
		}
	}

	// Clean up tenant maps
	for tenantID, rules := range r.byTenant {
		filtered := make([]*domain.IPRule, 0)
		for _, rule := range rules {
			if rule.ExpiresAt == nil || now.Before(*rule.ExpiresAt) {
				filtered = append(filtered, rule)
			}
		}
		r.byTenant[tenantID] = filtered
	}

	return count, nil
}
