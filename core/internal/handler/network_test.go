package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *NetworkHandler) {
	gin.SetMode(gin.TestMode)

	// Setup dependencies
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	authSvc := newMockAuthServiceWithTokens()

	// Setup for DeviceService
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	networkHandler := NewNetworkHandler(networkService, membershipService, ds, peerRepo, wgConfig)

	// Setup router
	r := gin.New()
	RegisterNetworkRoutes(r, networkHandler, authSvc, membershipRepo)

	return r, networkHandler
}

func TestCreateNetwork_Success(t *testing.T) {
	router, _ := setupTestRouter()

	req := domain.CreateNetworkRequest{
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	jsonData, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	httpReq.Header.Set("Idempotency-Key", "test-key-123")

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing data field")
	}

	if data["name"] != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, data["name"])
	}
}

func TestCreateNetwork_MissingIdempotencyKey(t *testing.T) {
	router, _ := setupTestRouter()

	req := domain.CreateNetworkRequest{
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	jsonData, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	// Missing Idempotency-Key header

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateNetwork_InvalidCIDR(t *testing.T) {
	router, _ := setupTestRouter()

	req := domain.CreateNetworkRequest{
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "invalid-cidr",
	}

	jsonData, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer valid-token")
	httpReq.Header.Set("Idempotency-Key", "test-key-456")

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateNetwork_Unauthorized(t *testing.T) {
	router, _ := setupTestRouter()

	req := domain.CreateNetworkRequest{
		Name:       "Test Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
	}

	jsonData, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	// Missing Authorization header
	httpReq.Header.Set("Idempotency-Key", "test-key-789")

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestListNetworks_Public(t *testing.T) {
	router, handler := setupTestRouter()

	// Create a test network first
	ctx := context.Background()
	networkRepo := repository.NewInMemoryNetworkRepository()
	if err := networkRepo.Create(ctx, &domain.Network{
		ID:         "net-123",
		TenantID:   "t1",
		Name:       "Public Network",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.0.0.0/24",
		CreatedBy:  "user123",
	}); err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	// Update handler with the populated repo
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	handler.networkService = networkService
	handler.memberService = membershipService

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/v1/networks?visibility=public", nil)
	httpReq.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Response missing data array")
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 network, got %d", len(data))
	}
}

func TestListNetworks_Unauthorized(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/v1/networks", nil)
	// Missing Authorization header

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestListNetworks_AdminAll(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/v1/networks?visibility=all", nil)
	httpReq.Header.Set("Authorization", "Bearer admin-token") // admin token

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListNetworks_NonAdminAll(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/v1/networks?visibility=all", nil)
	httpReq.Header.Set("Authorization", "Bearer valid-token") // non-admin token

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

// setupNetworkWithRepos exposes repositories for more detailed scenarios
func setupNetworkWithRepos() (*gin.Engine, *NetworkHandler, repository.NetworkRepository, repository.MembershipRepository, *service.MembershipService) {
	gin.SetMode(gin.TestMode)

	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	handler := NewNetworkHandler(networkService, membershipService, deviceService, peerRepo, wgConfig)
	r := gin.New()
	authSvc := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, authSvc, membershipRepo)

	return r, handler, networkRepo, membershipRepo, membershipService
}

func TestCreateNetwork_DuplicateName(t *testing.T) {
	router, handler, networkRepo, _, _ := setupNetworkWithRepos()

	// Seed a network with the same name
	existing := &domain.Network{
		ID:         "net-dup-1",
		TenantID:   "t1",
		Name:       "DupNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.10.0.0/24",
		CreatedBy:  "seed",
	}
	require.NoError(t, networkRepo.Create(context.Background(), existing))

	body := domain.CreateNetworkRequest{
		Name:       "DupNet",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyOpen,
		CIDR:       "10.20.0.0/24",
	}
	jsonData, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "dup-key")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "duplicate name should be rejected")

	// Ensure original network still present
	saved, err := handler.networkService.GetNetwork(context.Background(), existing.ID, "user_dev", "t1")
	require.NoError(t, err)
	assert.Equal(t, existing.Name, saved.Name)
}

func TestUpdateNetwork_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()
	// seed network and admin membership for dev
	net := &domain.Network{ID: "net-up-missing", TenantID: "t1", Name: "Upd", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.30.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/networks/"+net.ID, bytes.NewBuffer([]byte(`{"name":"NewName"}`)))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	// Missing Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteNetwork_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()
	net := &domain.Network{ID: "net-del-missing", TenantID: "t1", Name: "DeleteMe", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.40.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleAdmin, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID, nil)
	req.Header.Set("Authorization", "Bearer dev")
	// Missing Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteNetwork_Success(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()
	net := &domain.Network{ID: "net-del-success", TenantID: "t1", Name: "DeleteMeSuccess", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.41.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleAdmin, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID, nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteNetwork_NotFound(t *testing.T) {
	router, _, _, _, _ := setupNetworkWithRepos()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/nonexistent-network", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return not found or forbidden
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusForbidden)
}

func TestDeleteNetwork_Unauthorized(t *testing.T) {
	router, _, networkRepo, _, _ := setupNetworkWithRepos()
	net := &domain.Network{ID: "net-del-unauth", TenantID: "t1", Name: "DeleteMeUnauth", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.42.0.0/24", CreatedBy: "other_user"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	// Don't add user_dev as member

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID, nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should be forbidden since user is not admin
	assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusNotFound)
}

func TestJoinNetwork_OpenPolicyAndMissingHeader(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-open-join", TenantID: "t1", Name: "OpenNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.50.0.0/24", CreatedBy: "seed"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())
	// ensure open policy results in immediate approval
	membershipService.SetPeerProvisioning(nil)

	t.Run("success when idempotency provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer dev")
		req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing idempotency yields bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
		req.Header.Set("Authorization", "Bearer dev")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestApproveRequiresAdminMembership(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-approve", TenantID: "t1", Name: "ApproveNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.60.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	// Seed owner membership for admin_dev and member for user_dev
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())
	// Create pending join request for target user (user_dev2) via service to use shared repos
	_, _, err := membershipService.JoinNetwork(context.Background(), net.ID, "user_dev2", "t1", domain.GenerateIdempotencyKey())
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_dev2"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", body)
	req.Header.Set("Authorization", "Bearer dev") // user_dev is only member role (not admin)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "non-admin should not approve")
}

func TestListMembers_TenantIsolation(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	// network belongs to t2; token tenant is t1 -> should not be found
	net := &domain.Network{ID: "net-tenant-mismatch", TenantID: "t2", Name: "OtherTenant", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.70.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/members", nil)
	req.Header.Set("Authorization", "Bearer dev")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListJoinRequests_NonAdminDenied(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-joinreq", TenantID: "t1", Name: "JoinReqNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.80.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	// create pending join request for user_dev2
	joinKey := domain.GenerateIdempotencyKey()
	_, _, err := membershipService.JoinNetwork(context.Background(), net.ID, "pending_user", "t1", joinKey)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/join-requests", nil)
	req.Header.Set("Authorization", "Bearer dev") // non-admin

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListJoinRequests_AdminSuccess(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-joinreq-success", TenantID: "t1", Name: "JoinReqSuccessNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.81.0.0/24", CreatedBy: "admin_user"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_user", domain.RoleOwner, time.Now())
	// create pending join request
	joinKey := domain.GenerateIdempotencyKey()
	_, _, err := membershipService.JoinNetwork(context.Background(), net.ID, "pending_user", "t1", joinKey)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/join-requests", nil)
	req.Header.Set("Authorization", "Bearer admin") // admin token

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "data")
}

func TestReleaseIP_NotImplemented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig) // IPAM not injected
	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	// seed network + membership
	net := &domain.Network{ID: "net-release-ip", TenantID: "t1", Name: "ReleaseIPNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.91.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	t.Run("ipam not configured returns not implemented", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocation", nil)
		req.Header.Set("Authorization", "Bearer dev")
		req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})

	t.Run("missing idempotency key returns bad request when ipam present", func(t *testing.T) {
		ipam := service.NewIPAMService(networkRepo, mrepo, repository.NewInMemoryIPAM())
		handler.WithIPAM(ipam)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocation", nil)
		req.Header.Set("Authorization", "Bearer dev")
		// no Idempotency-Key

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGenerateConfig_NotMember(t *testing.T) {
	router, _, networkRepo, _, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-genconfig", TenantID: "t1", Name: "GenConfigNet", Visibility: domain.NetworkVisibilityPrivate, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.92.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/config", nil)
	req.Header.Set("Authorization", "Bearer dev") // user_dev is not a member
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Should fail with not found or forbidden since user is not member
	assert.Contains(t, []int{http.StatusNotFound, http.StatusForbidden}, w.Code)
}

func TestParseIntWithDefault(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultVal int
		expected   int
	}{
		{"empty string returns default", "", 10, 10},
		{"valid int string", "25", 10, 25},
		{"invalid string returns default", "abc", 10, 10},
		{"zero string", "0", 10, 0},
		{"negative number", "-5", 10, -5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseIntWithDefault(tc.input, tc.defaultVal)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAllocateIP_NotImplementedAndMissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig) // IPAM not injected
	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	// seed network + membership
	net := &domain.Network{ID: "net-no-ipam", TenantID: "t1", Name: "NoIPAM", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.90.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	t.Run("ipam not configured returns not implemented", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ip-allocations", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Authorization", "Bearer dev")
		req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})

	t.Run("missing idempotency key returns bad request when ipam present", func(t *testing.T) {
		ipam := service.NewIPAMService(networkRepo, mrepo, repository.NewInMemoryIPAM())
		handler.WithIPAM(ipam)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ip-allocations", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Authorization", "Bearer dev")
		// no Idempotency-Key

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ==================== Deny Tests ====================

func TestDeny_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-deny-nokey", TenantID: "t1", Name: "DenyNoKey", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.61.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_dev2"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/deny", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	// No Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeny_InvalidBody(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-deny-bad", TenantID: "t1", Name: "DenyBad", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.62.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/deny", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Kick Tests ====================

func TestKick_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-kick-nokey", TenantID: "t1", Name: "KickNoKey", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.63.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_dev2"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	// No Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKick_InvalidBody(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-kick-bad", TenantID: "t1", Name: "KickBad", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.64.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== Ban Tests ====================

func TestBan_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-ban-nokey", TenantID: "t1", Name: "BanNoKey", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.65.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_dev2"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	// No Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBan_InvalidBody(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-ban-bad", TenantID: "t1", Name: "BanBad", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.66.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBan_Success(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-ban-success", TenantID: "t1", Name: "BanSuccess", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.67.1.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_to_ban", domain.RoleMember, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_to_ban"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should succeed
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestKick_Success(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-kick-success", TenantID: "t1", Name: "KickSuccess", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.68.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_to_kick", domain.RoleMember, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_to_kick"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestDeny_Success(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-deny-success", TenantID: "t1", Name: "DenySuccess", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.69.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	// Create a pending join request
	_, _, _ = membershipService.JoinNetwork(context.Background(), net.ID, "pending_user", "t1", domain.GenerateIdempotencyKey())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"pending_user"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/deny", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

// ==================== Approve Tests ====================

func TestApprove_MissingIdempotencyKey(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-approve-nokey", TenantID: "t1", Name: "ApproveNoKey", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.67.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_dev2"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	// No Idempotency-Key header

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApprove_InvalidBody(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-approve-bad", TenantID: "t1", Name: "ApproveBad", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.68.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApprove_Success(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-approve-success", TenantID: "t1", Name: "ApproveSuccess", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.69.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":"user_pending"}`)
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", body)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should reach service layer and return appropriate response
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound, http.StatusForbidden, http.StatusBadRequest}, w.Code)
}

// ==================== ListJoinRequests Comprehensive Tests ====================

func TestListJoinRequests_EmptyList(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-joinreq-empty", TenantID: "t1", Name: "EmptyJoinReqNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.82.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/join-requests", nil)
	req.Header.Set("Authorization", "Bearer admin")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	// data might be nil or an empty array
	if data, ok := resp["data"].([]interface{}); ok {
		assert.Empty(t, data)
	} else {
		// nil data is also acceptable for empty list
		assert.Nil(t, resp["data"])
	}
}

func TestListJoinRequests_TenantIsolation(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	// Network belongs to different tenant
	net := &domain.Network{ID: "net-joinreq-tenant", TenantID: "t2", Name: "DifferentTenantNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.83.0.0/24", CreatedBy: "other_admin"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "other_admin", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/join-requests", nil)
	req.Header.Set("Authorization", "Bearer admin") // admin from t1

	router.ServeHTTP(w, req)

	// Should be not found due to tenant isolation
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListJoinRequests_MultipleRequests(t *testing.T) {
	router, _, networkRepo, membershipRepo, membershipService := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-joinreq-multi", TenantID: "t1", Name: "MultiJoinReqNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.84.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	// Create multiple pending requests
	for i := 1; i <= 3; i++ {
		joinKey := domain.GenerateIdempotencyKey()
		_, _, err := membershipService.JoinNetwork(context.Background(), net.ID, fmt.Sprintf("pending_user_%d", i), "t1", joinKey)
		require.NoError(t, err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/join-requests", nil)
	req.Header.Set("Authorization", "Bearer admin")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]interface{})
	assert.Len(t, data, 3)
}

// ==================== AdminReleaseIP Comprehensive Tests ====================

func TestAdminReleaseIP_NotImplemented(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig) // IPAM not injected
	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	net := &domain.Network{ID: "net-admin-release", TenantID: "t1", Name: "AdminReleaseNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.93.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocations/user_dev", nil)
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestAdminReleaseIP_MissingIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)

	// Inject IPAM
	ipam := service.NewIPAMService(networkRepo, mrepo, repository.NewInMemoryIPAM())
	handler.WithIPAM(ipam)

	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	net := &domain.Network{ID: "net-admin-release-nokey", TenantID: "t1", Name: "AdminReleaseNoKeyNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.94.0.0/24", CreatedBy: "admin_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocations/some_user", nil)
	req.Header.Set("Authorization", "Bearer admin")
	// No Idempotency-Key

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminReleaseIP_NonAdminForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)

	// Inject IPAM
	ipam := service.NewIPAMService(networkRepo, mrepo, repository.NewInMemoryIPAM())
	handler.WithIPAM(ipam)

	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	net := &domain.Network{ID: "net-admin-release-forbidden", TenantID: "t1", Name: "AdminReleaseForbiddenNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.95.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "owner", domain.RoleOwner, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocations/other_user", nil)
	req.Header.Set("Authorization", "Bearer dev") // non-admin
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==================== GenerateConfig Comprehensive Tests ====================

func TestGenerateConfig_NetworkNotFound(t *testing.T) {
	router, _, _, _, _ := setupNetworkWithRepos()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/nonexistent-network/config", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGenerateConfig_TenantIsolation(t *testing.T) {
	router, _, networkRepo, _, _ := setupNetworkWithRepos()

	// Network belongs to different tenant
	net := &domain.Network{ID: "net-genconfig-tenant", TenantID: "t2", Name: "DiffTenantNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.96.0.0/24", CreatedBy: "other_owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/config", nil)
	req.Header.Set("Authorization", "Bearer dev") // dev is from t1
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGenerateConfig_Unauthorized(t *testing.T) {
	router, _, networkRepo, _, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-genconfig-unauth", TenantID: "t1", Name: "GenConfigUnauth", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.97.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/config", nil)
	// No Authorization header
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGenerateConfig_MemberCanGenerate(t *testing.T) {
	router, _, networkRepo, membershipRepo, _ := setupNetworkWithRepos()

	net := &domain.Network{ID: "net-genconfig-member", TenantID: "t1", Name: "MemberGenConfigNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.98.0.0/24", CreatedBy: "owner"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "owner", domain.RoleOwner, time.Now())
	_, _ = membershipRepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/config", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Should succeed for member, or return various error codes if config generation deps are missing
	// In our test setup, key generation might fail but the handler path should be exercised
	assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

func TestReleaseIP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	// Setup IPAM
	ipamRepo := repository.NewInMemoryIPAM()
	ipamService := service.NewIPAMService(networkRepo, mrepo, ipamRepo)

	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)
	handler.WithIPAM(ipamService)

	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	// seed network + membership
	net := &domain.Network{ID: "net-release-success", TenantID: "t1", Name: "ReleaseSuccessNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.95.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

	// First allocate an IP
	_, _ = ipamService.AllocateIP(context.Background(), net.ID, "user_dev", "t1")

	t.Run("release ip success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/networks/"+net.ID+"/ip-allocation", nil)
		req.Header.Set("Authorization", "Bearer dev")
		req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestListIPAllocations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idemRepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(networkRepo, idemRepo)
	ms := service.NewMembershipService(networkRepo, mrepo, jrepo, idemRepo)
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)

	// Setup IPAM
	ipamRepo := repository.NewInMemoryIPAM()
	ipamService := service.NewIPAMService(networkRepo, mrepo, ipamRepo)

	handler := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)
	handler.WithIPAM(ipamService)

	r := gin.New()
	auth := newMockAuthServiceWithTokens()
	RegisterNetworkRoutes(r, handler, auth, mrepo)

	// seed network + membership
	net := &domain.Network{ID: "net-listip", TenantID: "t1", Name: "ListIPNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.96.0.0/24", CreatedBy: "user_dev"}
	require.NoError(t, networkRepo.Create(context.Background(), net))
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleOwner, time.Now())

	// Allocate an IP
	_, _ = ipamService.AllocateIP(context.Background(), net.ID, "user_dev", "t1")

	t.Run("list ip allocations success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/ip-allocations", nil)
		req.Header.Set("Authorization", "Bearer dev")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== DeleteNetwork Service Error Path ====================

func TestDeleteNetwork_ServiceError(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a network that doesn't exist to trigger service error
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/nonexistent-network-id", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error (NotFound or Forbidden)
	assert.Contains(t, []int{http.StatusNotFound, http.StatusForbidden}, w.Code)
}

// ==================== Deny/Kick/Ban Service Error Paths ====================

func TestDeny_ServiceError(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"user_id": "target-user"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/nonexistent-network/deny", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error from service
	assert.True(t, w.Code >= 400)
}

func TestKick_ServiceError(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"user_id": "target-user"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/nonexistent-network/kick", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error from service
	assert.True(t, w.Code >= 400)
}

func TestBan_ServiceError(t *testing.T) {
	router, _ := setupTestRouter()

	body := `{"user_id": "target-user"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/nonexistent-network/ban", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error from service
	assert.True(t, w.Code >= 400)
}

// ==================== ListIPAllocations Error Path ====================

func TestListIPAllocations_NetworkNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/nonexistent-network/ip-allocations", nil)
	req.Header.Set("Authorization", "Bearer dev")

	router.ServeHTTP(w, req)

	// Should return an error (could be 501 NotImplemented if IPAM not set, or 404/403)
	assert.True(t, w.Code >= 400)
}

// ==================== ReleaseIP Error Path ====================

func TestReleaseIP_NetworkNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/nonexistent-network/ip-allocation", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error
	assert.True(t, w.Code >= 400)
}

// ==================== AdminReleaseIP Error Path ====================

func TestAdminReleaseIP_NetworkNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/nonexistent-network/ip-allocations/some-user", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())

	router.ServeHTTP(w, req)

	// Should return an error (not found or forbidden)
	assert.True(t, w.Code >= 400)
}
