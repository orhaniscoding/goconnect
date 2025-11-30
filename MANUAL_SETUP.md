# GoConnect Manual Setup Guide

## Problem: Setup API works but browser shows 404

## Solution: Complete setup via API calls

### Step 1: Database Configuration
```bash
curl -X POST http://localhost:8080/setup \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "server": {"host": "0.0.0.0", "port": "8080"},
      "database": {
        "backend": "postgres",
        "host": "postgres", 
        "port": "5432",
        "user": "goconnect",
        "password": "goconnect_secret",
        "dbname": "goconnect"
      }
    }
  }'
```

### Step 2: JWT and WireGuard Configuration
```bash
curl -X POST http://localhost:8080/setup \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "server": {"host": "0.0.0.0", "port": "8080"},
      "database": {
        "backend": "postgres",
        "host": "postgres",
        "port": "5432", 
        "user": "goconnect",
        "password": "goconnect_secret",
        "dbname": "goconnect"
      },
      "jwt": {
        "secret": "change-me-in-production-minimum-32-chars",
        "access_token_ttl": "1h",
        "refresh_token_ttl": "24h"
      },
      "wireguard": {
        "server_endpoint": "vpn.example.com:51820",
        "server_pubkey": "sc5AnxyIlmfyG97Le1tSjP5GiedfwNQDYj1tSEBOWG8="
      }
    }
  }'
```

### Step 3: Finalize Setup with Restart
```bash
curl -X POST http://localhost:8080/setup \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "server": {"host": "0.0.0.0", "port": "8080"},
      "database": {
        "backend": "postgres",
        "host": "postgres",
        "port": "5432",
        "user": "goconnect", 
        "password": "goconnect_secret",
        "dbname": "goconnect"
      },
      "jwt": {
        "secret": "change-me-in-production-minimum-32-chars",
        "access_token_ttl": "1h", 
        "refresh_token_ttl": "24h"
      },
      "wireguard": {
        "server_endpoint": "vpn.example.com:51820",
        "server_pubkey": "sc5AnxyIlmfyG97Le1tSjP5GiedfwNQDYj1tSEBOWG8="
      }
    },
    "restart": true
  }'
```

### Step 4: Verify Setup
```bash
# Wait 5 seconds for restart
sleep 5

# Check health
curl http://localhost:8080/health

# Should return: {"mode":"production","ok":true,"service":"goconnect-server"}
```

### Step 5: Access GoConnect
```
http://localhost:8080
```

## Default Admin Access
After setup, the first user to access http://localhost:8080 will be prompted to create the admin account.

## Next Steps
1. Create admin account
2. Create first tenant ("Gaming Network")
3. Create network (Virtual LAN)
4. Invite members

## Troubleshooting
If setup fails:
1. Check Docker containers: `docker compose ps`
2. Check server logs: `docker compose logs server`
3. Restart server: `docker compose restart server`
4. Try setup again
