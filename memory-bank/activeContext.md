# activeContext

- Focus: Zero-config user experience (server-first).
- Completed: SQLite repos for core entities, migrations_sqlite aligned; config supports YAML load/save (`goconnect.yaml` default, `GOCONNECT_CONFIG_PATH` override) with env overrides; setup mode exposes `/setup` (GET status, POST persist to YAML) plus `/setup/status` & `/setup/validate`; steps metadata returned; `restart_required` flag in POST response; optional `restart:true` triggers graceful shutdown after response. Server tests pass via `cd server && go test ./...`.
- Completed (latest): Setup wizard endpoints now surface progress: `/setup` and `/setup/status` return step definitions, config presence/validity, completed/next step inference; `/setup/validate` echoes completed/next; `/setup` POST responds with progress metadata and restart_required, honoring restart flag to trigger graceful shutdown.
- Decisions: Idempotency repo stays in-memory for now; setup mode entered when config load fails; use env overrides only when non-empty.
- Next: Finish setup wizard UX (frontend flow/redirect + auto-restart wiring), encryption/keyring (SQLCipher or field-level + go-keyring) and config key handling; embed web UI into server; optional SQLite idempotency.
