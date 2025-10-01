package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerOnce sync.Once
	reqCounter   = prom.NewCounterVec(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "http_requests_total",
		Help:      "Total HTTP requests",
	}, []string{"method", "path", "status"})
	reqLatency = prom.NewHistogramVec(prom.HistogramOpts{
		Namespace: "goconnect",
		Name:      "http_request_duration_seconds",
		Help:      "Request duration seconds",
		Buckets:   prom.DefBuckets,
	}, []string{"method", "path"})
	auditEvents = prom.NewCounterVec(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_events_total",
		Help:      "Audit events emitted",
	}, []string{"action"})
	auditEvictions = prom.NewCounterVec(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_evictions_total",
		Help:      "Total audit events evicted from retention (memory/sqlite)",
	}, []string{"source"})
	auditFailures = prom.NewCounterVec(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_failures_total",
		Help:      "Total audit persistence failures (best-effort sinks)",
	}, []string{"reason"})
	auditInsertLatency = prom.NewHistogramVec(prom.HistogramOpts{
		Namespace: "goconnect",
		Name:      "audit_insert_duration_seconds",
		Help:      "Latency of audit event persistence operations",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"sink", "status"})
)

// Register all metrics (idempotent safe to call once at startup).
func Register() { registerOnce.Do(func() { prom.MustRegister(reqCounter, reqLatency, auditEvents, auditEvictions, auditFailures, auditInsertLatency) }) }

// GinMiddleware instruments incoming HTTP requests.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		reqLatency.WithLabelValues(c.Request.Method, path).Observe(duration)
		reqCounter.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	}
}

// Handler returns a standard promhttp handler.
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) { h.ServeHTTP(c.Writer, c.Request) }
}

// IncAudit increments audit event counter for given action.
func IncAudit(action string) { auditEvents.WithLabelValues(action).Inc() }

// IncAuditEviction increments the eviction counter.
// AddAuditEviction increments eviction counter for a specific source (e.g., "memory" or "sqlite").
func AddAuditEviction(source string, n int) {
	if n <= 0 {
		return
	}
	for i := 0; i < n; i++ {
		auditEvictions.WithLabelValues(source).Inc()
	}
}

// IncAuditFailure increments the failure counter.
// IncAuditFailure increments failure counter with a reason label.
func IncAuditFailure(reason string) { auditFailures.WithLabelValues(reason).Inc() }

// ObserveAuditInsert records insert latency for a sink with status (success|failure).
func ObserveAuditInsert(sink, status string, seconds float64) {
	auditInsertLatency.WithLabelValues(sink, status).Observe(seconds)
}
