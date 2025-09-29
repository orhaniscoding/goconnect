package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"sync"
	"time"
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
}
// Option configures store behavior.
type Option func(*InMemoryStore)

// WithHashing enables deterministic hashing of actor/object identifiers using provided secret.
// Uses HMAC-SHA256 and encodes first 18 bytes (to keep it short) in URL-safe base64 without padding.
func WithHashing(secret []byte) Option {
	return func(s *InMemoryStore) {
		if len(secret) == 0 { return }
		var h hash.Hash
		s.hasher = func(data string) string {
			h = hmac.New(sha256.New, secret)
			_, _ = h.Write([]byte(data))
			sum := h.Sum(nil)
			// truncate for readability (18 bytes ~ 144 bits) still collision-resistant for our scale
			truncated := sum[:18]
			enc := base64.RawURLEncoding.EncodeToString(truncated)
			return enc
		}
		s.redactAll = false
	}
}

// WithRedaction forces full redaction (default behavior) even if hashing enabled elsewhere.
func WithRedaction() Option { return func(s *InMemoryStore) { s.redactAll = true } }

// NewInMemoryStore creates a new in-memory audit store with optional functional options.
func NewInMemoryStore(opts ...Option) *InMemoryStore {
	store := &InMemoryStore{redactAll: true}
	for _, o := range opts { o(store) }
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
