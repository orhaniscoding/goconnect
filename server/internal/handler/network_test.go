package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
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
	networkHandler := NewNetworkHandler(networkService, membershipService)

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
	httpReq.Header.Set("Authorization", "Bearer dev")
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
	httpReq.Header.Set("Authorization", "Bearer dev")
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
	httpReq.Header.Set("Authorization", "Bearer dev")
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
		TenantID:   "default",
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
	httpReq.Header.Set("Authorization", "Bearer dev")

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
	httpReq.Header.Set("Authorization", "Bearer admin") // admin token

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestListNetworks_NonAdminAll(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/v1/networks?visibility=all", nil)
	httpReq.Header.Set("Authorization", "Bearer dev") // non-admin token

	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}
