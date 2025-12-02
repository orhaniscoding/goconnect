# 🏠 Self-Hosting Guide

Complete guide for running your own GoConnect server. Choose Docker (recommended) or manual installation.

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Prerequisites](#-prerequisites)
- [Docker Installation](#-docker-installation-recommended)
- [Manual Installation](#-manual-installation)
- [Configuration](#-configuration)
- [Reverse Proxy Setup](#-reverse-proxy-setup)
- [Troubleshooting](#-troubleshooting)
- [Security Checklist](#-security-checklist)

---

## ⚡ Quick Start

### Option 1: Docker (Fastest - 5 minutes)

```bash
# Download docker-compose file
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# Create environment file
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 16)
WG_SERVER_ENDPOINT=your-domain.com:51820
EOF

# Start everything
docker compose up -d

# Check status
docker compose ps
curl http://localhost:8081/health
```

### Option 2: Manual Binary (10 minutes)

```bash
# Download binary
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_linux_amd64.tar.gz
tar -xzf goconnect-server_*.tar.gz

# Create config
./goconnect-server --version  # Verify it works

# Run with defaults (SQLite)
./goconnect-server
```

---

## 📦 Prerequisites

### Minimum Requirements

| Component | Requirement |
|-----------|-------------|
| **OS** | Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+) |
| **CPU** | 1 core (2+ recommended) |
| **RAM** | 512 MB (1 GB+ recommended) |
| **Disk** | 1 GB free space |
| **Network** | Public IP or domain name |

### Required Software

- **Docker** (for Docker method): Docker 20.10+ and Docker Compose 2.0+
- **PostgreSQL** (for production): PostgreSQL 15+ (optional, SQLite works too)
- **WireGuard** tools: `wireguard-tools` package
- **Reverse Proxy** (recommended): Nginx or Caddy

### Install Prerequisites

**Ubuntu/Debian:**
```bash
# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# WireGuard tools
sudo apt update
sudo apt install -y wireguard-tools

# PostgreSQL (optional)
sudo apt install -y postgresql postgresql-contrib
```

**CentOS/RHEL:**
```bash
# Docker
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker

# WireGuard tools
sudo yum install -y epel-release
sudo yum install -y wireguard-tools

# PostgreSQL (optional)
sudo yum install -y postgresql-server postgresql-contrib
```

---

## 🐳 Docker Installation (Recommended)

### Why Docker?

✅ **Easy updates** - Just pull new image  
✅ **Isolated** - No conflicts with system packages  
✅ **Reproducible** - Same setup everywhere  
✅ **Includes dependencies** - PostgreSQL, Redis included  

### Step 1: Download Docker Compose

```bash
# Clone repository (or download files)
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# Or download just docker-compose.yml
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml
```

### Step 2: Configure Environment

Create a `.env` file:

```bash
cat > .env << 'EOF'
# Server Configuration
SERVER_PORT=8081
ENVIRONMENT=production

# Database (PostgreSQL)
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=change-this-strong-password
DB_SSLMODE=disable

# JWT Secret (REQUIRED - Generate a strong random key!)
JWT_SECRET=your-super-secret-key-minimum-32-characters-long-change-this

# WireGuard Configuration
WG_SERVER_ENDPOINT=vpn.yourdomain.com:51820
WG_INTERFACE_NAME=wg0
WG_DNS=1.1.1.1,1.0.0.1

# CORS (allow your frontend domains)
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,http://localhost:3000
EOF
```

**🔐 Generate secure secrets:**
```bash
# Generate JWT secret
openssl rand -base64 32

# Generate database password
openssl rand -base64 16
```

### Step 3: Start Services

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f server

# Check status
docker compose ps
```

### Step 4: Verify Installation

```bash
# Health check
curl http://localhost:8081/health

# Expected response:
# {"status":"healthy","version":"3.0.0"}
```

### Step 5: Configure Clients

Update your client configuration to point to your server:

**Desktop App:**
- Settings → Server URL → `https://vpn.yourdomain.com`

**CLI:**
```bash
goconnect setup
# Enter: https://vpn.yourdomain.com
```

### Docker Commands Reference

```bash
# Start services
docker compose up -d

# Stop services
docker compose stop

# View logs
docker compose logs -f server
docker compose logs -f postgres

# Restart a service
docker compose restart server

# Update to latest version
docker compose pull
docker compose up -d

# Remove everything (⚠️ deletes data!)
docker compose down -v
```

---

## 🖥️ Manual Installation

For those who prefer not to use Docker or need more control.

### Step 1: Download Binary

```bash
# Get latest version
VERSION=$(curl -s https://api.github.com/repos/orhaniscoding/goconnect/releases/latest | grep tag_name | cut -d'"' -f4)

# Download
curl -LO "https://github.com/orhaniscoding/goconnect/releases/download/${VERSION}/goconnect-server_linux_amd64.tar.gz"

# Extract
tar -xzf goconnect-server_*.tar.gz

# Move to system path
sudo mv goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# Verify
goconnect-server --version
```

### Step 2: Create System User

```bash
# Create dedicated user
sudo useradd -r -s /bin/false -d /var/lib/goconnect goconnect

# Create directories
sudo mkdir -p /etc/goconnect
sudo mkdir -p /var/lib/goconnect
sudo mkdir -p /var/log/goconnect

# Set ownership
sudo chown -R goconnect:goconnect /var/lib/goconnect /var/log/goconnect
```

### Step 3: Setup Database

**Option A: PostgreSQL (Production)**

```bash
# Install PostgreSQL (if not installed)
sudo apt install -y postgresql postgresql-contrib  # Ubuntu/Debian
# OR
sudo yum install -y postgresql-server postgresql-contrib  # CentOS/RHEL

# Create database and user
sudo -u postgres psql << EOF
CREATE USER goconnect WITH PASSWORD 'your-secure-password';
CREATE DATABASE goconnect OWNER goconnect;
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
EOF
```

**Option B: SQLite (Development/Simple)**

```bash
# SQLite is created automatically, just ensure directory exists
sudo mkdir -p /var/lib/goconnect
sudo chown goconnect:goconnect /var/lib/goconnect
```

### Step 4: Create Configuration File

```bash
sudo nano /etc/goconnect/goconnect.yaml
```

**PostgreSQL Configuration:**
```yaml
server:
  host: "0.0.0.0"
  port: "8081"
  environment: "production"
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  backend: "postgres"
  host: "localhost"
  port: "5432"
  user: "goconnect"
  password: "your-secure-password"
  dbname: "goconnect"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5

jwt:
  secret: "your-super-secret-key-minimum-32-characters-long"
  access_token_ttl: 15m
  refresh_token_ttl: 168h

wireguard:
  interface_name: "wg0"
  server_endpoint: "vpn.yourdomain.com:51820"
  server_pubkey: "YOUR_WIREGUARD_PUBLIC_KEY"
  private_key: "YOUR_WIREGUARD_PRIVATE_KEY"
  dns: "1.1.1.1,1.0.0.1"
  mtu: 1420
  keepalive: 25
  port: 51820

cors:
  allowed_origins:
    - "https://app.yourdomain.com"
    - "http://localhost:3000"
  allow_credentials: true
  max_age: 12h
```

**SQLite Configuration (Simpler):**
```yaml
server:
  host: "0.0.0.0"
  port: "8081"
  environment: "production"

database:
  backend: "sqlite"
  sqlite_path: "/var/lib/goconnect/goconnect.db"

jwt:
  secret: "your-super-secret-key-minimum-32-characters-long"
  access_token_ttl: 15m
  refresh_token_ttl: 168h

wireguard:
  interface_name: "wg0"
  server_endpoint: "vpn.yourdomain.com:51820"
  server_pubkey: "YOUR_WIREGUARD_PUBLIC_KEY"
  private_key: "YOUR_WIREGUARD_PRIVATE_KEY"
  dns: "1.1.1.1,1.0.0.1"
```

**🔐 Generate WireGuard Keys:**
```bash
# Generate private key
wg genkey | tee /tmp/wg_private.key

# Generate public key
cat /tmp/wg_private.key | wg pubkey | tee /tmp/wg_public.key

# Copy keys to config
# Private key → wireguard.private_key
# Public key → wireguard.server_pubkey
```

### Step 5: Create Systemd Service

```bash
sudo nano /etc/systemd/system/goconnect-server.service
```

```ini
[Unit]
Description=GoConnect Server
After=network-online.target postgresql.service
Wants=network-online.target
Documentation=https://github.com/orhaniscoding/goconnect

[Service]
Type=simple
User=goconnect
Group=goconnect
WorkingDirectory=/var/lib/goconnect

# Configuration file
ExecStart=/usr/local/bin/goconnect-server -config /etc/goconnect/goconnect.yaml

# Restart policy
Restart=on-failure
RestartSec=5

# Resource limits
LimitNOFILE=65535

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/goconnect /var/log/goconnect /etc/wireguard
PrivateTmp=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goconnect-server

[Install]
WantedBy=multi-user.target
```

### Step 6: Enable and Start

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable on boot
sudo systemctl enable goconnect-server

# Start service
sudo systemctl start goconnect-server

# Check status
sudo systemctl status goconnect-server

# View logs
sudo journalctl -u goconnect-server -f
```

### Step 7: Verify Installation

```bash
# Health check
curl http://localhost:8081/health

# Check service status
sudo systemctl status goconnect-server
```

---

## ⚙️ Configuration

### Environment Variables

All configuration can be set via environment variables (overrides config file):

```bash
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8081
export DB_BACKEND=postgres
export DB_HOST=localhost
export DB_USER=goconnect
export DB_PASSWORD=your-password
export DB_NAME=goconnect
export JWT_SECRET=your-secret-key
export WG_SERVER_ENDPOINT=vpn.yourdomain.com:51820
```

### Configuration File

Default location: `/etc/goconnect/goconnect.yaml` or `./goconnect.yaml`

**Full configuration reference:**

```yaml
server:
  host: "0.0.0.0"              # Listen address
  port: "8081"                 # HTTP port
  environment: "production"    # development | production
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  backend: "postgres"          # postgres | sqlite | memory
  host: "localhost"
  port: "5432"
  user: "goconnect"
  password: "password"
  dbname: "goconnect"
  sslmode: "disable"           # disable | require | verify-full
  sqlite_path: "/var/lib/goconnect/goconnect.db"  # For SQLite

jwt:
  secret: "..."                # Minimum 32 characters
  access_token_ttl: 15m
  refresh_token_ttl: 168h      # 7 days

wireguard:
  interface_name: "wg0"
  server_endpoint: "vpn.example.com:51820"
  server_pubkey: "..."         # 44 characters base64
  private_key: "..."           # Keep secret!
  dns: "1.1.1.1,1.0.0.1"
  mtu: 1420
  keepalive: 25
  port: 51820

cors:
  allowed_origins:
    - "https://app.yourdomain.com"
  allow_credentials: true
  max_age: 12h

redis:                          # Optional
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
```

---

## 🌐 Reverse Proxy Setup

### Why Use a Reverse Proxy?

✅ **SSL/TLS** - HTTPS encryption  
✅ **Domain name** - Use your own domain  
✅ **Security** - Hide internal ports  
✅ **Load balancing** - Multiple server instances  

### Option 1: Nginx (Most Common)

**Install Nginx:**
```bash
sudo apt install -y nginx  # Ubuntu/Debian
sudo yum install -y nginx  # CentOS/RHEL
```

**Create configuration:**
```bash
sudo nano /etc/nginx/sites-available/goconnect
```

```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name vpn.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name vpn.yourdomain.com;

    # SSL certificates (Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/vpn.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/vpn.yourdomain.com/privkey.pem;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy to GoConnect server
    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }

    # Health check endpoint (public)
    location /health {
        proxy_pass http://127.0.0.1:8081/health;
        access_log off;
    }
}
```

**Enable site:**
```bash
sudo ln -s /etc/nginx/sites-available/goconnect /etc/nginx/sites-enabled/
sudo nginx -t  # Test configuration
sudo systemctl reload nginx
```

**Get SSL certificate (Let's Encrypt):**
```bash
sudo apt install -y certbot python3-certbot-nginx  # Ubuntu/Debian
sudo certbot --nginx -d vpn.yourdomain.com
```

### Option 2: Caddy (Automatic HTTPS)

**Install Caddy:**
```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install -y caddy
```

**Create Caddyfile:**
```bash
sudo nano /etc/caddy/Caddyfile
```

```
vpn.yourdomain.com {
    reverse_proxy localhost:8081
    
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
    }
}
```

**Start Caddy:**
```bash
sudo systemctl enable caddy
sudo systemctl start caddy
```

---

## 🔧 Troubleshooting

### Server Won't Start

**Check logs:**
```bash
# Docker
docker compose logs server

# Systemd
sudo journalctl -u goconnect-server -n 50
```

**Common issues:**

| Error | Solution |
|-------|----------|
| `Failed to load configuration` | Check config file path and syntax |
| `Database connection failed` | Verify PostgreSQL is running and credentials |
| `JWT_SECRET is required` | Set JWT_SECRET environment variable |
| `Port already in use` | Change SERVER_PORT or stop conflicting service |
| `Permission denied` | Check file permissions and user ownership |

### Database Connection Issues

**PostgreSQL:**
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -h localhost -U goconnect -d goconnect

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

**SQLite:**
```bash
# Check file permissions
ls -la /var/lib/goconnect/goconnect.db

# Test SQLite
sqlite3 /var/lib/goconnect/goconnect.db "SELECT 1;"
```

### WireGuard Issues

**Check WireGuard module:**
```bash
# Load module
sudo modprobe wireguard

# Check if loaded
lsmod | grep wireguard

# Check interface
sudo wg show
```

**Firewall rules:**
```bash
# Allow WireGuard port (UDP 51820)
sudo ufw allow 51820/udp

# Allow HTTP port
sudo ufw allow 8081/tcp
```

### Network Connectivity

**Test server accessibility:**
```bash
# From server
curl http://localhost:8081/health

# From client
curl https://vpn.yourdomain.com/health
```

**Check firewall:**
```bash
# UFW (Ubuntu)
sudo ufw status

# Firewalld (CentOS)
sudo firewall-cmd --list-all
```

---

## 🔒 Security Checklist

### Essential Security Steps

- [ ] **Change default passwords** - JWT_SECRET, DB_PASSWORD
- [ ] **Use HTTPS** - Set up reverse proxy with SSL
- [ ] **Firewall** - Only expose necessary ports (80, 443, 51820)
- [ ] **Keep updated** - Regularly update server binary
- [ ] **Backup database** - Regular backups of PostgreSQL/SQLite
- [ ] **Monitor logs** - Check for suspicious activity
- [ ] **Use strong secrets** - Generate with `openssl rand -base64 32`
- [ ] **Limit CORS origins** - Only allow your frontend domains
- [ ] **Run as non-root** - Use dedicated user (goconnect)
- [ ] **Secure WireGuard keys** - Keep private keys secret

### Production Security Hardening

```bash
# 1. Disable root login (SSH)
sudo nano /etc/ssh/sshd_config
# Set: PermitRootLogin no

# 2. Enable firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 51820/udp
sudo ufw enable

# 3. Fail2ban (prevent brute force)
sudo apt install -y fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# 4. Automatic security updates
sudo apt install -y unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

---

## 📊 Monitoring

### Health Checks

```bash
# Basic health check
curl http://localhost:8081/health

# Prometheus metrics
curl http://localhost:8081/metrics
```

### Log Monitoring

```bash
# Docker logs
docker compose logs -f server

# Systemd logs
sudo journalctl -u goconnect-server -f

# Log rotation (systemd)
sudo nano /etc/systemd/journald.conf
# Set: SystemMaxUse=100M
sudo systemctl restart systemd-journald
```

### Backup

**PostgreSQL backup:**
```bash
# Manual backup
sudo -u postgres pg_dump goconnect > backup_$(date +%Y%m%d).sql

# Automated backup (cron)
0 2 * * * /usr/bin/pg_dump -U goconnect goconnect > /backups/goconnect_$(date +\%Y\%m\%d).sql
```

**SQLite backup:**
```bash
# Simple copy
cp /var/lib/goconnect/goconnect.db /backups/goconnect_$(date +%Y%m%d).db
```

---

## 🔄 Updates

### Docker Update

```bash
# Pull latest image
docker compose pull

# Restart with new image
docker compose up -d

# Cleanup old images
docker image prune -f
```

### Binary Update

```bash
# Stop service
sudo systemctl stop goconnect-server

# Backup current binary
sudo cp /usr/local/bin/goconnect-server /usr/local/bin/goconnect-server.backup

# Download new version
VERSION=v3.0.1
curl -LO "https://github.com/orhaniscoding/goconnect/releases/download/${VERSION}/goconnect-server_linux_amd64.tar.gz"
tar -xzf goconnect-server_*.tar.gz

# Replace binary
sudo mv goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# Start service
sudo systemctl start goconnect-server

# Verify
sudo systemctl status goconnect-server
```

---

## 📚 Additional Resources

- [Architecture Documentation](ARCHITECTURE.md)
- [User Guide](USER_GUIDE.md)
- [Security Best Practices](SECURITY.md)
- [GitHub Repository](https://github.com/orhaniscoding/goconnect)
- [Issue Tracker](https://github.com/orhaniscoding/goconnect/issues)

---

## 💬 Getting Help

- **GitHub Issues**: [Report bugs](https://github.com/orhaniscoding/goconnect/issues)
- **Discussions**: [Ask questions](https://github.com/orhaniscoding/goconnect/discussions)
- **Documentation**: Check other docs in `/docs` folder

---

**Made with ❤️ by the GoConnect team**

