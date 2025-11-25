# GoConnect Installation Guide

Complete installation instructions for all platforms and components.

---

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Server Installation](#-server-installation)
  - [Docker (Recommended)](#docker-recommended)
  - [Linux (Debian/Ubuntu)](#linux-debianubuntu)
  - [Linux (RHEL/Fedora)](#linux-rhelfedora)
  - [macOS](#macos)
  - [Windows](#windows)
- [Client Daemon Installation](#-client-daemon-installation)
  - [Linux](#linux)
  - [macOS](#macos-1)
  - [Windows](#windows-1)
- [Post-Installation](#-post-installation)
- [Troubleshooting](#-troubleshooting)

---

## 🚀 Quick Start

**Server (Docker - 30 seconds):**
```bash
docker run -d --name goconnect -p 8080:8080 ghcr.io/orhaniscoding/goconnect-server:latest
```

**Client (Linux - 1 minute):**
```bash
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-daemon_linux_amd64.deb
sudo dpkg -i goconnect-daemon_linux_amd64.deb
sudo systemctl enable --now goconnect-daemon
```

---

## 🖥️ Server Installation

The GoConnect Server is the central management component that handles:
- User authentication and authorization
- Device registration and management
- WireGuard configuration distribution
- REST API and WebSocket connections

### Docker (Recommended)

Docker is the fastest and most reliable way to deploy GoConnect Server.

**Prerequisites:**
- Docker Engine 20.10+
- Docker Compose v2+ (optional)

**Quick Start:**
```bash
# Pull the latest image
docker pull ghcr.io/orhaniscoding/goconnect-server:latest

# Run with default settings
docker run -d \
  --name goconnect-server \
  --restart unless-stopped \
  -p 8080:8080 \
  -v goconnect-data:/data \
  -e JWT_SECRET=$(openssl rand -base64 32) \
  ghcr.io/orhaniscoding/goconnect-server:latest
```

**Production Setup with Docker Compose:**
```yaml
# docker-compose.yml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect-server
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=goconnect
      - DB_USER=goconnect
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - goconnect-data:/data
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_DB=goconnect
      - POSTGRES_USER=goconnect
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - redis-data:/data

volumes:
  goconnect-data:
  postgres-data:
  redis-data:
```

```bash
# Create .env file
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 24)
EOF

# Start services
docker compose up -d

# Check status
docker compose ps
docker compose logs -f goconnect
```

---

### Linux (Debian/Ubuntu)

**Prerequisites:**
- Ubuntu 20.04+ or Debian 11+
- systemd
- wireguard-tools

**Installation:**
```bash
# 1. Download the package
VERSION="2.8.5"  # Replace with latest version
wget https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-server_${VERSION}_linux_amd64.deb

# 2. Install dependencies
sudo apt update
sudo apt install -y wireguard-tools postgresql redis-server

# 3. Install GoConnect Server
sudo dpkg -i goconnect-server_${VERSION}_linux_amd64.deb

# 4. Configure environment
sudo mkdir -p /etc/goconnect
sudo cat > /etc/goconnect/server.env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your-secure-password
REDIS_HOST=localhost
REDIS_PORT=6379
EOF
sudo chmod 600 /etc/goconnect/server.env

# 5. Setup PostgreSQL
sudo -u postgres createuser goconnect
sudo -u postgres createdb -O goconnect goconnect

# 6. Start the service
sudo systemctl enable goconnect-server
sudo systemctl start goconnect-server

# 7. Check status
sudo systemctl status goconnect-server
journalctl -u goconnect-server -f
```

---

### Linux (RHEL/Fedora)

**Prerequisites:**
- RHEL 8+ / Fedora 35+
- systemd
- wireguard-tools

**Installation:**
```bash
# 1. Download the package
VERSION="2.8.5"  # Replace with latest version
wget https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-server_${VERSION}_linux_amd64.rpm

# 2. Install dependencies
sudo dnf install -y wireguard-tools postgresql-server redis

# 3. Initialize PostgreSQL
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql redis

# 4. Install GoConnect Server
sudo rpm -i goconnect-server_${VERSION}_linux_amd64.rpm

# 5. Configure (same as Debian)
sudo mkdir -p /etc/goconnect
sudo cat > /etc/goconnect/server.env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your-secure-password
REDIS_HOST=localhost
REDIS_PORT=6379
EOF
sudo chmod 600 /etc/goconnect/server.env

# 6. Setup PostgreSQL
sudo -u postgres createuser goconnect
sudo -u postgres createdb -O goconnect goconnect

# 7. Start the service
sudo systemctl enable --now goconnect-server
sudo systemctl status goconnect-server
```

---

### macOS

**Prerequisites:**
- macOS 11+ (Big Sur or later)
- Homebrew
- WireGuard tools

**Installation:**
```bash
# 1. Install dependencies
brew install wireguard-tools postgresql@15 redis

# 2. Start databases
brew services start postgresql@15
brew services start redis

# 3. Download GoConnect Server
VERSION="2.8.5"  # Replace with latest version
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    curl -LO https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-server_${VERSION}_darwin_arm64.tar.gz
else
    curl -LO https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-server_${VERSION}_darwin_amd64.tar.gz
fi

# 4. Extract and install
tar -xzf goconnect-server_${VERSION}_darwin_*.tar.gz
sudo mkdir -p /usr/local/bin /etc/goconnect
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# 5. Create configuration
sudo cat > /etc/goconnect/server.env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=$USER
REDIS_HOST=localhost
REDIS_PORT=6379
EOF

# 6. Setup PostgreSQL
createdb goconnect

# 7. Create LaunchDaemon for auto-start
sudo cat > /Library/LaunchDaemons/com.goconnect.server.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.goconnect.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/goconnect-server</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>JWT_SECRET</key>
        <string>your-jwt-secret</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/goconnect-server.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/goconnect-server.error.log</string>
</dict>
</plist>
EOF

# 8. Start the service
sudo launchctl load /Library/LaunchDaemons/com.goconnect.server.plist

# 9. Verify
curl http://localhost:8080/health
```

---

### Windows

**Prerequisites:**
- Windows 10/11 or Windows Server 2019+
- PowerShell 5.1+
- Administrator access

**Installation (PowerShell as Administrator):**
```powershell
# 1. Download GoConnect Server
$VERSION = "2.8.5"  # Replace with latest version
$URL = "https://github.com/orhaniscoding/goconnect/releases/download/v$VERSION/goconnect-server_${VERSION}_windows_amd64.zip"
Invoke-WebRequest -Uri $URL -OutFile "goconnect-server.zip"

# 2. Extract to Program Files
$InstallPath = "$env:ProgramFiles\GoConnect\Server"
New-Item -ItemType Directory -Force -Path $InstallPath
Expand-Archive -Path "goconnect-server.zip" -DestinationPath $InstallPath -Force

# 3. Add to PATH
$env:Path += ";$InstallPath"
[Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::Machine)

# 4. Create configuration directory
New-Item -ItemType Directory -Force -Path "$env:ProgramData\GoConnect"

# 5. Create environment file
$JwtSecret = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }) -as [byte[]])
@"
JWT_SECRET=$JwtSecret
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your-secure-password
REDIS_HOST=localhost
REDIS_PORT=6379
"@ | Out-File -FilePath "$env:ProgramData\GoConnect\server.env" -Encoding UTF8

# 6. Install as Windows Service
New-Service -Name "GoConnectServer" `
    -BinaryPathName "$InstallPath\goconnect-server.exe" `
    -DisplayName "GoConnect VPN Server" `
    -Description "GoConnect VPN Management Server" `
    -StartupType Automatic

# 7. Start the service
Start-Service -Name "GoConnectServer"

# 8. Verify
Get-Service -Name "GoConnectServer"
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

**Note:** You'll also need PostgreSQL and Redis running on Windows. Consider using Docker Desktop for Windows or WSL2.

---

## 💻 Client Daemon Installation

The GoConnect Daemon runs on client devices and manages VPN connections.

### Linux

**Debian/Ubuntu (.deb) - Recommended:**
```bash
VERSION="2.8.5"
wget https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-daemon_${VERSION}_linux_amd64.deb
sudo dpkg -i goconnect-daemon_${VERSION}_linux_amd64.deb
sudo systemctl enable --now goconnect-daemon
sudo systemctl status goconnect-daemon
```

**RHEL/Fedora (.rpm):**
```bash
VERSION="2.8.5"
wget https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-daemon_${VERSION}_linux_amd64.rpm
sudo rpm -i goconnect-daemon_${VERSION}_linux_amd64.rpm
sudo systemctl enable --now goconnect-daemon
```

**Portable (.tar.gz):**
```bash
VERSION="2.8.5"
wget https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-daemon_${VERSION}_linux_amd64.tar.gz
tar -xzf goconnect-daemon_${VERSION}_linux_amd64.tar.gz
cd goconnect-daemon_*

# Install binary
sudo cp goconnect-daemon /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-daemon

# Install systemd service
sudo cp install/linux/goconnect-daemon.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now goconnect-daemon
```

---

### macOS

```bash
# 1. Download
VERSION="2.8.5"
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    curl -LO https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-daemon_${VERSION}_darwin_arm64.tar.gz
else
    curl -LO https://github.com/orhaniscoding/goconnect/releases/download/v${VERSION}/goconnect-daemon_${VERSION}_darwin_amd64.tar.gz
fi

# 2. Extract
tar -xzf goconnect-daemon_${VERSION}_darwin_*.tar.gz
cd goconnect-daemon_*

# 3. Install binary
sudo cp goconnect-daemon /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-daemon

# 4. Install LaunchDaemon
sudo cp install/macos/com.goconnect.daemon.plist /Library/LaunchDaemons/
sudo launchctl load /Library/LaunchDaemons/com.goconnect.daemon.plist

# 5. Verify
sudo launchctl list | grep goconnect
```

---

### Windows

**PowerShell (Run as Administrator):**
```powershell
# 1. Download
$VERSION = "2.8.5"
$URL = "https://github.com/orhaniscoding/goconnect/releases/download/v$VERSION/goconnect-daemon_${VERSION}_windows_amd64.zip"
Invoke-WebRequest -Uri $URL -OutFile "goconnect-daemon.zip"

# 2. Extract
$InstallPath = "$env:ProgramFiles\GoConnect\Daemon"
New-Item -ItemType Directory -Force -Path $InstallPath
Expand-Archive -Path "goconnect-daemon.zip" -DestinationPath $InstallPath -Force

# 3. Install as service
& "$InstallPath\install\windows\install.ps1"

# 4. Verify
Get-Service -Name "GoConnectDaemon"
```

---

## ⚙️ Post-Installation

### 1. Verify Server Health
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy"}
```

### 2. Access Web UI
Open `http://localhost:8080` in your browser.

### 3. Create Admin User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"secure-password","name":"Admin"}'
```

### 4. Configure Client
```bash
# On client machine, configure server URL
goconnect-daemon config set server_url https://vpn.yourdomain.com
goconnect-daemon login
```

---

## 🔧 Troubleshooting

### Server Won't Start

```bash
# Check logs
journalctl -u goconnect-server -n 50 --no-pager

# Verify database connection
psql -h localhost -U goconnect -d goconnect -c "SELECT 1"

# Check port availability
netstat -tlnp | grep 8080
```

### Daemon Won't Connect

```bash
# Check daemon logs
journalctl -u goconnect-daemon -n 50 --no-pager

# Test server connectivity
curl https://vpn.yourdomain.com/health

# Check WireGuard
sudo wg show
```

### Permission Denied

```bash
# Linux: Add user to wireguard group
sudo usermod -aG wireguard $USER

# macOS: Check System Preferences > Security & Privacy
```

---

## 📚 Next Steps

- [Configuration Guide](./CONFIG_FLAGS.md) - Environment variables and settings
- [Security Guide](./SECURITY.md) - Production hardening
- [API Reference](./API_EXAMPLES.http) - REST API documentation
- [Admin Guide](./ADMIN_GUIDE.md) - Dashboard usage
