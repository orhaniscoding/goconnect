package audit

import (
    "context"
    "sync"
    "time"
)

// EventRecord is an in-memory representation of an audit event (PII redacted).
type EventRecord struct {
    TS       time.Time         `json:"ts"`
    Action   string            `json:"action"`
    Actor    string            `json:"actor"`
    Object   string            `json:"object"`
    Details  map[string]any    `json:"details,omitempty"`
    RequestID string           `json:"request_id,omitempty"`
}

// InMemoryStore is a thread-safe in-memory Auditor implementation primarily for tests.
// It redacts actor/object similar to stdoutAuditor to ensure no PII is persisted.
type InMemoryStore struct {
    mu     sync.RWMutex
    events []EventRecord
}

// NewInMemoryStore creates a new in-memory audit store.
func NewInMemoryStore() *InMemoryStore { return &InMemoryStore{} }

// Event implements Auditor interface (PII redacted).
func (s *InMemoryStore) Event(ctx context.Context, action, actor, object string, details map[string]any) {
    if details == nil { details = map[string]any{} }
    rid, _ := ctx.Value("request_id").(string)
    rec := EventRecord{
        TS:       time.Now().UTC(),
        Action:   action,
        Actor:    "[redacted]",
        Object:   "[redacted]",
        Details:  details,
        RequestID: rid,
    }
    s.mu.Lock()
    s.events = append(s.events, rec)
    s.mu.Unlock()
}

// List returns a copy of stored events (for tests / diagnostics).
func (s *InMemoryStore) List() []EventRecord {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]EventRecord, len(s.events))
    copy(out, s.events)
    return out
}

// Clear removes all events (helper for tests).
func (s *InMemoryStore) Clear() {
    s.mu.Lock()
    s.events = nil
    s.mu.Unlock()
}
