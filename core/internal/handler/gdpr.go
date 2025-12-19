package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// GDPRHandler handles GDPR/DSR endpoints
type GDPRHandler struct {
	gdprService *service.GDPRService
	auditor     audit.Auditor
}

// NewGDPRHandler creates a new GDPR handler
func NewGDPRHandler(gdprService *service.GDPRService, auditor audit.Auditor) *GDPRHandler {
	return &GDPRHandler{
		gdprService: gdprService,
		auditor:     auditor,
	}
}

// ExportData handles GET /v1/me/export
// @Summary Export user data (GDPR)
// @Description Export all user data for GDPR compliance
// @Tags GDPR
// @Security bearerAuth
// @Produce json
// @Success 200 {object} service.GDPRExportData
// @Failure 401 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/me/export [get]
func (h *GDPRHandler) ExportData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}
	tenantID, _ := c.Get("tenant_id")

	data, err := h.gdprService.ExportUserData(c.Request.Context(), userID.(string), tenantID.(string))
	if err != nil {
		var domErr *domain.Error
		if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	// Audit log
	if h.auditor != nil {
		h.auditor.Event(c.Request.Context(), tenantID.(string), "GDPR_DATA_EXPORT", userID.(string), userID.(string), nil)
	}

	c.JSON(http.StatusOK, data)
}

// ExportDataDownload handles GET /v1/me/export/download
// @Summary Download user data as JSON file (GDPR)
// @Description Download all user data as a JSON file for GDPR compliance
// @Tags GDPR
// @Security bearerAuth
// @Produce application/json
// @Success 200 {file} file
// @Failure 401 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/me/export/download [get]
func (h *GDPRHandler) ExportDataDownload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}
	tenantID, _ := c.Get("tenant_id")

	jsonData, err := h.gdprService.ExportUserDataJSON(c.Request.Context(), userID.(string), tenantID.(string))
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	// Audit log
	if h.auditor != nil {
		h.auditor.Event(c.Request.Context(), tenantID.(string), "GDPR_DATA_DOWNLOAD", userID.(string), userID.(string), nil)
	}

	c.Header("Content-Disposition", "attachment; filename=goconnect-data-export.json")
	c.Data(http.StatusOK, "application/json", jsonData)
}

// RequestDeletionRequest is the request body for deletion
type RequestDeletionRequest struct {
	Confirmation string `json:"confirmation" binding:"required"` // Must be "DELETE MY ACCOUNT"
}

// RequestDeletion handles POST /v1/me/delete
// @Summary Request account deletion (GDPR)
// @Description Request account and data deletion for GDPR compliance
// @Tags GDPR
// @Security bearerAuth
// @Accept json
// @Produce json
// @Param body body RequestDeletionRequest true "Deletion confirmation"
// @Success 202 {object} service.GDPRDeleteRequest
// @Failure 400 {object} domain.Error
// @Failure 401 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/me/delete [post]
func (h *GDPRHandler) RequestDeletion(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Authentication required", nil))
		return
	}

	var req RequestDeletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request body", nil))
		return
	}

	// Require explicit confirmation
	if req.Confirmation != "DELETE MY ACCOUNT" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid confirmation. Must be exactly 'DELETE MY ACCOUNT'",
			map[string]string{"expected": "DELETE MY ACCOUNT"}))
		return
	}

	deleteReq, err := h.gdprService.RequestDeletion(c.Request.Context(), userID.(string))
	if err != nil {
		var domErr *domain.Error
		if errors.As(err, &domErr) {
			errorResponse(c, domErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, err.Error(), nil))
		}
		return
	}

	// Audit log
	if h.auditor != nil {
		tenantID, _ := c.Get("tenant_id")
		h.auditor.Event(c.Request.Context(), tenantID.(string), "GDPR_DELETE_REQUEST", userID.(string), userID.(string), map[string]interface{}{
			"request_id": deleteReq.ID,
		})
	}

	c.JSON(http.StatusAccepted, deleteReq)
}
