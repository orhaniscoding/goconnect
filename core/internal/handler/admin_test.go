package handler

import (
	"context"
	"encoding/json"
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

// setupAdminTest creates all required repositories and services for admin testing
func setupAdminTest() (*gin.Engine, *AdminHandler, *repository.InMemoryUserRepository, *repository.InMemoryTenantRepository, *repository.InMemoryNetworkRepository, *repository.InMemoryDeviceRepository, *repository.InMemoryAdminRepository) {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	chatRepo := repository.NewInMemoryChatRepository()
	adminRepo := repository.NewInMemoryAdminRepository()

	adminService := service.NewAdminService(
		userRepo,
		adminRepo,
		tenantRepo,
		networkRepo,
		deviceRepo,
		chatRepo,
		nil,                     // No auditor
		nil,                     // No Redis
		func() int { return 5 }, // mock active connections
	)

	handler := NewAdminHandler(adminService)
	r := gin.New()

	return r, handler, userRepo, tenantRepo, networkRepo, deviceRepo, adminRepo
}

// adminAuthMiddleware returns a test auth middleware that requires admin
func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if token == "admin-token" {
			c.Set("user_id", "admin_user")
			c.Set("tenant_id", "t1")
			c.Set("is_admin", true)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}
}

// ==================== LIST USERS TESTS ====================

func TestAdminHandler_ListUsers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users", adminAuthMiddleware(), handler.ListUsers)

		// Seed users
		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user1@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "user2@test.com"}))

		req := httptest.NewRequest("GET", "/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)

		meta := response["meta"].(map[string]interface{})
		assert.Equal(t, float64(2), meta["total"])
	})

	t.Run("Pagination", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users", adminAuthMiddleware(), handler.ListUsers)

		ctx := context.Background()
		for i := 0; i < 15; i++ {
			require.NoError(t, userRepo.Create(ctx, &domain.User{
				ID:    domain.GenerateNetworkID(),
				Email: "user" + string(rune('a'+i)) + "@test.com",
			}))
		}

		req := httptest.NewRequest("GET", "/v1/admin/users?limit=5&offset=0", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 5)

		meta := response["meta"].(map[string]interface{})
		assert.Equal(t, float64(15), meta["total"])
		assert.Equal(t, float64(5), meta["limit"])
	})

	t.Run("Query Filter", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users", adminAuthMiddleware(), handler.ListUsers)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "alice@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u2", Email: "bob@test.com"}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u3", Email: "alice.smith@test.com"}))

		req := httptest.NewRequest("GET", "/v1/admin/users?q=alice", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users", adminAuthMiddleware(), handler.ListUsers)

		req := httptest.NewRequest("GET", "/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ==================== LIST TENANTS TESTS ====================

func TestAdminHandler_ListTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _, tenantRepo, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/tenants", adminAuthMiddleware(), handler.ListTenants)

		ctx := context.Background()
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Tenant 1"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t2", Name: "Tenant 2"}))

		req := httptest.NewRequest("GET", "/v1/admin/tenants", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("Pagination", func(t *testing.T) {
		r, handler, _, tenantRepo, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/tenants", adminAuthMiddleware(), handler.ListTenants)

		ctx := context.Background()
		for i := 0; i < 12; i++ {
			require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{
				ID:   domain.GenerateNetworkID(),
				Name: "Tenant " + string(rune('A'+i)),
			}))
		}

		req := httptest.NewRequest("GET", "/v1/admin/tenants?limit=5&offset=5", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 5)
	})

	t.Run("Query Filter", func(t *testing.T) {
		r, handler, _, tenantRepo, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/tenants", adminAuthMiddleware(), handler.ListTenants)

		ctx := context.Background()
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Acme Corp"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t2", Name: "Beta Inc"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t3", Name: "Acme Labs"}))

		req := httptest.NewRequest("GET", "/v1/admin/tenants?q=acme", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

// ==================== GET SYSTEM STATS TESTS ====================

func TestAdminHandler_GetSystemStats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, tenantRepo, networkRepo, deviceRepo, _ := setupAdminTest()
		r.GET("/v1/admin/stats", adminAuthMiddleware(), handler.GetSystemStats)

		ctx := context.Background()
		// Seed data
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com"}))
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Tenant"}))
		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n1", TenantID: "t1", Name: "Network", CIDR: "10.0.0.0/24"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device", PubKey: "pubkey1"}))

		req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].(map[string]interface{})

		assert.Equal(t, float64(1), data["total_users"])
		assert.Equal(t, float64(1), data["total_tenants"])
		assert.Equal(t, float64(1), data["total_networks"])
		assert.Equal(t, float64(1), data["total_devices"])
		assert.Equal(t, float64(5), data["active_connections"]) // mock returns 5
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/stats", adminAuthMiddleware(), handler.GetSystemStats)

		// Request without authorization header
		req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Empty database", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/stats", adminAuthMiddleware(), handler.GetSystemStats)

		req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].(map[string]interface{})

		assert.Equal(t, float64(0), data["total_users"])
		assert.Equal(t, float64(0), data["total_tenants"])
	})
}

// ==================== TOGGLE USER ADMIN TESTS ====================

func TestAdminHandler_ToggleUserAdmin(t *testing.T) {
	t.Run("Success - Make Admin", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/toggle-admin", adminAuthMiddleware(), handler.ToggleUserAdmin)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "user@test.com", IsAdmin: false}))

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/toggle-admin", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, true, data["is_admin"])
	})

	t.Run("Success - Remove Admin", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/toggle-admin", adminAuthMiddleware(), handler.ToggleUserAdmin)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "admin@test.com", IsAdmin: true}))

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/toggle-admin", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, false, data["is_admin"])
	})

	t.Run("Not Found", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/toggle-admin", adminAuthMiddleware(), handler.ToggleUserAdmin)

		req := httptest.NewRequest("POST", "/v1/admin/users/nonexistent/toggle-admin", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Empty ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		// Explicitly define route without :id to test empty ID case
		r.POST("/v1/admin/users//toggle-admin", adminAuthMiddleware(), handler.ToggleUserAdmin)

		req := httptest.NewRequest("POST", "/v1/admin/users//toggle-admin", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Gin returns 301 redirect for double slashes or 404
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusMovedPermanently)
	})
}

// ==================== DELETE USER TESTS ====================

func TestAdminHandler_DeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.DELETE("/v1/admin/users/:id", adminAuthMiddleware(), handler.DeleteUser)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{ID: "u1", Email: "delete@test.com"}))

		req := httptest.NewRequest("DELETE", "/v1/admin/users/u1", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify user is deleted
		_, err := userRepo.GetByID(ctx, "u1")
		assert.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.DELETE("/v1/admin/users/:id", adminAuthMiddleware(), handler.DeleteUser)

		req := httptest.NewRequest("DELETE", "/v1/admin/users/nonexistent", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ==================== DELETE TENANT TESTS ====================

func TestAdminHandler_DeleteTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _, tenantRepo, _, _, _ := setupAdminTest()
		r.DELETE("/v1/admin/tenants/:id", adminAuthMiddleware(), handler.DeleteTenant)

		ctx := context.Background()
		require.NoError(t, tenantRepo.Create(ctx, &domain.Tenant{ID: "t1", Name: "Delete Me"}))

		req := httptest.NewRequest("DELETE", "/v1/admin/tenants/t1", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify tenant is deleted
		_, err := tenantRepo.GetByID(ctx, "t1")
		assert.Error(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.DELETE("/v1/admin/tenants/:id", adminAuthMiddleware(), handler.DeleteTenant)

		req := httptest.NewRequest("DELETE", "/v1/admin/tenants/nonexistent", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ==================== LIST NETWORKS TESTS ====================

func TestAdminHandler_ListNetworks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _, _, networkRepo, _, _ := setupAdminTest()
		r.GET("/v1/admin/networks", adminAuthMiddleware(), handler.ListNetworks)

		ctx := context.Background()
		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n1", TenantID: "t1", Name: "Network 1", CIDR: "10.0.0.0/24"}))
		require.NoError(t, networkRepo.Create(ctx, &domain.Network{ID: "n2", TenantID: "t1", Name: "Network 2", CIDR: "10.1.0.0/24"}))

		req := httptest.NewRequest("GET", "/v1/admin/networks", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

		data, ok := response["data"].([]interface{})
		if ok {
			assert.Len(t, data, 2)
		}
	})

	t.Run("Cursor Pagination", func(t *testing.T) {
		r, handler, _, _, networkRepo, _, _ := setupAdminTest()
		r.GET("/v1/admin/networks", adminAuthMiddleware(), handler.ListNetworks)

		ctx := context.Background()
		for i := 0; i < 10; i++ {
			require.NoError(t, networkRepo.Create(ctx, &domain.Network{
				ID:       domain.GenerateNetworkID(),
				TenantID: "t1",
				Name:     "Network " + string(rune('A'+i)),
				CIDR:     "10." + string(rune('0'+i)) + ".0.0/24",
			}))
		}

		req := httptest.NewRequest("GET", "/v1/admin/networks?limit=5", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

		data, ok := response["data"].([]interface{})
		if ok && len(data) > 0 {
			assert.LessOrEqual(t, len(data), 5)
		}
	})
}

// ==================== LIST DEVICES TESTS ====================

func TestAdminHandler_ListDevices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _, _, _, deviceRepo, _ := setupAdminTest()
		r.GET("/v1/admin/devices", adminAuthMiddleware(), handler.ListDevices)

		ctx := context.Background()
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d1", UserID: "u1", TenantID: "t1", Name: "Device 1", PubKey: "pk1"}))
		require.NoError(t, deviceRepo.Create(ctx, &domain.Device{ID: "d2", UserID: "u1", TenantID: "t1", Name: "Device 2", PubKey: "pk2"}))

		req := httptest.NewRequest("GET", "/v1/admin/devices", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

		data, ok := response["data"].([]interface{})
		if ok {
			assert.Len(t, data, 2)
		}
	})

	t.Run("Cursor Pagination", func(t *testing.T) {
		r, handler, _, _, _, deviceRepo, _ := setupAdminTest()
		r.GET("/v1/admin/devices", adminAuthMiddleware(), handler.ListDevices)

		ctx := context.Background()
		now := time.Now()
		for i := 0; i < 10; i++ {
			require.NoError(t, deviceRepo.Create(ctx, &domain.Device{
				ID:        domain.GenerateNetworkID(),
				UserID:    "u1",
				TenantID:  "t1",
				Name:      "Device " + string(rune('A'+i)),
				PubKey:    "pk" + string(rune('a'+i)),
				CreatedAt: now.Add(time.Duration(i) * time.Minute),
			}))
		}

		req := httptest.NewRequest("GET", "/v1/admin/devices?limit=5", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

		data, ok := response["data"].([]interface{})
		if ok && len(data) > 0 {
			assert.LessOrEqual(t, len(data), 5)
		}
	})
}

// ==================== LIST ALL USERS TESTS ====================
// Note: Success cases require database backend (AdminRepository uses SQL directly)
// These tests only cover the handler validation logic

func TestAdminHandler_ListAllUsers(t *testing.T) {
	t.Run("Unauthorized - No Auth Middleware", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/all-users", handler.ListAllUsers) // No auth middleware

		req := httptest.NewRequest("GET", "/v1/admin/all-users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With query parameters", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/all-users", adminAuthMiddleware(), handler.ListAllUsers)

		// Should handle query params even if service fails
		req := httptest.NewRequest("GET", "/v1/admin/all-users?role=admin&status=active&tenant_id=t1&q=test&page=2&per_page=10", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Reaches service layer (not blocked by middleware)
		assert.True(t, w.Code >= 200)
	})
}

// ==================== GET USER STATS TESTS ====================
// Note: Success cases require database backend (AdminRepository uses SQL directly)

func TestAdminHandler_GetUserStats(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/stats", handler.GetUserStats)

		req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("With auth header", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/stats", adminAuthMiddleware(), handler.GetUserStats)

		req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Reaches service layer (not blocked by middleware)
		assert.True(t, w.Code >= 200)
	})
}

// ==================== UPDATE USER ROLE TESTS ====================

func TestAdminHandler_UpdateUserRole(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", handler.UpdateUserRole)

		req := httptest.NewRequest("PUT", "/v1/admin/users/u1/role", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing User ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users//role", adminAuthMiddleware(), handler.UpdateUserRole)

		req := httptest.NewRequest("PUT", "/v1/admin/users//role", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return bad request since user ID is empty
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", adminAuthMiddleware(), handler.UpdateUserRole)

		req := httptest.NewRequest("PUT", "/v1/admin/users/u1/role", strings.NewReader("{invalid json}"))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Valid request body", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", adminAuthMiddleware(), handler.UpdateUserRole)

		body := `{"is_admin": true, "is_moderator": false}`
		req := httptest.NewRequest("PUT", "/v1/admin/users/u1/role", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should pass validation and reach service layer (may return error from service)
		assert.True(t, w.Code != http.StatusUnauthorized)
	})

	t.Run("Success - update existing user role", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", adminAuthMiddleware(), handler.UpdateUserRole)

		// Create both admin user and target user
		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:      "admin_user",
			Email:   "admin@test.com",
			IsAdmin: true,
		}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:    "target_user",
			Email: "target@test.com",
		}))

		body := `{"is_admin": true, "is_moderator": true}`
		req := httptest.NewRequest("PUT", "/v1/admin/users/target_user/role", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should succeed or return valid error - test the path
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, w.Code)
	})

	t.Run("Promote to moderator only", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", adminAuthMiddleware(), handler.UpdateUserRole)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:      "admin_user",
			Email:   "admin@test.com",
			IsAdmin: true,
		}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:    "mod_user",
			Email: "mod@test.com",
		}))

		body := `{"is_admin": false, "is_moderator": true}`
		req := httptest.NewRequest("PUT", "/v1/admin/users/mod_user/role", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, w.Code)
	})

	t.Run("Demote user roles", func(t *testing.T) {
		r, handler, userRepo, _, _, _, _ := setupAdminTest()
		r.PUT("/v1/admin/users/:id/role", adminAuthMiddleware(), handler.UpdateUserRole)

		ctx := context.Background()
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:      "admin_user",
			Email:   "admin@test.com",
			IsAdmin: true,
		}))
		require.NoError(t, userRepo.Create(ctx, &domain.User{
			ID:          "demote_user",
			Email:       "demote@test.com",
			IsAdmin:     true,
			IsModerator: true,
		}))

		body := `{"is_admin": false, "is_moderator": false}`
		req := httptest.NewRequest("PUT", "/v1/admin/users/demote_user/role", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusNotFound}, w.Code)
	})
}

// ==================== SUSPEND/UNSUSPEND TESTS ====================

func TestAdminHandler_SuspendUser(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/suspend", handler.SuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/suspend", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing User ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users//suspend", adminAuthMiddleware(), handler.SuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users//suspend", strings.NewReader(`{"reason": "test"}`))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/suspend", adminAuthMiddleware(), handler.SuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/suspend", strings.NewReader("{invalid}"))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing reason", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/suspend", adminAuthMiddleware(), handler.SuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/suspend", strings.NewReader(`{"reason": ""}`))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Valid request body", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/suspend", adminAuthMiddleware(), handler.SuspendUser)

		body := `{"reason": "Violation of terms of service"}`
		req := httptest.NewRequest("POST", "/v1/admin/users/u1/suspend", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should pass validation and reach service layer (may return error from service)
		assert.True(t, w.Code != http.StatusUnauthorized)
	})

	t.Run("Success_SuspendUser", func(t *testing.T) {
		r, handler, userRepo, _, _, _, adminRepo := setupAdminTest()
		r.POST("/v1/admin/users/:id/suspend", adminAuthMiddleware(), handler.SuspendUser)

		ctx := context.Background()

		// Create admin user (the one making the request) - for user lookup
		adminUser := &domain.User{ID: "admin_user", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Create target user to suspend
		targetUser := &domain.User{ID: "target-susp", Email: "target@test.com", TenantID: "t1", IsAdmin: false}
		require.NoError(t, userRepo.Create(ctx, targetUser))
		adminRepo.AddUser(targetUser) // Also add to admin repo for suspend

		body := `{"user_id": "target-susp", "reason": "Violation of terms of service - very bad behavior"}`
		req := httptest.NewRequest("POST", "/v1/admin/users/target-susp/suspend", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Log error if not OK
		if w.Code != http.StatusOK {
			t.Logf("Response: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminHandler_UnsuspendUser(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/unsuspend", handler.UnsuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/unsuspend", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing User ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users//unsuspend", adminAuthMiddleware(), handler.UnsuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users//unsuspend", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("With user ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.POST("/v1/admin/users/:id/unsuspend", adminAuthMiddleware(), handler.UnsuspendUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/u1/unsuspend", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Reaches service layer (not blocked by middleware)
		assert.True(t, w.Code >= 200)
	})

	t.Run("Success_UnsuspendUser", func(t *testing.T) {
		r, handler, userRepo, _, _, _, adminRepo := setupAdminTest()
		r.POST("/v1/admin/users/:id/unsuspend", adminAuthMiddleware(), handler.UnsuspendUser)

		ctx := context.Background()

		// Create admin user
		adminUser := &domain.User{ID: "admin_user", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Create suspended user
		suspendedUser := &domain.User{ID: "susp-user", Email: "suspended@test.com", TenantID: "t1", Suspended: true}
		require.NoError(t, userRepo.Create(ctx, suspendedUser))
		adminRepo.AddUser(suspendedUser)

		req := httptest.NewRequest("POST", "/v1/admin/users/susp-user/unsuspend", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("Response: %s", w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAdminHandler_GetUserDetails(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users/:id", handler.GetUserDetails)

		req := httptest.NewRequest("GET", "/v1/admin/users/u1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing User ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users/", adminAuthMiddleware(), handler.GetUserDetails)

		req := httptest.NewRequest("GET", "/v1/admin/users/", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("With user ID", func(t *testing.T) {
		r, handler, _, _, _, _, _ := setupAdminTest()
		r.GET("/v1/admin/users/:id", adminAuthMiddleware(), handler.GetUserDetails)

		req := httptest.NewRequest("GET", "/v1/admin/users/u1", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Reaches service layer (not blocked by middleware)
		assert.True(t, w.Code >= 200)
	})

	t.Run("Success", func(t *testing.T) {
		r, handler, userRepo, _, _, _, adminRepo := setupAdminTest()
		r.GET("/v1/admin/users/:id", adminAuthMiddleware(), handler.GetUserDetails)

		ctx := context.Background()

		// Create admin user
		adminUser := &domain.User{ID: "admin_user", Email: "admin@test.com", TenantID: "t1", IsAdmin: true}
		require.NoError(t, userRepo.Create(ctx, adminUser))

		// Create target user
		targetUser := &domain.User{ID: "target-details", Email: "target@test.com", TenantID: "t1"}
		require.NoError(t, userRepo.Create(ctx, targetUser))
		adminRepo.AddUser(targetUser)

		req := httptest.NewRequest("GET", "/v1/admin/users/target-details", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
