# 🖥️ GoConnect Desktop

GoConnect's desktop application built with Tauri. Create networks or join existing ones with a beautiful GUI.

## ✨ Features

- 🌐 **Create Network** - Start your own virtual LAN
- 🔗 **Join Network** - One-click join via invite link
- 💬 **Chat** - Built-in text channels
- 👥 **Member Management** - View online members
- 🎨 **Modern UI** - Clean dark theme
- 🔔 **System Tray** - Quick access to networks

## 🛠️ Technologies

| Layer | Technology |
|-------|------------|
| Framework | Tauri 2.0 |
| Frontend | React 19 + TypeScript |
| Styling | Tailwind CSS 4.0 |
| State | Zustand |
| Backend | Rust |
| IPC | gRPC (to CLI daemon) |

## 📦 Downloads

Download from [GitHub Releases](https://github.com/orhaniscoding/goconnect/releases/latest):

| Platform | File |
|----------|------|
| Windows | `GoConnect_*_x64-setup.exe` |
| Windows (MSI) | `GoConnect_*_x64_en-US.msi` |
| macOS (Apple Silicon) | `GoConnect_*_aarch64.dmg` |
| macOS (Intel) | `GoConnect_*_x64.dmg` |
| Linux (Debian/Ubuntu) | `GoConnect_*_amd64.deb` |
| Linux (Universal) | `GoConnect_*_amd64.AppImage` |

## 🚀 Development

### Requirements

- Node.js 20+
- Rust (via rustup)
- protoc (Protocol Buffers compiler)
- Platform dependencies:
  - **Windows**: WebView2 (usually pre-installed)
  - **macOS**: Xcode Command Line Tools
  - **Linux**: `webkit2gtk-4.1`, `libappindicator3`, `librsvg2`

### Setup

```bash
# Install dependencies
npm install

# Run in development mode
npm run tauri dev

# Production build
npm run tauri build
```

### Linux Dependencies

```bash
# Ubuntu/Debian
sudo apt install libwebkit2gtk-4.1-dev libappindicator3-dev librsvg2-dev

# Fedora
sudo dnf install webkit2gtk4.1-devel libappindicator-gtk3-devel librsvg2-devel

# Arch
sudo pacman -S webkit2gtk-4.1 libappindicator-gtk3 librsvg
```

## 🏗️ Project Structure

```
desktop/
├── src/                    # React frontend
│   ├── components/         # UI components
│   │   ├── Layout.tsx
│   │   ├── Sidebar.tsx
│   │   ├── NetworkList.tsx
│   │   └── ChatPanel.tsx
│   ├── lib/                # Utilities
│   │   ├── daemon.ts       # Daemon communication
│   │   └── hooks.ts        # React hooks
│   ├── App.tsx             # Main application
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles (Tailwind)
├── src-tauri/              # Rust backend
│   ├── src/
│   │   ├── main.rs         # Tauri application
│   │   ├── commands.rs     # Tauri commands
│   │   └── daemon.rs       # Daemon gRPC client
│   ├── Cargo.toml          # Rust dependencies
│   └── tauri.conf.json     # Tauri configuration
├── public/                 # Static assets
├── package.json            # Node dependencies
├── tailwind.config.js      # Tailwind configuration
├── vite.config.ts          # Vite configuration
└── tsconfig.json           # TypeScript configuration
```

## 🎨 UI Design

```
┌────────────────────────────────────────────────────────────┐
│  GoConnect                                        ─ □ ✕   │
├────┬───────────────┬───────────────────────────────────────┤
│ 🏠 │  Network      │  Main Content                         │
│────│  Name         │                                       │
│ 🎮 │               │  ┌─────────────────────────────────┐  │
│ 💼 │  NETWORKS     │  │  Connection Status              │  │
│ 👥 │  • Gaming     │  │  ● Connected                    │  │
│    │  • Work       │  │  IP: 10.0.1.5                   │  │
│ +  │               │  │  Latency: 12ms                  │  │
│    │  CHANNELS     │  └─────────────────────────────────┘  │
│ ⚙️ │  # general    │                                       │
│    │  # voice      │  ┌─────────────────────────────────┐  │
│ 👤 │               │  │  Online Members (3)             │  │
│    │               │  │  • Alice (10.0.1.2)             │  │
│    │               │  │  • Bob (10.0.1.3)               │  │
│    │               │  │  • You (10.0.1.5)               │  │
│    │               │  └─────────────────────────────────┘  │
└────┴───────────────┴───────────────────────────────────────┘
```

## ⚙️ Configuration

Tauri configuration in `src-tauri/tauri.conf.json`:

```json
{
  "productName": "GoConnect",
  "version": "3.0.0",
  "identifier": "com.goconnect.app",
  "bundle": {
    "active": true,
    "targets": "all",
    "category": "Network"
  }
}
```

## 🧪 Testing

```bash
# Frontend tests
npm test

# Type checking
npm run typecheck

# Lint
npm run lint
```

## 📦 Building for Release

```bash
# Build for current platform
npm run tauri build

# Build outputs location:
# - Windows: src-tauri/target/release/bundle/msi/
# - macOS: src-tauri/target/release/bundle/dmg/
# - Linux: src-tauri/target/release/bundle/deb/
```

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details.
