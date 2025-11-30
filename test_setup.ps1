# GoConnect Setup Test Script
# Run this after starting the server on port 8081

Write-Host "🚀 Testing GoConnect Setup Wizard..." -ForegroundColor Green

# Test 1: Check Setup Status
Write-Host "`n📋 Step 1: Checking setup status..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/setup/status" -Method GET
    Write-Host "✅ Setup API is working" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Cyan
    Write-Host "Next Step: $($response.next_step)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Setup API not responding" -ForegroundColor Red
    Write-Host "Make sure server is running on port 8081" -ForegroundColor Yellow
    exit 1
}

# Test 2: Configure Database (Personal/SQLite)
Write-Host "`n🗄️ Step 2: Configuring SQLite database..." -ForegroundColor Yellow
$dbConfig = @{
    config = @{
        server = @{
            host = "0.0.0.0"
            port = "8081"
        }
        database = @{
            backend = "sqlite"
            sqlite_path = "./goconnect.db"
        }
    }
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/setup" -Method POST -Body ($dbConfig | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ Database configured" -ForegroundColor Green
} catch {
    Write-Host "❌ Database configuration failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Configure JWT and Admin
Write-Host "`n🔐 Step 3: Configuring JWT and admin..." -ForegroundColor Yellow
$jwtConfig = @{
    config = @{
        server = @{
            host = "0.0.0.0"
            port = "8081"
        }
        database = @{
            backend = "sqlite"
            sqlite_path = "./goconnect.db"
        }
        jwt = @{
            secret = "ssssssssssssssssssssssssssssssss"
            access_token_ttl = "1h"
            refresh_token_ttl = "24h"
        }
        wireguard = @{
            server_endpoint = "auto"
            server_pubkey = "auto"
        }
    }
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/setup" -Method POST -Body ($jwtConfig | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ JWT and admin configured" -ForegroundColor Green
} catch {
    Write-Host "❌ JWT configuration failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Finalize Setup
Write-Host "`n🎉 Step 4: Finalizing setup..." -ForegroundColor Yellow
$finalConfig = @{
    config = @{
        server = @{
            host = "0.0.0.0"
            port = "8081"
        }
        database = @{
            backend = "sqlite"
            sqlite_path = "./goconnect.db"
        }
        jwt = @{
            secret = "ssssssssssssssssssssssssssssssss"
            access_token_ttl = "1h"
            refresh_token_ttl = "24h"
        }
        wireguard = @{
            server_endpoint = "example.com:51820"
            server_pubkey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
        }
    }
    restart = $true
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/setup" -Method POST -Body ($finalConfig | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ Setup completed successfully!" -ForegroundColor Green
    Write-Host "Server will restart and be ready at http://localhost:8081" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Setup finalization failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Verify Setup
Write-Host "`n🔍 Step 5: Verifying setup..." -ForegroundColor Yellow
Start-Sleep -Seconds 3  # Wait for server restart

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "✅ Server is running after setup!" -ForegroundColor Green
    Write-Host "Health Status: $($response.status)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Server health check failed after setup" -ForegroundColor Red
}

Write-Host "`n🎯 Setup test completed!" -ForegroundColor Green
Write-Host "Open http://localhost:8081 to access the web interface" -ForegroundColor Cyan
Write-Host "Default admin will be created on first access" -ForegroundColor Cyan
