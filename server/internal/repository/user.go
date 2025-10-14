package repository

import (
	"context"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// UserRepository defines the interface for user storage operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// InMemoryUserRepository is an in-memory implementation of UserRepository
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User // id -> user
	email map[string]*domain.User // email -> user
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*domain.User),
		email: make(map[string]*domain.User),
	}
}

// Create adds a new user to the repository
func (r *InMemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.email[user.Email]; exists {
		return domain.NewError(domain.ErrEmailAlreadyExists, "Email already registered", map[string]string{"email": user.Email})
	}

	r.users[user.ID] = user
	r.email[user.Email] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": id})
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.email[email]
	if !exists {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"email": email})
	}

	return user, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": user.ID})
	}

	r.users[user.ID] = user
	r.email[user.Email] = user
	return nil
}

// Delete removes a user from the repository
func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return domain.NewError(domain.ErrUserNotFound, "User not found", map[string]string{"user_id": id})
	}

	delete(r.users, id)
	delete(r.email, user.Email)
	return nil
}
