package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// IdempotencyRepository defines the interface for idempotency key operations
type IdempotencyRepository interface {
	Get(ctx context.Context, key string) (*domain.IdempotencyRecord, error)
	Set(ctx context.Context, record *domain.IdempotencyRecord) error
	Delete(ctx context.Context, key string) error
	Cleanup(ctx context.Context) error
}

// InMemoryIdempotencyRepository provides in-memory implementation
type InMemoryIdempotencyRepository struct {
	mu      sync.RWMutex
	records map[string]*domain.IdempotencyRecord
}

// NewInMemoryIdempotencyRepository creates a new in-memory idempotency repository
func NewInMemoryIdempotencyRepository() *InMemoryIdempotencyRepository {
	repo := &InMemoryIdempotencyRepository{
		records: make(map[string]*domain.IdempotencyRecord),
	}
	
	// Start cleanup goroutine
	go repo.periodicCleanup()
	
	return repo
}

// Get retrieves an idempotency record
func (r *InMemoryIdempotencyRepository) Get(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.records[key]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "Idempotency key not found", nil)
	}

	// Check if expired
	if record.IsExpired() {
		return nil, domain.NewError(domain.ErrNotFound, "Idempotency key expired", nil)
	}

	return record, nil
}

// Set stores an idempotency record
func (r *InMemoryIdempotencyRepository) Set(ctx context.Context, record *domain.IdempotencyRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.records[record.Key] = record
	return nil
}

// Delete removes an idempotency record
func (r *InMemoryIdempotencyRepository) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.records, key)
	return nil
}

// Cleanup removes expired records
func (r *InMemoryIdempotencyRepository) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, record := range r.records {
		if now.After(record.ExpiresAt) {
			delete(r.records, key)
		}
	}

	return nil
}

// periodicCleanup runs cleanup every hour
func (r *InMemoryIdempotencyRepository) periodicCleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		r.Cleanup(context.Background())
	}
}