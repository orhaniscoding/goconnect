package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== IPRuleType Tests ====================

func TestIPRuleType_Constants(t *testing.T) {
	t.Run("Types Are Correct", func(t *testing.T) {
		assert.Equal(t, IPRuleType("allow"), IPRuleTypeAllow)
		assert.Equal(t, IPRuleType("deny"), IPRuleTypeDeny)
	})
}

// ==================== IPRule.IsActive Tests ====================

func TestIPRule_IsActive(t *testing.T) {
	t.Run("Active When No Expiration", func(t *testing.T) {
		rule := &IPRule{
			ID:        "ipr123",
			CIDR:      "192.168.1.0/24",
			ExpiresAt: nil,
		}
		assert.True(t, rule.IsActive())
	})

	t.Run("Active When Not Expired", func(t *testing.T) {
		future := time.Now().Add(time.Hour)
		rule := &IPRule{
			ID:        "ipr123",
			CIDR:      "192.168.1.0/24",
			ExpiresAt: &future,
		}
		assert.True(t, rule.IsActive())
	})

	t.Run("Inactive When Expired", func(t *testing.T) {
		past := time.Now().Add(-time.Hour)
		rule := &IPRule{
			ID:        "ipr123",
			CIDR:      "192.168.1.0/24",
			ExpiresAt: &past,
		}
		assert.False(t, rule.IsActive())
	})
}

// ==================== IPRule.MatchesIP Tests ====================

func TestIPRule_MatchesIP(t *testing.T) {
	t.Run("Matches IP In CIDR Range", func(t *testing.T) {
		rule := &IPRule{CIDR: "192.168.1.0/24"}
		assert.True(t, rule.MatchesIP("192.168.1.50"))
		assert.True(t, rule.MatchesIP("192.168.1.1"))
		assert.True(t, rule.MatchesIP("192.168.1.254"))
	})

	t.Run("Does Not Match IP Outside CIDR Range", func(t *testing.T) {
		rule := &IPRule{CIDR: "192.168.1.0/24"}
		assert.False(t, rule.MatchesIP("192.168.2.1"))
		assert.False(t, rule.MatchesIP("10.0.0.1"))
	})

	t.Run("Matches Single IP", func(t *testing.T) {
		rule := &IPRule{CIDR: "10.0.0.5"}
		assert.True(t, rule.MatchesIP("10.0.0.5"))
		assert.False(t, rule.MatchesIP("10.0.0.6"))
	})

	t.Run("Returns False For Invalid IP", func(t *testing.T) {
		rule := &IPRule{CIDR: "192.168.1.0/24"}
		assert.False(t, rule.MatchesIP("invalid"))
		assert.False(t, rule.MatchesIP(""))
	})

	t.Run("Returns False For Invalid CIDR", func(t *testing.T) {
		rule := &IPRule{CIDR: "invalid"}
		assert.False(t, rule.MatchesIP("192.168.1.1"))
	})

	t.Run("Matches IPv6", func(t *testing.T) {
		rule := &IPRule{CIDR: "2001:db8::/32"}
		assert.True(t, rule.MatchesIP("2001:db8::1"))
		assert.False(t, rule.MatchesIP("2001:db9::1"))
	})
}

// ==================== ValidateIPRuleCIDR Tests ====================

func TestValidateIPRuleCIDR(t *testing.T) {
	t.Run("Valid CIDR", func(t *testing.T) {
		err := ValidateIPRuleCIDR("192.168.1.0/24")
		assert.NoError(t, err)
	})

	t.Run("Valid Single IP", func(t *testing.T) {
		err := ValidateIPRuleCIDR("10.0.0.5")
		assert.NoError(t, err)
	})

	t.Run("Valid IPv6 CIDR", func(t *testing.T) {
		err := ValidateIPRuleCIDR("2001:db8::/32")
		assert.NoError(t, err)
	})

	t.Run("Invalid CIDR Returns Error", func(t *testing.T) {
		err := ValidateIPRuleCIDR("invalid")
		require.Error(t, err)
		domErr, ok := err.(*Error)
		require.True(t, ok)
		assert.Equal(t, ErrInvalidRequest, domErr.Code)
	})
}

// ==================== GenerateIPRuleID Tests ====================

func TestGenerateIPRuleID(t *testing.T) {
	t.Run("Generates ID With Prefix", func(t *testing.T) {
		id := GenerateIPRuleID()
		assert.Contains(t, id, "ipr_")
	})

	t.Run("Generates Unique IDs", func(t *testing.T) {
		id1 := GenerateIPRuleID()
		id2 := GenerateIPRuleID()
		assert.NotEqual(t, id1, id2)
	})
}

// ==================== Struct Tests ====================

func TestIPRule(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		expires := now.Add(24 * time.Hour)
		rule := IPRule{
			ID:          "ipr123",
			TenantID:    "tenant123",
			Type:        IPRuleTypeAllow,
			CIDR:        "192.168.1.0/24",
			Description: "Office network",
			CreatedBy:   "user123",
			CreatedAt:   now,
			UpdatedAt:   now,
			ExpiresAt:   &expires,
		}

		assert.Equal(t, "ipr123", rule.ID)
		assert.Equal(t, IPRuleTypeAllow, rule.Type)
	})
}

func TestCreateIPRuleRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		req := CreateIPRuleRequest{
			TenantID:    "tenant123",
			Type:        IPRuleTypeDeny,
			CIDR:        "10.0.0.0/8",
			Description: "Block internal",
			CreatedBy:   "admin",
		}

		assert.Equal(t, IPRuleTypeDeny, req.Type)
		assert.Equal(t, "10.0.0.0/8", req.CIDR)
	})
}

func TestIPRuleResponse(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		resp := IPRuleResponse{
			ID:          "ipr123",
			TenantID:    "tenant123",
			Type:        IPRuleTypeAllow,
			CIDR:        "192.168.1.0/24",
			Description: "Office",
			CreatedBy:   "admin",
			CreatedAt:   time.Now(),
			IsActive:    true,
		}

		assert.True(t, resp.IsActive)
	})
}
