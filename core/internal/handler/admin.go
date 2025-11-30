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
	query := c.DefaultQuery("q", "")

	users, total, err := h.adminService.ListUsers(c.Request.Context(), limit, offset, query)
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
	query := c.Query("q")

	tenants, total, err := h.adminService.ListTenants(c.Request.Context(), limit, offset, query)
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
	query := c.Query("q")

	networks, nextCursor, err := h.adminService.ListNetworks(c.Request.Context(), limit, cursor, query)
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
	query := c.Query("q")

	devices, nextCursor, err := h.adminService.ListDevices(c.Request.Context(), limit, cursor, query)
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

// ====== NEW ADMIN USER MANAGEMENT HANDLERS ======

// ListAllUsers retrieves all users with filters and pagination
func (h *AdminHandler) ListAllUsers(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	// Parse filters
	filters := domain.UserFilters{
		Role:     c.Query("role"),      // "admin", "moderator", or empty
		Status:   c.Query("status"),    // "active", "suspended", or empty
		TenantID: c.Query("tenant_id"), // filter by tenant
		Search:   c.Query("q"),         // search in email/username
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	pagination := domain.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	users, totalCount, err := h.adminService.ListAllUsers(c.Request.Context(), adminUserID, filters, pagination)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"total":    totalCount,
			"page":     page,
			"per_page": perPage,
			"pages":    (totalCount + perPage - 1) / perPage,
		},
	})
}

// GetUserStats retrieves system-wide statistics
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	stats, err := h.adminService.GetUserStats(c.Request.Context(), adminUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// UpdateUserRole updates a user's role (admin/moderator flags)
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "User ID is required", nil))
		return
	}

	var req domain.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request body", nil))
		return
	}

	err := h.adminService.UpdateUserRole(c.Request.Context(), adminUserID, targetUserID, req.IsAdmin, req.IsModerator)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
	})
}

// SuspendUser suspends a user account
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "User ID is required", nil))
		return
	}

	var req domain.SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request body", nil))
		return
	}

	// Validate reason
	if err := req.Validate(); err != nil {
		if derr, ok := err.(*domain.Error); ok {
			errorResponse(c, derr)
		} else {
			errorResponse(c, &domain.Error{
				Code:    domain.ErrInvalidRequest,
				Message: err.Error(),
			})
		}
		return
	}

	err := h.adminService.SuspendUser(c.Request.Context(), adminUserID, targetUserID, req.Reason)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User suspended successfully",
	})
}

// UnsuspendUser unsuspends a user account
func (h *AdminHandler) UnsuspendUser(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "User ID is required", nil))
		return
	}

	err := h.adminService.UnsuspendUser(c.Request.Context(), adminUserID, targetUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User unsuspended successfully",
	})
}

// GetUserDetails retrieves full details of a user
func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	adminUserID := c.GetString("user_id")
	if adminUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "User ID is required", nil))
		return
	}

	user, err := h.adminService.GetUserDetails(c.Request.Context(), adminUserID, targetUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}
