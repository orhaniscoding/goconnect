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
	auditQueueDepth = prom.NewGauge(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "audit_queue_depth",
		Help:      "Current depth of async audit event queue",
	})
	auditDropped = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_events_dropped_total",
		Help:      "Total audit events dropped due to full async queue",
	})
	auditDispatchLatency = prom.NewHistogram(prom.HistogramOpts{
		Namespace: "goconnect",
		Name:      "audit_dispatch_duration_seconds",
		Help:      "Latency from enqueue to dispatch for async audit events",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	})
	auditQueueHighWatermark = prom.NewGauge(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "audit_queue_high_watermark",
		Help:      "Maximum observed async audit queue depth since process start",
	})
	auditDroppedByReason = prom.NewCounterVec(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_events_dropped_reason_total",
		Help:      "Total audit events dropped categorized by reason (full|shutdown)",
	}, []string{"reason"})
	auditWorkerRestarts = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_worker_restarts_total",
		Help:      "Total async audit worker restarts after panic recovery",
	})
	chainHeadAdvance = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_chain_head_advance_total",
		Help:      "Total times audit hash chain head advanced (events inserted)",
	})
	chainVerifyDuration = prom.NewHistogram(prom.HistogramOpts{
		Namespace: "goconnect",
		Name:      "audit_chain_verification_duration_seconds",
		Help:      "Duration of full hash chain verification runs",
		Buckets:   []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	})
	chainVerifyFailures = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_chain_verification_failures_total",
		Help:      "Total failed audit chain verification attempts",
	})
	chainAnchorCreated = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_chain_anchor_created_total",
		Help:      "Total anchor snapshots recorded for audit hash chain",
	})
	integrityExportCounter = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_integrity_export_total",
		Help:      "Total integrity export requests served",
	})
	integrityExportDuration = prom.NewHistogram(prom.HistogramOpts{
		Namespace: "goconnect",
		Name:      "audit_integrity_export_duration_seconds",
		Help:      "Duration of integrity export generation",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	})
	integritySignedCounter = prom.NewCounter(prom.CounterOpts{
		Namespace: "goconnect",
		Name:      "audit_integrity_signed_total",
		Help:      "Total integrity export snapshots successfully signed",
	})

	// WireGuard Metrics
	wgPeersTotal = prom.NewGauge(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "wg_peers_total",
		Help:      "Total number of peers configured on the WireGuard interface",
	})
	wgPeerRxBytes = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "wg_peer_rx_bytes_total",
		Help:      "Total bytes received from peer",
	}, []string{"public_key"})
	wgPeerTxBytes = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "wg_peer_tx_bytes_total",
		Help:      "Total bytes transmitted to peer",
	}, []string{"public_key"})
	wgPeerLastHandshake = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: "goconnect",
		Name:      "wg_peer_last_handshake_seconds",
		Help:      "Seconds since last handshake with peer",
	}, []string{"public_key"})
)

// Register all metrics (idempotent safe to call once at startup).
func Register() {
	registerOnce.Do(func() {
		prom.MustRegister(reqCounter, reqLatency, auditEvents, auditEvictions, auditFailures, auditInsertLatency, auditQueueDepth, auditDropped, auditDispatchLatency, auditQueueHighWatermark, auditDroppedByReason, auditWorkerRestarts, chainHeadAdvance, chainVerifyDuration, chainVerifyFailures, chainAnchorCreated, integrityExportCounter, integrityExportDuration, integritySignedCounter)
		prom.MustRegister(wgPeersTotal, wgPeerRxBytes, wgPeerTxBytes, wgPeerLastHandshake)
	})
}

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

// SetAuditQueueDepth sets the current queue depth gauge.
func SetAuditQueueDepth(n int) { auditQueueDepth.Set(float64(n)) }

// IncAuditDropped increments dropped events counter.
func IncAuditDropped() { auditDropped.Inc() }

// ObserveAuditDispatch records time from enqueue to dispatch.
func ObserveAuditDispatch(seconds float64) { auditDispatchLatency.Observe(seconds) }

// IncChainHead increments chain head advance counter.
func IncChainHead() { chainHeadAdvance.Inc() }

// ObserveChainVerification records verification duration and success/failure.
func ObserveChainVerification(seconds float64, ok bool) {
	chainVerifyDuration.Observe(seconds)
	if !ok {
		chainVerifyFailures.Inc()
	}
}

// IncChainAnchor increments anchor creation counter.
func IncChainAnchor() { chainAnchorCreated.Inc() }

// SetAuditQueueHighWatermark sets the high watermark gauge if higher than current.
func SetAuditQueueHighWatermark(n int) {
	// Gauge has no atomic get; rely on exported value monotonic by only setting when higher.
	// Simplicity: always set; caller ensures only when exceeding known max.
	auditQueueHighWatermark.Set(float64(n))
}

// IncAuditDroppedReason increments dropped counter with reason label.
func IncAuditDroppedReason(reason string) { auditDroppedByReason.WithLabelValues(reason).Inc() }

// IncAuditWorkerRestart increments worker restart counter.
func IncAuditWorkerRestart() { auditWorkerRestarts.Inc() }

// ObserveIntegrityExport records export duration and increments counter.
func ObserveIntegrityExport(seconds float64) {
	integrityExportCounter.Inc()
	integrityExportDuration.Observe(seconds)
}

// IncIntegritySigned increments the signed export counter.
func IncIntegritySigned() { integritySignedCounter.Inc() }

// SetWGPeersTotal sets the total number of peers.
func SetWGPeersTotal(n int) { wgPeersTotal.Set(float64(n)) }

// SetWGPeerStats sets the stats for a specific peer.
func SetWGPeerStats(pubKey string, rx, tx int64, lastHandshakeSeconds float64) {
	wgPeerRxBytes.WithLabelValues(pubKey).Set(float64(rx))
	wgPeerTxBytes.WithLabelValues(pubKey).Set(float64(tx))
	wgPeerLastHandshake.WithLabelValues(pubKey).Set(lastHandshakeSeconds)
}
