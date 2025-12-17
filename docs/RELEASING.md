# Release Process

This document describes the release process for GoConnect, including versioning standards, asset naming conventions, and how to perform a release.

## Versioning Standard

We follow [Semantic Versioning 2.0.0](https://semver.org/):
- **Major** (X.y.z): Incompatible API changes
- **Minor** (x.Y.z): Backwards-compatible functionality
- **Patch** (x.y.Z): Backwards-compatible bug fixes

**Tag Format**: `vMAJOR.MINOR.PATCH` (e.g., `v0.1.0`, `v1.2.3`).

## Release Assets

A complete release includes the following assets:

### Server
Naming: `goconnect-server_<version>_<os>_<arch>.<ext>`
- Linux (amd64/arm64): `.tar.gz`
- macOS (amd64/arm64): `.tar.gz`
- Windows (amd64/arm64): `.zip`

### CLI
Naming: `goconnect-cli_<version>_<os>_<arch>.<ext>`
- Linux (amd64/arm64): `.tar.gz`
- macOS (amd64/arm64): `.tar.gz`
- Windows (amd64/arm64): `.zip`

### Desktop App
- **Linux**: `.deb`, `.AppImage`
- **macOS**: `.dmg` (Universal or x64/arm64)
- **Windows**: `.msi`, `.exe`

### Metadata
- `checksums.txt`: SHA256 checksums for all artifacts
- `latest.json`: Update manifest for Tauri auto-updater

## How to Release

### Prerequisites
- Clean git working tree
- `gh` CLI installed and authenticated
- Valid signing keys in place (for Tauri updater)

### Automated Release (Recommended)

1. **Run the Release Script**:
   ```bash
   ./scripts/release.sh v0.1.0
   ```
   This script will:
   - Run all tests
   - Create the annotated tag
   - Push the tag to GitHub

2. **Watch the Pipeline**:
   - Go to GitHub Actions tab
   - Wait for "Release Pipeline" to complete
   - Verify assets appear on the Release page

### Manual Release

1. **Tag**: `git tag -a v0.1.0 -m "Release v0.1.0"`
2. **Push**: `git push origin v0.1.0`

## Rollback / Emergency

If a bad release is pushed:

1. **Delete Remote Release**:
   ```bash
   gh release delete v0.1.0 -y
   ```
2. **Delete Tag**:
   ```bash
   git tag -d v0.1.0
   git push origin --delete v0.1.0
   ```
3. **Fix & Re-release**: Bump patch version (e.g., `v0.1.1`) and release again.
