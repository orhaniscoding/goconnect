# Check for Administrator privileges
if (!([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Please run this script as Administrator!" -ForegroundColor Red
    exit 1
}

$InstallDir = "C:\Program Files\GoConnect"
$BinName = "goconnect-daemon.exe"
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

# Stop existing service if running
$Service = Get-Service "GoConnectDaemon" -ErrorAction SilentlyContinue
if ($Service) {
    Write-Host "Stopping existing service..."
    Stop-Service "GoConnectDaemon" -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

# Copy binary
Write-Host "Copying binary to $InstallDir..."
Copy-Item -Path $SourceBin -Destination $TargetBin -Force

# Create configuration directory and example config
$ConfigDir = "C:\ProgramData\GoConnect"
$ConfigFile = Join-Path $ConfigDir "config.yaml"
$ExampleConfig = Join-Path $PSScriptRoot "config.example.yaml"

if (!(Test-Path $ConfigDir)) {
    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
    Write-Host "Created config directory: $ConfigDir" -ForegroundColor Green
}

if (!(Test-Path $ConfigFile)) {
    if (Test-Path $ExampleConfig) {
        Copy-Item -Path $ExampleConfig -Destination $ConfigFile -Force
        Write-Host "Created example config: $ConfigFile" -ForegroundColor Green
        Write-Host "⚠️  IMPORTANT: Edit this file with your server URL before starting the service!" -ForegroundColor Yellow
    }
    else {
        # Create minimal config if example doesn't exist
        @"
# GoConnect Daemon Configuration
# REQUIRED: Set your server URL
server_url: "https://vpn.example.com:8080"

# Optional settings (defaults shown)
local_port: 12345
log_level: "info"
interface_name: "wg0"
"@ | Out-File -FilePath $ConfigFile -Encoding UTF8
        Write-Host "Created minimal config: $ConfigFile" -ForegroundColor Green
        Write-Host "⚠️  IMPORTANT: Edit this file with your server URL before starting the service!" -ForegroundColor Yellow
    }
}
else {
    Write-Host "Config file already exists: $ConfigFile" -ForegroundColor Cyan
}

# Create Service
if (!$Service) {
    Write-Host "Creating GoConnectDaemon service..."
    New-Service -Name "GoConnectDaemon" `
        -BinaryPathName "`"$TargetBin`"" `
        -DisplayName "GoConnect Daemon" `
        -Description "GoConnect VPN Client Daemon" `
        -StartupType Manual
    
    # Set recovery options (restart on failure)
    sc.exe failure GoConnectDaemon reset= 86400 actions= restart/60000/restart/60000/restart/60000 | Out-Null
}
else {
    Write-Host "Service already exists, updating binary..."
}

# Test if binary is executable
Write-Host "Testing binary..."
$TestResult = & $TargetBin --version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Warning: Binary test failed. Service may not start correctly." -ForegroundColor Yellow
    Write-Host "Error: $TestResult" -ForegroundColor Yellow
}

# Start Service
Write-Host "Starting service..."
try {
    Start-Service "GoConnectDaemon" -ErrorAction Stop
    Write-Host "✅ GoConnect Daemon installed and started successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Service Status:" -ForegroundColor Cyan
    Get-Service "GoConnectDaemon" | Format-Table -AutoSize
}
catch {
    Write-Host "⚠️  Service installed but failed to start." -ForegroundColor Yellow
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "REQUIRED: Configure the daemon before starting:" -ForegroundColor Yellow
    Write-Host "1. Edit: $ConfigFile" -ForegroundColor Yellow
    Write-Host "2. Set your server_url" -ForegroundColor Yellow
    Write-Host "3. Start service: Start-Service GoConnectDaemon" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "See config.example.yaml for all available options." -ForegroundColor Cyan
    exit 0
}
