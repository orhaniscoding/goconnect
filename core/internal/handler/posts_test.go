package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ==================== NewPostHandler Tests ====================

func TestNewPostHandler(t *testing.T) {
	t.Run("Creates Handler With Service", func(t *testing.T) {
		userRepo := repository.NewInMemoryUserRepository()
		svc := service.NewPostService(nil, userRepo)
		handler := NewPostHandler(svc)
		
		require.NotNil(t, handler)
		assert.Equal(t, svc, handler.postService)
	})
	
	t.Run("Creates Handler With Nil Service", func(t *testing.T) {
		handler := NewPostHandler(nil)
		require.NotNil(t, handler)
		assert.Nil(t, handler.postService)
	})
}

// ==================== GetPost Tests ====================

func TestPostHandler_GetPost_InvalidID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts/:id", handler.GetPost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts/invalid", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_GetPost_Success(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Post{
		ID: 1, Content: "Test post", UserID: 1,
	}, nil)
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts/:id", handler.GetPost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts/1", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_GetPost_NotFound(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.NewError(domain.ErrNotFound, "post not found", nil))
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts/:id", handler.GetPost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts/999", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

// ==================== LikePost Tests ====================

func TestPostHandler_LikePost_InvalidID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/posts/:id/like", handler.LikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/posts/invalid/like", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_LikePost_Success(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("IncrementLikes", mock.Anything, int64(1)).Return(nil)
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/posts/:id/like", handler.LikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/posts/1/like", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_LikePost_Error(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("IncrementLikes", mock.Anything, int64(1)).Return(domain.NewError(domain.ErrNotFound, "post not found", nil))
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/posts/:id/like", handler.LikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/posts/1/like", nil)
	r.ServeHTTP(w, req)
	
	// Service wraps all errors as InternalServerError
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// ==================== UnlikePost Tests ====================

func TestPostHandler_UnlikePost_InvalidID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/v1/posts/:id/like", handler.UnlikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/posts/invalid/like", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_UnlikePost_Success(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("DecrementLikes", mock.Anything, int64(1)).Return(nil)
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/v1/posts/:id/like", handler.UnlikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/posts/1/like", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_UnlikePost_Error(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("DecrementLikes", mock.Anything, int64(1)).Return(domain.NewError(domain.ErrNotFound, "post not found", nil))
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/v1/posts/:id/like", handler.UnlikePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/posts/1/like", nil)
	r.ServeHTTP(w, req)
	
	// Service wraps all errors as InternalServerError
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// ==================== DeletePost Tests ====================

func TestPostHandler_DeletePost_InvalidID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Next()
	})
	r.DELETE("/v1/posts/:id", handler.DeletePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/posts/invalid", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_DeletePost_Unauthorized(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No userID middleware
	r.DELETE("/v1/posts/:id", handler.DeletePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/posts/1", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== UpdatePost Tests ====================

func TestPostHandler_UpdatePost_InvalidID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/v1/posts/:id", handler.UpdatePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/posts/invalid", bytes.NewBufferString(`{"content":"test"}`))
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_UpdatePost_InvalidBody(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/v1/posts/:id", handler.UpdatePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/posts/1", bytes.NewBufferString(`invalid json`))
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_UpdatePost_Unauthorized(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No userID middleware
	r.PUT("/v1/posts/:id", handler.UpdatePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/v1/posts/1", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== CreatePost Tests ====================

func TestPostHandler_CreatePost_InvalidBody(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Next()
	})
	r.POST("/v1/posts", handler.CreatePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/posts", bytes.NewBufferString(`invalid json`))
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostHandler_CreatePost_Unauthorized(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(&repository.MockPostRepository{}, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No userID middleware
	r.POST("/v1/posts", handler.CreatePost)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/posts", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== Struct Tests ====================

func TestPostHandler_Struct(t *testing.T) {
	t.Run("Handler Has PostService", func(t *testing.T) {
		handler := &PostHandler{}
		assert.Nil(t, handler.postService)
	})
}

// ==================== GetPosts Tests ====================

func TestPostHandler_GetPosts_Success(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("GetAll", mock.Anything).Return([]*domain.Post{
		{ID: 1, Content: "Test post 1", UserID: 1},
		{ID: 2, Content: "Test post 2", UserID: 2},
	}, nil)
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts", handler.GetPosts)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_GetPosts_Error(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("GetAll", mock.Anything).Return(nil, domain.NewError(domain.ErrInternalServer, "database error", nil))
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts", handler.GetPosts)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_GetPosts_EmptyList(t *testing.T) {
	mockRepo := &repository.MockPostRepository{}
	mockRepo.On("GetAll", mock.Anything).Return([]*domain.Post{}, nil)
	
	userRepo := repository.NewInMemoryUserRepository()
	svc := service.NewPostService(mockRepo, userRepo)
	handler := NewPostHandler(svc)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/v1/posts", handler.GetPosts)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/posts", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}
