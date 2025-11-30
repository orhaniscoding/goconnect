package domain

import (
	"net"
	"time"
)

// IPRuleType defines the type of IP rule
type IPRuleType string

const (
	IPRuleTypeAllow IPRuleType = "allow"
	IPRuleTypeDeny  IPRuleType = "deny"
)

// IPRule represents an IP allow/deny rule for a tenant
type IPRule struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Type        IPRuleType `json:"type" db:"type"` // allow or deny
	CIDR        string     `json:"cidr" db:"cidr"` // IP or CIDR range (e.g., "192.168.1.0/24" or "10.0.0.1/32")
	Description string     `json:"description,omitempty" db:"description"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"` // Optional expiration
}

// CreateIPRuleRequest represents the request to create an IP rule
type CreateIPRuleRequest struct {
	TenantID    string     `json:"-"` // Set from context
	Type        IPRuleType `json:"type" binding:"required,oneof=allow deny"`
	CIDR        string     `json:"cidr" binding:"required"`
	Description string     `json:"description" binding:"omitempty,max=255"`
	CreatedBy   string     `json:"-"` // Set from context
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// IPRuleResponse represents the API response for an IP rule
type IPRuleResponse struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	Type        IPRuleType `json:"type"`
	CIDR        string     `json:"cidr"`
	Description string     `json:"description,omitempty"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// IsActive checks if the rule is currently active
func (r *IPRule) IsActive() bool {
	if r.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(*r.ExpiresAt)
}

// MatchesIP checks if the given IP matches this rule
func (r *IPRule) MatchesIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, cidrNet, err := net.ParseCIDR(r.CIDR)
	if err != nil {
		// Try parsing as single IP
		ruleIP := net.ParseIP(r.CIDR)
		if ruleIP != nil {
			return ip.Equal(ruleIP)
		}
		return false
	}

	return cidrNet.Contains(ip)
}

// ValidateIPRuleCIDR validates if the CIDR is valid
func ValidateIPRuleCIDR(cidr string) error {
	// Try as CIDR first
	_, _, err := net.ParseCIDR(cidr)
	if err == nil {
		return nil
	}

	// Try as single IP
	ip := net.ParseIP(cidr)
	if ip != nil {
		return nil
	}

	return NewError(ErrInvalidRequest, "Invalid IP or CIDR format", map[string]string{"cidr": cidr})
}

// GenerateIPRuleID generates a new ID for IP rules
func GenerateIPRuleID() string {
	return "ipr_" + GenerateNetworkID()[4:] // reuse network ID generator, change prefix
}
