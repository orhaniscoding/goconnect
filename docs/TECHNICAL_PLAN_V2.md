# GoConnect v2.0 - Kapsamlı Teknik Plan

> **Vizyon:** Discord + Tailscale/ZeroTier hibrit modeli
> **Tarih:** 2026-01-19
> **Durum:** Planlama Aşaması

---

## 1. Mimari Genel Bakış

### 1.1 Hiyerarşi Yapısı

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GoConnect Hiyerarşisi                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  SERVER (Sunucu/Topluluk)                                                   │
│  ├── Bağımsız instance (goconnect.io veya self-hosted)                     │
│  ├── Kendi kullanıcı veritabanı                                            │
│  ├── Server-wide roller ve yetkiler                                        │
│  │                                                                          │
│  ├── SECTION (Alt-Server / Bölüm) [YENİ]                                   │
│  │   ├── Büyük serverları organize etmek için                              │
│  │   ├── Kendi moderatörleri olabilir                                      │
│  │   ├── Arama ve filtreleme kolaylığı                                     │
│  │   │                                                                      │
│  │   ├── NETWORK (Ağ / LAN)                                                │
│  │   │   ├── WireGuard VPN ile P2P bağlantı                               │
│  │   │   ├── Ağa bağlı peers birbirini LAN'da görür                       │
│  │   │   ├── Ağ-specific roller                                            │
│  │   │   │                                                                  │
│  │   │   └── CHANNEL (Kanal)                                               │
│  │   │       ├── Text Channel (mesajlaşma)                                 │
│  │   │       ├── Voice Channel (sesli sohbet)                              │
│  │   │       └── Kanal-specific permission overrides                       │
│  │   │                                                                      │
│  │   └── CHANNEL (Section-level kanallar)                                  │
│  │       └── Ağa bağlı olmadan erişilebilir                               │
│  │                                                                          │
│  └── CHANNEL (Server-level kanallar)                                       │
│      └── #duyurular, #kurallar gibi genel kanallar                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Kimlik Mimarisi (Hibrit Model)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Hibrit Kimlik Modeli                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    OAuth Providers                                   │   │
│  │         Google    GitHub    Discord    Apple (gelecek)              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                  GoConnect Identity Service                          │   │
│  │                  (goconnect.io tarafından sunulur)                   │   │
│  │                                                                      │   │
│  │  Saklanan: OAuth ID, Display Name, Avatar, Email (opsiyonel)        │   │
│  │  Saklanmayan: Şifre, gerçek isim, hassas veri                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│         │                         │                         │               │
│         ▼                         ▼                         ▼               │
│  ┌─────────────┐          ┌─────────────┐          ┌─────────────┐         │
│  │ goconnect.io│          │ Self-Hosted │          │ Self-Hosted │         │
│  │  (Default)  │          │  Server A   │          │  Server B   │         │
│  │             │          │             │          │             │         │
│  │ GoConnect   │          │ Kendi Auth  │          │ GoConnect   │         │
│  │ Auth ZORUNLU│          │ (LDAP/SSO)  │          │ Auth KABUL  │         │
│  │             │          │ VEYA        │          │             │         │
│  │             │          │ GoConnect   │          │             │         │
│  └─────────────┘          └─────────────┘          └─────────────┘         │
│                                                                             │
│  NOT: Her server kendi kullanıcı profillerini saklar.                      │
│  GoConnect Identity sadece kimlik doğrulama için kullanılır.               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.3 Veri Akışı

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Veri Akış Diyagramı                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  VPN TRAFİĞİ (Tamamen P2P)                                                 │
│  ══════════════════════════                                                 │
│  ┌──────┐                                    ┌──────┐                      │
│  │Peer A│◄────────── WireGuard ─────────────►│Peer B│                      │
│  └──────┘          (Şifreli P2P)             └──────┘                      │
│      │                                           │                          │
│      └─────────────────┬─────────────────────────┘                          │
│                        │                                                    │
│                        ▼                                                    │
│                 ┌─────────────┐                                             │
│                 │   SERVER    │  ← Sadece: Peer discovery, Auth            │
│                 │ (Signaling) │    VPN trafiği buradan GEÇMEZ              │
│                 └─────────────┘                                             │
│                                                                             │
│  VOICE (WebRTC P2P)                                                        │
│  ═══════════════════                                                        │
│  ┌──────┐          ┌──────┐                                                │
│  │User A│◄── P2P ─►│User B│  (Mümkün olduğunda)                           │
│  └──────┘          └──────┘                                                │
│      │                 │                                                    │
│      └────────┬────────┘                                                    │
│               ▼                                                             │
│        ┌─────────────┐                                                      │
│        │ TURN Server │  ← Sadece NAT traversal başarısız olursa            │
│        └─────────────┘                                                      │
│                                                                             │
│  CHAT (Server-Side + E2E)                                                  │
│  ════════════════════════                                                   │
│  Public Channels  ──► Server DB (plaintext, searchable)                    │
│  DM'ler           ──► Server DB (E2E encrypted, server okuyamaz)           │
│  Ağ-içi Chat      ──► P2P (VPN üzerinden, server'a gitmez)                │
│                                                                             │
│  FILE TRANSFER                                                              │
│  ═════════════                                                              │
│  Ağ içinde     ──► Doğrudan P2P (WireGuard üzerinden)                     │
│  Ağ dışında    ──► WebRTC P2P (TURN fallback)                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Veritabanı Şeması

### 2.1 Yeni/Güncellenecek Tablolar

```sql
-- ═══════════════════════════════════════════════════════════════════════════
-- SECTION (Alt-Server / Bölüm) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    icon            VARCHAR(255),           -- Emoji veya URL
    position        INTEGER NOT NULL DEFAULT 0,

    -- Görünürlük
    visibility      VARCHAR(20) NOT NULL DEFAULT 'visible',  -- visible, hidden, archived

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,

    UNIQUE(server_id, name)
);

CREATE INDEX idx_sections_server ON sections(server_id);
CREATE INDEX idx_sections_position ON sections(server_id, position);

-- ═══════════════════════════════════════════════════════════════════════════
-- CHANNEL (Text/Voice Kanallar) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Hiyerarşi (biri dolu olmalı)
    server_id       UUID REFERENCES servers(id) ON DELETE CASCADE,
    section_id      UUID REFERENCES sections(id) ON DELETE CASCADE,
    network_id      UUID REFERENCES networks(id) ON DELETE CASCADE,

    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    type            VARCHAR(20) NOT NULL,   -- 'text', 'voice', 'announcement'
    position        INTEGER NOT NULL DEFAULT 0,

    -- Voice channel settings
    bitrate         INTEGER DEFAULT 64000,  -- 64kbps default
    user_limit      INTEGER DEFAULT 0,      -- 0 = unlimited

    -- Moderation
    slowmode        INTEGER DEFAULT 0,      -- Seconds between messages (0 = off)
    nsfw            BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,

    -- En az bir parent olmalı
    CONSTRAINT channel_has_parent CHECK (
        (server_id IS NOT NULL)::int +
        (section_id IS NOT NULL)::int +
        (network_id IS NOT NULL)::int = 1
    )
);

CREATE INDEX idx_channels_server ON channels(server_id) WHERE server_id IS NOT NULL;
CREATE INDEX idx_channels_section ON channels(section_id) WHERE section_id IS NOT NULL;
CREATE INDEX idx_channels_network ON channels(network_id) WHERE network_id IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════════
-- ROLE (Custom Roller) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,

    name            VARCHAR(100) NOT NULL,
    color           VARCHAR(7),             -- Hex color (#FF5733)
    icon            VARCHAR(255),           -- Emoji veya URL
    position        INTEGER NOT NULL,       -- Hiyerarşi (yüksek = güçlü)

    -- Sistem rolleri
    is_default      BOOLEAN DEFAULT FALSE,  -- Yeni üyelere otomatik atanır
    is_admin        BOOLEAN DEFAULT FALSE,  -- Tüm yetkilere sahip

    -- Mentionable
    mentionable     BOOLEAN DEFAULT FALSE,

    -- Görünürlük
    hoist           BOOLEAN DEFAULT FALSE,  -- Üye listesinde ayrı göster

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(server_id, name)
);

CREATE INDEX idx_roles_server ON roles(server_id);
CREATE INDEX idx_roles_position ON roles(server_id, position DESC);

-- ═══════════════════════════════════════════════════════════════════════════
-- PERMISSION (Granular Yetkiler) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

-- Permission tanımları (enum-like table)
CREATE TABLE permission_definitions (
    id              VARCHAR(100) PRIMARY KEY,   -- 'server.manage', 'channel.send_messages'
    category        VARCHAR(50) NOT NULL,       -- 'server', 'network', 'channel', 'member'
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    default_value   BOOLEAN DEFAULT FALSE
);

-- Role-level permissions
CREATE TABLE role_permissions (
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   VARCHAR(100) NOT NULL REFERENCES permission_definitions(id),
    allowed         BOOLEAN NOT NULL,           -- true = izin ver, false = reddet

    PRIMARY KEY (role_id, permission_id)
);

-- Channel-level permission overrides
CREATE TABLE channel_permission_overrides (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,

    -- Role VEYA User için override (biri dolu)
    role_id         UUID REFERENCES roles(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,

    permission_id   VARCHAR(100) NOT NULL REFERENCES permission_definitions(id),
    allowed         BOOLEAN,                    -- true = izin, false = reddet, NULL = inherit

    CONSTRAINT override_target CHECK (
        (role_id IS NOT NULL)::int + (user_id IS NOT NULL)::int = 1
    ),
    UNIQUE(channel_id, role_id, permission_id),
    UNIQUE(channel_id, user_id, permission_id)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- USER_ROLES (Kullanıcı-Rol İlişkisi) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE user_roles (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by     UUID REFERENCES users(id),

    PRIMARY KEY (user_id, role_id)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- MESSAGE (Geliştirilmiş) - GÜNCELLEME
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    author_id       UUID NOT NULL REFERENCES users(id),

    content         TEXT NOT NULL,

    -- Reply/Thread
    reply_to_id     UUID REFERENCES messages(id),
    thread_id       UUID REFERENCES messages(id),    -- Thread parent

    -- Attachments (JSON array)
    attachments     JSONB DEFAULT '[]',

    -- Embeds (link previews, rich content)
    embeds          JSONB DEFAULT '[]',

    -- Mentions
    mentions        UUID[] DEFAULT '{}',             -- @user mentions
    mention_roles   UUID[] DEFAULT '{}',             -- @role mentions
    mention_everyone BOOLEAN DEFAULT FALSE,

    -- Flags
    pinned          BOOLEAN DEFAULT FALSE,
    edited_at       TIMESTAMP,

    -- E2E encryption (DM'ler için)
    encrypted       BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP
);

CREATE INDEX idx_messages_channel ON messages(channel_id, created_at DESC);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_thread ON messages(thread_id) WHERE thread_id IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════════
-- REACTION (Emoji Tepkileri) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE reactions (
    message_id      UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji           VARCHAR(100) NOT NULL,      -- Unicode emoji veya custom emoji ID
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (message_id, user_id, emoji)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- VOICE_STATE (Ses Kanalı Durumu) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE voice_states (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,

    -- Durum
    self_mute       BOOLEAN DEFAULT FALSE,
    self_deaf       BOOLEAN DEFAULT FALSE,
    server_mute     BOOLEAN DEFAULT FALSE,      -- Mod tarafından susturuldu
    server_deaf     BOOLEAN DEFAULT FALSE,      -- Mod tarafından sağır edildi

    -- Bağlantı
    connected_at    TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- FRIEND (Arkadaşlık) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE friendships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, accepted, blocked

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    accepted_at     TIMESTAMP,

    UNIQUE(user_id, friend_id),
    CHECK(user_id != friend_id)
);

CREATE INDEX idx_friendships_user ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend ON friendships(friend_id, status);

-- ═══════════════════════════════════════════════════════════════════════════
-- DM_CHANNEL (Direkt Mesaj) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE dm_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            VARCHAR(20) NOT NULL DEFAULT 'dm',  -- 'dm', 'group_dm'

    -- Group DM için
    name            VARCHAR(100),
    icon            VARCHAR(255),
    owner_id        UUID REFERENCES users(id),

    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE dm_channel_members (
    channel_id      UUID NOT NULL REFERENCES dm_channels(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Ayarlar
    muted           BOOLEAN DEFAULT FALSE,

    joined_at       TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (channel_id, user_id)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- PRESENCE (Online Durumu) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE user_presence (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    status          VARCHAR(20) NOT NULL DEFAULT 'offline',  -- online, idle, dnd, invisible, offline
    custom_status   VARCHAR(128),

    -- Aktivite
    activity_type   VARCHAR(50),        -- playing, listening, watching, streaming
    activity_name   VARCHAR(128),

    -- Bağlantı
    last_seen       TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Cihaz durumları
    desktop_status  VARCHAR(20),
    mobile_status   VARCHAR(20),
    web_status      VARCHAR(20)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- SERVER_DISCOVERY (Sunucu Keşfi) - YENİ
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE server_discovery (
    server_id       UUID PRIMARY KEY REFERENCES servers(id) ON DELETE CASCADE,

    -- Discovery ayarları
    enabled         BOOLEAN DEFAULT FALSE,

    -- Kategoriler
    category        VARCHAR(50),            -- gaming, education, music, etc.
    tags            VARCHAR(50)[] DEFAULT '{}',

    -- Açıklama
    short_description VARCHAR(300),

    -- İstatistikler (cache)
    member_count    INTEGER DEFAULT 0,
    online_count    INTEGER DEFAULT 0,

    -- Öne çıkarma
    featured        BOOLEAN DEFAULT FALSE,
    verified        BOOLEAN DEFAULT FALSE,

    -- Timestamps
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_discovery_category ON server_discovery(category) WHERE enabled = TRUE;
CREATE INDEX idx_discovery_featured ON server_discovery(featured) WHERE enabled = TRUE;
```

### 2.2 Permission Definitions (Seed Data)

```sql
-- Server Permissions
INSERT INTO permission_definitions (id, category, name, description, default_value) VALUES
('server.manage', 'server', 'Manage Server', 'Edit server settings, name, icon', FALSE),
('server.delete', 'server', 'Delete Server', 'Permanently delete the server', FALSE),
('server.view_audit_log', 'server', 'View Audit Log', 'View server audit log', FALSE),
('server.manage_roles', 'server', 'Manage Roles', 'Create, edit, delete roles', FALSE),
('server.manage_channels', 'server', 'Manage Channels', 'Create, edit, delete channels', FALSE),
('server.manage_sections', 'server', 'Manage Sections', 'Create, edit, delete sections', FALSE),
('server.kick_members', 'server', 'Kick Members', 'Remove members from server', FALSE),
('server.ban_members', 'server', 'Ban Members', 'Ban members from server', FALSE),
('server.create_invite', 'server', 'Create Invite', 'Create server invite links', TRUE),

-- Network Permissions
('network.create', 'network', 'Create Network', 'Create VPN networks', FALSE),
('network.manage', 'network', 'Manage Network', 'Edit network settings', FALSE),
('network.delete', 'network', 'Delete Network', 'Delete networks', FALSE),
('network.connect', 'network', 'Connect to Network', 'Join VPN networks', TRUE),
('network.kick_peers', 'network', 'Kick Peers', 'Remove peers from network', FALSE),
('network.ban_peers', 'network', 'Ban Peers', 'Ban peers from network', FALSE),
('network.approve_join', 'network', 'Approve Join Requests', 'Approve pending join requests', FALSE),

-- Channel Permissions
('channel.view', 'channel', 'View Channel', 'See the channel', TRUE),
('channel.send_messages', 'channel', 'Send Messages', 'Send messages in text channels', TRUE),
('channel.embed_links', 'channel', 'Embed Links', 'Links show preview', TRUE),
('channel.attach_files', 'channel', 'Attach Files', 'Upload files', TRUE),
('channel.add_reactions', 'channel', 'Add Reactions', 'Add emoji reactions', TRUE),
('channel.mention_everyone', 'channel', 'Mention Everyone', 'Use @everyone and @here', FALSE),
('channel.manage_messages', 'channel', 'Manage Messages', 'Delete any message, pin messages', FALSE),
('channel.manage_threads', 'channel', 'Manage Threads', 'Manage thread settings', FALSE),

-- Voice Permissions
('voice.connect', 'voice', 'Connect', 'Join voice channels', TRUE),
('voice.speak', 'voice', 'Speak', 'Talk in voice channels', TRUE),
('voice.mute_members', 'voice', 'Mute Members', 'Server mute other members', FALSE),
('voice.deafen_members', 'voice', 'Deafen Members', 'Server deafen other members', FALSE),
('voice.move_members', 'voice', 'Move Members', 'Move members between channels', FALSE),
('voice.priority_speaker', 'voice', 'Priority Speaker', 'Be heard over others', FALSE);
```

---

## 3. API Tasarımı

### 3.1 Yeni Endpoints

```yaml
# ═══════════════════════════════════════════════════════════════════════════
# SECTION API
# ═══════════════════════════════════════════════════════════════════════════

POST   /api/v2/servers/{server_id}/sections
GET    /api/v2/servers/{server_id}/sections
GET    /api/v2/sections/{section_id}
PATCH  /api/v2/sections/{section_id}
DELETE /api/v2/sections/{section_id}
PATCH  /api/v2/sections/{section_id}/position

# ═══════════════════════════════════════════════════════════════════════════
# CHANNEL API
# ═══════════════════════════════════════════════════════════════════════════

# Server-level channels
POST   /api/v2/servers/{server_id}/channels
GET    /api/v2/servers/{server_id}/channels

# Section-level channels
POST   /api/v2/sections/{section_id}/channels
GET    /api/v2/sections/{section_id}/channels

# Network-level channels
POST   /api/v2/networks/{network_id}/channels
GET    /api/v2/networks/{network_id}/channels

# Channel operations
GET    /api/v2/channels/{channel_id}
PATCH  /api/v2/channels/{channel_id}
DELETE /api/v2/channels/{channel_id}

# Channel permissions
GET    /api/v2/channels/{channel_id}/permissions
PUT    /api/v2/channels/{channel_id}/permissions/{target_id}  # role or user
DELETE /api/v2/channels/{channel_id}/permissions/{target_id}

# ═══════════════════════════════════════════════════════════════════════════
# ROLE API
# ═══════════════════════════════════════════════════════════════════════════

POST   /api/v2/servers/{server_id}/roles
GET    /api/v2/servers/{server_id}/roles
GET    /api/v2/roles/{role_id}
PATCH  /api/v2/roles/{role_id}
DELETE /api/v2/roles/{role_id}
PATCH  /api/v2/roles/{role_id}/position

# Role permissions
GET    /api/v2/roles/{role_id}/permissions
PUT    /api/v2/roles/{role_id}/permissions

# Assign roles to users
PUT    /api/v2/servers/{server_id}/members/{user_id}/roles/{role_id}
DELETE /api/v2/servers/{server_id}/members/{user_id}/roles/{role_id}

# ═══════════════════════════════════════════════════════════════════════════
# MESSAGE API
# ═══════════════════════════════════════════════════════════════════════════

GET    /api/v2/channels/{channel_id}/messages
POST   /api/v2/channels/{channel_id}/messages
GET    /api/v2/channels/{channel_id}/messages/{message_id}
PATCH  /api/v2/channels/{channel_id}/messages/{message_id}
DELETE /api/v2/channels/{channel_id}/messages/{message_id}

# Reactions
PUT    /api/v2/channels/{channel_id}/messages/{message_id}/reactions/{emoji}
DELETE /api/v2/channels/{channel_id}/messages/{message_id}/reactions/{emoji}
GET    /api/v2/channels/{channel_id}/messages/{message_id}/reactions/{emoji}

# Pins
PUT    /api/v2/channels/{channel_id}/pins/{message_id}
DELETE /api/v2/channels/{channel_id}/pins/{message_id}
GET    /api/v2/channels/{channel_id}/pins

# ═══════════════════════════════════════════════════════════════════════════
# VOICE API
# ═══════════════════════════════════════════════════════════════════════════

POST   /api/v2/channels/{channel_id}/voice/join
POST   /api/v2/channels/{channel_id}/voice/leave
PATCH  /api/v2/channels/{channel_id}/voice/state   # mute, deaf, etc.
GET    /api/v2/channels/{channel_id}/voice/states  # who's in channel

# WebRTC Signaling (WebSocket)
WS     /api/v2/voice/signaling

# ═══════════════════════════════════════════════════════════════════════════
# DM API
# ═══════════════════════════════════════════════════════════════════════════

GET    /api/v2/users/@me/channels              # List DM channels
POST   /api/v2/users/{user_id}/dm              # Create/Get DM channel
POST   /api/v2/channels/group-dm               # Create group DM

# ═══════════════════════════════════════════════════════════════════════════
# FRIEND API
# ═══════════════════════════════════════════════════════════════════════════

GET    /api/v2/users/@me/friends
POST   /api/v2/users/@me/friends               # Send friend request
DELETE /api/v2/users/@me/friends/{user_id}     # Remove friend
PUT    /api/v2/users/@me/friends/{user_id}     # Accept request
POST   /api/v2/users/@me/blocks/{user_id}      # Block user
DELETE /api/v2/users/@me/blocks/{user_id}      # Unblock user

# ═══════════════════════════════════════════════════════════════════════════
# PRESENCE API
# ═══════════════════════════════════════════════════════════════════════════

PATCH  /api/v2/users/@me/presence              # Update own presence
GET    /api/v2/users/{user_id}/presence        # Get user presence

# ═══════════════════════════════════════════════════════════════════════════
# SERVER DISCOVERY API
# ═══════════════════════════════════════════════════════════════════════════

GET    /api/v2/discovery/servers               # List discoverable servers
GET    /api/v2/discovery/servers/search        # Search servers
GET    /api/v2/discovery/categories            # List categories
PATCH  /api/v2/servers/{server_id}/discovery   # Update discovery settings
```

### 3.2 WebSocket Events

```yaml
# ═══════════════════════════════════════════════════════════════════════════
# GATEWAY EVENTS
# ═══════════════════════════════════════════════════════════════════════════

# Connection
HELLO                   # Server greeting
HEARTBEAT               # Keep alive
HEARTBEAT_ACK           # Heartbeat response
IDENTIFY                # Client auth
READY                   # Initial state

# Message Events
MESSAGE_CREATE          # New message
MESSAGE_UPDATE          # Message edited
MESSAGE_DELETE          # Message deleted
MESSAGE_REACTION_ADD    # Reaction added
MESSAGE_REACTION_REMOVE # Reaction removed
TYPING_START            # User typing

# Channel Events
CHANNEL_CREATE          # New channel
CHANNEL_UPDATE          # Channel edited
CHANNEL_DELETE          # Channel deleted

# Voice Events
VOICE_STATE_UPDATE      # User voice state changed
VOICE_SERVER_UPDATE     # Voice server info (for WebRTC)

# Presence Events
PRESENCE_UPDATE         # User presence changed

# Member Events
MEMBER_JOIN             # User joined server
MEMBER_LEAVE            # User left server
MEMBER_UPDATE           # Member updated (roles, etc.)

# Network Events
NETWORK_PEER_JOIN       # Peer connected to VPN
NETWORK_PEER_LEAVE      # Peer disconnected from VPN
NETWORK_PEER_UPDATE     # Peer info changed

# Relationship Events
FRIEND_REQUEST          # Friend request received
FRIEND_ACCEPT           # Friend request accepted
FRIEND_REMOVE           # Friend removed
```

---

## 4. Implementation Phases

### Phase 1: Core Infrastructure (3 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 1: TEMEL ALTYAPI                                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 1: Database & Models                                                  │
│  ─────────────────────────                                                  │
│  □ Migration dosyaları oluştur                                             │
│  □ Domain modelleri güncelle (Section, Channel, Role, Permission)          │
│  □ Repository implementasyonları                                            │
│  □ Unit testler                                                             │
│                                                                             │
│  Week 2: Permission System                                                  │
│  ──────────────────────────                                                 │
│  □ Permission resolver (rol + channel override hesaplama)                  │
│  □ RBAC middleware güncellemesi                                             │
│  □ Permission check fonksiyonları                                          │
│  □ Integration testler                                                      │
│                                                                             │
│  Week 3: Channel System                                                     │
│  ───────────────────────                                                    │
│  □ Channel CRUD API                                                        │
│  □ Section CRUD API                                                        │
│  □ Message API güncellemesi (channel-based)                                │
│  □ WebSocket events                                                         │
│  □ E2E testler                                                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Phase 2: Voice & Real-time (2 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 2: VOICE & REAL-TIME                                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 4: Voice Channels                                                     │
│  ───────────────────────                                                    │
│  □ Voice state management                                                   │
│  □ WebRTC signaling server                                                  │
│  □ TURN integration (mevcut)                                               │
│  □ Voice permissions                                                        │
│                                                                             │
│  Week 5: Enhanced WebSocket                                                 │
│  ──────────────────────────                                                 │
│  □ Gateway refactor (Discord-like)                                         │
│  □ Presence system                                                          │
│  □ Typing indicators                                                        │
│  □ Connection state management                                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Phase 3: Social Features (2 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 3: SOCIAL FEATURES                                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 6: DM & Friends                                                       │
│  ─────────────────────                                                      │
│  □ DM channel system                                                        │
│  □ Friend system (add, accept, block)                                      │
│  □ E2E encryption for DMs                                                  │
│  □ Group DM                                                                 │
│                                                                             │
│  Week 7: Notifications & Misc                                               │
│  ────────────────────────────                                               │
│  □ Notification preferences                                                 │
│  □ @mentions                                                                │
│  □ Reactions                                                                │
│  □ Message pins                                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Phase 4: Discovery & OAuth (2 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 4: DISCOVERY & OAUTH                                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 8: OAuth Integration                                                  │
│  ──────────────────────────                                                 │
│  □ Google OAuth                                                             │
│  □ GitHub OAuth                                                             │
│  □ Discord OAuth                                                            │
│  □ Account linking                                                          │
│                                                                             │
│  Week 9: Server Discovery                                                   │
│  ─────────────────────────                                                  │
│  □ Discovery API                                                            │
│  □ Search & filter                                                          │
│  □ Categories & tags                                                        │
│  □ Invite system (link, code, QR)                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Phase 5: Desktop UI (2 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 5: DESKTOP UI                                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 10: Core UI Components                                                │
│  ────────────────────────────                                               │
│  □ Discord-like layout                                                      │
│  □ Server/Section/Channel sidebar                                          │
│  □ Message list & input                                                    │
│  □ Voice channel UI                                                         │
│                                                                             │
│  Week 11: Additional UI                                                     │
│  ───────────────────────                                                    │
│  □ Settings panels                                                          │
│  □ Role management UI                                                       │
│  □ Server discovery UI                                                      │
│  □ DM & Friends UI                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Phase 6: Polish & Launch (2 hafta)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PHASE 6: POLISH & LAUNCH                                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Week 12: Testing & Fixes                                                   │
│  ─────────────────────────                                                  │
│  □ Cross-platform testing (Win/Linux/macOS)                                │
│  □ Performance optimization                                                 │
│  □ Bug fixes                                                                │
│  □ Security audit                                                           │
│                                                                             │
│  Week 13: Launch Prep                                                       │
│  ─────────────────────                                                      │
│  □ goconnect.io deployment                                                 │
│  □ Documentation                                                            │
│  □ Beta launch                                                              │
│  □ Monitoring setup                                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Monetization Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    FREEMIUM MODEL (goconnect.io)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  🆓 FREE TIER                                                               │
│  ═════════════                                                              │
│  • 1 Section oluşturabilir                                                  │
│  • Ağ başına max 5 peer                                                     │
│  • Max 10 kanal/ağ                                                          │
│  • 50MB dosya upload limiti                                                 │
│  • Standart voice kalitesi (64kbps)                                        │
│  • Temel roller                                                             │
│                                                                             │
│  💎 PRO TIER ($5/ay veya $50/yıl)                                          │
│  ═══════════════════════════════                                            │
│  • Sınırsız Section                                                         │
│  • Ağ başına max 25 peer                                                    │
│  • Sınırsız kanal                                                           │
│  • 500MB dosya upload                                                       │
│  • Yüksek voice kalitesi (256kbps)                                         │
│  • Custom roller + emojiler                                                 │
│  • Öncelikli destek                                                         │
│                                                                             │
│  🏢 SERVER BOOST ($10/ay)                                                   │
│  ═══════════════════════                                                    │
│  • Ağ başına max 100 peer                                                   │
│  • Boosted badge                                                            │
│  • Discovery'de öne çıkma                                                   │
│  • Custom vanity URL                                                        │
│                                                                             │
│  🏗️ ENTERPRISE (Özel fiyat)                                                │
│  ═══════════════════════════                                                │
│  • Self-hosted kurulum desteği                                              │
│  • SSO/SAML entegrasyonu                                                    │
│  • SLA garantisi                                                            │
│  • Özel özellikler                                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Güvenlik Kontrol Listesi

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         GÜVENLİK CHECKLIST                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  □ Authentication                                                           │
│    ├── OAuth token validation                                               │
│    ├── JWT expiry (kısa süreli access, uzun süreli refresh)               │
│    ├── Device management                                                    │
│    └── Rate limiting                                                        │
│                                                                             │
│  □ Authorization                                                            │
│    ├── Permission check her endpoint'te                                     │
│    ├── Resource ownership validation                                        │
│    ├── Role hierarchy enforcement                                           │
│    └── Channel permission override calculation                              │
│                                                                             │
│  □ Input Validation                                                         │
│    ├── Server-side validation (client'a güvenme)                           │
│    ├── SQL injection prevention (parameterized queries)                    │
│    ├── XSS prevention (output encoding)                                    │
│    └── File upload validation (type, size)                                 │
│                                                                             │
│  □ Encryption                                                               │
│    ├── TLS everywhere                                                       │
│    ├── WireGuard (VPN traffic)                                             │
│    ├── E2E encryption (DMs)                                                │
│    └── At-rest encryption (sensitive data)                                 │
│                                                                             │
│  □ Privacy                                                                  │
│    ├── Minimal data collection                                              │
│    ├── GDPR/KVKK compliance                                                │
│    ├── Data export feature                                                  │
│    └── Account deletion                                                     │
│                                                                             │
│  □ Abuse Prevention                                                         │
│    ├── Rate limiting (API, WebSocket, messages)                            │
│    ├── Spam detection                                                       │
│    ├── Report system                                                        │
│    └── Auto-mod (optional)                                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Sonraki Adım

Bu plan onaylandıktan sonra, **Phase 1: Week 1 - Database & Models** ile başlayacağız:

1. Migration dosyaları oluşturma
2. Domain modelleri güncelleme
3. Repository implementasyonları
4. Unit testler

---

**Hazırlayan:** Claude Code
**Tarih:** 2026-01-19
**Versiyon:** 2.0-DRAFT
