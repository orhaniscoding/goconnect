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
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// TestAuditEvents_NetworkLifecycle verifies NETWORK_CREATED/UPDATED/DELETED audit emission.
func TestAuditEvents_NetworkLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nrepo := repository.NewInMemoryNetworkRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(nrepo, irepo)
	ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
	store := audit.NewInMemoryStore()
	ns.SetAuditor(store)
	ms.SetAuditor(store)
	authSvc := newMockAuthServiceWithTokens()

	// Setup for DeviceService
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, nrepo, wgConfig)

	h := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)
	r := gin.New()
	RegisterNetworkRoutes(r, h, authSvc, mrepo)

	// create network
	createReq := domain.CreateNetworkRequest{Name: "AuditNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.50.0.0/24"}
	body, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 create got %d body=%s", w.Code, w.Body.String())
	}

	// extract network id
	var resp struct {
		Data domain.Network `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// seed admin role to allow patch/delete
	_, _ = mrepo.UpsertApproved(context.Background(), resp.Data.ID, "user_dev", domain.RoleAdmin, time.Now())

	// update network name
	patch := map[string]string{"name": "AuditNet2"}
	pb, _ := json.Marshal(patch)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/v1/networks/"+resp.Data.ID, bytes.NewBuffer(pb))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 patch got %d", w.Code)
	}

	// delete network
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/v1/networks/"+resp.Data.ID, nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 delete got %d", w.Code)
	}

	events := store.List()
	// Expect at least create+update+delete
	want := map[string]bool{"NETWORK_CREATED": false, "NETWORK_UPDATED": false, "NETWORK_DELETED": false}
	for _, ev := range events {
		if _, ok := want[ev.Action]; ok {
			want[ev.Action] = true
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("missing audit event %s; got=%v", k, events)
		}
	}
}

// TestAuditEvents_MembershipFlow verifies join request + approve + ban events.
func TestAuditEvents_MembershipFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nrepo := repository.NewInMemoryNetworkRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(nrepo, irepo)
	ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
	store := audit.NewInMemoryStore()
	ns.SetAuditor(store)
	ms.SetAuditor(store)
	authSvc := newMockAuthServiceWithTokens()

	// Setup for DeviceService
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	ds := service.NewDeviceService(deviceRepo, userRepo, peerRepo, nrepo, wgConfig)

	h := NewNetworkHandler(ns, ms, ds, peerRepo, wgConfig)
	r := gin.New()
	RegisterNetworkRoutes(r, h, authSvc, mrepo)

	// create network (will emit created)
	createReq := domain.CreateNetworkRequest{Name: "JoinNet", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.51.0.0/24"}
	body, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 create got %d", w.Code)
	}
	var resp struct {
		Data domain.Network `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}

	// seed owner (admin_dev token maps to admin_dev user id) but we will use dev as actor; we need dev to be admin role
	_, _ = mrepo.UpsertApproved(context.Background(), resp.Data.ID, "user_dev", domain.RoleAdmin, time.Now())

	// join request by another user via token 'dev' is user_dev, so we need a different user id; simulate by direct repo insert pending? Better: create join via API with same user causes member; we want approval flow: use user_dev2 by forging Authorization? We'll simulate join request create via repo.
	jr, err := jrepo.CreatePending(context.Background(), resp.Data.ID, "candidate")
	if err != nil {
		t.Fatalf("create pending: %v", err)
	}
	_ = jr

	// approve candidate
	apprBody, _ := json.Marshal(map[string]string{"user_id": "candidate"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/networks/"+resp.Data.ID+"/approve", bytes.NewBuffer(apprBody))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 approve got %d body=%s", w.Code, w.Body.String())
	}

	// ban candidate
	banBody, _ := json.Marshal(map[string]string{"user_id": "candidate"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/networks/"+resp.Data.ID+"/ban", bytes.NewBuffer(banBody))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 ban got %d", w.Code)
	}

	events := store.List()
	want := map[string]bool{"NETWORK_CREATED": false, "NETWORK_JOIN_APPROVE": false, "NETWORK_MEMBER_BAN": false}
	for _, ev := range events {
		if _, ok := want[ev.Action]; ok {
			want[ev.Action] = true
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("missing audit event %s; events=%v", k, events)
		}
	}
}
