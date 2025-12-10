package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// InMemoryAdminRepository is an in-memory implementation of AdminRepositoryInterface for testing
type InMemoryAdminRepository struct {
	mu       sync.RWMutex
	users    map[string]*domain.User
	lastSeen map[string]*time.Time // Store last_seen separately since User doesn't have this field
}

// NewInMemoryAdminRepository creates a new in-memory admin repository
func NewInMemoryAdminRepository() *InMemoryAdminRepository {
	return &InMemoryAdminRepository{
		users:    make(map[string]*domain.User),
		lastSeen: make(map[string]*time.Time),
	}
}

// Ensure InMemoryAdminRepository implements AdminRepositoryInterface
var _ AdminRepositoryInterface = (*InMemoryAdminRepository)(nil)

// AddUser adds a user to the in-memory store (for test setup)
func (r *InMemoryAdminRepository) AddUser(user *domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
}

// ListAllUsers retrieves all users with filtering and pagination
func (r *InMemoryAdminRepository) ListAllUsers(ctx context.Context, filters domain.UserFilters, pagination domain.PaginationParams) ([]*domain.UserListItem, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*domain.UserListItem

	for _, user := range r.users {
		// Apply role filter
		if filters.Role == "admin" && !user.IsAdmin {
			continue
		}
		if filters.Role == "moderator" && !user.IsModerator {
			continue
		}
		if filters.Role == "user" && (user.IsAdmin || user.IsModerator) {
			continue
		}

		// Apply status filter
		if filters.Status == "active" && user.Suspended {
			continue
		}
		if filters.Status == "suspended" && !user.Suspended {
			continue
		}

		// Apply tenant filter
		if filters.TenantID != "" && user.TenantID != filters.TenantID {
			continue
		}

		// Apply search filter
		if filters.Search != "" {
			search := strings.ToLower(filters.Search)
			emailMatch := strings.Contains(strings.ToLower(user.Email), search)
			usernameMatch := user.Username != nil && strings.Contains(strings.ToLower(*user.Username), search)
			if !emailMatch && !usernameMatch {
				continue
			}
		}

		item := &domain.UserListItem{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			TenantID:    user.TenantID,
			IsAdmin:     user.IsAdmin,
			IsModerator: user.IsModerator,
			Suspended:   user.Suspended,
			CreatedAt:   user.CreatedAt,
			LastSeen:    r.lastSeen[user.ID],
		}
		filtered = append(filtered, item)
	}

	totalCount := len(filtered)

	// Apply pagination
	if pagination.PerPage > 0 {
		start := (pagination.Page - 1) * pagination.PerPage
		if start >= len(filtered) {
			return []*domain.UserListItem{}, totalCount, nil
		}
		end := start + pagination.PerPage
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	return filtered, totalCount, nil
}

// GetUserStats retrieves system-wide user statistics
func (r *InMemoryAdminRepository) GetUserStats(ctx context.Context) (*domain.SystemStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &domain.SystemStats{}

	for _, user := range r.users {
		stats.TotalUsers++
		if user.IsAdmin {
			stats.AdminUsers++
		}
		if user.IsModerator {
			stats.ModeratorUsers++
		}
		if user.Suspended {
			stats.SuspendedUsers++
		}
	}

	return stats, nil
}

// UpdateUserRole updates a user's admin or moderator status
func (r *InMemoryAdminRepository) UpdateUserRole(ctx context.Context, userID string, isAdmin, isModerator *bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if isAdmin != nil {
		user.IsAdmin = *isAdmin
	}
	if isModerator != nil {
		user.IsModerator = *isModerator
	}
	user.UpdatedAt = time.Now().UTC()

	return nil
}

// SuspendUser suspends a user account
func (r *InMemoryAdminRepository) SuspendUser(ctx context.Context, userID, reason, suspendedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Suspended = true
	now := time.Now().UTC()
	user.SuspendedAt = &now
	user.SuspendedReason = &reason
	user.SuspendedBy = &suspendedBy
	user.UpdatedAt = now

	return nil
}

// UnsuspendUser unsuspends a user account
func (r *InMemoryAdminRepository) UnsuspendUser(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Suspended = false
	user.SuspendedAt = nil
	user.SuspendedReason = nil
	user.SuspendedBy = nil
	user.UpdatedAt = time.Now().UTC()

	return nil
}

// GetUserByID retrieves a single user by ID
func (r *InMemoryAdminRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// UpdateLastSeen updates the last_seen timestamp for a user
func (r *InMemoryAdminRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	now := time.Now().UTC()
	r.lastSeen[userID] = &now

	return nil
}
