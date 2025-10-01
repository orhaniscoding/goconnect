package audit

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/metrics"
)

// EventRecord is an in-memory representation of an audit event (PII redacted).
type EventRecord struct {
	TS        time.Time      `json:"ts"`
	Action    string         `json:"action"`
	Actor     string         `json:"actor"`
	Object    string         `json:"object"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

// InMemoryStore is a thread-safe in-memory Auditor implementation primarily for tests.
// It redacts actor/object similar to stdoutAuditor to ensure no PII is persisted.
type InMemoryStore struct {
	mu        sync.RWMutex
	events    []EventRecord
	hasher    func(data string) string
	redactAll bool
	capacity  int // 0 = unbounded; when >0 acts as ring buffer (drop oldest)
}

// Option configures store behavior.
type Option func(*InMemoryStore)

// WithHashing enables deterministic hashing of actor/object identifiers using provided secret.
// Uses HMAC-SHA256 and encodes first 18 bytes (to keep it short) in URL-safe base64 without padding.
func WithHashing(secret []byte) Option { // legacy single-secret API
	return func(s *InMemoryStore) {
		s.hasher = newHasher(secret)
		if s.hasher != nil { s.redactAll = false }
	}
}

// WithHashSecrets configures multiple hashing secrets (rotation). First secret is active for new events.
func WithHashSecrets(secrets ...[]byte) Option {
	return func(s *InMemoryStore) {
		h, _, _ := multiSecretHasher(secrets)
		if h != nil {
			s.hasher = h
			s.redactAll = false
		}
	}
}

// WithRedaction forces full redaction (default behavior) even if hashing enabled elsewhere.
func WithRedaction() Option { return func(s *InMemoryStore) { s.redactAll = true } }

// NewInMemoryStore creates a new in-memory audit store with optional functional options.
func NewInMemoryStore(opts ...Option) *InMemoryStore {
	store := &InMemoryStore{redactAll: true}
	for _, o := range opts {
		o(store)
	}
	return store
}

// Event implements Auditor interface (PII redacted).
func (s *InMemoryStore) Event(ctx context.Context, action, actor, object string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	rid, _ := ctx.Value("request_id").(string)
	act := "[redacted]"
	obj := "[redacted]"
	if !s.redactAll && s.hasher != nil {
		// hash raw inputs (still considered pseudonymous, prevents raw leakage)
		act = s.hasher(actor)
		obj = s.hasher(object)
	}
	rec := EventRecord{
		TS:        time.Now().UTC(),
		Action:    action,
		Actor:     act,
		Object:    obj,
		Details:   details,
		RequestID: rid,
	}
	s.mu.Lock()
	if s.capacity > 0 && len(s.events) >= s.capacity {
		// drop oldest (index 0) by re-slicing; avoid realloc by copy shift
		copy(s.events[0:], s.events[1:])
		s.events[len(s.events)-1] = rec
		metrics.AddAuditEviction("memory", 1)
	} else {
		s.events = append(s.events, rec)
	}
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

// WithCapacity sets a max number of events retained in-memory (ring buffer style).
func WithCapacity(n int) Option {
	return func(s *InMemoryStore) {
		if n > 0 {
			s.capacity = n
		}
	}
}
