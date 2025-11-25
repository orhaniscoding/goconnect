package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
	Type        string     `json:"type"`
	CIDR        string     `json:"cidr"`
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

// httpErrorResponse writes a JSON error response
func httpErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}

// CreateIPRule handles POST /v1/admin/ip-rules
func (h *IPRuleHandler) CreateIPRule(w http.ResponseWriter, r *http.Request) {
	var req CreateIPRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get tenant ID and user ID from context (set by auth middleware)
	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	if tenantID == "" {
		httpErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
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

	rule, err := h.svc.CreateIPRule(r.Context(), createReq)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) {
			switch domainErr.Code {
			case domain.ErrValidation, domain.ErrInvalidRequest:
				httpErrorResponse(w, http.StatusBadRequest, domainErr.Message)
			case domain.ErrConflict:
				httpErrorResponse(w, http.StatusConflict, domainErr.Message)
			default:
				httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toIPRuleResponse(rule))
}

// ListIPRules handles GET /v1/admin/ip-rules
func (h *IPRuleHandler) ListIPRules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		httpErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	rules, err := h.svc.ListIPRules(r.Context(), tenantID)
	if err != nil {
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	responses := make([]IPRuleResponse, 0, len(rules))
	for _, rule := range rules {
		responses = append(responses, toIPRuleResponse(rule))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IPRulesListResponse{
		Rules: responses,
		Total: len(responses),
	})
}

// GetIPRule handles GET /v1/admin/ip-rules/{id}
func (h *IPRuleHandler) GetIPRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpErrorResponse(w, http.StatusBadRequest, "rule ID is required")
		return
	}

	rule, err := h.svc.GetIPRule(r.Context(), id)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) && domainErr.Code == domain.ErrNotFound {
			httpErrorResponse(w, http.StatusNotFound, "IP rule not found")
			return
		}
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Verify tenant ownership
	tenantID := r.Header.Get("X-Tenant-ID")
	if rule.TenantID != tenantID {
		httpErrorResponse(w, http.StatusNotFound, "IP rule not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toIPRuleResponse(rule))
}

// DeleteIPRule handles DELETE /v1/admin/ip-rules/{id}
func (h *IPRuleHandler) DeleteIPRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpErrorResponse(w, http.StatusBadRequest, "rule ID is required")
		return
	}

	// Verify tenant ownership before deleting
	rule, err := h.svc.GetIPRule(r.Context(), id)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) && domainErr.Code == domain.ErrNotFound {
			httpErrorResponse(w, http.StatusNotFound, "IP rule not found")
			return
		}
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if rule.TenantID != tenantID {
		httpErrorResponse(w, http.StatusNotFound, "IP rule not found")
		return
	}

	if err := h.svc.DeleteIPRule(r.Context(), id); err != nil {
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckIPRequest is the request body for checking an IP
type CheckIPRequest struct {
	IP string `json:"ip"`
}

// CheckIPResponse is the response for IP check
type CheckIPResponse struct {
	Allowed     bool            `json:"allowed"`
	MatchedRule *IPRuleResponse `json:"matched_rule,omitempty"`
}

// CheckIP handles POST /v1/admin/ip-rules/check
func (h *IPRuleHandler) CheckIP(w http.ResponseWriter, r *http.Request) {
	var req CheckIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.IP == "" {
		httpErrorResponse(w, http.StatusBadRequest, "IP address is required")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		httpErrorResponse(w, http.StatusBadRequest, "tenant ID is required")
		return
	}

	allowed, matchedRule, err := h.svc.CheckIP(r.Context(), tenantID, req.IP)
	if err != nil {
		httpErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response := CheckIPResponse{Allowed: allowed}
	if matchedRule != nil {
		ruleResponse := toIPRuleResponse(matchedRule)
		response.MatchedRule = &ruleResponse
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
