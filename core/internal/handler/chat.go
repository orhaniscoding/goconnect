package handler

import (
"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// ChatHandler handles chat HTTP requests
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) handleError(c *gin.Context, err error, message string) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		status := http.StatusBadRequest
		if domainErr.Code == domain.ErrNotFound {
			status = http.StatusNotFound
		} else if domainErr.Code == domain.ErrForbidden {
			status = http.StatusForbidden
		} else if domainErr.Code == domain.ErrUnauthorized {
			status = http.StatusUnauthorized
		}
		c.JSON(status, domainErr)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "ERR_INTERNAL_SERVER",
		"message": message,
		"details": gin.H{"error": err.Error()},
	})
}

// ListMessages handles GET /v1/chat
func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}

	// Parse query parameters
	scope := c.Query("scope")
	if scope == "" {
		scope = "host" // Default to host scope
	}

	// Parse time filters
	var sinceTime, beforeTime time.Time
	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			sinceTime = t
		}
	}
	if before := c.Query("before"); before != "" {
		if t, err := time.Parse(time.RFC3339, before); err == nil {
			beforeTime = t
		}
	}

	// Parse limit
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse cursor
	cursor := c.Query("cursor")

	// Parse include_deleted
	includeDeleted := c.Query("include_deleted") == "true"

	// Create filter
	filter := domain.ChatMessageFilter{
		Scope:          scope,
		Since:          sinceTime,
		Before:         beforeTime,
		Limit:          limit,
		Cursor:         cursor,
		IncludeDeleted: includeDeleted,
	}

	// List messages
	messages, nextCursor, err := h.chatService.ListMessages(c.Request.Context(), filter, tenantID)
	if err != nil {
		h.handleError(c, err, "Failed to list messages")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":    messages,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
	})
}

// GetMessage handles GET /v1/chat/:id
func (h *ChatHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("id")
	tenantID := c.GetString("tenant_id")

	message, err := h.chatService.GetMessage(c.Request.Context(), messageID, tenantID)
	if err != nil {
		h.handleError(c, err, "Failed to get message")
		return
	}

	c.JSON(http.StatusOK, message)
}

// SendMessage handles POST /v1/chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}

	var req struct {
		Scope       string   `json:"scope" binding:"required"`
		Body        string   `json:"body" binding:"required"`
		Attachments []string `json:"attachments"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid request body",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), userID, tenantID, req.Scope, req.Body, req.Attachments, "")
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(http.StatusBadRequest, domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to send message",
		})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// EditMessage handles PATCH /v1/chat/:id
func (h *ChatHandler) EditMessage(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	var req struct {
		Body string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid request body",
		})
		return
	}

	message, err := h.chatService.EditMessage(c.Request.Context(), messageID, userID, tenantID, req.Body, isAdmin)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			status := http.StatusBadRequest
			if domainErr.Code == domain.ErrForbidden {
				status = http.StatusForbidden
			} else if domainErr.Code == domain.ErrNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to edit message",
		})
		return
	}

	c.JSON(http.StatusOK, message)
}

// DeleteMessage handles DELETE /v1/chat/:id
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")
	isModerator := c.GetBool("is_moderator")

	mode := c.Query("mode")
	if mode == "" {
		mode = "soft" // Default to soft delete
	}

	if err := h.chatService.DeleteMessage(c.Request.Context(), messageID, userID, tenantID, mode, isAdmin, isModerator); err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			status := http.StatusBadRequest
			if domainErr.Code == domain.ErrForbidden {
				status = http.StatusForbidden
			} else if domainErr.Code == domain.ErrNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to delete message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "deleted",
		"mode":       mode,
		"message_id": messageID,
	})
}

// RedactMessage handles POST /v1/chat/:id/redact
func (h *ChatHandler) RedactMessage(c *gin.Context) {
	messageID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")
	isModerator := c.GetBool("is_moderator")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid request body",
		})
		return
	}

	redactedMsg, err := h.chatService.RedactMessage(c.Request.Context(), messageID, userID, tenantID, isAdmin, isModerator, req.Reason)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			status := http.StatusBadRequest
			if domainErr.Code == domain.ErrForbidden {
				status = http.StatusForbidden
			} else if domainErr.Code == domain.ErrNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to redact message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "redacted",
		"message_id": messageID,
		"new_body":   redactedMsg.Body,
	})
}

// GetEditHistory handles GET /v1/chat/:id/edits
func (h *ChatHandler) GetEditHistory(c *gin.Context) {
	messageID := c.Param("id")
	tenantID := c.GetString("tenant_id")

	edits, err := h.chatService.GetEditHistory(c.Request.Context(), messageID, tenantID)
	if err != nil {
		h.handleError(c, err, "Failed to get edit history")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message_id": messageID,
		"edits":      edits,
	})
}

// RegisterChatRoutes registers chat routes
func RegisterChatRoutes(r *gin.Engine, handler *ChatHandler, authMiddleware gin.HandlerFunc) {
	chat := r.Group("/v1/chat")
	chat.Use(authMiddleware) // All chat routes require authentication

	chat.GET("", handler.ListMessages)             // List messages in scope
	chat.POST("", handler.SendMessage)             // Send message
	chat.GET("/:id", handler.GetMessage)           // Get specific message
	chat.PATCH("/:id", handler.EditMessage)        // Edit message
	chat.DELETE("/:id", handler.DeleteMessage)     // Delete message (owner/admin/moderator)
	chat.GET("/:id/edits", handler.GetEditHistory) // Get edit history

	// Moderator-only routes
	chat.POST("/:id/redact", RequireModerator(), handler.RedactMessage) // Redact message (moderator/admin only)
}
