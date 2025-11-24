package handler

import (
	"context"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// newMockAuthService creates a mock auth service for testing
func newMockAuthService() *service.AuthService {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	return service.NewAuthService(userRepo, tenantRepo, nil)
}

// mockAuthServiceWithValidToken creates a mock auth service that accepts "dev" and "admin" tokens
type mockAuthService struct {
	*service.AuthService
}

func newMockAuthServiceWithTokens() *mockAuthService {
	return &mockAuthService{
		AuthService: newMockAuthService(),
	}
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	// Mock implementation for testing - accepts multiple test tokens
	// IMPORTANT: Tenant IDs must match test data (most tests use "t1")
	switch token {
	case "dev":
		return &domain.TokenClaims{
			UserID:   "user_dev",
			TenantID: "t1", // Changed from "tenant_dev" to match test data
			Email:    "dev@example.com",
			IsAdmin:  false,
		}, nil
	case "admin":
		return &domain.TokenClaims{
			UserID:   "admin_dev",
			TenantID: "t1", // Changed from "tenant_admin" to match test data
			Email:    "admin@example.com",
			IsAdmin:  true,
		}, nil
	case "valid-token":
		// Default valid token for tests - maps to user_dev for consistency
		return &domain.TokenClaims{
			UserID:   "user_dev",
			TenantID: "t1",
			Email:    "test@example.com",
			IsAdmin:  false, // Some tests expect non-admin
		}, nil
	case "member-token":
		// Non-admin member token for RBAC tests
		return &domain.TokenClaims{
			UserID:   "test_user_2",
			TenantID: "t1", // Changed from "test_tenant_2" to match test data
			Email:    "member@example.com",
			IsAdmin:  false,
		}, nil
	case "admin-token":
		// Admin token for RBAC tests (maps to user_admin)
		return &domain.TokenClaims{
			UserID:   "user_admin",
			TenantID: "t1",
			Email:    "admin-test@example.com",
			IsAdmin:  true,
		}, nil
	default:
		return nil, domain.NewError(domain.ErrUnauthorized, "Invalid token", nil)
	}
}

// createTestNetwork creates a network in the repository for testing
func createTestNetwork(ctx context.Context, networkRepo repository.NetworkRepository, createdBy, tenantID, networkID string) error {
	network := &domain.Network{
		ID:         networkID,
		Name:       "Test Network",
		TenantID:   tenantID,
		CreatedBy:  createdBy,
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return networkRepo.Create(ctx, network)
}

// createTestMembership creates a membership in the repository for testing
func createTestMembership(ctx context.Context, membershipRepo repository.MembershipRepository, networkID, userID string, role domain.MembershipRole) error {
	_, err := membershipRepo.UpsertApproved(ctx, networkID, userID, role, time.Now())
	return err
}
