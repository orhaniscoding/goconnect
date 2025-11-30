# GoConnect Daemon - macOS Installation

## Quick Install

```bash
tar -xzf goconnect-daemon_*_darwin_*.tar.gz
cd goconnect-daemon_*_darwin_*
chmod +x install.sh
sudo ./install.sh
```

The installer will:
1. Install binary to `/usr/local/bin/goconnect-daemon`
2. Create LaunchDaemon `com.goconnect.daemon.plist`
3. Create config directory `/etc/goconnect/`
4. Create example config `/etc/goconnect/config.yaml`

## Configuration

**REQUIRED**: Edit the configuration file before the service will work.

```bash
sudo nano /etc/goconnect/config.yaml
```

### Minimum Configuration

```yaml
# REQUIRED: Your GoConnect server URL
server_url: "https://vpn.example.com:8080"
```

See `config.example.yaml` for all available options.

## Service Management

```bash
# Start service
sudo launchctl load /Library/LaunchDaemons/com.goconnect.daemon.plist

# Stop service
sudo launchctl unload /Library/LaunchDaemons/com.goconnect.daemon.plist

# Check status
sudo launchctl list | grep goconnect

# View logs
sudo log show --predicate 'process == "goconnect-daemon"' --last 10m
```

## Uninstall

```bash
sudo ./uninstall.sh
```

## Troubleshooting

See `/docs/CONFIGURATION.md` for complete troubleshooting guide.

### Quick Checks

```bash
# Check configuration exists
ls -l /etc/goconnect/config.yaml

# Verify server URL is set
grep server_url /etc/goconnect/config.yaml

# Check service status
sudo launchctl list | grep goconnect

# Test binary
/usr/local/bin/goconnect-daemon --version
```

## Requirements

- macOS 11.0+ (Big Sur or newer)
- WireGuard tools: `brew install wireguard-tools`
- Administrator privileges for installation

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
