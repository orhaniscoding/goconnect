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
	auditEvictions = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_evictions_total",
		Help:      "Total audit events evicted from in-memory retention buffer",
	})
	auditFailures = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_failures_total",
		Help:      "Total audit persistence failures (best-effort sinks)",
	})
)

// Register all metrics (idempotent safe to call once at startup).
func Register() { registerOnce.Do(func() { prom.MustRegister(reqCounter, reqLatency, auditEvents, auditEvictions, auditFailures) }) }

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
func IncAuditEviction() { auditEvictions.Inc() }

// IncAuditFailure increments the failure counter.
func IncAuditFailure() { auditFailures.Inc() }
