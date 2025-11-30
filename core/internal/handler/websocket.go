package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	ws "github.com/orhaniscoding/goconnect/server/internal/websocket"
)

// WebSocketHandler handles WebSocket upgrade requests
type WebSocketHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Permissive for backwards compatibility
				// Use NewWebSocketHandlerWithConfig for CORS-aware version
				return true
			},
		},
	}
}

// NewWebSocketHandlerWithConfig creates a WebSocket handler with CORS configuration
func NewWebSocketHandlerWithConfig(hub *ws.Hub, corsConfig *config.CORSConfig) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     CheckOrigin(corsConfig),
		},
	}
}

// HandleUpgrade handles the WebSocket upgrade request
func (h *WebSocketHandler) HandleUpgrade(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required for WebSocket connection",
		})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	isAdmin, _ := c.Get("is_admin")
	isModerator, _ := c.Get("is_moderator")

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_UPGRADE_FAILED",
			"message": "Failed to upgrade connection to WebSocket",
			"details": map[string]string{"error": err.Error()},
		})
		return
	}

	// Create client
	userIDStr := userID.(string)
	tenantIDStr := ""
	if tenantID != nil {
		tenantIDStr = tenantID.(string)
	}

	isAdminBool := false
	if isAdmin != nil {
		isAdminBool = isAdmin.(bool)
	}

	isModeratorBool := false
	if isModerator != nil {
		isModeratorBool = isModerator.(bool)
	}

	client := ws.NewClient(h.hub, conn, userIDStr, tenantIDStr, isAdminBool, isModeratorBool)

	// Register client with hub
	h.hub.Register(client)

	// Auto-join user to their tenant room (for tenant-wide broadcasts)
	if tenantIDStr != "" {
		h.hub.JoinRoom(client, "tenant:"+tenantIDStr)
	}

	// Auto-join user to global "host" room
	h.hub.JoinRoom(client, "host")

	// Start client pumps
	go client.Run(c.Request.Context())
}

// RegisterWebSocketRoutes registers WebSocket routes
func RegisterWebSocketRoutes(r *gin.Engine, handler *WebSocketHandler, authMiddleware gin.HandlerFunc) {
	// WebSocket endpoint requires authentication
	r.GET("/v1/ws", authMiddleware, handler.HandleUpgrade)
}
