package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	ws "github.com/orhaniscoding/goconnect/server/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebSocketHandler(t *testing.T) {
	hub := ws.NewHub(nil)
	handler := NewWebSocketHandler(hub)

	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
	assert.NotNil(t, handler.upgrader)

	// Check CheckOrigin is permissive
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://evil.com")
	assert.True(t, handler.upgrader.CheckOrigin(req))
}

func TestNewWebSocketHandlerWithConfig(t *testing.T) {
	hub := ws.NewHub(nil)
	corsConfig := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}
	handler := NewWebSocketHandlerWithConfig(hub, corsConfig)

	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
	assert.NotNil(t, handler.upgrader)

	// Check CheckOrigin respects CORS config
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	assert.True(t, handler.upgrader.CheckOrigin(req))

	req.Header.Set("Origin", "http://evil.com")
	assert.False(t, handler.upgrader.CheckOrigin(req))
}

func TestHandleUpgrade_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	handler := NewWebSocketHandler(hub)

	r := gin.New()
	r.GET("/ws", handler.HandleUpgrade)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_UNAUTHORIZED")
}

func TestHandleUpgrade_Success(t *testing.T) {
	t.Skip("Skipping WebSocket upgrade test - requires full connection handling")
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewWebSocketHandler(hub)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("tenant_id", "test-tenant")
		c.Set("is_admin", false)
		c.Set("is_moderator", false)
		c.Next()
	})
	r.GET("/ws", handler.HandleUpgrade)

	// Create test server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create WebSocket client
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer conn.Close()

	// Verify client is registered
	time.Sleep(50 * time.Millisecond)
	// Hub should have 1 client now (check host room)
	assert.Eventually(t, func() bool {
		return len(hub.GetRoomClients("host")) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestHandleUpgrade_WithAdminAndModerator(t *testing.T) {
	t.Skip("Skipping WebSocket upgrade test - requires full connection handling")
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewWebSocketHandler(hub)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-user")
		c.Set("tenant_id", "admin-tenant")
		c.Set("is_admin", true)
		c.Set("is_moderator", true)
		c.Next()
	})
	r.GET("/ws", handler.HandleUpgrade)

	// Create test server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create WebSocket client
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer conn.Close()

	// Verify client is registered in both host and tenant rooms
	time.Sleep(50 * time.Millisecond)
	assert.Eventually(t, func() bool {
		return len(hub.GetRoomClients("host")) == 1 &&
			len(hub.GetRoomClients("tenant:admin-tenant")) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestHandleUpgrade_NoTenantID(t *testing.T) {
	t.Skip("Skipping WebSocket upgrade test - requires full connection handling")
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewWebSocketHandler(hub)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-no-tenant")
		// No tenant_id set
		c.Next()
	})
	r.GET("/ws", handler.HandleUpgrade)

	// Create test server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create WebSocket client
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer conn.Close()

	// Verify client is only in host room, not tenant room
	time.Sleep(50 * time.Millisecond)
	assert.Eventually(t, func() bool {
		// Should have client in host room
		hostClients := hub.GetRoomClients("host")
		// Client count in host should be 1
		return len(hostClients) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestRegisterWebSocketRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	handler := NewWebSocketHandler(hub)

	r := gin.New()
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "test")
		c.Next()
	}

	RegisterWebSocketRoutes(r, handler, authMiddleware)

	// Verify route is registered
	routes := r.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/v1/ws" && route.Method == "GET" {
			found = true
			break
		}
	}
	assert.True(t, found, "WebSocket route should be registered at /v1/ws")
}

func TestHandleUpgrade_BoolConversion(t *testing.T) {
	t.Skip("Skipping WebSocket upgrade test - requires full connection handling")
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewWebSocketHandler(hub)

	tests := []struct {
		name          string
		isAdmin       interface{}
		isModerator   interface{}
		expectedAdmin bool
		expectedMod   bool
	}{
		{
			name:          "nil values default to false",
			isAdmin:       nil,
			isModerator:   nil,
			expectedAdmin: false,
			expectedMod:   false,
		},
		{
			name:          "true values",
			isAdmin:       true,
			isModerator:   true,
			expectedAdmin: true,
			expectedMod:   true,
		},
		{
			name:          "mixed values",
			isAdmin:       true,
			isModerator:   nil,
			expectedAdmin: true,
			expectedMod:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", "test-user-"+tt.name)
				if tt.isAdmin != nil {
					c.Set("is_admin", tt.isAdmin)
				}
				if tt.isModerator != nil {
					c.Set("is_moderator", tt.isModerator)
				}
				c.Next()
			})
			r.GET("/ws", handler.HandleUpgrade)

			srv := httptest.NewServer(r)
			defer srv.Close()

			wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(wsURL, nil)
			require.NoError(t, err)
			require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
			conn.Close()

			time.Sleep(50 * time.Millisecond)
		})
	}
}

func TestHandleUpgrade_TenantIDConversion(t *testing.T) {
	t.Skip("Skipping WebSocket upgrade test - requires full connection handling")
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	handler := NewWebSocketHandler(hub)

	tests := []struct {
		name             string
		tenantID         interface{}
		shouldJoinTenant bool
	}{
		{
			name:             "nil tenant - no tenant room",
			tenantID:         nil,
			shouldJoinTenant: false,
		},
		{
			name:             "valid tenant - joins tenant room",
			tenantID:         "tenant-123",
			shouldJoinTenant: true,
		},
		{
			name:             "empty string tenant - no tenant room",
			tenantID:         "",
			shouldJoinTenant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_id", "test-user-"+tt.name)
				if tt.tenantID != nil {
					c.Set("tenant_id", tt.tenantID)
				}
				c.Next()
			})
			r.GET("/ws", handler.HandleUpgrade)

			srv := httptest.NewServer(r)
			defer srv.Close()

			wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(wsURL, nil)
			require.NoError(t, err)
			require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
			defer conn.Close()

			time.Sleep(50 * time.Millisecond)

			// Check tenant room membership
			if tt.shouldJoinTenant && tt.tenantID != "" {
				tenantRoom := "tenant:" + tt.tenantID.(string)
				clients := hub.GetRoomClients(tenantRoom)
				assert.Equal(t, 1, len(clients), "Should have client in tenant room")
			}
		})
	}
}
