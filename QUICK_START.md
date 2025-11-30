# GoConnect Quick Start (Manual Setup)

## Option 1: Docker (Recommended)
1. Start Docker Desktop
2. Run: `docker compose up -d`
3. Open: http://localhost:3000
4. Follow setup wizard

## Option 2: Local Setup (Troubleshooting)

### Server Issues?
If server won't bind to port, try:

```powershell
# Check what's using port 8080-8085
netstat -an | findstr :808

# Kill any conflicting processes
taskkill /F /IM server.exe

# Try different port
$env:SERVER_PORT="8085"
go run ./cmd/server --setup-mode --db-backend=sqlite
```

### PowerShell Script Issues?
```powershell
# Fix encoding
chcp 65001

# Run script with bypass
powershell -ExecutionPolicy Bypass -File "test_setup.ps1"
```

### Manual Setup Steps:
1. Start server on port 8085
2. Open: http://localhost:8085/setup
3. Configure:
   - Database: SQLite
   - Path: ./goconnect.db
   - JWT Secret: any-32-char-string
4. Complete setup
5. Access: http://localhost:8085

## Verification:
```bash
# Check server health
curl http://localhost:8085/health

# Check setup status
curl http://localhost:8085/setup/status
```

## Next Steps:
1. Create admin account
2. Create first tenant
3. Create network
4. Install daemon on client machines
