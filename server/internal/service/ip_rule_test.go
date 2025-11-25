package service

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPRuleService(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repo)
}

func TestIPRuleService_CreateIPRule_Success(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        domain.IPRuleTypeAllow,
		CIDR:        "192.168.1.0/24",
		Description: "Office network",
		CreatedBy:   "admin",
	}

	rule, err := service.CreateIPRule(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.NotEmpty(t, rule.ID)
	assert.Equal(t, "tenant-1", rule.TenantID)
	assert.Equal(t, domain.IPRuleTypeAllow, rule.Type)
	assert.Equal(t, "192.168.1.0/24", rule.CIDR)
	assert.Equal(t, "Office network", rule.Description)
	assert.Equal(t, "admin", rule.CreatedBy)
}

func TestIPRuleService_CreateIPRule_DenyRule(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        domain.IPRuleTypeDeny,
		CIDR:        "10.0.0.0/8",
		Description: "Block internal network",
		CreatedBy:   "admin",
	}

	rule, err := service.CreateIPRule(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, domain.IPRuleTypeDeny, rule.Type)
}

func TestIPRuleService_CreateIPRule_WithExpiration(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	expiresAt := time.Now().Add(24 * time.Hour)
	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        domain.IPRuleTypeAllow,
		CIDR:        "192.168.1.100/32",
		Description: "Temporary access",
		CreatedBy:   "admin",
		ExpiresAt:   &expiresAt,
	}

	rule, err := service.CreateIPRule(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, rule.ExpiresAt)
	assert.True(t, rule.ExpiresAt.After(time.Now()))
}

func TestIPRuleService_CreateIPRule_InvalidCIDR(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        domain.IPRuleTypeAllow,
		CIDR:        "invalid-cidr",
		Description: "Invalid rule",
		CreatedBy:   "admin",
	}

	rule, err := service.CreateIPRule(ctx, req)

	require.Error(t, err)
	assert.Nil(t, rule)
}

func TestIPRuleService_CreateIPRule_InvalidType(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        "invalid",
		CIDR:        "192.168.1.0/24",
		Description: "Invalid type",
		CreatedBy:   "admin",
	}

	rule, err := service.CreateIPRule(ctx, req)

	require.Error(t, err)
	assert.Nil(t, rule)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrValidation, domainErr.Code)
}

func TestIPRuleService_GetIPRule_Success(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create a rule first
	req := domain.CreateIPRuleRequest{
		TenantID:    "tenant-1",
		Type:        domain.IPRuleTypeAllow,
		CIDR:        "192.168.1.0/24",
		Description: "Office network",
		CreatedBy:   "admin",
	}
	created, _ := service.CreateIPRule(ctx, req)

	// Get the rule
	rule, err := service.GetIPRule(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, rule.ID)
}

func TestIPRuleService_GetIPRule_NotFound(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	rule, err := service.GetIPRule(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, rule)
}

func TestIPRuleService_ListIPRules_Success(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create multiple rules
	reqs := []domain.CreateIPRuleRequest{
		{TenantID: "tenant-1", Type: domain.IPRuleTypeAllow, CIDR: "192.168.1.0/24", CreatedBy: "admin"},
		{TenantID: "tenant-1", Type: domain.IPRuleTypeDeny, CIDR: "10.0.0.0/8", CreatedBy: "admin"},
		{TenantID: "tenant-2", Type: domain.IPRuleTypeAllow, CIDR: "172.16.0.0/12", CreatedBy: "admin"},
	}

	for _, req := range reqs {
		_, _ = service.CreateIPRule(ctx, req)
	}

	// List rules for tenant-1
	rules, err := service.ListIPRules(ctx, "tenant-1")

	require.NoError(t, err)
	assert.Len(t, rules, 2)
}

func TestIPRuleService_ListIPRules_EmptyTenant(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	rules, err := service.ListIPRules(ctx, "nonexistent-tenant")

	require.NoError(t, err)
	assert.Len(t, rules, 0)
}

func TestIPRuleService_DeleteIPRule_Success(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create a rule
	req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
	}
	created, _ := service.CreateIPRule(ctx, req)

	// Delete the rule
	err := service.DeleteIPRule(ctx, created.ID)

	require.NoError(t, err)

	// Verify it's deleted
	_, err = service.GetIPRule(ctx, created.ID)
	require.Error(t, err)
}

func TestIPRuleService_DeleteIPRule_NotFound(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	err := service.DeleteIPRule(ctx, "nonexistent")

	require.Error(t, err)
}

func TestIPRuleService_CheckIP_NoRules_AllowByDefault(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "192.168.1.100")

	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Nil(t, rule)
}

func TestIPRuleService_CheckIP_AllowRule_Match(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create allow rule
	req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, req)

	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "192.168.1.100")

	require.NoError(t, err)
	assert.True(t, allowed)
	assert.NotNil(t, rule)
	assert.Equal(t, domain.IPRuleTypeAllow, rule.Type)
}

func TestIPRuleService_CheckIP_AllowRule_NoMatch(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create allow rule for specific subnet
	req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, req)

	// Check IP outside the allowed range
	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "10.0.0.1")

	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Nil(t, rule)
}

func TestIPRuleService_CheckIP_DenyRule_Match(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create deny rule
	req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "10.0.0.0/8",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, req)

	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "10.1.2.3")

	require.NoError(t, err)
	assert.False(t, allowed)
	assert.NotNil(t, rule)
	assert.Equal(t, domain.IPRuleTypeDeny, rule.Type)
}

func TestIPRuleService_CheckIP_DenyTakesPrecedence(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create both allow and deny rules for overlapping ranges
	allowReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "10.0.0.0/8", // Wide allow
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, allowReq)

	denyReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "10.1.0.0/16", // Specific deny within allow
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, denyReq)

	// IP within deny range (should be denied)
	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "10.1.2.3")

	require.NoError(t, err)
	assert.False(t, allowed)
	assert.NotNil(t, rule)
	assert.Equal(t, domain.IPRuleTypeDeny, rule.Type)
}

func TestIPRuleService_CheckIP_AllowedOutsideDeny(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create both allow and deny rules
	allowReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "10.0.0.0/8",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, allowReq)

	denyReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "10.1.0.0/16",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, denyReq)

	// IP within allow but outside deny (should be allowed)
	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "10.2.0.1")

	require.NoError(t, err)
	assert.True(t, allowed)
	assert.NotNil(t, rule)
	assert.Equal(t, domain.IPRuleTypeAllow, rule.Type)
}

func TestIPRuleService_CheckIP_ExpiredRuleIgnored(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create expired deny rule
	pastTime := time.Now().Add(-1 * time.Hour)
	denyReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
		ExpiresAt: &pastTime,
	}
	_, _ = service.CreateIPRule(ctx, denyReq)

	// Expired deny rule should be ignored, allow by default
	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "192.168.1.100")

	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Nil(t, rule)
}

func TestIPRuleService_CheckIP_SingleIP(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create allow rule for single IP
	req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.100/32",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, req)

	// Exact match
	allowed, rule, err := service.CheckIP(ctx, "tenant-1", "192.168.1.100")
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.NotNil(t, rule)

	// Not exact match
	allowed2, rule2, err2 := service.CheckIP(ctx, "tenant-1", "192.168.1.101")
	require.NoError(t, err2)
	assert.False(t, allowed2)
	assert.Nil(t, rule2)
}

func TestIPRuleService_CleanupExpired_Success(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create expired and valid rules
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	expiredReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "10.0.0.0/8",
		CreatedBy: "admin",
		ExpiresAt: &pastTime,
	}
	_, _ = service.CreateIPRule(ctx, expiredReq)

	validReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
		ExpiresAt: &futureTime,
	}
	_, _ = service.CreateIPRule(ctx, validReq)

	noExpiryReq := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "172.16.0.0/12",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, noExpiryReq)

	// Cleanup expired rules
	deleted, err := service.CleanupExpired(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify only valid rules remain
	rules, _ := service.ListIPRules(ctx, "tenant-1")
	assert.Len(t, rules, 2)
}

func TestIPRuleService_MultiTenant_Isolation(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	service := NewIPRuleService(repo)
	ctx := context.Background()

	// Create rules for different tenants
	tenant1Req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeDeny,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, tenant1Req)

	tenant2Req := domain.CreateIPRuleRequest{
		TenantID:  "tenant-2",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		CreatedBy: "admin",
	}
	_, _ = service.CreateIPRule(ctx, tenant2Req)

	// Same IP should be denied for tenant-1 but allowed for tenant-2
	allowed1, _, _ := service.CheckIP(ctx, "tenant-1", "192.168.1.100")
	assert.False(t, allowed1)

	allowed2, _, _ := service.CheckIP(ctx, "tenant-2", "192.168.1.100")
	assert.True(t, allowed2)
}
