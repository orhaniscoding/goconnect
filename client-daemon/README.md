# GoConnect Client Daemon

The official client daemon for GoConnect, responsible for managing WireGuard interfaces, synchronizing configuration, and maintaining connectivity with the GoConnect Server.

## 🌟 Features

- **Automatic Configuration**: Fetches WireGuard configuration (Peers, Keys, Routes) from the server.
- **Cross-Platform Support**:
  - **Linux**: Uses `ip` and `wg` tools. Systemd service included.
  - **Windows**: Uses PowerShell and native Windows networking. Windows Service installer included.
  - **macOS**: Uses `ifconfig` and `route`. Launchd agent included.
- **Heartbeat System**: Reports device status, latency, and online/offline state to the server.
- **Resilient Networking**: Automatically re-applies configuration if the interface is tampered with.
- **Secure Storage**: Stores device credentials securely on the local filesystem.

## 🚀 Installation

### Linux

1.  **Build**:
    ```bash
    go build -o bin/goconnect-daemon ./cmd/daemon
    ```
2.  **Install Service**:
    ```bash
    sudo cp bin/goconnect-daemon /usr/local/bin/
    sudo make install-systemd
    ```
3.  **Start**:
    ```bash
    sudo systemctl start goconnect-daemon
    ```

### Windows

1.  **Build**:
    ```powershell
    go build -o bin/goconnect-daemon-windows-amd64.exe ./cmd/daemon
    ```
2.  **Install Service** (Run as Administrator):
    ```powershell
    ./service/windows/install.ps1
    ```
    This will install the daemon to `C:\Program Files\GoConnect` and start the Windows Service.

### macOS

1.  **Build**:
    ```bash
    go build -o bin/goconnect-daemon ./cmd/daemon
    ```
2.  **Install Agent**:
    ```bash
    make install-launchd
    ```

## 🔧 Configuration

The daemon uses a configuration file located at:
- **Linux**: `/etc/goconnect/config.json`
- **Windows**: `C:\ProgramData\GoConnect\config.json`
- **macOS**: `~/Library/Application Support/GoConnect/config.json`

### Environment Variables

- `GOCONNECT_SERVER_URL`: URL of the GoConnect Server (default: `http://localhost:8080`)
- `GOCONNECT_LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`)

## 🛠️ Development

Run the daemon in development mode (requires Admin/Root privileges for network operations):

```bash
# Linux/macOS
sudo go run ./cmd/daemon

# Windows (Admin PowerShell)
go run ./cmd/daemon
```
