package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// ══════════════════════════════════════════════════════════════════════════════
// RBAC MIDDLEWARE
// ══════════════════════════════════════════════════════════════════════════════
// Discord-style permission checking middleware for Gin framework
// Uses PermissionResolver to check user permissions in tenants/channels

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for user ID (set by auth middleware)
	UserIDKey ContextKey = "user_id"

	// TenantIDKey is the context key for tenant ID (extracted from request)
	TenantIDKey ContextKey = "tenant_id"

	// ChannelIDKey is the context key for channel ID (extracted from request)
	ChannelIDKey ContextKey = "channel_id"

	// PermissionResultKey is the context key for permission check result
	PermissionResultKey ContextKey = "permission_result"
)

// RBACMiddleware provides permission-based access control
type RBACMiddleware struct {
	permissionResolver *service.PermissionResolver
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(permissionResolver *service.PermissionResolver) *RBACMiddleware {
	return &RBACMiddleware{
		permissionResolver: permissionResolver,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CORE MIDDLEWARE FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

// RequirePermission checks if user has a specific permission in tenant/channel
// Usage in routes:
//   router.GET("/channels/:channelID", rbac.RequirePermission("channel.view"), handler)
//   router.POST("/channels/:channelID/messages", rbac.RequirePermission("channel.send_messages"), handler)
func (m *RBACMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Validate permission code
		if !m.permissionResolver.IsValidPermission(permission) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid permission configuration",
				"code":  "INVALID_PERMISSION",
			})
			c.Abort()
			return
		}

		// Step 2: Extract user ID from context (set by auth middleware)
		userID, exists := c.Get(string(UserIDKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user session",
				"code":  "INVALID_SESSION",
			})
			c.Abort()
			return
		}

		// Step 3: Extract tenant ID from request
		tenantID := m.extractTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "tenant context required",
				"code":  "TENANT_REQUIRED",
			})
			c.Abort()
			return
		}

		// Step 4: Extract channel ID (optional - depends on permission type)
		channelID := m.extractChannelID(c)

		// Step 5: Check permission
		result, err := m.permissionResolver.CheckPermission(
			c.Request.Context(),
			userIDStr,
			tenantID,
			channelID,
			permission,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check permissions",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		// Step 6: Store result in context for handlers to access
		c.Set(string(PermissionResultKey), result)

		// Step 7: Enforce permission
		if !result.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "insufficient permissions",
				"code":   "FORBIDDEN",
				"reason": result.Reason,
			})
			c.Abort()
			return
		}

		// Permission granted - continue to handler
		c.Next()
	}
}

// RequireAnyPermission checks if user has ANY of the specified permissions
// Usage: rbac.RequireAnyPermission("channel.view", "channel.manage_messages")
func (m *RBACMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate all permissions first
		for _, perm := range permissions {
			if !m.permissionResolver.IsValidPermission(perm) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "invalid permission configuration",
					"code":  "INVALID_PERMISSION",
				})
				c.Abort()
				return
			}
		}

		// Extract context
		userID, exists := c.Get(string(UserIDKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		userIDStr := userID.(string)
		tenantID := m.extractTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "tenant context required",
				"code":  "TENANT_REQUIRED",
			})
			c.Abort()
			return
		}

		channelID := m.extractChannelID(c)

		// Check multiple permissions (optimized batch check)
		results, err := m.permissionResolver.CheckMultiplePermissions(
			c.Request.Context(),
			userIDStr,
			tenantID,
			channelID,
			permissions,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check permissions",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		// Check if user has ANY of the permissions
		hasAnyPermission := false
		var grantedPermission string

		for perm, result := range results {
			if result.Allowed {
				hasAnyPermission = true
				grantedPermission = perm
				c.Set(string(PermissionResultKey), result)
				break
			}
		}

		if !hasAnyPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		// Store which permission was granted
		c.Set("granted_permission", grantedPermission)
		c.Next()
	}
}

// RequireAllPermissions checks if user has ALL of the specified permissions
// Usage: rbac.RequireAllPermissions("channel.view", "channel.send_messages")
func (m *RBACMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate all permissions first
		for _, perm := range permissions {
			if !m.permissionResolver.IsValidPermission(perm) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "invalid permission configuration",
					"code":  "INVALID_PERMISSION",
				})
				c.Abort()
				return
			}
		}

		// Extract context
		userID, exists := c.Get(string(UserIDKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		userIDStr := userID.(string)
		tenantID := m.extractTenantID(c)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "tenant context required",
				"code":  "TENANT_REQUIRED",
			})
			c.Abort()
			return
		}

		channelID := m.extractChannelID(c)

		// Check multiple permissions (optimized batch check)
		results, err := m.permissionResolver.CheckMultiplePermissions(
			c.Request.Context(),
			userIDStr,
			tenantID,
			channelID,
			permissions,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check permissions",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		// Check if user has ALL permissions
		missingPermissions := []string{}

		for perm, result := range results {
			if !result.Allowed {
				missingPermissions = append(missingPermissions, perm)
			}
		}

		if len(missingPermissions) > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "insufficient permissions",
				"code":                "FORBIDDEN",
				"missing_permissions": missingPermissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

// extractTenantID extracts tenant ID from request
// Priority: URL param > header > query param
func (m *RBACMiddleware) extractTenantID(c *gin.Context) string {
	// Try URL parameter (e.g., /tenants/:tenantID/...)
	if tenantID := c.Param("tenantID"); tenantID != "" {
		return tenantID
	}
	if tenantID := c.Param("tenant_id"); tenantID != "" {
		return tenantID
	}

	// Try header
	if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	// Try query parameter
	if tenantID := c.Query("tenant_id"); tenantID != "" {
		return tenantID
	}

	return ""
}

// extractChannelID extracts channel ID from request
// Priority: URL param > body (for POST/PUT) > query param
func (m *RBACMiddleware) extractChannelID(c *gin.Context) string {
	// Try URL parameter (e.g., /channels/:channelID/...)
	if channelID := c.Param("channelID"); channelID != "" {
		return channelID
	}
	if channelID := c.Param("channel_id"); channelID != "" {
		return channelID
	}

	// Try query parameter
	if channelID := c.Query("channel_id"); channelID != "" {
		return channelID
	}

	return ""
}

// ══════════════════════════════════════════════════════════════════════════════
// CONTEXT HELPERS (for handlers to use)
// ══════════════════════════════════════════════════════════════════════════════

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetTenantID retrieves the tenant ID from context
func GetTenantID(c *gin.Context) string {
	// First check if it's stored in context
	if tenantID, exists := c.Get(string(TenantIDKey)); exists {
		if tenantIDStr, ok := tenantID.(string); ok {
			return tenantIDStr
		}
	}

	// Otherwise extract from request
	rbac := &RBACMiddleware{}
	return rbac.extractTenantID(c)
}

// GetChannelID retrieves the channel ID from context
func GetChannelID(c *gin.Context) string {
	// First check if it's stored in context
	if channelID, exists := c.Get(string(ChannelIDKey)); exists {
		if channelIDStr, ok := channelID.(string); ok {
			return channelIDStr
		}
	}

	// Otherwise extract from request
	rbac := &RBACMiddleware{}
	return rbac.extractChannelID(c)
}

// GetPermissionResult retrieves the permission check result from context
func GetPermissionResult(c *gin.Context) (*service.PermissionResult, bool) {
	result, exists := c.Get(string(PermissionResultKey))
	if !exists {
		return nil, false
	}

	permResult, ok := result.(*service.PermissionResult)
	return permResult, ok
}

// ══════════════════════════════════════════════════════════════════════════════
// UTILITY MIDDLEWARE
// ══════════════════════════════════════════════════════════════════════════════

// SetTenantContext middleware extracts and stores tenant ID in context
// Useful for routes that don't need permission checks but need tenant context
func SetTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		rbac := &RBACMiddleware{}
		tenantID := rbac.extractTenantID(c)
		if tenantID != "" {
			c.Set(string(TenantIDKey), tenantID)
		}
		c.Next()
	}
}

// SetChannelContext middleware extracts and stores channel ID in context
func SetChannelContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		rbac := &RBACMiddleware{}
		channelID := rbac.extractChannelID(c)
		if channelID != "" {
			c.Set(string(ChannelIDKey), channelID)
		}
		c.Next()
	}
}

// AdminOnly middleware ensures user has ADMINISTRATOR permission
func (m *RBACMiddleware) AdminOnly() gin.HandlerFunc {
	return m.RequirePermission(domain.PermissionAdministrator)
}
