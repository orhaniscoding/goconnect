package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// NetworkRepository defines the interface for network data operations
type NetworkRepository interface {
	Create(ctx context.Context, network *domain.Network) error
	GetByID(ctx context.Context, id string) (*domain.Network, error)
	List(ctx context.Context, filter NetworkFilter) ([]*domain.Network, string, error)
	CheckCIDROverlap(ctx context.Context, cidr string, excludeID string) (bool, error)
	Update(ctx context.Context, id string, mutate func(n *domain.Network) error) (*domain.Network, error)
	SoftDelete(ctx context.Context, id string, at time.Time) error
}

// NetworkFilter represents filtering options for listing networks
type NetworkFilter struct {
	Visibility string // public|mine|all
	UserID     string // for "mine" filtering
	IsAdmin    bool   // for "all" filtering
	Search     string
	Limit      int
	Cursor     string
}

// InMemoryNetworkRepository provides in-memory implementation for development
type InMemoryNetworkRepository struct {
	mu       sync.RWMutex
	networks map[string]*domain.Network
	byTenant map[string][]*domain.Network
}

// NewInMemoryNetworkRepository creates a new in-memory network repository
func NewInMemoryNetworkRepository() *InMemoryNetworkRepository {
	return &InMemoryNetworkRepository{
		networks: make(map[string]*domain.Network),
		byTenant: make(map[string][]*domain.Network),
	}
}

// Create stores a new network
func (r *InMemoryNetworkRepository) Create(ctx context.Context, network *domain.Network) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if network with same name exists for tenant
	for _, existing := range r.byTenant[network.TenantID] {
		if existing.Name == network.Name && existing.SoftDeletedAt == nil {
			return domain.NewError(domain.ErrInvalidRequest,
				fmt.Sprintf("Network with name '%s' already exists", network.Name),
				map[string]string{"field": "name"})
		}
	}

	// Store the network
	r.networks[network.ID] = network
	r.byTenant[network.TenantID] = append(r.byTenant[network.TenantID], network)

	return nil
}

// GetByID retrieves a network by ID
func (r *InMemoryNetworkRepository) GetByID(ctx context.Context, id string) (*domain.Network, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	network, exists := r.networks[id]
	if !exists || network.SoftDeletedAt != nil {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}

	return network, nil
}

// List retrieves networks based on filter criteria
func (r *InMemoryNetworkRepository) List(ctx context.Context, filter NetworkFilter) ([]*domain.Network, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Network
	var startFound bool

	// If cursor is provided, find starting point
	if filter.Cursor == "" {
		startFound = true
	}

	for _, network := range r.networks {
		// Skip soft-deleted networks
		if network.SoftDeletedAt != nil {
			continue
		}

		// Handle cursor-based pagination
		if !startFound {
			if network.ID == filter.Cursor {
				startFound = true
			}
			continue
		}

		// Apply visibility filter
		if !r.matchesVisibilityFilter(network, filter) {
			continue
		}

		// Apply search filter
		if filter.Search != "" && !strings.Contains(strings.ToLower(network.Name), strings.ToLower(filter.Search)) {
			continue
		}

		result = append(result, network)

		// Limit results
		if len(result) >= filter.Limit {
			break
		}
	}

	// Generate next cursor
	var nextCursor string
	if len(result) == filter.Limit && len(result) > 0 {
		nextCursor = result[len(result)-1].ID
	}

	return result, nextCursor, nil
}

// CheckCIDROverlap checks if the given CIDR overlaps with existing networks
func (r *InMemoryNetworkRepository) CheckCIDROverlap(ctx context.Context, cidr string, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, network := range r.networks {
		// Skip soft-deleted networks and excluded network
		if network.SoftDeletedAt != nil || network.ID == excludeID {
			continue
		}

		overlap, err := domain.CheckCIDROverlap(cidr, network.CIDR)
		if err != nil {
			return false, fmt.Errorf("error checking CIDR overlap: %w", err)
		}

		if overlap {
			return true, nil
		}
	}

	return false, nil
}

// matchesVisibilityFilter checks if network matches visibility filter
func (r *InMemoryNetworkRepository) matchesVisibilityFilter(network *domain.Network, filter NetworkFilter) bool {
	switch filter.Visibility {
	case "public":
		return network.Visibility == domain.NetworkVisibilityPublic
	case "mine":
		return network.CreatedBy == filter.UserID
	case "all":
		return filter.IsAdmin // Only admins can see all networks
	default:
		return network.Visibility == domain.NetworkVisibilityPublic
	}
}

// Update mutates a network atomically applying validation (name uniqueness)
func (r *InMemoryNetworkRepository) Update(ctx context.Context, id string, mutate func(n *domain.Network) error) (*domain.Network, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.networks[id]
	if !ok || n.SoftDeletedAt != nil {
		return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}
	originalName := n.Name
	if err := mutate(n); err != nil {
		return nil, err
	}
	if n.Name != originalName { // enforce uniqueness
		for _, ex := range r.byTenant[n.TenantID] {
			if ex != n && ex.SoftDeletedAt == nil && ex.Name == n.Name {
				return nil, domain.NewError(domain.ErrInvalidRequest, fmt.Sprintf("Network with name '%s' already exists", n.Name), map[string]string{"field": "name"})
			}
		}
	}
	n.UpdatedAt = time.Now()
	return n, nil
}

// SoftDelete marks a network as deleted (soft) so it is excluded from listings
func (r *InMemoryNetworkRepository) SoftDelete(ctx context.Context, id string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.networks[id]
	if !ok || n.SoftDeletedAt != nil {
		return domain.NewError(domain.ErrNotFound, "Network not found", nil)
	}
	n.SoftDeletedAt = &at
	n.UpdatedAt = at
	return nil
}
