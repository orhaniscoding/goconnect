# 🏠 Self-Hosted GoConnect Kurulumu

Kendi GoConnect sunucunuzu kurmak, yönetmek ve güvence altına almak için tam kılavuz.

---

## 📑 İçindekiler

- [Neden Self-Hosted?](#neden-self-hosted)
- [Mimari Genel Bakış](#mimari-genel-bakış)
- [Hızlı Başlangıç](#hızlı-başlangıç)
- [Detaylı Kurulum](#detaylı-kurulum)
- [Konfigürasyon](#konfigürasyon)
- [Reverse Proxy](#reverse-proxy)
- [SSL/TLS Kurulumu](#ssltls-kurulumu)
- [Monitoring](#monitoring)
- [Yedekleme](#yedekleme)
- [Güvenlik](#güvenlik)

---

## 🤔 Neden Self-Hosted?

| Avantaj | Açıklama |
|---------|----------|
| 🔒 **Gizlilik** | Tüm veriler sizin sunucunuzda |
| 🎛️ **Kontrol** | Tam konfigürasyon kontrolü |
| 🚀 **Performans** | Kendi altyapınız |
| 💰 **Maliyet** | Uzun vadede ucuz |
| 🌍 **Konum** | Kendi bölgenizde |
| 🔄 **Özelleştirme** | Kendi kurallarınız |

### Kime Uygun?

- **Organizasyonlar**: Şirket içi kullanım
- **Gizlilik öncelikli kullanıcılar**: Veriler kendi sunucuda
- **Geliştiriciler**: Test ve development ortamı
- ** servis sağlayıcıları**: Müşterilerine hizmet

---

## 🏗️ Mimari Genel Bakış

```
┌─────────────────────────────────────────────────────────────┐
│                    Self-Hosted GoConnect                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │              Reverse Proxy (Nginx/Caddy)              │    │
│  │              SSL/TLS Termination                      │    │
│  └────────────────┬─────────────────────────────────────┘    │
│                   │                                            │
│  ┌────────────────▼─────────────────────────────────────┐    │
│  │              GoConnect Server (Core)                  │    │
│  │              HTTP API (port 8080)                     │    │
│  │              WebSocket (real-time)                    │    │
│  └────────────────┬─────────────────────────────────────┘    │
│                   │                                            │
│  ┌────────────────▼─────────────────────────────────────┐    │
│  │              Database (PostgreSQL/SQLite)             │    │
│  │              User Data, Networks, Chats              │    │
│  └────────────────┬─────────────────────────────────────┘    │
│                   │                                            │
│  ┌────────────────▼─────────────────────────────────────┐    │
│  │              WireGuard (port 51820/UDP)               │    │
│  │              Peer-to-Peer VPN                         │    │
│  └────────────────┬─────────────────────────────────────┘    │
│                   │                                            │
│              ┌─────┴─────┐                                    │
│              ▼           ▼                                    │
│         ┌─────────┐ ┌─────────┐                              │
│         │ Client 1│ │ Client 2│  ...                         │
│         └─────────┘ └─────────┘                              │
└─────────────────────────────────────────────────────────────┘
```

---

## ⚡ Hızlı Başlangıç

### Minimum Gereksinimler

| Kaynak | Minimum | Önerilen |
|--------|---------|----------|
| **CPU** | 2 core | 4+ core |
| **RAM** | 2 GB | 4+ GB |
| **Disk** | 20 GB | 50+ GB SSD |
| **Ağ** | 100 Mbps | 1 Gbps |

### 5 Dakikada Kurulum

```bash
# 1. Docker kurun (yoksa)
curl -fsSL https://get.docker.com | sh

# 2. Repo'yu clone edin
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect

# 3. docker-compose.yml indirin
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# 4. .env oluşturun
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DATABASE_URL=postgres://goconnect:$(openssl rand -base64 16)@db:5432/goconnect?sslmode=disable
WG_SERVER_ENDPOINT=$(curl -s ifconfig.me):51820
EOF

# 5. Başlatın
docker compose up -d

# 6. Logları kontrol edin
docker compose logs -f

# 7. Tarayıcıda açın
open http://localhost:8080
```

**İlk ayar:**
1. Admin hesabı oluşturun
2. Sunucu ayarlarını yapılandırın
3. WireGuard public key'i alın
4. İlk kullanıcıları davet edin

---

## 🔧 Detaylı Kurulum

### Docker ile Kurulum (Önerilen)

#### docker-compose.yml

```yaml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect
    restart: unless-stopped
    ports:
      - "8080:8080"      # HTTP API
      - "51820:51820/udp" # WireGuard
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - WG_SERVER_ENDPOINT=${WG_SERVER_ENDPOINT}
      - HTTP_PORT=8080
      - LOG_LEVEL=info
    volumes:
      - goconnect-data:/data
      - /dev/net/tun:/dev/net/tun
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    depends_on:
      - db
    networks:
      - goconnect-net

  db:
    image: postgres:15-alpine
    container_name: goconnect-db
    restart: unless-stopped
    environment:
      - POSTGRES_USER=goconnect
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=goconnect
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - goconnect-net

  # Opsiyonel: Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: goconnect-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - goconnect-net

  grafana:
    image: grafana/grafana:latest
    container_name: goconnect-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - goconnect-net

volumes:
  goconnect-data:
  postgres-data:
  prometheus-data:
  grafana-data:

networks:
  goconnect-net:
    driver: bridge
```

#### .env Dosyası

```bash
# Database
DB_PASSWORD=$(openssl rand -base64 24)
DATABASE_URL=postgres://goconnect:${DB_PASSWORD}@db:5432/goconnect?sslmode=disable

# Authentication
JWT_SECRET=$(openssl rand -base64 48)

# WireGuard
WG_SERVER_ENDPOINT=your-domain.com:51820
WG_SERVER_PUBKEY=<auto-generated on first run>
WG_SERVER_PRIVKEY=<auto-generated on first run>
WG_SUBNET=10.0.0.0/8

# Server
HTTP_PORT=8080
LOG_LEVEL=info

# Monitoring (opsiyonel)
GRAFANA_PASSWORD=$(openssl rand -base64 16)
```

---

### Manual Binary Installation

#### Linux (Debian/Ubuntu)

```bash
# 1. İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_linux_amd64.tar.gz
tar -xzf goconnect-server_linux_amd64.tar.gz
cd goconnect-server-linux-amd64

# 2. Kullanıcı ve dizin oluşturun
sudo useradd -r -s /bin/false goconnect
sudo mkdir -p /etc/goconnect /var/lib/goconnect /var/log/goconnect

# 3. Binary'yi kopyalayın
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# 4. Konfigürasyon
sudo cp config.example.env /etc/goconnect/.env
sudo chown -R goconnect:goconnect /etc/goconnect /var/lib/goconnect /var/log/goconnect
sudo nano /etc/goconnect/.env  # Edit config

# 5. Systemd service
sudo cat > /etc/systemd/system/goconnect.service << 'EOF'
[Unit]
Description=GoConnect Server
After=network.target postgresql.service
Wants=network-online.target

[Service]
Type=simple
User=goconnect
Group=goconnect
ExecStart=/usr/local/bin/goconnect-server -config /etc/goconnect/.env
Restart=on-failure
RestartSec=5s
AmbientCapabilities=CAP_NET_ADMIN

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/goconnect /var/log/goconnect

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goconnect

[Install]
WantedBy=multi-user.target
EOF

# 6. Başlatın
sudo systemctl daemon-reload
sudo systemctl enable goconnect
sudo systemctl start goconnect

# 7. Durumu kontrol edin
sudo systemctl status goconnect
sudo journalctl -u goconnect -f
```

#### macOS (Service)

```bash
# 1. İndirin ve kurun
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_darwin_arm64.tar.gz
tar -xzf goconnect-server_darwin_arm64.tar.gz
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# 2. Konfigürasyon
sudo mkdir -p /etc/goconnect
sudo cp config.example.env /etc/goconnect/.env
sudo nano /etc/goconnect/.env

# 3. LaunchAgent
~/Library/LaunchAgents/com.goconnect.server.plist

<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.goconnect.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/goconnect-server</string>
        <string>-config</string>
        <string>/etc/goconnect/.env</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/goconnect.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/goconnect.stderr.log</string>
</dict>
</plist>

# 4. Yükle
launchctl load ~/Library/LaunchAgents/com.goconnect.server.plist
```

#### Windows (Service)

```powershell
# 1. İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_windows_amd64.zip" -OutFile "server.zip"
Expand-Archive -Path "server.zip" -DestinationPath "C:\GoConnect"

# 2. Konfigürasyon
Copy-Item "C:\GoConnect\config.example.env" "C:\GoConnect\.env"
notepad "C:\GoConnect\.env"

# 3. NSSM ile service kurun
# İndir: https://nssm.cc/download
nssm install GoConnect "C:\GoConnect\goconnect-server.exe" "-config" "C:\GoConnect\.env"
nssm set GoConnect AppDirectory C:\GoConnect
nssm set GoConnect DisplayName GoConnect Server
nssm set GoConnect Description GoConnect P2P VPN Server
nssm set GoConnect Start SERVICE_AUTO_START
nssm start GoConnect
```

---

## ⚙️ Konfigürasyon

### Environment Variables

| Variable | Gerekli? | Varsayılan | Açıklama |
|----------|----------|------------|----------|
| `JWT_SECRET` | ✅ | - | JWT imzalama anahtarı (32+ karakter) |
| `DATABASE_URL` | ✅ | - | PostgreSQL/SQLite bağlantı dizesi |
| `WG_SERVER_ENDPOINT` | ✅ | - | Public endpoint: domain.com:51820 |
| `HTTP_PORT` | ❌ | 8080 | HTTP API portu |
| `WG_PORT` | ❌ | 51820 | WireGuard UDP portu |
| `WG_SUBNET` | ❌ | 10.0.0.0/8 | VPN subnet |
| `LOG_LEVEL` | ❌ | info | debug, info, warn, error |
| `CORS_ORIGINS` | ❌ | * | CORS izin verilen origins |

### PostgreSQL Konfigürasyonu

**postgresql.conf:**
```ini
# Performance
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100

# Connections
max_connections = 100

# Logging
log_min_duration_statement = 1000
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_connections = on
log_disconnections = on
log_duration = off
```

---

## 🌐 Reverse Proxy

### Nginx

```nginx
# /etc/nginx/sites-available/goconnect

server {
    listen 80;
    server_name goconnect.example.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name goconnect.example.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/goconnect.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/goconnect.example.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/goconnect-access.log;
    error_log /var/log/nginx/goconnect-error.log;

    # Proxy API
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # WebSocket
    location /v1/ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:8080;
        access_log off;
    }
}
```

### Caddy (Otomatik HTTPS)

```
# Caddyfile

goconnect.example.com {
    reverse_proxy localhost:8080

    # WebSocket support
    @websockets {
        header Connection *Upgrade*
        header Upgrade websocket
    }
    reverse_proxy @websockets localhost:8080

    # Logging
    log {
        output file /var/log/caddy/goconnect-access.log
        format json
    }

    # Security headers
    header {
        Strict-Transport-Security max-age=31536000;
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
    }
}
```

---

## 🔐 SSL/TLS Kurulumu

### Let's Encrypt (Certbot)

```bash
# Certbot kurun
sudo apt install certbot python3-certbot-nginx

# Sertifika alın
sudo certbot --nginx -d goconnect.example.com

# Otomatik yenileme (zaten ayarlı)
sudo certbot renew --dry-run
```

### Manuel SSL

```bash
# Self-signed (development için)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/goconnect.key \
  -out /etc/ssl/certs/goconnect.crt

# İzinleri ayarlayın
sudo chmod 600 /etc/ssl/private/goconnect.key
sudo chmod 644 /etc/ssl/certs/goconnect.crt
```

---

## 📊 Monitoring

### Prometheus Metrics

GoConnect `/metrics` endpoint'inde Prometheus metrics sunar.

**prometheus.yml:**
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'goconnect'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
```

### Grafana Dashboard

**Önerilen paneller:**
- Active connections
- Network traffic (bytes/sec)
- API request rate
- Database pool stats
- Error rate
- Latency (p50, p95, p99)

---

## 💾 Yedekleme

### Database Backup

```bash
# PostgreSQL
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/goconnect"
DB_CONTAINER="goconnect-db"

mkdir -p $BACKUP_DIR

# Backup al
docker exec $DB_CONTAINER pg_dump -U goconnect goconnect | gzip > $BACKUP_DIR/goconnect_$DATE.sql.gz

# 7 günden eski backup'ları sil
find $BACKUP_DIR -name "goconnect_*.sql.gz" -mtime +7 -delete

echo "Backup completed: goconnect_$DATE.sql.gz"
```

### Automated Backup (Cron)

```bash
# /etc/cron.d/goconnect-backup

# Her gün saat 03:00'da backup al
0 3 * * * root /usr/local/bin/backup-goconnect.sh
```

### Restore

```bash
# PostgreSQL'den restore
gunzip < goconnect_20250124_030000.sql.gz | docker exec -i goconnect-db psql -U goconnect goconnect
```

---

## 🔒 Güvenlik

### Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 51820/udp # WireGuard

sudo ufw enable
```

### Rate Limiting (Nginx)

```nginx
# Rate limiting zone
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/s;

# Apply to endpoints
location /v1/auth/ {
    limit_req zone=auth_limit burst=10 nodelay;
    proxy_pass http://localhost:8080;
}

location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
    proxy_pass http://localhost:8080;
}
```

### Fail2Ban

```ini
# /etc/fail2ban/jail.local

[goconnect-auth]
enabled = true
port = http,https
filter = goconnect-auth
logpath = /var/log/nginx/goconnect-error.log
maxretry = 5
bantime = 3600
findtime = 600
```

---

## 📚 Ek Kaynaklar

- [Reverse Proxy Guide](docs/self-hosted/REVERSE_PROXY.md)
- [SSL Setup Guide](docs/self-hosted/SSL_SETUP.md)
- [Monitoring Guide](docs/self-hosted/MONITORING.md)
- [Troubleshooting](../TROUBLESHOOTING.md)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
