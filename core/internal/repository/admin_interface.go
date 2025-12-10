package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// AdminRepositoryInterface defines the interface for admin repository operations
type AdminRepositoryInterface interface {
	ListAllUsers(ctx context.Context, filters domain.UserFilters, pagination domain.PaginationParams) ([]*domain.UserListItem, int, error)
	GetUserStats(ctx context.Context) (*domain.SystemStats, error)
	UpdateUserRole(ctx context.Context, userID string, isAdmin, isModerator *bool) error
	SuspendUser(ctx context.Context, userID, reason, suspendedBy string) error
	UnsuspendUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	UpdateLastSeen(ctx context.Context, userID string) error
}

// Ensure AdminRepository implements AdminRepositoryInterface
var _ AdminRepositoryInterface = (*AdminRepository)(nil)
