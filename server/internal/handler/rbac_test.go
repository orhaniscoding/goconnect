package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/orhaniscoding/goconnect/server/internal/domain"
    "github.com/orhaniscoding/goconnect/server/internal/repository"
    "github.com/orhaniscoding/goconnect/server/internal/service"
)

func setupRBAC() (*gin.Engine, repository.NetworkRepository, repository.MembershipRepository) {
    gin.SetMode(gin.TestMode)
    nrepo := repository.NewInMemoryNetworkRepository()
    irepo := repository.NewInMemoryIdempotencyRepository()
    mrepo := repository.NewInMemoryMembershipRepository()
    jrepo := repository.NewInMemoryJoinRequestRepository()

    ns := service.NewNetworkService(nrepo, irepo)
    ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)

    h := NewNetworkHandler(ns, ms)
    r := gin.New()
    RegisterNetworkRoutes(r, h)
    return r, nrepo, mrepo
}

func TestRBAC_ApproveForbiddenForMember(t *testing.T) {
    r, nrepo, mrepo := setupRBAC()
    net := &domain.Network{ID: "net-rbac-1", TenantID: "t1", Name: "N", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.8.0.0/24", CreatedBy: "owner_dev"}
    _ = nrepo.Create(context.Background(), net)
    // Seed only a regular member (not admin/owner)
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

    // user_dev tries to approve someone else -> forbidden
    body := map[string]string{"user_id": "another"}
    buf, _ := json.Marshal(body)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer dev")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden {
        t.Fatalf("expected 403 for non-admin approve, got %d body=%s", w.Code, w.Body.String())
    }
}

func TestRBAC_AdminCanApprove(t *testing.T) {
    r, nrepo, mrepo := setupRBAC()
    net := &domain.Network{ID: "net-rbac-2", TenantID: "t1", Name: "N2", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.9.0.0/24", CreatedBy: "owner_dev"}
    _ = nrepo.Create(context.Background(), net)
    // Seed owner as admin_dev
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

    // Pre-create pending join for user_dev
    // NOTE: membership service inside handler uses its own repo instance; instead, simulate via API:
    // user_dev join request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
    req.Header.Set("Authorization", "Bearer dev")
    req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
    r.ServeHTTP(w, req)
    if w.Code != http.StatusAccepted { t.Fatalf("expected 202 for join, got %d", w.Code) }

    // approve by admin
    w = httptest.NewRecorder()
    buf, _ := json.Marshal(map[string]string{"user_id": "user_dev"})
    req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer admin")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 approve by admin, got %d body=%s", w.Code, w.Body.String())
    }
}

func TestRBAC_OnlyAdminCanBanOrKick(t *testing.T) {
    r, nrepo, mrepo := setupRBAC()
    net := &domain.Network{ID: "net-rbac-3", TenantID: "t1", Name: "N3", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.10.0.0/24", CreatedBy: "owner_dev"}
    _ = nrepo.Create(context.Background(), net)
    // Seed owner and a member
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())
    _, _ = mrepo.UpsertApproved(context.Background(), net.ID, "user_dev", domain.RoleMember, time.Now())

    // member tries to ban -> forbidden
    w := httptest.NewRecorder()
    buf, _ := json.Marshal(map[string]string{"user_id": "admin_dev"})
    req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer dev")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403 non-admin ban, got %d", w.Code) }

    // admin bans member -> ok
    w = httptest.NewRecorder()
    buf, _ = json.Marshal(map[string]string{"user_id": "user_dev"})
    req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer admin")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 admin ban, got %d", w.Code) }

    // member tries to kick -> forbidden
    w = httptest.NewRecorder()
    buf, _ = json.Marshal(map[string]string{"user_id": "admin_dev"})
    req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer dev")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403 non-admin kick, got %d", w.Code) }

    // admin kicks (no-op if not member)
    w = httptest.NewRecorder()
    buf, _ = json.Marshal(map[string]string{"user_id": "user_dev"})
    req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", bytes.NewBuffer(buf))
    req.Header.Set("Authorization", "Bearer admin")
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 admin kick, got %d", w.Code) }
}
