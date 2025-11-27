# GoConnect Server - macOS Installation

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
