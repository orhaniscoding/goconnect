package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/middleware"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// ══════════════════════════════════════════════════════════════════════════════
// SECTION HANDLER
// ══════════════════════════════════════════════════════════════════════════════
// HTTP handlers for section management endpoints

type SectionHandler struct {
	sectionService *service.SectionService
}

// NewSectionHandler creates a new section handler
func NewSectionHandler(sectionService *service.SectionService) *SectionHandler {
	return &SectionHandler{
		sectionService: sectionService,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// REQUEST/RESPONSE TYPES
// ══════════════════════════════════════════════════════════════════════════════

type CreateSectionRequest struct {
	Name        string                     `json:"name" binding:"required,min=1,max=100"`
	Description string                     `json:"description,omitempty" binding:"max=500"`
	Icon        string                     `json:"icon,omitempty" binding:"max=255"`
	Visibility  domain.SectionVisibility   `json:"visibility,omitempty"`
}

type UpdateSectionRequest struct {
	Name        *string                    `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string                    `json:"description,omitempty" binding:"omitempty,max=500"`
	Icon        *string                    `json:"icon,omitempty" binding:"omitempty,max=255"`
	Visibility  *domain.SectionVisibility  `json:"visibility,omitempty"`
}

type UpdatePositionsRequest struct {
	Positions map[string]int `json:"positions" binding:"required"`
}

type SectionResponse struct {
	ID          string                   `json:"id"`
	TenantID    string                   `json:"tenant_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Icon        string                   `json:"icon,omitempty"`
	Position    int                      `json:"position"`
	Visibility  domain.SectionVisibility `json:"visibility"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

// ══════════════════════════════════════════════════════════════════════════════
// HANDLERS
// ══════════════════════════════════════════════════════════════════════════════

// Create creates a new section
// POST /api/v2/servers/:tenantID/sections
func (h *SectionHandler) Create(c *gin.Context) {
	// Extract tenant ID from URL
	tenantID := c.Param("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Parse request body
	var req CreateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Call service
	section, err := h.sectionService.Create(c.Request.Context(), service.CreateSectionInput{
		TenantID:    tenantID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Visibility:  req.Visibility,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	// Return response
	c.JSON(http.StatusCreated, toSectionResponse(section))
}

// GetByID retrieves a section by ID
// GET /api/v2/sections/:sectionID
func (h *SectionHandler) GetByID(c *gin.Context) {
	// Extract section ID from URL
	sectionID := c.Param("sectionID")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "section_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Call service
	section, err := h.sectionService.GetByID(c.Request.Context(), userID, sectionID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return response
	c.JSON(http.StatusOK, toSectionResponse(section))
}

// List retrieves all sections for a tenant
// GET /api/v2/servers/:tenantID/sections
func (h *SectionHandler) List(c *gin.Context) {
	// Extract tenant ID from URL
	tenantID := c.Param("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Call service
	sections, err := h.sectionService.ListByTenant(c.Request.Context(), userID, tenantID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return response
	response := make([]SectionResponse, len(sections))
	for i, section := range sections {
		response[i] = toSectionResponse(&section)
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": response,
		"count":    len(response),
	})
}

// Update updates a section
// PATCH /api/v2/sections/:sectionID
func (h *SectionHandler) Update(c *gin.Context) {
	// Extract section ID from URL
	sectionID := c.Param("sectionID")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "section_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Parse request body
	var req UpdateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Call service
	section, err := h.sectionService.Update(c.Request.Context(), service.UpdateSectionInput{
		UserID:      userID,
		SectionID:   sectionID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Visibility:  req.Visibility,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	// Return response
	c.JSON(http.StatusOK, toSectionResponse(section))
}

// Delete deletes a section
// DELETE /api/v2/sections/:sectionID
func (h *SectionHandler) Delete(c *gin.Context) {
	// Extract section ID from URL
	sectionID := c.Param("sectionID")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "section_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Call service
	err := h.sectionService.Delete(c.Request.Context(), userID, sectionID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return success
	c.JSON(http.StatusNoContent, nil)
}

// UpdatePositions updates positions for multiple sections
// PATCH /api/v2/servers/:tenantID/sections/positions
func (h *SectionHandler) UpdatePositions(c *gin.Context) {
	// Extract tenant ID from URL
	tenantID := c.Param("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tenant_id is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	// Parse request body
	var req UpdatePositionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Call service
	err := h.sectionService.UpdatePositions(c.Request.Context(), userID, tenantID, req.Positions)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"message": "positions updated successfully",
	})
}

// ══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ══════════════════════════════════════════════════════════════════════════════

func toSectionResponse(section *domain.Section) SectionResponse {
	return SectionResponse{
		ID:          section.ID,
		TenantID:    section.TenantID,
		Name:        section.Name,
		Description: section.Description,
		Icon:        section.Icon,
		Position:    section.Position,
		Visibility:  section.Visibility,
		CreatedAt:   section.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   section.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// handleError converts domain errors to HTTP responses
func handleError(c *gin.Context, err error) {
	// Check if it's a domain error
	if domainErr, ok := err.(*domain.Error); ok {
		statusCode := domainErrorToHTTPStatus(domainErr.Code)
		c.JSON(statusCode, gin.H{
			"error":   domainErr.Message,
			"code":    string(domainErr.Code),
			"details": domainErr.Details,
		})
		return
	}

	// Generic internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
		"code":  "INTERNAL_ERROR",
	})
}

// domainErrorToHTTPStatus maps domain error codes to HTTP status codes
func domainErrorToHTTPStatus(code string) int {
	switch code {
	case domain.ErrValidation:
		return http.StatusBadRequest
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrForbidden:
		return http.StatusForbidden
	case domain.ErrUnauthorized:
		return http.StatusUnauthorized
	case domain.ErrConflict:
		return http.StatusConflict
	case domain.ErrInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
