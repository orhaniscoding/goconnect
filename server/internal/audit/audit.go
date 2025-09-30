package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"time"
)

// Auditor defines a sink for audit events. PII must not be logged.
type Auditor interface {
	Event(ctx context.Context, action, actor, object string, details map[string]any)
}

type stdoutAuditor struct{ hasher func(string) string }

// NewStdoutAuditor returns a simple JSON stdout auditor suitable for dev/testing.
func NewStdoutAuditor() Auditor { return &stdoutAuditor{} }

// NewStdoutAuditorWithHashing returns a stdout auditor that hashes actor/object
// identifiers using HMAC-SHA256 with the provided secret (pseudonymous).
func NewStdoutAuditorWithHashing(secret []byte) Auditor {
	if len(secret) == 0 {
		return NewStdoutAuditor()
	}
	var h hash.Hash
	return &stdoutAuditor{hasher: func(v string) string {
		h = hmac.New(sha256.New, secret)
		_, _ = h.Write([]byte(v))
		sum := h.Sum(nil)
		return base64.RawURLEncoding.EncodeToString(sum[:18])
	}}
}

func (s *stdoutAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	// Redact PII: don't emit raw actor/object identifiers directly
	if details == nil {
		details = map[string]any{}
	}
	rid, _ := ctx.Value("request_id").(string)
	actorOut := "[redacted]"
	objectOut := "[redacted]"
	if s.hasher != nil {
		actorOut = s.hasher(actor)
		objectOut = s.hasher(object)
	}
	payload := map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339Nano),
		"request_id": rid,
		"action":     action,
		"actor":      actorOut,
		"object":     objectOut,
		"details":    details,
	}
	b, _ := json.Marshal(payload)
	fmt.Println(string(b))
}
