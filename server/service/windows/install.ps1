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
  Write-Host "The server requires configuration before it can run." -ForegroundColor Yellow
  Write-Host "Please configure at: C:\ProgramData\GoConnect\config.yaml" -ForegroundColor Yellow
  Write-Host "Then start with: Start-Service GoConnectServer" -ForegroundColor Yellow
  exit 0
}
