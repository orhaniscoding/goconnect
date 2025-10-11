features.json: beta_webchat, relay_enabled...

Audit configuration (server via environment):

- AUDIT_SQLITE_DSN: SQLite DSN/path (enables persistent auditor when set). Example: file:/var/lib/goconnect/audit.db?cache=shared&_busy_timeout=5000
- AUDIT_HASH_SECRETS_B64: Comma-separated base64/base64url secrets for actor/object hashing (first is active, others for rotation). Example: c2VjcmV0MQ, bXlzZWNyZXQy
- AUDIT_MAX_ROWS: Max number of events to retain (FIFO prune). Example: 100000
- AUDIT_MAX_AGE_SECONDS: Max age in seconds for events (time-based prune). Example: 2592000
- AUDIT_ANCHOR_INTERVAL: Insert anchor rows every N events. Example: 1000
- AUDIT_SIGNING_KEY_ED25519_B64: Ed25519 private key (64 bytes) in base64/base64url to sign integrity exports. Do not commit. Example: <base64-encoded-private-key>
- AUDIT_SIGNING_KID: Optional signing key identifier included in export payload (kid). Example: ops@goconnect.example

Notes:
- Provide either MAX_ROWS or MAX_AGE or both; pruning is best-effort.
- If signing key is present, GET /v1/audit/integrity includes signature and optional kid.
