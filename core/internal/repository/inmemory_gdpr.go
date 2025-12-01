package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// InMemoryDeletionRequestRepository is an in-memory implementation
type InMemoryDeletionRequestRepository struct {
	requests map[string]*domain.DeletionRequest
	mu       sync.RWMutex
}

// NewInMemoryDeletionRequestRepository creates a new in-memory repository
func NewInMemoryDeletionRequestRepository() *InMemoryDeletionRequestRepository {
	return &InMemoryDeletionRequestRepository{
		requests: make(map[string]*domain.DeletionRequest),
	}
}

func (r *InMemoryDeletionRequestRepository) Create(ctx context.Context, req *domain.DeletionRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.requests[req.ID]; exists {
		return fmt.Errorf("request already exists")
	}

	r.requests[req.ID] = req
	return nil
}

func (r *InMemoryDeletionRequestRepository) Get(ctx context.Context, id string) (*domain.DeletionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	req, exists := r.requests[id]
	if !exists {
		return nil, fmt.Errorf("request not found")
	}
	return req, nil
}

func (r *InMemoryDeletionRequestRepository) GetByUserID(ctx context.Context, userID string) (*domain.DeletionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, req := range r.requests {
		if req.UserID == userID {
			return req, nil
		}
	}
	return nil, nil // Not found, but not an error
}

func (r *InMemoryDeletionRequestRepository) ListPending(ctx context.Context) ([]*domain.DeletionRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var pending []*domain.DeletionRequest
	for _, req := range r.requests {
		if req.Status == domain.DeletionRequestStatusPending {
			pending = append(pending, req)
		}
	}
	return pending, nil
}

func (r *InMemoryDeletionRequestRepository) Update(ctx context.Context, req *domain.DeletionRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.requests[req.ID]; !exists {
		return fmt.Errorf("request not found")
	}

	r.requests[req.ID] = req
	return nil
}
