package audit

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
)

// mockAuditor collects events synchronously for verification.
type mockAuditor struct {
	mu     sync.Mutex
	events []string
}

func (m *mockAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, action+":"+actor+":"+object)
}
func (m *mockAuditor) Len() int { m.mu.Lock(); defer m.mu.Unlock(); return len(m.events) }

func TestAsyncAuditorDispatch(t *testing.T) {
	base := &mockAuditor{}
	async := NewAsyncAuditor(base, WithQueueSize(10), WithWorkers(2))
	defer async.Close()
	async.Event(context.Background(), "ACTION1", "alice", "net1", nil)
	async.Event(context.Background(), "ACTION2", "bob", "net2", nil)
	// allow worker to process
	time.Sleep(30 * time.Millisecond)
	if base.Len() != 2 {
		t.Fatalf("expected 2 events, got %d", base.Len())
	}
}

func TestAsyncAuditorDropOnFull(t *testing.T) {
	base := &mockAuditor{}
	async := NewAsyncAuditor(base, WithQueueSize(1), WithWorkers(1))
	defer async.Close()
	// First enqueues
	async.Event(context.Background(), "A1", "u1", "o1", nil)
	// Immediately attempt many to overrun queue (some should drop)
	for i := 0; i < 50; i++ {
		async.Event(context.Background(), "A2", "u2", "o2", nil)
	}
	time.Sleep(40 * time.Millisecond)
	l := base.Len()
	if l == 0 {
		t.Fatalf("expected at least one event dispatched")
	}
	if l > 10 {
		t.Fatalf("unexpectedly large processed count suggests no drops: %d", l)
	}
}

func TestAsyncAuditorCloseFlushes(t *testing.T) {
	base := &mockAuditor{}
	async := NewAsyncAuditor(base, WithQueueSize(20), WithWorkers(1))
	for i := 0; i < 5; i++ {
		async.Event(context.Background(), "AX", "u", "o", nil)
	}
	if err := async.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	if base.Len() != 5 {
		t.Fatalf("expected 5 flushed events got %d", base.Len())
	}
}

// panicAuditor triggers panic on first event then no-op.
type panicAuditor struct {
	mu       sync.Mutex
	panicked bool
}

func (p *panicAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	p.mu.Lock()
	if !p.panicked {
		p.panicked = true
		p.mu.Unlock()
		panic("boom")
	}
	p.mu.Unlock()
}

// AuditorFunc adapter
type AuditorFunc func(ctx context.Context, action, actor, object string, details map[string]any)
func (f AuditorFunc) Event(ctx context.Context, action, actor, object string, details map[string]any) { f(ctx, action, actor, object, details) }

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func TestAsyncWorkerRestartMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics.Register()
	pa := &panicAuditor{}
	async := NewAsyncAuditor(pa, WithQueueSize(10), WithWorkers(1))
	defer async.Close()
	async.Event(context.Background(), "A", "actor", "obj", nil)
	time.Sleep(50 * time.Millisecond)
	async.Event(context.Background(), "B", "actor", "obj", nil)
	r := gin.New(); r.GET("/metrics", metrics.Handler())
	w := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if !contains(body, "goconnect_audit_worker_restarts_total 1") {
		// Allow some slack; if restart not yet recorded, wait briefly once
		time.Sleep(25 * time.Millisecond)
		r.ServeHTTP(w, req)
		body = w.Body.String()
		if !contains(body, "goconnect_audit_worker_restarts_total 1") {
								 t.Fatalf("expected worker restart metric=1 body=\n%s", body)
		}
	}
}

func TestAsyncHighWatermarkMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics.Register()
	slow := AuditorFunc(func(ctx context.Context, action, actor, object string, details map[string]any) {
		time.Sleep(5 * time.Millisecond)
	})
	async := NewAsyncAuditor(slow, WithQueueSize(30), WithWorkers(1))
	defer async.Close()
	for i := 0; i < 20; i++ { async.Event(context.Background(), "HW", "a", "o", nil) }
	time.Sleep(150 * time.Millisecond)
	r := gin.New(); r.GET("/metrics", metrics.Handler())
	w := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if !contains(body, "goconnect_audit_queue_high_watermark") {
			 t.Fatalf("expected high watermark metric present body=\n%s", body)
	}
}

func TestAsyncDroppedReasonFull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics.Register()
	noOp := AuditorFunc(func(context.Context, string, string, string, map[string]any) {})
	async := NewAsyncAuditor(noOp, WithQueueSize(1), WithWorkers(1))
	defer async.Close()
	async.Event(context.Background(), "A", "a", "o", nil)
	async.Event(context.Background(), "B", "a", "o", nil) // should drop
	r := gin.New(); r.GET("/metrics", metrics.Handler())
	w := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if !contains(body, `goconnect_audit_events_dropped_reason_total{reason="full"}`) {
		 t.Fatalf("expected dropped reason=full metric present body=\n%s", body)
	}
}
