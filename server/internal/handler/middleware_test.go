package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestRequireModerator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Admin user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", true)
		c.Set("is_moderator", false)

		middleware := RequireModerator()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Success - Moderator user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		c.Set("is_moderator", true)

		middleware := RequireModerator()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Success - Admin and Moderator", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", true)
		c.Set("is_moderator", true)

		middleware := RequireModerator()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Forbidden - Regular user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		c.Set("is_moderator", false)

		middleware := RequireModerator()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - No flags set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// No flags set in context

		middleware := RequireModerator()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Admin user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", true)

		middleware := RequireAdmin()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Forbidden - Non-admin user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)

		middleware := RequireAdmin()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - No flag set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		middleware := RequireAdmin()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// Mock TokenValidator for testing
type mockTokenValidator struct {
	claims *domain.TokenClaims
	err    error
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	return m.claims, m.err
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		mockValidator := &mockTokenValidator{
			claims: &domain.TokenClaims{
				UserID:      "user-123",
				TenantID:    "tenant-456",
				Email:       "user@example.com",
				IsAdmin:     true,
				IsModerator: false,
			},
			err: nil,
		}

		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, "user-123", c.GetString("user_id"))
		assert.Equal(t, "tenant-456", c.GetString("tenant_id"))
		assert.True(t, c.GetBool("is_admin"))
		assert.False(t, c.GetBool("is_moderator"))
	})

	t.Run("Error - Missing Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		// No Authorization header

		mockValidator := &mockTokenValidator{}
		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Error - Invalid header format (no Bearer)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat token")

		mockValidator := &mockTokenValidator{}
		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Error - Missing token after Bearer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer")

		mockValidator := &mockTokenValidator{}
		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Error - Invalid token (validator error)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		mockValidator := &mockTokenValidator{
			claims: nil,
			err:    domain.NewError(domain.ErrInvalidToken, "Token expired", nil),
		}

		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Error - Generic validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer bad-token")

		mockValidator := &mockTokenValidator{
			claims: nil,
			err:    assert.AnError, // Generic error, not domain.Error
		}

		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success - Moderator user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer moderator-token")

		mockValidator := &mockTokenValidator{
			claims: &domain.TokenClaims{
				UserID:      "mod-123",
				TenantID:    "tenant-789",
				Email:       "mod@example.com",
				IsAdmin:     false,
				IsModerator: true,
			},
			err: nil,
		}

		middleware := AuthMiddleware(mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.False(t, c.GetBool("is_admin"))
		assert.True(t, c.GetBool("is_moderator"))
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Generate new request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		// No X-Request-Id header

		middleware := RequestIDMiddleware()
		middleware(c)

		assert.False(t, c.IsAborted())

		// Should have generated and set request ID
		requestID := c.GetString("request_id")
		assert.NotEmpty(t, requestID)

		// Should have set response header
		assert.Equal(t, requestID, w.Header().Get("X-Request-Id"))
	})

	t.Run("Success - Use provided request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		providedID := "custom-request-id-123"
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("X-Request-Id", providedID)

		middleware := RequestIDMiddleware()
		middleware(c)

		assert.False(t, c.IsAborted())

		// Should use provided request ID
		requestID := c.GetString("request_id")
		assert.Equal(t, providedID, requestID)

		// Should have set response header
		assert.Equal(t, providedID, w.Header().Get("X-Request-Id"))
	})

	t.Run("Request ID in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware := RequestIDMiddleware()
		middleware(c)

		// Should be accessible from gin context
		requestID := c.GetString("request_id")
		assert.NotEmpty(t, requestID)

		// Should also be in request context
		ctx := c.Request.Context()
		ctxRequestID := ctx.Value(requestIDKey)
		assert.Equal(t, requestID, ctxRequestID)
	})
}

func TestRequireNetworkAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Global admin bypass", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", true)
		// No membership_role set

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Success - Network owner", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		c.Set("membership_role", domain.RoleOwner)

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Success - Network admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		c.Set("membership_role", domain.RoleAdmin)

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Forbidden - Regular member", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		c.Set("membership_role", domain.RoleMember)

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - No membership role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("is_admin", false)
		// No membership_role set

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Forbidden - No context values at all", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// No values set

		middleware := RequireNetworkAdmin()
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Admin gets owner role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/networks/net-123/members", nil)
		c.Set("is_admin", true)
		c.Set("user_id", "admin-user")

		mrepo := repository.NewInMemoryMembershipRepository()
		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		role, exists := c.Get("membership_role")
		assert.True(t, exists)
		assert.Equal(t, domain.RoleOwner, role)
	})

	t.Run("Success - Member role from repository", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/networks/net-456/devices", nil)
		c.Set("is_admin", false)
		c.Set("user_id", "member-user")

		mrepo := repository.NewInMemoryMembershipRepository()
		// Create membership
		_, _ = mrepo.UpsertApproved(context.Background(), "net-456", "member-user", domain.RoleMember, time.Now())

		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		role, exists := c.Get("membership_role")
		assert.True(t, exists)
		assert.Equal(t, domain.RoleMember, role)
	})

	t.Run("Success - Network admin role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/v1/networks/net-789/join/approve", nil)
		c.Set("is_admin", false)
		c.Set("user_id", "network-admin")

		mrepo := repository.NewInMemoryMembershipRepository()
		// Create admin membership
		_, _ = mrepo.UpsertApproved(context.Background(), "net-789", "network-admin", domain.RoleAdmin, time.Now())

		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		role, exists := c.Get("membership_role")
		assert.True(t, exists)
		assert.Equal(t, domain.RoleAdmin, role)
	})

	t.Run("Success - No membership defaults to member", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/networks/net-999/info", nil)
		c.Set("is_admin", false)
		c.Set("user_id", "random-user")

		mrepo := repository.NewInMemoryMembershipRepository()
		// No membership created

		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		role, exists := c.Get("membership_role")
		assert.True(t, exists)
		assert.Equal(t, domain.RoleMember, role) // Default
	})

	t.Run("No-op - Non-network path", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/users/me", nil)
		c.Set("user_id", "user-123")

		mrepo := repository.NewInMemoryMembershipRepository()
		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		// Should not set membership_role for non-network paths
		_, exists := c.Get("membership_role")
		assert.False(t, exists)
	})

	t.Run("No-op - Short path", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1", nil)

		mrepo := repository.NewInMemoryMembershipRepository()
		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		_, exists := c.Get("membership_role")
		assert.False(t, exists)
	})

	t.Run("Success - No user_id in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/v1/networks/net-123/public", nil)
		// No user_id set (anonymous or pre-auth scenario)

		mrepo := repository.NewInMemoryMembershipRepository()
		mockValidator := &mockTokenValidator{}

		middleware := RoleMiddleware(mrepo, mockValidator)
		middleware(c)

		assert.False(t, c.IsAborted())
		role, exists := c.Get("membership_role")
		assert.True(t, exists)
		assert.Equal(t, domain.RoleMember, role) // Default for authenticated but not member
	})
}
