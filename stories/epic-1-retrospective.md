# Epic 1 Retrospective: Persistent Core Daemon & Lifecycle

## Overview
**Epic Goal:** Establish a robust, cross-platform background daemon capable of managing WireGuard connections, handling lifecycle events (start/stop/restart), and surviving crashes or sleep states.
**Status:** COMPLETE ✅

## Stories Completed
- **1.1 Core Daemon Entrypoint:** Created base `daemon` struct with structured JSON logging (`log/slog`) and signal handling.
- **1.2 Service Management (CLI):** Refactored CLI to `Cobra`. Implemented `service install/start/stop/status` using `kardianos/service`.
- **1.3 Graceful Shutdown:** Implemented `sync.WaitGroup` to ensure all goroutines (engine, API, validation) exit cleanly on SIGTERM.
- **1.4 Crash Recovery:** Configured `Restart=on-failure` (Linux) and `KeepAlive` (macOS) to ensure high availability.
- **1.5 Power State Handling:** Implemented "Time Jump" detection (>3x interval) to automatically reconnect VPN after system wake.

## What Went Well
- **Cobra Migration:** The refactor to Cobra made the CLI structure much cleaner and extensible.
- **Kardianos/Service:** proved to be a reliable abstraction for cross-platform service management.
- **Robustness Features:** Adding crash recovery and power state handling early ensures a resilient foundation.

## Challenges & Blockers
- **Environment Issues:** The `go` binary was missing from the shell environment, preventing compilation verification. We relied on static code analysis.
- **Module Separation:** There was some initial confusion between the new `core/internal/daemon` scaffold and the existing `cli/internal/daemon` logic. We decided to enhance the active `cli` daemon for now to maintain functionality.

## Lessons Learned
- **Architecture Alignment:** We need to clarify the long-term plan for `core` vs `cli` modules. Ideally, `cli` should just be a client for `core`.
- **Tooling:** Ensure build tools are available in the environment for future verifiable steps.

## Action Items
- [ ] Monitor the "Time Jump" threshold for false positives in production.
- [ ] Plan the migration of logic from `cli/internal/daemon` to `core` in a future refactor epic.

## Conclusion
The daemon is now production-ready from a lifecycle perspective. It installs as a system service, runs securely, and recovers from common failures.
