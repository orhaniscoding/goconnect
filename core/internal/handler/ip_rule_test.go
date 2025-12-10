package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIPRuleHandler() (*IPRuleHandler, *service.IPRuleService) {
	repo := repository.NewInMemoryIPRuleRepository()
	svc := service.NewIPRuleService(repo)
	handler := NewIPRuleHandler(svc)
	return handler, svc
}

func TestIPRuleHandler_CreateIPRule(t *testing.T) {
	handler, _ := setupIPRuleHandler()

	tests := []struct {
		name           string
		body           CreateIPRuleRequest
		tenantID       string
		userID         string
		expectedStatus int
	}{
		{
			name: "create allow rule with CIDR",
			body: CreateIPRuleRequest{
				Type:        "allow",
				CIDR:        "192.168.1.0/24",
				Description: "Office network",
			},
			tenantID:       "tenant-1",
			userID:         "admin-1",
			expectedStatus: http.StatusCreated,
		},
		{
			name: "create deny rule with single IP",
			body: CreateIPRuleRequest{
				Type:        "deny",
				CIDR:        "10.0.0.100",
				Description: "Blocked IP",
			},
			tenantID:       "tenant-1",
			userID:         "admin-1",
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid CIDR format",
			body: CreateIPRuleRequest{
				Type: "allow",
				CIDR: "not-an-ip",
			},
			tenantID:       "tenant-1",
			userID:         "admin-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid rule type",
			body: CreateIPRuleRequest{
				Type: "invalid",
				CIDR: "192.168.1.0/24",
			},
			tenantID:       "tenant-1",
			userID:         "admin-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing tenant ID",
			body: CreateIPRuleRequest{
				Type: "allow",
				CIDR: "192.168.1.0/24",
			},
			tenantID:       "",
			userID:         "admin-1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", tt.tenantID)
			req.Header.Set("X-User-ID", tt.userID)

			rr := httptest.NewRecorder()
			handler.CreateIPRule(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}

	// Additional test for conflict/success case (in-memory repo allows duplicates)
	t.Run("create duplicate rule allowed in memory", func(t *testing.T) {
		dupHandler, _ := setupIPRuleHandler()

		// Create first rule
		body1, _ := json.Marshal(CreateIPRuleRequest{
			Type:        "allow",
			CIDR:        "172.16.0.0/24",
			Description: "First rule",
		})
		req1 := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules", bytes.NewReader(body1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Tenant-ID", "tenant-1")
		req1.Header.Set("X-User-ID", "admin-1")
		rr1 := httptest.NewRecorder()
		dupHandler.CreateIPRule(rr1, req1)
		require.Equal(t, http.StatusCreated, rr1.Code)

		// Create second rule with same CIDR (in-memory allows it, real DB might conflict)
		body2, _ := json.Marshal(CreateIPRuleRequest{
			Type:        "allow",
			CIDR:        "172.16.0.0/24",
			Description: "Second rule",
		})
		req2 := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Tenant-ID", "tenant-1")
		req2.Header.Set("X-User-ID", "admin-1")
		rr2 := httptest.NewRecorder()
		dupHandler.CreateIPRule(rr2, req2)
		// In-memory repo doesn't enforce uniqueness, so 201 is expected
		assert.Contains(t, []int{http.StatusCreated, http.StatusConflict}, rr2.Code)
	})
}

func TestIPRuleHandler_ListIPRules(t *testing.T) {
	handler, svc := setupIPRuleHandler()

	// Create some rules
	ctx := context.Background()
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "192.168.1.0/24",
	})
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeDeny,
		CIDR:     "10.0.0.100",
	})
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-2",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "172.16.0.0/16",
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")

		rr := httptest.NewRecorder()
		handler.ListIPRules(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var response IPRulesListResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Total != 2 {
			t.Errorf("expected 2 rules for tenant-1, got %d", response.Total)
		}
	})

	t.Run("Missing tenant ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules", nil)
		// No X-Tenant-ID header

		rr := httptest.NewRecorder()
		handler.ListIPRules(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Empty tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules", nil)
		req.Header.Set("X-Tenant-ID", "tenant-empty")

		rr := httptest.NewRecorder()
		handler.ListIPRules(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response IPRulesListResponse
		json.NewDecoder(rr.Body).Decode(&response)
		assert.Equal(t, 0, response.Total)
	})
}

func TestIPRuleHandler_DeleteIPRule(t *testing.T) {
	handler, svc := setupIPRuleHandler()

	// Create a rule
	ctx := context.Background()
	rule, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "192.168.1.0/24",
	})

	// Setup mux for path parameter
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule.ID, nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify rule is deleted
	rules, _ := svc.ListIPRules(ctx, "tenant-1")
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after deletion, got %d", len(rules))
	}
}

func TestIPRuleHandler_CheckIP(t *testing.T) {
	handler, svc := setupIPRuleHandler()

	ctx := context.Background()
	// Create allow rule for a subnet
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "192.168.1.0/24",
	})
	// Create deny rule for specific IP within the subnet
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeDeny,
		CIDR:     "192.168.1.100",
	})

	tests := []struct {
		name          string
		ip            string
		tenantID      string
		expectAllowed bool
	}{
		{
			name:          "allowed IP in subnet",
			ip:            "192.168.1.50",
			tenantID:      "tenant-1",
			expectAllowed: true,
		},
		{
			name:          "denied IP",
			ip:            "192.168.1.100",
			tenantID:      "tenant-1",
			expectAllowed: false,
		},
		{
			name:          "IP outside allowed subnet",
			ip:            "10.0.0.1",
			tenantID:      "tenant-1",
			expectAllowed: false,
		},
		{
			name:          "tenant with no rules",
			ip:            "10.0.0.1",
			tenantID:      "tenant-2",
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(CheckIPRequest{IP: tt.ip})
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules/check", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", tt.tenantID)

			rr := httptest.NewRecorder()
			handler.CheckIP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rr.Code)
			}

			var response CheckIPResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Allowed != tt.expectAllowed {
				t.Errorf("expected allowed=%v, got %v", tt.expectAllowed, response.Allowed)
			}
		})
	}

	t.Run("Missing IP address", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-1")

		rr := httptest.NewRecorder()
		handler.CheckIP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Missing tenant ID", func(t *testing.T) {
		body, _ := json.Marshal(CheckIPRequest{IP: "192.168.1.50"})
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// No X-Tenant-ID header

		rr := httptest.NewRecorder()
		handler.CheckIP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/ip-rules/check", bytes.NewReader([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-1")

		rr := httptest.NewRecorder()
		handler.CheckIP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestIPFilterMiddleware(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	svc := service.NewIPRuleService(repo)
	middleware := NewIPFilterMiddleware(svc)

	// Create rules
	ctx := context.Background()
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "192.168.1.0/24",
	})
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeDeny,
		CIDR:     "192.168.1.100",
	})

	// Handler that returns OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		tenantID       string
		remoteAddr     string
		xForwardedFor  string
		expectedStatus int
	}{
		{
			name:           "allowed IP via RemoteAddr",
			tenantID:       "tenant-1",
			remoteAddr:     "192.168.1.50:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "denied IP via RemoteAddr",
			tenantID:       "tenant-1",
			remoteAddr:     "192.168.1.100:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "allowed IP via X-Forwarded-For",
			tenantID:       "tenant-1",
			remoteAddr:     "127.0.0.1:12345",
			xForwardedFor:  "192.168.1.50",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "denied IP via X-Forwarded-For",
			tenantID:       "tenant-1",
			remoteAddr:     "127.0.0.1:12345",
			xForwardedFor:  "192.168.1.100",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no tenant - skip filtering",
			tenantID:       "",
			remoteAddr:     "192.168.1.100:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "tenant with no rules - allow",
			tenantID:       "tenant-2",
			remoteAddr:     "10.0.0.1:12345",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Tenant-ID", tt.tenantID)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			rr := httptest.NewRecorder()
			middleware.Middleware(okHandler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestIPRuleExpiration(t *testing.T) {
	repo := repository.NewInMemoryIPRuleRepository()
	svc := service.NewIPRuleService(repo)

	ctx := context.Background()
	expired := time.Now().Add(-1 * time.Hour)
	notExpired := time.Now().Add(1 * time.Hour)

	// Create expired rule
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "192.168.1.0/24",
		ExpiresAt: &expired,
	})

	// Create non-expired rule
	svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID:  "tenant-1",
		Type:      domain.IPRuleTypeAllow,
		CIDR:      "10.0.0.0/8",
		ExpiresAt: &notExpired,
	})

	// List should only return non-expired
	rules, _ := svc.ListIPRules(ctx, "tenant-1")
	if len(rules) != 1 {
		t.Errorf("expected 1 non-expired rule, got %d", len(rules))
	}

	// Cleanup expired
	count, _ := svc.CleanupExpired(ctx)
	if count != 1 {
		t.Errorf("expected 1 expired rule cleaned up, got %d", count)
	}
}

func TestIPRuleHandler_GetIPRule(t *testing.T) {
	handler, svc := setupIPRuleHandler()

	// Create a rule
	ctx := context.Background()
	rule, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
		TenantID: "tenant-1",
		Type:     domain.IPRuleTypeAllow,
		CIDR:     "192.168.1.0/24",
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules/"+rule.ID, nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		req.SetPathValue("id", rule.ID)

		rr := httptest.NewRecorder()
		handler.GetIPRule(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Missing ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules/", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		// No path value set

		rr := httptest.NewRecorder()
		handler.GetIPRule(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules/nonexistent", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		req.SetPathValue("id", "nonexistent")

		rr := httptest.NewRecorder()
		handler.GetIPRule(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})

	t.Run("Wrong Tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/ip-rules/"+rule.ID, nil)
		req.Header.Set("X-Tenant-ID", "wrong-tenant")
		req.SetPathValue("id", rule.ID)

		rr := httptest.NewRecorder()
		handler.GetIPRule(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for wrong tenant, got %d", rr.Code)
		}
	})
}

// ==================== DeleteIPRule Comprehensive Tests ====================

func TestIPRuleHandler_DeleteIPRule_Comprehensive(t *testing.T) {
	t.Run("Success - Delete existing rule", func(t *testing.T) {
		handler, svc := setupIPRuleHandler()

		ctx := context.Background()
		rule, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
			TenantID: "tenant-1",
			Type:     domain.IPRuleTypeAllow,
			CIDR:     "192.168.2.0/24",
		})

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule.ID, nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
		}

		// Verify rule is deleted
		rules, _ := svc.ListIPRules(ctx, "tenant-1")
		for _, r := range rules {
			if r.ID == rule.ID {
				t.Errorf("rule should be deleted but still exists")
			}
		}
	})

	t.Run("Missing ID", func(t *testing.T) {
		handler, _ := setupIPRuleHandler()

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")
		// No path value set

		rr := httptest.NewRecorder()
		handler.DeleteIPRule(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		handler, _ := setupIPRuleHandler()

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/nonexistent-rule", nil)
		req.Header.Set("X-Tenant-ID", "tenant-1")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})

	t.Run("Wrong Tenant - Cannot Delete Other's Rule", func(t *testing.T) {
		handler, svc := setupIPRuleHandler()

		ctx := context.Background()
		rule, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
			TenantID: "tenant-1",
			Type:     domain.IPRuleTypeAllow,
			CIDR:     "192.168.3.0/24",
		})

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule.ID, nil)
		req.Header.Set("X-Tenant-ID", "tenant-2") // Different tenant

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for wrong tenant, got %d", rr.Code)
		}

		// Verify rule still exists for original tenant
		rules, _ := svc.ListIPRules(ctx, "tenant-1")
		found := false
		for _, r := range rules {
			if r.ID == rule.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rule should not be deleted by wrong tenant")
		}
	})

	t.Run("Delete Multiple Rules Sequentially", func(t *testing.T) {
		handler, svc := setupIPRuleHandler()

		ctx := context.Background()
		rule1, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
			TenantID: "tenant-1",
			Type:     domain.IPRuleTypeAllow,
			CIDR:     "192.168.10.0/24",
		})
		rule2, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
			TenantID: "tenant-1",
			Type:     domain.IPRuleTypeDeny,
			CIDR:     "192.168.20.0/24",
		})

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

		// Delete first rule
		req1 := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule1.ID, nil)
		req1.Header.Set("X-Tenant-ID", "tenant-1")
		rr1 := httptest.NewRecorder()
		mux.ServeHTTP(rr1, req1)
		if rr1.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for first delete, got %d", rr1.Code)
		}

		// Delete second rule
		req2 := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule2.ID, nil)
		req2.Header.Set("X-Tenant-ID", "tenant-1")
		rr2 := httptest.NewRecorder()
		mux.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for second delete, got %d", rr2.Code)
		}

		// Verify all rules deleted
		rules, _ := svc.ListIPRules(ctx, "tenant-1")
		if len(rules) != 0 {
			t.Errorf("expected 0 rules, got %d", len(rules))
		}
	})

	t.Run("Idempotent Delete", func(t *testing.T) {
		handler, svc := setupIPRuleHandler()

		ctx := context.Background()
		rule, _ := svc.CreateIPRule(ctx, domain.CreateIPRuleRequest{
			TenantID: "tenant-1",
			Type:     domain.IPRuleTypeAllow,
			CIDR:     "192.168.50.0/24",
		})

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /v1/admin/ip-rules/{id}", handler.DeleteIPRule)

		// First delete
		req1 := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule.ID, nil)
		req1.Header.Set("X-Tenant-ID", "tenant-1")
		rr1 := httptest.NewRecorder()
		mux.ServeHTTP(rr1, req1)
		if rr1.Code != http.StatusNoContent {
			t.Errorf("expected status 204 for first delete, got %d", rr1.Code)
		}

		// Second delete (should return not found or no content depending on implementation)
		req2 := httptest.NewRequest(http.MethodDelete, "/v1/admin/ip-rules/"+rule.ID, nil)
		req2.Header.Set("X-Tenant-ID", "tenant-1")
		rr2 := httptest.NewRecorder()
		mux.ServeHTTP(rr2, req2)
		// Either 204 (idempotent) or 404 (strict) is acceptable
		if rr2.Code != http.StatusNoContent && rr2.Code != http.StatusNotFound {
			t.Errorf("expected status 204 or 404, got %d", rr2.Code)
		}
	})
}
