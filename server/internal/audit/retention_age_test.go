package audit

import (
	"context"
	"testing"
	"time"
)

func TestAgeRetentionPrunesOldEvents(t *testing.T) {
	aud, err := NewSqliteAuditor(":memory:", WithMaxAge(50*time.Millisecond))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	ctx := context.Background()
	aud.Event(ctx, "t1", "A1", "actor", "obj", nil)
	time.Sleep(60 * time.Millisecond)
	aud.Event(ctx, "t1", "A2", "actor", "obj", nil)
	// Trigger pruning by inserting another event
	aud.Event(ctx, "t1", "A3", "actor", "obj", nil)
	// Count should be 2 (A2, A3) if pruning worked
	c, err := aud.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if c != 2 {
		t.Fatalf("expected 2 events retained got %d", c)
	}
	if err := aud.VerifyChain(ctx); err != nil {
		t.Fatalf("verify after prune: %v", err)
	}
}

func TestAgeRetentionTamperDetected(t *testing.T) {
	aud, err := NewSqliteAuditor(":memory:", WithMaxAge(30*time.Millisecond))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	ctx := context.Background()
	aud.Event(ctx, "t1", "A1", "actor", "obj", nil)
	time.Sleep(40 * time.Millisecond)
	aud.Event(ctx, "t1", "A2", "actor", "obj", nil) // A1 should be pruned on next insert
	aud.Event(ctx, "t1", "A3", "actor", "obj", nil) // triggers pruning of A1
	// Tamper last event hash
	_, err = aud.db.ExecContext(ctx, `UPDATE audit_events SET chain_hash='deadbeef' WHERE seq=(SELECT MAX(seq) FROM audit_events)`)
	if err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := aud.VerifyChain(ctx); err == nil {
		t.Fatalf("expected verification failure after tamper")
	}
}
