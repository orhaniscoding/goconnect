# GoConnect Daemon - Windows Installation

## Quick Install

1. Extract the ZIP file to a temporary location
2. Right-click `install.ps1` and select "Run as Administrator"
3. Edit the configuration file: `C:\ProgramData\GoConnect\config.yaml`
4. Start the service: `Start-Service GoConnectDaemon`

## What Gets Installed

- **Binary**: `C:\Program Files\GoConnect\goconnect.exe`
- **Service**: `GoConnectDaemon` (Windows Service, Manual startup)
- **Config**: `C:\ProgramData\GoConnect\config.yaml` (auto-created from example)

## Configuration

The installer automatically creates a configuration file with defaults.

### Required Configuration

Edit `C:\ProgramData\GoConnect\config.yaml` and set your server URL:

```yaml
# REQUIRED: Your GoConnect server URL
server_url: "https://vpn.example.com:8080"
```

### Optional Settings

```yaml
# Local API port (default: 12345)
local_port: 12345

# Logging level: debug, info, warn, error (default: info)
log_level: "info"

# WireGuard interface name (default: wg0)
interface_name: "wg0"

# Connection settings
reconnect_interval: 30        # seconds
health_check_interval: 60     # seconds
```

See `config.example.yaml` for all available options and detailed documentation.

## Service Management

### Start Service
```powershell
Start-Service GoConnect
```

### Stop Service
```powershell
Stop-Service GoConnect
```

### Check Service Status
```powershell
Get-Service GoConnect
```

### View Service Logs
```powershell
# View recent service events
Get-EventLog -LogName Application -Source GoConnectDaemon -Newest 20

# Or if log_file is configured in config.yaml:
Get-Content C:\ProgramData\GoConnect\daemon.log -Tail 50 -Wait
```

## Uninstall

Run `uninstall.ps1` as Administrator to remove the service and binary.

Configuration files in `C:\ProgramData\GoConnect` are not deleted automatically.

## Troubleshooting

### Service won't start

1. **Check configuration**:
   ```powershell
   Test-Path "C:\ProgramData\GoConnect\config.yaml"
   Get-Content "C:\ProgramData\GoConnect\config.yaml"
   ```

2. **Verify server URL is set**:
   ```powershell
   Select-String -Path "C:\ProgramData\GoConnect\config.yaml" -Pattern "server_url"
   ```

3. **Check service logs**:
   ```powershell
   Get-EventLog -LogName Application -Source GoConnectDaemon -Newest 20
   ```

4. **Test binary manually**:
   ```powershell
   & "C:\Program Files\GoConnect\goconnect.exe" --version
   ```

### Connection issues

- Verify server URL is correct and accessible
- Check Windows Firewall settings
- Ensure server is running and reachable

### WireGuard errors

- Install WireGuard for Windows: https://www.wireguard.com/install/
- Ensure WireGuard kernel driver is installed
- Check interface name in config matches system

## Manual Installation

If the script doesn't work:

1. **Copy binary**:
   ```powershell
   New-Item -ItemType Directory -Force -Path "C:\Program Files\GoConnect"
   Copy-Item goconnect.exe "C:\Program Files\GoConnect\"
   ```

2. **Create config directory**:
   ```powershell
   New-Item -ItemType Directory -Force -Path "C:\ProgramData\GoConnect"
   Copy-Item config.example.yaml "C:\ProgramData\GoConnect\config.yaml"
   ```

3. **Create service**:
   ```powershell
   New-Service -Name "GoConnectDaemon" `
       -BinaryPathName '"C:\Program Files\GoConnect\goconnect.exe"' `
       -DisplayName "GoConnect Daemon" `
       -Description "GoConnect VPN Client Daemon" `
       -StartupType Manual
   ```

4. **Edit config** and start service:
   ```powershell
   notepad "C:\ProgramData\GoConnect\config.yaml"
   Start-Service GoConnect
   ```

## Additional Help

For more detailed configuration information, see:
- `config.example.yaml` - Configuration file reference
- `/docs/CONFIGURATION.md` - Complete configuration guide
- GitHub Issues: https://github.com/orhaniscoding/goconnect/issues
