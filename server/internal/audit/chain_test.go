package audit

import (
    "context"
    "database/sql"
    "testing"
)

func TestHashChainVerificationSuccess(t *testing.T) {
    aud, err := NewSqliteAuditor(":memory:")
    if err != nil { t.Fatalf("new sqlite auditor: %v", err) }
    ctx := context.Background()
    aud.Event(ctx, "A1", "actor", "obj", map[string]any{"k":1})
    aud.Event(ctx, "A2", "actor", "obj", map[string]any{"k":2})
    if err := aud.VerifyChain(ctx); err != nil { t.Fatalf("expected success verify, got %v", err) }
}

func TestHashChainVerificationMismatch(t *testing.T) {
    aud, err := NewSqliteAuditor(":memory:")
    if err != nil { t.Fatalf("new sqlite auditor: %v", err) }
    ctx := context.Background()
    aud.Event(ctx, "A1", "actor", "obj", nil)
    aud.Event(ctx, "A2", "actor", "obj", nil)
    // Tamper: update chain_hash of first row
    _, err = aud.db.ExecContext(ctx, `UPDATE audit_events SET chain_hash='deadbeef' WHERE seq=1`)
    if err != nil { t.Fatalf("tamper update: %v", err) }
    if err := aud.VerifyChain(ctx); err == nil { t.Fatalf("expected mismatch error, got nil") }
}

// Ensure chain_hash column exists for legacy migration scenario
func TestHashChainColumnExists(t *testing.T) {
    aud, err := NewSqliteAuditor(":memory:")
    if err != nil { t.Fatalf("new sqlite auditor: %v", err) }
    row := aud.db.QueryRow(`PRAGMA table_info(audit_events)`)
    // just check at least one row returns without error and later query specifically for chain_hash
    var cid int; var name, ctype string; var notnull, pk int; var dflt sql.NullString
    _ = row.Scan(&cid,&name,&ctype,&notnull,&dflt,&pk)
    // verify chain_hash column present
    rows, err := aud.db.Query(`PRAGMA table_info(audit_events)`)
    if err != nil { t.Fatalf("pragma table_info: %v", err) }
    defer rows.Close()
    found := false
    for rows.Next() {
        var cid2 int; var name2, ctype2 string; var nn2, pk2 int; var dflt2 sql.NullString
        if err := rows.Scan(&cid2,&name2,&ctype2,&nn2,&dflt2,&pk2); err != nil { t.Fatalf("scan: %v", err) }
        if name2 == "chain_hash" { found = true; break }
    }
    if !found { t.Fatalf("expected chain_hash column present") }
}