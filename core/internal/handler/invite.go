package handler

import (
"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// InviteHandler handles invite token HTTP endpoints
type InviteHandler struct {
	inviteService *service.InviteService
}

// NewInviteHandler creates a new invite handler
func NewInviteHandler(inviteService *service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
	}
}

// CreateInvite handles POST /v1/networks/:id/invites
// @Summary Create network invite token
// @Description Create a new invite token for a network (requires admin/owner role)
// @Tags Invites
// @Security bearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Network ID"
// @Param body body domain.CreateInviteRequest true "Invite options"
// @Success 201 {object} domain.InviteTokenResponse
// @Failure 400 {object} domain.Error
// @Failure 401 {object} domain.Error
// @Failure 403 {object} domain.Error
// @Router /v1/networks/{id}/invites [post]
func (h *InviteHandler) CreateInvite(c *gin.Context) {
	networkID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}
	tenantID, _ := c.Get("tenant_id")

	var req domain.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body - use defaults
		req = domain.CreateInviteRequest{}
	}

	opts := service.CreateInviteOptions{
		ExpiresIn: req.ExpiresIn,
		UsesMax:   req.UsesMax,
	}

	response, err := h.inviteService.CreateInvite(c.Request.Context(), networkID, tenantID.(string), userID.(string), opts)
	if err != nil {
		var domErr *domain.Error; if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListInvites handles GET /v1/networks/:id/invites
// @Summary List network invite tokens
// @Description List all active invite tokens for a network (requires admin/owner role)
// @Tags Invites
// @Security bearerAuth
// @Produce json
// @Param id path string true "Network ID"
// @Success 200 {array} domain.InviteTokenResponse
// @Failure 401 {object} domain.Error
// @Failure 403 {object} domain.Error
// @Router /v1/networks/{id}/invites [get]
func (h *InviteHandler) ListInvites(c *gin.Context) {
	networkID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}

	invites, err := h.inviteService.ListInvites(c.Request.Context(), networkID, userID.(string))
	if err != nil {
		var domErr *domain.Error; if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invites})
}

// GetInvite handles GET /v1/networks/:id/invites/:invite_id
// @Summary Get invite token details
// @Description Get details of a specific invite token
// @Tags Invites
// @Security bearerAuth
// @Produce json
// @Param id path string true "Network ID"
// @Param invite_id path string true "Invite Token ID"
// @Success 200 {object} domain.InviteTokenResponse
// @Failure 401 {object} domain.Error
// @Failure 404 {object} domain.Error
// @Router /v1/networks/{id}/invites/{invite_id} [get]
func (h *InviteHandler) GetInvite(c *gin.Context) {
	inviteID := c.Param("invite_id")

	response, err := h.inviteService.GetInviteByID(c.Request.Context(), inviteID)
	if err != nil {
		var domErr *domain.Error; if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RevokeInvite handles DELETE /v1/networks/:id/invites/:invite_id
// @Summary Revoke invite token
// @Description Revoke an invite token (requires admin/owner role)
// @Tags Invites
// @Security bearerAuth
// @Param id path string true "Network ID"
// @Param invite_id path string true "Invite Token ID"
// @Success 204 "No Content"
// @Failure 401 {object} domain.Error
// @Failure 403 {object} domain.Error
// @Failure 404 {object} domain.Error
// @Router /v1/networks/{id}/invites/{invite_id} [delete]
func (h *InviteHandler) RevokeInvite(c *gin.Context) {
	networkID := c.Param("id")
	inviteID := c.Param("invite_id")
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}
	tenantID, _ := c.Get("tenant_id")

	err := h.inviteService.RevokeInvite(c.Request.Context(), inviteID, networkID, tenantID.(string), userID.(string))
	if err != nil {
		var domErr *domain.Error; if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ValidateInvite handles GET /v1/invites/:token/validate
// @Summary Validate invite token
// @Description Validate an invite token without using it (public endpoint)
// @Tags Invites
// @Produce json
// @Param token path string true "Invite Token"
// @Success 200 {object} object{valid:bool,network_id:string,network_name:string}
// @Failure 400 {object} domain.Error
// @Router /v1/invites/{token}/validate [get]
func (h *InviteHandler) ValidateInvite(c *gin.Context) {
	tokenStr := c.Param("token")

	token, err := h.inviteService.ValidateInvite(c.Request.Context(), tokenStr)
	if err != nil {
		// For invalid tokens, return a generic response
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"network_id": token.NetworkID,
		"expires_at": token.ExpiresAt,
	})
}

// RegisterInviteRoutes registers invite routes
func RegisterInviteRoutes(r *gin.Engine, h *InviteHandler, authMiddleware gin.HandlerFunc) {
	// Public validation endpoint
	r.GET("/v1/invites/:token/validate", h.ValidateInvite)

	// Protected invite management endpoints
	invites := r.Group("/v1/networks/:id/invites")
	invites.Use(authMiddleware)
	{
		invites.POST("", h.CreateInvite)
		invites.GET("", h.ListInvites)
		invites.GET("/:invite_id", h.GetInvite)
		invites.DELETE("/:invite_id", h.RevokeInvite)
	}
}
