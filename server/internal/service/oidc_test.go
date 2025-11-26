package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== OIDC SERVICE TESTS ====================

func TestNewOIDCService_NoConfig(t *testing.T) {
	t.Run("Returns Nil When Not Configured", func(t *testing.T) {
		// Clear all OIDC env vars
		os.Unsetenv("OIDC_ISSUER")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Unsetenv("OIDC_CLIENT_SECRET")
		os.Unsetenv("OIDC_REDIRECT_URL")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, err, "Should not return error when OIDC not configured")
		assert.Nil(t, svc, "Should return nil service when OIDC not configured")
	})
}

func TestNewOIDCService_PartialConfig(t *testing.T) {
	t.Run("Returns Nil When Partially Configured - Missing Issuer", func(t *testing.T) {
		os.Unsetenv("OIDC_ISSUER")
		os.Setenv("OIDC_CLIENT_ID", "test-client")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")
		defer func() {
			os.Unsetenv("OIDC_CLIENT_ID")
			os.Unsetenv("OIDC_CLIENT_SECRET")
			os.Unsetenv("OIDC_REDIRECT_URL")
		}()

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, err)
		assert.Nil(t, svc)
	})

	t.Run("Returns Nil When Partially Configured - Missing ClientID", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://example.com")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")
		defer func() {
			os.Unsetenv("OIDC_ISSUER")
			os.Unsetenv("OIDC_CLIENT_SECRET")
			os.Unsetenv("OIDC_REDIRECT_URL")
		}()

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, err)
		assert.Nil(t, svc)
	})

	t.Run("Returns Nil When Partially Configured - Missing Secret", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://example.com")
		os.Setenv("OIDC_CLIENT_ID", "test-client")
		os.Unsetenv("OIDC_CLIENT_SECRET")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")
		defer func() {
			os.Unsetenv("OIDC_ISSUER")
			os.Unsetenv("OIDC_CLIENT_ID")
			os.Unsetenv("OIDC_REDIRECT_URL")
		}()

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, err)
		assert.Nil(t, svc)
	})

	t.Run("Returns Nil When Partially Configured - Missing RedirectURL", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://example.com")
		os.Setenv("OIDC_CLIENT_ID", "test-client")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Unsetenv("OIDC_REDIRECT_URL")
		defer func() {
			os.Unsetenv("OIDC_ISSUER")
			os.Unsetenv("OIDC_CLIENT_ID")
			os.Unsetenv("OIDC_CLIENT_SECRET")
		}()

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, err)
		assert.Nil(t, svc)
	})
}

func TestUserInfo_Struct(t *testing.T) {
	t.Run("UserInfo Fields", func(t *testing.T) {
		info := &UserInfo{
			Email: "user@example.com",
			Name:  "Test User",
			Sub:   "user-123",
		}

		assert.Equal(t, "user@example.com", info.Email)
		assert.Equal(t, "Test User", info.Name)
		assert.Equal(t, "user-123", info.Sub)
	})
}
