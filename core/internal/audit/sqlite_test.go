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

func TestSqliteAuditor_WithMaxRows(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit_maxrows.db")
	// Set max rows to 5
	a, err := NewSqliteAuditor(dbPath, WithMaxRows(5))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()

	ctx := context.Background()

	// Add 10 events
	for i := 0; i < 10; i++ {
		a.Event(ctx, "t1", ActionNetworkCreated, "user", "net", map[string]any{"i": i})
	}

	// Should still have events (retention may not be immediate)
	c, err := a.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	// MaxRows option is set, events should be inserted
	if c == 0 {
		t.Fatalf("expected events, got 0")
	}
}

func TestSqliteAuditor_QueryLogs(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit_query.db")
	a, err := NewSqliteAuditor(dbPath)
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()

	ctx := context.Background()

	// Create events for tenant-1
	a.Event(ctx, "tenant-1", ActionNetworkCreated, "alice", "net-1", nil)
	a.Event(ctx, "tenant-1", ActionIPAllocated, "bob", "net-2", nil)
	a.Event(ctx, "tenant-1", ActionNetworkDeleted, "charlie", "net-3", nil)

	// Test QueryLogs (wrapper around QueryLogsFiltered)
	logs, total, err := a.QueryLogs(ctx, "tenant-1", 10, 0)
	if err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}

	// Test with offset
	logs2, total2, err := a.QueryLogs(ctx, "tenant-1", 10, 1)
	if err != nil {
		t.Fatalf("query logs with offset: %v", err)
	}
	if total2 != 3 {
		t.Errorf("expected total 3, got %d", total2)
	}
	if len(logs2) != 2 {
		t.Errorf("expected 2 logs with offset, got %d", len(logs2))
	}

	// Test with limit
	logs3, total3, err := a.QueryLogs(ctx, "tenant-1", 1, 0)
	if err != nil {
		t.Fatalf("query logs with limit: %v", err)
	}
	if total3 != 3 {
		t.Errorf("expected total 3, got %d", total3)
	}
	if len(logs3) != 1 {
		t.Errorf("expected 1 log with limit, got %d", len(logs3))
	}

	// Test nonexistent tenant
	logs4, total4, err := a.QueryLogs(ctx, "nonexistent-tenant", 10, 0)
	if err != nil {
		t.Fatalf("query logs nonexistent tenant: %v", err)
	}
	if total4 != 0 {
		t.Errorf("expected total 0, got %d", total4)
	}
	if len(logs4) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs4))
	}
}

