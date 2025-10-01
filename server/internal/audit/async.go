package audit

import (
    "context"
    "errors"
    "sync"
    "time"

    "github.com/orhaniscoding/goconnect/server/internal/metrics"
)

// AsyncAuditor is a non-blocking wrapper that enqueues events and dispatches them via a worker.
// Drops events when queue is full (best-effort) and records metrics.
type AsyncAuditor struct {
    next    Auditor
    ch      chan queuedEvent
    stopCh  chan struct{}
    once    sync.Once
    wg      sync.WaitGroup
    started bool
}

type queuedEvent struct {
    enqueueTS time.Time
    ctx       context.Context
    action    string
    actor     string
    object    string
    details   map[string]any
}

// AsyncOption configures AsyncAuditor behavior.
type AsyncOption func(*asyncConfig)

type asyncConfig struct {
    queueSize int
    workers   int
}

// WithQueueSize sets the bounded queue size (default 1024).
func WithQueueSize(n int) AsyncOption { return func(c *asyncConfig) { if n > 0 { c.queueSize = n } } }

// WithWorkers sets number of dispatch workers (default 1).
func WithWorkers(n int) AsyncOption { return func(c *asyncConfig) { if n > 0 { c.workers = n } } }

// NewAsyncAuditor wraps an underlying auditor with an asynchronous queue.
func NewAsyncAuditor(next Auditor, opts ...AsyncOption) *AsyncAuditor {
    cfg := asyncConfig{queueSize: 1024, workers: 1}
    for _, o := range opts { o(&cfg) }
    a := &AsyncAuditor{
        next:   next,
        ch:     make(chan queuedEvent, cfg.queueSize),
        stopCh: make(chan struct{}),
    }
    a.start(cfg.workers)
    return a
}

func (a *AsyncAuditor) start(workers int) {
    a.once.Do(func() {
        a.started = true
        for i := 0; i < workers; i++ {
            a.wg.Add(1)
            go a.run()
        }
    })
}

// Event enqueues an event or drops it if the queue is full, updating metrics accordingly.
func (a *AsyncAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
    if !a.started { return }
    ev := queuedEvent{enqueueTS: time.Now(), ctx: ctx, action: action, actor: actor, object: object, details: details}
    select {
    case a.ch <- ev:
        metrics.SetAuditQueueDepth(len(a.ch))
    default:
        metrics.IncAuditDropped()
    }
}

func (a *AsyncAuditor) run() {
    defer a.wg.Done()
    for {
        select {
        case <-a.stopCh:
            // drain remaining events before exit
            for {
                select {
                case ev := <-a.ch:
                    a.dispatch(ev)
                default:
                    return
                }
            }
        case ev := <-a.ch:
            a.dispatch(ev)
        }
    }
}

func (a *AsyncAuditor) dispatch(ev queuedEvent) {
    metrics.ObserveAuditDispatch(time.Since(ev.enqueueTS).Seconds())
    a.next.Event(ev.ctx, ev.action, ev.actor, ev.object, ev.details)
    metrics.SetAuditQueueDepth(len(a.ch))
}

// Close gracefully stops workers, flushing queued events.
func (a *AsyncAuditor) Close() error {
    if !a.started { return errors.New("async auditor not started") }
    close(a.stopCh)
    a.wg.Wait()
    return nil
}