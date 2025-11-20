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
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit.db")
	a, err := NewSqliteAuditor(dbPath, WithIntegritySigningKey(priv))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()

	ctx := context.Background()
	a.Event(ctx, "t1", "ACTION", "actor-1", "object-1", map[string]any{"k": "v"})

	exp, err := a.ExportIntegrity(ctx, 5)
	if err != nil {
		t.Fatalf("export integrity: %v", err)
	}
	if exp.Signature == "" {
		t.Fatalf("expected signature present")
	}

	// Reconstruct canonical payload (export minus signature) and verify.
	expCopy := exp
	expCopy.Signature = ""
	payload, err := json.Marshal(expCopy)
	if err != nil {
		t.Fatalf("marshal copy: %v", err)
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(exp.Signature)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if !ed25519.Verify(pub, payload, sigBytes) {
		t.Fatalf("signature verification failed")
	}
}

func TestIntegrityExportSigningWithKeyID(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "audit_kid.db")
	kid := "key-2025-10-01"
	a, err := NewSqliteAuditor(dbPath, WithIntegritySigningKeyID(kid, priv))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	defer a.Close()
	ctx := context.Background()
	a.Event(ctx, "t1", "ACTION", "actor-1", "object-1", map[string]any{"k": "v"})
	exp, err := a.ExportIntegrity(ctx, 5)
	if err != nil {
		t.Fatalf("export integrity: %v", err)
	}
	if exp.KeyID != kid {
		t.Fatalf("expected kid %s got %s", kid, exp.KeyID)
	}
	if exp.Signature == "" {
		t.Fatalf("expected signature present")
	}
	// verify signature includes kid (already part of payload)
	copyExp := exp
	copyExp.Signature = ""
	payload, err := json.Marshal(copyExp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(exp.Signature)
	if err != nil {
		t.Fatalf("decode sig: %v", err)
	}
	if !ed25519.Verify(pub, payload, sigBytes) {
		t.Fatalf("signature verify failed with kid payload")
	}
}
