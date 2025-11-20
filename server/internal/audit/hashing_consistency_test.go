package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestHashingConsistency ensures that hashing output between InMemoryStore and SqliteAuditor
// is identical given the same secret and raw identifiers.
func TestHashingConsistency(t *testing.T) {
	secret := []byte("consistency-secret")
	mem := NewInMemoryStore(WithHashing(secret))

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit.db")
	sqliteAud, err := NewSqliteAuditor(dbPath, WithSqliteHashing(secret))
	if err != nil {
		t.Fatalf("failed to create sqlite auditor: %v", err)
	}
	t.Cleanup(func() { _ = sqliteAud.Close(); _ = os.Remove(dbPath) })

	ctx := context.Background()
	actor := "user-123"
	object := "network-456"

	// Emit events
	mem.Event(ctx, "t1", "ACTION", actor, object, nil)
	sqliteAud.Event(ctx, "t1", "ACTION", actor, object, nil)

	memEvents := mem.List()
	if len(memEvents) != 1 {
		t.Fatalf("expected 1 mem event, got %d", len(memEvents))
	}

	sqlEvents, err := sqliteAud.ListRecent(ctx, 1)
	if err != nil {
		t.Fatalf("failed listing sqlite events: %v", err)
	}
	if len(sqlEvents) != 1 {
		t.Fatalf("expected 1 sqlite event, got %d", len(sqlEvents))
	}

	mh := memEvents[0].Actor
	sh := sqlEvents[0].Actor
	if mh != sh {
		t.Fatalf("hash mismatch mem=%s sqlite=%s", mh, sh)
	}
	if len(mh) != 24 { // 18 bytes base64url -> 24 chars
		t.Fatalf("unexpected hash length: %d", len(mh))
	}
}
