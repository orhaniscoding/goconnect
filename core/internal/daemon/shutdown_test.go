package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShutdownManager_Execute(t *testing.T) {
	t.Run("successful_tasks", func(t *testing.T) {
		sm := NewShutdownManager()
		var count int32

		sm.Register("task1", func() error {
			atomic.AddInt32(&count, 1)
			return nil
		})
		sm.Register("task2", func() error {
			atomic.AddInt32(&count, 1)
			return nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sm.Execute(ctx)
		// Small delay to ensure goroutines complete after Execute returns
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, int32(2), atomic.LoadInt32(&count))
	})

	t.Run("failing_task", func(t *testing.T) {
		sm := NewShutdownManager()
		sm.Register("fail_task", func() error {
			return errors.New("boom")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Should not panic or hang
		sm.Execute(ctx)
	})

	t.Run("timeout_task", func(t *testing.T) {
		sm := NewShutdownManager()
		var completed int32

		sm.Register("slow_task", func() error {
			time.Sleep(500 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil
		})

		// Set a very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		sm.Execute(ctx)

		// Task should not have completed because of short timeout
		assert.Equal(t, int32(0), atomic.LoadInt32(&completed))
	})
}
