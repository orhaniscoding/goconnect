#
# Version bump script for GoConnect (PowerShell)
# Usage: .\scripts\bump-version.ps1 <new-version>
# Example: .\scripts\bump-version.ps1 3.1.0
#
# This script updates version numbers across all project files:
# - desktop/package.json
# - desktop/src-tauri/tauri.conf.json
# - desktop/src-tauri/Cargo.toml
#
# After running this script, commit the changes and create a tag:
#   git add -A
#   git commit -m "chore: bump version to v$NewVersion"
#   git tag v$NewVersion
#   git push origin main --tags
#

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$NewVersion
)

$ErrorActionPreference = "Stop"

# Get project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

# Validate version format (semver)
if ($NewVersion -notmatch '^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$') {
    Write-Host "❌ Invalid version format: $NewVersion" -ForegroundColor Red
    Write-Host "ℹ️  Version must be in semver format: X.Y.Z or X.Y.Z-prerelease" -ForegroundColor Blue
    exit 1
}

Write-Host "ℹ️  Bumping version to: $NewVersion" -ForegroundColor Blue
Write-Host ""

# Paths
$PackageJsonPath = Join-Path $ProjectRoot "desktop\package.json"
$TauriConfPath = Join-Path $ProjectRoot "desktop\src-tauri\tauri.conf.json"
$CargoTomlPath = Join-Path $ProjectRoot "desktop\src-tauri\Cargo.toml"

# Get current versions
$PackageJson = Get-Content $PackageJsonPath | ConvertFrom-Json
$TauriConf = Get-Content $TauriConfPath | ConvertFrom-Json
$CargoToml = Get-Content $CargoTomlPath -Raw
$CargoVersionMatch = [regex]::Match($CargoToml, '^version = "([^"]+)"', [System.Text.RegularExpressions.RegexOptions]::Multiline)

Write-Host "ℹ️  Current versions:" -ForegroundColor Blue
Write-Host "  - package.json:      $($PackageJson.version)"
Write-Host "  - tauri.conf.json:   $($TauriConf.version)"
Write-Host "  - Cargo.toml:        $($CargoVersionMatch.Groups[1].Value)"
Write-Host ""

# Update package.json
Write-Host "ℹ️  Updating desktop/package.json..." -ForegroundColor Blue
$PackageJson.version = $NewVersion
$PackageJson | ConvertTo-Json -Depth 100 | Set-Content $PackageJsonPath -Encoding UTF8
Write-Host "✅ Updated package.json" -ForegroundColor Green

# Update tauri.conf.json
Write-Host "ℹ️  Updating desktop/src-tauri/tauri.conf.json..." -ForegroundColor Blue
$TauriConf.version = $NewVersion
$TauriConf | ConvertTo-Json -Depth 100 | Set-Content $TauriConfPath -Encoding UTF8
Write-Host "✅ Updated tauri.conf.json" -ForegroundColor Green

# Update Cargo.toml
Write-Host "ℹ️  Updating desktop/src-tauri/Cargo.toml..." -ForegroundColor Blue
$CargoToml = $CargoToml -replace '^version = "[^"]+"', "version = `"$NewVersion`""
Set-Content $CargoTomlPath -Value $CargoToml -Encoding UTF8 -NoNewline
Write-Host "✅ Updated Cargo.toml" -ForegroundColor Green

Write-Host ""
Write-Host "✅ All versions updated to: $NewVersion" -ForegroundColor Green
Write-Host ""

# Verify updates
Write-Host "ℹ️  Verifying updates..." -ForegroundColor Blue
$UpdatedPackage = (Get-Content $PackageJsonPath | ConvertFrom-Json).version
$UpdatedTauri = (Get-Content $TauriConfPath | ConvertFrom-Json).version
$UpdatedCargoContent = Get-Content $CargoTomlPath -Raw
$UpdatedCargoMatch = [regex]::Match($UpdatedCargoContent, '^version = "([^"]+)"', [System.Text.RegularExpressions.RegexOptions]::Multiline)
$UpdatedCargo = $UpdatedCargoMatch.Groups[1].Value

if ($UpdatedPackage -eq $NewVersion -and $UpdatedTauri -eq $NewVersion -and $UpdatedCargo -eq $NewVersion) {
    Write-Host "✅ All versions verified!" -ForegroundColor Green
} else {
    Write-Host "❌ Version mismatch detected!" -ForegroundColor Red
    Write-Host "  - package.json:      $UpdatedPackage"
    Write-Host "  - tauri.conf.json:   $UpdatedTauri"
    Write-Host "  - Cargo.toml:        $UpdatedCargo"
    exit 1
}

Write-Host ""
Write-Host "ℹ️  Next steps:" -ForegroundColor Blue
Write-Host "  1. Review changes: git diff"
Write-Host "  2. Commit: git add -A; git commit -m 'chore: bump version to v$NewVersion'"
Write-Host "  3. Tag: git tag v$NewVersion"
Write-Host "  4. Push: git push origin main --tags"
Write-Host ""
Write-Host "✅ Done!" -ForegroundColor Green
