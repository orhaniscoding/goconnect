package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// IPRuleHandler handles IP rule HTTP endpoints
type IPRuleHandler struct {
	svc *service.IPRuleService
}

// NewIPRuleHandler creates a new IP rule handler
func NewIPRuleHandler(svc *service.IPRuleService) *IPRuleHandler {
	return &IPRuleHandler{svc: svc}
}

// CreateIPRuleRequest is the request body for creating an IP rule
type CreateIPRuleRequest struct {
	Type        string     `json:"type" binding:"required"`
	CIDR        string     `json:"cidr" binding:"required"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// IPRuleResponse is the response for an IP rule
type IPRuleResponse struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	Type        string     `json:"type"`
	CIDR        string     `json:"cidr"`
	Description string     `json:"description,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// IPRulesListResponse is the response for listing IP rules
type IPRulesListResponse struct {
	Rules []IPRuleResponse `json:"rules"`
	Total int              `json:"total"`
}

func toIPRuleResponse(rule *domain.IPRule) IPRuleResponse {
	return IPRuleResponse{
		ID:          rule.ID,
		TenantID:    rule.TenantID,
		Type:        string(rule.Type),
		CIDR:        rule.CIDR,
		Description: rule.Description,
		CreatedBy:   rule.CreatedBy,
		CreatedAt:   rule.CreatedAt,
		ExpiresAt:   rule.ExpiresAt,
	}
}

// CreateIPRule handles POST /v1/admin/ip-rules
func (h *IPRuleHandler) CreateIPRule(c *gin.Context) {
	var req CreateIPRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrValidation, "Invalid request body: "+err.Error(), nil))
		return
	}

	// Get tenant ID and user ID from context (set by auth middleware)
	// Fallback to headers for compatibility/testing if context keys missing (though middleware should set them)
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}
	userID := c.GetString("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if tenantID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Tenant ID is required", nil))
		return
	}

	createReq := domain.CreateIPRuleRequest{
		TenantID:    tenantID,
		Type:        domain.IPRuleType(req.Type),
		CIDR:        req.CIDR,
		Description: req.Description,
		CreatedBy:   userID,
		ExpiresAt:   req.ExpiresAt,
	}

	rule, err := h.svc.CreateIPRule(c.Request.Context(), createReq)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
			return
		}
		slog.Error("CreateIPRule: Failed to create rule", "error", err, "user_id", userID, "tenant_id", tenantID)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}
	slog.Info("IP rule created successfully", "rule_id", rule.ID, "user_id", userID, "tenant_id", tenantID)
	c.JSON(http.StatusCreated, toIPRuleResponse(rule))
}

// ListIPRules handles GET /v1/admin/ip-rules
func (h *IPRuleHandler) ListIPRules(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}
	if tenantID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Tenant ID is required", nil))
		return
	}

	rules, err := h.svc.ListIPRules(c.Request.Context(), tenantID)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	responses := make([]IPRuleResponse, 0, len(rules))
	for _, rule := range rules {
		responses = append(responses, toIPRuleResponse(rule))
	}

	c.JSON(http.StatusOK, IPRulesListResponse{
		Rules: responses,
		Total: len(responses),
	})
}

// GetIPRule handles GET /v1/admin/ip-rules/{id}
func (h *IPRuleHandler) GetIPRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Rule ID is required", nil))
		return
	}

	rule, err := h.svc.GetIPRule(c.Request.Context(), id)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	// Verify tenant ownership
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}

	if rule.TenantID != tenantID {
		errorResponse(c, domain.NewError(domain.ErrNotFound, "IP rule not found", nil))
		return
	}

	c.JSON(http.StatusOK, toIPRuleResponse(rule))
}

// DeleteIPRule handles DELETE /v1/admin/ip-rules/{id}
func (h *IPRuleHandler) DeleteIPRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Rule ID is required", nil))
		return
	}

	// Verify tenant ownership before deleting
	rule, err := h.svc.GetIPRule(c.Request.Context(), id)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			errorResponse(c, domainErr)
			return
		}
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}
	if rule.TenantID != tenantID {
		errorResponse(c, domain.NewError(domain.ErrNotFound, "IP rule not found", nil))
		return
	}

	if err := h.svc.DeleteIPRule(c.Request.Context(), id); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	c.Status(http.StatusNoContent)
}

// CheckIPRequest is the request body for checking an IP
type CheckIPRequest struct {
	IP string `json:"ip" binding:"required"`
}

// CheckIPResponse is the response for IP check
type CheckIPResponse struct {
	Allowed     bool            `json:"allowed"`
	MatchedRule *IPRuleResponse `json:"matched_rule,omitempty"`
}

// CheckIP handles POST /v1/admin/ip-rules/check
func (h *IPRuleHandler) CheckIP(c *gin.Context) {
	var req CheckIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrValidation, "Invalid request body: "+err.Error(), nil))
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}
	if tenantID == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Tenant ID is required", nil))
		return
	}

	allowed, matchedRule, err := h.svc.CheckIP(c.Request.Context(), tenantID, req.IP)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		return
	}

	response := CheckIPResponse{Allowed: allowed}
	if matchedRule != nil {
		ruleResponse := toIPRuleResponse(matchedRule)
		response.MatchedRule = &ruleResponse
	}

	c.JSON(http.StatusOK, response)
}
