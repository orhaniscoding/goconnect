package handler

import (
    "bytes"
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

func setupIPRelease() (*gin.Engine, repository.MembershipRepository, string) {
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
    net := &domain.Network{ID: "net-rel-1", TenantID: "t1", Name: "RelNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.90.0.0/30", CreatedBy: "user_dev", CreatedAt: time.Now(), UpdatedAt: time.Now()}
    if err := nrepo.Create(context.Background(), net); err != nil { panic(err) }
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())
    return r, mrepo, net.ID
}

func TestIPAllocationReleaseFlow(t *testing.T) {
    r, _, netID := setupIPRelease()
    // allocate
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/v1/networks/"+netID+"/ip-allocations", bytes.NewBuffer([]byte("{}")))
    req.Header.Set("Authorization", "Bearer dev")
    req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("allocate expected 200 got %d", w.Code) }

    // release
    w = httptest.NewRecorder()
    delReq, _ := http.NewRequest("DELETE", "/v1/networks/"+netID+"/ip-allocation", nil)
    delReq.Header.Set("Authorization", "Bearer dev")
    delReq.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
    r.ServeHTTP(w, delReq)
    if w.Code != http.StatusNoContent { t.Fatalf("release expected 204 got %d", w.Code) }

    // allocate again -> should reuse same IP (repository logic) but handler test just checks 200
    w = httptest.NewRecorder()
    req2, _ := http.NewRequest("POST", "/v1/networks/"+netID+"/ip-allocations", bytes.NewBuffer([]byte("{}")))
    req2.Header.Set("Authorization", "Bearer dev")
    req2.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
    r.ServeHTTP(w, req2)
    if w.Code != http.StatusOK { t.Fatalf("re-allocate expected 200 got %d", w.Code) }
}

// NOTE: Negative release (non-member) HTTP test omitted due to auth middleware always mapping token -> user_dev; service layer covers non-member case separately.