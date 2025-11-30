package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// TenantHandler handles HTTP requests for tenant operations
type TenantHandler struct {
	tenantService *service.TenantMembershipService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *service.TenantMembershipService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// ==================== TENANT ROUTES ====================

// CreateTenant handles POST /v1/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req domain.CreateTenantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	userID := c.GetString("user_id")

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), userID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": tenant,
	})
}

// GetTenant handles GET /v1/tenants/:tenantId
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")

	tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tenant,
	})
}

// UpdateTenant handles PATCH /v1/tenants/:tenantId
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    tenant,
		"message": "Tenant updated successfully",
	})
}

// DeleteTenant handles DELETE /v1/tenants/:tenantId
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	err := h.tenantService.DeleteTenant(c.Request.Context(), userID, tenantID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant deleted successfully",
	})
}

// ==================== DISCOVERY ROUTES ====================

// ListPublicTenants handles GET /v1/tenants/public
func (h *TenantHandler) ListPublicTenants(c *gin.Context) {
	var req domain.ListTenantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(), nil))
		return
	}

	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 20
	}

	results, nextCursor, err := h.tenantService.ListPublicTenants(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := gin.H{
		"data": results,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// SearchTenants handles GET /v1/tenants/search
func (h *TenantHandler) SearchTenants(c *gin.Context) {
	var req domain.ListTenantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(), nil))
		return
	}

	query := strings.TrimSpace(c.DefaultQuery("q", req.Search))
	if query == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Missing search query", map[string]string{"field": "q"}))
		return
	}
	req.Search = query

	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 20
	}

	results, nextCursor, err := h.tenantService.ListPublicTenants(c.Request.Context(), &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := gin.H{
		"data": results,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// ==================== MEMBERSHIP ROUTES ====================

// JoinTenant handles POST /v1/tenants/:tenantId/join
func (h *TenantHandler) JoinTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.JoinTenantRequest
	// Password is optional, so ignore binding errors
	_ = c.ShouldBindJSON(&req)

	member, err := h.tenantService.JoinTenant(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    member,
		"message": "Successfully joined tenant",
	})
}

// JoinByCode handles POST /v1/tenants/join-by-code
func (h *TenantHandler) JoinByCode(c *gin.Context) {
	userID := c.GetString("user_id")

	var req domain.JoinByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	member, err := h.tenantService.JoinByCode(c.Request.Context(), userID, req.Code)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    member,
		"message": "Successfully joined tenant via invite code",
	})
}

// LeaveTenant handles POST /v1/tenants/:tenantId/leave
func (h *TenantHandler) LeaveTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	err := h.tenantService.LeaveTenant(c.Request.Context(), userID, tenantID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully left tenant",
	})
}

// GetUserTenants handles GET /v1/users/me/tenants
func (h *TenantHandler) GetUserTenants(c *gin.Context) {
	userID := c.GetString("user_id")

	memberships, err := h.tenantService.GetUserTenants(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": memberships,
	})
}

// GetTenantMembers handles GET /v1/tenants/:tenantId/members
func (h *TenantHandler) GetTenantMembers(c *gin.Context) {
	tenantID := c.Param("tenantId")

	var req domain.ListTenantMembersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(), nil))
		return
	}

	// Default limit
	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 50
	}

	members, nextCursor, err := h.tenantService.GetTenantMembers(c.Request.Context(), tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := gin.H{
		"data": members,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}
	c.JSON(http.StatusOK, response)
}

// UpdateMemberRole handles PATCH /v1/tenants/:tenantId/members/:memberId
func (h *TenantHandler) UpdateMemberRole(c *gin.Context) {
	tenantID := c.Param("tenantId")
	memberID := c.Param("memberId")
	actorID := c.GetString("user_id")

	var req domain.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	err := h.tenantService.UpdateMemberRole(c.Request.Context(), actorID, tenantID, memberID, req.Role)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member role updated successfully",
	})
}

// RemoveMember handles DELETE /v1/tenants/:tenantId/members/:memberId
func (h *TenantHandler) RemoveMember(c *gin.Context) {
	tenantID := c.Param("tenantId")
	memberID := c.Param("memberId")
	actorID := c.GetString("user_id")

	err := h.tenantService.RemoveMember(c.Request.Context(), actorID, tenantID, memberID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed successfully",
	})
}

// BanMember handles POST /v1/tenants/:tenantId/members/:memberId/ban
func (h *TenantHandler) BanMember(c *gin.Context) {
	tenantID := c.Param("tenantId")
	memberID := c.Param("memberId")
	actorID := c.GetString("user_id")

	err := h.tenantService.BanMember(c.Request.Context(), actorID, tenantID, memberID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member banned successfully",
	})
}

// UnbanMember handles DELETE /v1/tenants/:tenantId/members/:memberId/ban
func (h *TenantHandler) UnbanMember(c *gin.Context) {
	tenantID := c.Param("tenantId")
	memberID := c.Param("memberId")
	actorID := c.GetString("user_id")

	err := h.tenantService.UnbanMember(c.Request.Context(), actorID, tenantID, memberID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member unbanned successfully",
	})
}

// ListBannedMembers handles GET /v1/tenants/:tenantId/members/banned
func (h *TenantHandler) ListBannedMembers(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	members, err := h.tenantService.ListBannedMembers(c.Request.Context(), userID, tenantID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
	})
}

// ==================== INVITE ROUTES ====================

// CreateInvite handles POST /v1/tenants/:tenantId/invites
func (h *TenantHandler) CreateInvite(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.CreateTenantInviteRequest
	// All fields are optional, so ignore binding errors
	_ = c.ShouldBindJSON(&req)

	// Defaults
	if req.MaxUses == 0 {
		req.MaxUses = 1
	}
	if req.ExpiresIn == 0 {
		req.ExpiresIn = 86400 // 24 hours default
	}

	invite, err := h.tenantService.CreateInvite(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": invite,
	})
}

// ListInvites handles GET /v1/tenants/:tenantId/invites
func (h *TenantHandler) ListInvites(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	invites, err := h.tenantService.ListInvites(c.Request.Context(), userID, tenantID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invites,
	})
}

// RevokeInvite handles DELETE /v1/tenants/:tenantId/invites/:inviteId
func (h *TenantHandler) RevokeInvite(c *gin.Context) {
	tenantID := c.Param("tenantId")
	inviteID := c.Param("inviteId")
	userID := c.GetString("user_id")

	err := h.tenantService.RevokeInvite(c.Request.Context(), userID, tenantID, inviteID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invite revoked successfully",
	})
}

// ==================== ANNOUNCEMENT ROUTES ====================

// CreateAnnouncement handles POST /v1/tenants/:tenantId/announcements
func (h *TenantHandler) CreateAnnouncement(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	announcement, err := h.tenantService.CreateAnnouncement(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": announcement,
	})
}

// ListAnnouncements handles GET /v1/tenants/:tenantId/announcements
func (h *TenantHandler) ListAnnouncements(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.ListAnnouncementsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(), nil))
		return
	}

	if req.Limit == 0 || req.Limit > 50 {
		req.Limit = 20
	}

	announcements, nextCursor, err := h.tenantService.GetAnnouncements(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := gin.H{
		"data": announcements,
	}
	if nextCursor != "" {
		response["next_cursor"] = nextCursor
	}
	c.JSON(http.StatusOK, response)
}

// UpdateAnnouncement handles PATCH /v1/tenants/:tenantId/announcements/:announcementId
func (h *TenantHandler) UpdateAnnouncement(c *gin.Context) {
	tenantID := c.Param("tenantId")
	announcementID := c.Param("announcementId")
	userID := c.GetString("user_id")

	var req domain.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	err := h.tenantService.UpdateAnnouncement(c.Request.Context(), userID, tenantID, announcementID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement updated successfully",
	})
}

// DeleteAnnouncement handles DELETE /v1/tenants/:tenantId/announcements/:announcementId
func (h *TenantHandler) DeleteAnnouncement(c *gin.Context) {
	tenantID := c.Param("tenantId")
	announcementID := c.Param("announcementId")
	userID := c.GetString("user_id")

	err := h.tenantService.DeleteAnnouncement(c.Request.Context(), userID, tenantID, announcementID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Announcement deleted successfully",
	})
}

// ==================== CHAT ROUTES ====================

// SendChatMessage handles POST /v1/tenants/:tenantId/chat/messages
func (h *TenantHandler) SendChatMessage(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.SendChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(), nil))
		return
	}

	message, err := h.tenantService.SendChatMessage(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": message,
	})
}

// GetChatHistory handles GET /v1/tenants/:tenantId/chat/messages
func (h *TenantHandler) GetChatHistory(c *gin.Context) {
	tenantID := c.Param("tenantId")
	userID := c.GetString("user_id")

	var req domain.ListChatMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid query parameters: "+err.Error(), nil))
		return
	}

	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 50
	}

	messages, err := h.tenantService.GetChatHistory(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

// DeleteChatMessage handles DELETE /v1/tenants/:tenantId/chat/messages/:messageId
func (h *TenantHandler) DeleteChatMessage(c *gin.Context) {
	tenantID := c.Param("tenantId")
	messageID := c.Param("messageId")
	userID := c.GetString("user_id")

	err := h.tenantService.DeleteChatMessage(c.Request.Context(), userID, tenantID, messageID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
	})
}

// ==================== HELPER FUNCTIONS ====================

// handleServiceError handles service layer errors and sends appropriate HTTP responses
func handleServiceError(c *gin.Context, err error) {
	if domainErr, ok := err.(*domain.Error); ok {
		errorResponse(c, domainErr)
	} else {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
	}
}

// RegisterTenantRoutes registers all tenant-related routes
func (h *TenantHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Public tenant discovery routes (no authentication required)
	publicTenants := rg.Group("/tenants")
	{
		publicTenants.GET("/public", h.ListPublicTenants)
		publicTenants.GET("/search", h.SearchTenants)
	}

	// Tenant routes (authenticated)
	tenants := rg.Group("/tenants")
	tenants.Use(authMiddleware)
	{
		// Tenant CRUD
		tenants.POST("", h.CreateTenant)
		tenants.GET("/:tenantId", h.GetTenant)
		tenants.PATCH("/:tenantId", h.UpdateTenant)
		tenants.DELETE("/:tenantId", h.DeleteTenant)

		// Membership
		tenants.POST("/:tenantId/join", h.JoinTenant)
		tenants.POST("/join-by-code", h.JoinByCode)
		tenants.POST("/:tenantId/leave", h.LeaveTenant)

		// Members management
		tenants.GET("/:tenantId/members", h.GetTenantMembers)
		tenants.GET("/:tenantId/members/banned", h.ListBannedMembers)
		tenants.PATCH("/:tenantId/members/:memberId", h.UpdateMemberRole)
		tenants.DELETE("/:tenantId/members/:memberId", h.RemoveMember)
		tenants.POST("/:tenantId/members/:memberId/ban", h.BanMember)
		tenants.DELETE("/:tenantId/members/:memberId/ban", h.UnbanMember)

		// Invites
		tenants.POST("/:tenantId/invites", h.CreateInvite)
		tenants.GET("/:tenantId/invites", h.ListInvites)
		tenants.DELETE("/:tenantId/invites/:inviteId", h.RevokeInvite)

		// Announcements
		tenants.POST("/:tenantId/announcements", h.CreateAnnouncement)
		tenants.GET("/:tenantId/announcements", h.ListAnnouncements)
		tenants.PATCH("/:tenantId/announcements/:announcementId", h.UpdateAnnouncement)
		tenants.DELETE("/:tenantId/announcements/:announcementId", h.DeleteAnnouncement)

		// Chat
		tenants.POST("/:tenantId/chat/messages", h.SendChatMessage)
		tenants.GET("/:tenantId/chat/messages", h.GetChatHistory)
		tenants.DELETE("/:tenantId/chat/messages/:messageId", h.DeleteChatMessage)
	}

	// User's tenants route
	users := rg.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me/tenants", h.GetUserTenants)
	}
}
