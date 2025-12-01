# Active Context

## Current Focus
The project is in **Implementation Phase 1** - establishing the IPC foundation.
-   **Completed:** PRD, Architecture, Project Context, Memory Bank, gRPC Proto & Server Skeleton.
-   **Active:** Wiring CLI TUI to gRPC client.
-   **Next:** Full gRPC implementation + IPC security.

## Recent Changes
-   **2025-12-01:** Implemented gRPC server skeleton (`cli/internal/daemon/grpc_server.go`).
-   **2025-12-01:** Expanded `daemon.proto` with 6 services (Daemon, Network, Peer, Chat, Transfer, Settings).
-   **2025-12-01:** Integrated gRPC server into daemon service lifecycle.
-   **2025-12-01:** Added `make proto` Makefile target.

## Active Decisions
-   **gRPC IPC:** Server running on localhost:34101 (Windows) or Unix socket (Linux/macOS).
-   **Gradual Migration:** HTTP bridge remains alongside gRPC for backward compatibility.
-   **Service Stubs:** gRPC methods return `Unimplemented` for features not yet in engine.

## Current Implementation State
```
cli/internal/proto/
├── daemon.pb.go        # Generated message types
└── daemon_grpc.pb.go   # Generated gRPC server/client

cli/internal/daemon/
├── service.go          # Daemon service (HTTP + gRPC)
└── grpc_server.go      # NEW: gRPC service implementations
```

## Next Steps
1.  **Wire CLI TUI:** Create gRPC client wrapper in TUI for status/peers/chat.
2.  **Implement Engine Methods:** Add missing methods (JoinNetwork, CreateNetwork, etc.).
3.  **IPC Security:** Add token-based authentication to gRPC interceptors.
4.  **Desktop Integration:** Create Rust gRPC client for Tauri app.
