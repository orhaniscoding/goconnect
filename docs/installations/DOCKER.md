# 🐳 Docker ile Kurulum

Docker kullanarak GoConnect kurulumu ve deployment.

---

## 📋 İçindekiler

- [Hızlı Başlangıç](#hızlı-başlangıç)
- [Docker Compose](#docker-compose)
- [Konfigürasyon](#konfigürasyon)
- [Production Deployment](#production-deployment)
- [Sorun Giderme](#sorun-giderme)

---

## ⚡ Hızlı Başlangıç

### Prerequisites

Docker kurulu olmalı:
- [Docker Desktop](https://www.docker.com/products/docker-desktop) (Windows/macOS)
- [Docker Engine](https://docs.docker.com/engine/install/) (Linux)

### 3 Dakikada Kurulum

```bash
# docker-compose.yml indirin
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# .env oluşturun
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DATABASE_URL=postgres://goconnect:$(openssl rand -base64 16)@db:5432/goconnect?sslmode=disable
WG_SERVER_ENDPOINT=$(curl -s ifconfig.me):51820
EOF

# Başlatın
docker compose up -d

# Logları görün
docker compose logs -f

# Tarayıcıda açın
open http://localhost:8080
```

---

## 🐳 Docker Compose

### Minimal docker-compose.yml

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
    volumes:
      - goconnect-data:/data
    cap_add:
      - NET_ADMIN
```

---

### Production docker-compose.yml

```yaml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
      - "51820:51820/udp"
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

volumes:
  goconnect-data:
  postgres-data:

networks:
  goconnect-net:
    driver: bridge
```

---

## ⚙️ Konfigürasyon

### Environment Variables (.env)

```bash
# JWT Secret (minimum 32 karakter)
JWT_SECRET=your-super-secret-key-at-least-32-characters-change-this

# Database
DB_PASSWORD=change-this-password
DATABASE_URL=postgres://goconnect:${DB_PASSWORD}@db:5432/goconnect?sslmode=disable

# WireGuard
WG_SERVER_ENDPOINT=your-domain.com:51820
WG_SERVER_PUBKEY=
WG_SERVER_PRIVKEY=
WG_SUBNET=10.0.0.0/8

# Server
HTTP_PORT=8080
LOG_LEVEL=info
CORS_ORIGINS=*
```

---

### Volumes

**Persistent data:**
```yaml
volumes:
  - goconnect-data:/data       # Config, logs, state
  - postgres-data:/var/lib/postgresql/data  # Database
```

**Bind mount (development):**
```yaml
volumes:
  - ./config:/data
```

---

## 🚀 Production Deployment

### Reverse Proxy (Nginx)

```nginx
upstream goconnect {
    server localhost:8080;
}

server {
    listen 80;
    server_name goconnect.example.com;

    location / {
        proxy_pass http://goconnect;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

---

### Docker Daemon Configuration

**/etc/docker/daemon.json:**
```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "live-restore": true
}
```

---

### Resource Limits

```yaml
services:
  goconnect:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

---

## 🔧 Komutlar

### Temel Komutlar

```bash
# Başlat
docker compose up -d

# Durdur
docker compose down

# Restart
docker compose restart

# Logları görüntüle
docker compose logs -f

# Sadece goconnect logları
docker compose logs -f goconnect

# Son 100 satır
docker compose logs --tail 100
```

---

### Güncelleme

```bash
# Yeni image'i çekin
docker compose pull

# Recreate containers
docker compose up -d --force-recreate

# Eski image'leri temizleyin
docker image prune -a
```

---

### Backup

```bash
# Database backup
docker exec goconnect-db pg_dump -U goconnect goconnect | gzip > backup_$(date +%Y%m%d).sql.gz

# Volume backup
docker run --rm -v goconnect-data:/data -v $(pwd):/backup alpine tar czf /backup/goconnect-data-backup.tar.gz /data
```

---

### Restore

```bash
# Database restore
gunzip < backup_20250124.sql.gz | docker exec -i goconnect-db psql -U goconnect goconnect

# Volume restore
docker run --rm -v goconnect-data:/data -v $(pwd):/backup alpine tar xzf /backup/goconnect-data-backup.tar.gz -C /
```

---

## 🔧 Sorun Giderme

### Container başlamıyor

```bash
# Logları kontrol edin
docker compose logs goconnect

# Container durumunu kontrol edin
docker ps -a

# Shell'e girin
docker compose exec goconnect sh

# Environment variables'ı kontrol edin
docker compose exec goconnect env | sort
```

---

### Port conflict

```bash
# Port kullanan process'i bulun
sudo lsof -i :8080

# Docker'ı farklı portta çalıştırın
docker compose -f docker-compose.yml -p dev up -d
```

---

### Permission denied (/dev/net/tun)

```bash
# /dev/net/tun'u kontrol edin
ls -l /dev/net/tun

# İzinleri düzeltin
sudo chmod 666 /dev/net/tun

# Veya container'ı privileged çalıştırın (DEĞİL ÖNERİLEN)
docker compose run --privileged ...
```

---

### Out of memory

```bash
# Docker disk kullanımı
docker system df

# Temizlik
docker system prune -a --volumes

# Log rotation yapılandırın
# (bkz: Docker Daemon Configuration)
```

---

## 📚 Ek Kaynaklar

- [Self-Hosted Setup](../SELF_HOSTED_SETUP.md)
- [Reverse Proxy Guide](self-hosted/REVERSE_PROXY.md)
- [Docker Documentation](https://docs.docker.com/)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
