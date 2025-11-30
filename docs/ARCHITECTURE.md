# 🏗️ GoConnect Architecture Documentation

This document explains GoConnect's technical architecture in detail.

---

## 📋 Contents

1. [Overview](#1-overview)
2. [Components](#2-components)
3. [Data Flow](#3-data-flow)
4. [Network Architecture](#4-network-architecture)
5. [Security](#5-security)
6. [Scalability](#6-scalability)

---

## 1. Overview

### 1.1 Design Philosophy

GoConnect is built on these principles:

| Principle | Description |
|-----------|-------------|
| **Simplicity** | Users need no technical knowledge |
| **Single App** | Both host and client in one application |
| **Cross-Platform** | Windows, macOS, Linux support |
| **P2P First** | Direct connection when possible, relay otherwise |
| **Security** | End-to-end encryption with WireGuard |

### 1.2 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER DEVICES                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐        │
│   │ GoConnect App │   │ GoConnect App │   │ GoConnect CLI │        │
│   │   (Host)      │   │   (Client)    │   │   (Client)    │        │
│   │               │   │               │   │               │        │
│   │ Windows/Mac   │   │ Windows/Mac   │   │    Linux      │        │
│   │   /Linux      │   │   /Linux      │   │   Server      │        │
│   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘        │
│           │                   │                   │                 │
│           │ WireGuard         │ WireGuard         │ WireGuard       │
│           │ Tunnel            │ Tunnel            │ Tunnel          │
│           │                   │                   │                 │
│           └─────────┬─────────┴─────────┬─────────┘                 │
│                     │                   │                           │
│                     ▼                   ▼                           │
│           ┌─────────────────────────────────────────┐               │
│           │        VIRTUAL LAN (10.0.1.0/24)        │               │
│           │                                         │               │
│           │   Host: 10.0.1.1                       │               │
│           │   Client1: 10.0.1.2                    │               │
│           │   Client2: 10.0.1.3                    │               │
│           │                                         │               │
│           └─────────────────────────────────────────┘               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                │ Coordination
                                │ (Signaling)
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      GOCONNECT INFRASTRUCTURE                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐  │
│   │ Signaling       │   │ STUN/TURN       │   │ Relay           │  │
│   │ Server          │   │ Servers         │   │ Servers         │  │
│   │                 │   │                 │   │                 │  │
│   │ - Peer discovery│   │ - NAT traversal │   │ - Last resort   │  │
│   │ - Invite links  │   │ - Public IP     │   │ - When P2P fails│  │
│   │ - Coordination  │   │   discovery     │   │                 │  │
│   └─────────────────┘   └─────────────────┘   └─────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. Components

### 2.1 GoConnect App (Tauri)

Desktop application providing both host and client functionality.

```
desktop-client/
├── src/                    # React Frontend
│   ├── components/         # UI components
│   │   ├── Sidebar.tsx     # Left sidebar
│   │   ├── NetworkList.tsx # Network list
│   │   ├── MemberList.tsx  # Member list
│   │   └── Chat.tsx        # Chat panel
│   ├── pages/              # Pages
│   │   ├── Home.tsx        # Home page
│   │   ├── Create.tsx      # Create network
│   │   └── Join.tsx        # Join network
│   ├── lib/                # Utilities
│   │   ├── api.ts          # API calls
│   │   └── wireguard.ts    # WG integration
│   └── App.tsx             # Main app
├── src-tauri/              # Rust Backend
│   ├── src/
│   │   ├── main.rs         # Entry point
│   │   ├── commands.rs     # Tauri commands
│   │   ├── wireguard.rs    # WireGuard management
│   │   └── network.rs      # Network operations
│   └── Cargo.toml
└── package.json
```

**Technologies:**
- Frontend: React 19, TypeScript, Tailwind CSS
- Backend: Rust, Tauri 2.0
- Package size: ~15 MB (1/10th of Electron)

### 2.2 GoConnect CLI

Terminal application with the same functionality.

```
client-daemon/
├── cmd/
│   └── daemon/
│       └── main.go         # Entry point
├── internal/
│   ├── tui/                # Terminal UI (Bubbletea)
│   │   ├── model.go        # TUI model
│   │   ├── views.go        # Screens
│   │   └── styles.go       # Styles
│   ├── network/            # Network management
│   ├── wireguard/          # WireGuard
│   └── config/             # Configuration
└── go.mod
```

**Technologies:**
- Language: Go 1.24+
- TUI: Bubbletea, Lipgloss
- Single binary, no dependencies

### 2.3 GoConnect Core

Shared Go library containing business logic.

```
server/
├── internal/
│   ├── network/
│   │   ├── network.go      # Network structure
│   │   ├── host.go         # Host functions
│   │   └── client.go       # Client functions
│   ├── wireguard/
│   │   ├── config.go       # WG configuration
│   │   ├── interface.go    # Interface management
│   │   └── peer.go         # Peer management
│   ├── auth/
│   │   ├── token.go        # JWT operations
│   │   └── invite.go       # Invite links
│   └── service/            # Business logic
└── go.mod
```

---

## 3. Data Flow

### 3.1 Network Creation Flow

```
┌──────────┐                              ┌──────────────┐
│  Host    │                              │  Signaling   │
│  App     │                              │  Server      │
└────┬─────┘                              └──────┬───────┘
     │                                           │
     │  1. Click "Create Network"                │
     │────────────────────────────────────────▶  │
     │                                           │
     │  2. Network info + WG public key          │
     │────────────────────────────────────────▶  │
     │                                           │
     │  3. Network ID + Invite link              │
     │◀────────────────────────────────────────  │
     │                                           │
     │  4. Create local WireGuard interface      │
     │  ┌─────────────────────────────┐          │
     │  │ wg0: 10.0.1.1/24            │          │
     │  │ private key: xxx            │          │
     │  │ listen port: 51820          │          │
     │  └─────────────────────────────┘          │
     │                                           │
     │  5. Establish WebSocket connection        │
     │◀────────────────────────────────────────▶ │
     │     (for peer updates)                    │
     │                                           │
```

### 3.2 Network Join Flow

```
┌──────────┐      ┌──────────────┐      ┌──────────┐
│  Client  │      │  Signaling   │      │   Host   │
│  App     │      │  Server      │      │   App    │
└────┬─────┘      └──────┬───────┘      └────┬─────┘
     │                   │                   │
     │ 1. Paste invite   │                   │
     │    link           │                   │
     │                   │                   │
     │ 2. Join request   │                   │
     │──────────────────▶│                   │
     │                   │ 3. Notify host    │
     │                   │──────────────────▶│
     │                   │                   │
     │                   │ 4. Accept/Reject  │
     │                   │◀──────────────────│
     │                   │                   │
     │ 5. Peer info      │                   │
     │◀──────────────────│                   │
     │   (host IP,       │                   │
     │    public key,    │                   │
     │    endpoint)      │                   │
     │                   │                   │
     │ 6. Establish WG   │                   │
     │═══════════════════╪═══════════════════│
     │     WireGuard P2P Connection          │
     │═══════════════════╪═══════════════════│
     │                   │                   │
```

### 3.3 NAT Traversal Flow

```
┌──────────┐      ┌──────────────┐      ┌──────────┐
│  Peer A  │      │  STUN/TURN   │      │  Peer B  │
│ (NAT ✓)  │      │  Server      │      │ (NAT ✓)  │
└────┬─────┘      └──────┬───────┘      └────┬─────┘
     │                   │                   │
     │ 1. Binding req    │                   │
     │──────────────────▶│                   │
     │                   │                   │
     │ 2. Public IP:Port │                   │
     │◀──────────────────│                   │
     │   (203.0.113.1:   │                   │
     │    54321)         │                   │
     │                   │                   │
     │                   │ 3. Binding req    │
     │                   │◀──────────────────│
     │                   │                   │
     │                   │ 4. Public IP:Port │
     │                   │──────────────────▶│
     │                   │   (198.51.100.1:  │
     │                   │    12345)         │
     │                   │                   │
     │ 5. Exchange endpoints via Signaling   │
     │◀─────────────────────────────────────▶│
     │                   │                   │
     │ 6. UDP hole punch │                   │
     │═══════════════════╪═══════════════════│
     │     Direct P2P Connection             │
     │═══════════════════╪═══════════════════│
```

---

## 4. Network Architecture

### 4.1 WireGuard Configuration

**Host Side:**
```ini
[Interface]
PrivateKey = <host_private_key>
Address = 10.0.1.1/24
ListenPort = 51820

[Peer]
# Client 1
PublicKey = <client1_public_key>
AllowedIPs = 10.0.1.2/32
Endpoint = <client1_endpoint>

[Peer]
# Client 2
PublicKey = <client2_public_key>
AllowedIPs = 10.0.1.3/32
Endpoint = <client2_endpoint>
```

**Client Side:**
```ini
[Interface]
PrivateKey = <client_private_key>
Address = 10.0.1.2/24

[Peer]
# Host
PublicKey = <host_public_key>
AllowedIPs = 10.0.1.0/24
Endpoint = <host_endpoint>
PersistentKeepalive = 25
```

### 4.2 IP Addressing

| Network Type | Subnet | Usage |
|--------------|--------|-------|
| Default | 10.0.x.0/24 | Normal networks |
| Large | 10.0.x.0/16 | 65K+ devices |
| Custom | User-defined | Advanced |

**IP Assignment:**
- Host: Always `.1` (e.g., 10.0.1.1)
- Clients: Sequential `.2`, `.3`, `.4`...
- Broadcast: `.255` (e.g., 10.0.1.255)

---

## 5. Security

### 5.1 Encryption

| Layer | Protocol | Description |
|-------|----------|-------------|
| Tunnel | WireGuard | ChaCha20-Poly1305 |
| Key Exchange | Noise Protocol | Curve25519 |
| Signaling | TLS 1.3 | HTTPS/WSS |
| Invite Links | JWT | HS256 signed |

### 5.2 Key Management

```
┌─────────────────────────────────────────────────────┐
│                   Key Lifecycle                      │
├─────────────────────────────────────────────────────┤
│                                                     │
│  1. Generation                                      │
│     └─▶ wg genkey > private.key                    │
│     └─▶ wg pubkey < private.key > public.key       │
│                                                     │
│  2. Storage                                         │
│     └─▶ Private key: OS keychain                   │
│         - Windows: Credential Manager              │
│         - macOS: Keychain                          │
│         - Linux: Secret Service                    │
│                                                     │
│  3. Exchange                                        │
│     └─▶ Public key: Via signaling server           │
│                                                     │
│  4. Rotation                                        │
│     └─▶ Every 30 days (optional)                   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 5.3 Threat Model

| Threat | Protection |
|--------|------------|
| Man-in-the-middle | WireGuard public key verification |
| Replay attack | Nonce-based encryption |
| Unauthorized access | Invite link + approval system |
| Brute force | Rate limiting + CAPTCHA |

---

## 6. Scalability

### 6.1 Single Network Limits

| Metric | Limit | Note |
|--------|-------|------|
| Member count | 65,534 | /16 subnet |
| Concurrent connections | ~1,000 | Depends on host capacity |
| Bandwidth | Unlimited* | P2P, excludes relay |

### 6.2 Federation (Future)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  GoConnect  │     │  GoConnect  │     │  GoConnect  │
│  Instance 1 │◀───▶│  Instance 2 │◀───▶│  Instance 3 │
│  (Region A) │     │  (Region B) │     │  (Region C) │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       └───────────────────┴───────────────────┘
                           │
                    Federation Protocol
                    (ActivityPub-like)
```

---

<div align="center">

**[← Home](../README.md)** | **[User Guide →](USER_GUIDE.md)**

</div>
