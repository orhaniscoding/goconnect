package metrics

import (
    "net/http/httptest"
    "testing"
    "strings"
    "github.com/gin-gonic/gin"
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
    if w1.Code != 200 { t.Fatalf("expected 200, got %d", w1.Code) }

    // fetch /metrics
    w2 := httptest.NewRecorder()
    req2 := httptest.NewRequest("GET", "/metrics", nil)
    r.ServeHTTP(w2, req2)
    if w2.Code != 200 { t.Fatalf("metrics endpoint not 200: %d", w2.Code) }
    body := w2.Body.String()
    if !strings.Contains(body, "goconnect_http_requests_total") || !strings.Contains(body, "goconnect_http_request_duration_seconds_bucket") || !strings.Contains(body, "goconnect_audit_events_total") {
        t.Fatalf("expected metrics not found in body:\n%s", body)
    }
}

