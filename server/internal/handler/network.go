package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// NetworkHandler handles HTTP requests for network operations
type NetworkHandler struct {
	networkService *service.NetworkService
	memberService  *service.MembershipService
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(networkService *service.NetworkService, memberService *service.MembershipService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
		memberService:  memberService,
	}
}

// CreateNetwork handles POST /v1/networks
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var req domain.CreateNetworkRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	// Extract user info from context
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	// Extract idempotency key (required for mutations)
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Idempotency-Key header is required for mutation operations",
			map[string]string{"required_header": "Idempotency-Key"}))
		return
	}

	// Call service
	network, err := h.networkService.CreateNetwork(c.Request.Context(), &req, userIDStr, idempotencyKey)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	// Return created network
	c.JSON(http.StatusCreated, gin.H{
		"data": network,
	})
}

// ListNetworks handles GET /v1/networks
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	var req domain.ListNetworksRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Visibility == "" {
		req.Visibility = "public"
	}

	// Extract user info from context
	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	userIDStr := userID.(string)
	isAdminBool := isAdmin.(bool)

	// Call service
	networks, nextCursor, err := h.networkService.ListNetworks(c.Request.Context(), &req, userIDStr, isAdminBool)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	// Build response
	response := gin.H{
		"data": networks,
		"pagination": gin.H{
			"limit": req.Limit,
		},
	}

	// Add next cursor if available
	if nextCursor != "" {
		response["pagination"].(gin.H)["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// GetNetwork handles GET /v1/networks/:id
func (h *NetworkHandler) GetNetwork(c *gin.Context) {
	// TODO: Implement get network by ID
	errorResponse(c, domain.NewError(domain.ErrNotImplemented, "Get network by ID not implemented yet", nil))
}

// UpdateNetwork handles PATCH /v1/networks/:id
func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
	// TODO: Implement network updates
	errorResponse(c, domain.NewError(domain.ErrNotImplemented, "Update network not implemented yet", nil))
}

// DeleteNetwork handles DELETE /v1/networks/:id
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	// TODO: Implement network deletion (soft delete)
	errorResponse(c, domain.NewError(domain.ErrNotImplemented, "Delete network not implemented yet", nil))
}

// RegisterNetworkRoutes registers all network-related routes
func RegisterNetworkRoutes(r *gin.Engine, handler *NetworkHandler) {
	v1 := r.Group("/v1")
	v1.Use(RequestIDMiddleware())
	v1.Use(CORSMiddleware())

	// Network routes
	networks := v1.Group("/networks")
	networks.Use(AuthMiddleware()) // All network operations require authentication

	networks.POST("", handler.CreateNetwork)
	networks.GET("", handler.ListNetworks)
	networks.GET("/:id", handler.GetNetwork)
	networks.PATCH("/:id", handler.UpdateNetwork)
	networks.DELETE("/:id", handler.DeleteNetwork)

	// Membership & Join flow
	networks.POST("/:id/join", handler.JoinNetwork)
	networks.POST("/:id/approve", handler.Approve)
	networks.POST("/:id/deny", handler.Deny)
	networks.POST("/:id/kick", handler.Kick)
	networks.POST("/:id/ban", handler.Ban)
	networks.GET("/:id/members", handler.ListMembers)
}

// Helper function to convert string to int with default
func parseIntWithDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultVal
}

// JoinNetwork handles POST /v1/networks/:id/join
func (h *NetworkHandler) JoinNetwork(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	idem := c.GetHeader("Idempotency-Key")
	if idem == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	m, jr, err := h.memberService.JoinNetwork(c.Request.Context(), networkID, userID, idem)
	if err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	if m != nil {
		c.JSON(http.StatusOK, gin.H{"data": m})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"data": jr})
}

// Approve handles POST /v1/networks/:id/approve
func (h *NetworkHandler) Approve(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	m, err := h.memberService.Approve(c.Request.Context(), networkID, body.UserID, actor)
	if err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *NetworkHandler) Deny(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Deny(c.Request.Context(), networkID, body.UserID, actor); err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NetworkHandler) Kick(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Kick(c.Request.Context(), networkID, body.UserID, actor); err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NetworkHandler) Ban(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Ban(c.Request.Context(), networkID, body.UserID, actor); err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NetworkHandler) ListMembers(c *gin.Context) {
	networkID := c.Param("id")
	status := c.Query("status")
	limit := parseIntWithDefault(c.Query("limit"), 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cursor := c.Query("cursor")
	data, next, err := h.memberService.ListMembers(c.Request.Context(), networkID, status, limit, cursor)
	if err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	resp := gin.H{"data": data, "pagination": gin.H{"limit": limit}}
	if next != "" {
		resp["pagination"].(gin.H)["next_cursor"] = next
	}
	c.JSON(http.StatusOK, resp)
}
