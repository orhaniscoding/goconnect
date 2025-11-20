package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// DeviceHandler handles device HTTP requests
type DeviceHandler struct {
	deviceService *service.DeviceService
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// RegisterDevice handles POST /v1/devices
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}

	var req domain.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid request body",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	device, err := h.deviceService.RegisterDevice(c.Request.Context(), userID, tenantID, &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to register device",
		})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// ListDevices handles GET /v1/devices
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	// Parse query parameters
	platform := c.Query("platform")

	// Parse active filter
	var activePtr *bool
	if activeStr := c.Query("active"); activeStr != "" {
		active := activeStr == "true"
		activePtr = &active
	}

	// Parse limit
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse cursor
	cursor := c.Query("cursor")

	// Create filter
	filter := domain.DeviceFilter{
		Platform: platform,
		Active:   activePtr,
		Limit:    limit,
		Cursor:   cursor,
	}

	devices, nextCursor, err := h.deviceService.ListDevices(c.Request.Context(), userID, tenantID, isAdmin, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to list devices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices":     devices,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
	})
}

// GetDevice handles GET /v1/devices/:id
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	device, err := h.deviceService.GetDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to get device",
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// UpdateDevice handles PATCH /v1/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	var req domain.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid request body",
		})
		return
	}

	device, err := h.deviceService.UpdateDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin, &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to update device",
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice handles DELETE /v1/devices/:id
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	if err := h.deviceService.DeleteDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to delete device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "deleted",
		"device_id": deviceID,
	})
}

// Heartbeat handles POST /v1/devices/:id/heartbeat
func (h *DeviceHandler) Heartbeat(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	var req domain.DeviceHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if err := h.deviceService.Heartbeat(c.Request.Context(), deviceID, userID, tenantID, &req); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to process heartbeat",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// DisableDevice handles POST /v1/devices/:id/disable
func (h *DeviceHandler) DisableDevice(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	if err := h.deviceService.DisableDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to disable device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "disabled",
		"device_id": deviceID,
	})
}

// GetDeviceConfig handles GET /v1/devices/:id/config
func (h *DeviceHandler) GetDeviceConfig(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	config, err := h.deviceService.GetDeviceConfig(c.Request.Context(), deviceID, userID, tenantID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to get device config",
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// EnableDevice handles POST /v1/devices/:id/enable
func (h *DeviceHandler) EnableDevice(c *gin.Context) {
	deviceID := c.Param("id")
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	isAdmin := c.GetBool("is_admin")

	if err := h.deviceService.EnableDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to enable device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "enabled",
		"device_id": deviceID,
	})
}

// RegisterDeviceRoutes registers device routes
func RegisterDeviceRoutes(r *gin.Engine, handler *DeviceHandler, authMiddleware gin.HandlerFunc) {
	devices := r.Group("/v1/devices")
	devices.Use(authMiddleware) // All device routes require authentication

	devices.POST("", handler.RegisterDevice)     // Register new device
	devices.GET("", handler.ListDevices)         // List user's devices
	devices.GET("/:id", handler.GetDevice)       // Get specific device
	devices.PATCH("/:id", handler.UpdateDevice)  // Update device info
	devices.DELETE("/:id", handler.DeleteDevice) // Delete device

	devices.POST("/:id/heartbeat", handler.Heartbeat)   // Device heartbeat
	devices.POST("/:id/disable", handler.DisableDevice) // Disable device
	devices.POST("/:id/enable", handler.EnableDevice)   // Enable device
	devices.GET("/:id/config", handler.GetDeviceConfig) // Get device config
}
