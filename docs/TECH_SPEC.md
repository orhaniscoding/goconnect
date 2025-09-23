# GoConnect — Teknik Proje Dökümanı (FINAL • Ultra-Detay • Branding + OS Uyumlu) 

**Ürün adı:** **GoConnect**
**Yapımcı/Marka:** **orhaniscoding** (© Orhan Tüzer) — Lisans: **MIT**
**Binaries (tüm OS’lerde sabit isimler):**

* **Host/Server:** `goconnect-server` (Windows’ta `goconnect-server.exe`)
* **Client-Daemon:** `goconnect-daemon` (Windows’ta `goconnect-daemon.exe`)

**`--version` çıktısı (server & daemon):**

```
goconnect-<name> vX.Y.Z (commit <hash>, build <YYYY-MM-DD>) built by orhaniscoding
```

## 0) Ürün Amacı & Kapsam (Hamachi/ZeroTier benzeri; web-first)

GoConnect, internet üzerindeki cihazları **aynı sanal ağa** dahil eder.

* **Host (Server)**: Ağların, üyeliklerin, yapılandırmaların, sohbetlerin ve tüm audit/veri durumunun **tek kaynağı**.
* **Client-Daemon**: Cihazda çalışan hafif ajan; **WireGuard** tünelini uygular, **Localhost Bridge** ile Web UI’ye köprü olur.
* **Unified Web UI (Next.js)**: Hem son kullanıcı hem admin işlemleri tek web arayüzünde. TR/EN i18n + A11Y.

> **GENEL KURAL:** “Durum (state) host’ta.” Client-daemon sadece “uygulayıcı” (ephemeral). Mümkün olan her işlem **host** üzerinde kayıt altına alınır.

---

## 1) Alan Adları, CORS & CSP

* **Web UI:** `https://app.goconnect.example`
* **API/WS:** `https://api.goconnect.example`  (WS: `wss://api.goconnect.example/v1/ws`)
* **Localhost Bridge:** `http://127.0.0.1:<random_port>`

**CORS (sıkı):** Sadece `https://app.goconnect.example` (dev’de `http://localhost:*` izinli).
**CSP (özet):**

```
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline';
img-src 'self' data:;
connect-src 'self' https://api.goconnect.example wss://api.goconnect.example http://127.0.0.1:*;
frame-ancestors 'none';
base-uri 'none';
form-action 'self';
```

---

## 2) Teknoloji & Temel İlkeler

* **Go 1.22+**, **Node 20+**, **Postgres 15+**, **Redis 7+**, **Next.js 14+**
* Kimlikler: **ULID**; Zaman: **UTC**; Günlük: **JSON (zap)**
* İzlenebilirlik: **OpenTelemetry** trace + **Prometheus** metrikleri
* Şifreleme: **Argon2**; Tokenlar: **JWT** + **JWKS rotation (kid)**
* Tüm **mutation** REST isteklerinde **Idempotency-Key** (24h TTL; gövde değişirse **409**)

**Standart Hata Şeması (REST):**

```json
{ "code": "ERR_SNAKE_CASE", "message": "Human readable", "details": { }, "retry_after": 0 }
```

---

## 3) Repo Yerleşimi (Kanonik)

```
README.md
LICENSE (MIT)
copilot-instructions.md
go.work
goreleaser.yml
package.json (dev tooling: commitlint/husky)
.editorconfig  .gitignore  .golangci.yml  .commitlintrc.cjs  .vscode/settings.json
.github/workflows/  (ci.yml, release-please.yml, auto-readme.yml)
release-please-config.json
.release-please-manifest.json
scripts/gen_readme.mjs

docs/
  TECH_SPEC.md  (bu dosya)
  SUPER_PROMPT.md
  LOCAL_BRIDGE.md
  THREAT_MODEL.md  RUNBOOKS.md  CONFIG_FLAGS.md  SECURITY.md
  SSO_2FA.md  I18N_A11Y.md  API_EXAMPLES.http  WS_MESSAGES.md  SEQ_DIAGRAMS.md
  README.tpl.md

server/
  go.mod
  docker-compose.yaml (dev: Postgres+Redis)
  Makefile
  cmd/server/main.go              (--version ldflags uygular)
  openapi/openapi.yaml            (kaynak doğruluk)
  internal/...                    (domain, repo, handlers, services, ipam,...)
  migrations/...                  (golang-migrate)

client-daemon/
  go.mod
  Makefile
  cmd/daemon/main.go              (--version ldflags uygular)
  service/linux/goconnect-daemon.service
  service/windows/install.ps1
  service/macos/com.goconnect.daemon.plist

web-ui/
  package.json  next.config.js  .env.example
  public/locales/tr/common.json
  public/locales/en/common.json
  src/lib/{api.ts, bridge.ts}
  src/components/Footer.tsx
  src/app/(public)/login/page.tsx
  src/app/(protected)/dashboard/page.tsx
```

---

## 4) Mimari: Bileşenler & Sorumluluklar

### 4.1 Host (Server, Go)

* **REST `/v1`** ve **WS `/v1/ws`**
* **Ağ** oluşturma/güncelleme/silme, **üyelik** yönetimi (join/approve/deny/kick/ban)
* **WireGuard** profil üretimi (IPAM, AllowedIPs, MTU, DNS, Keepalive=25)
* **Chat** (host chat + ağ chat): edit, soft-delete, hard-delete, **moderation\_redacted** (kismi sansür), sürüm geçmişi
* **RBAC:** tenant → (owner/admin/moderator/member)
* **Denied/Quota/Rate-limit:** login, join, chat, invite vb.
* **Audit (immutable) + PII redaction**
* **Event Outbox** (WS/Jobs) + Redis Pub/Sub fan-out
* **Multi-tenant:** tüm veri `tenant_id` ile scope’lu

### 4.2 Client-Daemon (Go)

* **WireGuard** tünel uygular:

  * **Windows:** WireGuardNT (Wintun), servis + firewall kural
  * **macOS/Linux:** `wg-quick` ile; gerekli **CAP\_NET\_ADMIN** / LaunchDaemon plist
* **Localhost Bridge** sunar: `http://127.0.0.1:<random_port>`

  * `GET /status`, `POST /wg/apply`, `POST /wg/down`, `GET /peers`
* **Host heartbeat** + auto-reconnect
* **Zero local state** (kalıcı durum host’ta)

### 4.3 Unified Web UI (Next.js)

* **User**: Login, dashboard, host chati, public/private ağ listesi & katılım, ağ chat
* **Admin**: Onay kuyruğu, ağ yönetimi, moderasyon, ban/kick, audit arama
* **i18n TR/EN**; **A11Y**; **Footer**: “© 20XX GoConnect — Built by orhaniscoding”

---

## 5) RBAC ve Yetkiler

* **owner**: tenant yönetimi, tüm ağlar üzerinde tam yetki
* **admin**: tenant içindeki tüm ağları yönetir
* **moderator**: chat ve üyelik moderasyonu (kick/ban, redact/soft/hard delete)
* **member**: ağlara katılır, chat kullanır
* **Kontrat:** Erişim kontrolleri hem REST hem WS tarafında uygulanır; audit’e işlenir.

---

## 6) Veri Modeli (Özet Şema)

**Tablolar (kritik alanlar):**

* **tenants**(id, name, ...), **users**(id, email, pass\_hash, locale, 2fa, ...), **devices**(id, user\_id, platform, pubkey, …)
* **networks**(id, tenant\_id, name, visibility\[public|private], join\_policy\[open|invite|approval], cidr, dns, mtu, split\_tunnel, created\_by, …)
* **memberships**(id, network\_id, user\_id, device\_id?, role, status\[pending|approved|banned], device\_pubkey, joined\_at, …)
* **wg\_peers**(id, network\_id, device\_id, assigned\_ip, allowed\_ips\[], endpoint\_hint, rotation\_at, …)
* **chat\_messages**(id, scope\[host|network:<id>], user\_id, body, attachments?, **deleted\_at?**, **redacted\:boolean**, **redaction\_mask?**, created\_at)
* **chat\_message\_edits**(id, message\_id, prev\_body, new\_body, editor\_id, edited\_at)
* **audit\_logs**(id, tenant\_id, actor\_id, action, object\_type, object\_id, before?, after?, **immutable** bool, ts)
* **bans**(id, tenant\_id, user\_id?, device\_id?, reason, expires\_at?)
* **invite\_tokens**(id, network\_id, token, expires\_at, uses\_max, uses\_left)
* **idem\_keys**(key, request\_hash, response\_hash, expire\_at)

**Kısıt & İndeks:**

* Unique: `(tenant_id, network_name)`, `(network_id, device_id)`, `(network_id, assigned_ip)`
* Index: chat `scope, created_at DESC`; audit `tenant_id, created_at DESC`
* **FK’ler ON DELETE RESTRICT**, audit için “before/after snapshot” JSONB

---

## 7) IPAM (Ağ Adresleme) & WireGuard Profil

### Networks API Implementation Flow

**Implemented in v0.1.0:** Basic CRUD operations with IPAM and idempotency

#### POST /v1/networks Implementation:
1. **Idempotency Check**: Validate key+body hash, return cached response if exists
2. **Input Validation**: CIDR format validation using Go's `net.ParseCIDR`
3. **CIDR Overlap Detection**: Check against existing networks using IP containment
4. **Business Logic**: Name uniqueness per tenant, audit logging
5. **Repository**: In-memory store with thread-safe operations
6. **Response**: Standard format with proper HTTP status codes

#### GET /v1/networks Implementation:
1. **RBAC Filtering**: visibility=public|mine|all with admin-only enforcement
2. **Cursor Pagination**: Stable ordering with configurable limits
3. **Search**: Case-insensitive name matching
4. **Repository**: Efficient in-memory filtering and sorting

#### Architecture Pattern:
```
Handler → Service → Repository
   ↓         ↓         ↓
Input    Business   Data
Valid    Logic      Access
```

#### Error Handling:
- **ERR_CIDR_OVERLAP**: Network CIDR conflicts with existing range
- **ERR_IDEMPOTENCY_CONFLICT**: Same key, different request body
- **ERR_FORBIDDEN**: Insufficient privileges (admin-only operations)
- **ERR_CIDR_INVALID**: Malformed CIDR or host address instead of network

#### Testing Coverage:
- **Unit Tests**: CIDR validation, overlap detection, idempotency logic
- **Integration Tests**: HTTP handlers, RBAC enforcement, error scenarios
- **Repository Tests**: Data consistency, concurrent access patterns

* **CIDR**: Ağ başına benzersiz; çakışma → **409 `ERR_CIDR_OVERLAP`**
* **IP atama**: Havuzdan, gateway ve broadcast hariç; `assigned_ip` unique
* **Profil üretimi (wg-quick)**:

  * `[Interface]` → `Address=<assigned_ip>`, `DNS=<dns>`, `MTU=<mtu>` (opsiyon), `PrivateKey=<client>`
  * `[Peer]` → `PublicKey=<server_pubkey>`, `Endpoint=<host_ip:port>`, `AllowedIPs=<network_cidr veya split list>`, `PersistentKeepalive=25`

**Döndürülen profil** host’ta üretilir ve **audit** log’a “profile\_rendered” olarak işlenir.

---

## 8) REST API (Sözleşme)

**Temel Başlıklar:**

* `Authorization: Bearer <JWT>`
* `Idempotency-Key: <uuid>`  (mutasyonlarda **zorunlu**)
* `X-Request-Id: <ulid>` (server da üretip döner)

**Çekirdek Uçlar (özet):**

* **Auth**

  * `POST /v1/auth/register` → {email, password}
  * `POST /v1/auth/login` → {email, password}
  * `POST /v1/auth/refresh`
  * `POST /v1/auth/logout`
* **Networks**

  * `POST /v1/networks` → {name, visibility, join\_policy, cidr, dns?, mtu?, split\_tunnel?}
    - **Idempotency-Key**: Required header for mutation safety (24h TTL)
    - **CIDR Validation**: Network format validated, overlap detection enforced
    - **RBAC**: Authenticated users can create networks
    - **Business Rules**: Name unique per tenant, CIDR non-overlapping
    - **Errors**: 409 ERR_CIDR_OVERLAP, 409 ERR_IDEMPOTENCY_CONFLICT
  * `GET /v1/networks` (query: visibility=public|mine|all, search, paging)
    - **Visibility Filters**: `public` (default), `mine` (user's networks), `all` (admin-only)
    - **Pagination**: Cursor-based with configurable limit (1-100, default 20)
    - **Search**: Optional name-based filtering
    - **RBAC**: Authenticated users see public+mine, admins see all
  * `GET /v1/networks/:id`
  * `PATCH /v1/networks/:id` (rename, visibility, join\_policy, …)
  * `DELETE /v1/networks/:id`
* **Memberships**

  * `POST /v1/networks/:id/join` (open → auto approve; invite/approval → pending)
  * `POST /v1/networks/:id/approve` (admin/moderator)
  * `POST /v1/networks/:id/deny`
  * `POST /v1/networks/:id/kick`
  * `POST /v1/networks/:id/ban`
  * `GET /v1/networks/:id/members`
* **WireGuard**

  * `POST /v1/networks/:id/wg/keypair`
  * `GET /v1/networks/:id/wg/profile` (client CFG; **audit** kaydı)
  * `POST /v1/networks/:id/wg/rotate` (server/peers rotasyonu)
* **Chat**

  * `GET /v1/chat?scope=host|network:<id>` (paging)
  * `POST /v1/chat` (send), `PATCH /v1/chat/:id` (edit),
  * `DELETE /v1/chat/:id?mode=soft|hard`
  * `POST /v1/chat/:id/redact` (partial; `redaction_mask` döner)
* **Audit**

  * `GET /v1/audit?actor=&action=&object_type=&from=&to=`

**Örnek Hata Yanıtı:**

```json
{ "code": "ERR_NOT_AUTHORIZED", "message": "You are not allowed to approve members.", "details": { "required_role": "moderator" } }
```

---

## 9) WebSocket (Sözleşme)

* **Endpoint:** `wss://api.goconnect.example/v1/ws` (JWT zorunlu)
* **Inbound** (hepsi `op_id` içerir):

  * `auth.refresh`
  * `chat.send` `{scope, body, attachments?}`
  * `chat.edit` `{message_id, new_body}`
  * `chat.delete` `{message_id, mode: "soft"|"hard"}`
  * `chat.redact` `{message_id, mask}`
* **Outbound:**

  * `chat.message` `{id, scope, user_id, body, redacted, ...}`
  * `chat.edited`, `chat.deleted`, `chat.redacted`
  * `member.joined`, `member.left`
  * `request.join.pending`
  * `admin.kick`, `admin.ban`
  * `net.updated`

**Ölçekleme:** Redis Pub/Sub fan-out; sticky gerekmez. **Backpressure:** oda başına kuyruk sınırı; sınırı aşınca oldest-drop (audit’e yazılır).

---

## 10) Localhost Bridge (Web ↔ Daemon)

* **Origin:** `http://127.0.0.1:<random_port>`
* **Kimlik:** OAuth2 + **PKCE** + custom URL scheme `goconnect://pair?...`
* **Header:** `X-Loopback-Token` (**10 dk TTL**, sliding). Yeniden kullanım → **401 `ERR_LOOPBACK_REUSE`**
* **Uçlar:**

  * `GET /status` → {running, wg: {active, iface, ip, peers}}
  * `POST /wg/apply` → WireGuard profilini uygula
  * `POST /wg/down` → Tüneli kapat
  * `GET /peers` → Aktif WG peers listesi
* **CORS:** Sadece `https://app.goconnect.example`

---

## 11) Güvenlik (Detay)

* **JWT** kısa ömürlü; **JWKS rotation (kid)** + **7 gün grace**
* **Argon2** (parola hash), 2FA (TOTP/WebAuthn) opsiyonel
* **RBAC** kontrolleri **handler** katında (server) + **policy** testleri
* **Rate-limit**: login (IP ve hesap bazlı), join, chat, invite
* **IP allow/deny**: admin panelden düzenlenebilir (tenant scope)
* **GDPR/DSR**: `/me/export`, `/me/delete` (queue + asenkron iş; audit kaydı)
* **Log redaction**: e-posta, token, IP gibi PII alanlar maske
* **Hata Kodları (kısa katalog):**

  * `ERR_INVALID_CIDR`, `ERR_CIDR_OVERLAP`
  * `ERR_NOT_AUTHORIZED`, `ERR_QUOTA_EXCEEDED`
  * `ERR_LOOPBACK_REUSE`, `ERR_RATE_LIMITED`

---

## 12) Observability (Metrik & İzleme)

* **Prometheus**

  * `http_requests_total{path,method,status}`
  * `http_request_duration_seconds_bucket{path,method}`
  * `ws_events_total{room,event}`
  * `errors_total{code}`
  * `outbox_pending`, `dlq_size`
* **OTEL Traces**

  * `ipam.allocate`, `wg.profile.render`, `ws.broadcast`, `auth.login`
* **Logs** (zap JSON)

  * `correlation_id` / `X-Request-Id`, düzey: info→error; hassas alanlar maskelenir.

---

## 13) Test Stratejisi (Tam Kapsam)

* **Unit (Go)**: RBAC policy, IPAM (CIDR overlap), error katalogu, rate-limit, outbox, daemon bridge adapter
* **Integration**: `docker-compose` (PG+Redis+Server) — **senaryoların uçtan uca**:

  1. register→login→create\_network→join\_request→approve→wg\_profile
  2. chat send→edit→soft\_delete→redact→hard\_delete (audit doğrulaması)
* **E2E (Playwright)**: Web UI login, network list/join, admin approvals, chat mod (TR/EN)
* **WS Harness**: inbound/outbound, `op_id` idempotency ve outbox fan-out
* **Contract (Schemathesis)**: `server/openapi/openapi.yaml` üzerinden property-based
* **Fuzz**: JSON decoder/validator, WG profile renderer
* **Load (k6)**: 1K WS client ile chat fan-out kısa koşu
* **Chaos (Toxiproxy)**: PG/Redis latency/loss/reset; **DLQ replay** senaryosu
* **Coverage**: **≥ %70** (server + daemon, satır + branch)

**Test Matris:**

* OS: Win 10/11, macOS 13+, Ubuntu 22.04+
* Arch: amd64, arm64
* Go: son minör + n-1; Node: LTS

---

## 14) CI/CD, Sürümleme & Paketleme (Branding’li)

* **CI** (`.github/workflows/ci.yml`):

  * Server/Daemon: `go test ./... -race -cover`
  * Web UI: `npm i` + `next build`
  * Concurrency: **cancel-in-progress**
* **README otomasyonu**: `scripts/gen_readme.mjs` + `auto-readme.yml`
* **Sürümleme**: **Release Please (monorepo)**

  * `release-please-config.json`, `.release-please-manifest.json`, `release-please.yml`
  * Otomatik **Release PR** + **Changelog**
* **Dağıtımlar**: **GoReleaser** (`goreleaser.yml`)

  * `ldflags` ile `--version` bilgilerinin gömülmesi (**built by orhaniscoding**)
  * **Linux**: `nfpm` → `.deb/.rpm`, `systemd` unit dosyaları; Maintainer: orhaniscoding
  * **Windows**: **Scoop** manifest + **Winget** şablonu; Service installer `install.ps1`
  * **macOS**: `.pkg` (notarization pipeline — Apple Developer hesap/secret gerektirir), LaunchDaemon plist
  * **SBOM**/**cosign** imzalama (ileriki aşama pipeline adımı)

**GoReleaser — ldflags örneği:**

```yaml
ldflags:
  - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=orhaniscoding
```

---

## 15) i18n & A11Y & UI Branding

* **i18n**: `public/locales/{tr,en}/common.json` — **UI stringleri hard-code etmeyin.**
* **A11Y**: Kontrast, odak halkası, ARIA, klavye erişimi; komponent PR’larında lint
* **Footer**:

  * TR: `© 20XX GoConnect — orhaniscoding tarafından geliştirildi`
  * EN: `© 20XX GoConnect — Built by orhaniscoding`
* **About sayfası**: `--version` çıktısını (server/daemon) gösterir.

---

## 16) Hata Kodları (Katalog / Örnekler)

* **Ağ & IPAM**: `ERR_INVALID_CIDR`, `ERR_CIDR_OVERLAP`
* **Kimlik/RBAC**: `ERR_NOT_AUTHORIZED`
* **Kota/Oran**: `ERR_QUOTA_EXCEEDED`, `ERR_RATE_LIMITED`
* **Bridge**: `ERR_LOOPBACK_REUSE`
* **Genel**: `ERR_CONFLICT`, `ERR_BAD_REQUEST`, `ERR_NOT_FOUND`, `ERR_INTERNAL`

---

## 17) Güvenlik İnce Noktaları

* Refresh token reuse → tüm oturumları revoke + audit
* Chat moderasyon: **soft-delete** (geri döndürülebilir UI işareti), **hard-delete** (kalıcı), **redaction** (kısmi sansür + `redaction_mask`), **edit** (sürüm geçmişi)
* Dosya/ek: AV taraması (opsiyonel), dosya türü/uzunluk sınırı
* Rate-limit’ler kaynak/role/pattern bazlı (çarpanlar api config’te)

---

## 18) Definition of Ready / Done

**Ready (bir iş başlamadan):**

* Kullanım hikâyesi, kabul kriterleri, RBAC etkisi, i18n/A11Y gereksinimi net; testler & metrikler tanımlı.

**Done:**

* Kod + test + dokümantasyon + metrikler **tam**; **coverage ≥ %70**;
* OpenAPI/WS dokümanı güncel;
* README otomasyonu ve Release Please yeşil;
* Binary isimleri ve **branding** korunmuş;
* PR tek mantık birimi, Conventional Commits ile.

---

## 19) Örnekler

**OpenAPI info (branding teması):**

```yaml
openapi: 3.0.3
info:
  title: GoConnect API
  version: 0.1.0
  contact:
    name: orhaniscoding
    url: https://github.com/orhaniscoding/goconnect
servers:
  - url: https://api.goconnect.example
```

**WS chat.send (inbound) örneği:**

```json
{
  "op_id": "01J..ULID",
  "op": "chat.send",
  "data": { "scope": "network:8b9..", "body": "Merhaba!" }
}
```

**Chat redaction (outbound) örneği:**

```json
{
  "ev": "chat.redacted",
  "data": { "message_id": "01J..", "redacted": true, "redaction_mask": "[***]" }
}
```

---

# EK: OS Servis/Paketleme Notları

### Linux (systemd)

* `/usr/local/bin/goconnect-daemon`
* Unit: `client-daemon/service/linux/goconnect-daemon.service`
* Paket: `nfpm` (`goreleaser.yml` içinde)

### Windows (Service + Scoop/Winget)

* Binary’ler: `goconnect-*.exe`
* Service: `client-daemon/service/windows/install.ps1` (`sc.exe create ...`)
* Dağıtım: `packaging/windows/scoop.json` + `packaging/winget/manifest.yaml`

### macOS (LaunchDaemon + Notarization)

* Plist: `client-daemon/service/macos/com.goconnect.daemon.plist`
* `.pkg` için `packaging/macos/pkgbuild.sh` (CI’da Apple Notarization gerekir)

---

# EK: Kod Kalitesi & Stil

* **Go**: `golangci-lint` (govet, staticcheck, errcheck, revive), `gofmt`
* **TS/React**: ESLint + TypeScript strict, `no-any`, `no-implicit-any`
* **Commit**: Conventional Commits (`feat:`, `fix:`, `chore:` …)
* **PR Şablonu**: kontrol listesi (RBAC, i18n, A11Y, test, docs, metrics)

---

> Bu doküman **Copilot/ChatGPT için kanonik kaynaktır**. Her türlü özellik/patch **bu şartnameye** göre uygulanmalı; farklılıklar açıkça gerekçelendirilmeli ve dokümana yansıtılmalıdır.

---
