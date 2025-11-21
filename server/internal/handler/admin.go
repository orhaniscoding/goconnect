package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) handleError(c *gin.Context, err error) {
	if derr, ok := err.(*domain.Error); ok {
		errorResponse(c, derr)
	} else {
		errorResponse(c, &domain.Error{
			Code:    domain.ErrInternalServer,
			Message: err.Error(),
		})
	}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.adminService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *AdminHandler) ListTenants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tenants, total, err := h.adminService.ListTenants(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tenants,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.adminService.GetSystemStats(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

func (h *AdminHandler) ToggleUserAdmin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		errorResponse(c, &domain.Error{
			Code:    domain.ErrInvalidRequest,
			Message: "User ID is required",
		})
		return
	}

	user, err := h.adminService.ToggleUserAdmin(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		errorResponse(c, &domain.Error{
			Code:    domain.ErrInvalidRequest,
			Message: "User ID is required",
		})
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

func (h *AdminHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		errorResponse(c, &domain.Error{
			Code:    domain.ErrInvalidRequest,
			Message: "Tenant ID is required",
		})
		return
	}

	if err := h.adminService.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant deleted successfully",
	})
}

func (h *AdminHandler) ListNetworks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	cursor := c.DefaultQuery("cursor", "")

	networks, nextCursor, err := h.adminService.ListNetworks(c.Request.Context(), limit, cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": networks,
		"meta": gin.H{
			"limit":       limit,
			"next_cursor": nextCursor,
			"has_more":    nextCursor != "",
		},
	})
}

func (h *AdminHandler) ListDevices(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	cursor := c.DefaultQuery("cursor", "")

	devices, nextCursor, err := h.adminService.ListDevices(c.Request.Context(), limit, cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": devices,
		"meta": gin.H{
			"limit":       limit,
			"next_cursor": nextCursor,
			"has_more":    nextCursor != "",
		},
	})
}
