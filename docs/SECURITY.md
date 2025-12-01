# 🔐 GoConnect Security Guide

Bu doküman GoConnect'in güvenlik mimarisini, en iyi uygulamaları ve self-hosted kurulumlar için güvenlik yapılandırmasını açıklar.

## 🏗️ Security Architecture

### Zero-Trust Model

GoConnect, Zero-Trust güvenlik modeli üzerine inşa edilmiştir:

```
┌─────────────────────────────────────────────────────────────┐
│                     ZERO-TRUST LAYERS                        │
├─────────────────────────────────────────────────────────────┤
│  1. Identity Layer     │ WireGuard Ed25519 Keys             │
│  2. Transport Layer    │ WireGuard Encryption (ChaCha20)    │
│  3. Application Layer  │ JWT + Refresh Tokens               │
│  4. Authorization      │ Role-Based Access Control (RBAC)   │
│  5. Audit              │ Comprehensive Event Logging        │
└─────────────────────────────────────────────────────────────┘
```

### Encryption

| Layer | Algorithm | Purpose |
|-------|-----------|---------|
| Identity | Ed25519 | Key pairs for identity |
| Key Exchange | X25519 (Curve25519) | DH key agreement |
| Symmetric | ChaCha20-Poly1305 | Data encryption |
| Hashing | BLAKE2s | Fast hashing |
| Passwords | Argon2id | Password hashing |

---

## 🔑 Authentication

### JWT Token Flow

```
1. User Login → Server validates credentials
2. Server issues:
   - Access Token (15 min TTL)
   - Refresh Token (7 day TTL)
3. Client stores tokens securely
4. Access Token sent with every request
5. Refresh Token used to get new Access Token
```

### Password Security

- **Algorithm**: Argon2id (OWASP recommended)
- **Parameters**: 
  - Memory: 64 MB
  - Iterations: 3
  - Parallelism: 4
  - Salt: 16 bytes random
  - Hash: 32 bytes

### Two-Factor Authentication (2FA)

GoConnect supports TOTP-based 2FA:

1. Enable 2FA in settings
2. Scan QR code with authenticator app (Google Authenticator, Authy, etc.)
3. Enter 6-digit code to verify
4. Recovery codes provided for backup

---

## 🛡️ Authorization (RBAC)

### Roles

| Role | Permissions |
|------|-------------|
| **Owner** | Full network control, delete network, transfer ownership |
| **Admin** | Manage members, kick/ban, create channels |
| **Moderator** | Kick users, manage channels |
| **Member** | Connect, chat, file transfer |

### Permission Matrix

| Action | Owner | Admin | Mod | Member |
|--------|-------|-------|-----|--------|
| Connect to Network | ✅ | ✅ | ✅ | ✅ |
| Send Messages | ✅ | ✅ | ✅ | ✅ |
| Transfer Files | ✅ | ✅ | ✅ | ✅ |
| Kick Users | ✅ | ✅ | ✅ | ❌ |
| Ban Users | ✅ | ✅ | ❌ | ❌ |
| Invite Users | ✅ | ✅ | ✅ | ❌ |
| Manage Channels | ✅ | ✅ | ❌ | ❌ |
| Delete Network | ✅ | ❌ | ❌ | ❌ |

---

## 🌐 Network Security

### WireGuard Integration

GoConnect uses WireGuard for secure tunneling:

- **Protocol**: UDP-based, stateless
- **Encryption**: ChaCha20-Poly1305
- **Key Exchange**: Curve25519 (X25519)
- **Perfect Forward Secrecy**: Yes

### Key Management

```
┌─────────────────────────────────────────────────┐
│              KEY GENERATION FLOW                │
├─────────────────────────────────────────────────┤
│  1. Private key generated locally (never sent)  │
│  2. Public key derived from private key         │
│  3. Public key registered with server           │
│  4. Server distributes peer public keys         │
│  5. Direct P2P encryption established           │
└─────────────────────────────────────────────────┘
```

**Important**: Private keys NEVER leave your device.

### NAT Traversal

```
1. STUN Discovery (Public IP detection)
2. UDP Hole Punching (Direct P2P)
3. TURN Relay (Fallback if P2P fails)
```

---

## 🏠 Self-Hosted Security

### Production Checklist

- [ ] **HTTPS Only**: Use TLS 1.3 with valid certificates
- [ ] **Firewall**: Only expose required ports
- [ ] **Database**: Use strong passwords, enable encryption at rest
- [ ] **JWT Secret**: Generate cryptographically random secret (32+ bytes)
- [ ] **Rate Limiting**: Enable to prevent brute-force attacks
- [ ] **Updates**: Keep GoConnect and dependencies updated

### Recommended Firewall Rules

```bash
# HTTP API (behind reverse proxy)
# Internal only, reverse proxy handles external traffic

# WireGuard UDP
sudo ufw allow 51820/udp

# Block everything else
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw enable
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name api.goconnect.example.com;

    ssl_certificate /etc/letsencrypt/live/api.goconnect.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.goconnect.example.com/privkey.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Environment Variables

```bash
# Generate secure JWT secret
JWT_SECRET=$(openssl rand -base64 48)

# Use strong database password
DATABASE_URL="postgres://goconnect:$(openssl rand -base64 24)@localhost/goconnect"

# Disable debug mode
LOG_LEVEL=info
DEBUG=false
```

---

## 🔍 Audit Logging

GoConnect logs security-relevant events:

| Event | Logged Data |
|-------|-------------|
| Login Attempt | User, IP, Success/Failure, Timestamp |
| Token Refresh | User, IP, Timestamp |
| Network Create | User, Network ID, Timestamp |
| Member Join/Leave | User, Network, Action, Timestamp |
| Permission Change | Actor, Target, Old/New Role, Timestamp |
| 2FA Enable/Disable | User, Action, Timestamp |

### Log Location

- **Server**: `stdout` (Docker) or `/var/log/goconnect/audit.log`
- **CLI**: `~/.config/goconnect/logs/`

---

## 🐛 Security Vulnerability Reporting

If you discover a security vulnerability:

1. **DO NOT** open a public GitHub issue
2. Email: security@goconnect.io (if available)
3. Or use GitHub Security Advisory (private)

We will respond within 48 hours and work with you to fix the issue.

---

## 📚 References

- [WireGuard Protocol](https://www.wireguard.com/protocol/)
- [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
