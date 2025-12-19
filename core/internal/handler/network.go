package handler

import (
	"errors"
	"fmt" // Keep fmt for fmt.Sprintf in Content-Disposition
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
)

// NetworkHandler handles HTTP requests for network operations
type NetworkHandler struct {
	networkService *service.NetworkService
	memberService  *service.MembershipService
	ipamService    *service.IPAMService
	deviceService  *service.DeviceService
	peerRepo       repository.PeerRepository
	wgConfig       config.WireGuardConfig
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(
	networkService *service.NetworkService,
	memberService *service.MembershipService,
	deviceService *service.DeviceService,
	peerRepo repository.PeerRepository,
	wgConfig config.WireGuardConfig,
) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
		memberService:  memberService,
		deviceService:  deviceService,
		peerRepo:       peerRepo,
		wgConfig:       wgConfig,
	}
}

// WithIPAM allows late injection of IPAM service to avoid breaking existing tests
func (h *NetworkHandler) WithIPAM(ipam *service.IPAMService) *NetworkHandler {
	h.ipamService = ipam
	return h
}

// CreateNetwork handles POST /v1/networks
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var req domain.CreateNetworkRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("CreateNetwork: Invalid request body", "error", err)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	// Apply defaults for optional fields (CLI compat: sends only name)
	req.ApplyDefaults()

	// Extract user info from context
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr := tenantID.(string)

	// Extract idempotency key (required for mutations)
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		slog.Warn("CreateNetwork: Idempotency-Key header is required")
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Idempotency-Key header is required for mutation operations",
			map[string]string{"required_header": "Idempotency-Key"}))
		return
	}

	// Call service
	network, err := h.networkService.CreateNetwork(c.Request.Context(), &req, userIDStr, tenantIDStr, idempotencyKey)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			slog.Error("CreateNetwork: Service error", "error", domainErr, "user_id", userIDStr, "tenant_id", tenantIDStr)
			errorResponse(c, domainErr)
		} else {
			slog.Error("CreateNetwork: Internal server error", "error", err, "user_id", userIDStr, "tenant_id", tenantIDStr)
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	slog.Info("Network created successfully", "network_id", network.ID, "user_id", userIDStr, "tenant_id", tenantIDStr)
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
		slog.Warn("ListNetworks: Invalid query parameters", "error", err)
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
	tenantID, _ := c.Get("tenant_id")
	isAdmin, _ := c.Get("is_admin")

	userIDStr := userID.(string)
	tenantIDStr := tenantID.(string)
	isAdminBool := isAdmin.(bool)

	// Call service
	networks, nextCursor, err := h.networkService.ListNetworks(c.Request.Context(), &req, userIDStr, tenantIDStr, isAdminBool)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			slog.Error("ListNetworks: Service error", "error", domainErr, "user_id", userIDStr, "tenant_id", tenantIDStr)
			errorResponse(c, domainErr)
		} else {
			slog.Error("ListNetworks: Internal server error", "error", err, "user_id", userIDStr, "tenant_id", tenantIDStr)
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	slog.Info("Networks listed successfully", "count", len(networks), "user_id", userIDStr, "tenant_id", tenantIDStr)
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
	id := c.Param("id")
	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	net, err := h.networkService.GetNetwork(c.Request.Context(), id, userID, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("GetNetwork: Service error", "error", derr, "network_id", id, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("GetNetwork: Internal server error", "error", err, "network_id", id, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Network retrieved successfully", "network_id", id, "user_id", userID, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"data": net})
}

// UpdateNetwork handles PATCH /v1/networks/:id
func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("UpdateNetwork: Idempotency-Key header is required")
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	id := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		slog.Warn("UpdateNetwork: Invalid body", "error", err, "network_id", id, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", map[string]string{"details": err.Error()}))
		return
	}
	updated, err := h.networkService.UpdateNetwork(c.Request.Context(), id, actor, tenantID, patch)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("UpdateNetwork: Service error", "error", derr, "network_id", id, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("UpdateNetwork: Internal server error", "error", err, "network_id", id, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Network updated successfully", "network_id", id, "actor_id", actor, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// DeleteNetwork handles DELETE /v1/networks/:id
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("DeleteNetwork: Idempotency-Key header is required")
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	id := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	if err := h.networkService.DeleteNetwork(c.Request.Context(), id, actor, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("DeleteNetwork: Service error", "error", derr, "network_id", id, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("DeleteNetwork: Internal server error", "error", err, "network_id", id, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Network deleted successfully", "network_id", id, "actor_id", actor, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RegisterNetworkRoutes registers all network-related routes
func RegisterNetworkRoutes(r *gin.Engine, handler *NetworkHandler, authService TokenValidator, mrepo repository.MembershipRepository) {
	v1 := r.Group("/v1")
	v1.Use(RequestIDMiddleware())
	// Note: CORS is applied globally via router.Use(handler.NewCORSMiddleware(&cfg.CORS)) in main.go
	// Basic rate limiting per user/IP to protect write endpoints; configurable via env
	// Defaults: capacity=5 tokens per 1s window
	rl := NewRateLimiterFromEnv(5, time.Second)

	// Network routes
	networks := v1.Group("/networks")
	networks.Use(AuthMiddleware(authService))        // All network operations require authentication
	networks.Use(RoleMiddleware(mrepo, authService)) // Resolve membership role after auth

	networks.POST("", rl, handler.CreateNetwork)
	networks.GET("", handler.ListNetworks)
	networks.GET("/:id", handler.GetNetwork)
	networks.PATCH("/:id", rl, RequireNetworkAdmin(), handler.UpdateNetwork)
	networks.DELETE("/:id", rl, RequireNetworkAdmin(), handler.DeleteNetwork)

	// IP allocation (member-level; no admin required). Mutation -> requires Idempotency-Key
	networks.POST("/:id/ip-allocations", rl, handler.AllocateIP)
	networks.GET("/:id/ip-allocations", handler.ListIPAllocations)
	networks.DELETE("/:id/ip-allocation", rl, handler.ReleaseIP)
	// Admin/Owner release of another member's allocation
	networks.DELETE("/:id/ip-allocations/:user_id", rl, RequireNetworkAdmin(), handler.AdminReleaseIP)

	// Membership & Join flow
	networks.POST("/:id/join", rl, handler.JoinNetwork)
	networks.POST("/join", rl, handler.JoinNetworkByInvite) // Compat route for CLI (invite_code in body)
	networks.POST("/:id/approve", rl, RequireNetworkAdmin(), handler.Approve)
	networks.POST("/:id/deny", rl, RequireNetworkAdmin(), handler.Deny)
	networks.POST("/:id/kick", rl, RequireNetworkAdmin(), handler.Kick)
	networks.POST("/:id/config", rl, handler.GenerateConfig)

	networks.POST("/:id/ban", rl, RequireNetworkAdmin(), handler.Ban)
	networks.GET("/:id/members", handler.ListMembers)
	networks.GET("/:id/join-requests", RequireNetworkAdmin(), handler.ListJoinRequests)
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
	tenantID := c.MustGet("tenant_id").(string)
	idem := c.GetHeader("Idempotency-Key")
	if idem == "" {
		slog.Warn("JoinNetwork: Idempotency-Key header is required", "network_id", networkID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	m, jr, err := h.memberService.JoinNetwork(c.Request.Context(), networkID, userID, tenantID, idem)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("JoinNetwork: Service error", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("JoinNetwork: Internal server error", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	if m != nil {
		slog.Info("User joined network successfully", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		c.JSON(http.StatusOK, gin.H{"data": m})
		return
	}
	slog.Info("Join request submitted successfully", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
	c.JSON(http.StatusAccepted, gin.H{"data": jr})
}

// Approve handles POST /v1/networks/:id/approve
func (h *NetworkHandler) Approve(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("Approve: Idempotency-Key header is required", "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.Warn("Approve: Invalid body", "error", err, "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	m, err := h.memberService.Approve(c.Request.Context(), networkID, body.UserID, actor, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("Approve: Service error", "error", derr, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("Approve: Internal server error", "error", err, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Member approved successfully", "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *NetworkHandler) Deny(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("Deny: Idempotency-Key header is required", "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.Warn("Deny: Invalid body", "error", err, "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Deny(c.Request.Context(), networkID, body.UserID, actor, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("Deny: Service error", "error", derr, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("Deny: Internal server error", "error", err, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Join request denied successfully", "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NetworkHandler) Kick(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("Kick: Idempotency-Key header is required", "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.Warn("Kick: Invalid body", "error", err, "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Kick(c.Request.Context(), networkID, body.UserID, actor, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("Kick: Service error", "error", derr, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("Kick: Internal server error", "error", err, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Member kicked successfully", "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NetworkHandler) Ban(c *gin.Context) {
	networkID := c.Param("id")
	actor := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("Ban: Idempotency-Key header is required", "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.Warn("Ban: Invalid body", "error", err, "network_id", networkID, "actor_id", actor)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid body", nil))
		return
	}
	if err := h.memberService.Ban(c.Request.Context(), networkID, body.UserID, actor, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("Ban: Service error", "error", derr, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("Ban: Internal server error", "error", err, "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Member banned successfully", "network_id", networkID, "target_user_id", body.UserID, "actor_id", actor, "tenant_id", tenantID)
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
	tenantID := c.MustGet("tenant_id").(string)
	data, next, err := h.memberService.ListMembers(c.Request.Context(), networkID, status, tenantID, limit, cursor)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("ListMembers: Service error", "error", derr, "network_id", networkID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("ListMembers: Internal server error", "error", err, "network_id", networkID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("Members listed successfully", "network_id", networkID, "status", status, "count", len(data), "tenant_id", tenantID)
	resp := gin.H{"data": data, "pagination": gin.H{"limit": limit}}
	if next != "" {
		resp["pagination"].(gin.H)["next_cursor"] = next
	}
	c.JSON(http.StatusOK, resp)
}

// ListJoinRequests handles GET /v1/networks/:id/join-requests (admin/owner only)
func (h *NetworkHandler) ListJoinRequests(c *gin.Context) {
	networkID := c.Param("id")
	tenantID := c.MustGet("tenant_id").(string)

	requests, err := h.memberService.ListJoinRequests(c.Request.Context(), networkID, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("ListJoinRequests: Service error", "error", derr, "network_id", networkID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("ListJoinRequests: Internal server error", "error", err, "network_id", networkID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	slog.Info("Join requests listed successfully", "network_id", networkID, "count", len(requests), "tenant_id", tenantID)
	c.JSON(http.StatusOK, gin.H{"data": requests})
}

// AllocateIP handles POST /v1/networks/:id/ip-allocations
func (h *NetworkHandler) AllocateIP(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)

	if h.ipamService == nil {
		slog.Error("AllocateIP: IPAM service not available", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "IPAM not available", nil))
		return
	}
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("AllocateIP: Idempotency-Key header is required", "network_id", networkID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}

	if h.memberService == nil { // safety
		slog.Error("AllocateIP: Membership service unavailable", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Membership service unavailable", nil))
		return
	}
	// We need membership repository access; quick path: call ListMembers with limit 1 filtering not implemented -> fallback to internal repo not exposed.
	// Simplify: attempt allocation; if network Get ok and not admin maybe enforce membership by future improvement.
	// For now we enforce by checking membership role presence using service private method isn't accessible -> compromise: treat any user as allowed (will adjust later when repository accessible here).
	alloc, err := h.ipamService.AllocateIP(c.Request.Context(), networkID, userID, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("AllocateIP: Service error", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("AllocateIP: Internal server error", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("IP allocated successfully", "network_id", networkID, "user_id", userID, "tenant_id", tenantID, "allocated_ip", alloc.IP)
	c.JSON(http.StatusOK, gin.H{"data": alloc})
}

// ListIPAllocations handles GET /v1/networks/:id/ip-allocations
func (h *NetworkHandler) ListIPAllocations(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)

	if h.ipamService == nil {
		slog.Error("ListIPAllocations: IPAM service not available", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "IPAM not available", nil))
		return
	}
	allocs, err := h.ipamService.ListAllocations(c.Request.Context(), networkID, userID, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("ListIPAllocations: Service error", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("ListIPAllocations: Internal server error", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("IP allocations listed successfully", "network_id", networkID, "user_id", userID, "tenant_id", tenantID, "count", len(allocs))
	c.JSON(http.StatusOK, gin.H{"data": allocs})
}

// ReleaseIP handles DELETE /v1/networks/:id/ip-allocation (self-release)
func (h *NetworkHandler) ReleaseIP(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)

	if h.ipamService == nil {
		slog.Error("ReleaseIP: IPAM service not available", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "IPAM not available", nil))
		return
	}
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("ReleaseIP: Idempotency-Key header is required", "network_id", networkID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	if err := h.ipamService.ReleaseIP(c.Request.Context(), networkID, userID, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("ReleaseIP: Service error", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("ReleaseIP: Internal server error", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("IP released successfully", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
	c.Status(http.StatusNoContent)
}

// AdminReleaseIP handles DELETE /v1/networks/:id/ip-allocations/:user_id (admin/owner releasing another member's allocation)
func (h *NetworkHandler) AdminReleaseIP(c *gin.Context) {
	networkID := c.Param("id")
	targetUserID := c.Param("user_id")
	actorUserID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)

	if h.ipamService == nil {
		slog.Error("AdminReleaseIP: IPAM service not available", "network_id", networkID, "target_user_id", targetUserID, "actor_id", actorUserID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "IPAM not available", nil))
		return
	}
	if c.GetHeader("Idempotency-Key") == "" {
		slog.Warn("AdminReleaseIP: Idempotency-Key header is required", "network_id", networkID, "target_user_id", targetUserID, "actor_id", actorUserID)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Idempotency-Key header is required for mutation operations", map[string]string{"required_header": "Idempotency-Key"}))
		return
	}
	if err := h.ipamService.ReleaseIPForActor(c.Request.Context(), networkID, actorUserID, targetUserID, tenantID); err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("AdminReleaseIP: Service error", "error", derr, "network_id", networkID, "target_user_id", targetUserID, "actor_id", actorUserID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("AdminReleaseIP: Internal server error", "error", err, "network_id", networkID, "target_user_id", targetUserID, "actor_id", actorUserID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("IP released by admin successfully", "network_id", networkID, "target_user_id", targetUserID, "actor_id", actorUserID, "tenant_id", tenantID)
	c.Status(http.StatusNoContent)
}

// GenerateConfig handles POST /v1/networks/:id/config
func (h *NetworkHandler) GenerateConfig(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	userEmail := c.GetString("user_email")

	slog.Info("GenerateConfig request received", "network_id", networkID, "user_id", userID, "tenant_id", tenantID)

	// 1. Check if user is a member of the network
	network, err := h.networkService.GetNetwork(c.Request.Context(), networkID, userID, tenantID)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("GenerateConfig: Failed to get network for membership check", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("GenerateConfig: Internal server error during network check", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	// 2. Generate Key Pair
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		slog.Error("GenerateConfig: Failed to generate WireGuard key pair", "error", err, "network_id", networkID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to generate keys", nil))
		return
	}
	slog.Debug("GenerateConfig: WireGuard key pair generated", "network_id", networkID, "user_id", userID, "public_key", keyPair.PublicKey)

	// 3. Register a new "Manual" Device
	deviceName := fmt.Sprintf("Manual Config %s", time.Now().Format("2006-01-02 15:04"))
	regReq := &domain.RegisterDeviceRequest{
		Name:     deviceName,
		Platform: "linux", // Generic
		PubKey:   keyPair.PublicKey,
	}

	device, err := h.deviceService.RegisterDevice(c.Request.Context(), userID, tenantID, regReq)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("GenerateConfig: Failed to register device", "error", derr, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
			errorResponse(c, derr)
			return
		}
		slog.Error("GenerateConfig: Internal server error during device registration", "error", err, "network_id", networkID, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to register device", nil))
		return
	}
	slog.Info("GenerateConfig: Device registered successfully", "device_id", device.ID, "device_name", device.Name, "network_id", networkID, "user_id", userID)

	// 4. Get the Peer (IP Allocation) for this network
	peer, err := h.peerRepo.GetByNetworkAndDevice(c.Request.Context(), networkID, device.ID)
	if err != nil {
		slog.Error("GenerateConfig: Failed to retrieve peer allocation", "error", err, "network_id", networkID, "device_id", device.ID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to retrieve peer allocation", nil))
		return
	}
	slog.Debug("GenerateConfig: Peer allocation retrieved", "network_id", networkID, "device_id", device.ID, "peer_id", peer.ID)

	// 5. Generate Config
	gen := wireguard.NewProfileGenerator(
		h.wgConfig.ServerEndpoint,
		h.wgConfig.ServerPubKey,
		h.wgConfig.DNS,
		h.wgConfig.MTU,
		h.wgConfig.Keepalive,
	)

	// Parse CIDR to get prefix length
	_, ipNet, _ := net.ParseCIDR(network.CIDR)
	prefixLen, _ := ipNet.Mask.Size()

	// Extract IP from AllowedIPs
	if len(peer.AllowedIPs) == 0 {
		slog.Error("GenerateConfig: No IP allocated for peer", "network_id", networkID, "device_id", device.ID, "peer_id", peer.ID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "No IP allocated", nil))
		return
	}
	deviceIP := peer.AllowedIPs[0]
	if idx := strings.Index(deviceIP, "/"); idx != -1 {
		deviceIP = deviceIP[:idx]
	}

	profReq := &wireguard.ProfileRequest{
		NetworkID:        network.ID,
		NetworkName:      network.Name,
		NetworkCIDR:      network.CIDR,
		DeviceID:         device.ID,
		DeviceName:       device.Name,
		DeviceIP:         deviceIP,
		DevicePrivateKey: keyPair.PrivateKey,
		PrefixLen:        prefixLen,
		UserID:           userID,
		UserEmail:        userEmail,
	}

	configContent, err := gen.GenerateClientConfig(c.Request.Context(), profReq)
	if err != nil {
		slog.Error("GenerateConfig: Failed to generate client config", "error", err, "network_id", networkID, "device_id", device.ID, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to generate config", nil))
		return
	}

	slog.Info("GenerateConfig: WireGuard config generated successfully", "network_id", networkID, "device_id", device.ID, "user_id", userID)
	// 6. Return Config
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.conf\"", network.Name))
	c.Data(http.StatusOK, "application/x-wireguard-profile", []byte(configContent))
}

// JoinNetworkByInvite handles POST /v1/networks/join (compat route for CLI)
// This endpoint accepts an invite_code in the body and resolves the network internally
func (h *NetworkHandler) JoinNetworkByInvite(c *gin.Context) {
	var req struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("JoinNetworkByInvite: Invalid request body", "error", err)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"invite_code is required",
			map[string]string{"details": err.Error()}))
		return
	}

	userID := c.MustGet("user_id").(string)
	tenantID := c.MustGet("tenant_id").(string)
	idem := c.GetHeader("Idempotency-Key")
	if idem == "" {
		slog.Warn("JoinNetworkByInvite: Idempotency-Key header is required", "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Idempotency-Key header is required for mutation operations",
			map[string]string{"required_header": "Idempotency-Key"}))
		return
	}

	// Join using invite code - the service will resolve the network ID
	m, jr, err := h.memberService.JoinByInviteCode(c.Request.Context(), req.InviteCode, userID, tenantID, idem)
	if err != nil {
		var derr *domain.Error
		if errors.As(err, &derr) {
			slog.Error("JoinNetworkByInvite: Service error", "error", derr, "invite_code", req.InviteCode, "user_id", userID)
			errorResponse(c, derr)
			return
		}
		slog.Error("JoinNetworkByInvite: Internal server error", "error", err, "invite_code", req.InviteCode, "user_id", userID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	if m != nil {
		slog.Info("User joined network via invite code successfully", "invite_code", req.InviteCode, "user_id", userID)
		c.JSON(http.StatusOK, gin.H{"data": m})
		return
	}
	slog.Info("Join request via invite code submitted successfully", "invite_code", req.InviteCode, "user_id", userID)
	c.JSON(http.StatusAccepted, gin.H{"data": jr})
}
