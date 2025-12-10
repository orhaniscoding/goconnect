package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryDeletionRequestRepository_Create(t *testing.T) {
	repo := NewInMemoryDeletionRequestRepository()
	ctx := context.Background()

	req := &domain.DeletionRequest{
		ID:          "req-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: time.Now(),
	}

	// Create new request
	err := repo.Create(ctx, req)
	require.NoError(t, err)

	// Try to create duplicate
	err = repo.Create(ctx, req)
	assert.Error(t, err)
}

func TestInMemoryDeletionRequestRepository_Get(t *testing.T) {
	repo := NewInMemoryDeletionRequestRepository()
	ctx := context.Background()

	// Get non-existent
	_, err := repo.Get(ctx, "nonexistent")
	assert.Error(t, err)

	// Create and get
	req := &domain.DeletionRequest{
		ID:          "req-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, req))

	got, err := repo.Get(ctx, "req-1")
	require.NoError(t, err)
	assert.Equal(t, "user-1", got.UserID)
}

func TestInMemoryDeletionRequestRepository_GetByUserID(t *testing.T) {
	repo := NewInMemoryDeletionRequestRepository()
	ctx := context.Background()

	// Get non-existent (returns nil, nil)
	got, err := repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	assert.Nil(t, got)

	// Create and get by user ID
	req := &domain.DeletionRequest{
		ID:          "req-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, req))

	got, err = repo.GetByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "req-1", got.ID)
}

func TestInMemoryDeletionRequestRepository_ListPending(t *testing.T) {
	repo := NewInMemoryDeletionRequestRepository()
	ctx := context.Background()

	// Empty list
	pending, err := repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Empty(t, pending)

	// Create pending request
	req1 := &domain.DeletionRequest{
		ID:          "req-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, req1))

	// Create completed request
	req2 := &domain.DeletionRequest{
		ID:          "req-2",
		UserID:      "user-2",
		Status:      domain.DeletionRequestStatusCompleted,
		RequestedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, req2))

	// List pending
	pending, err = repo.ListPending(ctx)
	require.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, "req-1", pending[0].ID)
}

func TestInMemoryDeletionRequestRepository_Update(t *testing.T) {
	repo := NewInMemoryDeletionRequestRepository()
	ctx := context.Background()

	// Update non-existent
	req := &domain.DeletionRequest{
		ID:          "req-1",
		UserID:      "user-1",
		Status:      domain.DeletionRequestStatusPending,
		RequestedAt: time.Now(),
	}
	err := repo.Update(ctx, req)
	assert.Error(t, err)

	// Create and update
	require.NoError(t, repo.Create(ctx, req))

	req.Status = domain.DeletionRequestStatusCompleted
	now := time.Now()
	req.CompletedAt = &now

	err = repo.Update(ctx, req)
	require.NoError(t, err)

	// Verify update
	got, err := repo.Get(ctx, "req-1")
	require.NoError(t, err)
	assert.Equal(t, domain.DeletionRequestStatusCompleted, got.Status)
	assert.NotNil(t, got.CompletedAt)
}
