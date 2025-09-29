package audit

import (
	"context"
	"sync"
	"testing"
)

// TestInMemoryStoreConcurrent verifies the store is safe under concurrent Event calls.
func TestInMemoryStoreConcurrent(t *testing.T) {
	store := NewInMemoryStore()
	const workers = 64
	const per = 50
	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < per; j++ {
				store.Event(ctx, "TEST_EVENT", "actor", "object", map[string]any{"i": id, "j": j})
			}
		}(i)
	}
	wg.Wait()
	events := store.List()
	want := workers * per
	if len(events) != want {
		t.Fatalf("expected %d events got %d", want, len(events))
	}
	// ensure redaction applied
	if events[0].Actor != "[redacted]" || events[0].Object != "[redacted]" {
		t.Fatalf("expected redacted fields, got actor=%s object=%s", events[0].Actor, events[0].Object)
	}
}
