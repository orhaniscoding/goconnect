package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

func setupNetworkRouterBasic() (*gin.Engine, repository.NetworkRepository) {
	gin.SetMode(gin.TestMode)
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	h := NewNetworkHandler(networkService, membershipService)
	r := gin.New()
	RegisterNetworkRoutes(r, h)
	return r, networkRepo
}

func TestGetNetwork_FoundAndNotFound(t *testing.T) {
	r, nrepo := setupNetworkRouterBasic()
	// seed
	n := &domain.Network{ID: "net-get-1", TenantID: "t1", Name: "GetTest", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.9.0.0/24", CreatedBy: "user_dev", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := nrepo.Create(context.Background(), n); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	// success
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/networks/"+n.ID, nil)
	req.Header.Set("Authorization", "Bearer dev")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/does-not-exist", nil)
	req.Header.Set("Authorization", "Bearer dev")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found, got %d", w.Code)
	}
}
