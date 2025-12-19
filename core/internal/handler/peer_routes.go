package handler

import (
	"github.com/gin-gonic/gin"
)

// RegisterPeerRoutes registers all peer-related routes
func RegisterPeerRoutes(r *gin.Engine, h *PeerHandler, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/v1")
	v1.Use(authMiddleware)
	{
		// Peer CRUD operations
		v1.POST("/peers", h.CreatePeer)
		v1.GET("/peers/:id", h.GetPeer)
		v1.PATCH("/peers/:id", h.UpdatePeer)
		v1.DELETE("/peers/:id", h.DeletePeer)

		// Peer statistics
		v1.GET("/peers/:id/stats", h.GetPeerStats)
		v1.POST("/peers/:id/stats", h.UpdatePeerStats)

		// Key rotation
		v1.POST("/peers/:id/rotate-keys", h.RotatePeerKeys)

		// Network-specific peer routes
		v1.GET("/networks/:id/peers", h.GetPeersByNetwork)
		v1.GET("/networks/:id/peers/active", h.GetActivePeers)
		v1.GET("/networks/:id/peers/stats", h.GetNetworkPeerStats)

		// Device-specific peer routes
		v1.GET("/devices/:id/peers", h.GetPeersByDevice)
	}
}
