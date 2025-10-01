package audit

import (
    "context"
    "testing"
    "time"
    "sync"
)

// mockAuditor collects events synchronously for verification.
type mockAuditor struct { 
    mu sync.Mutex
    events []string 
}
func (m *mockAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
    m.mu.Lock(); defer m.mu.Unlock()
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
    if base.Len() != 2 { t.Fatalf("expected 2 events, got %d", base.Len()) }
}

func TestAsyncAuditorDropOnFull(t *testing.T) {
    base := &mockAuditor{}
    async := NewAsyncAuditor(base, WithQueueSize(1), WithWorkers(1))
    defer async.Close()
    // First enqueues
    async.Event(context.Background(), "A1", "u1", "o1", nil)
    // Immediately attempt many to overrun queue (some should drop)
    for i:=0; i<50; i++ { async.Event(context.Background(), "A2", "u2", "o2", nil) }
    time.Sleep(40 * time.Millisecond)
    l := base.Len()
    if l == 0 { t.Fatalf("expected at least one event dispatched") }
    if l > 10 { t.Fatalf("unexpectedly large processed count suggests no drops: %d", l) }
}

func TestAsyncAuditorCloseFlushes(t *testing.T) {
    base := &mockAuditor{}
    async := NewAsyncAuditor(base, WithQueueSize(20), WithWorkers(1))
    for i:=0; i<5; i++ { async.Event(context.Background(), "AX", "u", "o", nil) }
    if err := async.Close(); err != nil { t.Fatalf("close error: %v", err) }
    if base.Len() != 5 { t.Fatalf("expected 5 flushed events got %d", base.Len()) }
}