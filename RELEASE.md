# Release Process

This project uses an automated release process powered by **Release Please** and **Goreleaser**.

## How it works

1.  **Conventional Commits**: All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.
    *   `feat: ...` -> Minor version bump (0.1.0 -> 0.2.0)
    *   `fix: ...` -> Patch version bump (0.1.0 -> 0.1.1)
    *   `feat!: ...` or `BREAKING CHANGE:` -> Major version bump (0.1.0 -> 1.0.0)

2.  **Release PR**:
    *   When changes are pushed to the `main` branch, the **Release Please** workflow runs.
    *   It analyzes the commits since the last release.
    *   It creates or updates a "Release PR" (e.g., `chore(main): release 1.0.0`).
    *   This PR contains the updated `CHANGELOG.md` and version bumps in `package.json`, `version.go`, etc.

3.  **Triggering a Release**:
    *   Review the Release PR.
    *   **Merge** the Release PR into `main`.
    *   **Release Please** will automatically:
        *   Create a new Git tag (e.g., `v1.0.0`).
        *   Create a GitHub Release with the changelog.

4.  **Build & Publish Artifacts**:
    *   The creation of the GitHub Release triggers the **Goreleaser** workflow.
    *   Goreleaser builds the binaries for:
        *   **Server**: Linux, Windows, macOS (amd64, arm64)
        *   **Client Daemon**: Linux, Windows, macOS (amd64, arm64)
    *   It uploads these artifacts (binaries, archives, deb/rpm packages) to the GitHub Release page.

## Manual Actions (If needed)

If the automation fails or you need to force a release:
1.  Ensure you are on `main` and up to date.
2.  Tag the commit manually: `git tag v1.0.1`
3.  Push the tag: `git push origin v1.0.1`
4.  This will trigger Goreleaser, but it won't generate the Changelog automatically.

## Supported Platforms

The release system automatically builds for:
*   **Linux**: amd64, arm64 (deb, rpm, tar.gz)
*   **Windows**: amd64, arm64 (zip)
*   **macOS**: amd64, arm64 (tar.gz)
