package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)
	handler := NewAuthHandler(authService, nil)

	// Dummy auth middleware
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	}

	// Should not panic
	RegisterAuthRoutes(r, handler, authMiddleware)

	// Verify routes are registered
	routes := r.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	assert.True(t, routePaths["POST:/v1/auth/register"], "register route should be registered")
	assert.True(t, routePaths["POST:/v1/auth/login"], "login route should be registered")
	assert.True(t, routePaths["POST:/v1/auth/refresh"], "refresh route should be registered")
	assert.True(t, routePaths["POST:/v1/auth/logout"], "logout route should be registered")
	assert.True(t, routePaths["GET:/v1/auth/me"], "me route should be registered")
}

func TestRegisterChatRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	chatRepo := repository.NewInMemoryChatRepository()
	chatService := service.NewChatService(chatRepo, nil)
	handler := NewChatHandler(chatService)

	// Dummy auth middleware
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("tenant_id", "test-tenant")
		c.Next()
	}

	// Should not panic
	RegisterChatRoutes(r, handler, authMiddleware)

	// Verify routes are registered
	routes := r.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	assert.True(t, routePaths["GET:/v1/chat"], "list messages route should be registered")
	assert.True(t, routePaths["POST:/v1/chat"], "create message route should be registered")
}

func TestRegisterDeviceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, config.WireGuardConfig{})
	handler := NewDeviceHandler(deviceService)

	// Dummy auth middleware
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("tenant_id", "test-tenant")
		c.Next()
	}

	// Should not panic
	RegisterDeviceRoutes(r, handler, authMiddleware)

	// Verify routes are registered
	routes := r.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	assert.True(t, routePaths["POST:/v1/devices"], "register device route should be registered")
	assert.True(t, routePaths["GET:/v1/devices"], "list devices route should be registered")
	assert.True(t, routePaths["GET:/v1/devices/:id"], "get device route should be registered")
}

func TestRegisterInviteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	inviteRepo := repository.NewInMemoryInviteTokenRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	inviteService := service.NewInviteService(inviteRepo, networkRepo, membershipRepo, "http://localhost")
	handler := NewInviteHandler(inviteService)

	// Dummy auth middleware
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("tenant_id", "test-tenant")
		c.Next()
	}

	// Should not panic
	RegisterInviteRoutes(r, handler, authMiddleware)

	// Verify routes are registered
	routes := r.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	assert.True(t, routePaths["POST:/v1/networks/:id/invites"], "create invite route should be registered")
	assert.True(t, routePaths["GET:/v1/networks/:id/invites"], "list invites route should be registered")
}
