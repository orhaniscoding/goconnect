# GoConnect Server - Linux Installation

## Quick Install

```bash
tar -xzf goconnect-server_*_linux_*.tar.gz
cd goconnect-server_*_linux_*
chmod +x install.sh
sudo ./install.sh
```

## What Gets Installed

- Binary: `/usr/local/bin/goconnect-server`
- Service: `goconnect-server.service` (systemd)
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
  interface: wg0
  subnet: 10.8.0.0/24
```

## Service Management

```bash
# Start
sudo systemctl start goconnect-server

# Stop
sudo systemctl stop goconnect-server

# Status
sudo systemctl status goconnect-server

# Enable on boot
sudo systemctl enable goconnect-server

# Logs
sudo journalctl -u goconnect-server -f
```

## Uninstall

```bash
sudo ./uninstall.sh
```

## Requirements

- Linux kernel 5.6+
- PostgreSQL 12+
- Redis 6+
- systemd
