package handler

import (
	"context"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// newMockAuthService creates a mock auth service for testing
func newMockAuthService() *service.AuthService {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	return service.NewAuthService(userRepo, tenantRepo)
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
	// Mock implementation for testing
	switch token {
	case "dev":
		return &domain.TokenClaims{
			UserID:   "user_dev",
			TenantID: "tenant_dev",
			Email:    "dev@example.com",
			IsAdmin:  false,
		}, nil
	case "admin":
		return &domain.TokenClaims{
			UserID:   "admin_dev",
			TenantID: "tenant_admin",
			Email:    "admin@example.com",
			IsAdmin:  true,
		}, nil
	default:
		return nil, domain.NewError(domain.ErrUnauthorized, "Invalid token", nil)
	}
}
