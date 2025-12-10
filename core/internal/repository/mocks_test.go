package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockPostRepository_Create(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	post := &domain.Post{ID: 1, Content: "Test Post", UserID: 1}

	// Test successful create
	mockRepo.On("Create", ctx, post).Return(post, nil).Once()
	result, err := mockRepo.Create(ctx, post)
	assert.NoError(t, err)
	assert.Equal(t, post, result)
	mockRepo.AssertExpectations(t)

	// Test create with error
	mockRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("create error")).Once()
	result, err = mockRepo.Create(ctx, post)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_GetAll(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	posts := []*domain.Post{{ID: 1, Content: "Post 1"}, {ID: 2, Content: "Post 2"}}

	// Test successful get all
	mockRepo.On("GetAll", ctx).Return(posts, nil).Once()
	result, err := mockRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)

	// Test get all with error
	mockRepo.On("GetAll", ctx).Return(nil, errors.New("get all error")).Once()
	result, err = mockRepo.GetAll(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_GetByID(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	post := &domain.Post{ID: 1, Content: "Test Post"}

	// Test successful get by ID
	mockRepo.On("GetByID", ctx, int64(1)).Return(post, nil).Once()
	result, err := mockRepo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, post, result)
	mockRepo.AssertExpectations(t)

	// Test get by ID not found
	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found")).Once()
	result, err = mockRepo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_Update(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	post := &domain.Post{ID: 1, Content: "Updated Post"}

	// Test successful update
	mockRepo.On("Update", ctx, post).Return(post, nil).Once()
	result, err := mockRepo.Update(ctx, post)
	assert.NoError(t, err)
	assert.Equal(t, post, result)
	mockRepo.AssertExpectations(t)

	// Test update with error
	mockRepo.On("Update", ctx, mock.Anything).Return(nil, errors.New("update error")).Once()
	result, err = mockRepo.Update(ctx, post)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_Delete(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	// Test successful delete
	mockRepo.On("Delete", ctx, int64(1)).Return(nil).Once()
	err := mockRepo.Delete(ctx, 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test delete with error
	mockRepo.On("Delete", ctx, int64(999)).Return(errors.New("delete error")).Once()
	err = mockRepo.Delete(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_IncrementLikes(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	// Test successful increment
	mockRepo.On("IncrementLikes", ctx, int64(1)).Return(nil).Once()
	err := mockRepo.IncrementLikes(ctx, 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test increment with error
	mockRepo.On("IncrementLikes", ctx, int64(999)).Return(errors.New("increment error")).Once()
	err = mockRepo.IncrementLikes(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockPostRepository_DecrementLikes(t *testing.T) {
	mockRepo := new(MockPostRepository)
	ctx := context.Background()

	// Test successful decrement
	mockRepo.On("DecrementLikes", ctx, int64(1)).Return(nil).Once()
	err := mockRepo.DecrementLikes(ctx, 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test decrement with error
	mockRepo.On("DecrementLikes", ctx, int64(999)).Return(errors.New("decrement error")).Once()
	err = mockRepo.DecrementLikes(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
