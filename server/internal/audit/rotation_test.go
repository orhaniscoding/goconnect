package audit

import (
    "context"
    "os"
    "path/filepath"
    "testing"
)

// TestKeyRotationDualSecrets ensures that when secrets are rotated the first (new) secret
// produces different hashes for new events while historical events hashed with the old secret
// remain stable. We simulate this by creating events with secretA, then re-opening with
// [secretB, secretA] and emitting another event. We expect the first event hash to differ
// from the second (since different active secret) and that re-computing the old hash using
// legacy single-secret API matches the stored historical value.
func TestKeyRotationDualSecrets(t *testing.T) {
    secretA := []byte("rot-secret-A")
    secretB := []byte("rot-secret-B")

    // InMemoryStore scenario
    memLegacy := NewInMemoryStore(WithHashing(secretA))
    ctx := context.Background()
    memLegacy.Event(ctx, "ACTION", "actor-1", "object-1", nil)
    if len(memLegacy.List()) != 1 { t.Fatalf("expected 1 legacy event") }
    oldHash := memLegacy.List()[0].Actor

    // Recreate store with rotated secrets (B active, A fallback conceptually)
    memRotated := NewInMemoryStore(WithHashSecrets(secretB, secretA))
    memRotated.Event(ctx, "ACTION", "actor-1", "object-1", nil)
    if len(memRotated.List()) != 1 { t.Fatalf("expected 1 rotated event") }
    newHash := memRotated.List()[0].Actor
    if oldHash == newHash { t.Fatalf("expected different hashes after rotation, got same %s", oldHash) }

    // Validate deterministic hashing for old secret still matches stored historical event
    singleHasher := newHasher(secretA)
    if singleHasher("actor-1") != oldHash { t.Fatalf("rehash with old secret mismatch") }
    if newHasher(secretB)("actor-1") != newHash { t.Fatalf("rehash with new secret mismatch") }

    // Sqlite scenario
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "audit.db")
    sqlLegacy, err := NewSqliteAuditor(dbPath, WithSqliteHashing(secretA))
    if err != nil { t.Fatalf("sqlite legacy: %v", err) }
    defer sqlLegacy.Close()
    sqlLegacy.Event(ctx, "ACTION", "actor-1", "object-1", nil)
    eventsA, err := sqlLegacy.ListRecent(ctx, 1)
    if err != nil || len(eventsA) != 1 { t.Fatalf("expected 1 legacy sqlite event: %v", err) }
    legacyHash := eventsA[0].Actor

    // Re-open with rotated secrets
    sqlRotated, err := NewSqliteAuditor(dbPath, WithSqliteHashSecrets(secretB, secretA))
    if err != nil { t.Fatalf("sqlite rotated: %v", err) }
    defer sqlRotated.Close()
    sqlRotated.Event(ctx, "ACTION", "actor-1", "object-1", nil)
    eventsB, err := sqlRotated.ListRecent(ctx, 2)
    if err != nil || len(eventsB) != 2 { t.Fatalf("expected 2 events after rotation: %v", err) }
    // eventsB[0] is newest (uses new secret), eventsB[1] is historical (old secret)
    newestHash := eventsB[0].Actor
    historicalHash := eventsB[1].Actor
    if historicalHash != legacyHash { t.Fatalf("historical hash changed; expected %s got %s", legacyHash, historicalHash) }
    if newestHash == legacyHash { t.Fatalf("expected new hash after rotation, still %s", newestHash) }

    // Re-hash validations
    if newHasher(secretA)("actor-1") != legacyHash { t.Fatalf("old secret rehash mismatch sqlite") }
    if newHasher(secretB)("actor-1") != newestHash { t.Fatalf("new secret rehash mismatch sqlite") }

    // Cleanup
    _ = os.Remove(dbPath)
}
