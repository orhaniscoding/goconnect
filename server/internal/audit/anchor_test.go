package audit

import (
	"context"
	"testing"
)

func TestAnchorCreationAndPartialVerify(t *testing.T) {
	aud, err := NewSqliteAuditor(":memory:", WithAnchorInterval(2))
	if err != nil {
		t.Fatalf("new sqlite auditor: %v", err)
	}
	ctx := context.Background()
	// Insert 5 events -> anchors expected at seq 2 and 4
	for i := 0; i < 5; i++ {
		aud.Event(ctx, "t1", "ACT", "actor", "obj", map[string]any{"i": i})
	}
	anchors, err := aud.ListAnchors(ctx)
	if err != nil {
		t.Fatalf("list anchors: %v", err)
	}
	if len(anchors) != 2 {
		t.Fatalf("expected 2 anchors got %d", len(anchors))
	}
	if anchors[0].Seq != 2 || anchors[1].Seq != 4 {
		t.Fatalf("unexpected anchor seqs: %+v", anchors)
	}
	// Partial verify from second anchor
	if err := aud.VerifyFromAnchor(ctx, anchors[1].Seq); err != nil {
		t.Fatalf("partial verify failed: %v", err)
	}
}

func TestPartialVerifyMismatchAfterAnchor(t *testing.T) {
	aud, err := NewSqliteAuditor(":memory:", WithAnchorInterval(1))
	if err != nil {
		t.Fatalf("new sqlite auditor: %v", err)
	}
	ctx := context.Background()
	aud.Event(ctx, "t1", "A1", "actor", "obj", nil) // seq1 anchor
	aud.Event(ctx, "t1", "A2", "actor", "obj", nil) // seq2 anchor
	// tamper row 2 chain_hash
	_, err = aud.db.ExecContext(ctx, `UPDATE audit_events SET chain_hash='deadbeef' WHERE seq=2`)
	if err != nil {
		t.Fatalf("tamper: %v", err)
	}
	// Partial verify from seq2 should fail
	if err := aud.VerifyFromAnchor(ctx, 2); err == nil {
		t.Fatalf("expected mismatch from anchor seq2")
	}
}
