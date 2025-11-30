# Check for Administrator privileges
if (!([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
  Write-Host "Please run this script as Administrator!" -ForegroundColor Red
  exit 1
}

$InstallDir = "C:\Program Files\GoConnect"
$BinName = "goconnect-server.exe"
$SourceBin = Join-Path $PSScriptRoot $BinName
$TargetBin = Join-Path $InstallDir $BinName

# Create installation directory
if (!(Test-Path $InstallDir)) {
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Write-Host "Created directory: $InstallDir" -ForegroundColor Green
}

# Check if source binary exists
if (!(Test-Path $SourceBin)) {
  Write-Host "Error: Could not find $BinName in the same directory as this script." -ForegroundColor Red
  Write-Host "Expected location: $SourceBin" -ForegroundColor Yellow
  exit 1
}

# Stop existing service
$Service = Get-Service "GoConnectServer" -ErrorAction SilentlyContinue
if ($Service) {
  Write-Host "Stopping existing service..."
  Stop-Service "GoConnectServer" -Force -ErrorAction SilentlyContinue
  Start-Sleep -Seconds 2
}

# Copy binary
Write-Host "Copying binary to $InstallDir..."
Copy-Item -Path $SourceBin -Destination $TargetBin -Force

# Create configuration directory and example config
$ConfigDir = "C:\ProgramData\GoConnect"
$ConfigFile = Join-Path $ConfigDir ".env"
$ExampleConfig = Join-Path $PSScriptRoot "config.example.env"

if (!(Test-Path $ConfigDir)) {
  New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
  Write-Host "Created config directory: $ConfigDir" -ForegroundColor Green
}

if (!(Test-Path $ConfigFile)) {
  if (Test-Path $ExampleConfig) {
    Copy-Item -Path $ExampleConfig -Destination $ConfigFile -Force
    Write-Host "Created example config: $ConfigFile" -ForegroundColor Green
    Write-Host "⚠️  IMPORTANT: Edit this file with your database and WireGuard settings!" -ForegroundColor Yellow
  }
  else {
    # Create minimal config if example doesn't exist
    @"
# GoConnect Server Configuration
# Copy to .env and configure before starting

SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goconnect
DB_USER=goconnect
DB_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_min_32_chars
WG_SERVER_ENDPOINT=vpn.example.com:51820
WG_SERVER_PUBKEY=your_wireguard_public_key_44_chars
"@ | Out-File -FilePath $ConfigFile -Encoding UTF8
    Write-Host "Created minimal config: $ConfigFile" -ForegroundColor Green
    Write-Host "⚠️  IMPORTANT: Edit this file before starting the service!" -ForegroundColor Yellow
  }
}
else {
  Write-Host "Config file already exists: $ConfigFile" -ForegroundColor Cyan
}

# Create Service
if (!$Service) {
  Write-Host "Creating GoConnectServer service..."
  New-Service -Name "GoConnectServer" `
    -BinaryPathName "`"$TargetBin`"" `
    -DisplayName "GoConnect Server" `
    -Description "GoConnect VPN Management Server" `
    -StartupType Manual
    
  # Set recovery options
  sc.exe failure GoConnectServer reset= 86400 actions= restart/60000/restart/60000/restart/60000 | Out-Null
}
else {
  Write-Host "Service already exists, updating binary..."
}

# Test binary
Write-Host "Testing binary..."
$TestResult = & $TargetBin --version 2>&1
if ($LASTEXITCODE -ne 0) {
  Write-Host "Warning: Binary test failed. Service may not start correctly." -ForegroundColor Yellow
}

# Start Service
Write-Host "Starting service..."
try {
  Start-Service "GoConnectServer" -ErrorAction Stop
  Write-Host "✅ GoConnect Server installed and started successfully!" -ForegroundColor Green
  Write-Host ""
  Write-Host "Service Status:" -ForegroundColor Cyan
  Get-Service "GoConnectServer" | Format-Table -AutoSize
}
catch {
  Write-Host "⚠️  Service installed but failed to start." -ForegroundColor Yellow
  Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
  Write-Host ""
  Write-Host "REQUIRED: Configure before starting:" -ForegroundColor Yellow
  Write-Host "1. Edit: $ConfigFile" -ForegroundColor Yellow
  Write-Host "2. Set database credentials (DB_*)" -ForegroundColor Yellow
  Write-Host "3. Set JWT_SECRET (min 32 chars)" -ForegroundColor Yellow
  Write-Host "4. Set WireGuard settings (WG_*)" -ForegroundColor Yellow
  Write-Host "5. Start service: Start-Service GoConnectServer" -ForegroundColor Yellow
  Write-Host ""
  Write-Host "See config.example.env for all available options." -ForegroundColor Cyan
  exit 0
}
