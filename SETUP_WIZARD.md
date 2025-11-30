# GoConnect Setup Wizard Guide

## Quick Start (5 Minutes)

### Step 1: Start Server
```powershell
cd server
$env:SERVER_PORT="8081"
go run ./cmd/server --setup-mode --db-backend=sqlite
```

### Step 2: Open Setup Wizard
```
http://localhost:8081/setup
```

### Step 3: Configure Database
Select "Personal (SQLite)" mode:
- Database backend: sqlite
- SQLite path: ./goconnect.db
- Server port: 8081

### Step 4: Create Admin Account
- JWT Secret: auto-generated (32 chars)
- Admin Email: admin@goconnect.local
- Admin Password: admin123

### Step 5: WireGuard Configuration
- Server Endpoint: auto-detect
- Server Public Key: auto-generate

### Step 6: Finalize Setup
- Click "Complete Setup"
- Server restarts automatically
- Ready for use!

## Post-Setup

### Access Web Interface
```
http://localhost:8081
```

### Default Admin Login
- Email: admin@goconnect.local
- Password: admin123

### Create First Tenant
1. Login as admin
2. Click "Create Tenant"
3. Name: "My Gaming Network"
4. Description: "Private VPN for friends"

### Create First Network
1. Go to tenant dashboard
2. Click "Create Network"
3. Name: "Gaming LAN"
4. CIDR: 10.0.0.0/24 (auto-suggested)

## Testing the Setup

### 1. Verify Server
```bash
curl http://localhost:8081/health
# Should return: {"status":"ok"}
```

### 2. Test Web UI
- Open browser to http://localhost:8081
- Login with admin credentials
- Create tenant and network

### 3. Test Client Daemon
```bash
cd client-daemon
go run ./cmd/daemon
```

## Troubleshooting

### Port 8080 Conflict
If PostgreSQL is on 8080, use port 8081:
```powershell
$env:SERVER_PORT="8081"
```

### Database Issues
- Ensure SQLite path is writable
- Check file permissions for goconnect.db

### Setup Wizard Not Loading
- Verify server is running with --setup-mode
- Check browser console for errors
- Try http://localhost:8081/setup/status

## Next Steps

After setup completion:
1. Invite team members
2. Install client daemon on their machines
3. Connect to the network
4. Start gaming/file sharing!

## Production Deployment

For production use:
1. Use PostgreSQL instead of SQLite
2. Configure proper JWT secrets
3. Set up HTTPS with SSL certificates
4. Use Docker Compose for deployment

See: [DEPLOYMENT.md](docs/DEPLOYMENT.md) for details
