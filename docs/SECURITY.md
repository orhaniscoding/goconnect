# GoConnect Security Guide

This document provides comprehensive security guidance for deploying and operating GoConnect in production environments.

---

## 🔐 Authentication & Authorization

### JWT Token Security

GoConnect uses JWT (JSON Web Tokens) for authentication. Configure securely:

```bash
# Generate a strong secret (REQUIRED in production)
export JWT_SECRET=$(openssl rand -base64 32)

# Token expiration settings
export JWT_ACCESS_TOKEN_EXPIRY=15m    # Short-lived access tokens
export JWT_REFRESH_TOKEN_EXPIRY=7d    # Longer refresh tokens
```

**Best Practices:**
- Never use default secrets in production
- Rotate secrets periodically
- Use at least 256-bit (32-byte) secrets
- Store secrets in secure vaults (HashiCorp Vault, AWS Secrets Manager)

### Password Security (Argon2id)

GoConnect uses Argon2id for password hashing with these recommended parameters:

| Parameter   | Development | Production |
| ----------- | ----------- | ---------- |
| Memory      | 64 MB       | 64 MB      |
| Iterations  | 1           | 3          |
| Parallelism | 2           | 4          |
| Salt Length | 16 bytes    | 16 bytes   |
| Key Length  | 32 bytes    | 32 bytes   |

```bash
# Configure via environment
export ARGON2_MEMORY=65536      # 64 MB in KB
export ARGON2_ITERATIONS=3
export ARGON2_PARALLELISM=4
```

### JWKS Rotation

For enterprise deployments, implement JWKS (JSON Web Key Set) rotation:

1. Generate new key pair
2. Add new public key to JWKS endpoint
3. Wait for token expiry window
4. Remove old key from JWKS
5. Update signing key

---

## 🛡️ Rate Limiting

Protect your server from abuse with rate limiting:

```bash
# Global rate limits
export RATE_LIMIT_REQUESTS_PER_MINUTE=100
export RATE_LIMIT_BURST=20

# Authentication endpoints (stricter)
export AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE=10
export AUTH_RATE_LIMIT_BURST=5

# API rate limits per user
export API_RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

**Recommended Limits:**

| Endpoint Type         | Requests/min | Burst |
| --------------------- | ------------ | ----- |
| Public APIs           | 100          | 20    |
| Auth (login/register) | 10           | 5     |
| Password reset        | 3            | 1     |
| API (authenticated)   | 60           | 10    |

---

## 🔒 Transport Security (TLS/HTTPS)

### Production TLS Setup

**Option 1: Let's Encrypt (Recommended)**
```bash
# Using certbot
sudo certbot certonly --standalone -d vpn.yourdomain.com

# Configure GoConnect
export TLS_CERT_FILE=/etc/letsencrypt/live/vpn.yourdomain.com/fullchain.pem
export TLS_KEY_FILE=/etc/letsencrypt/live/vpn.yourdomain.com/privkey.pem
export TLS_ENABLED=true
```

**Option 2: Reverse Proxy (nginx)**
```nginx
server {
    listen 443 ssl http2;
    server_name vpn.yourdomain.com;
    
    ssl_certificate /etc/ssl/certs/vpn.crt;
    ssl_certificate_key /etc/ssl/private/vpn.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 🗄️ Database Security

### PostgreSQL Hardening

```bash
# Use SSL connections
export DB_SSLMODE=require

# Connection string with SSL
export DATABASE_URL="postgres://goconnect:password@db.example.com:5432/goconnect?sslmode=require"
```

**Checklist:**
- [ ] Use strong, unique database passwords
- [ ] Enable SSL/TLS for database connections
- [ ] Restrict database network access (firewall/VPC)
- [ ] Use separate database users per service
- [ ] Enable query logging for audit
- [ ] Regular backups with encryption

### Redis Security

```bash
# Enable authentication
export REDIS_PASSWORD=your-strong-password

# Use TLS if available
export REDIS_TLS_ENABLED=true
```

---

## 📊 Audit Logging

GoConnect maintains comprehensive audit logs for security and compliance:

```bash
# Enable persistent audit logging
export AUDIT_SQLITE_DSN=file:/var/lib/goconnect/audit.db

# Hash sensitive data in logs
export AUDIT_HASH_SECRETS_B64=$(openssl rand -base64 32)

# Retention settings
export AUDIT_MAX_ROWS=100000           # Keep last 100k events
export AUDIT_MAX_AGE_SECONDS=7776000   # 90 days

# Sign audit exports for integrity
export AUDIT_SIGNING_KEY_ED25519_B64=$(openssl genpkey -algorithm ed25519 | openssl pkey -outform der | base64)
```

**Logged Events:**
- User authentication (login/logout/failed attempts)
- User management (create/update/delete)
- Device registration/removal
- Network changes
- Administrative actions
- API access patterns

---

## 🔑 WireGuard Key Security

### Key Generation Best Practices

- Generate keys on the device itself (never transmit private keys)
- Use secure random number generators
- Store private keys with restricted permissions (0600)
- Never log or expose private keys

```bash
# Generate WireGuard keys securely
wg genkey | tee privatekey | wg pubkey > publickey
chmod 600 privatekey
```

### Key Rotation

Implement periodic key rotation:

1. Generate new key pair on device
2. Register new public key with server
3. Update WireGuard configuration
4. Revoke old public key

---

## 🌐 Network Security

### Firewall Configuration

**Server (Linux with UFW):**
```bash
# Allow HTTPS
sudo ufw allow 443/tcp

# Allow WireGuard
sudo ufw allow 51820/udp

# Allow API (if exposed)
sudo ufw allow 8080/tcp

# Enable firewall
sudo ufw enable
```

**Docker Deployment:**
```yaml
# docker-compose.yml
services:
  goconnect:
    ports:
      - "127.0.0.1:8080:8080"  # Bind to localhost only
    networks:
      - internal
```

### IP Allowlisting

For sensitive endpoints:
```bash
export ADMIN_ALLOWED_IPS=10.0.0.0/8,192.168.0.0/16
```

---

## 🔏 GDPR & Data Privacy

### Data Subject Rights (DSR)

GoConnect supports GDPR compliance:

| Right                  | Implementation                  |
| ---------------------- | ------------------------------- |
| Right to Access        | Export user data via API        |
| Right to Erasure       | Delete user and associated data |
| Right to Portability   | JSON export of all user data    |
| Right to Rectification | Update user profile             |

**User Data Export:**
```bash
# Export user data
curl -X GET https://api.example.com/v1/users/{id}/export \
  -H "Authorization: Bearer $TOKEN"
```

**User Data Deletion:**
```bash
# Delete user and all associated data
curl -X DELETE https://api.example.com/v1/users/{id} \
  -H "Authorization: Bearer $TOKEN"
```

### Data Retention

Configure retention policies:
```bash
export USER_DATA_RETENTION_DAYS=365
export AUDIT_LOG_RETENTION_DAYS=90
export SESSION_DATA_RETENTION_DAYS=30
```

---

## 🚨 Incident Response

### Security Monitoring

1. **Log Aggregation**: Send logs to SIEM (Splunk, ELK, etc.)
2. **Alerting**: Configure alerts for suspicious activity
3. **Metrics**: Monitor authentication failures, rate limit hits

### Incident Checklist

- [ ] Identify scope of breach
- [ ] Rotate compromised credentials
- [ ] Revoke affected tokens/sessions
- [ ] Notify affected users
- [ ] Document incident timeline
- [ ] Implement remediation measures
- [ ] Post-incident review

---

## 🚫 User Suspension & Session Management

### Real-Time Suspension Enforcement

GoConnect implements **immediate suspension enforcement** to address security incidents in real-time.

**Security Guarantees:**
- ✅ Suspended users cannot login (even with correct password)
- ✅ Existing JWT tokens from suspended users are rejected immediately
- ✅ All active sessions blacklisted within seconds of suspension
- ✅ Zero grace period for suspension enforcement

**How It Works:**
1. **At Login:** Suspended users blocked before token generation
2. **At Token Validation:** Database check rejects suspended users' tokens
3. **Auto-Blacklist on Suspension:** All active sessions invalidated when admin suspends user
4. **Session Tracking:** Redis tracks all active JWT IDs for instant revocation

**Implementation:**
```bash
# Redis keys:
user_sessions:{userID}  → Set of active JTIs (JWT IDs)
blacklist:{JTI}        → "suspended" (TTL: 7 days)
```

**Admin Endpoints:**
```http
### Suspend User
POST /v1/admin/users/{user_id}/suspend
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "reason": "Security policy violation"
}

### Unsuspend User
POST /v1/admin/users/{user_id}/unsuspend
Authorization: Bearer {admin_token}
```

**Graceful Degradation:**
- If Redis unavailable: suspension still enforced via database queries
- Performance impact without Redis: ~2-5ms per request

**Best Practices:**
- Monitor Redis availability for real-time enforcement
- Review suspension audit logs regularly
- Document suspension reasons for compliance

---

## ✅ Security Checklist

### Pre-Production

- [ ] Strong JWT_SECRET configured
- [ ] TLS/HTTPS enabled
- [ ] Database SSL enabled
- [ ] Rate limiting configured
- [ ] Audit logging enabled
- [ ] Firewall rules applied
- [ ] Secrets stored securely
- [ ] Redis configured for session tracking

### Ongoing

- [ ] Regular security updates
- [ ] Credential rotation schedule
- [ ] Audit log review
- [ ] Penetration testing (annual)
- [ ] Backup verification
- [ ] Incident response drills
- [ ] Review suspended users weekly

---

## 📞 Security Contact

Report security vulnerabilities responsibly. We follow responsible disclosure practices.
