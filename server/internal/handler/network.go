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
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(networkService *service.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
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