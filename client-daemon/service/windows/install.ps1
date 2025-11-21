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
    # Try looking in bin/ folder relative to script if not in same folder
    $SourceBin = Join-Path $PSScriptRoot "..\..\bin\goconnect-daemon-windows-amd64.exe"
    if (!(Test-Path $SourceBin)) {
        Write-Host "Error: Could not find $BinName in script directory or bin folder." -ForegroundColor Red
        exit 1
    }
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

# Create Service
if (!$Service) {
    Write-Host "Creating GoConnectDaemon service..."
    New-Service -Name "GoConnectDaemon" `
        -BinaryPathName "$TargetBin" `
        -DisplayName "GoConnect Daemon" `
        -Description "GoConnect VPN Client Daemon" `
        -StartupType Automatic
    
    # Set recovery options (restart on failure)
    sc.exe failure GoConnectDaemon reset= 86400 actions= restart/60000/restart/60000/restart/60000
}
else {
    Write-Host "Service already exists, updating binary..."
}

# Start Service
Write-Host "Starting service..."
Start-Service "GoConnectDaemon"
Write-Host "✅ GoConnect Daemon installed and started successfully!" -ForegroundColor Green
