# System Patterns

## System Architecture
GoConnect uses a **Daemon-Client Architecture** within a **Polyglot Monorepo**.

### Components
1.  **Core Daemon (Go):**
    -   **Role:** The "Brain". Handles networking, state, and logic.
    -   **Responsibilities:** WireGuard tunnel management, NAT traversal, SQLite storage, gRPC server.
    -   **Characteristics:** Headless, privileged, long-running.
2.  **CLI Client (Go):**
    -   **Role:** Terminal Interface.
    -   **Responsibilities:** User interaction via TUI, communicating with Daemon.
    -   **Characteristics:** Ephemeral, unprivileged.
3.  **Desktop App (Rust/Tauri + React):**
    -   **Role:** Graphical Interface.
    -   **Responsibilities:** Rich UI, system tray, notifications.
    -   **Characteristics:** Long-running UI, unprivileged.

### Communication
-   **Protocol:** **gRPC** over Unix Domain Sockets (macOS/Linux) or Named Pipes (Windows).
-   **Contract:** **Protobuf v3** (`core/proto`) is the single source of truth.
-   **Security:** Token-based authentication for IPC.

## Design Patterns

### Data Architecture
-   **Single Source of Truth:** The Daemon's **SQLite** database holds the authoritative state.
-   **Optimistic UI:** Clients (CLI/Desktop) use optimistic updates (via TanStack Query) for responsiveness, confirmed by Daemon events.
-   **Configuration:** **YAML** file for user settings, managed by the Daemon.

### Security Patterns
-   **Zero-Trust IPC:** All local connections must be authenticated.
-   **Identity:** **WireGuard Public Keys** (Ed25519) identify peers.
-   **Local Keys:** Private keys never leave the device.

### Implementation Patterns
-   **Naming:**
    -   Go: `PascalCase` (Exported), `camelCase` (Internal).
    -   Rust: `snake_case` (Funcs), `PascalCase` (Types).
    -   Proto: `PascalCase` (Messages/RPCs), `snake_case` (Fields).
-   **Error Handling:**
    -   Go: Return `error` last, wrap with context.
    -   Rust: `Result<T, E>` with `thiserror`/`anyhow`.
    -   gRPC: Standard status codes + `ErrorDetail` metadata.
-   **Events:** `NounVerbPast` naming (e.g., `PeerConnected`).

## Project Structure
```text
goconnect/
├── core/       # Go: Daemon, Logic, Proto, SQLite
├── cli/        # Go: Bubbletea TUI
└── desktop/    # Rust/Tauri: React Frontend
```
