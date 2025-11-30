package audit

import (
	"context"
	"testing"
)

func TestInMemoryStoreCapacityEvictsOldest(t *testing.T) {
	store := NewInMemoryStore(WithCapacity(3))
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		store.Event(ctx, "t1", "EVT", "actor", "obj", map[string]any{"i": i})
	}
	evs := store.List()
	if len(evs) != 3 {
		t.Fatalf("expected 3 retained events, got %d", len(evs))
	}
	// Oldest two should be evicted; remaining should have details i=2,3,4
	for idx, expected := range []int{2, 3, 4} {
		if evs[idx].Details["i"].(int) != expected {
			t.Fatalf("expected detail i=%d at idx=%d, got %v", expected, idx, evs[idx].Details["i"])
		}
	}
}
