# 🖥️ GoConnect Desktop

GoConnect's desktop application. Create networks (host) or join existing networks (client) with a single app.

## ✨ Features

- 🌐 **Create Network** - Start your own virtual LAN
- 🔗 **Join Network** - One-click join via invite link
- 💬 **Chat** - Built-in text channels
- 👥 **Member Management** - Invite, kick, ban
- 🎨 **Modern UI** - Dark theme, user-friendly

## 🛠️ Technologies

| Layer | Technology |
|-------|------------|
| Framework | Tauri 2.0 |
| Frontend | React 19 + TypeScript |
| Styling | Tailwind CSS |
| Backend | Rust |

## 📦 Development

### Requirements

- Node.js 20+
- Rust (via rustup)
- Platform dependencies:
  - **Windows:** WebView2 (usually installed)
  - **macOS:** Xcode Command Line Tools
  - **Linux:** `webkit2gtk`, `libappindicator`

### Setup

```bash
# Install dependencies
npm install

# Run in development mode
npm run tauri dev

# Production build
npm run tauri build
```

### Project Structure

```
desktop/
├── src/                # React frontend
│   ├── App.tsx         # Main application
│   ├── main.tsx        # Entry point
│   └── index.css       # Global styles
├── src-tauri/          # Rust backend
│   ├── src/
│   │   └── main.rs     # Tauri application
│   ├── Cargo.toml      # Rust dependencies
│   └── tauri.conf.json # Tauri configuration
├── package.json
├── tailwind.config.js
└── vite.config.ts
```

## 🎨 UI Structure

```
┌────────────────────────────────────────────────────────────┐
│  GoConnect                                        ─ □ ✕   │
├────┬──────────────┬────────────────────────────────────────┤
│ 🏠 │  Network     │  Main content area                     │
│────│  Name        │                                        │
│ 🎮 │  NETWORKS    │  Connection status, members,          │
│ 💼 │  • Minecraft │  chat etc.                            │
│ 👥 │  • Work VPN  │                                        │
│    │              │                                        │
│ +  │  CHANNELS    │                                        │
│    │  # general   │                                        │
│ 👤 │  # announce  │                                        │
└────┴──────────────┴────────────────────────────────────────┘
```

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
