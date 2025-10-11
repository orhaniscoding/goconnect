package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

type MembershipRepository interface {
	UpsertApproved(ctx context.Context, networkID, userID string, role domain.MembershipRole, joinedAt time.Time) (*domain.Membership, error)
	Get(ctx context.Context, networkID, userID string) (*domain.Membership, error)
	SetStatus(ctx context.Context, networkID, userID string, status domain.MembershipStatus) error
	List(ctx context.Context, networkID string, status string, limit int, cursor string) ([]*domain.Membership, string, error)
	Remove(ctx context.Context, networkID, userID string) error
}

type JoinRequestRepository interface {
	CreatePending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error)
	GetPending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error)
	Decide(ctx context.Context, id string, approve bool) error
}

type InMemoryMembershipRepository struct {
	mu        sync.RWMutex
	byKey     map[string]*domain.Membership // key: networkID|userID
	byNetwork map[string][]*domain.Membership
}

func NewInMemoryMembershipRepository() *InMemoryMembershipRepository {
	return &InMemoryMembershipRepository{byKey: map[string]*domain.Membership{}, byNetwork: map[string][]*domain.Membership{}}
}

func key(nid, uid string) string { return nid + "|" + uid }

func (r *InMemoryMembershipRepository) UpsertApproved(ctx context.Context, networkID, userID string, role domain.MembershipRole, joinedAt time.Time) (*domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(networkID, userID)
	if m, ok := r.byKey[k]; ok {
		m.Status = domain.StatusApproved
		m.Role = role
		m.JoinedAt = &joinedAt
		m.UpdatedAt = time.Now()
		return m, nil
	}
	m := &domain.Membership{ID: domain.GenerateNetworkID(), NetworkID: networkID, UserID: userID, Role: role, Status: domain.StatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.JoinedAt = &joinedAt
	r.byKey[k] = m
	r.byNetwork[networkID] = append(r.byNetwork[networkID], m)
	return m, nil
}

func (r *InMemoryMembershipRepository) Get(ctx context.Context, networkID, userID string) (*domain.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.byKey[key(networkID, userID)]; ok {
		return m, nil
	}
	return nil, domain.NewError(domain.ErrNotFound, "Membership not found", nil)
}

func (r *InMemoryMembershipRepository) SetStatus(ctx context.Context, networkID, userID string, status domain.MembershipStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(networkID, userID)
	if m, ok := r.byKey[k]; ok {
		m.Status = status
		m.UpdatedAt = time.Now()
		return nil
	}
	return domain.NewError(domain.ErrNotFound, "Membership not found", nil)
}

func (r *InMemoryMembershipRepository) List(ctx context.Context, networkID string, status string, limit int, cursor string) ([]*domain.Membership, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	arr := r.byNetwork[networkID]
	// stable sort: by JoinedAt asc then ID asc
	sort.SliceStable(arr, func(i, j int) bool {
		var ti, tj time.Time
		if arr[i].JoinedAt != nil {
			ti = *arr[i].JoinedAt
		}
		if arr[j].JoinedAt != nil {
			tj = *arr[j].JoinedAt
		}
		if !ti.Equal(tj) {
			return ti.Before(tj)
		}
		return arr[i].ID < arr[j].ID
	})
	start := 0
	if cursor != "" {
		for idx, m := range arr {
			if m.ID == cursor {
				start = idx + 1
				break
			}
		}
	}
	out := make([]*domain.Membership, 0, limit)
	for i := start; i < len(arr) && len(out) < limit; i++ {
		if status != "" && string(arr[i].Status) != status {
			continue
		}
		out = append(out, arr[i])
	}
	next := ""
	if len(out) == limit && len(out) > 0 {
		next = out[len(out)-1].ID
	}
	return out, next, nil
}

func (r *InMemoryMembershipRepository) Remove(ctx context.Context, networkID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(networkID, userID)
	m, ok := r.byKey[k]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Membership not found", nil)
	}
	delete(r.byKey, k)
	arr := r.byNetwork[networkID]
	for i, it := range arr {
		if it.ID == m.ID {
			r.byNetwork[networkID] = append(arr[:i], arr[i+1:]...)
			break
		}
	}
	return nil
}

type InMemoryJoinRequestRepository struct {
	mu    sync.RWMutex
	byKey map[string]*domain.JoinRequest // networkID|userID
	byID  map[string]*domain.JoinRequest
}

func NewInMemoryJoinRequestRepository() *InMemoryJoinRequestRepository {
	return &InMemoryJoinRequestRepository{byKey: map[string]*domain.JoinRequest{}, byID: map[string]*domain.JoinRequest{}}
}

func (r *InMemoryJoinRequestRepository) CreatePending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key(networkID, userID)
	if jr, ok := r.byKey[k]; ok && jr.Status == "pending" {
		return jr, domain.NewError(domain.ErrAlreadyRequested, "Join request already pending", nil)
	}
	jr := &domain.JoinRequest{ID: domain.GenerateNetworkID(), NetworkID: networkID, UserID: userID, Status: "pending", CreatedAt: time.Now()}
	r.byKey[k] = jr
	r.byID[jr.ID] = jr
	return jr, nil
}

func (r *InMemoryJoinRequestRepository) GetPending(ctx context.Context, networkID, userID string) (*domain.JoinRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if jr, ok := r.byKey[key(networkID, userID)]; ok && jr.Status == "pending" {
		return jr, nil
	}
	return nil, domain.NewError(domain.ErrNotFound, "No pending join request", nil)
}

func (r *InMemoryJoinRequestRepository) Decide(ctx context.Context, id string, approve bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	jr, ok := r.byID[id]
	if !ok {
		return domain.NewError(domain.ErrNotFound, "Join request not found", nil)
	}
	if jr.Status != "pending" {
		return nil
	}
	jr.Status = map[bool]string{true: "approved", false: "denied"}[approve]
	now := time.Now()
	jr.DecidedAt = &now
	return nil
}
