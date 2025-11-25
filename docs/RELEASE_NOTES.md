# Release Notes v2.8.0

## 🚀 New Features

### Server Infrastructure
- **Graceful Shutdown:** HTTP server, background workers (device offline detection, WireGuard sync, metrics), WebSocket hub, and audit pipelines now shut down cleanly on SIGINT/SIGTERM.
- **WireGuard Profile DNS:** JSON WireGuard profile responses now include DNS servers from network configuration instead of empty arrays.

## 🔧 Technical Details
- Root context wired through all background goroutines for coordinated cancellation.
- Async auditor and SQLite auditor properly close on shutdown.
- Redis client connection released on shutdown.
- HTTP server uses 20-second graceful drain timeout.

## 📦 Upgrade Instructions
No database migrations required. Existing deployments benefit automatically from graceful shutdown behavior.
