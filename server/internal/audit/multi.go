package audit

import "context"

// MultiAuditor fans out events to multiple underlying auditors.
// Failures in one sink do not stop others (best-effort: current interface has no error).
type MultiAuditor struct {
    sinks []Auditor
}

// NewMultiAuditor constructs a fan-out auditor.
func NewMultiAuditor(sinks ...Auditor) *MultiAuditor { return &MultiAuditor{sinks: sinks} }

// Event dispatches the event to all underlying sinks.
func (m *MultiAuditor) Event(ctx context.Context, action, actor, object string, details map[string]any) {
    for _, s := range m.sinks {
        if s != nil {
            s.Event(ctx, action, actor, object, details)
        }
    }
}
