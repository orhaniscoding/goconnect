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

# 🎯 SYSTEM ARCHITECTURE OVERVIEW

## Core Concepts (Hiyerarşi):

```
TENANT (Server/Organization)
├── Visibility: public / private
├── Access: open / password / invite_only
├── Settings: network_creation rules, default_network, etc.
├── Announcements (Admin only yazabilir)
├── General Chat (Tüm üyeler)
├── Members (owner > admin > moderator > vip > member)
│
└── NETWORKS (VPN Ağları)
    ├── Visibility: public / private
    ├── JoinPolicy: open / invite / approval
    ├── RequiredRole: member / vip / admin / owner
    ├── Network Chat
    │
    └── MEMBERSHIPS
        └── User → Network connection with role
```

## Entity Relationships:

```
┌─────────────┐     1:N      ┌─────────────────┐
│   TENANT    │─────────────►│ TENANT_MEMBERS  │
│             │              │ (user_id, role) │
└──────┬──────┘              └────────┬────────┘
       │                              │
       │ 1:N                          │ N:1
       ▼                              ▼
┌─────────────┐              ┌─────────────┐
│   NETWORK   │              │    USER     │
│             │              │             │
└──────┬──────┘              └──────┬──────┘
       │                            │
       │ 1:N                        │ 1:N
       ▼                            ▼
┌─────────────┐              ┌─────────────┐
│ MEMBERSHIP  │◄─────────────│   DEVICE    │
│ (network)   │              │             │
└─────────────┘              └──────┬──────┘
                                    │
                                    │ 1:N (per network)
                                    ▼
                             ┌─────────────┐
                             │    PEER     │
                             │ (WireGuard) │
                             └─────────────┘
```

## Key Design Decisions:
1. **User can join multiple Tenants** (N:N via tenant_members)
2. **User has a "default" tenant_id** for backwards compatibility
3. **Network inherits Tenant's tenant_id**
4. **Network access can be restricted by role** (required_role field)
5. **Invite codes work for both Tenant and Network level**

---

# ✅ IMPLEMENTED FEATURES (v2.12.0+)

## Authentication & Security (COMPLETE)
- **JWT Authentication**: Access tokens (15min) + Refresh tokens (7d)
- **TOTP 2FA**: Time-based One-Time Password with QR code setup
- **Recovery Codes**: 10 single-use codes for account recovery
- **SSO/OIDC**: External identity provider support
- **Password Hashing**: Argon2id algorithm
- **Rate Limiting**: All endpoints protected

## Network Management (COMPLETE)
- **Network CRUD**: Create, read, update, delete networks
- **Peer Management**: Device peers with WireGuard keys
- **Invite Tokens**: Time-limited, usage-limited tokens (PostgreSQL repo)
- **IP Rules**: Allow/block CIDR ranges per tenant (PostgreSQL repo)

## Multi-Tenancy Base (COMPLETE)
- **Tenant Model**: ID, Name, CreatedAt, UpdatedAt
- **User belongs to Tenant**: user.TenantID (single tenant - legacy)
- **Network belongs to Tenant**: network.TenantID
- **Role-Based Access Control**: via `server/internal/rbac/`

## Web UI (COMPLETE)
- **Next.js 14 App Router** with i18n (tr/en)
- **Dashboard**: Network list, status
- **Settings Page**: Profile, 2FA setup, Recovery Codes management
- **Login/Register**: Public pages
- **Network Chat**: Real-time chat within networks (WebSocket)
- **Tenant Chat**: Real-time chat within tenants (WebSocket)

## Infrastructure (COMPLETE)
- **PostgreSQL Repositories**: User, Tenant, Network, Session, RecoveryCode, InviteToken, IPRule
- **Redis**: Session caching, rate limiting
- **Prometheus Metrics**: `/metrics` endpoint
- **WebSocket**: Real-time peer updates
- **Client Daemon**: Windows/Linux/macOS service files

---

# 📋 CURRENT DOMAIN MODELS (server/internal/domain/)

```go
// User - system user
type User struct {
    ID             string     `json:"id"`
    Email          string     `json:"email"`
    PasswordHash   string     `json:"-"`
    DisplayName    string     `json:"display_name"`
    Role           string     `json:"role"`          // system_admin, user
    TenantID       string     `json:"tenant_id"`     // default tenant (legacy)
    TOTPSecret     string     `json:"-"`             // 2FA secret
    TOTPEnabled    bool       `json:"totp_enabled"`
    RecoveryUsed   int        `json:"recovery_used"` // count of used recovery codes
    LastLoginAt    *time.Time `json:"last_login_at"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

// Tenant - organization/server container
type Tenant struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Network - VPN network within a tenant
type Network struct {
    ID          string    `json:"id"`
    TenantID    string    `json:"tenant_id"`
    Name        string    `json:"name"`
    Subnet      string    `json:"subnet"`      // e.g., "10.0.0.0/24"
    ListenPort  int       `json:"listen_port"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Device - user's device (computer/phone)
type Device struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Name      string    `json:"name"`
    OS        string    `json:"os"`           // windows, linux, macos, ios, android
    CreatedAt time.Time `json:"created_at"`
}

// Peer - WireGuard peer in a network
type Peer struct {
    ID           string    `json:"id"`
    NetworkID    string    `json:"network_id"`
    DeviceID     string    `json:"device_id"`
    UserID       string    `json:"user_id"`
    PublicKey    string    `json:"public_key"`
    AllowedIPs   string    `json:"allowed_ips"` // assigned IP in network
    LastSeen     time.Time `json:"last_seen"`
    CreatedAt    time.Time `json:"created_at"`
}

// InviteToken - network invitation (PostgreSQL repo ready)
type InviteToken struct {
    ID          string     `json:"id"`
    NetworkID   string     `json:"network_id"`
    Token       string     `json:"token"`
    MaxUses     int        `json:"max_uses"`
    UseCount    int        `json:"use_count"`
    ExpiresAt   *time.Time `json:"expires_at"`
    CreatedBy   string     `json:"created_by"`
    CreatedAt   time.Time  `json:"created_at"`
    RevokedAt   *time.Time `json:"revoked_at"`
}

// IPRule - IP allow/block rules (PostgreSQL repo ready)
type IPRule struct {
    ID          string     `json:"id"`
    TenantID    string     `json:"tenant_id"`
    CIDR        string     `json:"cidr"`          // e.g., "192.168.1.0/24"
    Action      string     `json:"action"`        // "allow" or "block"
    Description string     `json:"description"`
    CreatedBy   string     `json:"created_by"`
    ExpiresAt   *time.Time `json:"expires_at"`
    CreatedAt   time.Time  `json:"created_at"`
}

// RecoveryCode - 2FA backup codes
type RecoveryCode struct {
    ID        string     `json:"id"`
    UserID    string     `json:"user_id"`
    CodeHash  string     `json:"-"`             // bcrypt hashed
    UsedAt    *time.Time `json:"used_at"`
    CreatedAt time.Time  `json:"created_at"`
}

// Session - user login session
type Session struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    RefreshToken string    `json:"-"`
    UserAgent    string    `json:"user_agent"`
    IPAddress    string    `json:"ip_address"`
    ExpiresAt    time.Time `json:"expires_at"`
    CreatedAt    time.Time `json:"created_at"`
}

// Membership - user's membership in a network (current)
type Membership struct {
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    NetworkID  string    `json:"network_id"`
    Role       string    `json:"role"`          // owner, admin, member
    JoinedAt   time.Time `json:"joined_at"`
}
```

---

# ✅ IMPLEMENTED: TENANT MULTI-MEMBERSHIP SYSTEM (v2.12.0)

## Overview
Multi-tenant membership system (Discord-like model) - **FULLY IMPLEMENTED**.

## New Domain Models (TO BE ADDED)

```go
// TenantRole - user's role within a tenant
type TenantRole string
const (
    TenantRoleOwner     TenantRole = "owner"     // Full control
    TenantRoleAdmin     TenantRole = "admin"     // Manage users, networks
    TenantRoleModerator TenantRole = "moderator" // Manage chat, announcements
    TenantRoleVIP       TenantRole = "vip"       // Premium access
    TenantRoleMember    TenantRole = "member"    // Basic access
)

// TenantVisibility - who can see/find the tenant
type TenantVisibility string
const (
    TenantVisibilityPublic  TenantVisibility = "public"  // Discoverable
    TenantVisibilityPrivate TenantVisibility = "private" // Invite only
)

// TenantAccessType - how users join
type TenantAccessType string
const (
    TenantAccessOpen       TenantAccessType = "open"        // Anyone can join
    TenantAccessPassword   TenantAccessType = "password"    // Requires password
    TenantAccessInviteOnly TenantAccessType = "invite_only" // Requires invite
)

// ENHANCED Tenant - with new fields
type Tenant struct {
    ID           string           `json:"id"`
    Name         string           `json:"name"`
    Description  string           `json:"description"`
    IconURL      string           `json:"icon_url"`
    Visibility   TenantVisibility `json:"visibility"`    // public/private
    AccessType   TenantAccessType `json:"access_type"`   // open/password/invite_only
    PasswordHash string           `json:"-"`             // for password access
    MaxMembers   int              `json:"max_members"`   // 0 = unlimited
    MemberCount  int              `json:"member_count"`  // computed
    OwnerID      string           `json:"owner_id"`
    CreatedAt    time.Time        `json:"created_at"`
    UpdatedAt    time.Time        `json:"updated_at"`
}

// NEW: TenantMember - user's membership in a tenant (N:N)
type TenantMember struct {
    ID        string     `json:"id"`
    TenantID  string     `json:"tenant_id"`
    UserID    string     `json:"user_id"`
    Role      TenantRole `json:"role"`
    Nickname  string     `json:"nickname"`    // tenant-specific display name
    JoinedAt  time.Time  `json:"joined_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    BannedAt  *time.Time `json:"banned_at"`   // When user was banned
    BannedBy  string     `json:"banned_by"`   // Who banned the user
}

// NEW: TenantInvite - tenant-level invitation (like Steam friend codes)
type TenantInvite struct {
    ID        string     `json:"id"`
    TenantID  string     `json:"tenant_id"`
    Code      string     `json:"code"`        // short code like "ABC123"
    MaxUses   int        `json:"max_uses"`
    UseCount  int        `json:"use_count"`
    ExpiresAt *time.Time `json:"expires_at"`
    CreatedBy string     `json:"created_by"`
    CreatedAt time.Time  `json:"created_at"`
}

// ENHANCED Network - with role restrictions
type Network struct {
    ID           string     `json:"id"`
    TenantID     string     `json:"tenant_id"`
    Name         string     `json:"name"`
    Description  string     `json:"description"`
    Subnet       string     `json:"subnet"`
    ListenPort   int        `json:"listen_port"`
    RequiredRole TenantRole `json:"required_role"` // minimum role to access
    IsHidden     bool       `json:"is_hidden"`     // hidden from list
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

// NEW: TenantAnnouncement - admin announcements
type TenantAnnouncement struct {
    ID        string    `json:"id"`
    TenantID  string    `json:"tenant_id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    AuthorID  string    `json:"author_id"`
    IsPinned  bool      `json:"is_pinned"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// NEW: TenantChatMessage - tenant general chat
type TenantChatMessage struct {
    ID        string    `json:"id"`
    TenantID  string    `json:"tenant_id"`
    UserID    string    `json:"user_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    EditedAt  *time.Time `json:"edited_at"`
}
```

## New Database Tables (Migration File)

```sql
-- 000007_tenant_multi_membership.sql

-- Enhance tenants table
ALTER TABLE tenants
ADD COLUMN description TEXT DEFAULT '',
ADD COLUMN icon_url TEXT DEFAULT '',
ADD COLUMN visibility VARCHAR(20) DEFAULT 'private',
ADD COLUMN access_type VARCHAR(20) DEFAULT 'invite_only',
ADD COLUMN password_hash TEXT,
ADD COLUMN max_members INTEGER DEFAULT 0,
ADD COLUMN owner_id UUID NOT NULL REFERENCES users(id);

-- Tenant members (N:N relationship)
CREATE TABLE tenant_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    nickname VARCHAR(100),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

-- Tenant invites (Steam-like codes)
CREATE TABLE tenant_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(20) UNIQUE NOT NULL,
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tenant announcements
CREATE TABLE tenant_announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id),
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tenant chat messages
CREATE TABLE tenant_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    edited_at TIMESTAMP WITH TIME ZONE
);

-- Enhance networks table
ALTER TABLE networks
ADD COLUMN description TEXT DEFAULT '',
ADD COLUMN required_role VARCHAR(20) DEFAULT 'member',
ADD COLUMN is_hidden BOOLEAN DEFAULT FALSE;

-- Indexes
CREATE INDEX idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX idx_tenant_members_tenant ON tenant_members(tenant_id);
CREATE INDEX idx_tenant_invites_code ON tenant_invites(code);
CREATE INDEX idx_tenant_announcements_tenant ON tenant_announcements(tenant_id);
CREATE INDEX idx_tenant_chat_tenant ON tenant_chat_messages(tenant_id, created_at);
CREATE INDEX idx_tenants_visibility ON tenants(visibility);
```

## Planned API Endpoints (~30 new)

```
# Tenant Discovery
GET  /api/tenants/public             # List public tenants (paginated)
GET  /api/tenants/search?q=          # Search public tenants

# Tenant Membership
POST /api/tenants/{id}/join          # Join open tenant
POST /api/tenants/join-by-code       # Join via invite code
POST /api/tenants/{id}/leave         # Leave tenant
GET  /api/users/me/tenants           # List my tenants

# Tenant Management (owner/admin)
PATCH /api/tenants/{id}              # Update tenant settings
DELETE /api/tenants/{id}             # Delete tenant

# Member Management
GET  /api/tenants/{id}/members       # List members
PATCH /api/tenants/{id}/members/{uid}  # Update member role
DELETE /api/tenants/{id}/members/{uid} # Remove member
POST /api/tenants/{id}/members/{uid}/ban # Ban member
DELETE /api/tenants/{id}/members/{uid}/ban # Unban member

# Tenant Invites
POST /api/tenants/{id}/invites       # Create invite
GET  /api/tenants/{id}/invites       # List invites
DELETE /api/tenants/{id}/invites/{iid} # Revoke invite

# Announcements
POST /api/tenants/{id}/announcements # Create announcement
GET  /api/tenants/{id}/announcements # List announcements
PATCH /api/tenants/{id}/announcements/{aid} # Update
DELETE /api/tenants/{id}/announcements/{aid} # Delete

# General Chat (WebSocket primarily, REST for history)
GET  /api/tenants/{id}/chat/messages # Get chat history
POST /api/tenants/{id}/chat/messages # Post message (REST fallback)
DELETE /api/tenants/{id}/chat/messages/{mid} # Delete message

# WebSocket Messages (new types)
tenant:chat:message   # Real-time chat
tenant:chat:delete    # Message deleted
tenant:announcement   # New announcement
tenant:member:join    # Member joined
tenant:member:leave   # Member left
tenant:member:update  # Role changed
```

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
| Add domain model | `server/internal/domain/*.go` |
| Add repository | `server/internal/repository/*.go` |
| Add service | `server/internal/service/*.go` |
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
| Add database migration | `server/migrations/000XXX_*.sql` |
| Add UI component | `web-ui/src/components/*.tsx` |
| Add protected page | `web-ui/src/app/[locale]/(protected)/*` |
| Add public page | `web-ui/src/app/[locale]/(public)/*` |

---

# 21) 📦 EXISTING REPOSITORY IMPLEMENTATIONS

| Repository | Interface | PostgreSQL | In-Memory | Tests |
|------------|-----------|------------|-----------|-------|
| UserRepository | ✅ | ✅ | ✅ | ✅ |
| TenantRepository | ✅ | ✅ | ✅ | ✅ |
| NetworkRepository | ✅ | ✅ | ✅ | ✅ |
| SessionRepository | ✅ | ✅ | ✅ | ✅ |
| RecoveryCodeRepository | ✅ | ✅ | ✅ | ✅ |
| InviteTokenRepository | ✅ | ✅ | ✅ | ✅ (21 tests) |
| IPRuleRepository | ✅ | ✅ | ✅ | ✅ (24 tests) |
| TenantMemberRepository | ✅ | ✅ | ✅ | ✅ (30 tests) |
| TenantInviteRepository | ✅ | ✅ | ✅ | ✅ (26 tests) |
| TenantAnnouncementRepository | ✅ | ✅ | ✅ | ✅ (21 tests) |
| TenantChatRepository | ✅ | ✅ | ✅ | ✅ (24 tests) |
| DeviceRepository | ✅ | ✅ | ✅ | ✅ (35 tests) |
| PeerRepository | ✅ | ✅ | ✅ | ✅ (50 tests) |
| MembershipRepository | ✅ | ✅ | ✅ | ✅ (15 tests) |

Legend: ✅ = Complete, ⏳ = Needs implementation/tests

---

# 22) 🌐 WEB UI PAGES

| Route | File | Auth | Description |
|-------|------|------|-------------|
| `/[locale]/login` | `(public)/login/page.tsx` | No | Login page |
| `/[locale]/register` | `(public)/register/page.tsx` | No | Registration |
| `/[locale]/dashboard` | `(protected)/dashboard/page.tsx` | Yes | Network list |
| `/[locale]/networks/[id]` | `(protected)/networks/[id]/page.tsx` | Yes | Network details |
| `/[locale]/networks/[id]/chat` | `(protected)/networks/[id]/chat/page.tsx` | Yes | Network chat |
| `/[locale]/settings` | `(protected)/settings/page.tsx` | Yes | User settings (2FA, Recovery) |
| `/[locale]/tenants` | `(protected)/tenants/page.tsx` | Yes | Tenant discovery |
| `/[locale]/tenants/[id]` | `(protected)/tenants/[id]/page.tsx` | Yes | Tenant details |
| `/[locale]/tenants/[id]/chat` | `(protected)/tenants/[id]/chat/page.tsx` | Yes | Tenant chat |

---

# 23) 🔑 AUTHENTICATION FLOW

```
1. Login: POST /api/auth/login
   ├─ Email/Password → returns JWT + refresh token
   ├─ If 2FA enabled → returns { requires_2fa: true, temp_token }
   └─ Then: POST /api/auth/verify-2fa with TOTP code

2. Token Refresh: POST /api/auth/refresh
   └─ Refresh token → new JWT

3. Recovery Flow:
   POST /api/auth/recovery
   ├─ temp_token + recovery_code
   └─ Returns new JWT (2FA bypassed)
```

---

# 24) 📊 TEST COVERAGE

| Package | Tests | Coverage |
|---------|-------|----------|
| `repository/invite_token_test.go` | 21 | High |
| `repository/ip_rule_test.go` | 24 | High |
| `repository/tenant_member_test.go` | 30 | High |
| `repository/tenant_invite_test.go` | 26 | High |
| `repository/tenant_announcement_test.go` | 21 | High |
| `repository/tenant_chat_test.go` | 24 | High |
| `repository/device_test.go` | 35 | High |
| `repository/peer_test.go` | 50 | High |
| `repository/membership_test.go` | 15 | High |
| `service/ip_rule_test.go` | 22 | High |
| `service/tenant_membership_test.go` | 30 | High |
| `service/auth_test.go` | ~15 | Medium |
| `service/recovery_test.go` | ~10 | Medium |

---

**Last Updated:** 2025-11-26 | **Version:** v2.12.0+
