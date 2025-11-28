# Plan: Client Daemon - Unified Service & Keyring Integration

## 1. Executive Summary
To achieve the "Zero-Config" goal for the client, the daemon must run seamlessly in the background as a system service. It must also securely store authentication tokens without requiring user intervention in config files.
This plan focuses on refactoring the `client-daemon` to use a unified service library and integrating OS-level keyring storage.

## 2. Goals
1.  **Unified Service Command:** `goconnect-daemon install`, `start`, `stop`, `uninstall` commands that work across Windows, Linux, and macOS using `kardianos/service`.
2.  **Secure Storage:** Replace plain text token storage in `config.yaml` with OS Keyring integration (`zalando/go-keyring`).
3.  **Auto-Connect:** The service should automatically connect to the last used network on boot.

## 3. Technical Architecture

### 3.1. Service Management (`kardianos/service`)
Refactor `client-daemon/cmd/daemon/main.go` to implement the `service.Interface`.
*   **Windows:** Uses Service Control Manager (SCM).
*   **Linux:** Uses Systemd (preferred) or SysV.
*   **macOS:** Uses Launchd.

### 3.2. Token Storage Strategy
*   **Current:** Tokens stored in `config.yaml` (insecure).
*   **New:**
    *   `config.yaml` stores non-sensitive settings (server URL, log level).
    *   `access_token` and `refresh_token` are stored in OS Keyring.
    *   **Service:** `GoConnect Daemon`
    *   **User:** The system user (or specific user context).

### 3.3. Commands
The binary will support subcommands:
*   `run` (default): Run in foreground (or as service if detecting service context).
*   `install`: Register the service.
*   `uninstall`: Unregister.
*   `start`: Start the service.
*   `stop`: Stop the service.
*   `login`: CLI login flow (optional, primarily for headless).

## 4. Implementation Steps

### Phase 1: Service Refactor
- [ ] Add `github.com/kardianos/service` dependency.
- [ ] Create `internal/daemon/service.go` to wrap the main logic in a struct implementing `service.Interface`.
- [ ] Update `main.go` to handle flags/commands for service control.

### Phase 2: Keyring Integration
- [ ] Add `github.com/zalando/go-keyring` dependency.
- [ ] Create `internal/storage/keyring.go`.
- [ ] Modify `internal/config` to read/write tokens via Keyring instead of YAML.
- [ ] Add fallback for headless Linux servers (encrypted file if no keyring available).

### Phase 3: Auto-Connect Logic
- [ ] Ensure the `Start()` method triggers the connection logic immediately if tokens exist.
- [ ] Implement a retry loop for network availability on boot.

## 5. Success Criteria
*   Running `sudo ./goconnect-daemon install` on Linux creates a systemd service.
*   Running `goconnect-daemon.exe install` on Windows creates a Windows Service.
*   Tokens are NOT visible in `config.yaml`.
*   Rebooting the machine automatically reconnects the VPN.
