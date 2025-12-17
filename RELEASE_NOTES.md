# GoConnect Release Notes

## v1.1.0 (Current)

### New Features

#### 🔔 Push Notifications
- Cross-platform native notifications
- Linux: `notify-send` integration
- macOS: AppleScript notifications
- Windows: PowerShell toast notifications
- Settings: Enable/Disable, Do Not Disturb, Sound control

#### 📊 Metrics Dashboard
- Real-time network monitoring
- 6 Prometheus metrics:
  - WebSocket connections & rooms
  - Active networks & peers
  - Message counts & membership stats
- JSON summary endpoint for integrations
- Auto-refresh every 10 seconds

#### 🔄 Auto-Update
- CLI: `goconnect update` command
- Desktop: Tauri updater plugin
- Automatic version checking
- Download progress tracking
- Release notes display

---

## v1.0.0

### Initial Release

Complete virtual LAN platform with:

- **Networking**: WireGuard-based secure tunnels
- **Desktop App**: Tauri 2.0 + React
- **CLI**: Go + Bubbletea TUI
- **Server**: Self-hosted with PostgreSQL/SQLite
- **Chat**: Real-time messaging
- **Voice**: WebRTC communication
- **File Transfers**: P2P sharing
- **Security**: 2FA, JWT, audit logging

---

For full changelog, see [CHANGELOG.md](CHANGELOG.md).
