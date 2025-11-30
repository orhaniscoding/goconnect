# GoConnect Setup Completion Script
# Run this to complete the setup wizard manually

Write-Host "🚀 Completing GoConnect Setup..." -ForegroundColor Green

# Step 1: Database Configuration (PostgreSQL)
Write-Host "`n📊 Step 1: Configuring PostgreSQL database..." -ForegroundColor Yellow

$dbConfig = @{
    config = @{
        server = @{
            host = "0.0.0.0"
            port = "8080"
        }
        database = @{
            backend = "postgres"
            host = "postgres"
            port = "5432"
            user = "postgres"
            password = "postgres"
            dbname = "goconnect"
        }
    }
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/setup" -Method POST -Body ($dbConfig | ConvertTo-Json -Depth 10) -ContentType "application/json"
    Write-Host "✅ Database configured" -ForegroundColor Green
} catch {
    Write-Host "❌ Database config failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Content)" -ForegroundColor Yellow
}

# Step 2: JWT and WireGuard Configuration
Write-Host "`n🔐 Step 2: Configuring JWT and WireGuard..." -ForegroundColor Yellow

$fullConfig = @{
    config = @{
        server = @{
            host = "0.0.0.0"
            port = "8080"
        }
        database = @{
            backend = "postgres"
            host = "postgres"
            port = "5432"
            user = "postgres"
            password = "postgres"
            dbname = "goconnect"
        }
        jwt = @{
            secret = "ssssssssssssssssssssssssssssssss"
            access_token_ttl = "1h"
            refresh_token_ttl = "24h"
        }
        wireguard = @{
            server_endpoint = "localhost:51820"
            server_pubkey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
        }
    }
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/setup" -Method POST -Body ($fullConfig | ConvertTo-Json -Depth 10) -ContentType "application/json"
    Write-Host "✅ JWT and WireGuard configured" -ForegroundColor Green
} catch {
    Write-Host "❌ JWT config failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Content)" -ForegroundColor Yellow
}

# Step 3: Finalize Setup with Restart
Write-Host "`n🎉 Step 3: Finalizing setup..." -ForegroundColor Yellow

$finalConfig = @{
    config = $fullConfig.config
    restart = $true
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/setup" -Method POST -Body ($finalConfig | ConvertTo-Json -Depth 10) -ContentType "application/json"
    Write-Host "✅ Setup completed! Server will restart..." -ForegroundColor Green
    Write-Host "Restart required: $($response.restart_required)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Setup finalization failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Content)" -ForegroundColor Yellow
}

# Step 4: Verify Setup
Write-Host "`n🔍 Step 4: Verifying setup..." -ForegroundColor Yellow
Start-Sleep -Seconds 5  # Wait for server restart

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✅ Server is running after setup!" -ForegroundColor Green
    Write-Host "Mode: $($response.mode)" -ForegroundColor Cyan
    Write-Host "Status: $($response.ok)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Server health check failed" -ForegroundColor Red
}

Write-Host "`n🎯 Setup completed!" -ForegroundColor Green
Write-Host "Open http://localhost:8080 to access GoConnect" -ForegroundColor Cyan
Write-Host "Default admin will be created on first access" -ForegroundColor Cyan
