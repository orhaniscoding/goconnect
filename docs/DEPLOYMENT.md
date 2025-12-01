# 🚀 GoConnect Deployment Guide

Bu doküman GoConnect Server'ı production ortamında deploy etmek için gereken adımları açıklar.

---

## 📋 Prerequisites

- Docker 24+ veya Go 1.24+
- PostgreSQL 15+ (production) veya SQLite (development)
- WireGuard tools (`wg`, `wg-quick`)
- Reverse proxy (Nginx/Caddy) - önerilen
- Valid SSL certificate (Let's Encrypt)

---

## 🐳 Docker Deployment (Önerilen)

### Quick Start

```bash
# Docker Compose ile hızlı kurulum
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.prod.yml
docker compose -f docker-compose.prod.yml up -d
```

### Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect-server
    restart: unless-stopped
    ports:
      - "8080:8080"           # HTTP API (reverse proxy arkasında)
      - "51820:51820/udp"     # WireGuard
    volumes:
      - goconnect-data:/data
      - /etc/wireguard:/etc/wireguard
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=postgres://goconnect:${DB_PASSWORD}@postgres:5432/goconnect?sslmode=disable
      - WG_SERVER_ENDPOINT=${WG_ENDPOINT}:51820
      - LOG_LEVEL=info
    cap_add:
      - NET_ADMIN
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
    depends_on:
      - postgres
    networks:
      - goconnect-net

  postgres:
    image: postgres:16-alpine
    container_name: goconnect-db
    restart: unless-stopped
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=goconnect
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=goconnect
    networks:
      - goconnect-net

volumes:
  goconnect-data:
  postgres-data:

networks:
  goconnect-net:
    driver: bridge
```

### Environment File

```bash
# .env
JWT_SECRET=your-super-secret-key-at-least-32-characters-long
DB_PASSWORD=your-database-password
WG_ENDPOINT=vpn.example.com
```

---

## 🖥️ Binary Deployment

### Download & Install

```bash
# Download latest release
VERSION=$(curl -s https://api.github.com/repos/orhaniscoding/goconnect/releases/latest | grep tag_name | cut -d'"' -f4)
curl -LO "https://github.com/orhaniscoding/goconnect/releases/download/${VERSION}/goconnect-server_linux_amd64.tar.gz"
tar -xzf goconnect-server_linux_amd64.tar.gz

# Move to system path
sudo mv goconnect-server /usr/local/bin/

# Create config directory
sudo mkdir -p /etc/goconnect
sudo mkdir -p /var/lib/goconnect
```

### Systemd Service

```ini
# /etc/systemd/system/goconnect-server.service
[Unit]
Description=GoConnect Server
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
Type=simple
User=goconnect
Group=goconnect
EnvironmentFile=/etc/goconnect/server.env
ExecStart=/usr/local/bin/goconnect-server
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/goconnect /etc/wireguard

[Install]
WantedBy=multi-user.target
```

### Enable & Start

```bash
# Create user
sudo useradd -r -s /bin/false goconnect

# Set permissions
sudo chown -R goconnect:goconnect /var/lib/goconnect

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable goconnect-server
sudo systemctl start goconnect-server

# Check status
sudo systemctl status goconnect-server
sudo journalctl -u goconnect-server -f
```

---

## 🌐 Reverse Proxy Configuration

### Nginx

```nginx
# /etc/nginx/sites-available/goconnect
upstream goconnect_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name api.goconnect.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.goconnect.example.com;

    # SSL (Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/api.goconnect.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.goconnect.example.com/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # API Routes
    location / {
        proxy_pass http://goconnect_backend;
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

    # Rate limiting
    limit_req zone=goconnect_limit burst=20 nodelay;
}

# Rate limit zone (in http block)
# limit_req_zone $binary_remote_addr zone=goconnect_limit:10m rate=10r/s;
```

### Caddy (Simpler Alternative)

```
# /etc/caddy/Caddyfile
api.goconnect.example.com {
    reverse_proxy localhost:8080
    
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
    }
}
```

---

## 🗄️ Database Setup

### PostgreSQL

```bash
# Create database
sudo -u postgres psql

CREATE USER goconnect WITH PASSWORD 'your-secure-password';
CREATE DATABASE goconnect OWNER goconnect;
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q

# Run migrations
goconnect-server -migrate
```

### SQLite (Development Only)

```bash
# SQLite is created automatically
DATABASE_URL=sqlite:///var/lib/goconnect/goconnect.db
```

---

## ☁️ Cloud Deployment

### AWS (ECS/Fargate)

```yaml
# task-definition.json
{
  "family": "goconnect-server",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "goconnect-server",
      "image": "ghcr.io/orhaniscoding/goconnect-server:latest",
      "portMappings": [
        {"containerPort": 8080, "protocol": "tcp"},
        {"containerPort": 51820, "protocol": "udp"}
      ],
      "environment": [
        {"name": "DATABASE_URL", "value": "..."},
        {"name": "JWT_SECRET", "value": "..."}
      ],
      "linuxParameters": {
        "capabilities": {
          "add": ["NET_ADMIN"]
        }
      }
    }
  ]
}
```

### DigitalOcean (Recommended for Simple Setup)

```bash
# Create droplet with Docker
doctl compute droplet create goconnect-server \
  --image docker-20-04 \
  --size s-2vcpu-4gb \
  --region nyc1 \
  --ssh-keys $SSH_KEY_ID

# SSH and run Docker Compose
ssh root@<droplet-ip>
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.prod.yml
docker compose up -d
```

---

## 📊 Monitoring

### Health Check

```bash
# HTTP health endpoint
curl https://api.goconnect.example.com/health

# Expected response
{"status": "healthy", "version": "3.0.0"}
```

### Prometheus Metrics

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'goconnect'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Log Aggregation

```bash
# Docker logs
docker logs -f goconnect-server

# Systemd logs
journalctl -u goconnect-server -f

# Send to external service (e.g., Loki)
docker run -d \
  -v /var/log:/var/log:ro \
  grafana/promtail:latest
```

---

## 🔄 Updates

### Docker

```bash
# Pull latest
docker compose pull

# Restart with new image
docker compose up -d

# Cleanup old images
docker image prune -f
```

### Binary

```bash
# Stop service
sudo systemctl stop goconnect-server

# Download new version
VERSION=v3.0.1
curl -LO "https://github.com/orhaniscoding/goconnect/releases/download/${VERSION}/goconnect-server_linux_amd64.tar.gz"
tar -xzf goconnect-server_linux_amd64.tar.gz
sudo mv goconnect-server /usr/local/bin/

# Run migrations (if needed)
goconnect-server -migrate

# Start service
sudo systemctl start goconnect-server
```

---

## 🛟 Troubleshooting

### Common Issues

| Problem | Solution |
|---------|----------|
| "NET_ADMIN capability required" | Add `cap_add: NET_ADMIN` to Docker or run with sudo |
| "WireGuard interface creation failed" | Ensure `wireguard` kernel module is loaded |
| "Database connection failed" | Check `DATABASE_URL` and PostgreSQL status |
| "JWT validation failed" | Ensure `JWT_SECRET` is consistent across restarts |

### Debug Mode

```bash
# Enable debug logging
LOG_LEVEL=debug goconnect-server
```

---

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
