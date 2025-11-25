package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// IPRuleService handles IP rule business logic
type IPRuleService struct {
	repo repository.IPRuleRepository
}

// NewIPRuleService creates a new IP rule service
func NewIPRuleService(repo repository.IPRuleRepository) *IPRuleService {
	return &IPRuleService{repo: repo}
}

// CreateIPRule creates a new IP rule
func (s *IPRuleService) CreateIPRule(ctx context.Context, req domain.CreateIPRuleRequest) (*domain.IPRule, error) {
	// Validate CIDR format
	if err := domain.ValidateIPRuleCIDR(req.CIDR); err != nil {
		return nil, err
	}

	// Validate rule type
	if req.Type != domain.IPRuleTypeAllow && req.Type != domain.IPRuleTypeDeny {
		return nil, domain.NewError(domain.ErrValidation, "invalid rule type, must be 'allow' or 'deny'", nil)
	}

	now := time.Now()
	rule := &domain.IPRule{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Type:        req.Type,
		CIDR:        req.CIDR,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// GetIPRule gets an IP rule by ID
func (s *IPRuleService) GetIPRule(ctx context.Context, id string) (*domain.IPRule, error) {
	return s.repo.GetByID(ctx, id)
}

// ListIPRules lists all IP rules for a tenant
func (s *IPRuleService) ListIPRules(ctx context.Context, tenantID string) ([]*domain.IPRule, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

// DeleteIPRule deletes an IP rule
func (s *IPRuleService) DeleteIPRule(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CheckIP checks if an IP is allowed for a tenant
// Returns: allowed bool, matchedRule *IPRule, error
func (s *IPRuleService) CheckIP(ctx context.Context, tenantID, ipAddr string) (bool, *domain.IPRule, error) {
	rules, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return false, nil, err
	}

	// If no rules, allow by default
	if len(rules) == 0 {
		return true, nil, nil
	}

	// Check deny rules first (deny takes precedence)
	for _, rule := range rules {
		if rule.Type == domain.IPRuleTypeDeny && rule.IsActive() && rule.MatchesIP(ipAddr) {
			return false, rule, nil
		}
	}

	// Check if there are any allow rules
	hasAllowRules := false
	for _, rule := range rules {
		if rule.Type == domain.IPRuleTypeAllow && rule.IsActive() {
			hasAllowRules = true
			if rule.MatchesIP(ipAddr) {
				return true, rule, nil
			}
		}
	}

	// If there are allow rules but none matched, deny
	if hasAllowRules {
		return false, nil, nil
	}

	// No allow rules, so allow by default
	return true, nil, nil
}

// CleanupExpired removes expired IP rules
func (s *IPRuleService) CleanupExpired(ctx context.Context) (int, error) {
	return s.repo.DeleteExpired(ctx)
}
