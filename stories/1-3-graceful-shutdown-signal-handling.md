# Story 1.3: Graceful Shutdown & Signal Handling

## Story
As a System Administrator,
I want the daemon to shut down cleanly when receiving system signals (SIGINT, SIGTERM) or service stop commands,
So that data corruption is prevented and connections are closed properly.

## Acceptance Criteria
- [ ] **Given** the daemon is running (foreground or service)
- [ ] **When** it receives SIGINT or SIGTERM
- [ ] **Then** it logs "Received signal..."
- [ ] **And** it cancels the main context
- [ ] **And** it waits for active goroutines (e.g., WireGuard interface close) to finish (with timeout)
- [ ] **And** it logs "Daemon exited cleanly" before process exit
- [ ] **Given** the daemon is running as a service
- [ ] **When** `goconnect service stop` is run
- [ ] **Then** the `Stop()` method is called and performs the same cleanup

## Tasks/Subtasks
- [ ] Analyze `cli` vs `core` daemon overlap.
- [ ] Integrate `kardianos/service` Logger for system logs.
- [ ] Implement `service.Interface` methods correctly in `cli/internal/svc/manager.go`.
- [ ] Refactor `cli/cmd/legacy.go` (run command) to share cleanup logic.
- [ ] Ensure `context.Context` is propagated to key components.

## Dev Notes
- `kardianos/service` behaves differently on interactive vs service mode.
- In `Run()` (foreground), `service.Run()` blocks.
- We need a `Program` struct that implements `Start` (non-blocking) and `Stop`.
- **Constraint:** We must unify the "Core Daemon" (created in 1.1) with the "CLI Daemon" (refactored in 1.2). The CLI should import/use the Core logic.

## Dev Agent Record
- **Plan:**
  1. Check `cli/internal/daemon`.
  2. If it's legacy, replace/wrap with `core/internal/daemon`.
  3. Wire up signals.

## File List
- cli/internal/svc/manager.go
- cli/cmd/legacy.go
- core/internal/daemon/daemon.go

## Change Log
- Initial creation.
- Implemented graceful shutdown with `sync.WaitGroup`.

## Status
done
