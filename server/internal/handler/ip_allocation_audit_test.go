package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// TestIPAllocationReleaseAudit verifies IP_ALLOCATED and IP_RELEASED (self) events.
func TestIPAllocationReleaseAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nrepo := repository.NewInMemoryNetworkRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	iprepo := repository.NewInMemoryIPAM()
	ns := service.NewNetworkService(nrepo, irepo)
	ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
	ips := service.NewIPAMService(nrepo, mrepo, iprepo)
	store := audit.NewInMemoryStore()
	ns.SetAuditor(store)
	ms.SetAuditor(store)
	ips.SetAuditor(store)
	h := NewNetworkHandler(ns, ms).WithIPAM(ips)
	r := gin.New()
	r.Use(RoleMiddleware(mrepo))
	RegisterNetworkRoutes(r, h)
	netw := &domain.Network{ID: "audit-net-1", TenantID: "t1", Name: "AuditIP", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.70.0.0/30", CreatedBy: "user_dev", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := nrepo.Create(context.Background(), netw); err != nil {
		t.Fatalf("create network: %v", err)
	}
	_, _ = mrepo.UpsertApproved(context.Background(), netw.ID, "user_dev", domain.RoleMember, time.Now())

	// allocate
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/v1/networks/"+netw.ID+"/ip-allocations", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("alloc expected 200 got %d body=%s", w.Code, w.Body.String())
	}

	// self release
	w = httptest.NewRecorder()
	delReq, _ := http.NewRequest(http.MethodDelete, "/v1/networks/"+netw.ID+"/ip-allocation", nil)
	delReq.Header.Set("Authorization", "Bearer dev")
	delReq.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, delReq)
	if w.Code != http.StatusNoContent {
		t.Fatalf("release expected 204 got %d", w.Code)
	}

	foundAlloc := false
	foundRelease := false
	for _, ev := range store.List() {
		switch ev.Action {
		case "IP_ALLOCATED":
			foundAlloc = true
		case "IP_RELEASED":
			foundRelease = true
		}
	}
	if !foundAlloc || !foundRelease {
		t.Fatalf("expected both IP_ALLOCATED and IP_RELEASED events; got=%v", store.List())
	}
}

// TestAdminReleaseAudit verifies admin release emits IP_RELEASED with released_for detail.
func TestAdminReleaseAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nrepo := repository.NewInMemoryNetworkRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	iprepo := repository.NewInMemoryIPAM()
	ns := service.NewNetworkService(nrepo, irepo)
	ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
	ips := service.NewIPAMService(nrepo, mrepo, iprepo)
	store := audit.NewInMemoryStore()
	ns.SetAuditor(store)
	ms.SetAuditor(store)
	ips.SetAuditor(store)
	h := NewNetworkHandler(ns, ms).WithIPAM(ips)
	r := gin.New()
	r.Use(RoleMiddleware(mrepo))
	RegisterNetworkRoutes(r, h)
	netw := &domain.Network{ID: "audit-net-2", TenantID: "t1", Name: "AuditIP2", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.71.0.0/30", CreatedBy: "user_dev", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := nrepo.Create(context.Background(), netw); err != nil {
		t.Fatalf("create network: %v", err)
	}
	// actor admin
	_, _ = mrepo.UpsertApproved(context.Background(), netw.ID, "user_dev", domain.RoleAdmin, time.Now())
	// target member
	_, _ = mrepo.UpsertApproved(context.Background(), netw.ID, "memberB", domain.RoleMember, time.Now())
	if _, err := ips.AllocateIP(context.Background(), netw.ID, "memberB"); err != nil {
		t.Fatalf("alloc target: %v", err)
	}
	// admin release target
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/v1/networks/"+netw.ID+"/ip-allocations/memberB", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("admin release expected 204 got %d body=%s", w.Code, w.Body.String())
	}
	// find release event with released_for
	found := false
	for _, ev := range store.List() {
		if ev.Action == "IP_RELEASED" {
			if rf, ok := ev.Details["released_for"]; ok && rf == "memberB" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected IP_RELEASED event with released_for detail; events=%v", store.List())
	}
}
