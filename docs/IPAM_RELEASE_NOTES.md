# IPAM Release Notes

## Scope
Initial in-memory IP Address Management (IPAM) implementation providing:
- Deterministic sequential allocation of IPv4 addresses within a network CIDR
- Stable per-user allocation (idempotent: repeated allocation returns same IP)
- Release & reuse semantics (LIFO free list of offsets)
- Admin/Owner capability to release another member's allocation
- Membership enforcement (only approved members may allocate/list/release)
- Concurrency-safety (mutex protected repository; race tests pass)
- Audit events: IP_ALLOCATED, IP_RELEASED (with `released_for` detail for admin releases)

## Design Overview
Each `IPAllocation` stores:
- NetworkID
- UserID
- IP (string, computed from network CIDR + offset)
- Offset (int, not exposed via API)

Repository tracks per-network:
- `nextOffset` (monotonic host offset to assign when free list empty)
- `freeOffsets` stack (slice functioning as LIFO for reuse)
- Allocations keyed by user

Allocation steps:
1. If user already has allocation, return it (stable).
2. Else if `freeOffsets` not empty: pop last offset and compute IP.
3. Else assign `nextOffset`, increment, compute IP; guard against exhaustion by verifying host count.

Release steps:
1. If user has allocation: remove entry and push its offset onto `freeOffsets` (LIFO reuse).
2. If user absent: noop (idempotent success).

### Offset to IP Computation
`NextIP(baseCIDR, offset)` implemented in domain layer; IP math uses standard library + big.Int increment logic. Host ordering matches sequential integer offsets excluding network & broadcast addresses automatically by offset spacing.

### Concurrency Considerations
- Repository uses a mutex; operations are O(1) for allocate/release (amortized) except IP parsing for new networks.
- Stress test performs parallel allocate/release cycles ensuring no duplicate IPs and race-detector clean.

### Idempotency
Service layer enforces membership; allocation is naturally idempotent for an existing user. Separate global mutation idempotency (via Idempotency-Key header) applies to other resources but is not required for IP allocation because allocate is effectively a upsert per (network,user).

## Error Model
- Non-members / non-approved members receive `ERR_NOT_AUTHORIZED`.
- Exhaustion returns `ERR_IP_EXHAUSTED` (mapped to HTTP 409 Conflict via domain mapping).
- Network not found surfaces `ERR_NOT_FOUND`.

## Audit Events
| Action       | Details                             | Emitted When                                                                                       |
| ------------ | ----------------------------------- | -------------------------------------------------------------------------------------------------- |
| IP_ALLOCATED | `ip`                                | User receives allocation (new or first call)                                                       |
| IP_RELEASED  | optional `released_for` target user | User releases own allocation (no details) OR admin/owner releases another member (includes target) |

Actor/Object identifiers are redacted (`[redacted]`) at storage layer to prevent PII leakage; correlation relies on request-scoped metadata (future extension: hashed identifiers with salt).

## Future Enhancements
1. IPv6 support (dual stack) and CIDR partitioning.
2. Pool fragmentation handling and defragmentation metrics.
3. Reserved address ranges per network (e.g., skip .1 for gateway usage).
4. Bulk allocation for gateways / service accounts.
5. Persistence layer (SQLite / Badger / Postgres) replacing in-memory maps.
6. Metrics: allocation latency, utilization %, churn rate.
7. Soft quota per user or role-based allocation limits.
8. Structured audit correlation IDs (span / trace integration).
9. Garbage collection for stale allocations (lease model with renewal heartbeat) optional.

## Testing Summary
- Service tests cover sequential allocation, stability, invalid network, membership denial, release reuse, idempotent release, admin release scenarios, and concurrency stress.
- Handler tests validate HTTP flow, role gating, admin release endpoint, and audit emission.
- Race detector: all tests pass with `-race`.

## Migration Notes
Current state is in-memory only; any restart loses allocations. Introducing persistence will require a data model:
```
table ip_allocations (
  network_id text,
  user_id text,
  ip text,
  offset integer,
  allocated_at timestamp,
  primary key(network_id, user_id)
)
```
And an offsets tracking table or deriving nextOffset via MAX(offset) plus managing free list persistence.

## Open Questions
- Should we treat repeated allocation calls for a user as an audit event each time (currently only first matters)? (Decided: only first allocation for signal clarity.)
- Add explicit endpoint for admin listing all allocations? (Current list requires membership; admin same path.)
- Introduce lease expiry to prevent "zombie" allocations after member removal? (Planned with membership revoke hook.)

---
Revision: v1 (initial) – Date: 2025-09-29
