---
mode: agent
---
**Görev:** /v1/networks için create + list’i uygula.
**Yanıt formatın:** PLAN → PATCHES → TESTS → DOCS → COMMIT
**Kabul kriterleri:**
1) Model/Schema: Network{id,name,visibility,join_policy,cidr,created_by,created_at,updated_at,soft_deleted_at,moderation_redacted}
2) POST /v1/networks: Idempotency-Key zorunlu; CIDR valid + overlap=409 ERR_CIDR_OVERLAP; audit=NETWORK_CREATE
3) GET /v1/networks: visibility=public|mine|all, limit+cursor; stabil sıralama; audit=NETWORK_LIST
4) RBAC: create=authenticated; all=admin
5) Error: tek tip {code,message,details}; idempotency body mismatch=409 ERR_IDEMPOTENCY_CONFLICT
6) TESTS: unit (CIDR/overlap, idempotency, RBAC), integration (duplicate create, list filter/paging)
7) DOCS: server/openapi/openapi.yaml ve docs/TECH_SPEC.md güncelle
8) CI yeşil kalmalı
9) Commit: feat(server): add /v1/networks create+list with IPAM overlap, idempotency and audit
