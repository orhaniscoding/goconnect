package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Auditor defines a sink for audit events. PII must not be logged.
type Auditor interface {
	Event(ctx context.Context, action, actor, object string, details map[string]any)
}

type stdoutAuditor struct{}

// NewStdoutAuditor returns a simple JSON stdout auditor suitable for dev/testing.
func NewStdoutAuditor() Auditor { return &stdoutAuditor{} }

func (s *stdoutAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	// Redact PII: don't emit raw actor/object identifiers directly
	if details == nil {
		details = map[string]any{}
	}
	rid, _ := ctx.Value("request_id").(string)
	payload := map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339Nano),
		"request_id": rid,
		"action":     action,
		"actor":      "[redacted]",
		"object":     "[redacted]",
		"details":    details,
	}
	b, _ := json.Marshal(payload)
	fmt.Println(string(b))
}
