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

// sets up a router with explicit repos/services so we can seed state
func setupMembershipsRouter() (*gin.Engine, *service.MembershipService, repository.MembershipRepository, repository.NetworkRepository) {
	gin.SetMode(gin.TestMode)

	// repos
	networkRepo := repository.NewInMemoryNetworkRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()

	// services
	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)

	// handler
	h := NewNetworkHandler(networkService, membershipService)
	r := gin.New()
	RegisterNetworkRoutes(r, h)

	return r, membershipService, membershipRepo, networkRepo
}

func TestJoinApproveFlow_ListApproved(t *testing.T) {
	r, _, mrepo, nrepo := setupMembershipsRouter()

	// seed network with approval policy
	net := &domain.Network{ID: "net-appr-1", TenantID: "t1", Name: "N", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.5.0.0/24", CreatedBy: "admin_dev"}
	if err := nrepo.Create(context.Background(), net); err != nil {
		t.Fatalf("create network: %v", err)
	}

	// seed admin membership for actor admin_dev
	if _, err := mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now()); err != nil {
		t.Fatalf("seed admin membership: %v", err)
	}

	// user_dev requests join
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for join request, got %d body=%s", w.Code, w.Body.String())
	}

	// admin approves user_dev
	w = httptest.NewRecorder()
	body := map[string]string{"user_id": "user_dev"}
	buf, _ := json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 approve, got %d body=%s", w.Code, w.Body.String())
	}

	// list approved members should contain both admin_dev and user_dev
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/"+net.ID+"/members?status=approved", nil)
	req.Header.Set("Authorization", "Bearer admin")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 list members, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []domain.Membership `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	found := map[string]bool{}
	for _, m := range resp.Data {
		found[m.UserID] = true
	}
	if !found["admin_dev"] || !found["user_dev"] {
		t.Fatalf("expected approved members to include admin_dev and user_dev: %+v", found)
	}
}

func TestDenyRemainsUnapproved(t *testing.T) {
	r, _, mrepo, nrepo := setupMembershipsRouter()
	net := &domain.Network{ID: "net-appr-2", TenantID: "t1", Name: "N2", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.6.0.0/24", CreatedBy: "admin_dev"}
	if err := nrepo.Create(context.Background(), net); err != nil { t.Fatalf("create network: %v", err) }
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	// join
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	// deny by admin
	w = httptest.NewRecorder()
	body := map[string]string{"user_id": "user_dev"}
	buf, _ := json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/deny", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 deny, got %d", w.Code)
	}

	// list approved: should only contain admin_dev
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/"+net.ID+"/members?status=approved", nil)
	req.Header.Set("Authorization", "Bearer admin")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 list, got %d", w.Code)
	}
	var resp struct {
		Data []domain.Membership `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	foundUser := false
	foundAdmin := false
	for _, m := range resp.Data {
		if m.UserID == "user_dev" {
			foundUser = true
		}
		if m.UserID == "admin_dev" {
			foundAdmin = true
		}
	}
	if foundUser || !foundAdmin {
		t.Fatalf("deny should not approve user; got user=%v admin=%v", foundUser, foundAdmin)
	}
}

func TestBanAndKickFlows(t *testing.T) {
	r, _, mrepo, nrepo := setupMembershipsRouter()
	net := &domain.Network{ID: "net-appr-3", TenantID: "t1", Name: "N3", Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyApproval, CIDR: "10.7.0.0/24", CreatedBy: "admin_dev"}
	if err := nrepo.Create(context.Background(), net); err != nil { t.Fatalf("create network: %v", err) }
	_, _ = mrepo.UpsertApproved(context.Background(), net.ID, "admin_dev", domain.RoleOwner, time.Now())

	// user joins and admin approves
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks/"+net.ID+"/join", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	buf, _ := json.Marshal(map[string]string{"user_id": "user_dev"})
	req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/approve", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 approve, got %d", w.Code)
	}

	// ban the user
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/ban", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 ban, got %d", w.Code)
	}

	// list banned members should include user_dev
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/"+net.ID+"/members?status=banned", nil)
	req.Header.Set("Authorization", "Bearer admin")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 list banned, got %d", w.Code)
	}
	var respB struct {
		Data []domain.Membership `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respB)
	banned := false
	for _, m := range respB.Data {
		if m.UserID == "user_dev" {
			banned = true
		}
	}
	if !banned {
		t.Fatalf("expected user_dev to be banned")
	}

	// approve again to become approved then kick
	// First, set status to approved again
	_ = mrepo.SetStatus(context.Background(), net.ID, "user_dev", domain.StatusApproved)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/networks/"+net.ID+"/kick", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer admin")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 kick, got %d", w.Code)
	}

	// approved list should not include user_dev anymore
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/"+net.ID+"/members?status=approved", nil)
	req.Header.Set("Authorization", "Bearer admin")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 list approved, got %d", w.Code)
	}
	var respA struct {
		Data []domain.Membership `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respA)
	still := false
	for _, m := range respA.Data {
		if m.UserID == "user_dev" {
			still = true
		}
	}
	if still {
		t.Fatalf("user_dev should have been kicked and not in approved list")
	}
}
