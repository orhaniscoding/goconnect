package repository

import (
	"context"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// IPAMRepository defines allocation persistence
type IPAMRepository interface {
	// GetOrAllocate returns existing allocation for user or allocates next available IP.
	GetOrAllocate(ctx context.Context, networkID, userID, cidr string) (*domain.IPAllocation, error)
	// List returns all allocations for a network (for testing/introspection)
	List(ctx context.Context, networkID string) ([]*domain.IPAllocation, error)
}

type inMemoryIPAM struct {
	mu sync.Mutex
	// allocations[networkID][userID] = *IPAllocation
	allocations map[string]map[string]*domain.IPAllocation
	// nextOffset[networkID] = uint32 (next host offset to attempt)
	nextOffset map[string]uint32
}

// NewInMemoryIPAM creates a new in-memory IPAM repository
func NewInMemoryIPAM() IPAMRepository {
	return &inMemoryIPAM{allocations: make(map[string]map[string]*domain.IPAllocation), nextOffset: make(map[string]uint32)}
}

func (r *inMemoryIPAM) GetOrAllocate(ctx context.Context, networkID, userID, cidr string) (*domain.IPAllocation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.allocations[networkID]; !ok {
		r.allocations[networkID] = make(map[string]*domain.IPAllocation)
		r.nextOffset[networkID] = 1 // start from first usable host
	}
	if alloc, ok := r.allocations[networkID][userID]; ok {
		return alloc, nil
	}
	// attempt next IP
	for {
		offset := r.nextOffset[networkID]
		ip, err := domain.NextIP(cidr, offset)
		if err != nil {
			return nil, domain.NewError(domain.ErrInternalServer, "Failed to compute next IP", nil)
		}
		if ip == "" { // exhausted
			return nil, domain.NewError(domain.ErrIPExhausted, "No available IP addresses remaining", nil)
		}
		r.nextOffset[networkID] = offset + 1
		// ensure uniqueness (scan existing) - simple for in-memory
		collision := false
		for _, existing := range r.allocations[networkID] {
			if existing.IP == ip {
				collision = true
				break
			}
		}
		if collision {
			continue // try next offset
		}
		alloc := &domain.IPAllocation{NetworkID: networkID, UserID: userID, IP: ip}
		r.allocations[networkID][userID] = alloc
		return alloc, nil
	}
}

func (r *inMemoryIPAM) List(ctx context.Context, networkID string) ([]*domain.IPAllocation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*domain.IPAllocation
	for _, a := range r.allocations[networkID] {
		list = append(list, a)
	}
	return list, nil
}
