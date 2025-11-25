# Development Workflow Policy

Agreed single-thread flow on `main` branch.

---

## 🔀 Branching Rules
1. **STRICTLY NO BRANCHING.** All development occurs on `main`.
2. No feature branches. No merge commits.
3. Commit often, push when stable.

---

## 📝 Commit Strategy
1. Keep commits focused and logically grouped.
2. Use Conventional Commit prefixes (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
3. Ensure code compiles and tests pass before pushing to `main`.

---

## ⚠️ Conflict Avoidance
1. Pull `main` frequently (`git pull --rebase`) to stay up to date.
2. If a conflict occurs, resolve it locally before pushing.

---

## 🔧 Tooling Expectations

### Before Every Push:
```bash
# Server
cd server
go build ./...
go test ./... -race
golangci-lint run

# Web UI
cd web-ui
npm run typecheck
npm run build
```

### Required Tool Versions:
| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.24+ | Required for latest features |
| golangci-lint | v1.64.8+ | Required for Go 1.24 compatibility |
| Node.js | 20 LTS | For web-ui |
| npm | 10+ | Comes with Node 20 |

---

## 📁 Files to Update by Change Type

### API Changes:
- `server/openapi/openapi.yaml` - OpenAPI spec
- `docs/API_EXAMPLES.http` - HTTP examples
- `docs/CONFIG_FLAGS.md` - If new env vars added

### WebSocket Changes:
- `docs/WEBSOCKET_API.md` - Message types
- `docs/WS_MESSAGES.md` - Detailed message format

### UI Changes:
- `web-ui/src/locales/en/common.json` - English strings
- `web-ui/src/locales/tr/common.json` - Turkish strings
- Both files MUST be updated together!

### Security Changes:
- `docs/SECURITY.md` - Security practices
- `docs/THREAT_MODEL.md` - Threat analysis
- `docs/SSO_2FA.md` - Auth changes

### Schema/Storage Changes:
- `docs/AUDIT_NOTES.md` - Audit schema changes
- `server/internal/database/migrations/` - New migrations
- `docs/POSTGRESQL_SETUP.md` - If setup changes

---

## 🐳 Docker-Specific Rules

### web-ui/next.config.js MUST have:
```javascript
output: 'standalone',
```

### web-ui/public/ folder MUST exist:
- Keep `.gitkeep` file if folder is empty
- Docker COPY command requires this folder

---

## 🌐 Internationalization Rules

When adding UI text:
1. Add key to BOTH `en/common.json` and `tr/common.json`
2. Use consistent key naming: `section.subsection.key`
3. Support variables: `"text": "Hello {name}"`

---

## 🧪 Testing Requirements

| Change Type | Required Tests |
|-------------|----------------|
| New endpoint | Unit + integration test |
| Bug fix | Regression test |
| Security fix | Security test case |
| Business logic | Unit tests |

Use `sync/atomic` for shared counters in concurrent tests to avoid data races.

---

## 🚀 Release Process

Releases are **fully automated**:
1. Push with conventional commit → Release Please creates PR
2. Merge Release Please PR → GoReleaser builds assets
3. Docker images auto-pushed to ghcr.io

**DO NOT** manually create releases or tags.

---

## 📊 CI Pipeline Checks

All pushes trigger:
- `ci.yml` - Build & test
- `lint.yml` - golangci-lint + TypeScript
- `codeql.yml` - Security analysis
- `release-please.yml` - Version management

---

Revision: v3 – Date: 2025-11-25
