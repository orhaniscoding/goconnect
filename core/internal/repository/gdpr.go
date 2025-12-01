package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// DeletionRequestRepository handles storage of deletion requests
type DeletionRequestRepository interface {
	Create(ctx context.Context, req *domain.DeletionRequest) error
	Get(ctx context.Context, id string) (*domain.DeletionRequest, error)
	GetByUserID(ctx context.Context, userID string) (*domain.DeletionRequest, error)
	ListPending(ctx context.Context) ([]*domain.DeletionRequest, error)
	Update(ctx context.Context, req *domain.DeletionRequest) error
}
