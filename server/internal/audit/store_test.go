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
	if events[0].Actor != "[redacted]" || events[0].Object != "[redacted]" {
		t.Fatalf("expected redacted fields, got actor=%s object=%s", events[0].Actor, events[0].Object)
	}
}

func TestInMemoryStoreHashing(t *testing.T) {
	secret := []byte("test-secret-key")
	store := NewInMemoryStore(WithHashing(secret))
	ctx := context.Background()
	store.Event(ctx, "ACTION", "userA", "net1", nil)
	store.Event(ctx, "ACTION", "userA", "net1", nil)
	store.Event(ctx, "ACTION", "userB", "net1", nil)
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
