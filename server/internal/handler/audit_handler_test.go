package handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"net/http/httptest"
	"testing"
)

func TestAuditIntegrityHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics.Register()
	aud, err := audit.NewSqliteAuditor(":memory:", audit.WithAnchorInterval(2))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	ctx := context.Background()
	// produce events -> anchors at 2,4
	for i := 0; i < 5; i++ {
		aud.Event(ctx, "ACT", "actor", "obj", map[string]any{"i": i})
	}
	r := gin.New()
	r.GET("/v1/audit/integrity", AuditIntegrityHandler(aud))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/audit/integrity?anchors=5", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	var resp struct {
		Head struct {
			Seq  int64  `json:"seq"`
			Hash string `json:"hash"`
			TS   string `json:"ts"`
		} `json:"head"`
		Anchors []struct {
			Seq  int64  `json:"seq"`
			Hash string `json:"hash"`
			TS   string `json:"ts"`
		} `json:"anchors"`
		LatestSeq int64 `json:"latest_seq"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Head.Seq == 0 || resp.Head.Hash == "" {
		t.Fatalf("missing head in response: %+v", resp)
	}
	if resp.LatestSeq != resp.Head.Seq {
		t.Fatalf("latest_seq mismatch head.seq")
	}
	if len(resp.Anchors) == 0 {
		t.Fatalf("expected anchors present")
	}
	// Ensure anchors ascending by seq
	for i := 1; i < len(resp.Anchors); i++ {
		if resp.Anchors[i-1].Seq >= resp.Anchors[i].Seq {
			t.Fatalf("anchors not strictly ascending: %+v", resp.Anchors)
		}
	}
	// Last anchor seq must be <= head seq
	if resp.Anchors[len(resp.Anchors)-1].Seq > resp.Head.Seq {
		t.Fatalf("last anchor seq exceeds head seq")
	}
}
