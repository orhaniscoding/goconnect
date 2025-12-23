# Story 2.2: Secure Local gRPC Socket Setup

## Story
As a Security Engineer,
I want the daemon to listen on a secure, local-only IPC channel (Unix Domain Socket or Named Pipe),
So that only the authorized CLI or desktop app running as the same user (or root/admin) can control it, preventing unauthorized local access.

## Acceptance Criteria
- [ ] **Given** the daemon configuration
- [ ] **When** starting the gRPC server
- [ ] **Then** it must use a Unix Domain Socket (Linux/macOS) at a secure path (e.g., `/var/run/goconnect.sock` or `~/.goconnect/daemon.sock`).
- [ ] **And** the socket file permissions must be restricted (e.g., `0600` - owner read/write only).
- [ ] **And** on Windows, it uses a secure Named Pipe with ACLs (if applicable, or restricted TCP loopback with token).
- [ ] **And** verify TCP listeners are NOT exposed unless explicitly configured for specific interfaces (defaults to IPC only).

## Tasks/Subtasks
- [ ] Analyze `cli/internal/daemon` for existing gRPC listener logic.
- [ ] Implement/Enforce UDS listener on Linux/macOS.
- [ ] Implement file permission setting (`chmod 0600`) immediately after socket creation.
- [ ] Update `cli` client connection logic to dial via UDS.

## Dev Notes
- `net.Listen("unix", path)`
- Remove potential existing `net.Listen("tcp", ...)` unless needed for specific bridge cases (Story 1.1 might have left some TCP).
- Ensure `os.Remove(path)` is called before listening to clean up stale sockets.

## File List
- cli/internal/daemon/grpc.go (Server)
- cli/internal/client/grpc.go (Client - presumably)
- cli/internal/config/config.go (Socket path definition)

## Change Log
- Initial creation.
- Added `EnableDesktopIPC` to config (defaults to false).
- Implemented stale socket cleanup and `0600` permission enforcement in `grpc_server.go`.
- Made TCP fallback listener conditional.

## Status
done
