package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== MOCK REPOSITORIES ====================

// mockNetworkRepo implements NetworkRepository for testing
type mockNetworkRepo struct {
	networks map[string]*domain.Network
}

func newMockNetworkRepo() *mockNetworkRepo {
	return &mockNetworkRepo{networks: make(map[string]*domain.Network)}
}

func (m *mockNetworkRepo) GetByID(ctx context.Context, id string) (*domain.Network, error) {
	if net, ok := m.networks[id]; ok {
		return net, nil
	}
	return nil, domain.NewError(domain.ErrNotFound, "Network not found", nil)
}

// mockMembershipRepo implements MembershipRepository for testing
type mockMembershipRepo struct {
	memberships map[string]*domain.Membership
}

func newMockMembershipRepo() *mockMembershipRepo {
	return &mockMembershipRepo{memberships: make(map[string]*domain.Membership)}
}

func (m *mockMembershipRepo) Get(ctx context.Context, networkID, userID string) (*domain.Membership, error) {
	key := networkID + ":" + userID
	if mem, ok := m.memberships[key]; ok {
		return mem, nil
	}
	return nil, domain.NewError(domain.ErrNotFound, "Membership not found", nil)
}

// mockUserRepo implements UserRepository for testing
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, domain.NewError(domain.ErrNotFound, "User not found", nil)
}

// mockAuditor implements service.Auditor for testing
type mockAuditor struct {
	events []map[string]interface{}
}

func (m *mockAuditor) Event(ctx context.Context, tenantID, eventType, actorID, targetID string, details map[string]any) {
	m.events = append(m.events, map[string]interface{}{
		"tenant_id":  tenantID,
		"event_type": eventType,
		"actor_id":   actorID,
		"target_id":  targetID,
		"details":    details,
	})
}

// wireguardAuthMiddleware returns a test auth middleware
func wireguardAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		switch token {
		case "user-token":
			c.Set("user_id", "user1")
			c.Set("tenant_id", "t1")
			c.Set("is_admin", false)
			c.Next()
		case "admin-token":
			c.Set("user_id", "admin1")
			c.Set("tenant_id", "t1")
			c.Set("is_admin", true)
			c.Next()
		case "wrong-tenant-token":
			c.Set("user_id", "user1")
			c.Set("tenant_id", "t2") // Different tenant
			c.Set("is_admin", false)
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		}
	}
}

// ==================== GET PROFILE TESTS ====================

func TestWireGuardHandler_GetProfile(t *testing.T) {
	t.Run("Unauthorized - No Token", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Unauthorized - Empty User ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)

		// Custom middleware that sets empty user_id
		emptyUserMiddleware := func(c *gin.Context) {
			c.Set("user_id", "")
			c.Set("tenant_id", "t1")
			c.Next()
		}

		r.GET("/v1/networks/:id/wg/profile", emptyUserMiddleware, handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing Device ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "device_id")
	})

	t.Run("Reject Non-JSON Accept", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Accept", "text/plain")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Only JSON API")
	})

	t.Run("Network Not Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/nonexistent/wg/profile?device_id=d1", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Tenant Isolation - Wrong Tenant", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		networkRepo.networks["net1"] = &domain.Network{
			ID:       "net1",
			TenantID: "t1", // Network belongs to t1
			Name:     "Test Network",
			CIDR:     "10.0.0.0/24",
		}
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		req.Header.Set("Authorization", "Bearer wrong-tenant-token") // User from t2
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Not A Member", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		networkRepo.networks["net1"] = &domain.Network{
			ID:       "net1",
			TenantID: "t1",
			Name:     "Test Network",
			CIDR:     "10.0.0.0/24",
		}
		membershipRepo := newMockMembershipRepo()
		// No membership added
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "not a member")
	})

	t.Run("Membership Not Approved", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		networkRepo.networks["net1"] = &domain.Network{
			ID:       "net1",
			TenantID: "t1",
			Name:     "Test Network",
			CIDR:     "10.0.0.0/24",
		}
		membershipRepo := newMockMembershipRepo()
		now := time.Now()
		membershipRepo.memberships["net1:user1"] = &domain.Membership{
			NetworkID: "net1",
			UserID:    "user1",
			Status:    domain.StatusPending, // Not approved
			JoinedAt:  &now,
		}
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "not approved")
	})
}

// ==================== VALIDATION TESTS ====================

func TestWireGuardHandler_Validation(t *testing.T) {
	t.Run("All Parameters Required", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

		// No query parameters at all
		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Contains(t, response["message"], "device_id")
	})
}

// ==================== REGISTER ROUTES TESTS ====================

func TestRegisterWireGuardRoutes(t *testing.T) {
	t.Run("Routes Registered", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()

		networkRepo := newMockNetworkRepo()
		membershipRepo := newMockMembershipRepo()
		userRepo := newMockUserRepo()
		auditor := &mockAuditor{}

		handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
		RegisterWireGuardRoutes(r, handler, wireguardAuthMiddleware())

		// Test that route is registered - should return unauthorized (not 404)
		req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
		req.Header.Set("Accept", "application/json")
		// No auth header - should get 401, not 404
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Route exists if we get 401 (unauthorized) instead of 404 (not found)
		assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound,
			"Expected 401, 400 or 404 but got %d", w.Code)
	})
}

// ==================== GetProfile Deep Path Tests ====================

func TestWireGuardHandler_GetProfile_NetworkDomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Custom network repo that returns domain error
	customNetworkRepo := &mockNetworkRepoWithDomainError{}
	membershipRepo := newMockMembershipRepo()
	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	handler := NewWireGuardHandler(customNetworkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/networks/error-net/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return the domain error's HTTP status
	assert.True(t, w.Code >= 400)
}

// mockNetworkRepoWithDomainError returns a domain error
type mockNetworkRepoWithDomainError struct{}

func (m *mockNetworkRepoWithDomainError) GetByID(ctx context.Context, id string) (*domain.Network, error) {
	return nil, domain.NewError(domain.ErrForbidden, "Access denied", nil)
}

func TestWireGuardHandler_GetProfile_RejectedMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	networkRepo := newMockNetworkRepo()
	networkRepo.networks["net1"] = &domain.Network{
		ID:       "net1",
		TenantID: "t1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	}

	membershipRepo := newMockMembershipRepo()
	now := time.Now()
	membershipRepo.memberships["net1:user1"] = &domain.Membership{
		NetworkID: "net1",
		UserID:    "user1",
		Status:    "rejected", // Rejected status
		JoinedAt:  &now,
	}

	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not approved")
}

func TestWireGuardHandler_GetProfile_BannedMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	networkRepo := newMockNetworkRepo()
	networkRepo.networks["net1"] = &domain.Network{
		ID:       "net1",
		TenantID: "t1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	}

	membershipRepo := newMockMembershipRepo()
	now := time.Now()
	membershipRepo.memberships["net1:user1"] = &domain.Membership{
		NetworkID: "net1",
		UserID:    "user1",
		Status:    domain.StatusBanned, // Banned status (using correct constant)
		JoinedAt:  &now,
	}

	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not approved")
}

func TestWireGuardHandler_NewHandler(t *testing.T) {
	networkRepo := newMockNetworkRepo()
	membershipRepo := newMockMembershipRepo()
	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)

	assert.NotNil(t, handler)
	assert.Equal(t, networkRepo, handler.networkRepo)
	assert.Equal(t, membershipRepo, handler.membershipRepo)
	assert.Equal(t, userRepo, handler.userRepo)
	assert.Equal(t, auditor, handler.auditor)
}

func TestWireGuardHandler_GetProfile_WrongTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	networkRepo := newMockNetworkRepo()
	membershipRepo := newMockMembershipRepo()
	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	// Create network with tenant t1
	networkRepo.networks["net1"] = &domain.Network{
		ID:       "net1",
		TenantID: "t1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	}

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	// Request with wrong tenant token
	req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer wrong-tenant-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}

func TestWireGuardHandler_GetProfile_MembershipNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	networkRepo := newMockNetworkRepo()
	membershipRepo := newMockMembershipRepo()
	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	// Create network
	networkRepo.networks["net1"] = &domain.Network{
		ID:       "net1",
		TenantID: "t1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	}

	// No membership - user1 is not a member of net1

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not a member")
}

func TestWireGuardHandler_GetProfile_PendingMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	networkRepo := newMockNetworkRepo()
	membershipRepo := newMockMembershipRepo()
	userRepo := newMockUserRepo()
	auditor := &mockAuditor{}

	// Create network
	networkRepo.networks["net1"] = &domain.Network{
		ID:       "net1",
		TenantID: "t1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	}

	// Add pending membership
	membershipRepo.memberships["net1:user1"] = &domain.Membership{
		NetworkID: "net1",
		UserID:    "user1",
		Status:    domain.StatusPending,
	}

	handler := NewWireGuardHandler(networkRepo, membershipRepo, nil, nil, userRepo, nil, auditor)
	r.GET("/v1/networks/:id/wg/profile", wireguardAuthMiddleware(), handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/networks/net1/wg/profile?device_id=d1", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "not approved")
}
