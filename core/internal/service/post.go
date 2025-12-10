package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// PostService handles business logic for posts
type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

// NewPostService creates a new post service
func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(ctx context.Context, req *domain.CreatePostRequest) (*domain.Post, error) {
	if req.Content == "" {
		return nil, domain.NewError(domain.ErrInvalidRequest, "Content is required", nil)
	}

	now := time.Now()
	post := &domain.Post{
		UserID:    req.UserID,
		Content:   req.Content,
		ImageURL:  req.ImageURL,
		Likes:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdPost, err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to create post", nil)
	}

	// Load author information
	author, err := s.userRepo.GetByID(ctx, fmt.Sprintf("%d", createdPost.UserID))
	if err == nil {
		createdPost.Author = author
	}

	return createdPost, nil
}

// GetPosts retrieves all posts
func (s *PostService) GetPosts(ctx context.Context) ([]*domain.Post, error) {
	posts, err := s.postRepo.GetAll(ctx)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to retrieve posts", nil)
	}

	return posts, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, postID int64) (*domain.Post, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, domain.NewError(domain.ErrNotFound, "Post not found", nil)
	}

	// Load author information
	author, err := s.userRepo.GetByID(ctx, fmt.Sprintf("%d", post.UserID))
	if err == nil {
		post.Author = author
	}

	return post, nil
}

// UpdatePost updates an existing post
func (s *PostService) UpdatePost(ctx context.Context, req *domain.UpdatePostRequest) (*domain.Post, error) {
	// Check if post exists and belongs to user
	existingPost, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, domain.NewError(domain.ErrNotFound, "Post not found", nil)
	}

	if existingPost.UserID != req.UserID {
		return nil, domain.NewError(domain.ErrForbidden, "You can only update your own posts", nil)
	}

	existingPost.Content = req.Content
	existingPost.ImageURL = req.ImageURL
	existingPost.UpdatedAt = time.Now()

	updatedPost, err := s.postRepo.Update(ctx, existingPost)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to update post", nil)
	}

	// Load author information
	author, err := s.userRepo.GetByID(ctx, fmt.Sprintf("%d", updatedPost.UserID))
	if err == nil {
		updatedPost.Author = author
	}

	return updatedPost, nil
}

// DeletePost deletes a post
func (s *PostService) DeletePost(ctx context.Context, postID, userID int64) error {
	// Check if post exists and belongs to user
	existingPost, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return domain.NewError(domain.ErrNotFound, "Post not found", nil)
	}

	if existingPost.UserID != userID {
		return domain.NewError(domain.ErrForbidden, "You can only delete your own posts", nil)
	}

	err = s.postRepo.Delete(ctx, postID)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to delete post", nil)
	}

	return nil
}

// LikePost increments the like count of a post
func (s *PostService) LikePost(ctx context.Context, postID int64) error {
	err := s.postRepo.IncrementLikes(ctx, postID)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to like post", nil)
	}

	return nil
}

// UnlikePost decrements the like count of a post
func (s *PostService) UnlikePost(ctx context.Context, postID int64) error {
	err := s.postRepo.DecrementLikes(ctx, postID)
	if err != nil {
		return domain.NewError(domain.ErrInternalServer, "Failed to unlike post", nil)
	}

	return nil
}
