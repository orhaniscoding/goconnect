# Tech Context

## Technology Stack

### Core & CLI (Go)
-   **Language:** Go 1.24+
-   **TUI Framework:** Bubbletea + Lipgloss
-   **IPC:** `grpc-go`
-   **Service:** `kardianos/service`
-   **Database:** `mattn/go-sqlite3` (CGo enabled)

### Desktop (Rust + TypeScript)
-   **Host:** Rust (Tauri 2.x)
-   **Frontend:** React + TypeScript + Vite
-   **Styling:** Tailwind CSS
-   **State:** TanStack Query v5
-   **IPC:** `tonic` (Rust gRPC client)

### Shared Infrastructure
-   **Protocol:** Protobuf v3
-   **Network:** WireGuard (userspace/kernel)
-   **Build:** Makefiles, `go.work`

## Development Environment
-   **OS Support:** Windows 10/11, macOS 12+, Linux (systemd).
-   **Requirements:**
    -   Go 1.24+
    -   Rust (latest stable)
    -   Node.js (LTS) + pnpm
    -   GCC/MinGW (for CGo)

## Technical Constraints
-   **Privileges:** Daemon requires Admin/Root for network interface management.
-   **Latency:** <50ms overhead target requires efficient native code (Go/Rust).
-   **Binary Size:** Keep binaries small for easy distribution.
-   **Offline:** Must handle network interruptions gracefully (auto-reconnect).
