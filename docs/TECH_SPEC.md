# GoConnect — Teknik Proje Dökümanı (FINAL • Ultra-Detay)

Ürün: **GoConnect** · Yapımcı: **orhaniscoding** · Lisans: **MIT** · © 2025  
Binaries: `goconnect-server`, `goconnect-daemon` (Win: `.exe`) · `--version` çıktısı: `goconnect-<name> vX.Y.Z (commit <hash>, build <date>) built by orhaniscoding`

## 0) Amaç & Kapsam
Host (state), Daemon (WG uygulayıcı), Web UI (user+admin). Hamachi/ZeroTier benzeri.

## 1) Alan Adı/CORS/CSP
Web: https://app.goconnect.example · API/WS: https://api.goconnect.example (`/v1`, `/v1/ws`) · Bridge: http://127.0.0.1:<port>  
CORS: sadece web origin; CSP connect-src: API/WS + 127.0.0.1 (dev: localhost serbest).

## 2) Yığın & İlkeler
Go 1.22+, Node 20+, Postgres 15+, Redis 7+, Next 14+. ULID/UTC/zap/OTEL/Prom. Argon2, JWT+JWKS rotation.  
Idempotency-Key (POST mutasyon zorunlu, TTL24h, gövde değişirse 409).  
Hata: {code,message,details?,retry_after?}.

## 3) Repo
README, LICENSE, copilot-instructions, go.work, goreleaser.yml, workflows, scripts, docs/*, server/*, client-daemon/*, web-ui/* (tam liste root’ta).

## 4) Host (Server)
REST/WS, Networks+Memberships, WG profile (IPAM), Chat (edit/soft/hard-delete, redact, history), RBAC, Rate-limit, Audit immutable, Outbox+Redis fan-out, Multi-tenant.

## 5) Daemon
WG (Win: WireGuardNT; macOS/Linux: wg-quick). Localhost Bridge: /status, /wg/apply, /wg/down, /peers (OAuth2+PKCE + custom scheme; X-Loopback-Token 10dk).

## 6) Web UI
Next.js, TR/EN i18n, A11Y, footer “Built by orhaniscoding”. User+Admin sayfaları.

## 7) RBAC
owner/admin/moderator/member; hem REST hem WS’de policy + test.

## 8) Veri Modeli
tenants, users, devices, networks(visibility/join_policy/cidr/dns/mtu/split), memberships(status/role/device_pubkey), wg_peers(assigned_ip/allowed_ips/endpoint_hint/rotation), chat_messages(deleted_at,redacted,redaction_mask), chat_message_edits, audit_logs, bans, invite_tokens, idem_keys.

## 9) IPAM & WG
CIDR overlap→409 ERR_CIDR_OVERLAP. Profil: Address,DNS,MTU,Keepalive=25; Peer(server pubkey/endpoint), AllowedIPs.

## 10) REST (özet)
Auth (register/login/refresh/logout); Networks (create/list/get/patch/delete); Memberships (join/approve/deny/kick/ban/get); WireGuard (keypair/profile/rotate); Chat (list/send/edit/delete/redact); Audit list.

## 11) WS
Inbound(op_id): auth.refresh, chat.send|edit|delete|redact. Outbound: chat.message/edited/deleted/redacted, member.joined/left, request.*, admin.kick|ban, net.updated. Backpressure (oda), oldest-drop.

## 12) Güvenlik
JWT kısa; JWKS rotation+grace; Argon2; RBAC; rate-limit; GDPR; log redaction; IP allow/deny; SSO/2FA (OIDC, TOTP/WebAuthn).

## 13) Observability
Prometheus: http/ws/errors_total{code}, outbox/dlq. OTEL: ipam.allocate, wg.profile.render, ws.broadcast. zap JSON + correlation_id.

## 14) Test
Unit, Integration (PG+Redis+Server), E2E (Playwright), WS harness (op_id), Contract (Schemathesis), Fuzz, Load (k6), Chaos (Toxiproxy). Coverage ≥70%.

## 15) CI/CD & Paketleme
CI: build+test; README otomasyonu. Release Please (PR+changelog). GoReleaser: çoklu OS, ldflags brand, nfpm deb/rpm, Scoop/Winget, macOS pkg (notarization pipeline). SBOM/cosign ileride.

## 16) OS Detayları
Linux: systemd unit; Windows: Service + Scoop/Winget manifest; macOS: LaunchDaemon + notarized pkg. Footer ve --version her OS’de brand içerir.

## 17) DoD
TECH_SPEC uyumu; test+docs+metrics; coverage≥70; Release Please yeşil; brand/binary isimleri sabit; i18n/A11Y kuralları sağlandı.
