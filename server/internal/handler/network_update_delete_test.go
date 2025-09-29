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

<<<<<<< HEAD
func setupUpdateDelete() (*gin.Engine, repository.NetworkRepository, repository.MembershipRepository) {
	gin.SetMode(gin.TestMode)
	nrepo := repository.NewInMemoryNetworkRepository()
	irepo := repository.NewInMemoryIdempotencyRepository()
	mrepo := repository.NewInMemoryMembershipRepository()
	jrepo := repository.NewInMemoryJoinRequestRepository()
	ns := service.NewNetworkService(nrepo, irepo)
	ms := service.NewMembershipService(nrepo, mrepo, jrepo, irepo)
	h := NewNetworkHandler(ns, ms)
	r := gin.New()
	// inject role middleware BEFORE route registration so it executes earlier
	r.Use(RoleMiddleware(mrepo))
	RegisterNetworkRoutes(r, h)
	return r, nrepo, mrepo
}

func seedNet(t *testing.T, repo repository.NetworkRepository, id, name string) {
	n := &domain.Network{ID: id, TenantID: "t1", Name: name, Visibility: domain.NetworkVisibilityPublic, JoinPolicy: domain.JoinPolicyOpen, CIDR: "10.20.0.0/24", CreatedBy: "u", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(context.Background(), n); err != nil {
		t.Fatalf("seed create: %v", err)
	}
}

func TestNetworkUpdateNameAndDuplicate(t *testing.T) {
	r, repo, mrepo := setupUpdateDelete()
	seedNet(t, repo, "net-up-1", "Alpha")
	seedNet(t, repo, "net-up-2", "Beta")
	// seed admin membership for dev user
	// give user_dev owner role on first network and admin on second to test both
	_, _ = mrepo.UpsertApproved(context.Background(), "net-up-1", "user_dev", domain.RoleOwner, time.Now())
	_, _ = mrepo.UpsertApproved(context.Background(), "net-up-2", "user_dev", domain.RoleAdmin, time.Now())
	// update name of first to new unique
	body := map[string]string{"name": "Gamma"}
	buf, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/v1/networks/net-up-1", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 update, got %d body=%s", w.Code, w.Body.String())
	}

	// duplicate name attempt (rename second to Gamma) should 400
	body["name"] = "Gamma"
	buf, _ = json.Marshal(body)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/v1/networks/net-up-2", bytes.NewBuffer(buf))
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate name, got %d", w.Code)
	}
}

func TestNetworkSoftDeleteExclusion(t *testing.T) {
	r, repo, mrepo := setupUpdateDelete()
	seedNet(t, repo, "net-del-1", "DelNet")
	_, _ = mrepo.UpsertApproved(context.Background(), "net-del-1", "user_dev", domain.RoleAdmin, time.Now())
	// delete
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/networks/net-del-1", nil)
	req.Header.Set("Authorization", "Bearer dev")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 delete, got %d", w.Code)
	}
	// get should 404
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/networks/net-del-1", nil)
	req.Header.Set("Authorization", "Bearer dev")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after soft delete, got %d", w.Code)
	}
}
