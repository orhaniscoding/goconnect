# GoConnect Server - Windows Installation

## Quick Install

1. Extract the ZIP file
2. Right-click `install.ps1` and select "Run as Administrator"
3. Configure the server
4. Start the service

## What Gets Installed

- Binary: `C:\Program Files\GoConnect\goconnect-server.exe`
- Service: `GoConnectServer` (Windows Service)
- Config: `C:\ProgramData\GoConnect\config.yaml` (create manually)

## Configuration

Create `C:\ProgramData\GoConnect\config.yaml`:

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

```powershell
# Start service
Start-Service GoConnectServer

# Stop service
Stop-Service GoConnectServer

# Check status
Get-Service GoConnectServer

# View logs
Get-Content C:\ProgramData\GoConnect\server.log -Tail 50 -Wait
```

## Uninstall

Run `uninstall.ps1` as Administrator.

## Troubleshooting

### Service won't start
- Check config file exists and is valid
- Verify database connection
- Verify Redis connection
- Run manually: `& "C:\Program Files\GoConnect\goconnect-server.exe"`

### Port already in use
- Change `server.port` in config
- Check if another service uses port 8080

## Requirements

- Windows 10/11 or Server 2019+
- PostgreSQL 12+
- Redis 6+
- Administrator privileges for installation
