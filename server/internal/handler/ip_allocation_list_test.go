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

// setupAllocList creates network, membership and one allocation
func setupAllocList(t *testing.T) (*gin.Engine, string) {
    gin.SetMode(gin.TestMode)
    nrepo := repository.NewInMemoryNetworkRepository()
    irepo := repository.NewInMemoryIdempotencyRepository()
    mrepo := repository.NewInMemoryMembershipRepository()
    jrepo := repository.NewInMemoryJoinRequestRepository()
    iprepo := repository.NewInMemoryIPAM()
    ns := service.NewNetworkService(nrepo, irepo)
    ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
    ips := service.NewIPAMService(nrepo, mrepo, iprepo)
    h := NewNetworkHandler(ns, ms).WithIPAM(ips)
    r := gin.New()
    r.Use(RoleMiddleware(mrepo))
    RegisterNetworkRoutes(r, h)
    net := &domain.Network{ID: "net-list-1", TenantID: "t1", Name: "NetList", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.90.0.0/30", CreatedBy: "user_dev", CreatedAt: time.Now(), UpdatedAt: time.Now()}
    if err := nrepo.Create(context.Background(), net); err != nil { t.Fatalf("create network: %v", err) }
    // add membership & allocate
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())
    if _, err := ips.AllocateIP(context.Background(), net.ID, "user_dev"); err != nil { t.Fatalf("allocate: %v", err) }
    return r, net.ID
}

func TestListIPAllocationsSuccess(t *testing.T) {
    r, netID := setupAllocList(t)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/v1/networks/"+netID+"/ip-allocations", nil)
    req.Header.Set("Authorization", "Bearer dev")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String()) }
    if len(w.Body.Bytes()) == 0 { t.Fatalf("empty body") }
}

func TestListIPAllocationsNonMemberDenied(t *testing.T) {
    gin.SetMode(gin.TestMode)
    nrepo := repository.NewInMemoryNetworkRepository()
    irepo := repository.NewInMemoryIdempotencyRepository()
    mrepo := repository.NewInMemoryMembershipRepository()
    jrepo := repository.NewInMemoryJoinRequestRepository()
    iprepo := repository.NewInMemoryIPAM()
    ns := service.NewNetworkService(nrepo, irepo)
    ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
    ips := service.NewIPAMService(nrepo, mrepo, iprepo)
    h := NewNetworkHandler(ns, ms).WithIPAM(ips)
    r := gin.New()
    r.Use(RoleMiddleware(mrepo))
    RegisterNetworkRoutes(r, h)
    net := &domain.Network{ID: "net-list-2", TenantID: "t1", Name: "NetList2", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.91.0.0/30", CreatedBy: "user_other", CreatedAt: time.Now(), UpdatedAt: time.Now()}
    if err := nrepo.Create(context.Background(), net); err != nil { t.Fatalf("create network: %v", err) }
    // no membership for user_dev
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/v1/networks/"+net.ID+"/ip-allocations", nil)
    req.Header.Set("Authorization", "Bearer dev")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403 got %d body=%s", w.Code, w.Body.String()) }
}
