# Audit System Notes

## Overview
Current audit subsystem provides a lightweight, in-memory, append-only store used in tests plus a stdout JSON auditor for development. It captures security-relevant and lifecycle events across network, membership, and IPAM operations while intentionally redacting direct identifiers to avoid PII leakage.

## Goals
- Provide tamper-evident, append-only event trail (future: hash chain / signature)
- Avoid logging raw user or network identifiers directly (privacy & compliance)
- Support correlation via request_id and structured details
- Remain low overhead for hot paths

## Components
| Component                  | Purpose                                                                    |
| -------------------------- | -------------------------------------------------------------------------- |
| `Auditor` interface        | Pluggable sink abstraction (`Event(ctx, action, actor, object, details)` ) |
| `stdoutAuditor`            | Dev-oriented JSON emitter (redacts actor/object)                           |
| `InMemoryStore`            | Thread-safe slice-based store for tests (redacted fields)                  |
| Service `SetAuditor` hooks | Allow injection without changing constructors                              |

## Event Schema
```
EventRecord {
  ts: RFC3339Nano UTC timestamp,
  action: string (ENUM-like constant),
  actor: "[redacted]" | hashed_id,  // hashed when hashing enabled (HMAC-SHA256 truncated)
  object: "[redacted]" | hashed_id, // ditto
  details: map[string]any (action-specific non-PII),
  request_id: optional request correlation id
}
```

## Current Actions (Non-Exhaustive)
- NETWORK_CREATED / NETWORK_UPDATED / NETWORK_DELETED
- NETWORK_JOIN_APPROVE / NETWORK_MEMBER_BAN (membership lifecycle)
- IP_ALLOCATED / IP_RELEASED

### Detail Conventions
- Names and raw IDs are NOT stored. Minimal operational metadata only (e.g., `{ "ip": "10.1.0.5" }`).
- Admin-triggered IP release includes `released_for` specifying (currently raw) target user id — this is a temporary compromise; will migrate to hashed token.

## Redaction Strategy & Roadmap
Current redaction is coarse (constant `[redacted]`). Upcoming refinements:
1. (DONE – memory + stdout + sqlite) Deterministic hashing via HMAC-SHA256 (first 144 bits base64url) of actor & object when store constructed with `WithHashing(secret)`, `WithSqliteHashing(secret)` or using `NewStdoutAuditorWithHashing(secret)`. Default remains full redaction.
2. Configurable retention (memory cap + ring buffer / eviction policy).
3. Streaming exporter: channel fan-out to external sinks (stdout, OpenTelemetry, file, webhook).
4. Tamper-evidence: append hash chain: `event_hash = H(prev_hash || canonical_json)`. Persist head hash for integrity audits.
5. Backpressure controls & async buffering (bounded queue + worker) for production throughput.

## Integrity & Ordering
- In-memory slice preserves insertion order; no reordering.
- Concurrency test ensures safety under parallel writes.
- Future persistence should use monotonic sequence IDs + WAL.

## Performance Considerations
- Allocations: Single mutex lock/unlock per event append; adequate for current throughput. Target < 5µs per append locally.
- Planned optimization: lock-free ring buffer or sharded mutex if event volume grows.

## Testing
- Concurrency test validates exact event count under 64 * 50 parallel writes.
- Handler & service tests assert event presence for critical flows.

## Future API Extensions
| Feature                               | Rationale                                       |
| ------------------------------------- | ----------------------------------------------- |
| Query API (time range, action filter) | Operational troubleshooting & compliance export |
| Export format (NDJSON streaming)      | Integrate with SIEM / log pipeline              |
| Action constants package              | Prevent typos & enable IDE completion           |
| Multi-sink fan-out with retries       | Reliability under sink outage                   |
| Privacy budget / classification tags  | Data governance & selective retention           |

## Security Considerations
- No PII at rest: ensure future details map excludes user-supplied raw strings unless classified safe.
- Potential side-channel: IP values may reveal network structure; acceptable for operational logs but consider masking last octet for user allocations if policy requires.
- Rotation of hashing keys must include forward-security (previous events not rehashable to raw IDs).

## Open Questions
- Do we require per-tenant logical segregation or encryption-at-rest for audit logs? (Not yet; backlog item for multi-tenant maturity.)
- Should audit failures fail the parent operation? (Current: best-effort fire-and-forget. Future: configurable critical actions.)
- Introduce sampling for high-volume benign events? (Optional toggle.)

## Immediate Next Steps
1. Structured exporter interface + noop + stdout + channel fan-out.
2. Hash chain (tamper evident) layering atop new SQLite `audit_events` table (use monotonic seq as chain order anchor).
3. Retention / eviction policy (bounded ring buffer for in-memory mode) + optional SQLite pruning (time/row cap).
4. HMAC key rotation mechanism (dual-key window + rehash forward).
5. Metrics instrumentation (event/sec, failures, queue latency when async buffering added).

## Persistence (SQLite Prototype)

Implemented `SqliteAuditor` (`internal/audit/sqlite.go`) using pure Go driver `modernc.org/sqlite` (CGO-free). Schema:

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
- Fire-and-forget insert inside `Event` (errors ignored intentionally; future: metric + backpressure channel).
- Optional hashing (`WithSqliteHashing(secret)`) mirrors in-memory/stdout truncation (Base64 URL, first 18 bytes of HMAC digest (144 bits)).
- `ListRecent` + `Count` helper methods exist solely for tests / diagnostics; no public HTTP exposure yet.
- Monotonic `seq` + high-resolution `ts` prepare for hash chain (`chain_hash = H(prev_chain_hash || canonical_event)` future work).
- Portability: no system SQLite dependency required.

Planned Enhancements:
- Config toggles for WAL + performance pragmas (journal_mode=WAL, synchronous=NORMAL, busy_timeout).
- Streaming iterator / pagination API for operational export.
- Integrity chain table referencing `seq` with periodic snapshot anchors for audit verification.
- Optional durability escalation: fail-fast on insert error for security-critical actions (configurable policy).

---
Revision: v1 (initial) – Date: 2025-09-29
