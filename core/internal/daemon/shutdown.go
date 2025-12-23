package daemon

import (
	"context"
	"sync"

	"github.com/orhaniscoding/goconnect/server/internal/logger"
)

// CleanupTask represents a named cleanup function.
type CleanupTask struct {
	Name string
	Fn   func() error
}

// ShutdownManager handles registered cleanup tasks upon application exit.
type ShutdownManager struct {
	mu    sync.Mutex
	tasks []CleanupTask
}

// NewShutdownManager creates a new ShutdownManager.
func NewShutdownManager() *ShutdownManager {
	return &ShutdownManager{
		tasks: make([]CleanupTask, 0),
	}
}

// Register adds a new cleanup task to the manager.
func (s *ShutdownManager) Register(name string, fn func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, CleanupTask{Name: name, Fn: fn})
}

// Execute runs all registered cleanup tasks within the provided context's deadline.
func (s *ShutdownManager) Execute(ctx context.Context) {
	logger.Info("Executing cleanup tasks...")

	var wg sync.WaitGroup
	for _, task := range s.tasks {
		wg.Add(1)
		go func(t CleanupTask) {
			defer wg.Done()
			
			// We use a separate channel to detect function completion
			done := make(chan error, 1)
			go func() {
				done <- t.Fn()
			}()

			select {
			case err := <-done:
				if err != nil {
					logger.Error("Cleanup task failed", "task", t.Name, "error", err)
				} else {
					logger.Debug("Cleanup task succeeded", "task", t.Name)
				}
			case <-ctx.Done():
				logger.Error("Cleanup task timed out", "task", t.Name)
			}
		}(task)
	}

	// Wait for all goroutines to finish OR context timeout
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		logger.Info("All cleanup tasks completed")
	case <-ctx.Done():
		logger.Error("Shutdown process timed out")
	}
}
