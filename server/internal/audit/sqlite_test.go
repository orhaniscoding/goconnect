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

func TestSqliteAuditor_QueryLogsFiltered(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit_filter.db")
	a, err := NewSqliteAuditor(dbPath)
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()

	ctx := context.Background()

	// Create test events
	// Note: Without hashing, actor/object are stored as "[redacted]"
	// So we can only filter by action and tenant_id reliably without hashing
	a.Event(ctx, "tenant-1", ActionNetworkCreated, "alice", "net-1", nil)
	a.Event(ctx, "tenant-1", ActionIPAllocated, "bob", "net-1", nil)
	a.Event(ctx, "tenant-1", ActionNetworkCreated, "charlie", "net-2", nil)
	a.Event(ctx, "tenant-2", ActionNetworkCreated, "david", "net-3", nil)

	tests := []struct {
		name          string
		tenantID      string
		filter        AuditFilter
		expectedCount int
	}{
		{
			name:          "no filter - tenant-1",
			tenantID:      "tenant-1",
			filter:        AuditFilter{},
			expectedCount: 3,
		},
		{
			name:          "filter by actor - redacted (all actors redacted without hasher)",
			tenantID:      "tenant-1",
			filter:        AuditFilter{Actor: "[redacted]"},
			expectedCount: 3, // All actors are [redacted] when no hasher is set
		},
		{
			name:          "filter by action - NETWORK_CREATED",
			tenantID:      "tenant-1",
			filter:        AuditFilter{Action: ActionNetworkCreated},
			expectedCount: 2,
		},
		{
			name:          "filter by action - IP_ALLOCATED",
			tenantID:      "tenant-1",
			filter:        AuditFilter{Action: ActionIPAllocated},
			expectedCount: 1,
		},
		{
			name:          "different tenant",
			tenantID:      "tenant-2",
			filter:        AuditFilter{},
			expectedCount: 1,
		},
		{
			name:          "nonexistent actor",
			tenantID:      "tenant-1",
			filter:        AuditFilter{Actor: "nonexistent"},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, total, err := a.QueryLogsFiltered(ctx, tt.tenantID, tt.filter, 100, 0)
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			if total != tt.expectedCount {
				t.Errorf("expected total %d, got %d", tt.expectedCount, total)
			}
			if len(logs) != tt.expectedCount {
				t.Errorf("expected %d logs, got %d", tt.expectedCount, len(logs))
			}
		})
	}
}
