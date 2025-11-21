$Version = "2.4.0"
$DistDir = "dist"
$ReleaseDir = "$DistDir/v2.4.0"
$SourceDir = "."

New-Item -ItemType Directory -Force -Path $ReleaseDir

$Targets = @(
  @{ Os = "linux"; Arch = "amd64"; Ext = "tar.gz" },
  @{ Os = "linux"; Arch = "arm64"; Ext = "tar.gz" },
  @{ Os = "darwin"; Arch = "amd64"; Ext = "tar.gz" },
  @{ Os = "darwin"; Arch = "arm64"; Ext = "tar.gz" },
  @{ Os = "windows"; Arch = "amd64"; Ext = "zip" },
  @{ Os = "windows"; Arch = "arm64"; Ext = "zip" }
)

foreach ($Target in $Targets) {
  $Os = $Target.Os
  $Arch = $Target.Arch
  $Ext = $Target.Ext
    
  $TempDir = "$DistDir/temp_${Os}_${Arch}"
  New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
    
  $ServerBin = "goconnect-server-${Os}-${Arch}"
  $DaemonBin = "goconnect-daemon-${Os}-${Arch}"
    
  if ($Os -eq "windows") {
    $ServerBin += ".exe"
    $DaemonBin += ".exe"
    $ExtBin = ".exe"
  }
  else {
    $ExtBin = ""
  }
    
  # Copy Binaries
  Copy-Item "$DistDir/$ServerBin" -Destination "$TempDir/goconnect-server$ExtBin"
  Copy-Item "$DistDir/$DaemonBin" -Destination "$TempDir/goconnect-daemon$ExtBin"
    
  # Copy Docs
  Copy-Item "$SourceDir/README.md" -Destination $TempDir
  Copy-Item "$SourceDir/LICENSE" -Destination $TempDir
    
  $ArchiveName = "goconnect_${Version}_${Os}_${Arch}.$Ext"
  $ArchivePath = "$ReleaseDir/$ArchiveName"
    
  Write-Host "Creating $ArchiveName..."
    
  if ($Ext -eq "zip") {
    Compress-Archive -Path "$TempDir/*" -DestinationPath $ArchivePath -Force
  }
  else {
    # PowerShell doesn't have built-in tar.gz support in older versions, but tar might be available
    # If tar is available (Windows 10+ usually has it)
    tar -czf $ArchivePath -C $TempDir .
  }
    
  Remove-Item -Recurse -Force $TempDir
}

Write-Host "Packaging complete."
