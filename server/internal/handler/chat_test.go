package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChatTest() (*gin.Engine, *ChatHandler, *service.ChatService, repository.UserRepository) {
	gin.SetMode(gin.TestMode)

	// Setup repositories
	chatRepo := repository.NewInMemoryChatRepository()
	userRepo := repository.NewInMemoryUserRepository()

	// Create test user
	testUser := &domain.User{
		ID:       "user-123",
		TenantID: "tenant-1",
		Email:    "test@example.com",
		PasswordHash: "dummy",
	}
	userRepo.Create(context.Background(), testUser)

	// Setup service
	chatService := service.NewChatService(chatRepo, userRepo)

	// Setup handler
	handler := NewChatHandler(chatService)

	// Setup router
	r := gin.New()

	return r, handler, chatService, userRepo
}

func TestChatHandler_SendMessage(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	r.POST("/v1/chat", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.SendMessage(c)
	})

	t.Run("Success - send message", func(t *testing.T) {
		body := map[string]interface{}{
			"scope": "host",
			"body":  "Hello, world!",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "user-123", response["user_id"])
		assert.Equal(t, "host", response["scope"])
		assert.Equal(t, "Hello, world!", response["body"])
		assert.False(t, response["redacted"].(bool))
	})

	t.Run("Success - send message with attachments", func(t *testing.T) {
		body := map[string]interface{}{
			"scope":       "network:123",
			"body":        "Check these files",
			"attachments": []string{"file1.pdf", "file2.png"},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		attachments := response["attachments"].([]interface{})
		assert.Len(t, attachments, 2)
		assert.Equal(t, "file1.pdf", attachments[0])
	})

	t.Run("Unauthorized - missing user_id", func(t *testing.T) {
		r2 := gin.New()
		r2.POST("/v1/chat", handler.SendMessage) // No user_id in context

		body := map[string]interface{}{
			"scope": "host",
			"body":  "Test",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Validation - empty body", func(t *testing.T) {
		body := map[string]interface{}{
			"scope": "host",
			"body":  "",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/chat", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Verify message was actually saved
	t.Run("Verify message saved", func(t *testing.T) {
		messages, _, err := chatService.ListMessages(context.TODO(), domain.ChatMessageFilter{
			Scope:    "host",
			TenantID: "tenant-1",
			Limit:    10,
		}, "tenant-1")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(messages), 1)
	})
}

func TestChatHandler_GetMessage(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	// Create test message
	msg, _ := chatService.SendMessage(context.TODO(),
		"user-123", "tenant-1", "host", "Test message", nil)

	r.GET("/v1/chat/:id", func(c *gin.Context) {
		c.Set("tenant_id", "tenant-1")
		handler.GetMessage(c)
	})

	t.Run("Success - get message", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/v1/chat/%s", msg.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, msg.ID, response["id"])
		assert.Equal(t, "Test message", response["body"])
	})

	t.Run("Not found - invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/chat/non-existent-id", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestChatHandler_ListMessages(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	// Create test messages
	ctx := context.TODO()
	chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Message 1", nil)
	time.Sleep(10 * time.Millisecond)
	chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Message 2", nil)
	time.Sleep(10 * time.Millisecond)
	chatService.SendMessage(ctx, "user-123", "tenant-1", "network:123", "Network message", nil)

	r.GET("/v1/chat", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.ListMessages(c)
	})

	t.Run("Success - list all host messages", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/chat?scope=host", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		messages := response["messages"].([]interface{})
		assert.Len(t, messages, 2) // Only host messages
		assert.False(t, response["has_more"].(bool))
	})

	t.Run("Success - filter by scope", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/chat?scope=network:123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		messages := response["messages"].([]interface{})
		assert.Len(t, messages, 1)
	})

	t.Run("Success - pagination with limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/chat?scope=host&limit=1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		messages := response["messages"].([]interface{})
		assert.Len(t, messages, 1)
		assert.True(t, response["has_more"].(bool))
		assert.NotEmpty(t, response["next_cursor"])
	})

	t.Run("Success - default scope (host)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/chat", nil) // No scope param
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unauthorized - missing user_id", func(t *testing.T) {
		r2 := gin.New()
		r2.GET("/v1/chat", handler.ListMessages) // No user_id

		req := httptest.NewRequest("GET", "/v1/chat", nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChatHandler_EditMessage(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	// Create test message
	ctx := context.TODO()
	msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Original text", nil)

	r.PATCH("/v1/chat/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		c.Set("is_admin", false)
		handler.EditMessage(c)
	})

	t.Run("Success - owner edits message", func(t *testing.T) {
		body := map[string]interface{}{
			"body": "Edited text",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/chat/%s", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "Edited text", response["body"])
	})

	t.Run("Success - admin edits any message", func(t *testing.T) {
		r2 := gin.New()
		r2.PATCH("/v1/chat/:id", func(c *gin.Context) {
			c.Set("user_id", "admin-456")
			c.Set("tenant_id", "tenant-1")
			c.Set("is_admin", true)
			handler.EditMessage(c)
		})

		body := map[string]interface{}{
			"body": "Admin edit",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/chat/%s", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation - empty body", func(t *testing.T) {
		body := map[string]interface{}{
			"body": "",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/chat/%s", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not found - invalid message ID", func(t *testing.T) {
		body := map[string]interface{}{
			"body": "Edit",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/chat/non-existent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestChatHandler_DeleteMessage(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	r.DELETE("/v1/chat/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		c.Set("is_admin", false)
		c.Set("is_moderator", false)
		handler.DeleteMessage(c)
	})

	t.Run("Success - soft delete by owner", func(t *testing.T) {
		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "To delete", nil)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/v1/chat/%s?mode=soft", msg.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "deleted", response["status"])
		assert.Equal(t, "soft", response["mode"])
	})

	t.Run("Success - hard delete by owner", func(t *testing.T) {
		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "To delete", nil)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/v1/chat/%s?mode=hard", msg.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success - moderator deletes any message", func(t *testing.T) {
		r2 := gin.New()
		r2.DELETE("/v1/chat/:id", func(c *gin.Context) {
			c.Set("user_id", "mod-789")
			c.Set("tenant_id", "tenant-1")
			c.Set("is_admin", false)
			c.Set("is_moderator", true)
			handler.DeleteMessage(c)
		})

		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Test", nil)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/v1/chat/%s?mode=soft", msg.ID), nil)
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not found - invalid message ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/chat/non-existent?mode=soft", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestChatHandler_RedactMessage(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	r.POST("/v1/chat/:id/redact", func(c *gin.Context) {
		c.Set("user_id", "mod-789")
		c.Set("tenant_id", "tenant-1")
		c.Set("is_admin", false)
		c.Set("is_moderator", true)
		handler.RedactMessage(c)
	})

	t.Run("Success - moderator redacts message", func(t *testing.T) {
		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Inappropriate", nil)

		body := map[string]interface{}{
			"reason": "Violates policy",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/chat/%s/redact", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "redacted", response["status"])
		assert.Equal(t, "[REDACTED]", response["new_body"])
	})

	t.Run("Forbidden - regular user cannot redact", func(t *testing.T) {
		r2 := gin.New()
		r2.POST("/v1/chat/:id/redact", func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Set("tenant_id", "tenant-1")
			c.Set("is_admin", false)
			c.Set("is_moderator", false)
			handler.RedactMessage(c)
		})

		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Test", nil)

		body := map[string]interface{}{
			"reason": "No reason",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/chat/%s/redact", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Validation - missing reason", func(t *testing.T) {
		ctx := context.TODO()
		msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Test", nil)

		body := map[string]interface{}{} // Missing reason
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", fmt.Sprintf("/v1/chat/%s/redact", msg.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChatHandler_GetEditHistory(t *testing.T) {
	r, handler, chatService, _ := setupChatTest()

	// Create and edit message
	ctx := context.TODO()
	msg, _ := chatService.SendMessage(ctx, "user-123", "tenant-1", "host", "Original", nil)
	chatService.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "Edited v1", false)
	chatService.EditMessage(ctx, msg.ID, "user-123", "tenant-1", "Edited v2", false)

	r.GET("/v1/chat/:id/edits", func(c *gin.Context) {
		c.Set("tenant_id", "tenant-1")
		handler.GetEditHistory(c)
	})

	t.Run("Success - get edit history", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/v1/chat/%s/edits", msg.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, msg.ID, response["message_id"])
		edits := response["edits"].([]interface{})
		assert.Len(t, edits, 2)
	})
}
