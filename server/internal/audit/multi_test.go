package audit

import (
	"context"
	"testing"
)

type captureAuditor struct{ events []EventRecord }

func (c *captureAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	c.events = append(c.events, EventRecord{Action: action, Actor: actor, Object: object, Details: details})
}

func TestMultiAuditorFanOut(t *testing.T) {
	c1 := &captureAuditor{}
	c2 := &captureAuditor{}
	multi := NewMultiAuditor(c1, c2)
	multi.Event(context.Background(), "ACTION_X", "a", "o", map[string]any{"k": "v"})
	if len(c1.events) != 1 || len(c2.events) != 1 {
		t.Fatalf("expected both sinks to receive event")
	}
}
