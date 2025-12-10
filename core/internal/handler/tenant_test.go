package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

// setupTenantTest creates all required repositories and services for tenant testing
func setupTenantTest() (*gin.Engine, *TenantHandler, *service.TenantMembershipService, *mockAuthService) {
	gin.SetMode(gin.TestMode)

	// Create repositories
	tenantMemberRepo := repository.NewInMemoryTenantMemberRepository()
	tenantInviteRepo := repository.NewInMemoryTenantInviteRepository()
	tenantAnnouncementRepo := repository.NewInMemoryTenantAnnouncementRepository()
	tenantChatRepo := repository.NewInMemoryTenantChatRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	userRepo := repository.NewInMemoryUserRepository()

	// Create service
	tenantService := service.NewTenantMembershipService(
		tenantMemberRepo,
		tenantInviteRepo,
		tenantAnnouncementRepo,
		tenantChatRepo,
		tenantRepo,
		userRepo,
	)

	// Create handler
	handler := NewTenantHandler(tenantService)

	// Create router with auth middleware mock
	r := gin.New()
	mockAuth := newMockAuthServiceWithTokens()

	return r, handler, tenantService, mockAuth
}

// authMiddleware returns a test auth middleware that sets user context
func authMiddleware(mockAuth *mockAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := mockAuth.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

// ==================== CREATE TENANT TESTS ====================

func TestTenantHandler_CreateTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.POST("/v1/tenants", authMiddleware(mockAuth), handler.CreateTenant)

		body := map[string]interface{}{
			"name":        "Test Tenant",
			"description": "A test tenant",
			"visibility":  "public",
			"access_type": "open",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Tenant", data["name"])
		assert.NotEmpty(t, data["id"])
	})

	t.Run("Missing name", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.POST("/v1/tenants", authMiddleware(mockAuth), handler.CreateTenant)

		body := map[string]interface{}{
			"description": "A test tenant without name",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.POST("/v1/tenants", authMiddleware(mockAuth), handler.CreateTenant)

		body := map[string]interface{}{
			"name": "Test Tenant",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Duplicate name", func(t *testing.T) {
		t.Skip("Duplicate name validation not implemented yet")
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants", authMiddleware(mockAuth), handler.CreateTenant)

		ctx := context.Background()
		// First successful creation
		_, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name: "Unique Tenant Name",
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"name":        "Unique Tenant Name", // Duplicate name
			"description": "Another test tenant",
			"visibility":  "public",
			"access_type": "open",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code) // Expect 409 Conflict
	})
}

// ==================== GET TENANT TESTS ====================

func TestTenantHandler_GetTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.GetTenant)

		// Create a tenant first
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Tenant", data["name"])
	})

	t.Run("Not found", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.GetTenant)

		req := httptest.NewRequest("GET", "/v1/tenants/non-existent", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Service returns error for non-existent tenant
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
	})
}

// ==================== UPDATE TENANT TESTS ====================

func TestTenantHandler_UpdateTenant(t *testing.T) {
	t.Run("Success - Owner updates", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.UpdateTenant)

		// Create tenant as user_dev (token user)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Original Name",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"name":        "Updated Name",
			"description": "New description",
			"visibility":  "public",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Updated Name", data["name"])
		assert.Equal(t, "New description", data["description"])
		assert.Equal(t, "public", data["visibility"])
	})

	t.Run("Forbidden - Non-member", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.UpdateTenant)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Other Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"name": "Trying to update",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - Regular member", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.UpdateTenant)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Other Tenant",
			Visibility: domain.TenantVisibilityPublic,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		// user_dev joins as member
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		body := map[string]interface{}{
			"name": "Trying to update as member",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Password required for password access", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.UpdateTenant)

		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		// Try to change to password access without providing password
		body := map[string]interface{}{
			"access_type": "password",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTenantHandler_DeleteTenant(t *testing.T) {
	t.Run("Success - Owner deletes tenant", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.DeleteTenant)

		// Create tenant as user_dev (token user)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Tenant to Delete",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Tenant deleted successfully", response["message"])

		// Verify tenant is deleted
		_, err = tenantService.GetTenant(ctx, tenant.ID)
		assert.Error(t, err)
	})

	t.Run("Forbidden - Non-member cannot delete", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.DeleteTenant)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "other_user", &domain.CreateTenantRequest{
			Name:       "Other Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - Admin cannot delete (only owner)", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.DeleteTenant)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Owner Tenant",
			Visibility: domain.TenantVisibilityPublic,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		// user_dev joins and is promoted to admin
		membership, err := tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		err = tenantService.UpdateMemberRole(ctx, "owner_user", tenant.ID, membership.ID, domain.TenantRoleAdmin)
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Not Found - Tenant does not exist", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.DeleteTenant)

		req := httptest.NewRequest("DELETE", "/v1/tenants/nonexistent-tenant-id", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code) // Not found shows as forbidden since user is not a member
	})

	t.Run("Failure - Delete with active members", func(t *testing.T) {
		t.Skip("Delete with active members validation not implemented yet")
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId", authMiddleware(mockAuth), handler.DeleteTenant)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name: "Tenant with Members",
		})
		require.NoError(t, err)

		// Add another member
		_, err = tenantService.JoinTenant(ctx, "active_member", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code) // Expect 409 Conflict if tenant has members
	})
}

// ==================== PUBLIC TENANT DISCOVERY TESTS ====================

func TestTenantHandler_ListPublicTenants(t *testing.T) {
	r, handler, tenantService, _ := setupTenantTest()
	r.GET("/v1/tenants/public", handler.ListPublicTenants)

	ctx := context.Background()
	_, err := tenantService.CreateTenant(ctx, "owner_alpha", &domain.CreateTenantRequest{
		Name:       "Alpha Ops",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	require.NoError(t, err)
	_, err = tenantService.CreateTenant(ctx, "owner_beta", &domain.CreateTenantRequest{
		Name:       "Beta Private",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
	})
	require.NoError(t, err)
	_, err = tenantService.CreateTenant(ctx, "owner_gamma", &domain.CreateTenantRequest{
		Name:       "Gamma Ops",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/v1/tenants/public?limit=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestTenantHandler_RegisterRoutes_PublicVsProtected(t *testing.T) {
	r, handler, tenantService, _ := setupTenantTest()

	authMiddleware := func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("user_id", "user_dev")
		c.Next()
	}

	handler.RegisterRoutes(r.Group("/v1"), authMiddleware)

	ctx := context.Background()
	tenant, err := tenantService.CreateTenant(ctx, "owner", &domain.CreateTenantRequest{
		Name:       "Public Ops",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	require.NoError(t, err)

	// Public listing should succeed without auth header
	publicReq := httptest.NewRequest("GET", "/v1/tenants/public", nil)
	publicResp := httptest.NewRecorder()
	r.ServeHTTP(publicResp, publicReq)
	assert.Equal(t, http.StatusOK, publicResp.Code)

	// Protected tenant detail should require auth
	protectedReq := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID, nil)
	protectedResp := httptest.NewRecorder()
	r.ServeHTTP(protectedResp, protectedReq)
	assert.Equal(t, http.StatusUnauthorized, protectedResp.Code)

	// Same endpoint with auth header should pass
	protectedReqAuth := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID, nil)
	protectedReqAuth.Header.Set("Authorization", "Bearer test")
	protectedRespAuth := httptest.NewRecorder()
	r.ServeHTTP(protectedRespAuth, protectedReqAuth)
	assert.Equal(t, http.StatusOK, protectedRespAuth.Code)

	// Users/me/tenants also needs auth
	userTenantsReq := httptest.NewRequest("GET", "/v1/users/me/tenants", nil)
	userTenantsResp := httptest.NewRecorder()
	r.ServeHTTP(userTenantsResp, userTenantsReq)
	assert.Equal(t, http.StatusUnauthorized, userTenantsResp.Code)

	userTenantsReqAuth := httptest.NewRequest("GET", "/v1/users/me/tenants", nil)
	userTenantsReqAuth.Header.Set("Authorization", "Bearer test")
	userTenantsRespAuth := httptest.NewRecorder()
	r.ServeHTTP(userTenantsRespAuth, userTenantsReqAuth)
	assert.Equal(t, http.StatusOK, userTenantsRespAuth.Code)
}

func TestTenantHandler_SearchTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, _ := setupTenantTest()
		r.GET("/v1/tenants/search", handler.SearchTenants)

		ctx := context.Background()
		_, err := tenantService.CreateTenant(ctx, "owner_alpha", &domain.CreateTenantRequest{
			Name:       "Alpha Ops",
			Visibility: domain.TenantVisibilityPublic,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)
		_, err = tenantService.CreateTenant(ctx, "owner_beta", &domain.CreateTenantRequest{
			Name:       "Beta Ops",
			Visibility: domain.TenantVisibilityPublic,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/v1/tenants/search?q=alpha&limit=5", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("Missing query", func(t *testing.T) {
		r, handler, _, _ := setupTenantTest()
		r.GET("/v1/tenants/search", handler.SearchTenants)

		req := httptest.NewRequest("GET", "/v1/tenants/search", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ==================== JOIN TENANT TESTS ====================

func TestTenantHandler_JoinTenant(t *testing.T) {
	t.Run("Success - Open tenant", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		// Create a public open tenant
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Open Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["message"], "Successfully joined")
	})

	t.Run("Already a member - idempotent", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		// Create tenant (owner is auto-member)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Try to join as owner (already member)
		// Service allows idempotent join - returns existing membership
		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should succeed (idempotent) or conflict based on implementation
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusConflict)
	})

	t.Run("Password protected requires password", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Password Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessPassword,
			Password:   "super-secret",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Password protected success with valid password", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Password Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessPassword,
			Password:   "super-secret",
		})
		require.NoError(t, err)

		body := map[string]string{"password": "super-secret"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invite only requires code", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Invite Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessInviteOnly,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Capacity limit enforced", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Limited Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
			MaxMembers: 1,
		})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ==================== JOIN BY CODE TESTS ====================

func TestTenantHandler_JoinByCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/join-by-code", authMiddleware(mockAuth), handler.JoinByCode)

		// Create tenant and invite
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Invite Only Tenant",
			Visibility: "private",
			AccessType: "invite_only",
		})
		require.NoError(t, err)

		// Create invite
		invite, err := tenantService.CreateInvite(ctx, "owner_user", tenant.ID, &domain.CreateTenantInviteRequest{
			MaxUses:   10,
			ExpiresIn: 3600,
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"code": invite.Code,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/join-by-code", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid code", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/join-by-code", authMiddleware(mockAuth), handler.JoinByCode)

		body := map[string]interface{}{
			"code": "invalid-code-123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/join-by-code", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Service returns BadRequest or NotFound for invalid code
		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
	})
}

// ==================== LEAVE TENANT TESTS ====================

func TestTenantHandler_LeaveTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/leave", authMiddleware(mockAuth), handler.LeaveTenant)

		// Create tenant
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Join as another user
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Leave
		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/leave", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Owner cannot leave", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/leave", authMiddleware(mockAuth), handler.LeaveTenant)

		// Create tenant as user_dev (the token user)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Try to leave as owner
		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/leave", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ==================== GET USER TENANTS TESTS ====================

func TestTenantHandler_GetUserTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/users/me/tenants", authMiddleware(mockAuth), handler.GetUserTenants)

		// Create tenants
		ctx := context.Background()
		_, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Tenant 1",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		_, err = tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Tenant 2",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/v1/users/me/tenants", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})

	t.Run("Empty list", func(t *testing.T) {
		r, handler, _, mockAuth := setupTenantTest()
		r.GET("/v1/users/me/tenants", authMiddleware(mockAuth), handler.GetUserTenants)

		req := httptest.NewRequest("GET", "/v1/users/me/tenants", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// Data can be nil or empty array
		if data, ok := response["data"].([]interface{}); ok {
			assert.Len(t, data, 0)
		} else {
			// nil is also acceptable for empty list
			assert.Nil(t, response["data"])
		}
	})
}

// ==================== GET TENANT MEMBERS TESTS ====================

func TestTenantHandler_GetTenantMembers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/members", authMiddleware(mockAuth), handler.GetTenantMembers)

		// Create tenant
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/members", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 1) // Owner is a member
	})

	t.Run("With pagination", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/members", authMiddleware(mockAuth), handler.GetTenantMembers)

		// Create tenant
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add more members
		for i := 0; i < 5; i++ {
			tenantService.JoinTenant(ctx, "member_"+string(rune('a'+i)), tenant.ID, &domain.JoinTenantRequest{})
		}

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/members?limit=3", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		// Should return up to limit members
		assert.True(t, len(data) <= 3)
	})
}

// ==================== UPDATE MEMBER ROLE TESTS ====================

func TestTenantHandler_UpdateMemberRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/members/:memberId", authMiddleware(mockAuth), handler.UpdateMemberRole)

		// Create tenant as user_dev (token user)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add a member and capture membership ID for handler path
		member, err := tenantService.JoinTenant(ctx, "member_user", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		body := map[string]interface{}{
			"role": "admin",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/members/"+member.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden - not admin", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/members/:memberId", authMiddleware(mockAuth), handler.UpdateMemberRole)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// user_dev joins as member (actor)
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Add target member to attempt promoting
		targetMember, err := tenantService.JoinTenant(ctx, "target_user", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		body := map[string]interface{}{
			"role": "admin",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/members/"+targetMember.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Non-admin trying to update role should fail with Forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Success - Promote member to owner (transfer ownership)", func(t *testing.T) {
		t.Skip("Owner promotion via UpdateMemberRole not supported - use separate transfer ownership endpoint")
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/members/:memberId", authMiddleware(mockAuth), handler.UpdateMemberRole)

		ctx := context.Background()
		// Create tenant as initial_owner
		initialOwnerID := "initial_owner"
		tenant, err := tenantService.CreateTenant(ctx, initialOwnerID, &domain.CreateTenantRequest{
			Name:       "Owner Transfer Tenant",
			Visibility: domain.TenantVisibilityPrivate,
			AccessType: domain.TenantAccessOpen,
		})
		require.NoError(t, err)

		// Token user ('user_dev') joins as member and is promoted to admin
		promotedAdminMembership, err := tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		err = tenantService.UpdateMemberRole(ctx, initialOwnerID, tenant.ID, promotedAdminMembership.ID, domain.TenantRoleAdmin)
		require.NoError(t, err)

		// Create target user and make them a member (the one to be promoted to owner)
		targetUserID := "new_owner_candidate"
		targetMembership, err := tenantService.JoinTenant(ctx, targetUserID, tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// 'user_dev' (admin) tries to promote 'new_owner_candidate' to owner
		body := map[string]interface{}{
			"role": domain.TenantRoleOwner, // Promote to owner
		}
		jsonBody, _ := json.Marshal(body)

		// 'valid-token' maps to user_dev who was promoted to admin above
		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/members/"+targetMembership.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Member role updated successfully", response["message"])

		// Verify ownership transfer
		updatedTenant, err := tenantService.GetTenant(ctx, tenant.ID)
		require.NoError(t, err)
		assert.Equal(t, targetUserID, updatedTenant.OwnerID, "New owner ID should match target user")

		// Verify old owner's role is now admin
		members, _, err := tenantService.GetTenantMembers(ctx, tenant.ID, &domain.ListTenantMembersRequest{Limit: 100})
		require.NoError(t, err)

		var oldOwnerMember *domain.TenantMember
		for _, m := range members {
			if m.UserID == initialOwnerID {
				oldOwnerMember = m
				break
			}
		}
		require.NotNil(t, oldOwnerMember, "Old owner's membership should exist")
		assert.Equal(t, domain.TenantRoleAdmin, oldOwnerMember.Role, "Old owner should be demoted to admin")
	})
}

// ==================== REMOVE MEMBER TESTS ====================

func TestTenantHandler_RemoveMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/members/:memberId", authMiddleware(mockAuth), handler.RemoveMember)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add a member and use returned membership ID
		memberToRemove, err := tenantService.JoinTenant(ctx, "member_to_remove", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/members/"+memberToRemove.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Owner should be able to remove member
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failure - Cannot remove self", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/members/:memberId", authMiddleware(mockAuth), handler.RemoveMember)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Get owner's membership via GetTenantMembers
		members, _, err := tenantService.GetTenantMembers(ctx, tenant.ID, &domain.ListTenantMembersRequest{Limit: 100})
		require.NoError(t, err)
		var ownerMembershipID string
		for _, m := range members {
			if m.UserID == "user_dev" {
				ownerMembershipID = m.ID
				break
			}
		}
		require.NotEmpty(t, ownerMembershipID, "Owner membership should exist")

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/members/"+ownerMembershipID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should not be able to remove self, especially if owner
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ==================== BAN MEMBER TESTS ====================

func TestTenantHandler_BanMember(t *testing.T) {
	t.Run("Success_OwnerBansMember", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.BanMember)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add a member to ban
		memberToBan, err := tenantService.JoinTenant(ctx, "member_to_ban", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/members/"+memberToBan.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Owner should be able to ban member
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Member banned successfully", response["message"])
	})

	t.Run("Failure_MemberCannotBan", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.BanMember)

		// Create tenant as different user (not the token user)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "different_owner", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add token user (user_dev) as a regular member
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Add another member to try to ban
		memberToBan, err := tenantService.JoinTenant(ctx, "member_to_ban", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/members/"+memberToBan.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Regular member should NOT be able to ban
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Failure_CannotBanSelf", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.BanMember)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "different_owner", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Promote user_dev to admin so they can try to ban
		tokenUserMember, err := tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		err = tenantService.UpdateMemberRole(ctx, "different_owner", tenant.ID, tokenUserMember.ID, domain.TenantRoleAdmin)
		require.NoError(t, err)

		// Try to ban self
		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/members/"+tokenUserMember.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should NOT be able to ban self
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Failure_BannedUserCannotRejoin", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.BanMember)
		r.POST("/v1/tenants/:tenantId/join", authMiddleware(mockAuth), handler.JoinTenant)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Ban member using service directly
		memberToBan, err := tenantService.JoinTenant(ctx, "test_user_2", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		err = tenantService.BanMember(ctx, "user_dev", tenant.ID, memberToBan.ID)
		require.NoError(t, err)

		// Try to rejoin using member-token (test_user_2)
		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer member-token") // Maps to test_user_2
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Banned user should NOT be able to rejoin
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ==================== UNBAN MEMBER TESTS ====================

func TestTenantHandler_UnbanMember(t *testing.T) {
	t.Run("Success_AdminUnbansMember", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.UnbanMember)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add and ban a member
		memberToBan, err := tenantService.JoinTenant(ctx, "member_to_ban", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		err = tenantService.BanMember(ctx, "user_dev", tenant.ID, memberToBan.ID)
		require.NoError(t, err)

		// Unban the member
		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/members/"+memberToBan.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Member unbanned successfully", response["message"])
	})

	t.Run("Failure_MemberCannotUnban", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.UnbanMember)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "different_owner", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add and ban a member
		memberToBan, err := tenantService.JoinTenant(ctx, "member_to_ban", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		err = tenantService.BanMember(ctx, "different_owner", tenant.ID, memberToBan.ID)
		require.NoError(t, err)

		// Add token user (user_dev) as a regular member
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Try to unban as regular member
		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/members/"+memberToBan.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Regular member should NOT be able to unban
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Failure_UnbanNotBannedMember", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/members/:memberId/ban", authMiddleware(mockAuth), handler.UnbanMember)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add a member but don't ban
		member, err := tenantService.JoinTenant(ctx, "normal_member", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Try to unban a non-banned member
		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/members/"+member.ID+"/ban", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should fail - member is not banned
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ==================== LIST BANNED MEMBERS TESTS ====================

func TestTenantHandler_ListBannedMembers(t *testing.T) {
	t.Run("Success_ModeratorListsBanned", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/members/banned", authMiddleware(mockAuth), handler.ListBannedMembers)

		// Create tenant as user_dev (token user is owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add and ban some members
		member1, err := tenantService.JoinTenant(ctx, "banned_user_1", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)
		member2, err := tenantService.JoinTenant(ctx, "banned_user_2", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		err = tenantService.BanMember(ctx, "user_dev", tenant.ID, member1.ID)
		require.NoError(t, err)
		err = tenantService.BanMember(ctx, "user_dev", tenant.ID, member2.ID)
		require.NoError(t, err)

		// List banned members
		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/members/banned", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		members := response["members"].([]interface{})
		assert.Len(t, members, 2)
	})

	t.Run("Failure_MemberCannotListBanned", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/members/banned", authMiddleware(mockAuth), handler.ListBannedMembers)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "different_owner", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Add token user (user_dev) as a regular member
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		// Try to list banned members as regular member
		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/members/banned", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Regular member should NOT be able to list banned
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// ==================== INVITE TESTS ====================

func TestTenantHandler_CreateInvite(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/invites", authMiddleware(mockAuth), handler.CreateInvite)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"max_uses":   5,
			"expires_in": 3600,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/invites", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["code"])
	})
}

func TestTenantHandler_ListInvites(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/invites", authMiddleware(mockAuth), handler.ListInvites)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Create some invites
		tenantService.CreateInvite(ctx, "user_dev", tenant.ID, &domain.CreateTenantInviteRequest{MaxUses: 5, ExpiresIn: 3600})
		tenantService.CreateInvite(ctx, "user_dev", tenant.ID, &domain.CreateTenantInviteRequest{MaxUses: 10, ExpiresIn: 7200})

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/invites", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

func TestTenantHandler_RevokeInvite(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/invites/:inviteId", authMiddleware(mockAuth), handler.RevokeInvite)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Create invite
		invite, err := tenantService.CreateInvite(ctx, "user_dev", tenant.ID, &domain.CreateTenantInviteRequest{MaxUses: 5, ExpiresIn: 3600})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/invites/"+invite.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== ANNOUNCEMENT TESTS ====================

func TestTenantHandler_CreateAnnouncement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/announcements", authMiddleware(mockAuth), handler.CreateAnnouncement)

		// Create tenant as user_dev (owner can create announcements)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"title":   "Test Announcement",
			"content": "This is a test announcement",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/announcements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Announcement", data["title"])
	})

	t.Run("Forbidden - not admin", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/announcements", authMiddleware(mockAuth), handler.CreateAnnouncement)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// user_dev joins as member
		_, err = tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})
		require.NoError(t, err)

		body := map[string]interface{}{
			"title":   "Unauthorized Announcement",
			"content": "Should fail",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/announcements", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestTenantHandler_ListAnnouncements(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/announcements", authMiddleware(mockAuth), handler.ListAnnouncements)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Create announcements
		tenantService.CreateAnnouncement(ctx, "user_dev", tenant.ID, &domain.CreateAnnouncementRequest{
			Title:   "Announcement 1",
			Content: "Content 1",
		})
		tenantService.CreateAnnouncement(ctx, "user_dev", tenant.ID, &domain.CreateAnnouncementRequest{
			Title:   "Announcement 2",
			Content: "Content 2",
		})

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/announcements", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

func TestTenantHandler_DeleteAnnouncement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/announcements/:announcementId", authMiddleware(mockAuth), handler.DeleteAnnouncement)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Create announcement
		ann, err := tenantService.CreateAnnouncement(ctx, "user_dev", tenant.ID, &domain.CreateAnnouncementRequest{
			Title:   "To Delete",
			Content: "Will be deleted",
		})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/announcements/"+ann.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== CHAT TESTS ====================

func TestTenantHandler_SendChatMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/chat/messages", authMiddleware(mockAuth), handler.SendChatMessage)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"content": "Hello, this is a test message!",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/chat/messages", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Hello, this is a test message!", data["content"])
	})

	t.Run("Forbidden - not a member", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.POST("/v1/tenants/:tenantId/chat/messages", authMiddleware(mockAuth), handler.SendChatMessage)

		// Create tenant as different user
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Private Tenant",
			Visibility: "private",
			AccessType: "invite_only",
		})
		require.NoError(t, err)

		body := map[string]interface{}{
			"content": "Unauthorized message",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/tenants/"+tenant.ID+"/chat/messages", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestTenantHandler_GetChatHistory(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.GET("/v1/tenants/:tenantId/chat/messages", authMiddleware(mockAuth), handler.GetChatHistory)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Send some messages
		tenantService.SendChatMessage(ctx, "user_dev", tenant.ID, &domain.SendChatMessageRequest{Content: "Message 1"})
		time.Sleep(1 * time.Millisecond)
		tenantService.SendChatMessage(ctx, "user_dev", tenant.ID, &domain.SendChatMessageRequest{Content: "Message 2"})

		req := httptest.NewRequest("GET", "/v1/tenants/"+tenant.ID+"/chat/messages", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].([]interface{})
		assert.Len(t, data, 2)
	})
}

func TestTenantHandler_DeleteChatMessage(t *testing.T) {
	t.Run("Success - own message", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.DELETE("/v1/tenants/:tenantId/chat/messages/:messageId", authMiddleware(mockAuth), handler.DeleteChatMessage)

		// Create tenant as user_dev
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Send a message
		msg, err := tenantService.SendChatMessage(ctx, "user_dev", tenant.ID, &domain.SendChatMessageRequest{Content: "To delete"})
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/v1/tenants/"+tenant.ID+"/chat/messages/"+msg.ID, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTenantHandler_UpdateAnnouncement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/announcements/:announcementId", authMiddleware(mockAuth), handler.UpdateAnnouncement)

		// Create tenant as user_dev (owner)
		ctx := context.Background()
		tenant, err := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})
		require.NoError(t, err)

		// Create an announcement
		announcement, err := tenantService.CreateAnnouncement(ctx, "user_dev", tenant.ID, &domain.CreateAnnouncementRequest{
			Title:   "Original Title",
			Content: "Original Content",
		})
		require.NoError(t, err)

		// Update the announcement
		body := map[string]interface{}{
			"title":   "Updated Title",
			"content": "Updated Content",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/announcements/"+announcement.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/announcements/:announcementId", authMiddleware(mockAuth), handler.UpdateAnnouncement)

		ctx := context.Background()
		tenant, _ := tenantService.CreateTenant(ctx, "user_dev", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/announcements/some-id", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forbidden - not admin", func(t *testing.T) {
		r, handler, tenantService, mockAuth := setupTenantTest()
		r.PATCH("/v1/tenants/:tenantId/announcements/:announcementId", authMiddleware(mockAuth), handler.UpdateAnnouncement)

		// Create tenant as different owner
		ctx := context.Background()
		tenant, _ := tenantService.CreateTenant(ctx, "owner_user", &domain.CreateTenantRequest{
			Name:       "Test Tenant",
			Visibility: "public",
			AccessType: "open",
		})

		// user_dev joins as member
		tenantService.JoinTenant(ctx, "user_dev", tenant.ID, &domain.JoinTenantRequest{})

		// Create announcement as owner
		announcement, _ := tenantService.CreateAnnouncement(ctx, "owner_user", tenant.ID, &domain.CreateAnnouncementRequest{
			Title:   "Owner Announcement",
			Content: "Content",
		})

		// user_dev (member) tries to update
		body := map[string]interface{}{
			"title": "Hacked Title",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/tenants/"+tenant.ID+"/announcements/"+announcement.ID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
