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
Date: 2025-09-30 (amended: +retention +multi-sink)

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

## Redaction, Hashing & Export Roadmap
1. (DONE) Deterministic actor/object hashing (HMAC-SHA256, first 144 bits base64url) unified helper.
2. (PARTIAL DONE) In-memory retention via ring buffer (`WithCapacity(n)`) + (DONE) SQLite row-cap pruning (`WithMaxRows`) with labeled eviction metrics.
3. (DONE) Multi-sink fan-out (`MultiAuditor`) – foundation for exporter layer.
4. (PARTIAL DONE) Metrics: HTTP req counters/histograms + audit event counter + eviction & labeled failure counters + insert latency histogram + async queue metrics (depth, drops, dispatch latency) + (NEW) chain metrics (head advance, verification duration, failures).
5. (DONE) Tamper-evidence (phase 1): per-row `chain_hash = H(prev_chain_hash || canonical_event_json)` persisted inline.
6. (DONE) Tamper-evidence (phase 2): periodic anchor snapshots every N events (configurable) stored in `audit_chain_anchors(seq, ts, chain_hash)` enabling partial verification windows (O(k) after anchor vs full O(n)). New metric: `goconnect_audit_chain_anchor_created_total`.
6. Backpressure & async buffering (bounded queue + worker) + reliability (retry with jitter / dead-letter).
7. HMAC key rotation (dual-key window + forward-only correlation; old hashes non-reversible).

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

## Immediate Next Steps (Updated)
1. (DONE) SQLite time-based pruning variant (age cutoff) & dual-mode retention policy (row AND/OR age). Age pruning occurs post-insert; prunes events older than configurable duration, emits eviction metrics, removes orphan anchors. Verification treats the first retained event as a new baseline (empty previous hash) while still detecting tampering in the retained window.
2. Async buffering enhancements: backpressure policy (adaptive drop / priority), worker restart counter, enqueue failure reasons.
3. Integrity export: endpoint / tool to emit current head + recent anchors for remote verification.
4. Exporter integration & remote verification endpoint (serve current head + last anchor) + integrity alerting hook.
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
