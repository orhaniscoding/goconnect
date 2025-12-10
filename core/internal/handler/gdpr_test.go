package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGDPRTest creates all required repositories and services for GDPR testing
func setupGDPRTest() (*gin.Engine, *GDPRHandler, *repository.InMemoryUserRepository, *repository.InMemoryDeviceRepository) {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	deletionRepo := repository.NewInMemoryDeletionRequestRepository()

	gdprService := service.NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo, deletionRepo)
	handler := NewGDPRHandler(gdprService, nil) // nil auditor for tests

	r := gin.New()
	return r, handler, userRepo, deviceRepo
}

// gdprAuthMiddleware returns a test auth middleware for GDPR tests
func gdprAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		switch token {
		case "valid-user":
			c.Set("user_id", "user1")
			c.Set("tenant_id", "t1")
			c.Set("is_admin", false)
			c.Next()
		case "admin-user":
			c.Set("user_id", "admin1")
			c.Set("tenant_id", "t1")
			c.Set("is_admin", true)
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		}
	}
}

// ==================== EXPORT DATA TESTS ====================

func TestGDPRHandler_ExportData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, deviceRepo := setupGDPRTest()
		r.GET("/v1/me/export", gdprAuthMiddleware(), handler.ExportData)

		ctx := context.Background()
		// Seed user
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:       "user1",
			TenantID: "t1",
			Email:    "user@test.com",
		}))
		// Seed device
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
			ID:        "d1",
			UserID:    "user1",
			TenantID:  "t1",
			Name:      "Test Device",
			PubKey:    "pubkey1",
			CreatedAt: time.Now(),
		}))

		req := httptest.NewRequest("GET", "/v1/me/export", nil)
		req.Header.Set("Authorization", "Bearer valid-user")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.NotNil(t, response["user"])
		assert.NotNil(t, response["exported_at"])
	})

	t.Run("Unauthorized - No Token", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.GET("/v1/me/export", gdprAuthMiddleware(), handler.ExportData)

		req := httptest.NewRequest("GET", "/v1/me/export", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Unauthorized - Invalid Token", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.GET("/v1/me/export", gdprAuthMiddleware(), handler.ExportData)

		req := httptest.NewRequest("GET", "/v1/me/export", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("User Not Found", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.GET("/v1/me/export", gdprAuthMiddleware(), handler.ExportData)

		// Don't seed user - will cause not found
		req := httptest.NewRequest("GET", "/v1/me/export", nil)
		req.Header.Set("Authorization", "Bearer valid-user")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ==================== EXPORT DATA DOWNLOAD TESTS ====================

func TestGDPRHandler_ExportDataDownload(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, _ := setupGDPRTest()
		r.GET("/v1/me/export/download", gdprAuthMiddleware(), handler.ExportDataDownload)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:       "user1",
			TenantID: "t1",
			Email:    "user@test.com",
		}))

		req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
		req.Header.Set("Authorization", "Bearer valid-user")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.GET("/v1/me/export/download", gdprAuthMiddleware(), handler.ExportDataDownload)

		req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ==================== REQUEST DELETION TESTS ====================

func TestGDPRHandler_RequestDeletion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, _ := setupGDPRTest()
		r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:       "user1",
			TenantID: "t1",
			Email:    "user@test.com",
		}))

		body := `{"confirmation": "DELETE MY ACCOUNT"}`
		req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer valid-user")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "pending", response["status"])
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

		body := `{"confirmation": "DELETE MY ACCOUNT"}`
		req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Confirmation", func(t *testing.T) {
		r, handler, userRepo, _ := setupGDPRTest()
		r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:       "user1",
			TenantID: "t1",
			Email:    "user@test.com",
		}))

		body := `{"confirmation": "delete my account"}` // lowercase - should fail
		req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer valid-user")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Confirmation", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

		body := `{}`
		req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer valid-user")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		r, handler, _, _ := setupGDPRTest()
		r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

		body := `{invalid json}`
		req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer valid-user")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ==================== ExportDataDownload Comprehensive Tests ====================

func TestGDPRHandler_ExportDataDownload_UserNotFound(t *testing.T) {
	r, handler, _, _ := setupGDPRTest()
	r.GET("/v1/me/export/download", gdprAuthMiddleware(), handler.ExportDataDownload)

	// Don't seed user - will cause not found
	req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
	req.Header.Set("Authorization", "Bearer valid-user")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGDPRHandler_ExportDataDownload_WithDevices(t *testing.T) {
	r, handler, userRepo, deviceRepo := setupGDPRTest()
	r.GET("/v1/me/export/download", gdprAuthMiddleware(), handler.ExportDataDownload)

	ctx := context.Background()
	require.NoError(t, userRepo.Create(ctx, &domain.User{
		ID:       "user1",
		TenantID: "t1",
		Email:    "user@test.com",
	}))
	// Add multiple devices
	require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
		ID:        "d1",
		UserID:    "user1",
		TenantID:  "t1",
		Name:      "Device 1",
		PubKey:    "pubkey1",
		Platform:  "linux",
		CreatedAt: time.Now(),
	}))
	require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
		ID:        "d2",
		UserID:    "user1",
		TenantID:  "t1",
		Name:      "Device 2",
		PubKey:    "pubkey2",
		Platform:  "macos",
		CreatedAt: time.Now(),
	}))

	req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
	req.Header.Set("Authorization", "Bearer valid-user")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify content contains user data
	body := w.Body.String()
	assert.Contains(t, body, "user1")
}

func TestGDPRHandler_ExportDataDownload_InvalidToken(t *testing.T) {
	r, handler, _, _ := setupGDPRTest()
	r.GET("/v1/me/export/download", gdprAuthMiddleware(), handler.ExportDataDownload)

	req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGDPRHandler_ExportDataDownload_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	deletionRepo := repository.NewInMemoryDeletionRequestRepository()

	gdprService := service.NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo, deletionRepo)
	handler := NewGDPRHandler(gdprService, nil)

	// Middleware that doesn't set user_id
	r.GET("/v1/me/export/download", func(c *gin.Context) {
		// Don't set user_id
		c.Set("tenant_id", "t1")
		c.Next()
	}, handler.ExportDataDownload)

	req := httptest.NewRequest("GET", "/v1/me/export/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== ExportData Additional Tests ====================

func TestGDPRHandler_ExportData_WithMultipleDevices(t *testing.T) {
	r, handler, userRepo, deviceRepo := setupGDPRTest()
	r.GET("/v1/me/export", gdprAuthMiddleware(), handler.ExportData)

	ctx := context.Background()
	require.NoError(t, userRepo.Create(ctx, &domain.User{
		ID:       "user1",
		TenantID: "t1",
		Email:    "user@test.com",
	}))
	// Add devices
	for i := 1; i <= 3; i++ {
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
			ID:        fmt.Sprintf("d%d", i),
			UserID:    "user1",
			TenantID:  "t1",
			Name:      fmt.Sprintf("Device %d", i),
			PubKey:    fmt.Sprintf("pubkey%d", i),
			CreatedAt: time.Now(),
		}))
	}

	req := httptest.NewRequest("GET", "/v1/me/export", nil)
	req.Header.Set("Authorization", "Bearer valid-user")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.NotNil(t, response["user"])
	assert.NotNil(t, response["devices"])
}

func TestGDPRHandler_ExportData_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := repository.NewInMemoryUserRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	deletionRepo := repository.NewInMemoryDeletionRequestRepository()

	gdprService := service.NewGDPRService(userRepo, deviceRepo, networkRepo, membershipRepo, deletionRepo)
	handler := NewGDPRHandler(gdprService, nil)

	// Middleware that doesn't set user_id
	r.GET("/v1/me/export", func(c *gin.Context) {
		// Don't set user_id
		c.Next()
	}, handler.ExportData)

	req := httptest.NewRequest("GET", "/v1/me/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== RequestDeletion Additional Tests ====================

func TestGDPRHandler_RequestDeletion_UserNotFound(t *testing.T) {
	r, handler, _, _ := setupGDPRTest()
	r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

	// User not created
	body := `{"confirmation": "DELETE MY ACCOUNT"}`
	req := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-user")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return accepted even if user not found (deletion request is queued)
	assert.True(t, w.Code == http.StatusAccepted || w.Code == http.StatusNotFound)
}

func TestGDPRHandler_RequestDeletion_DuplicateRequest(t *testing.T) {
	r, handler, userRepo, _ := setupGDPRTest()
	r.POST("/v1/me/delete", gdprAuthMiddleware(), handler.RequestDeletion)

	ctx := context.Background()
	require.NoError(t, userRepo.Create(ctx, &domain.User{
		ID:       "user1",
		TenantID: "t1",
		Email:    "user@test.com",
	}))

	body := `{"confirmation": "DELETE MY ACCOUNT"}`

	// First request
	req1 := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
	req1.Header.Set("Authorization", "Bearer valid-user")
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusAccepted, w1.Code)

	// Second request (duplicate)
	req2 := httptest.NewRequest("POST", "/v1/me/delete", strings.NewReader(body))
	req2.Header.Set("Authorization", "Bearer valid-user")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Should either accept or return conflict
	assert.True(t, w2.Code == http.StatusAccepted || w2.Code == http.StatusConflict)
}
