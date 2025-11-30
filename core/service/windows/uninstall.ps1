# Check for Administrator privileges
if (!([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Please run this script as Administrator!" -ForegroundColor Red
    exit 1
}

$InstallDir = "C:\Program Files\GoConnect"
$ServiceName = "GoConnectServer"

# Stop service
$Service = Get-Service $ServiceName -ErrorAction SilentlyContinue
if ($Service) {
    if ($Service.Status -eq "Running") {
        Write-Host "Stopping service..."
        Stop-Service $ServiceName -Force
        Start-Sleep -Seconds 2
    }
    
    # Remove service
    Write-Host "Removing service..."
    sc.exe delete $ServiceName | Out-Null
    Write-Host "Service removed." -ForegroundColor Green
}

# Remove installation directory
if (Test-Path $InstallDir) {
    Write-Host "Removing installation directory..."
    Remove-Item -Path $InstallDir -Recurse -Force
    Write-Host "Installation directory removed." -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ GoConnect Server uninstalled successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Note: Config files in C:\ProgramData\GoConnect were preserved." -ForegroundColor Yellow
Write-Host "To remove: Remove-Item -Path 'C:\ProgramData\GoConnect' -Recurse -Force" -ForegroundColor Yellow
