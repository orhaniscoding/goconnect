package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

func TestRateLimit_Join429(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a network as admin (separate bucket)
	create := domain.CreateNetworkRequest{
		Name:       "RL Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.99.0.0/24",
	}
	payload, _ := json.Marshal(create)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 create network, got %d body=%s", w.Code, w.Body.String())
	}
	var cresp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cresp); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	netID := cresp.Data.ID
	if netID == "" {
		t.Fatalf("network id empty in create response: %s", w.Body.String())
	}

	// Now send 6 join requests as user_dev within 1s
	var last *httptest.ResponseRecorder
	for i := 0; i < 6; i++ {
		last = httptest.NewRecorder()
		jreq, _ := http.NewRequest("POST", "/v1/networks/"+netID+"/join", nil)
		jreq.Header.Set("Authorization", "Bearer valid-token")
		jreq.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
		router.ServeHTTP(last, jreq)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 6th request, got %d body=%s", last.Code, last.Body.String())
	}
	var errResp domain.Error
	if err := json.Unmarshal(last.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal rate limit error: %v", err)
	}
	if errResp.Code != domain.ErrRateLimited {
		t.Fatalf("expected code %s, got %s", domain.ErrRateLimited, errResp.Code)
	}
}
