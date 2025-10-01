package metrics_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	a "github.com/orhaniscoding/goconnect/server/internal/audit"
	m "github.com/orhaniscoding/goconnect/server/internal/metrics"
)

func TestAuditEvictionMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m.Register()
	store := a.NewInMemoryStore(a.WithCapacity(1))
	r := gin.New()
	r.GET("/metrics", m.Handler())

	store.Event(context.Background(), "TEST_ACTION", "actor1", "object1", nil)
	store.Event(context.Background(), "TEST_ACTION", "actor2", "object2", nil) // triggers eviction

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	// Expect labeled eviction metric source="memory"
	if !strings.Contains(body, `goconnect_audit_evictions_total{source="memory"} 1`) {
		 t.Fatalf("expected memory eviction metric =1, body=\n%s", body)
	}
}

func TestAuditFailureMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m.Register()
	auditor, err := a.NewSqliteAuditor(":memory:")
	if err != nil {
		t.Fatalf("sqlite auditor create: %v", err)
	}
	// Close DB to force failure on insert
	_ = auditor.Close()
	auditor.Event(context.Background(), "TEST_ACTION", "actor", "object", nil)

	r := gin.New()
	r.GET("/metrics", m.Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "goconnect_audit_failures_total 1") {
		t.Fatalf("expected failure metric =1, body=\n%s", body)
	}
}

func TestSqliteRetentionEvictionMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m.Register()
	auditor, err := a.NewSqliteAuditor(":memory:", a.WithMaxRows(3))
	if err != nil { t.Fatalf("sqlite auditor create: %v", err) }
	ctx := context.Background()
	// Insert 5 events; expect 2 evictions (pruned oldest) leaving 3
	for i := 0; i < 5; i++ {
		auditor.Event(ctx, "TEST_ACTION", "actor", "object", map[string]any{"i": i})
	}
	r := gin.New()
	r.GET("/metrics", m.Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
	body := w.Body.String()
	if !strings.Contains(body, `goconnect_audit_evictions_total{source="sqlite"} 2`) {
		t.Fatalf("expected sqlite eviction metric =2 body=\n%s", body)
	}
}
