# Detailed Tech Spec: User-Friendly "Zero-Config" Architecture

## 1. Executive Summary
This plan outlines the architectural transformation of the GoConnect Server and Client Daemon into a "Consumer Ready" product.
**Core Philosophy:**
*   **Zero-Touch Start:** Binaries must run immediately without crashing.
*   **Encrypted Persistence:** Local databases (SQLite) must be encrypted at rest.
*   **Service Continuity:** Applications run as OS services, surviving reboots.
*   **Cross-Platform Parity:** Windows, Linux, and macOS share the exact same setup flow and capabilities.

## 2. Server Architecture: "Personal Edition" (Embedded & Encrypted)

### 2.1. Data Persistence Layer (Encrypted SQLite)
*   **Current State:** PostgreSQL (Production) or In-Memory (Dev).
*   **New State:** Add `Repository` implementations using `sqlite3` with **SQLCipher** (or application-side AES-GCM encryption for fields).
    *   **Driver:** Use `github.com/mattn/go-sqlite3` with encryption extensions or `modernc.org/sqlite` if CGO-free is required (prefer standard with encryption).
    *   **Schema:** Use Go-migrate to manage SQLite migrations parallel to Postgres.
    *   **Encryption Key:**
        *   First run: Generate a cryptographically secure master key.
        *   Storage: Store this master key in the OS Keyring/Keychain (using `github.com/zalando/go-keyring`) where possible, or a protected file with strict permissions if headless.

### 2.2. Configuration Strategy
*   **Priority Loading:**
    1.  Environment Variables (Enterprise/Docker override)
    2.  `goconnect.yaml` (Encrypted/Protected Config File)
    3.  **Wizard Mode** (Default if 1 & 2 are missing)
*   **Setup Wizard:**
    *   If no config/DB is found, server starts HTTP on port `:8080`.
    *   Any request redirects to `/setup`.
    *   **Wizard Steps:**
        1.  **Mode Selection:** "Personal" (SQLite) vs "Enterprise" (Connect to Postgres).
        2.  **Admin Creation:** Define superuser.
        3.  **Finalize:** Generates keys, initializes DB, saves config, and **restarts the internal engine**.

### 2.3. Embedded Frontend
*   **Mechanism:** Go `embed`.
*   **Path:** `//go:embed web-ui/out/*`
*   **Routing:** `server/internal/handler/web.go` will serve static assets for `/`, `/dashboard`, etc., and API handlers for `/api/v1`.

## 3. Client Daemon: Service & Desktop Controller

### 3.1. Architecture: Separation of Concerns
Instead of one monolithic binary, we will use a **Controller/Service** model.
1.  **The Service (`goconnect-service`):**
    *   Headless, high-privilege (root/system) background process.
    *   Handles WireGuard networking, DNS, and connectivity.
    *   Starts automatically on boot.
    *   Exposes a **Local API** (via Unix Socket or localhost HTTP with mTLS) for the Controller.
2.  **The Controller (`goconnect-ui`):**
    *   User-mode desktop application (System Tray / Menu Bar).
    *   Connects to the Service's Local API.
    *   Displays status, prompts for login, manages settings.

### 3.2. Cross-Platform Service Management
*   **Library:** Use `github.com/kardianos/service`.
*   **Capabilities:**
    *   `goconnect install`: Registers the service (Windows Service, Systemd, Launchd).
    *   `goconnect start/stop`: Controls the background process.
*   **Flow:**
    *   Installer (MSI/Deb/Pkg) puts binaries in place.
    *   Post-install script runs `goconnect install`.

### 3.3. Encrypted Local Storage
*   **File:** `device.enc.json` (replacing plain JSON).
*   **Encryption:** AES-GCM 256-bit.
*   **Key Management:** OS Keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service).

## 4. Documentation Overhaul
The entire documentation set in `/docs` will be rewritten to reflect "Product" vs "Dev" usage.

*   `docs/GETTING_STARTED.md`: Replaces INSTALLATION.md. Focuses on "Download Installer -> Run Wizard".
*   `docs/DEPLOYMENT.md`: For Enterprise/Docker users (the old way).
*   `docs/ARCHITECTURE.md`: Updated to explain Service/Controller split and SQLite encryption.
*   `docs/SECURITY.md`: Detailed explanation of local storage encryption and key management.

## 5. Implementation Roadmap

### Phase 1: Server Core (Week 1)
*   [ ] Add `internal/repository/sqlite` implementation.
*   [ ] Implement Application-Level Field Encryption for SQLite.
*   [ ] Build "Setup Mode" state machine in Server.

### Phase 2: Client Architecture (Week 2)
*   [ ] Refactor `client-daemon` into `cmd/service` and `cmd/ui`.
*   [ ] Implement `kardianos/service` for unified service management.
*   [ ] Implement OS Keyring integration for config encryption.

### Phase 3: Polishing & Embedding (Week 3)
*   [ ] Embed Web UI into Server.
*   [ ] Create Desktop Tray Application (using `fyne` or `systray`).
*   [ ] Update all Documentation.

---

## Progress (Codex, latest)
- SQLite kalıcılığı sağlandı: user/tenant/network/membership/join_request/device/peer/chat/network_chat/ip_rule/invite_token/tenant_member/tenant_invite/tenant_announcement/tenant_chat. IPAM SQLite eklendi. Idempotency bilinçli olarak in-memory bırakıldı.
- `migrations_sqlite` PostgreSQL şemasına hizalandı (ör. `moderation_redacted`, `revoked_at`).
- Tüm server testleri yeşil: `go test ./...` geçiyor.
- Config için YAML dosyası desteği eklendi (`goconnect.yaml` varsayılan, `GOCONNECT_CONFIG_PATH` ile override). Dosya varsa boş olmayan env değerleri üstüne yazar, dosya yoksa eski env yükleme yoluna düşer. Setup modunda `/setup` GET status ve POST persist (YAML’e yaz) eklendi, config path status’te dönüyor. İlgili testler eklendi ve geçiyor.
- Setup modunda durum/validasyon endpointleri eklendi (`/setup/status`, `/setup/validate`), config dosyası mevcut/valid mi bilgisi ve validasyon hatası döndürülüyor.
- Setup adım listesi eklendi (`steps`), `/setup` yanıtında dönüyor; `/setup` POST artık `restart_required` döndürüyor (manuel restart akışını netleştirmek için).
- `/setup` POST `restart:true` gönderildiğinde HTTP yanıtı flushlandıktan sonra sunucu graceful shutdown (stop cancel) tetikleniyor.

## Açık işler (sonraki)
1) Setup wizard akışı (config yoksa /setup, adımlar + config yazma + restart otomasyonu).
2) Şifreleme/keyring (SQLCipher/field-level + go-keyring) ve config anahtar yönetimi.
3) Embedded web UI servisleme.
4) Idempotency için SQLite sürümü istenirse eklenebilir (şu an in-memory).
