package audit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/metrics"
)

// AsyncAuditor is a non-blocking wrapper that enqueues events and dispatches them via a worker.
// Drops events when queue is full (best-effort) and records metrics.
type AsyncAuditor struct {
	next     Auditor
	ch       chan queuedEvent
	stopCh   chan struct{}
	once     sync.Once
	wg       sync.WaitGroup
	started  bool
	maxDepth int   // track high watermark
	inFlight int64 // queued + currently processing
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
func WithQueueSize(n int) AsyncOption {
	return func(c *asyncConfig) {
		if n > 0 {
			c.queueSize = n
		}
	}
}

// WithWorkers sets number of dispatch workers (default 1).
func WithWorkers(n int) AsyncOption {
	return func(c *asyncConfig) {
		if n > 0 {
			c.workers = n
		}
	}
}

// NewAsyncAuditor wraps an underlying auditor with an asynchronous queue.
func NewAsyncAuditor(next Auditor, opts ...AsyncOption) *AsyncAuditor {
	cfg := asyncConfig{queueSize: 1024, workers: 1}
	for _, o := range opts {
		o(&cfg)
	}
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
			a.launchWorker()
		}
	})
}

func (a *AsyncAuditor) launchWorker() {
	a.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				metrics.IncAuditWorkerRestart()
				// Relaunch replacement worker separately to avoid nested wg misuse
				go a.launchWorker()
			}
			a.wg.Done()
		}()
		a.run()
	}()
}

// Event enqueues an event or drops it if the queue is full, updating metrics accordingly.
func (a *AsyncAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	if !a.started {
		return
	}
	ev := queuedEvent{enqueueTS: time.Now(), ctx: ctx, action: action, actor: actor, object: object, details: details}
	// Treat the system as full when the number of in-flight events (queued + currently
	// processing) is already at capacity. This prevents bursts from overwhelming the
	// underlying auditor even if the worker consumes quickly.
	if atomic.LoadInt64(&a.inFlight) >= int64(cap(a.ch)) {
		metrics.IncAuditDropped()
		metrics.IncAuditDroppedReason("full")
		return
	}
	select {
	case a.ch <- ev:
		atomic.AddInt64(&a.inFlight, 1)
		cur := len(a.ch)
		metrics.SetAuditQueueDepth(cur)
		if cur > a.maxDepth {
			a.maxDepth = cur
			metrics.SetAuditQueueHighWatermark(cur)
		}
	default:
		metrics.IncAuditDropped()
		metrics.IncAuditDroppedReason("full")
	}
}

func (a *AsyncAuditor) run() {
	for {
		select {
		case <-a.stopCh:
			// drain remaining events before exit
			for {
				select {
				case ev := <-a.ch:
					a.dispatch(ev)
				default:
					// mark shutdown drops for any still queued (should be none after drain loop)
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
	atomic.AddInt64(&a.inFlight, -1)
}

// Close gracefully stops workers, flushing queued events.
func (a *AsyncAuditor) Close() error {
	if !a.started {
		return errors.New("async auditor not started")
	}
	close(a.stopCh)
	a.wg.Wait()
	return nil
}
