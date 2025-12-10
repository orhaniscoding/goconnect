package audit

import (
	"context"
	"strings"
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
				store.Event(ctx, "t1", "TEST_EVENT", "actor", "object", map[string]any{"i": id, "j": j})
			}
		}(i)
	}
	wg.Wait()
	events := store.List()
	want := workers * per
	if len(events) != want {
		t.Fatalf("expected %d events got %d", want, len(events))
	}
	if events[0].Actor != "[redacted]" || events[0].Object != "[redacted]" {
		t.Fatalf("expected redacted fields, got actor=%s object=%s", events[0].Actor, events[0].Object)
	}
}

func TestInMemoryStoreHashing(t *testing.T) {
	secret := []byte("test-secret-key")
	store := NewInMemoryStore(WithHashing(secret))
	ctx := context.Background()
	store.Event(ctx, "t1", "ACTION", "userA", "net1", nil)
	store.Event(ctx, "t1", "ACTION", "userA", "net1", nil)
	store.Event(ctx, "t1", "ACTION", "userB", "net1", nil)
	evs := store.List()
	if len(evs) != 3 {
		t.Fatalf("expected 3 events")
	}
	a1 := evs[0].Actor
	a2 := evs[1].Actor
	a3 := evs[2].Actor
	if a1 != a2 {
		t.Fatalf("same actor should hash identically: %s vs %s", a1, a2)
	}
	if a1 == a3 {
		t.Fatalf("different actors should have different hashes")
	}
	if strings.Contains(a1, "redacted") {
		t.Fatalf("expected hashed value, got redacted")
	}
	if len(a1) != 24 { // 18 bytes base64url => 24 chars
		t.Fatalf("unexpected hash length: %d", len(a1))
	}
}

func TestInMemoryStore_Clear(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Add some events
	store.Event(ctx, "t1", "ACTION1", "actor1", "object1", nil)
	store.Event(ctx, "t1", "ACTION2", "actor2", "object2", nil)
	store.Event(ctx, "t1", "ACTION3", "actor3", "object3", nil)

	// Verify events are stored
	events := store.List()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Clear events
	store.Clear()

	// Verify all events are removed
	events = store.List()
	if len(events) != 0 {
		t.Fatalf("expected 0 events after Clear, got %d", len(events))
	}
}

func TestInMemoryStore_WithRedaction(t *testing.T) {
	// Create store with redaction enabled
	store := NewInMemoryStore(WithRedaction())
	ctx := context.Background()

	store.Event(ctx, "t1", "ACTION", "user123", "resource456", nil)

	events := store.List()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// With redaction, actor and object should be [redacted]
	if events[0].Actor != "[redacted]" {
		t.Fatalf("expected [redacted], got %s", events[0].Actor)
	}
	if events[0].Object != "[redacted]" {
		t.Fatalf("expected [redacted], got %s", events[0].Object)
	}
}
