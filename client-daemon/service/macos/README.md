# GoConnect Daemon - macOS Installation

## Quick Install

```bash
chmod +x install.sh
sudo ./install.sh
```

## What Gets Installed

- Binary: `/usr/local/bin/goconnect-daemon`
- LaunchDaemon: `/Library/LaunchDaemons/com.goconnect.daemon.plist`
- Config: `/etc/goconnect/config.yaml`

## Configuration

Create `/etc/goconnect/config.yaml`:

```yaml
server:
  url: https://your-vpn-server.com
  api_key: your-api-key-here

wireguard:
  interface_name: utun3
  listen_port: 51820

logging:
  level: info
  file: /var/log/goconnect-daemon.log
```

## Service Management

```bash
# Load and start service
sudo launchctl load /Library/LaunchDaemons/com.goconnect.daemon.plist

# Stop and unload service
sudo launchctl unload /Library/LaunchDaemons/com.goconnect.daemon.plist

# Check if running
sudo launchctl list | grep goconnect

# View logs
tail -f /var/log/goconnect-daemon.log
```

## Uninstall

```bash
sudo launchctl unload /Library/LaunchDaemons/com.goconnect.daemon.plist
sudo rm /usr/local/bin/goconnect-daemon
sudo rm /Library/LaunchDaemons/com.goconnect.daemon.plist
```

## Requirements

- macOS 10.15+ (Catalina or later)
- WireGuard Tools: `brew install wireguard-tools`
- Root privileges for installation
