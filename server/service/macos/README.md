# GoConnect Server - macOS Installation

## Quick Install

```bash
tar -xzf goconnect-server_*_darwin_*.tar.gz
cd goconnect-server_*_darwin_*
chmod +x install.sh
sudo ./install.sh
```

The installer will:
1. Install binary to `/usr/local/bin/goconnect-server`
2. Create LaunchDaemon `com.goconnect.server.plist`
3. Create config directory `/etc/goconnect/`
4. Create example config `/etc/goconnect/.env`

## Configuration

**REQUIRED**: Edit the configuration file before starting.

```bash
sudo nano /etc/goconnect/.env
```

### Essential Settings

```bash
# Database (PostgreSQL required)
DB_HOST=localhost
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password

# JWT Secret (min 32 chars)
JWT_SECRET=your_jwt_secret_min_32_chars

# WireGuard
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=your_44_char_public_key
WG_PRIVATE_KEY=your_private_key
```

### Generate Keys

```bash
# JWT Secret
openssl rand -base64 32

# WireGuard Keys
wg genkey  # Private
echo "YOUR_PRIVATE_KEY" | wg pubkey  # Public
```

See `config.example.env` for all available options.

## Database Setup

```bash
# Install PostgreSQL
brew install postgresql
brew services start postgresql

# Create database
psql postgres
CREATE DATABASE goconnect;
CREATE USER goconnect WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
```

## Run Migrations

```bash
sudo /usr/local/bin/goconnect-server -migrate
```

## Service Management

```bash
# Start
sudo launchctl load /Library/LaunchDaemons/com.goconnect.server.plist

# Stop
sudo launchctl unload /Library/LaunchDaemons/com.goconnect.server.plist

# Status
sudo launchctl list | grep goconnect

# Logs
sudo log show --predicate 'process == "goconnect-server"' --last 10m
```

## Uninstall

```bash
sudo ./uninstall.sh
```

## Troubleshooting

See `/docs/CONFIGURATION.md` for complete guide.

### Quick Checks

```bash
# Check config
ls -l /etc/goconnect/.env

# Test database
psql -h localhost -U goconnect -d goconnect

# Check service
sudo launchctl list | grep goconnect
```

## Requirements

- macOS 11.0+ (Big Sur or newer)
- PostgreSQL 12+: `brew install postgresql`
- WireGuard: `brew install wireguard-tools`
- Administrator privileges

## Quick Install

```bash
tar -xzf goconnect-server_*_darwin_*.tar.gz
cd goconnect-server_*_darwin_*
chmod +x install.sh
sudo ./install.sh
```

## What Gets Installed

- Binary: `/usr/local/bin/goconnect-server`
- LaunchDaemon: `/Library/LaunchDaemons/com.goconnect.server.plist`
- Config: `/etc/goconnect/config.yaml`

## Configuration

Create `/etc/goconnect/config.yaml`:

```yaml
server:
  port: 8080
  host: 0.0.0.0

database:
  host: localhost
  port: 5432
  user: goconnect
  password: your-password
  dbname: goconnect

jwt:
  secret: your-secret-key-min-32-chars-long

redis:
  addr: localhost:6379
  password: ""
  db: 0

wireguard:
  interface: utun0
  subnet: 10.8.0.0/24
```

## Service Management

```bash
# Load and start
sudo launchctl load /Library/LaunchDaemons/com.goconnect.server.plist

# Stop and unload
sudo launchctl unload /Library/LaunchDaemons/com.goconnect.server.plist

# Check if running
sudo launchctl list | grep goconnect

# View logs
tail -f /var/log/goconnect-server.log
```

## Uninstall

```bash
sudo ./uninstall.sh
```

## Requirements

- macOS 11+ (Big Sur or later)
- PostgreSQL: `brew install postgresql`
- Redis: `brew install redis`
