package handler

import (
	"github.com/gin-gonic/gin"
)

// ══════════════════════════════════════════════════════════════════════════════
// SECTION ROUTES
// ══════════════════════════════════════════════════════════════════════════════

// RegisterSectionRoutes registers all section-related routes
// Note: RBAC permission checks are handled in the service layer
func RegisterSectionRoutes(r *gin.RouterGroup, handler *SectionHandler, authMiddleware gin.HandlerFunc) {
	// Public routes (none for sections - all require auth)

	// Protected routes - require authentication
	sections := r.Group("/api/v2/servers/:tenantID/sections")
	sections.Use(authMiddleware)
	{
		sections.POST("", handler.Create)
		sections.GET("", handler.List)
		sections.PATCH("/positions", handler.UpdatePositions)
	}

	// Individual section routes
	section := r.Group("/api/v2/sections")
	section.Use(authMiddleware)
	{
		section.GET("/:sectionID", handler.GetByID)
		section.PATCH("/:sectionID", handler.Update)
		section.DELETE("/:sectionID", handler.Delete)
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CHANNEL ROUTES
// ══════════════════════════════════════════════════════════════════════════════

// RegisterChannelRoutes registers all channel-related routes
func RegisterChannelRoutes(r *gin.RouterGroup, handler *ChannelHandler, authMiddleware gin.HandlerFunc) {
	// Server-level channels
	serverChannels := r.Group("/api/v2/servers/:tenantID/channels")
	serverChannels.Use(authMiddleware)
	{
		serverChannels.POST("", handler.CreateInServer)
		serverChannels.GET("", handler.ListInServer)
	}

	// Section-level channels
	sectionChannels := r.Group("/api/v2/sections/:sectionID/channels")
	sectionChannels.Use(authMiddleware)
	{
		sectionChannels.POST("", handler.CreateInSection)
		sectionChannels.GET("", handler.ListInSection)
	}

	// Individual channel routes
	channels := r.Group("/api/v2/channels")
	channels.Use(authMiddleware)
	{
		channels.GET("/:channelID", handler.GetByID)
		channels.PATCH("/:channelID", handler.Update)
		channels.DELETE("/:channelID", handler.Delete)
	}
}
