# GoConnect Server - Windows Installation

## Quick Install

1. Extract the ZIP file
2. Right-click `install.ps1` and select "Run as Administrator"
3. Edit configuration: `C:\ProgramData\GoConnect\.env`
4. Run migrations: `& "C:\Program Files\GoConnect\goconnect-server.exe" -migrate`
5. Start service: `Start-Service GoConnectServer`

## What Gets Installed

- **Binary**: `C:\Program Files\GoConnect\goconnect-server.exe`
- **Service**: `GoConnectServer` (Windows Service, Manual startup)
- **Config**: `C:\ProgramData\GoConnect\.env` (auto-created from example)

## Configuration

The installer automatically creates a configuration file with defaults.

### Required Configuration

Edit `C:\ProgramData\GoConnect\.env` and configure these essential settings:

```bash
# Database (PostgreSQL required)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password_here

# JWT Secret (minimum 32 characters)
JWT_SECRET=your_very_secure_random_jwt_secret_key_min_32_chars

# WireGuard Configuration
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=your_wireguard_public_key_44_chars_base64
WG_PRIVATE_KEY=your_wireguard_private_key_here
```

### Generate Required Keys

```powershell
# Generate JWT Secret (32+ characters)
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))

# Generate WireGuard keys (requires WireGuard installed)
wg genkey  # Private key
echo "YOUR_PRIVATE_KEY" | wg pubkey  # Public key
```

See `config.example.env` for all available options and detailed documentation.

## Database Setup

### Install PostgreSQL

Download from: https://www.postgresql.org/download/windows/

### Create Database

```powershell
# Using psql
psql -U postgres
```

```sql
CREATE DATABASE goconnect;
CREATE USER goconnect WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE goconnect TO goconnect;
\q
```

## Run Migrations

Before starting the server for the first time:

```powershell
& "C:\Program Files\GoConnect\goconnect-server.exe" -migrate
```

## Service Management

```powershell
# Start service
Start-Service GoConnectServer

# Stop service
Stop-Service GoConnectServer

# Check status
Get-Service GoConnectServer

# View service logs
Get-EventLog -LogName Application -Source GoConnectServer -Newest 20
```

## Uninstall

Run `uninstall.ps1` as Administrator.

Configuration files in `C:\ProgramData\GoConnect` are not deleted automatically.

## Troubleshooting

### Service won't start

1. **Check configuration**:
   ```powershell
   Test-Path "C:\ProgramData\GoConnect\.env"
   Get-Content "C:\ProgramData\GoConnect\.env"
   ```

2. **Verify database connection**:
   ```powershell
   # Test PostgreSQL connection
   psql -h localhost -p 5432 -U goconnect -d goconnect
   ```

3. **Check JWT secret length**:
   ```powershell
   # Must be at least 32 characters
   (Get-Content "C:\ProgramData\GoConnect\.env" | Select-String "JWT_SECRET").Line.Length
   ```

4. **Verify WireGuard keys**:
   ```powershell
   # Public key must be exactly 44 characters
   (Get-Content "C:\ProgramData\GoConnect\.env" | Select-String "WG_SERVER_PUBKEY").Line
   ```

5. **Check service logs**:
   ```powershell
   Get-EventLog -LogName Application -Source GoConnectServer -Newest 20
   ```

### Database connection failed

- Verify PostgreSQL is running: `Get-Service postgresql*`
- Check credentials in `.env` file
- Ensure database exists: `psql -U postgres -l`
- Test connection manually

### Migrations failed

- Ensure database user has proper permissions
- Check if migrations were already run
- Verify database is accessible

### Port already in use

- Change `SERVER_PORT` in `.env` file
- Check what's using port 8080: `netstat -ano | findstr :8080`

## Requirements

- **Windows**: 10/11 or Server 2019+
- **PostgreSQL**: 12+ (required)
- **Redis**: 6+ (optional, for session management)
- **WireGuard**: For VPN functionality
- **Administrator**: Privileges required for installation

## Additional Help

For more detailed configuration information, see:
- `config.example.env` - Environment variable reference
- `/docs/CONFIGURATION.md` - Complete configuration guide
- `/docs/POSTGRESQL_SETUP.md` - Database setup guide
- GitHub Issues: https://github.com/orhaniscoding/goconnect/issues
