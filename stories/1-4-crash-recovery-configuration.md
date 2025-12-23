# Story 1.4: Crash Recovery Configuration

## Story
As a DevOps Engineer,
I want the daemon service to automatically restart if it crashes,
So that network connectivity is restored without manual intervention.

## Acceptance Criteria
- [ ] **Given** the daemon service is installed
- [ ] **When** the daemon process crashes or is killed
- [ ] **Then** the service manager (systemd, etc.) automatically restarts it
- [ ] **And** restart delays/limits are configured (e.g., "Restart=on-failure", 5s delay)

## Tasks/Subtasks
- [ ] Analyze `kardianos/service` support for restart options.
- [ ] Update `cli/internal/svc/manager.go` `service.Config` with restart options.
- [ ] Verify generated service files (if possible) or configuration.

## Dev Notes
- `kardianos/service` exposes platform-specific options via `Config.Option`.
- **Linux (systemd):** `Restart=on-failure`, `RestartSec=5s`.
- **macOS (launchd):** `KeepAlive=true`, `RunAtLoad=true` (usually default).
- **Windows:** `Recovery` actions.

## File List
- cli/internal/svc/manager.go

## Change Log
- Initial creation.
- Implemented `Restart=on-failure` (Linux) and `KeepAlive` (macOS) in `cli/internal/svc/manager.go`.

## Status
done
