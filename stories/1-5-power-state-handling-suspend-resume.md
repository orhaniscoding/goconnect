# Story 1.5: Power State Handling (Suspend/Resume)

## Story
As a Mobile User,
I want the VPN connection to automatically reconnect when my laptop wakes up from sleep,
So that I don't lose connectivity or have to manually restart the service.

## Acceptance Criteria
- [ ] **Given** the daemon is running and connected
- [ ] **When** the system suspends and resumes
- [ ] **Then** the daemon detects the wake event (e.g., via time jump or OS signal)
- [ ] **And** it triggers a `Reconnect()` attempt to re-establish WireGuard handshake.

## Tasks/Subtasks
- [ ] Implement "Time Jump" detection in the main run loop.
    - If `time.Since(lastTick) > threshold` (e.g., 15s when interval is 5s), assume sleep occurred.
- [ ] Implement `Reconnect()` logic in `cli/internal/daemon/service.go`.
- [ ] Add logging for "System resumed from sleep".

## Dev Notes
- **Strategy:** "Time Drift" detection is a robust, cross-platform way to detect resume without heavy dependencies (cgo/dbus).
- We already have a `ticker` in the `run()` loop.
- Threshold: 2x or 3x the health check interval.

## File List
- cli/internal/daemon/service.go

## Change Log
- Initial creation.
- Implemented time jump detection (>3x interval) to handle system resume in `cli/internal/daemon/service.go`.

## Status
done
