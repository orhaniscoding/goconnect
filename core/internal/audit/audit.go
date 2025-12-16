package audit

import (
	"context"
	"log/slog"
)

// Auditor defines a sink for audit events. PII must not be logged.
type Auditor interface {
	Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any)
}

type stdoutAuditor struct{ hasher func(string) string }

// NewStdoutAuditor returns a simple JSON stdout auditor suitable for dev/testing.
func NewStdoutAuditor() Auditor { return &stdoutAuditor{} }

// NewStdoutAuditorWithHashing returns a stdout auditor that hashes actor/object
// identifiers using HMAC-SHA256 with the provided secret (pseudonymous).
func NewStdoutAuditorWithHashing(secret []byte) Auditor { // legacy single-secret helper
	if len(secret) == 0 {
		return NewStdoutAuditor()
	}
	return &stdoutAuditor{hasher: newHasher(secret)}
}

// NewStdoutAuditorWithHashSecrets supports key rotation (first secret active).
func NewStdoutAuditorWithHashSecrets(secrets ...[]byte) Auditor {
	if len(secrets) == 0 || len(secrets[0]) == 0 {
		return NewStdoutAuditor()
	}
	h, _, _ := multiSecretHasher(secrets)
	return &stdoutAuditor{hasher: h}
}

func (s *stdoutAuditor) Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
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

	slog.Info("AUDIT_EVENT",
		"request_id", rid,
		"tenant_id", tenantID,
		"action", action,
		"actor", actorOut,
		"object", objectOut,
		"details", details,
	)
}

// metricsAuditor wraps another Auditor to collect per-action counts.
type metricsAuditor struct {
	next Auditor
	inc  func(action string)
}

func (m *metricsAuditor) Event(ctx context.Context, tenantID, action, actor, object string, details map[string]any) {
	if m.inc != nil {
		m.inc(action)
	}
	m.next.Event(ctx, tenantID, action, actor, object, details)
}

// WrapWithMetrics decorates the provided auditor with a counter increment callback.
func WrapWithMetrics(base Auditor, inc func(action string)) Auditor {
	return &metricsAuditor{next: base, inc: inc}
}
