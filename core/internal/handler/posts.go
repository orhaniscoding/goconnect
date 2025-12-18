package handler

import (
"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService *service.PostService
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost handles POST /v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req domain.CreatePostRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	req.UserID = userID.(int64)

	post, err := h.postService.CreatePost(c.Request.Context(), &req)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": post,
	})
}

// GetPosts handles GET /v1/posts
func (h *PostHandler) GetPosts(c *gin.Context) {
	posts, err := h.postService.GetPosts(c.Request.Context())
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": posts,
	})
}

// GetPost handles GET /v1/posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid post ID", nil))
		return
	}

	post, err := h.postService.GetPost(c.Request.Context(), postID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// UpdatePost handles PUT /v1/posts/:id
func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid post ID", nil))
		return
	}

	var req domain.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	req.ID = postID
	req.UserID = userID.(int64)

	post, err := h.postService.UpdatePost(c.Request.Context(), &req)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": post,
	})
}

// DeletePost handles DELETE /v1/posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid post ID", nil))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	err = h.postService.DeletePost(c.Request.Context(), postID, userID.(int64))
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// LikePost handles POST /v1/posts/:id/like
func (h *PostHandler) LikePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid post ID", nil))
		return
	}

	err = h.postService.LikePost(c.Request.Context(), postID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post liked successfully",
	})
}

// UnlikePost handles DELETE /v1/posts/:id/like
func (h *PostHandler) UnlikePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid post ID", nil))
		return
	}

	err = h.postService.UnlikePost(c.Request.Context(), postID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post unliked successfully",
	})
}
