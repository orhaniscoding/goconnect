# Release Report: v2.4.0

**Date:** 2025-11-22
**Version:** 2.4.0
**Author:** orhaniscoding

## Summary
This release introduces comprehensive search and filtering capabilities to the Admin Dashboard, allowing administrators to efficiently manage Users, Tenants, Networks, and Devices. It also includes the implementation of device management within the dashboard.

## Build Status
| Component | Platform | Arch  | Status    | Artifact                             |
| --------- | -------- | ----- | --------- | ------------------------------------ |
| Server    | Linux    | amd64 | ✅ Success | `goconnect-server-linux-amd64`       |
| Server    | Linux    | arm64 | ✅ Success | `goconnect-server-linux-arm64`       |
| Server    | macOS    | amd64 | ✅ Success | `goconnect-server-darwin-amd64`      |
| Server    | macOS    | arm64 | ✅ Success | `goconnect-server-darwin-arm64`      |
| Server    | Windows  | amd64 | ✅ Success | `goconnect-server-windows-amd64.exe` |
| Server    | Windows  | arm64 | ✅ Success | `goconnect-server-windows-arm64.exe` |
| Daemon    | Linux    | amd64 | ✅ Success | `goconnect-daemon-linux-amd64`       |
| Daemon    | Linux    | arm64 | ✅ Success | `goconnect-daemon-linux-arm64`       |
| Daemon    | macOS    | amd64 | ✅ Success | `goconnect-daemon-darwin-amd64`      |
| Daemon    | macOS    | arm64 | ✅ Success | `goconnect-daemon-darwin-arm64`      |
| Daemon    | Windows  | amd64 | ✅ Success | `goconnect-daemon-windows-amd64.exe` |
| Daemon    | Windows  | arm64 | ✅ Success | `goconnect-daemon-windows-arm64.exe` |

## Packaging
All binaries have been packaged into platform-specific archives in `dist/v2.4.0/`.

## Documentation
- Updated `CHANGELOG.md` with new features.
- Updated `README.md` with version bump.
- Updated `server/openapi/openapi.yaml` with search parameters.
- Updated `docs/API_EXAMPLES.http` with search examples.
- Created `dist/v2.4.0/RELEASE_NOTES.md`.

## Security
- No new dependencies introduced.
- Search inputs are sanitized and validated in the backend repositories.
- SQL injection prevention is handled via parameterized queries (or ORM equivalent in current in-memory/mock implementation).

## Next Steps
1. Upload artifacts to GitHub Release.
2. Deploy server update.
3. Notify users of client update.
