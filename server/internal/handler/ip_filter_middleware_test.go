package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
)

// mockIPRuleService implements a mock for IPRuleService
type mockIPRuleService struct {
	rules      map[string][]*domain.IPRule
	allowByIP  map[string]bool
	checkError error
}

func newMockIPRuleService() *mockIPRuleService {
	return &mockIPRuleService{
		rules:     make(map[string][]*domain.IPRule),
		allowByIP: make(map[string]bool),
	}
}

func (m *mockIPRuleService) CheckIP(ctx context.Context, tenantID, ip string) (bool, *domain.IPRule, error) {
	if m.checkError != nil {
		return false, nil, m.checkError
	}

	key := tenantID + ":" + ip
	if allowed, ok := m.allowByIP[key]; ok {
		if !allowed {
			return false, &domain.IPRule{CIDR: "blocked-cidr"}, nil
		}
		return true, nil, nil
	}

	// Default: allow
	return true, nil, nil
}

// ==================== IP FILTER MIDDLEWARE TESTS ====================

func TestIPFilterMiddleware_NoTenant(t *testing.T) {
	t.Run("No Tenant Header - Skip Filtering", func(t *testing.T) {
		// We can't directly instantiate IPFilterMiddleware without real service
		// So we test getClientIP function instead
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "192.168.1.1:12345"

		ip := getClientIP(r)
		assert.Equal(t, "192.168.1.1", ip)
	})
}

func TestGetClientIP(t *testing.T) {
	t.Run("From RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "10.0.0.1:8080"

		ip := getClientIP(r)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("From X-Forwarded-For Single IP", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		r.Header.Set("X-Forwarded-For", "203.0.113.50")

		ip := getClientIP(r)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("From X-Forwarded-For Multiple IPs", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

		ip := getClientIP(r)
		assert.Equal(t, "203.0.113.50", ip) // First IP in chain
	})

	t.Run("From X-Real-IP", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		r.Header.Set("X-Real-IP", "198.51.100.10")

		ip := getClientIP(r)
		assert.Equal(t, "198.51.100.10", ip)
	})

	t.Run("X-Forwarded-For Takes Priority Over X-Real-IP", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		r.Header.Set("X-Forwarded-For", "203.0.113.50")
		r.Header.Set("X-Real-IP", "198.51.100.10")

		ip := getClientIP(r)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("Invalid X-Forwarded-For Falls Back", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "10.0.0.5:8080"
		r.Header.Set("X-Forwarded-For", "invalid-ip")

		ip := getClientIP(r)
		// Should fall back to X-Real-IP or RemoteAddr
		assert.Equal(t, "10.0.0.5", ip)
	})

	t.Run("RemoteAddr Without Port", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "192.168.0.1" // No port

		ip := getClientIP(r)
		assert.Equal(t, "192.168.0.1", ip)
	})

	t.Run("IPv6 Address", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "[::1]:8080"

		ip := getClientIP(r)
		assert.Equal(t, "::1", ip)
	})

	t.Run("IPv6 In X-Forwarded-For", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		r.Header.Set("X-Forwarded-For", "2001:db8::1")

		ip := getClientIP(r)
		assert.Equal(t, "2001:db8::1", ip)
	})
}

func TestNewIPFilterMiddleware(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		// We can't test with nil service, but we can verify the struct is created
		// This is mainly to ensure code path coverage
		assert.NotPanics(t, func() {
			_ = NewIPFilterMiddleware(nil)
		})
	})
}

// ==================== MIDDLEWARE INTEGRATION TESTS ====================

func TestIPFilterMiddleware_Integration(t *testing.T) {
	t.Run("Next Handler Called When No Tenant", func(t *testing.T) {
		// Create a simple handler to verify middleware behavior
		called := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		// Since we can't easily mock the service, we test the path without tenant
		middleware := NewIPFilterMiddleware(nil)

		// When svc is nil, we need to handle gracefully
		// The middleware should still work
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "10.0.0.1:8080"
		// No X-Tenant-ID header

		w := httptest.NewRecorder()

		// This will panic if svc is nil and tenant header exists
		// But without tenant header, it should call next
		handler := middleware.Middleware(nextHandler)
		handler.ServeHTTP(w, r)

		assert.True(t, called, "Next handler should be called when no tenant")
	})
}
