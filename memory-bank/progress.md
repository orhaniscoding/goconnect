# progress

- What works: SQLite persistence for user/tenant/network/membership/join_request/device/peer/chat/network_chat/ip_rule/invite_token/tenant_member/tenant_invite/tenant_announcement/tenant_chat; IPAM on SQLite; migrations_sqlite aligned with Postgres; config YAML load/save with env override; setup mode `/setup` GET status + POST persist + `/setup/status` + `/setup/validate`; steps metadata + `restart_required` in POST response; `/setup` and `/setup/status` now report progress (completed/next steps, config presence/validity); `restart:true` triggers graceful shutdown; `cd server && go test ./...` green.
- What's left: Finish setup wizard UX (frontend multi-step + restart wiring), add encryption/keyring (SQLCipher/field-level + go-keyring) and key management, embed web UI, optional SQLite idempotency repo.
- Known issues: None noted beyond pending features.
