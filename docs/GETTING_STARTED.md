# Getting Started with GoConnect

GoConnect is designed to be up and running in minutes. This guide covers the "Personal Edition" setup, which uses an embedded database (SQLite) and requires no external dependencies.

For enterprise or production deployments (Docker, PostgreSQL), see [Deployment Guide](DEPLOYMENT.md).

## 1. Download & Run

### Linux / macOS
1.  **Download** the latest binary for your platform from the [Releases Page](https://github.com/orhaniscoding/goconnect/releases).
2.  **Make executable**: `chmod +x goconnect-server`
3.  **Run**: `./goconnect-server`

### Windows
1.  **Download** the `goconnect-server.exe` from the [Releases Page](https://github.com/orhaniscoding/goconnect/releases).
2.  **Double-click** to run.

## 2. Setup Wizard

On the first run, GoConnect will detect that it hasn't been configured and will start in **Setup Mode**.

1.  Open your browser and navigate to `http://localhost:8080`.
2.  You will be redirected to the **Setup Wizard**.
3.  **Mode Selection**: Choose **"Personal"** (uses built-in SQLite database).
4.  **Admin Account**: Create your primary administrator account.
5.  **Network Settings**: The wizard will generate secure keys for WireGuard automatically.
6.  **Finish**: Click "Save & Restart".

The server will write your configuration to `goconnect.yaml` and restart automatically. You can now log in!

## 3. Connecting Clients

Once your server is running:

1.  Download the **GoConnect Client** for your device.
2.  Launch the client.
3.  Enter your server URL (e.g., `http://your-server-ip:8080`).
4.  Log in with the user account you created.
5.  The client will automatically configure WireGuard and connect.

## 4. Running as a Service (Optional)

To keep GoConnect running in the background after you close the terminal:

### Linux (Systemd)
```bash
# Generate a service file (assuming binary is in /usr/local/bin)
sudo goconnect-server install
sudo systemctl start goconnect-server
```

### Windows
The installer (coming soon) handles this automatically. For manual setup, you can use Task Scheduler or run as a Service using NSSM.

## Next Steps
- [Configuration Reference](CONFIGURATION.md) - Customize ports, security, and more.
- [Admin Guide](ADMIN_GUIDE.md) - Learn how to manage networks and peers.
