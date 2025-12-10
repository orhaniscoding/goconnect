package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPostRepository is a test double for PostRepository
type mockPostRepository struct {
	mu        sync.RWMutex
	posts     map[int64]*domain.Post
	nextID    int64
	createErr error
	getErr    error
	getAllErr error
	updateErr error
	deleteErr error
	likeErr   error
	unlikeErr error
}

func newMockPostRepository() *mockPostRepository {
	return &mockPostRepository{
		posts:  make(map[int64]*domain.Post),
		nextID: 1,
	}
}

func (m *mockPostRepository) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	post.ID = m.nextID
	m.nextID++
	now := time.Now()
	if post.CreatedAt.IsZero() {
		post.CreatedAt = now
	}
	if post.UpdatedAt.IsZero() {
		post.UpdatedAt = now
	}
	m.posts[post.ID] = post
	return post, nil
}

func (m *mockPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	posts := make([]*domain.Post, 0, len(m.posts))
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (m *mockPostRepository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	post, exists := m.posts[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	// Return a copy to avoid pointer issues in tests
	p := *post
	return &p, nil
}

func (m *mockPostRepository) Update(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.posts[post.ID]; !exists {
		return nil, errors.New("post not found")
	}
	post.UpdatedAt = time.Now()
	m.posts[post.ID] = post
	return post, nil
}

func (m *mockPostRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.posts[id]; !exists {
		return errors.New("post not found")
	}
	delete(m.posts, id)
	return nil
}

func (m *mockPostRepository) IncrementLikes(ctx context.Context, id int64) error {
	if m.likeErr != nil {
		return m.likeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	post, exists := m.posts[id]
	if !exists {
		return errors.New("post not found")
	}
	post.Likes++
	return nil
}

func (m *mockPostRepository) DecrementLikes(ctx context.Context, id int64) error {
	if m.unlikeErr != nil {
		return m.unlikeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	post, exists := m.posts[id]
	if !exists {
		return errors.New("post not found")
	}
	if post.Likes > 0 {
		post.Likes--
	}
	return nil
}

// setupPostServiceTest creates a PostService with mock repositories
func setupPostServiceTest() (*PostService, *mockPostRepository, *repository.InMemoryUserRepository) {
	postRepo := newMockPostRepository()
	userRepo := repository.NewInMemoryUserRepository()
	svc := NewPostService(postRepo, userRepo)
	return svc, postRepo, userRepo
}

// ==================== CreatePost Tests ====================

func TestPostService_CreatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, userRepo := setupPostServiceTest()
		ctx := context.Background()

		// Setup user
		username := "testuser"
		user := &domain.User{ID: "1", Email: "test@example.com", Username: &username}
		require.NoError(t, userRepo.Create(ctx, user))

		req := &domain.CreatePostRequest{
			UserID:  1,
			Content: "Hello world",
		}

		post, err := svc.CreatePost(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, "Hello world", post.Content)
		assert.Equal(t, int64(1), post.UserID)
		assert.NotNil(t, post.Author)
		assert.Equal(t, "testuser", *post.Author.Username)
	})

	t.Run("Empty Content Fails", func(t *testing.T) {
		svc, _, _ := setupPostServiceTest()
		ctx := context.Background()

		req := &domain.CreatePostRequest{
			UserID:  1,
			Content: "",
		}

		_, err := svc.CreatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInvalidRequest, err.(*domain.Error).Code)
	})

	t.Run("Repo Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.createErr = errors.New("db error")
		ctx := context.Background()

		req := &domain.CreatePostRequest{
			UserID:  1,
			Content: "content",
		}

		_, err := svc.CreatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInternalServer, err.(*domain.Error).Code)
	})
}

// ==================== GetPosts Tests ====================

func TestPostService_GetPosts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		// Add dummy posts
		postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Post 1"})
		postRepo.Create(ctx, &domain.Post{UserID: 2, Content: "Post 2"})

		posts, err := svc.GetPosts(ctx)
		require.NoError(t, err)
		assert.Len(t, posts, 2)
	})

	t.Run("Repo Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.getAllErr = errors.New("db error")
		ctx := context.Background()

		_, err := svc.GetPosts(ctx)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInternalServer, err.(*domain.Error).Code)
	})
}

// ==================== GetPost Tests ====================

func TestPostService_GetPost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, postRepo, userRepo := setupPostServiceTest()
		ctx := context.Background()

		// Setup user & post
		user := &domain.User{ID: "1", Email: "test@example.com"}
		userRepo.Create(ctx, user)
		
		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Post 1"})

		post, err := svc.GetPost(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, post.ID)
		assert.NotNil(t, post.Author)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, _, _ := setupPostServiceTest()
		ctx := context.Background()

		_, err := svc.GetPost(ctx, 999)
		require.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err.(*domain.Error).Code)
	})
}

// ==================== UpdatePost Tests ====================

func TestPostService_UpdatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, postRepo, userRepo := setupPostServiceTest()
		ctx := context.Background()

		user := &domain.User{ID: "1", Email: "test@example.com"}
		userRepo.Create(ctx, user)

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Original"})

		req := &domain.UpdatePostRequest{
			ID:      created.ID,
			UserID:  1,
			Content: "Updated",
		}

		updated, err := svc.UpdatePost(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Content)
	})

	t.Run("Forbidden", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Original"})

		req := &domain.UpdatePostRequest{
			ID:      created.ID,
			UserID:  2, // Different user
			Content: "Updated",
		}

		_, err := svc.UpdatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrForbidden, err.(*domain.Error).Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, _, _ := setupPostServiceTest()
		ctx := context.Background()

		req := &domain.UpdatePostRequest{ID: 999, UserID: 1, Content: "Updated"}
		_, err := svc.UpdatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err.(*domain.Error).Code)
	})
}

// ==================== DeletePost Tests ====================

func TestPostService_DeletePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "To delete"})

		err := svc.DeletePost(ctx, created.ID, 1)
		require.NoError(t, err)

		_, err = postRepo.GetByID(ctx, created.ID)
		assert.Error(t, err)
	})

	t.Run("Forbidden", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "To delete"})

		err := svc.DeletePost(ctx, created.ID, 2)
		require.Error(t, err)
		assert.Equal(t, domain.ErrForbidden, err.(*domain.Error).Code)
	})
}

// ==================== Like/Unlike Tests ====================

func TestPostService_Likes(t *testing.T) {
	t.Run("Like Success", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Like me"})

		err := svc.LikePost(ctx, created.ID)
		require.NoError(t, err)

		updated, _ := postRepo.GetByID(ctx, created.ID)
		assert.Equal(t, 1, updated.Likes)
	})

	t.Run("Unlike Success", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Unlike me", Likes: 1})

		err := svc.UnlikePost(ctx, created.ID)
		require.NoError(t, err)

		updated, _ := postRepo.GetByID(ctx, created.ID)
		assert.Equal(t, 0, updated.Likes)
	})
	
	t.Run("Like Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.likeErr = errors.New("db error")
		ctx := context.Background()
		
		err := svc.LikePost(ctx, 1) // id doesn't matter for mock error
		require.Error(t, err)
	})

	t.Run("Unlike Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.unlikeErr = errors.New("db error")
		ctx := context.Background()

		err := svc.UnlikePost(ctx, 1)
		require.Error(t, err)
	})
}

// ==================== Author Lookup Tests ====================

func TestPostService_AuthorLookupFailure(t *testing.T) {
	t.Run("CreatePost Only Warns On Author Lookup Fail", func(t *testing.T) {
		svc, _, _ := setupPostServiceTest()
		// We don't create the user in userRepo, so lookup will fail
		
		req := &domain.CreatePostRequest{UserID: 1, Content: "No author"}
		post, err := svc.CreatePost(context.Background(), req)
		require.NoError(t, err)
		assert.Nil(t, post.Author) // Should be nil because lookup failed
	})

	t.Run("GetPost Only Warns On Author Lookup Fail", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		created, _ := postRepo.Create(context.Background(), &domain.Post{UserID: 1, Content: "No author"})
		
		post, err := svc.GetPost(context.Background(), created.ID)
		require.NoError(t, err)
		assert.Nil(t, post.Author)
	})

	t.Run("UpdatePost Only Warns On Author Lookup Fail", func(t *testing.T) {
		svc, postRepo, userRepo := setupPostServiceTest()
		// Make sure initial user exists so we can update
		userRepo.Create(context.Background(), &domain.User{ID: "1", Email: "@"})
		
		created, _ := postRepo.Create(context.Background(), &domain.Post{UserID: 1, Content: "Original"})
		
		// Delete user to cause lookup fail during Update
		userRepo.Delete(context.Background(), "1")

		req := &domain.UpdatePostRequest{ID: created.ID, UserID: 1, Content: "Updated"}
		updated, err := svc.UpdatePost(context.Background(), req)
		require.NoError(t, err)
		assert.Nil(t, updated.Author)
	})
}

// ==================== Update/Delete Error Tests ====================

func TestPostService_RepoErrors(t *testing.T) {
	t.Run("UpdateRepo Error", func(t *testing.T) {
		svc, postRepo, userRepo := setupPostServiceTest()
		postRepo.updateErr = errors.New("db error")
		ctx := context.Background()

		// User needs to exist for ownership check
		userRepo.Create(ctx, &domain.User{ID: "1", Email: "@"})
		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "Original"})

		req := &domain.UpdatePostRequest{ID: created.ID, UserID: 1, Content: "Updated"}
		_, err := svc.UpdatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInternalServer, err.(*domain.Error).Code)
	})

	t.Run("DeleteRepo Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.deleteErr = errors.New("db error")
		ctx := context.Background()

		created, _ := postRepo.Create(ctx, &domain.Post{UserID: 1, Content: "To delete"})
		
		err := svc.DeletePost(ctx, created.ID, 1)
		require.Error(t, err)
		assert.Equal(t, domain.ErrInternalServer, err.(*domain.Error).Code)
	})
	
	t.Run("DeleteRepo Get Post Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.getErr = errors.New("db error")
		ctx := context.Background()

		err := svc.DeletePost(ctx, 1, 1)
		require.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err.(*domain.Error).Code)
	})
	
	t.Run("UpdateRepo Get Post Error", func(t *testing.T) {
		svc, postRepo, _ := setupPostServiceTest()
		postRepo.getErr = errors.New("db error")
		ctx := context.Background()

		req := &domain.UpdatePostRequest{ID: 1, UserID: 1}
		_, err := svc.UpdatePost(ctx, req)
		require.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err.(*domain.Error).Code)
	})
}
