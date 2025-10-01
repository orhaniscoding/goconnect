package audit

import (
    "context"
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "path/filepath"
    "testing"
)

// TestIntegrityExportSigning verifies that an integrity export is signed when a signing key is configured
// and that the signature validates over the canonical JSON payload (without signature field).
func TestIntegrityExportSigning(t *testing.T) {
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil { t.Fatalf("generate key: %v", err) }

    dir := t.TempDir()
    dbPath := filepath.Join(dir, "audit.db")
    a, err := NewSqliteAuditor(dbPath, WithIntegritySigningKey(priv))
    if err != nil { t.Fatalf("new auditor: %v", err) }
    defer a.Close()

    ctx := context.Background()
    a.Event(ctx, "ACTION", "actor-1", "object-1", map[string]any{"k":"v"})

    exp, err := a.ExportIntegrity(ctx, 5)
    if err != nil { t.Fatalf("export integrity: %v", err) }
    if exp.Signature == "" { t.Fatalf("expected signature present") }

    // Reconstruct canonical payload (export minus signature) and verify.
    expCopy := exp
    expCopy.Signature = ""
    payload, err := json.Marshal(expCopy)
    if err != nil { t.Fatalf("marshal copy: %v", err) }
    sigBytes, err := base64.RawURLEncoding.DecodeString(exp.Signature)
    if err != nil { t.Fatalf("decode signature: %v", err) }
    if !ed25519.Verify(pub, payload, sigBytes) {
        t.Fatalf("signature verification failed")
    }
}
