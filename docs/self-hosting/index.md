# Self-Hosting Guide

**Complete guide to running your own GoConnect server.**

---

## 📋 Table of Contents

1. [Why Self-Host?](#why-self-host)
2. [Quick Start (Docker)](#quick-start-docker)
3. [Manual Installation](#manual-installation)
4. [Configuration](#configuration)
5. [Reverse Proxy](#reverse-proxy)
6. [Security Checklist](#security-checklist)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

---

## Why Self-Host?

### Benefits

**Privacy 🔒**
- Your data stays on your server
- No third-party tracking
- Complete control over logs

**Control 🎛️**
- Customize settings
- Manage users
- Set your own policies

**Performance 🚀**
- No intermediate servers
- Low latency
- Dedicated resources

**Cost 💰**
- Free (with your own server)
- No subscription fees
- Pay only for hosting

---

## Quick Start (Docker)

**Prerequisites:**
- Docker 20.10+
- Docker Compose 2.0+
- 1GB RAM minimum
- 10GB disk space

### Step 1: Create Directory

```bash
mkdir goconnect-server
cd goconnect-server
```

### Step 2: Download Docker Compose File

```bash
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml
```

Or create `docker-compose.yml` manually:

```yaml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "51820:51820/udp"
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DB_PASSWORD=${DB_PASSWORD}
      - WG_SERVER_ENDPOINT=${WG_SERVER_ENDPOINT}
    volumes:
      - goconnect-data:/app/data
      - ./config:/app/config
    networks:
      - goconnect-net

  postgres:
    image: postgres:15-alpine
    container_name: goconnect-db
    restart: unless-stopped
    environment:
      - POSTGRES_DB=goconnect
      - POSTGRES_USER=goconnect
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - goconnect-net

volumes:
  goconnect-data:
  postgres-data:

networks:
  goconnect-net:
    driver: bridge
```

### Step 3: Create Environment File

```bash
cat > .env << EOF
# Generate random secrets
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 16)

# Replace with your actual domain or public IP
WG_SERVER_ENDPOINT=your-domain.com:51820
EOF
```

**IMPORTANT:** Replace `your-domain.com` with your actual domain or public IP address.

### Step 4: Start Server

```bash
docker compose up -d
```

### Step 5: Verify

```bash
# Check logs
docker compose logs -f

# Check status
docker compose ps
```

Your server is now running on `http://localhost:8080`

---

## Manual Installation

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- WireGuard tools
- 1GB RAM minimum
- 10GB disk space

### Step 1: Install Dependencies

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install -y postgresql wireguard-tools
```

**macOS:**

```bash
brew install postgresql wireguard-tools
```

### Step 2: Download GoConnect

```bash
# Download latest release
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server-linux-amd64.tar.gz

# Extract
tar -xzf goconnect-server-linux-amd64.tar.gz

# Move to /usr/local/bin
sudo mv goconnect-server /usr/local/bin/
```

### Step 3: Setup Database

```bash
# Create database
sudo -u postgres psql

CREATE DATABASE goconnect;
CREATE USER goconnect WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
```

### Step 4: Configure

Create `/etc/goconnect/config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5432
  name: "goconnect"
  user: "goconnect"
  password: "your_password"

jwt:
  secret: "your-jwt-secret-min-32-chars"
  expiration: "24h"

wireguard:
  endpoint: "your-domain.com:51820"
  interface: "wg0"
```

### Step 5: Run Server

```bash
# Run directly
goconnect-server --config /etc/goconnect/config.yaml

# Or create a systemd service (recommended)
sudo nano /etc/systemd/system/goconnect.service
```

**Systemd Service File:**

```ini
[Unit]
Description=GoConnect Server
After=network.target postgresql.service

[Service]
Type=simple
User=goconnect
ExecStart=/usr/local/bin/goconnect-server --config /etc/goconnect/config.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable goconnect
sudo systemctl start goconnect
```

---

## Configuration

### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `JWT_SECRET` | Yes | Secret for JWT tokens | Random 32+ chars |
| `DB_PASSWORD` | Yes | PostgreSQL password | Random 16+ chars |
| `WG_SERVER_ENDPOINT` | Yes | Public endpoint | `domain.com:51820` |
| `LOG_LEVEL` | No | Log verbosity | `info`, `debug`, `warn` |
| `RATE_LIMIT` | No | Requests per minute | `100` |

### Advanced Config

**Custom Network Range:**

```yaml
wireguard:
  cidr: "10.0.0.0/24"
```

**Custom Port:**

```yaml
server:
  port: 9000
```

**Enable TLS:**

```yaml
server:
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

---

## Reverse Proxy

### Nginx

**Install Nginx:**

```bash
sudo apt install nginx
```

**Config file `/etc/nginx/sites-available/goconnect`:**

```nginx
server {
    listen 80;
    server_name your-domain.com;

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

**Enable:**

```bash
sudo ln -s /etc/nginx/sites-available/goconnect /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Caddy (Automatic HTTPS)

**Caddyfile:**

```
your-domain.com {
    reverse_proxy localhost:8080
}
```

---

## Security Checklist

### ✅ Before Going Public

- [ ] Change all default passwords
- [ ] Generate strong JWT_SECRET (32+ random characters)
- [ ] Set up firewall (only expose ports 80, 443, 51820)
- [ ] Enable HTTPS/TLS
- [ ] Configure rate limiting
- [ ] Set up log monitoring
- [ ] Backup database regularly
- [ ] Keep software updated

### Firewall Rules

**UFW (Ubuntu):**

```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 51820/udp
sudo ufw enable
```

---

## Monitoring

### Logs

```bash
# Docker
docker compose logs -f goconnect

# Systemd
journalctl -u goconnect -f
```

### Health Check

```bash
curl http://localhost:8080/health
```

---

## Troubleshooting

### Server Won't Start

**Check logs:**
```bash
docker compose logs goconnect
```

### WireGuard Connection Failed

**Check endpoint:**
```bash
curl ifconfig.me
```

---

**Last Updated:** 2025-01-24
**Version:** 1.0.0
