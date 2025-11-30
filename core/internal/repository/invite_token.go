package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// InviteTokenRepository defines the interface for invite token operations
type InviteTokenRepository interface {
	Create(ctx context.Context, token *domain.InviteToken) error
	GetByID(ctx context.Context, id string) (*domain.InviteToken, error)
	GetByToken(ctx context.Context, token string) (*domain.InviteToken, error)
	ListByNetwork(ctx context.Context, networkID string) ([]*domain.InviteToken, error)
	UseToken(ctx context.Context, token string) (*domain.InviteToken, error)
	Revoke(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int, error)
}

// InMemoryInviteTokenRepository is an in-memory implementation
type InMemoryInviteTokenRepository struct {
	mu        sync.RWMutex
	byID      map[string]*domain.InviteToken
	byToken   map[string]*domain.InviteToken
	byNetwork map[string][]*domain.InviteToken
}

// NewInMemoryInviteTokenRepository creates a new in-memory invite token repository
func NewInMemoryInviteTokenRepository() *InMemoryInviteTokenRepository {
	return &InMemoryInviteTokenRepository{
		byID:      make(map[string]*domain.InviteToken),
		byToken:   make(map[string]*domain.InviteToken),
		byNetwork: make(map[string][]*domain.InviteToken),
	}
}

// Create stores a new invite token
func (r *InMemoryInviteTokenRepository) Create(ctx context.Context, token *domain.InviteToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byID[token.ID]; exists {
		return domain.NewError(domain.ErrConflict, "Invite token ID already exists", nil)
	}
	if _, exists := r.byToken[token.Token]; exists {
		return domain.NewError(domain.ErrConflict, "Invite token already exists", nil)
	}

	r.byID[token.ID] = token
	r.byToken[token.Token] = token
	r.byNetwork[token.NetworkID] = append(r.byNetwork[token.NetworkID], token)

	return nil
}

// GetByID retrieves an invite token by its ID
func (r *InMemoryInviteTokenRepository) GetByID(ctx context.Context, id string) (*domain.InviteToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, exists := r.byID[id]
	if !exists {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}
	return token, nil
}

// GetByToken retrieves an invite token by the token string
func (r *InMemoryInviteTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.InviteToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	token, exists := r.byToken[tokenStr]
	if !exists {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}
	return token, nil
}

// ListByNetwork lists all invite tokens for a network
func (r *InMemoryInviteTokenRepository) ListByNetwork(ctx context.Context, networkID string) ([]*domain.InviteToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokens := r.byNetwork[networkID]
	result := make([]*domain.InviteToken, 0, len(tokens))
	for _, t := range tokens {
		// Only return non-expired, non-revoked tokens
		if t.RevokedAt == nil && time.Now().Before(t.ExpiresAt) {
			result = append(result, t)
		}
	}
	return result, nil
}

// UseToken decrements the uses_left counter and returns the token
func (r *InMemoryInviteTokenRepository) UseToken(ctx context.Context, tokenStr string) (*domain.InviteToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, exists := r.byToken[tokenStr]
	if !exists {
		return nil, domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}

	if !token.IsValid() {
		if token.RevokedAt != nil {
			return nil, domain.NewError(domain.ErrInviteTokenRevoked, "Invite token has been revoked", nil)
		}
		return nil, domain.NewError(domain.ErrInviteTokenExpired, "Invite token has expired or reached max uses", nil)
	}

	if err := token.DecrementUse(); err != nil {
		return nil, err
	}

	return token, nil
}

// Revoke marks an invite token as revoked
func (r *InMemoryInviteTokenRepository) Revoke(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, exists := r.byID[id]
	if !exists {
		return domain.NewError(domain.ErrInviteTokenNotFound, "Invite token not found", nil)
	}

	now := time.Now()
	token.RevokedAt = &now
	return nil
}

// DeleteExpired removes all expired tokens from memory
func (r *InMemoryInviteTokenRepository) DeleteExpired(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	now := time.Now()

	for id, token := range r.byID {
		if now.After(token.ExpiresAt) {
			delete(r.byID, id)
			delete(r.byToken, token.Token)
			count++
		}
	}

	// Also clean up byNetwork slices
	for networkID, tokens := range r.byNetwork {
		filtered := make([]*domain.InviteToken, 0)
		for _, t := range tokens {
			if now.Before(t.ExpiresAt) {
				filtered = append(filtered, t)
			}
		}
		r.byNetwork[networkID] = filtered
	}

	return count, nil
}
