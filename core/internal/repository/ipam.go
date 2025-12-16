package repository

import (
	"context"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// IPAMRepository defines allocation persistence
type IPAMRepository interface {
	// GetOrAllocate returns existing allocation for subject (user or device) or allocates next available IP.
	// subjectID can be either a user ID (legacy) or device ID (preferred for device-based allocation).
	GetOrAllocate(ctx context.Context, networkID, subjectID, cidr string) (*domain.IPAllocation, error)
	// List returns all allocations for a network (for testing/introspection)
	List(ctx context.Context, networkID string) ([]*domain.IPAllocation, error)
	// Release removes a subject's allocation (idempotent). Implementation may keep a free list for reuse.
	Release(ctx context.Context, networkID, subjectID string) error
}

type inMemoryIPAM struct {
	mu sync.Mutex
	// allocations[networkID][userID] = *IPAllocation
	allocations map[string]map[string]*domain.IPAllocation
	// nextOffset[networkID] = uint32 (next host offset to attempt)
	nextOffset map[string]uint32
	// freeOffsets[networkID] = slice of previously used offsets available for reuse (LIFO for cache locality)
	freeOffsets map[string][]uint32
}

// NewInMemoryIPAM creates a new in-memory IPAM repository
func NewInMemoryIPAM() IPAMRepository {
	return &inMemoryIPAM{
		allocations: make(map[string]map[string]*domain.IPAllocation),
		nextOffset:  make(map[string]uint32),
		freeOffsets: make(map[string][]uint32),
	}
}

func (r *inMemoryIPAM) GetOrAllocate(ctx context.Context, networkID, subjectID, cidr string) (*domain.IPAllocation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.allocations[networkID]; !ok {
		r.allocations[networkID] = make(map[string]*domain.IPAllocation)
		r.nextOffset[networkID] = 1 // start from first usable host
	}
	if alloc, ok := r.allocations[networkID][subjectID]; ok {
		return alloc, nil
	}
	// attempt reuse from free list first (LIFO)
	for {
		var offset uint32
		if free := r.freeOffsets[networkID]; len(free) > 0 {
			offset = free[len(free)-1]
			r.freeOffsets[networkID] = free[:len(free)-1]
		} else {
			offset = r.nextOffset[networkID]
			r.nextOffset[networkID] = offset + 1
		}
		ip, err := domain.NextIP(cidr, offset)
		if err != nil {
			return nil, domain.NewError(domain.ErrInternalServer, "Failed to compute next IP", nil)
		}
		if ip == "" { // exhausted
			return nil, domain.NewError(domain.ErrIPExhausted, "No available IP addresses remaining", nil)
		}
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
		// Use DeviceID for new allocations (subjectID is device ID when device-based)
		alloc := &domain.IPAllocation{NetworkID: networkID, DeviceID: subjectID, IP: ip, Offset: offset}
		r.allocations[networkID][subjectID] = alloc
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

// Release implements idempotent release; if allocation exists it is deleted and its offset appended to free list.
// If subject had no allocation, it is a no-op (idempotent success). Offsets are reused in LIFO order to favor cache locality
// and quick recycling; this behavior is suitable for small pools and predictable reuse in tests.
func (r *inMemoryIPAM) Release(ctx context.Context, networkID, subjectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subjects, ok := r.allocations[networkID]
	if !ok {
		return nil
	}
	alloc, ok := subjects[subjectID]
	if !ok {
		return nil
	}
	// reclaim offset
	if alloc.Offset > 0 { // offset starts at 1
		r.freeOffsets[networkID] = append(r.freeOffsets[networkID], alloc.Offset)
	}
	delete(subjects, subjectID)
	return nil
}
