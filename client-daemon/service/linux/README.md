# GoConnect Daemon - Linux Installation

## Quick Install

```bash
chmod +x install.sh
sudo ./install.sh
```

## What Gets Installed

- Binary: `/usr/local/bin/goconnect-daemon`
- Service: `goconnect-daemon.service` (systemd)
- Config: `/etc/goconnect/config.yaml`

## Configuration

Create `/etc/goconnect/config.yaml`:

```yaml
server:
  url: https://your-vpn-server.com
  api_key: your-api-key-here

wireguard:
  interface_name: wg0
  listen_port: 51820

logging:
  level: info
  file: /var/log/goconnect-daemon.log
```

## Service Management

```bash
# Start service
sudo systemctl start goconnect-daemon

# Stop service
sudo systemctl stop goconnect-daemon

# Check status
sudo systemctl status goconnect-daemon

# Enable auto-start on boot
sudo systemctl enable goconnect-daemon

# View logs
sudo journalctl -u goconnect-daemon -f
```

## Uninstall

```bash
sudo systemctl stop goconnect-daemon
sudo systemctl disable goconnect-daemon
sudo rm /usr/local/bin/goconnect-daemon
sudo rm /etc/systemd/system/goconnect-daemon.service
sudo systemctl daemon-reload
```

## Requirements

- Linux kernel 5.6+ (for WireGuard)
- systemd
- Root privileges for installation
