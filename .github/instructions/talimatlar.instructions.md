---
applyTo: '**'
---
Provide project context and coding guidelines that AI should follow when generating code, answering questions, or reviewing changes.

GoConnect – Çalışma Protokolü (kanonik kaynak = docs/TECH_SPEC.md)

Her görevde çıktı formatın PLAN → PATCHES → TESTS → DOCS → COMMIT.

docs/TECH_SPEC.md’deki sözleşmeler bağlayıcıdır (RBAC, idempotency, hata kodları, chat moderasyon/soft-delete/redact, IPAM).

Server (Go): handler → service → repo katmanı; tek tip hata modeli {code,message,details}; mutasyonlarda Idempotency-Key zorunlu.

Web (Next.js App Router): her segmentte layout.tsx; build-time fetch yok; client-side fetch (bridge/api); TR/EN i18n; A11Y.

CI’ı kırma: go test ./... -race -cover ve npm i && npm run typecheck && npm run build yeşil kalmalı.

Dokümantasyon: OpenAPI + docs/TECH_SPEC.md senkron; README otomasyonu bozulmasın.

Güvenlik: RBAC, rate limit, audit; log’larda PII yok; sürüm bilgisi --version ldflags ile.

Commitler: Conventional Commits, signed.

(İpucu) “Project index / repo indexing” varsa docs/, server/, web-ui/’yi dahil et.
