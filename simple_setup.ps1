# Simple GoConnect Setup
Write-Host "Starting GoConnect setup..." -ForegroundColor Green

# Database config
$dbBody = @{
    config = @{
        server = @{host = "0.0.0.0"; port = "8080"}
        database = @{backend = "postgres"; host = "postgres"; port = "5432"; user = "postgres"; password = "postgres"; dbname = "goconnect"}
    }
}

try {
    Invoke-RestMethod -Uri "http://localhost:8080/setup" -Method POST -Body ($dbBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Database configured" -ForegroundColor Green
} catch {
    Write-Host "Database failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Full config
$fullBody = @{
    config = @{
        server = @{host = "0.0.0.0"; port = "8080"}
        database = @{backend = "postgres"; host = "postgres"; port = "5432"; user = "postgres"; password = "postgres"; dbname = "goconnect"}
        jwt = @{secret = "ssssssssssssssssssssssssssssssss"; access_token_ttl = "1h"; refresh_token_ttl = "24h"}
        wireguard = @{server_endpoint = "localhost:51820"; server_pubkey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}
    }
    restart = $true
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/setup" -Method POST -Body ($fullBody | ConvertTo-Json -Depth 10) -ContentType "application/json"
    Write-Host "Setup completed! Server restarting..." -ForegroundColor Green
} catch {
    Write-Host "Setup failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Check http://localhost:8080 after restart" -ForegroundColor Cyan
