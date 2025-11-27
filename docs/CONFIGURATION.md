# GoConnect Configuration Guide

This guide explains how to configure both the GoConnect Server and Daemon.

## Table of Contents

- [Quick Start](#quick-start)
- [Daemon Configuration](#daemon-configuration)
- [Server Configuration](#server-configuration)
- [Platform-Specific Setup](#platform-specific-setup)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Daemon (Client)

The daemon requires **only one setting** to get started:

```yaml
server_url: "https://vpn.example.com:8080"
```

### Server

The server requires these essential settings:

1. **Database**: PostgreSQL connection details
2. **JWT Secret**: At least 32 characters for token security
3. **WireGuard**: Server endpoint and keys

## Daemon Configuration

### Configuration File Location

- **Windows**: `C:\ProgramData\GoConnect\config.yaml`
- **Linux**: `/etc/goconnect/config.yaml`
- **macOS**: `/etc/goconnect/config.yaml`

### Required Settings

```yaml
# REQUIRED: Your GoConnect server URL
server_url: "https://vpn.example.com:8080"
```

### Optional Settings

```yaml
# Local API port for web UI communication (default: 12345)
local_port: 12345

# Logging level: debug, info, warn, error (default: info)
log_level: "info"

# Log file path (empty = stdout)
log_file: ""

# WireGuard interface name (default: wg0)
interface_name: "wg0"

# Reconnection settings
reconnect_interval: 30        # seconds
health_check_interval: 60     # seconds
```

### After Installation

1. **Edit config file** with your server URL
2. **Start the service**:
   - Windows: `Start-Service GoConnectDaemon`
   - Linux: `sudo systemctl start goconnect-daemon`
   - macOS: Service starts automatically

## Server Configuration

### Configuration File Location

The server uses environment variables, typically stored in:

- **All platforms**: `/etc/goconnect/.env` or `.env` in working directory

### Required Settings

#### 1. Server Settings

```bash
# Server listen configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_ENVIRONMENT=production  # or 'development'
```

#### 2. Database (PostgreSQL)

```bash
# PostgreSQL connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password_here

# Connection pool
DB_SSLMODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s
```

**Setup PostgreSQL**:

```bash
# Create database and user
sudo -u postgres psql
CREATE DATABASE goconnect;
CREATE USER goconnect WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
```

#### 3. JWT Configuration

```bash
# Generate a secure secret (minimum 32 characters)
JWT_SECRET=your_very_secure_random_jwt_secret_key_min_32_chars

# Token lifetimes
JWT_ACCESS_TTL=15m          # Access token: 15 minutes
JWT_REFRESH_TTL=168h        # Refresh token: 7 days
JWT_REFRESH_SECRET=         # Optional: separate refresh token secret
```

**Generate JWT Secret**:

```bash
# Linux/macOS
openssl rand -base64 32

# Windows PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
```

#### 4. WireGuard Configuration

```bash
# Generate WireGuard keys
WG_PRIVATE_KEY=your_wireguard_private_key_here
WG_SERVER_PUBKEY=your_wireguard_public_key_44_chars_base64

# Server endpoint (what clients connect to)
WG_SERVER_ENDPOINT=vpn.example.com:51820

# Interface settings
WG_INTERFACE_NAME=wg0
WG_PORT=51820
WG_DNS=1.1.1.1,8.8.8.8
WG_MTU=1420
WG_KEEPALIVE=25
```

**Generate WireGuard Keys**:

```bash
# Generate private key
wg genkey

# Generate public key from private key
echo "YOUR_PRIVATE_KEY" | wg pubkey
```

### Optional Settings

#### Redis (Session Management)

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

#### Audit Logging

```bash
AUDIT_SQLITE_DSN=./audit.db
AUDIT_HASH_SECRETS=your_base64_secret_for_hashing
AUDIT_ASYNC=true
AUDIT_QUEUE_SIZE=1024
AUDIT_WORKER_COUNT=1
AUDIT_FLUSH_INTERVAL=5s
```

#### CORS

```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.example.com
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=3600
```

### After Configuration

1. **Run database migrations**:
   ```bash
   ./goconnect-server -migrate
   ```

2. **Start the service**:
   - Windows: `Start-Service GoConnectServer`
   - Linux: `sudo systemctl start goconnect-server`
   - macOS: Service starts automatically

## Platform-Specific Setup

### Windows

#### Daemon
```powershell
# Edit config
notepad "C:\ProgramData\GoConnect\config.yaml"

# Start service
Start-Service GoConnectDaemon

# Check status
Get-Service GoConnectDaemon
```

#### Server
```powershell
# Edit config
notepad "C:\ProgramData\GoConnect\.env"

# Run migrations
& "C:\Program Files\GoConnect\goconnect-server.exe" -migrate

# Start service
Start-Service GoConnectServer
```

### Linux

#### Daemon
```bash
# Edit config
sudo nano /etc/goconnect/config.yaml

# Start and enable
sudo systemctl start goconnect-daemon
sudo systemctl enable goconnect-daemon

# Check status
sudo systemctl status goconnect-daemon
```

#### Server
```bash
# Edit config
sudo nano /etc/goconnect/.env

# Run migrations
sudo /usr/local/bin/goconnect-server -migrate

# Start and enable
sudo systemctl start goconnect-server
sudo systemctl enable goconnect-server

# Check status
sudo systemctl status goconnect-server
```

### macOS

#### Daemon
```bash
# Edit config
sudo nano /etc/goconnect/config.yaml

# Service starts automatically, or:
sudo launchctl load /Library/LaunchDaemons/com.goconnect.daemon.plist

# Check status
sudo launchctl list | grep goconnect
```

#### Server
```bash
# Edit config
sudo nano /etc/goconnect/.env

# Run migrations
sudo /usr/local/bin/goconnect-server -migrate

# Service starts automatically, or:
sudo launchctl load /Library/LaunchDaemons/com.goconnect.server.plist

# Check status
sudo launchctl list | grep goconnect
```

## Security Best Practices

### JWT Secrets

- **Minimum 32 characters** (preferably 64+)
- Use cryptographically random values
- Never commit secrets to version control
- Rotate regularly (requires re-authentication)

### Database

- **Use strong passwords** (16+ characters)
- Enable SSL/TLS (`DB_SSLMODE=require` or higher)
- Create dedicated database user with minimal privileges
- Regular backups

### WireGuard

- **Keep private keys secret** - never share or commit
- Use strong, random keys generated by `wg genkey`
- Public keys can be shared freely
- Rotate keys periodically

### File Permissions

Configuration files contain secrets and should be protected:

```bash
# Linux/macOS
sudo chmod 600 /etc/goconnect/config.yaml
sudo chmod 600 /etc/goconnect/.env
sudo chown root:root /etc/goconnect/config.yaml
sudo chown root:root /etc/goconnect/.env

# Windows (PowerShell)
$acl = Get-Acl "C:\ProgramData\GoConnect\config.yaml"
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule("SYSTEM","FullControl","Allow")
$acl.AddAccessRule($rule)
Set-Acl "C:\ProgramData\GoConnect\config.yaml" $acl
```

### Network Security

- Use HTTPS for server endpoint (Let's Encrypt)
- Configure firewall rules for WireGuard port
- Use internal network for database connection
- Consider VPN/SSH tunneling for database access

## Troubleshooting

### Daemon Won't Start

1. **Check configuration file exists**:
   ```bash
   # Windows
   Test-Path "C:\ProgramData\GoConnect\config.yaml"
   
   # Linux/macOS
   ls -l /etc/goconnect/config.yaml
   ```

2. **Verify server URL is set**:
   ```bash
   # Linux/macOS
   grep server_url /etc/goconnect/config.yaml
   
   # Windows
   Select-String -Path "C:\ProgramData\GoConnect\config.yaml" -Pattern "server_url"
   ```

3. **Check logs**:
   ```bash
   # Windows
   Get-EventLog -LogName Application -Source GoConnectDaemon -Newest 20
   
   # Linux
   sudo journalctl -u goconnect-daemon -n 50
   
   # macOS
   sudo log show --predicate 'process == "goconnect-daemon"' --last 10m
   ```

### Server Won't Start

1. **Verify environment variables loaded**:
   ```bash
   # Check if .env file exists
   ls -l /etc/goconnect/.env
   
   # Test configuration
   goconnect-server --version
   ```

2. **Database connection issues**:
   ```bash
   # Test PostgreSQL connection
   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
   
   # Check PostgreSQL is running
   sudo systemctl status postgresql  # Linux
   pg_isready  # Any platform with psql
   ```

3. **JWT secret validation**:
   ```bash
   # Ensure JWT_SECRET is at least 32 characters
   echo $JWT_SECRET | wc -c
   ```

4. **WireGuard key validation**:
   ```bash
   # Public key must be exactly 44 characters
   echo $WG_SERVER_PUBKEY | wc -c
   ```

5. **Check server logs**:
   ```bash
   # Linux
   sudo journalctl -u goconnect-server -n 50 -f
   
   # macOS
   sudo log show --predicate 'process == "goconnect-server"' --last 10m
   
   # Windows
   Get-EventLog -LogName Application -Source GoConnectServer -Newest 20
   ```

### Configuration Not Applied

- **Restart service** after configuration changes
- Check file permissions (must be readable by service)
- Verify YAML syntax (no tabs, correct indentation)
- For server, ensure environment variables are exported

### Permission Issues

```bash
# Linux/macOS: Run with sudo
sudo systemctl restart goconnect-server

# Windows: Run as Administrator
Start-Service GoConnectServer
```

### Connection Issues

1. **Check server is reachable**:
   ```bash
   curl https://vpn.example.com:8080/health
   ```

2. **Verify WireGuard port is open**:
   ```bash
   nc -zv vpn.example.com 51820
   ```

3. **Check firewall rules**:
   ```bash
   # Linux (UFW)
   sudo ufw status
   
   # Linux (iptables)
   sudo iptables -L -n
   
   # Windows
   Get-NetFirewallRule | Where-Object {$_.DisplayName -like "*GoConnect*"}
   ```

## Getting Help

If you continue to experience issues:

1. Check the logs (see troubleshooting section)
2. Review the example configuration files
3. Verify all required settings are present
4. Ensure services have proper permissions
5. Check GitHub issues for similar problems

For more help, visit: https://github.com/orhaniscoding/goconnect/issues
