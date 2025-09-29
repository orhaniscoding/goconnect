package handler

import (
	"context"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// AuthMiddleware validates JWT tokens and extracts user information
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authorization header required", nil))
			c.Abort()
			return
		}

		// Extract bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Invalid authorization header format", nil))
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Implement proper JWT validation
		// For now, use mock validation
		userID, isAdmin, err := validateToken(token)
		if err != nil {
			errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Invalid or expired token", nil))
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", userID)
		c.Set("is_admin", isAdmin)
		c.Next()
	}
}

// RequireAdmin ensures the user has admin privileges
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			errorResponse(c, domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware generates and adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-Id", requestID)
		c.Set("request_id", requestID)
		// also propagate into request context for downstream services
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Allow specific origins (simplified for development)
		allowedOrigins := []string{
			"https://app.goconnect.example",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-Id")
	c.Header("Access-Control-Expose-Headers", "X-Request-Id, Retry-After")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RoleMiddleware resolves the actor's membership role (if any) for a network route and injects into context.
// It expects a membership repository (in-memory for now). For non-network paths it no-ops.
func RoleMiddleware(mrepo repository.MembershipRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only attempt if path contains /v1/networks/{id}
		parts := strings.Split(c.Request.URL.Path, "/")
		// expect: "", "v1", "networks", ":id", ...
		if len(parts) < 4 || parts[1] != "v1" || parts[2] != "networks" {
			c.Next()
			return
		}
		networkID := parts[3]
		userID, _ := c.Get("user_id")
		uid, _ := userID.(string)
		// If auth middleware not yet executed (ordering), attempt lightweight token parse
		if uid == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				seg := strings.SplitN(authHeader, " ", 2)
				if len(seg) == 2 && seg[0] == "Bearer" {
					if tid, isAdm, err := validateToken(seg[1]); err == nil {
						uid = tid
						// If global admin, short-circuit by granting elevated role without membership lookup
						if isAdm {
							c.Set("membership_role", domain.RoleOwner)
							c.Set("user_id", uid)
							c.Next()
							return
						}
					}
				}
			}
		}
		role := domain.RoleMember // default implicit role if authenticated but not member
		if uid != "" {
			if m, err := mrepo.Get(c.Request.Context(), networkID, uid); err == nil {
				role = m.Role
			}
		}
		c.Set("membership_role", role)
		c.Next()
	}
}

// RequireNetworkAdmin ensures membership role is admin or owner (for network-scoped mutations) OR is_admin flag.
func RequireNetworkAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// global admin bypass (is_admin from token)
		if ia, ok := c.Get("is_admin"); ok && ia.(bool) {
			c.Next(); return
		}
		roleAny, ok := c.Get("membership_role")
		if !ok {
			errorResponse(c, domain.NewError(domain.ErrForbidden, "Membership role required", nil))
			c.Abort(); return
		}
		role, _ := roleAny.(domain.MembershipRole)
		if role != domain.RoleAdmin && role != domain.RoleOwner {
			errorResponse(c, domain.NewError(domain.ErrForbidden, "Administrator privileges required", nil))
			c.Abort(); return
		}
		c.Next()
	}
}


// validateToken validates JWT token and returns user info
// TODO: Replace with proper JWT validation
func validateToken(token string) (userID string, isAdmin bool, err error) {
	// Mock implementation for development
	switch token {
	case "dev":
		return "user_dev", false, nil
	case "admin":
		return "admin_dev", true, nil
	default:
		return "", false, domain.NewError(domain.ErrUnauthorized, "Invalid token", nil)
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return domain.GenerateNetworkID() // Reuse the ID generation logic
}

// errorResponse sends a standardized error response
func errorResponse(c *gin.Context, derr *domain.Error) {
    status := derr.ToHTTPStatus()
    if derr.Code == domain.ErrForbidden || derr.Code == domain.ErrUnauthorized {
        // unify outward code while preserving computed status (401 vs 403)
        derr = &domain.Error{Code: domain.ErrNotAuthorized, Message: derr.Message, Details: derr.Details, RetryAfter: derr.RetryAfter}
    }
    c.JSON(status, derr)
}
