package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/middleware"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// ══════════════════════════════════════════════════════════════════════════════
// CHANNEL HANDLER
// ══════════════════════════════════════════════════════════════════════════════
// HTTP handlers for channel management endpoints

type ChannelHandler struct {
	channelService *service.ChannelService
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// REQUEST/RESPONSE TYPES
// ══════════════════════════════════════════════════════════════════════════════

type CreateChannelRequest struct {
	Name        string             `json:"name" binding:"required,min=1,max=100"`
	Description string             `json:"description,omitempty" binding:"max=500"`
	Type        domain.ChannelType `json:"type" binding:"required,oneof=text voice announcement"`
	Bitrate     *int               `json:"bitrate,omitempty" binding:"omitempty,min=8000,max=384000"`
	UserLimit   *int               `json:"user_limit,omitempty" binding:"omitempty,min=0,max=99"`
	Slowmode    *int               `json:"slowmode,omitempty" binding:"omitempty,min=0,max=21600"`
	NSFW        bool               `json:"nsfw,omitempty"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
	Bitrate     *int    `json:"bitrate,omitempty" binding:"omitempty,min=8000,max=384000"`
	UserLimit   *int    `json:"user_limit,omitempty" binding:"omitempty,min=0,max=99"`
	Slowmode    *int    `json:"slowmode,omitempty" binding:"omitempty,min=0,max=21600"`
	NSFW        *bool   `json:"nsfw,omitempty"`
}

type ChannelResponse struct {
	ID          string             `json:"id"`
	TenantID    *string            `json:"tenant_id,omitempty"`
	SectionID   *string            `json:"section_id,omitempty"`
	NetworkID   *string            `json:"network_id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        domain.ChannelType `json:"type"`
	Position    int                `json:"position"`
	Bitrate     int                `json:"bitrate,omitempty"`
	UserLimit   int                `json:"user_limit,omitempty"`
	Slowmode    int                `json:"slowmode"`
	NSFW        bool               `json:"nsfw"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLERS - Server-Level Channels
// ══════════════════════════════════════════════════════════════════════════════

// CreateInServer creates a new channel at server level
// POST /api/v2/servers/:tenantID/channels
func (h *ChannelHandler) CreateInServer(c *gin.Context) {
	tenantID := c.Param("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "INVALID_REQUEST", "details": err.Error()})
		return
	}

	// Create channel at server level
	channel, err := h.channelService.Create(c.Request.Context(), service.CreateChannelInput{
		UserID:      userID,
		TenantID:    &tenantID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Bitrate:     getIntValue(req.Bitrate, 64000),
		UserLimit:   getIntValue(req.UserLimit, 0),
		Slowmode:    getIntValue(req.Slowmode, 0),
		NSFW:        req.NSFW,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toChannelResponse(channel))
}

// ListInServer lists all channels at server level
// GET /api/v2/servers/:tenantID/channels
func (h *ChannelHandler) ListInServer(c *gin.Context) {
	tenantID := c.Param("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	// Parse query parameters
	channelType := c.Query("type")
	var typeFilter *domain.ChannelType
	if channelType != "" {
		ct := domain.ChannelType(channelType)
		typeFilter = &ct
	}

	// Call service
	channels, nextCursor, err := h.channelService.List(c.Request.Context(), service.ListChannelsInput{
		UserID:   userID,
		TenantID: &tenantID,
		Type:     typeFilter,
		Limit:    50,
		Cursor:   c.Query("cursor"),
	})

	if err != nil {
		handleError(c, err)
		return
	}

	// Build response
	response := make([]ChannelResponse, len(channels))
	for i, channel := range channels {
		response[i] = toChannelResponse(&channel)
	}

	result := gin.H{
		"channels": response,
		"count":    len(response),
	}
	if nextCursor != "" {
		result["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, result)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLERS - Section-Level Channels
// ══════════════════════════════════════════════════════════════════════════════

// CreateInSection creates a new channel in a section
// POST /api/v2/sections/:sectionID/channels
func (h *ChannelHandler) CreateInSection(c *gin.Context) {
	sectionID := c.Param("sectionID")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "INVALID_REQUEST", "details": err.Error()})
		return
	}

	// Create channel in section
	channel, err := h.channelService.Create(c.Request.Context(), service.CreateChannelInput{
		UserID:      userID,
		SectionID:   &sectionID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Bitrate:     getIntValue(req.Bitrate, 64000),
		UserLimit:   getIntValue(req.UserLimit, 0),
		Slowmode:    getIntValue(req.Slowmode, 0),
		NSFW:        req.NSFW,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toChannelResponse(channel))
}

// ListInSection lists all channels in a section
// GET /api/v2/sections/:sectionID/channels
func (h *ChannelHandler) ListInSection(c *gin.Context) {
	sectionID := c.Param("sectionID")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	// Parse query parameters
	channelType := c.Query("type")
	var typeFilter *domain.ChannelType
	if channelType != "" {
		ct := domain.ChannelType(channelType)
		typeFilter = &ct
	}

	// Call service
	channels, nextCursor, err := h.channelService.List(c.Request.Context(), service.ListChannelsInput{
		UserID:    userID,
		SectionID: &sectionID,
		Type:      typeFilter,
		Limit:     50,
		Cursor:    c.Query("cursor"),
	})

	if err != nil {
		handleError(c, err)
		return
	}

	// Build response
	response := make([]ChannelResponse, len(channels))
	for i, channel := range channels {
		response[i] = toChannelResponse(&channel)
	}

	result := gin.H{
		"channels": response,
		"count":    len(response),
	}
	if nextCursor != "" {
		result["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, result)
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLERS - Channel Operations
// ══════════════════════════════════════════════════════════════════════════════

// GetByID retrieves a channel by ID
// GET /api/v2/channels/:channelID
func (h *ChannelHandler) GetByID(c *gin.Context) {
	channelID := c.Param("channelID")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), userID, channelID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toChannelResponse(channel))
}

// Update updates a channel
// PATCH /api/v2/channels/:channelID
func (h *ChannelHandler) Update(c *gin.Context) {
	channelID := c.Param("channelID")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "INVALID_REQUEST", "details": err.Error()})
		return
	}

	channel, err := h.channelService.Update(c.Request.Context(), service.UpdateChannelInput{
		UserID:      userID,
		ChannelID:   channelID,
		Name:        req.Name,
		Description: req.Description,
		Bitrate:     req.Bitrate,
		UserLimit:   req.UserLimit,
		Slowmode:    req.Slowmode,
		NSFW:        req.NSFW,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toChannelResponse(channel))
}

// Delete deletes a channel
// DELETE /api/v2/channels/:channelID
func (h *ChannelHandler) Delete(c *gin.Context) {
	channelID := c.Param("channelID")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required", "code": "INVALID_REQUEST"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required", "code": "AUTH_REQUIRED"})
		return
	}

	err := h.channelService.Delete(c.Request.Context(), userID, channelID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

func toChannelResponse(channel *domain.Channel) ChannelResponse {
	return ChannelResponse{
		ID:          channel.ID,
		TenantID:    channel.TenantID,
		SectionID:   channel.SectionID,
		NetworkID:   channel.NetworkID,
		Name:        channel.Name,
		Description: channel.Description,
		Type:        channel.Type,
		Position:    channel.Position,
		Bitrate:     channel.Bitrate,
		UserLimit:   channel.UserLimit,
		Slowmode:    channel.Slowmode,
		NSFW:        channel.NSFW,
		CreatedAt:   channel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   channel.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func getIntValue(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
