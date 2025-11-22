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
   - Archives (`.tar.gz`, `.zip`) are created.
   - Linux packages (`.deb`, `.rpm`) are generated.
   - Checksums (`checksums.txt`) are calculated.
4. **Publishing:**
   - A new **GitHub Release** is created.
   - All artifacts (binaries, packages, checksums) are uploaded.
   - Release notes are posted.

## 5. Post-Release Verification
- Verify the release appears on the [GitHub Releases Page](https://github.com/orhaniscoding/goconnect/releases).
- Ensure the Docker image (if configured) is pushed to the registry.
