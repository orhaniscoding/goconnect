package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSqliteAuditor_Basic(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit.db")
	a, err := NewSqliteAuditor(dbPath)
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()

	ctx := context.Background()
	a.Event(ctx, "t1", ActionNetworkCreated, "alice", "net-1", map[string]any{"foo": "bar"})
	a.Event(ctx, "t1", ActionIPAllocated, "alice", "net-1", map[string]any{"ip": "10.0.0.2"})

	if c, err := a.Count(ctx); err != nil || c != 2 {
		t.Fatalf("expected 2 events count=%d err=%v", c, err)
	}
	evs, err := a.ListRecent(ctx, 10)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}
	if len(evs) != 2 {
		t.Fatalf("expected 2 events got %d", len(evs))
	}
	// No-op: we only assert that two events exist; order may vary.
}

func TestSqliteAuditor_Hashing(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit_hash.db")
	secret := []byte("super-secret")
	a, err := NewSqliteAuditor(dbPath, WithSqliteHashing(secret))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()
	ctx := context.Background()
	a.Event(ctx, "t1", ActionNetworkCreated, "bob", "net-9", nil)
	evs, err := a.ListRecent(ctx, 5)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event got %d", len(evs))
	}
	if evs[0].Actor == "bob" || evs[0].Actor == "[redacted]" {
		t.Fatalf("expected hashed actor got %s", evs[0].Actor)
	}
	// Re-open and ensure persistence
	a.Close()
	a2, err := NewSqliteAuditor(dbPath, WithSqliteHashing(secret))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer a2.Close()
	evs2, err := a2.ListRecent(ctx, 5)
	if err != nil {
		t.Fatalf("list2: %v", err)
	}
	if len(evs2) != 1 {
		t.Fatalf("expected 1 after reopen got %d", len(evs2))
	}
}

// Benchmark to gauge basic insert throughput (optional)
func BenchmarkSqliteAuditor_Event(b *testing.B) {
	dir := b.TempDir()
	dbPath := filepath.Join(dir, "bench.db")
	a, err := NewSqliteAuditor(dbPath)
	if err != nil {
		b.Fatalf("new auditor: %v", err)
	}
	defer a.Close()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Event(ctx, "t1", ActionIPAllocated, "bench-actor", "net", map[string]any{"i": i})
	}
	if _, err := os.Stat(dbPath); err != nil {
		b.Fatalf("db not created: %v", err)
	}
}
