package repository

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockPostRepository is a mock implementation of PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	args := m.Called(ctx, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Post), args.Error(1)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	args := m.Called(ctx, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) IncrementLikes(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) DecrementLikes(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
