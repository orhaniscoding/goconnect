# Audit System Notes
- `ListRecent` & `Count` helpers for tests / diagnostics (not public API yet).
- Monotonic `seq` + high-resolution `ts` prepare for hash chain (`chain_hash = H(prev_chain_hash || canonical_event)`).
- CGO-free portability; future pragmas (WAL, busy_timeout) configurable.

Planned Enhancements:
- Config toggles for WAL & performance pragmas.
- Streaming iterator / pagination API.
- Integrity chain storage (anchor snapshots) + verification tool.
- Optional durability escalation: fail-fast for designated critical actions.

## Hash Chain (Planned Design Snapshot)
- Each persisted event will compute `chain_hash = H(prev_chain_hash || canonical_event_json)`.
- Head hash stored in a metadata table; periodic snapshot rows for O(log n) verification.
- Verification tool walks chain and reports first divergence.

---
Revision: v2 – Consolidated after hashing & pre-persistence merge
---
Revision: v2 – Consolidated after hashing & pre-persistence merge
Date: 2025-09-30

# Audit System Notes

## Overview
Current audit subsystem provides a lightweight, pluggable auditing layer: an in-memory append-only store (used in tests), a stdout JSON auditor (dev diagnostics), and a SQLite-backed auditor (persistence prototype). It captures security-relevant and lifecycle events across network, membership, and IPAM operations while intentionally redacting or hashing direct identifiers to avoid PII leakage.

## Goals
- Provide tamper-evident, append-only event trail (future: hash chain / signature)
- Avoid logging raw user or network identifiers directly (privacy & compliance)
- Support correlation via request_id and structured details
- Remain low overhead for hot paths

## Components
| Component                   | Purpose                                                                   |
| --------------------------- | ------------------------------------------------------------------------- |
| `Auditor` interface         | Pluggable sink abstraction (`Event(ctx, action, actor, object, details)`) |
| `stdoutAuditor`             | Dev-oriented JSON emitter (redacts or hashes actor/object)                |
| `InMemoryStore`             | Thread-safe slice-based store for tests (redacted or hashed fields)       |
| `SqliteAuditor` (prototype) | Durable persistence (optional hashing)                                    |
| Service `SetAuditor` hooks  | Allow injection without changing constructors                             |

## Event Schema
```
EventRecord {
  ts: RFC3339Nano UTC timestamp,
  action: string,
  actor: "[redacted]" | hashed_id,
  object: "[redacted]" | hashed_id,
  details: map[string]any (action-specific non-PII),
  request_id: optional request correlation id
}

// Hashing: When auditor/store constructed with a hashing secret (in-memory, stdout, sqlite),
// actor/object become deterministic HMAC-SHA256(secret, raw_id) truncated to first 18 bytes
// (144 bits) and Base64 URL encoded (≈24 chars). Otherwise literal "[redacted]".
```

## Current Actions (Non-Exhaustive)
- NETWORK_CREATED / NETWORK_UPDATED / NETWORK_DELETED
- NETWORK_JOIN_REQUEST / NETWORK_JOIN_APPROVE / NETWORK_JOIN_DENY
- NETWORK_MEMBER_KICK / NETWORK_MEMBER_BAN / NETWORK_MEMBER_UNBAN
- IP_ALLOCATED / IP_RELEASED

### Detail Conventions
- Names and raw IDs are NOT stored. Minimal operational metadata only (e.g., `{ "ip": "10.1.0.5" }`).
- Admin-triggered IP release includes `released_for` specifying (currently raw) target user id — temporary; will migrate to hashed token.

## Redaction & Hashing Roadmap
1. (DONE) Deterministic actor/object hashing (HMAC-SHA256, first 144 bits base64url) via optional constructors (`WithHashing(secret)`, `NewStdoutAuditorWithHashing(secret)`, `WithSqliteHashing(secret)`).
2. Configurable retention (in-memory ring buffer + SQLite pruning by time / row cap).
3. Structured exporter interface + multi-sink fan-out (stdout, channel, future: OpenTelemetry / webhook / file).
4. Tamper-evidence: hash chain (`event_chain_hash = H(prev_chain_hash || canonical_event_json)`) with persisted head and periodic anchor snapshots.
5. Backpressure & async buffering (bounded queue + worker) + reliability (retry with jitter).
6. Metrics (events/sec, failures, queue depth, insertion latency).
7. HMAC key rotation (dual-key window + forward-only correlation; old hashes remain non-reversible).

## Integrity & Ordering
- In-memory slice preserves insertion order; concurrency test ensures guarantees.
- SQLite uses monotonic `seq` (AUTOINCREMENT) + timestamp for deterministic ordering.
- Future hash chain will reference `seq` (or previous chain hash) to ensure tamper evidence.

## Performance Considerations
- In-memory: single mutex per append; acceptable at current scale (< few µs target per event).
- SQLite prototype: synchronous insert; future optimizations (WAL, batching, async) after exporter layer.
- Planned optimization: lock-free ring buffer or sharded mutex if event volume grows.

## Testing
- Concurrency test validates exact event count under parallel writes.
- Handler & service tests assert event presence for critical flows.
- (Planned) Cross-implementation hashing consistency test (same secret & raw id ⇒ identical 24-char token across all auditors).

## Future API Extensions
| Feature                               | Rationale                                       |
| ------------------------------------- | ----------------------------------------------- |
| Query API (time range, action filter) | Operational troubleshooting & compliance export |
| Export format (NDJSON streaming)      | Integrate with SIEM / log pipeline              |
| Action constants package              | Prevent typos & enable IDE completion           |
| Multi-sink fan-out with retries       | Reliability under sink outage                   |
| Privacy classification tags           | Data governance & selective retention           |
| Pagination / streaming iterator       | Efficient large export                          |

## Security Considerations
- No PII at rest: ensure `details` excludes user-supplied raw strings unless classified safe.
- Side-channel: IP values may reveal topology; consider masking last octet if policy evolves.
- Hash truncation (144 bits) collision risk negligible for projected volumes (<10^9 events) but monitored.
- Key rotation must maintain forward secrecy (old events not rehashable to raw IDs).

## Open Questions
- Per-tenant logical segregation or encryption-at-rest? (Backlog for multi-tenant maturity.)
- Should audit failures ever fail parent operation? (Current: best-effort; future: policy tier for critical actions.)
- Sampling for high-volume benign events? (Optional toggle.)
- Dedicated integrity verification command / endpoint?

## Immediate Next Steps
1. Unify hashing helper across all auditors.
2. Exporter interface + stdout + channel fan-out (foundation for async & multi-sink).
3. Retention policies (ring buffer + SQLite pruning) + metrics instrumentation.
4. Hash chain prototype (persist head hash, verification tool/test).
5. Key rotation framework (dual active secrets + migration tests).

## Persistence (SQLite Prototype)
Implemented `SqliteAuditor` (`internal/audit/sqlite.go`) using pure Go driver (`modernc.org/sqlite`). Schema:
```
CREATE TABLE IF NOT EXISTS audit_events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  action TEXT NOT NULL,
  actor TEXT NOT NULL,
  object TEXT NOT NULL,
  details TEXT,
  request_id TEXT
);
CREATE INDEX IF NOT EXISTS idx_audit_events_action_ts ON audit_events(action, ts);
```

Characteristics:
- Fire-and-forget insert inside `Event` (errors currently ignored; future: metric + optional escalation policy).
- Optional hashing mirrors in-memory/stdout truncation (Base64 URL, first 18 bytes of HMAC digest / 144 bits).
- `ListRecent` & `Count` helpers for tests / diagnostics (not public API yet).
- Monotonic `seq` + high-resolution `ts` prepare for hash chain.
- CGO-free portability; future pragmas (WAL, busy_timeout) configurable.

Planned Enhancements:
- Config toggles for WAL & performance pragmas.
- Streaming iterator / pagination API.
- Integrity chain storage (anchor snapshots) + verification tool.
- Optional durability escalation: fail-fast for designated critical actions.

## Hash Chain (Planned Design Snapshot)
- Each persisted event will compute `chain_hash = H(prev_chain_hash || canonical_event_json)`.
- Head hash stored in a metadata table; periodic snapshot rows for O(log n) verification.
- Verification tool walks chain and reports first divergence.
- Handler & service tests assert event presence for critical flows.
