# GoConnect • Copilot Süper Prompt (FINAL)

## 0) Genel Davranış

* **Bu proje adı:** `GoConnect`
* **Yapımcı:** `orhaniscoding`
* **TECH\_SPEC.md** → **Kanonik Şartname**. Onu oku, sapma.
* **Tüm çıktıların** TECH\_SPEC.md’ye uygun, eksiksiz, testli, belgeli, güvenli, brand’li olması **zorunlu**.
* **Asla tahmin etme**. Emin olmadığında “TECH\_SPEC.md’ye göre boş bırakıldı, açıklama istiyorum” de.
* **Her yanıtın formatı:**

  1. **PLAN** — yapılacak işin açıklaması
  2. **PATCHES** — dosya bazlı diff veya yeni dosya içerikleri
  3. **TESTS** — unit/integration/e2e/WS testleri
  4. **DOCS** — docs/ klasöründe güncellemeler (OpenAPI, WS\_MESSAGES, README otomasyonu vb.)
  5. **COMMIT** — Conventional Commit mesajı

---

## 1) Repo Yapısı

* Klasörler: `server/`, `client-daemon/`, `web-ui/`, `docs/`, `scripts/`
* Dosyalar: `.editorconfig`, `.gitignore`, `.golangci.yml`, `.vscode/settings.json`, `copilot-instructions.md`
* CI/CD: `.github/workflows/{ci.yml,release-please.yml,auto-readme.yml}`, `release-please-config.json`, `.release-please-manifest.json`
* Paketleme: `goreleaser.yml`, `client-daemon/service/{linux,windows,macos}/` servis dosyaları, `packaging/{windows,macos}` şablonları

---

## 2) Binaries & Branding

* Binary adları **sabit**:

  * `goconnect-server` (`.exe` Win’de)
  * `goconnect-daemon` (`.exe` Win’de)

* **--version** çıktısı:

  ```
  goconnect-<name> vX.Y.Z (commit <hash>, build <YYYY-MM-DD>) built by orhaniscoding
  ```

* **GoReleaser ldflags** ile gömülecek (`goreleaser.yml`).

* **Web UI Footer:**

  * EN: `© 20XX GoConnect — Built by orhaniscoding`
  * TR: `© 20XX GoConnect — orhaniscoding tarafından geliştirildi`

* **OpenAPI info.contact:**

  ```yaml
  contact:
    name: orhaniscoding
    url: https://github.com/orhaniscoding/goconnect
  ```

---

## 3) Çekirdek İlkeler

* **Her patch**: Kod + Test + Docs **birlikte** gelir.
* **REST ve WS kontratları** OpenAPI (`server/openapi/openapi.yaml`) ve `docs/WS_MESSAGES.md` dosyalarına güncellenir.
* **Idempotency-Key** → tüm mutasyon REST çağrılarında zorunlu, test ile kanıtlanmalı.
* **RBAC enforcement** → handler + test.
* **Audit logging** → tüm kritik işlemler immutable log’da.
* **i18n** → Web UI stringleri locale JSON’dan; asla hard-coded string ekleme.
* **A11Y** → her yeni komponentte aria-label ve klavye erişim testi zorunlu.
* **Coverage**: Go kodları için ≥%70.

---

## 4) CI/CD & Release

* **CI**:

  * Go: `go test ./... -race -cover`
  * Web UI: `npm i && npm run lint && npm run build`
* **README**: `scripts/gen_readme.mjs` → `auto-readme.yml`
* **Release**: Release Please (monorepo) → sürüm PR + changelog
* **GoReleaser**: Linux `.deb/.rpm` (nfpm), Windows (Scoop+Winget), macOS `.pkg` (notarization), ldflags ile branding.
* **SBOM & cosign**: release pipeline’da (gelecek aşama).

---

## 5) Test Sistemi

* **Unit**: RBAC, IPAM, error katalogu, rate-limit, outbox
* **Integration**: PG+Redis+Server senaryoları (register→join→approve→wg\_profile)
* **E2E**: Playwright (login, network join, admin approval, chat flows TR/EN)
* **WS Harness**: inbound/outbound op\_id testleri
* **Contract**: Schemathesis → OpenAPI
* **Fuzz**: JSON decoder, WG profile renderer
* **Load**: k6 → 1K WS clients
* **Chaos**: Toxiproxy (PG/Redis latency/loss/reset)

---

## 6) Yapılmaması Gerekenler

* Binary adını **asla** değiştirme.
* `--version` çıktısında “orhaniscoding” yazmıyorsa **hata**.
* Web UI stringleri **hard-coded** olamaz.
* Test/dokümansız PR olmaz.
* Branding/metadata her zaman korunur.

---

## 7) Örnek İlk Görev (Copilot’a Verilecek)

> **“`/v1/networks create` endpointini TECH\_SPEC.md’deki kontrata göre ekle. PLAN → PATCHES → TESTS → DOCS → COMMIT formatında yanıt ver.”**

Copilot’un yapması gereken:

* Server’da handler + service + repo ekle
* Integration test (network create → duplicate name → ERR\_CONFLICT)
* OpenAPI güncellemesi
* Audit log entry
* Conventional commit: `feat(server): implement /v1/networks create endpoint`
