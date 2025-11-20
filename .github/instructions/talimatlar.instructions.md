---
GoConnect – Çalışma Protokolü (kanonik kaynak = docs/TECH_SPEC.md)

⚠️ ÖNEMLİ: Geliştirme yapmadan önce Reports/ klasöründeki tüm raporları oku ve analiz et!
Reports/ klasörü projenin durumunu, eksik özellikleri, teknik borcu ve öncelikli görevleri içerir.
Her rapor tarih-saat damgalıdır ve proje kararlarında kritik referans kaynağıdır.

Her görevde çıktı formatın PLAN → PATCHES → TESTS → DOCS → COMMIT.

docs/TECH_SPEC.md'deki sözleşmeler bağlayıcıdır (RBAC, idempotency, hata kodları, chat moderasyon/soft-delete/redact, IPAM).To: '**'
---
Provide project context and coding guidelines that AI should follow when generating code, answering questions, or reviewing changes.

GoConnect – Çalışma Protokolü (kanonik kaynak = docs/TECH_SPEC.md)

Her görevde çıktı formatın PLAN → PATCHES → TESTS → DOCS → COMMIT.

docs/TECH_SPEC.md’deki sözleşmeler bağlayıcıdır (RBAC, idempotency, hata kodları, chat moderasyon/soft-delete/redact, IPAM).

Server (Go): handler → service → repo katmanı; tek tip hata modeli {code,message,details}; mutasyonlarda Idempotency-Key zorunlu.

Web (Next.js App Router): her segmentte layout.tsx; build-time fetch yok; client-side fetch (bridge/api); TR/EN i18n; A11Y.

CI Gating:
- Go: go test ./... -race -cover (≥ 60% target başlangıç; artırılabilir)
- Web: npm ci && npm run typecheck && npm run build
- Lint: golangci-lint clean, go vet clean, (ileride eslint) warnings yok
- Security: CodeQL / gosec kritik bulgu yok (aksi halde işaretle & tartış)

Dokümantasyon: OpenAPI + docs/TECH_SPEC.md senkron; README otomasyonu bozulmasın.

Güvenlik: RBAC, rate limit, audit; log’larda PII yok; sürüm bilgisi --version ldflags ile.

Commitler: Conventional Commits, signed.

PR Checklist (otomatik / manuel doğrula):
- [ ] Testler yeşil (race dahil)
- [ ] Linter temiz
- [ ] OpenAPI ve TECH_SPEC drift yok
- [ ] Audit / PII ihlali yok
- [ ] Idempotency-Key tüm mutasyonlarda korunuyor
- [ ] RBAC error outward code = ERR_NOT_AUTHORIZED

Release Akışı:
1. release-please → versiyon PR
2. Merge → tag → goreleaser workflow (ikili + paket)
3. README otomasyonu güncel tag / tarih basar

Güvenlik Notu:
- `validateToken` stub üretim için değildir; gerçek JWT/OIDC doğrulaması eklenmeden dağıtma.

(Gelecek) Observability Standardı:
- Prometheus metrics: http_requests_total, http_request_duration_seconds, audit_events_total
- Error log’larında PII yok, sadece hash / redacted alanlar.

(İpucu) “Project index / repo indexing” varsa docs/, server/, web-ui/’yi dahil et.
