package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
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
		aud.Event(ctx, "test-tenant", "ACT", "actor", "obj", map[string]any{"i": i})
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

func TestAuditListHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aud, err := audit.NewSqliteAuditor(":memory:", audit.WithAnchorInterval(10))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer aud.Close()

	ctx := context.Background()
	// produce events
	for i := 0; i < 5; i++ {
		aud.Event(ctx, "test-tenant", "ACT", "actor", "obj", map[string]any{"i": i})
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/v1/audit", AuditListHandler(aud))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/audit?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := resp["data"]; !ok {
		t.Fatalf("expected data field in response")
	}
	if _, ok := resp["pagination"]; !ok {
		t.Fatalf("expected pagination field in response")
	}
}

func TestAuditListHandler_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aud, err := audit.NewSqliteAuditor(":memory:", audit.WithAnchorInterval(10))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer aud.Close()

	r := gin.New()
	// No tenant_id middleware
	r.GET("/v1/audit", AuditListHandler(aud))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/audit", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("expected 401 got %d", w.Code)
	}
}

func TestAuditListHandler_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aud, err := audit.NewSqliteAuditor(":memory:", audit.WithAnchorInterval(10))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer aud.Close()

	ctx := context.Background()
	aud.Event(ctx, "test-tenant", "ACTION", "specific-actor", "specific-obj", nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/v1/audit", AuditListHandler(aud))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/audit?actor=specific-actor&action=ACTION", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAuditListHandler_InvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	aud, err := audit.NewSqliteAuditor(":memory:", audit.WithAnchorInterval(10))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer aud.Close()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/v1/audit", AuditListHandler(aud))

	w := httptest.NewRecorder()
	// page=-1, limit=999 - should be corrected internally
	req := httptest.NewRequest("GET", "/v1/audit?page=-1&limit=999", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	pagination := resp["pagination"].(map[string]interface{})
	// Invalid page should be corrected to 1
	if int(pagination["page"].(float64)) != 1 {
		t.Fatalf("expected page=1, got %v", pagination["page"])
	}
	// Invalid limit should be corrected to 20 (default)
	if int(pagination["limit"].(float64)) != 20 {
		t.Fatalf("expected limit=20, got %v", pagination["limit"])
	}
}
