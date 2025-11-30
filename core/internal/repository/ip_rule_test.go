package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test IP rule
func mkIPRule(id, tenantID string, ruleType domain.IPRuleType, cidr, description, createdBy string, expiresAt *time.Time) *domain.IPRule {
	now := time.Now()
	return &domain.IPRule{
		ID:          id,
		TenantID:    tenantID,
		Type:        ruleType,
		CIDR:        cidr,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
	}
}

func TestNewInMemoryIPRuleRepository(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.byID)
	assert.NotNil(t, repo.byTenant)
	assert.Equal(t, 0, len(repo.byID))
}

func TestIPRuleRepository_Create_Success(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()
	rule := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Office network", "admin", nil)

	err := repo.Create(ctx, rule)

	require.NoError(t, err)
	assert.Equal(t, 1, len(repo.byID))
	assert.Equal(t, 1, len(repo.byTenant["tenant-1"]))
}

func TestIPRuleRepository_Create_DuplicateID(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()
	rule1 := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Rule 1", "admin", nil)
	rule2 := mkIPRule("ipr-1", "tenant-2", domain.IPRuleTypeDeny, "10.0.0.0/8", "Rule 2", "admin", nil)

	err1 := repo.Create(ctx, rule1)
	require.NoError(t, err1)

	err2 := repo.Create(ctx, rule2)

	require.Error(t, err2)
	domainErr, ok := err2.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrConflict, domainErr.Code)
}

func TestIPRuleRepository_Create_MultipleRules(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()
	rules := []*domain.IPRule{
		mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Office", "admin", nil),
		mkIPRule("ipr-2", "tenant-1", domain.IPRuleTypeDeny, "10.0.0.0/8", "VPN", "admin", nil),
		mkIPRule("ipr-3", "tenant-2", domain.IPRuleTypeAllow, "172.16.0.0/12", "Cloud", "admin", nil),
	}

	for _, rule := range rules {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(repo.byID))
	assert.Equal(t, 2, len(repo.byTenant["tenant-1"]))
	assert.Equal(t, 1, len(repo.byTenant["tenant-2"]))
}

func TestIPRuleRepository_GetByID_Success(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()
	rule := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Office", "admin", nil)
	_ = repo.Create(ctx, rule)

	result, err := repo.GetByID(ctx, "ipr-1")

	require.NoError(t, err)
	assert.Equal(t, "ipr-1", result.ID)
	assert.Equal(t, "192.168.1.0/24", result.CIDR)
	assert.Equal(t, domain.IPRuleTypeAllow, result.Type)
}

func TestIPRuleRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	result, err := repo.GetByID(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestIPRuleRepository_ListByTenant_Success(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	// Create rules for different tenants
	rule1 := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Rule 1", "admin", nil)
	rule2 := mkIPRule("ipr-2", "tenant-1", domain.IPRuleTypeDeny, "10.0.0.0/8", "Rule 2", "admin", nil)
	rule3 := mkIPRule("ipr-3", "tenant-2", domain.IPRuleTypeAllow, "172.16.0.0/12", "Rule 3", "admin", nil)

	_ = repo.Create(ctx, rule1)
	_ = repo.Create(ctx, rule2)
	_ = repo.Create(ctx, rule3)

	result, err := repo.ListByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestIPRuleRepository_ListByTenant_ExcludesExpired(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	// Create expired and valid rules
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	expiredRule := mkIPRule("ipr-expired", "tenant-1", domain.IPRuleTypeDeny, "10.0.0.0/8", "Expired", "admin", &pastTime)
	validRule := mkIPRule("ipr-valid", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Valid", "admin", &futureTime)
	noExpiryRule := mkIPRule("ipr-noexpiry", "tenant-1", domain.IPRuleTypeAllow, "172.16.0.0/12", "No expiry", "admin", nil)

	_ = repo.Create(ctx, expiredRule)
	_ = repo.Create(ctx, validRule)
	_ = repo.Create(ctx, noExpiryRule)

	result, err := repo.ListByTenant(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify expired rule is excluded
	for _, r := range result {
		assert.NotEqual(t, "ipr-expired", r.ID)
	}
}

func TestIPRuleRepository_ListByTenant_EmptyTenant(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	result, err := repo.ListByTenant(ctx, "nonexistent-tenant")

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestIPRuleRepository_Delete_Success(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()
	rule := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Office", "admin", nil)
	_ = repo.Create(ctx, rule)

	err := repo.Delete(ctx, "ipr-1")

	require.NoError(t, err)
	assert.Equal(t, 0, len(repo.byID))
	assert.Equal(t, 0, len(repo.byTenant["tenant-1"]))
}

func TestIPRuleRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrNotFound, domainErr.Code)
}

func TestIPRuleRepository_DeleteExpired_Success(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	// Create expired and valid rules
	pastTime1 := time.Now().Add(-2 * time.Hour)
	pastTime2 := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	expiredRule1 := mkIPRule("ipr-exp-1", "tenant-1", domain.IPRuleTypeDeny, "10.0.0.0/8", "Expired 1", "admin", &pastTime1)
	expiredRule2 := mkIPRule("ipr-exp-2", "tenant-1", domain.IPRuleTypeDeny, "172.16.0.0/12", "Expired 2", "admin", &pastTime2)
	validRule := mkIPRule("ipr-valid", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Valid", "admin", &futureTime)
	noExpiryRule := mkIPRule("ipr-noexpiry", "tenant-1", domain.IPRuleTypeAllow, "0.0.0.0/0", "No expiry", "admin", nil)

	_ = repo.Create(ctx, expiredRule1)
	_ = repo.Create(ctx, expiredRule2)
	_ = repo.Create(ctx, validRule)
	_ = repo.Create(ctx, noExpiryRule)

	deleted, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 2, deleted)
	assert.Equal(t, 2, len(repo.byID))
}

func TestIPRuleRepository_DeleteExpired_NoExpired(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	rule := mkIPRule("ipr-1", "tenant-1", domain.IPRuleTypeAllow, "192.168.1.0/24", "Valid", "admin", nil)
	_ = repo.Create(ctx, rule)

	deleted, err := repo.DeleteExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
	assert.Equal(t, 1, len(repo.byID))
}

func TestIPRuleRepository_Concurrency(t *testing.T) {
	repo := NewInMemoryIPRuleRepository()
	ctx := context.Background()

	// Run concurrent Create operations
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func(idx int) {
			rule := mkIPRule(
				"ipr-"+string(rune('a'+idx%26))+"-"+time.Now().Format("150405.000000000"),
				"tenant-1",
				domain.IPRuleTypeAllow,
				"192.168.1."+string(rune('0'+idx%10))+"/32",
				"Test rule",
				"admin",
				nil,
			)
			_ = repo.Create(ctx, rule)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// Verify no data corruption
	assert.True(t, len(repo.byID) > 0)
}

func TestIPRule_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{
			name:      "No expiration",
			expiresAt: nil,
			expected:  true,
		},
		{
			name:      "Future expiration",
			expiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			expected:  true,
		},
		{
			name:      "Past expiration",
			expiresAt: func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(),
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule := &domain.IPRule{
				ID:        "ipr-1",
				TenantID:  "tenant-1",
				Type:      domain.IPRuleTypeAllow,
				CIDR:      "192.168.1.0/24",
				ExpiresAt: tc.expiresAt,
			}

			result := rule.IsActive()

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIPRule_MatchesIP(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		ip       string
		expected bool
	}{
		{
			name:     "CIDR match",
			cidr:     "192.168.1.0/24",
			ip:       "192.168.1.100",
			expected: true,
		},
		{
			name:     "CIDR no match",
			cidr:     "192.168.1.0/24",
			ip:       "192.168.2.100",
			expected: false,
		},
		{
			name:     "Single IP match",
			cidr:     "10.0.0.1",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "Single IP no match",
			cidr:     "10.0.0.1",
			ip:       "10.0.0.2",
			expected: false,
		},
		{
			name:     "Invalid IP",
			cidr:     "192.168.1.0/24",
			ip:       "invalid",
			expected: false,
		},
		{
			name:     "Wide CIDR",
			cidr:     "10.0.0.0/8",
			ip:       "10.255.255.255",
			expected: true,
		},
		{
			name:     "IPv6 no match IPv4 CIDR",
			cidr:     "192.168.1.0/24",
			ip:       "::1",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule := &domain.IPRule{
				ID:       "ipr-1",
				TenantID: "tenant-1",
				Type:     domain.IPRuleTypeAllow,
				CIDR:     tc.cidr,
			}

			result := rule.MatchesIP(tc.ip)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateIPRuleCIDR(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{
			name:        "Valid CIDR /24",
			cidr:        "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "Valid CIDR /8",
			cidr:        "10.0.0.0/8",
			expectError: false,
		},
		{
			name:        "Valid single IP",
			cidr:        "192.168.1.1",
			expectError: false,
		},
		{
			name:        "Valid IPv6 CIDR",
			cidr:        "2001:db8::/32",
			expectError: false,
		},
		{
			name:        "Invalid CIDR",
			cidr:        "192.168.1.0/33",
			expectError: true,
		},
		{
			name:        "Invalid IP",
			cidr:        "invalid",
			expectError: true,
		},
		{
			name:        "Empty string",
			cidr:        "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateIPRuleCIDR(tc.cidr)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
