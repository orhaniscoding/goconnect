# GoConnect Daemon - Linux Installation

## Quick Install

```bash
tar -xzf goconnect-daemon_*_linux_*.tar.gz
cd goconnect-daemon_*_linux_*
chmod +x install.sh
sudo ./install.sh
```

The installer will:
1. Install binary to `/usr/local/bin/goconnect-daemon`
2. Create systemd service `goconnect-daemon.service`
3. Create config directory `/etc/goconnect/`
4. Create example config `/etc/goconnect/config.yaml`

## Configuration

**REQUIRED**: Edit the configuration file before starting the service.

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

# Check service logs
sudo journalctl -u goconnect-daemon -n 50

# Test binary
/usr/local/bin/goconnect-daemon --version
```

## Requirements

- Linux kernel 5.6+ (for WireGuard)
- systemd
- WireGuard tools: `sudo apt install wireguard-tools`
- Root privileges for installation
