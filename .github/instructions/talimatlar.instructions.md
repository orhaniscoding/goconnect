The following instructions are **always active** while the AI is working on this project.
The AI must automatically follow these rules in **every** response.

---

# 🏗️ PROJECT OVERVIEW - GoConnect

GoConnect is a **WireGuard-based VPN management system** with:
- **Server** (Go): REST API, WebSocket, PostgreSQL/SQLite, Redis
- **Client Daemon** (Go): System service for VPN connectivity
- **Web UI** (Next.js 14): React dashboard with i18n support
- **Docker Images**: ghcr.io/orhaniscoding/goconnect-server, goconnect-webui

**Current Version**: Check `.release-please-manifest.json` for latest version

---

# 1) CORE ENGINEERING PRINCIPLES

* The AI must always choose **the most logical, necessary, and correct** development step based on the current state of the project.
* The AI acts as a **Developer, Architect, QA Engineer, Security Engineer, DevOps Engineer, and Release Manager** simultaneously.
* Unnecessary complexity, unnecessary files, unnecessary operations, and unnecessary commits are **strictly forbidden**.
* Code quality, maintainability, readability, performance, and security are **top priorities**.

---

# 2) GIT RULES (STRICTLY NO BRANCHING)

* All development must occur **exclusively on the main branch**.
* No new branches are created. No merges. No feature branches.
* Commit/push only when meaningful and necessary.

### Commit message format (Conventional Commits):

feat:     (New feature)
fix:      (Bug fix)
docs:     (Documentation only)
refactor: (Code change that neither fixes a bug nor adds a feature)
chore:    (Build process, aux tools, release preparation)
test:     (Adding missing tests or correcting existing tests)
report:   (Adding analysis reports)

---

# 3) 📁 PROJECT STRUCTURE (MUST MAINTAIN)

```
goconnect/
├── .github/
│   ├── instructions/           # AI instruction files
│   │   └── talimatlar.instructions.md  # THIS FILE - primary AI guide
│   └── workflows/              # GitHub Actions CI/CD
│       ├── ci.yml              # Tests, builds, coverage
│       ├── lint.yml            # golangci-lint, TypeScript checks
│       ├── release-please.yml  # Auto-release with GoReleaser & Docker
│       ├── goreleaser.yml      # Binary builds (19 assets)
│       ├── codeql.yml          # Security scanning
│       ├── security-scan.yml   # Dependency scanning
│       └── auto-readme.yml     # README auto-generation
├── server/                     # Go backend server
│   ├── cmd/server/main.go      # Entry point
│   ├── internal/               # Private packages
│   │   ├── audit/              # Audit logging with SQLite
│   │   ├── config/             # Configuration management
│   │   ├── database/           # PostgreSQL, SQLite, Redis clients
│   │   ├── domain/             # Domain models
│   │   ├── handler/            # HTTP handlers (Gin)
│   │   ├── metrics/            # Prometheus metrics
│   │   ├── rbac/               # Role-based access control
│   │   ├── repository/         # Data access layer
│   │   ├── service/            # Business logic
│   │   ├── websocket/          # Real-time communication
│   │   └── wireguard/          # WireGuard key/profile management
│   ├── go.mod                  # Go module dependencies
│   ├── .golangci.yml           # Linter configuration
│   └── Dockerfile              # Server container build
├── client-daemon/              # Go VPN client daemon
│   ├── cmd/daemon/main.go      # Entry point
│   ├── internal/               # Daemon logic
│   └── service/                # OS service files
│       ├── windows/install.ps1           # Windows Service installer
│       ├── linux/goconnect-daemon.service # systemd unit file
│       └── macos/com.goconnect.daemon.plist # launchd plist
├── web-ui/                     # Next.js 14 frontend
│   ├── src/
│   │   ├── app/                # App Router pages
│   │   │   ├── [locale]/       # i18n routing (tr/en)
│   │   │   │   ├── (protected)/ # Auth-required pages
│   │   │   │   └── (public)/   # Public pages (login, register)
│   │   ├── components/         # React components
│   │   ├── lib/                # Utilities
│   │   │   ├── api.ts          # API client functions
│   │   │   ├── bridge.ts       # Daemon bridge communication
│   │   │   ├── i18n.ts         # Internationalization config
│   │   │   └── i18n-context.tsx # Translation context & useT hook
│   │   └── locales/            # Translation files
│   │       ├── en/common.json  # English translations
│   │       └── tr/common.json  # Turkish translations
│   ├── public/                 # Static assets (needs .gitkeep)
│   ├── next.config.js          # Next.js config (output: 'standalone' required!)
│   ├── Dockerfile              # WebUI container build
│   └── package.json            # Node dependencies
├── docs/                       # Documentation
│   ├── INSTALLATION.md         # Complete setup guides (all platforms)
│   ├── SECURITY.md             # Security hardening guide
│   ├── CONFIG_FLAGS.md         # Environment variables
│   ├── RELEASE_PROCESS.md      # Release workflow
│   └── ...                     # Other docs
├── goreleaser.yml              # GoReleaser config (release notes here!)
├── release-please-config.json  # Release Please config
├── .release-please-manifest.json # Version tracking
├── CHANGELOG.md                # Auto-generated changelog
└── README.md                   # Auto-generated from docs/README.tpl.md
```

---

# 4) 🔧 TECHNOLOGY STACK (VERSIONS MATTER!)

## Backend (Server & Daemon)
| Component | Version | Config File |
|-----------|---------|-------------|
| Go | 1.24+ | `go.mod` |
| Gin | Latest | `go.mod` |
| golangci-lint | v1.64.8+ | `.golangci.yml`, `.github/workflows/lint.yml` |
| PostgreSQL | 15+ | `docs/POSTGRESQL_SETUP.md` |
| Redis | 7+ | `docs/CONFIG_FLAGS.md` |
| WireGuard | Latest | System requirement |

## Frontend (Web UI)
| Component | Version | Config File |
|-----------|---------|-------------|
| Node.js | 20 LTS | `.github/workflows/*.yml` |
| Next.js | 14 | `web-ui/package.json` |
| React | 18 | `web-ui/package.json` |
| TypeScript | 5+ | `web-ui/tsconfig.json` |

## CI/CD
| Tool | Purpose | Config |
|------|---------|--------|
| GitHub Actions | CI/CD | `.github/workflows/` |
| Release Please | Versioning | `release-please-config.json` |
| GoReleaser | Binary builds | `goreleaser.yml` |
| Docker Buildx | Multi-arch images | `release-please.yml` |

---

# 5) 📋 MANDATORY FILES TO UPDATE

## When Adding New Features:
| File | Update Required |
|------|-----------------|
| `docs/CONFIG_FLAGS.md` | New environment variables |
| `docs/API_EXAMPLES.http` | New API endpoints |
| `docs/WEBSOCKET_API.md` | New WebSocket messages |
| `server/openapi/openapi.yaml` | OpenAPI spec updates |
| `web-ui/src/locales/*/common.json` | New UI strings (BOTH en & tr!) |

## When Changing Security:
| File | Update Required |
|------|-----------------|
| `docs/SECURITY.md` | Security practices |
| `docs/THREAT_MODEL.md` | New threat considerations |
| `docs/SSO_2FA.md` | Auth changes |

## When Changing Architecture:
| File | Update Required |
|------|-----------------|
| `docs/TECH_SPEC.md` | Architecture changes |
| `docs/LOCAL_BRIDGE.md` | Daemon communication |
| This file | Project structure changes |

## When Releasing:
| File | Auto-Updated By |
|------|-----------------|
| `CHANGELOG.md` | Release Please |
| `.release-please-manifest.json` | Release Please |
| `README.md` | auto-readme.yml workflow |

---

# 6) 🚀 RELEASE SYSTEM (FULLY AUTOMATED)

## Release Flow:
```
1. Push to main with conventional commit (feat:/fix:/etc.)
   ↓
2. Release Please creates/updates PR with version bump
   ↓
3. Merge Release Please PR
   ↓
4. GoReleaser builds 19 binary assets
   ↓
5. Docker images pushed to ghcr.io
   ↓
6. GitHub Release created with detailed notes
```

## Release Notes Location:
**`goreleaser.yml`** contains release header and footer templates.
- Header: File descriptions table, package extensions guide
- Footer: Installation instructions for ALL platforms (Docker, Linux, macOS, Windows)

## When Updating Release Notes:
1. Edit `goreleaser.yml` → `release.header` and `release.footer`
2. Include setup for ALL platforms
3. Keep documentation links table updated

---

# 7) 🔨 BUILD & TEST COMMANDS

## Server
```bash
cd server
go build ./...           # Build
go test ./... -race      # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Linting (use v1.64.8+)
```

## Client Daemon
```bash
cd client-daemon
go build ./...
go test ./...
```

## Web UI
```bash
cd web-ui
npm ci                   # Install dependencies
npm run build            # Production build (creates .next/standalone)
npm run typecheck        # TypeScript check (tsc --noEmit)
npm run dev              # Development server
```

## Docker (Local Testing)
```bash
docker compose up -d     # Start all services
docker compose logs -f   # View logs
```

---

# 8) 🌐 INTERNATIONALIZATION (i18n)

## Translation Files:
- `web-ui/src/locales/en/common.json` - English
- `web-ui/src/locales/tr/common.json` - Turkish

## Rules:
1. **ALWAYS update BOTH files** when adding UI strings
2. Use dot notation for keys: `"section.subsection.key": "value"`
3. Nested objects are supported: `{ "section": { "key": "value" } }`
4. Variable interpolation: `"greeting": "Hello {name}"` → `t('greeting', { name: 'John' })`

## Translation Hook Usage:
```tsx
import { useT } from '@/lib/i18n-context'

function Component() {
  const t = useT()
  return <p>{t('key')}</p>                    // Simple
  return <p>{t('key', { var: 'value' })}</p>  // With variables
  return <p>{t('key', 'fallback')}</p>        // With fallback
}
```

---

# 9) 🐳 DOCKER CONFIGURATION

## Critical Settings:

### `web-ui/next.config.js` - MUST have:
```javascript
output: 'standalone',  // Required for Docker!
```

### `web-ui/Dockerfile` - Copies:
```dockerfile
COPY --from=builder /app/public ./public/
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
```

### `web-ui/public/` - MUST exist:
Keep `.gitkeep` file to ensure folder exists in git.

### `server/docker-compose.yaml` - Production Deployment:
```yaml
services:
  server:
    restart: always    # Auto-restart on crash/reboot
  db:
    restart: always
    volumes: [postgres_data:/var/lib/postgresql/data]  # Persistent!
  redis:
    restart: always
    volumes: [redis_data:/data]  # Persistent!
```

---

# 9.5) 🔧 SERVICE/DAEMON CONFIGURATION (ALL PLATFORMS)

## Overview - Auto-Start & Recovery:

| Platform | Method | Config File | Auto-Start | Crash Recovery |
|----------|--------|-------------|------------|----------------|
| **Windows** | Windows Service | `client-daemon/service/windows/install.ps1` | ✅ `StartupType Automatic` | ✅ 3x retry in 60s |
| **Linux** | systemd | `client-daemon/service/linux/goconnect-daemon.service` | ✅ `WantedBy=multi-user.target` | ✅ `Restart=always` |
| **macOS** | launchd | `client-daemon/service/macos/com.goconnect.daemon.plist` | ✅ `RunAtLoad=true` | ✅ `KeepAlive` |
| **Server (Docker)** | Docker Compose | `server/docker-compose.yaml` | ✅ `restart: always` | ✅ `restart: always` |

## Windows Service (`install.ps1`):
```powershell
# Key settings:
New-Service -Name "GoConnectDaemon" -StartupType Automatic
sc.exe failure GoConnectDaemon reset= 86400 actions= restart/60000/restart/60000/restart/60000
```
- Runs as LocalSystem
- Auto-restarts 3 times on failure (60s intervals)
- Resets failure count after 24h

## Linux systemd (`goconnect-daemon.service`):
```ini
[Service]
Restart=always
RestartSec=5
StartLimitInterval=0   # Unlimited restarts

[Install]
WantedBy=multi-user.target  # Start at boot
```
- Commands: `systemctl enable/start/stop goconnect-daemon`

## macOS launchd (`com.goconnect.daemon.plist`):
```xml
<key>RunAtLoad</key><true/>           <!-- Start at boot -->
<key>KeepAlive</key>
<dict>
    <key>SuccessfulExit</key><false/>  <!-- Restart if exits normally -->
    <key>Crashed</key><true/>          <!-- Restart if crashes -->
</dict>
<key>ThrottleInterval</key><integer>5</integer>  <!-- Min 5s between restarts -->
```
- Log files: `/var/log/goconnect-daemon.log`, `/var/log/goconnect-daemon.err`
- Runs as root (required for VPN operations)
- Commands: `launchctl load/unload /Library/LaunchDaemons/com.goconnect.daemon.plist`

## When Modifying Service Files:

| Change | Files to Update |
|--------|-----------------|
| Daemon binary path | All 3 service files + `goreleaser.yml` |
| Restart policy | Respective platform service file |
| Logging config | macOS plist (StandardOutPath), Linux service (journald) |
| Service name | All 3 service files + `docs/INSTALLATION.md` |

## CRITICAL Rules:
1. **ALL platforms must have equivalent functionality** - same auto-start, same crash recovery
2. **Test on all platforms** before release
3. **Update `docs/INSTALLATION.md`** when service install steps change
4. **Update `goreleaser.yml` release notes** to reflect service commands

---

# 10) 🧪 LINTING CONFIGURATION

## `server/.golangci.yml` - Key Exclusions:
- Test files: errcheck, unused, ineffassign excluded
- websocket/: errcheck excluded (fire-and-forget pattern)
- gocritic: exitAfterDefer, ifElseChain, elseif disabled
- Handler w.Write errors excluded

## `web-ui/` - TypeScript:
- `tsc --noEmit` for type checking
- Strict mode enabled

---

# 11) ⚠️ COMMON ISSUES & SOLUTIONS

| Issue | Cause | Solution |
|-------|-------|----------|
| Docker build fails: `/app/.next/standalone not found` | Missing `output: 'standalone'` | Add to `next.config.js` |
| Docker build fails: `/app/public not found` | Empty public folder | Add `.gitkeep` to `web-ui/public/` |
| golangci-lint fails with Go 1.24 | Old lint version | Use v1.64.8+ in `.github/workflows/lint.yml` |
| TypeScript error: nested translation | Wrong type in i18n.ts | Use `Record<string, any>` for Dictionary type |
| Data race in tests | Shared variable access | Use `sync/atomic` for counters |
| Release Please PR not created | No conventional commits | Use `feat:`, `fix:`, etc. |

---

# 12) 🔐 SECURITY REQUIREMENTS

## Environment Variables (NEVER commit!):
```
JWT_SECRET=<random-32-byte-base64>
DATABASE_URL=postgres://...
REDIS_URL=redis://...
OIDC_CLIENT_SECRET=...
```

## Password Hashing:
- Algorithm: Argon2id
- See `docs/SECURITY.md` for parameters

## API Security:
- Rate limiting on all endpoints
- JWT with short expiry (15min access, 7d refresh)
- CORS properly configured

---

# 13) 📝 DOCUMENTATION STANDARDS

## When Creating/Updating Docs:
1. Use clear headers with emoji indicators
2. Include code examples
3. Cover ALL platforms (Linux, macOS, Windows)
4. Keep tables for quick reference
5. Link to related docs

## README.md Auto-Generation:
- Template: `docs/README.tpl.md`
- Workflow: `.github/workflows/auto-readme.yml`
- Don't edit README.md directly!

---

# 14) ✅ PRE-COMMIT CHECKLIST

Before every commit, ensure:

- [ ] Code compiles: `go build ./...`
- [ ] Tests pass: `go test ./...`
- [ ] Lint passes: `golangci-lint run`
- [ ] TypeScript compiles: `npm run typecheck`
- [ ] Translations updated (both en & tr)
- [ ] Relevant docs updated
- [ ] Conventional commit message used

---

# 15) 🔄 KEEPING PROJECT IN SYNC

## After Any Change, Check:

### Code Changes:
- [ ] Tests added/updated?
- [ ] API docs updated? (`docs/API_EXAMPLES.http`)
- [ ] WebSocket docs updated? (`docs/WEBSOCKET_API.md`)
- [ ] Config docs updated? (`docs/CONFIG_FLAGS.md`)

### UI Changes:
- [ ] Both translation files updated?
- [ ] TypeScript types correct?
- [ ] Responsive design maintained?

### Security Changes:
- [ ] `docs/SECURITY.md` updated?
- [ ] `docs/THREAT_MODEL.md` reviewed?
- [ ] No secrets in code?

### Release Changes:
- [ ] `goreleaser.yml` release notes current?
- [ ] `docs/INSTALLATION.md` accurate?
- [ ] All platform instructions included?

---

# 16) 📊 METRICS & MONITORING

## Prometheus Metrics:
- Endpoint: `/metrics`
- Custom metrics in `server/internal/metrics/`

## Health Check:
- Endpoint: `/health`
- Returns: server status, DB connection, Redis connection

---

# 17) 🎯 RESPONSE GUIDELINES

Every AI response MUST:

1. **Analyze** the current situation
2. **Select** the most logical next step
3. **Implement** the change
4. **Update** related docs (if needed)
5. **Test** the change works
6. **Commit** with conventional message (if needed)
7. **Verify** CI passes

---

# 18) 📞 COMMUNICATION STYLE

* Short, clear, technical
* No unnecessary explanation
* Code examples when helpful
* Turkish or English based on user preference

---

# 19) 🚨 CRITICAL REMINDERS

1. **Never create branches** - main only
2. **Never skip translations** - both en & tr
3. **Never commit secrets** - use env vars
4. **Never ignore linting** - fix all errors
5. **Never forget Docker settings** - standalone output required
6. **Always test before push** - CI failures waste time
7. **Always update docs** - outdated docs cause confusion

---

# 20) 📚 QUICK REFERENCE - FILE LOCATIONS

| Need To... | Edit This File |
|------------|----------------|
| Add API endpoint | `server/internal/handler/*.go` |
| Add WebSocket message | `server/internal/websocket/handler.go` |
| Add translation | `web-ui/src/locales/*/common.json` |
| Add config option | `server/internal/config/config.go` + `docs/CONFIG_FLAGS.md` |
| Update release notes | `goreleaser.yml` |
| Fix linting rules | `server/.golangci.yml` |
| Update CI | `.github/workflows/*.yml` |
| Add documentation | `docs/*.md` |
| Modify Windows service | `client-daemon/service/windows/install.ps1` |
| Modify Linux service | `client-daemon/service/linux/goconnect-daemon.service` |
| Modify macOS service | `client-daemon/service/macos/com.goconnect.daemon.plist` |
| Change Docker deployment | `server/docker-compose.yaml` |
| Add daemon functionality | `client-daemon/cmd/daemon/main.go` |

---

**Last Updated:** 2025-11-25 | **Version:** v2.8.8+
