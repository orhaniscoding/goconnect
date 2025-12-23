# Story 1.2: Service Management Actions (CLI)

## Story
As a System Administrator,
I want to install and manage the daemon as a system service using CLI commands,
So that it runs automatically on boot and recovers from crashes.

## Acceptance Criteria
- [ ] **Given** the CLI is installed
- [ ] **When** I run `goconnect service install` (with sudo/admin)
- [ ] **Then** the daemon is registered as a system service (systemd/launchd/SCM)
- [ ] **When** I run `goconnect service start`
- [ ] **Then** the service starts and persists
- [ ] **When** I run `goconnect service status`
- [ ] **Then** it reports the current service status

## Tasks/Subtasks
- [ ] Initialize `cli` module if missing
- [ ] Implement `service` command structure (Cobra)
- [ ] integrate `kardianos/service` library
- [ ] Implement `install` action
- [ ] Implement `start` action
- [ ] Implement `status` action
- [ ] Implement `stop`/`uninstall` actions

## Dev Notes
- Use `github.com/kardianos/service` for cross-platform service abstraction.
- The CLI needs to locate the daemon binary. For now, assume it's in the same directory or standard path.
- **Privilege Check:** These commands typically require root/admin.

## Dev Agent Record
- **Plan:**
  1. Check `cli` structure.
  2. Add `cobra` and `kardianos/service` dependencies.
  3. Implement commands.

## File List
- cli/cmd/root.go (if missing)
- cli/cmd/service.go
- cli/internal/service/manager.go

## Change Log
- Initial creation.
- Code Review: Fixed service binary path, added run command, improved error handling.

## Status
done
