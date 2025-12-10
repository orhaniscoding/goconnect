package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsEndpointAndCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()
	r := gin.New()
	r.Use(GinMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	r.GET("/metrics", Handler())
	// simulate audit event
	IncAudit("test_action")

	// hit /ping
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	// fetch /metrics
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("metrics endpoint not 200: %d", w2.Code)
	}
	body := w2.Body.String()
	if !strings.Contains(body, "goconnect_http_requests_total") || !strings.Contains(body, "goconnect_http_request_duration_seconds_bucket") || !strings.Contains(body, "goconnect_audit_events_total") {
		t.Fatalf("expected metrics not found in body:\n%s", body)
	}
}

func TestRegister_Idempotent(t *testing.T) {
	// Register should be safe to call multiple times
	Register()
	Register()
	Register()
	// If it panics, test fails
}

func TestIncAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncAudit("user.login")
	IncAudit("user.logout")
	IncAudit("user.login") // Increment again

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `goconnect_audit_events_total{action="user.login"}`)
	assert.Contains(t, body, `goconnect_audit_events_total{action="user.logout"}`)
}

func TestAddAuditEviction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	// Add multiple evictions
	AddAuditEviction("memory", 5)
	AddAuditEviction("sqlite", 3)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `goconnect_audit_evictions_total{source="memory"} 5`)
	assert.Contains(t, body, `goconnect_audit_evictions_total{source="sqlite"} 3`)
}

func TestAddAuditEviction_ZeroAndNegative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	// Zero and negative should not increment
	AddAuditEviction("test", 0)
	AddAuditEviction("test", -5)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// Should not have test source metric (or it's 0)
	if strings.Contains(body, `goconnect_audit_evictions_total{source="test"} `) {
		// If it exists, should be 0
		assert.NotContains(t, body, `goconnect_audit_evictions_total{source="test"} 1`)
	}
}

func TestIncAuditFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncAuditFailure("network_error")
	IncAuditFailure("disk_full")
	IncAuditFailure("network_error") // Increment again

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `goconnect_audit_failures_total{reason="network_error"}`)
	assert.Contains(t, body, `goconnect_audit_failures_total{reason="disk_full"}`)
}

func TestObserveAuditInsert(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	ObserveAuditInsert("sqlite", "success", 0.005)
	ObserveAuditInsert("sqlite", "failure", 0.010)
	ObserveAuditInsert("postgres", "success", 0.002)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `goconnect_audit_insert_duration_seconds_bucket{sink="sqlite",status="success"`)
	assert.Contains(t, body, `goconnect_audit_insert_duration_seconds_bucket{sink="sqlite",status="failure"`)
	assert.Contains(t, body, `goconnect_audit_insert_duration_seconds_bucket{sink="postgres",status="success"`)
}

func TestSetAuditQueueDepth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	SetAuditQueueDepth(42)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_queue_depth 42")
}

func TestIncAuditDropped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncAuditDropped()
	IncAuditDropped()
	IncAuditDropped()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_events_dropped_total")
}

func TestObserveAuditDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	ObserveAuditDispatch(0.001)
	ObserveAuditDispatch(0.005)
	ObserveAuditDispatch(0.010)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_dispatch_duration_seconds_bucket")
	assert.Contains(t, body, "goconnect_audit_dispatch_duration_seconds_count")
}

func TestSetAuditQueueHighWatermark(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	SetAuditQueueHighWatermark(100)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_queue_high_watermark 100")
}

func TestIncAuditDroppedReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncAuditDroppedReason("full")
	IncAuditDroppedReason("shutdown")
	IncAuditDroppedReason("full") // Increment full again

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `goconnect_audit_events_dropped_reason_total{reason="full"}`)
	assert.Contains(t, body, `goconnect_audit_events_dropped_reason_total{reason="shutdown"}`)
}

func TestIncAuditWorkerRestart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncAuditWorkerRestart()
	IncAuditWorkerRestart()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_worker_restarts_total")
}

func TestIncChainHead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncChainHead()
	IncChainHead()
	IncChainHead()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_chain_head_advance_total")
}

func TestObserveChainVerification_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	ObserveChainVerification(0.025, true)
	ObserveChainVerification(0.050, true)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_chain_verification_duration_seconds_bucket")
	// Failures should be 0
	assert.Contains(t, body, "goconnect_audit_chain_verification_failures_total 0")
}

func TestObserveChainVerification_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	ObserveChainVerification(0.030, false)
	ObserveChainVerification(0.040, false)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_chain_verification_duration_seconds_bucket")
	// Should have failure count
	assert.Contains(t, body, "goconnect_audit_chain_verification_failures_total")
}

func TestIncChainAnchor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncChainAnchor()
	IncChainAnchor()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_chain_anchor_created_total")
}

func TestObserveIntegrityExport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	ObserveIntegrityExport(0.100)
	ObserveIntegrityExport(0.200)

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_integrity_export_total")
	assert.Contains(t, body, "goconnect_audit_integrity_export_duration_seconds_bucket")
}

func TestIncIntegritySigned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	IncIntegritySigned()
	IncIntegritySigned()
	IncIntegritySigned()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "goconnect_audit_integrity_signed_total")
}

func TestGinMiddleware_Metrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	r := gin.New()
	r.Use(GinMiddleware())
	r.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(200, "ok")
	})
	r.GET("/metrics", Handler())

	// Make request
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Check metrics
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w2, req2)

	body := w2.Body.String()
	assert.Contains(t, body, `goconnect_http_requests_total{method="GET",path="/test",status="200"}`)
	assert.Contains(t, body, `goconnect_http_request_duration_seconds_bucket{method="GET",path="/test"`)
}

func TestGinMiddleware_404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	r := gin.New()
	r.Use(GinMiddleware())
	r.GET("/metrics", Handler())

	// Make request to non-existent endpoint
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/nonexistent", nil)
	r.ServeHTTP(w1, req1)

	// Check metrics
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w2, req2)

	body := w2.Body.String()
	// Should track 404 with actual URL path (no route match, so path = URL path)
	assert.Contains(t, body, `goconnect_http_requests_total{method="GET",path="/nonexistent",status="404"}`)
}

func TestGinMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	r := gin.New()
	r.Use(GinMiddleware())
	r.GET("/api/v1/users", func(c *gin.Context) { c.String(200, "users") })
	r.POST("/api/v1/users", func(c *gin.Context) { c.String(201, "created") })
	r.GET("/metrics", Handler())

	// Multiple GET requests
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		r.ServeHTTP(w, req)
	}

	// POST request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", nil)
	r.ServeHTTP(w, req)

	// Check metrics
	wMetrics := httptest.NewRecorder()
	reqMetrics := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(wMetrics, reqMetrics)

	body := wMetrics.Body.String()
	assert.Contains(t, body, `goconnect_http_requests_total{method="GET",path="/api/v1/users",status="200"}`)
	assert.Contains(t, body, `goconnect_http_requests_total{method="POST",path="/api/v1/users",status="201"}`)
}

func TestHandler_ReturnsPrometheusMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	r := gin.New()
	r.GET("/metrics", Handler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body := w.Body.String()

	// Verify Prometheus format
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
	assert.Contains(t, body, "goconnect_")
}

func TestMetrics_AllCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Register()

	// Trigger all metrics
	IncAudit("test")
	AddAuditEviction("memory", 1)
	IncAuditFailure("test")
	ObserveAuditInsert("sqlite", "success", 0.001)
	SetAuditQueueDepth(10)
	IncAuditDropped()
	ObserveAuditDispatch(0.002)
	SetAuditQueueHighWatermark(20)
	IncAuditDroppedReason("full")
	IncAuditWorkerRestart()
	IncChainHead()
	ObserveChainVerification(0.01, true)
	IncChainAnchor()
	ObserveIntegrityExport(0.05)
	IncIntegritySigned()

	r := gin.New()
	r.GET("/metrics", Handler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// Verify all metrics are present
	expectedMetrics := []string{
		"goconnect_audit_events_total",
		"goconnect_audit_evictions_total",
		"goconnect_audit_failures_total",
		"goconnect_audit_insert_duration_seconds",
		"goconnect_audit_queue_depth",
		"goconnect_audit_events_dropped_total",
		"goconnect_audit_dispatch_duration_seconds",
		"goconnect_audit_queue_high_watermark",
		"goconnect_audit_events_dropped_reason_total",
		"goconnect_audit_worker_restarts_total",
		"goconnect_audit_chain_head_advance_total",
		"goconnect_audit_chain_verification_duration_seconds",
		"goconnect_audit_chain_verification_failures_total",
		"goconnect_audit_chain_anchor_created_total",
		"goconnect_audit_integrity_export_total",
		"goconnect_audit_integrity_export_duration_seconds",
		"goconnect_audit_integrity_signed_total",
	}

	for _, metric := range expectedMetrics {
		assert.Contains(t, body, metric, "Expected metric %s not found", metric)
	}
}

func TestSetWGPeersTotal(t *testing.T) {
	Register()

	// Should not panic
	SetWGPeersTotal(10)
	SetWGPeersTotal(0)
	SetWGPeersTotal(100)

	// Verify metric is registered by checking /metrics endpoint
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", Handler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "goconnect_wg_peers_total")
}

func TestSetWGPeerStats(t *testing.T) {
	Register()

	// Should not panic
	SetWGPeerStats("pubkey1234567890", 1024, 2048, 60.5)
	SetWGPeerStats("pubkey0987654321", 4096, 8192, 120.0)

	// Verify metrics are registered by checking /metrics endpoint
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", Handler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "goconnect_wg_peer_rx_bytes")
	assert.Contains(t, body, "goconnect_wg_peer_tx_bytes")
	assert.Contains(t, body, "goconnect_wg_peer_last_handshake_seconds")
}
