# GoConnect Daemon - Windows Installation

## Quick Install

1. Extract the ZIP file to a temporary location
2. Right-click `install.ps1` and select "Run as Administrator"
3. The daemon will be installed as a Windows service

## What Gets Installed

- Binary: `C:\Program Files\GoConnect\goconnect-daemon.exe`
- Service: `GoConnectDaemon` (Windows Service)
- Config: `C:\ProgramData\GoConnect\config.yaml` (you need to create this)

## Configuration

Before the service can start, you need to create a configuration file:

1. Create directory: `C:\ProgramData\GoConnect\`
2. Create file: `C:\ProgramData\GoConnect\config.yaml`
3. Add your server configuration:

```yaml
server:
  url: https://your-vpn-server.com
  api_key: your-api-key-here

wireguard:
  interface_name: wg0
  listen_port: 51820

logging:
  level: info
  file: C:\ProgramData\GoConnect\daemon.log
```

## Service Management

### Start Service
```powershell
Start-Service GoConnectDaemon
```

### Stop Service
```powershell
Stop-Service GoConnectDaemon
```

### Check Service Status
```powershell
Get-Service GoConnectDaemon
```

### View Logs
Check `C:\ProgramData\GoConnect\daemon.log`

## Uninstall

Run `uninstall.ps1` as Administrator to remove the service and binary.

## Troubleshooting

### Service won't start
- Check if config file exists: `C:\ProgramData\GoConnect\config.yaml`
- Verify config syntax is valid YAML
- Check logs at: `C:\ProgramData\GoConnect\daemon.log`
- Run manually to see errors: `& "C:\Program Files\GoConnect\goconnect-daemon.exe"`

### Permission errors
- Ensure you ran `install.ps1` as Administrator
- Check Windows Firewall isn't blocking the service

### WireGuard errors
- Install WireGuard for Windows: https://www.wireguard.com/install/
- Ensure WireGuard kernel driver is installed

## Manual Installation

If the script doesn't work:

1. Copy `goconnect-daemon.exe` to `C:\Program Files\GoConnect\`
2. Create the service:
```powershell
New-Service -Name "GoConnectDaemon" `
    -BinaryPathName "C:\Program Files\GoConnect\goconnect-daemon.exe" `
    -DisplayName "GoConnect Daemon" `
    -Description "GoConnect VPN Client Daemon" `
    -StartupType Manual
```
3. Create config file (see Configuration section)
4. Start service: `Start-Service GoConnectDaemon`
