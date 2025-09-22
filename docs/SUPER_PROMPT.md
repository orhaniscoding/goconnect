# GoConnect • Copilot Süper Prompt (FINAL)

- Proje adı: GoConnect · Yapımcı: orhaniscoding · Lisans: MIT
- Kanonik şartname: `docs/TECH_SPEC.md`. Ondan sapma.
- Binaries: `goconnect-server`, `goconnect-daemon` (Win: .exe). `--version` çıktısı brand içerir.
- UI footer TR/EN: “Built by orhaniscoding”. i18n JSON’dan, hard-coded yok.

YANIT FORMATIN:
1) PLAN – kapsam, dosyalar, riskler
2) PATCHES – dosya bazlı diff/icerik
3) TESTS – unit/integration/e2e/ws/contract/fuzz
4) DOCS – OpenAPI, WS, README.tpl, örnekler
5) COMMIT – Conventional subject

ZORUNLULAR:
- Idempotency-Key tüm mutasyonlarda (TTL24h, body mismatch→409)
- RBAC policy & tests
- Audit immutable kayıt
- OpenAPI + WS_MESSAGES güncel
- Coverage ≥70%
- CI yeşil; Release Please + GoReleaser uyumlu
- i18n & A11Y kuralları

YAPMA:
- Binary adı ve brand değiştirme
- Hard-coded UI string
- Test/dokümansız değişiklik

ÖRNEK İŞ:
“/v1/networks create” → handler+service+repo; integration test; OpenAPI güncelle; audit; commit: `feat(server): add /v1/networks create`
