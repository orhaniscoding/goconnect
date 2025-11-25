# Release Process (CI/CD Driven)

This document outlines the automated release process for GoConnect, managed via GitHub Actions, Release Please, and GoReleaser.

## 1. Development & Commits
All changes must follow **Conventional Commits** to ensure the release automation works correctly.

- `feat: ...` -> Triggers a **Minor** version bump (v1.1.0 -> v1.2.0)
- `fix: ...` -> Triggers a **Patch** version bump (v1.1.0 -> v1.1.1)
- `feat!: ...` or `BREAKING CHANGE:` -> Triggers a **Major** version bump (v1.0.0 -> v2.0.0)

## 2. Release Preparation (Automated PR)
The **Release Please** GitHub Action monitors the `main` branch. When it detects new commits:
1. It creates or updates a Pull Request titled `chore(main): release vX.Y.Z`.
2. This PR includes:
   - Updated `CHANGELOG.md`.
   - Updated version numbers in `package.json`, `version.go`, etc.

## 3. Triggering the Release
To publish a new release:
1. Review the **Release PR** created by the bot.
2. **Merge** the PR into `main`.

## 4. Build & Publish (Automated Pipeline)
Once the Release PR is merged, the **GoReleaser** workflow is triggered automatically:
1. **Tagging:** A new git tag (e.g., `v2.5.0`) is created.
2. **Building:** Binaries are cross-compiled for:
   - Windows (amd64, arm64)
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
3. **Packaging:**
   - Archives (`.tar.gz` for Unix, `.zip` for Windows) are created.
   - Linux packages (`.deb`, `.rpm`) are generated with systemd service.
   - Docker images are built and pushed to GitHub Container Registry.
   - Checksums (`checksums.txt`) are calculated.
4. **Publishing:**
   - A new **GitHub Release** is created with detailed release notes.
   - All artifacts (binaries, packages, checksums) are uploaded.
   - Release notes include installation instructions per platform.

## 5. Release Artifacts

Each release includes:

### 📦 Portable Archives
| Platform | Architecture | File                                   |
| -------- | ------------ | -------------------------------------- |
| Linux    | amd64        | `goconnect-daemon_linux_amd64.tar.gz`  |
| Linux    | arm64        | `goconnect-daemon_linux_arm64.tar.gz`  |
| macOS    | amd64        | `goconnect-daemon_darwin_amd64.tar.gz` |
| macOS    | arm64        | `goconnect-daemon_darwin_arm64.tar.gz` |
| Windows  | amd64        | `goconnect-daemon_windows_amd64.zip`   |
| Windows  | arm64        | `goconnect-daemon_windows_arm64.zip`   |

### 🐧 Linux Packages (Easy Install)
| Format        | Architecture | File                         |
| ------------- | ------------ | ---------------------------- |
| Debian/Ubuntu | amd64        | `goconnect-daemon_amd64.deb` |
| Debian/Ubuntu | arm64        | `goconnect-daemon_arm64.deb` |
| RHEL/Fedora   | amd64        | `goconnect-daemon_amd64.rpm` |
| RHEL/Fedora   | arm64        | `goconnect-daemon_arm64.rpm` |

### 🐳 Docker Images
| Image                                           | Description |
| ----------------------------------------------- | ----------- |
| `ghcr.io/orhaniscoding/goconnect-server:latest` | API Server  |
| `ghcr.io/orhaniscoding/goconnect-web:latest`    | Web UI      |

### 🔐 Verification
| File            | Description                    |
| --------------- | ------------------------------ |
| `checksums.txt` | SHA256 checksums for all files |

## 6. Installation Methods

### Quick Install (Linux)
```bash
# Debian/Ubuntu
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-daemon_amd64.deb
sudo dpkg -i goconnect-daemon_amd64.deb

# RHEL/Fedora
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-daemon_amd64.rpm
sudo rpm -i goconnect-daemon_amd64.rpm
```

### Quick Install (macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/orhaniscoding/goconnect/main/client-daemon/service/macos/install.sh | sudo bash
```

### Quick Install (Windows)
```powershell
irm https://raw.githubusercontent.com/orhaniscoding/goconnect/main/client-daemon/service/windows/install.ps1 | iex
```

### Docker Compose (Full Stack)
```bash
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect
docker-compose up -d
```

## 7. Post-Release Verification
- Verify the release appears on the [GitHub Releases Page](https://github.com/orhaniscoding/goconnect/releases).
- Ensure Docker images are pushed to GitHub Container Registry.
- Test installation scripts work correctly on target platforms.
- Verify checksums match downloaded files.
