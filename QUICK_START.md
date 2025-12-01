# 🚀 GoConnect Quick Start

Get GoConnect running in 5 minutes.

---

## 📋 Contents

1. [Download](#1-download)
2. [Install](#2-install)
3. [Launch](#3-launch)
4. [Create or Join Network](#4-create-or-join-network)
5. [Use](#5-use)

---

## 1. Download

### Desktop Application (Recommended)

Download from [GitHub Releases](https://github.com/orhaniscoding/goconnect/releases/latest):

| Operating System | File |
|------------------|------|
| Windows | `GoConnect_*_x64-setup.exe` |
| macOS (Apple Silicon) | `GoConnect_*_aarch64.dmg` |
| macOS (Intel) | `GoConnect_*_x64.dmg` |
| Linux (Debian/Ubuntu) | `GoConnect_*_amd64.deb` |
| Linux (Other) | `GoConnect_*_amd64.AppImage` |

### Terminal Application (CLI)

For command-line users and servers:

```bash
# Linux (x64)
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli_v3.0.0_linux_amd64.tar.gz
tar -xzf goconnect-cli_v3.0.0_linux_amd64.tar.gz
sudo mv goconnect-cli_v3.0.0_linux_amd64 /usr/local/bin/goconnect

# macOS (Apple Silicon)
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli_v3.0.0_darwin_arm64.tar.gz
tar -xzf goconnect-cli_v3.0.0_darwin_arm64.tar.gz
sudo mv goconnect-cli_v3.0.0_darwin_arm64 /usr/local/bin/goconnect
```

```powershell
# Windows PowerShell
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-cli_v3.0.0_windows_amd64.zip" -OutFile "goconnect-cli.zip"
Expand-Archive -Path "goconnect-cli.zip" -DestinationPath "."
```

---

## 2. Install

### Windows
1. Double-click `GoConnect_*_x64-setup.exe`
2. Follow the installation wizard
3. Click "Finish"

### macOS
1. Open the `.dmg` file
2. Drag GoConnect to Applications
3. On first launch, right-click → "Open" (or allow in Security settings)

### Linux (Debian/Ubuntu)
```bash
sudo dpkg -i GoConnect_*_amd64.deb
```

### Linux (AppImage)
```bash
chmod +x GoConnect_*_amd64.AppImage
./GoConnect_*_amd64.AppImage
```

---

## 3. Launch

### Desktop Application

1. Start GoConnect from your applications menu
2. You'll see the welcome screen:

```
┌──────────────────────────────────────┐
│         🔗 Welcome to GoConnect      │
│                                      │
│    "Virtual LAN made simple."        │
│                                      │
│   ┌────────────────────────────┐     │
│   │     🌐 Create Network      │     │
│   │     Start your own         │     │
│   └────────────────────────────┘     │
│                                      │
│   ┌────────────────────────────┐     │
│   │     🔗 Join Network        │     │
│   │     Join with invite link  │     │
│   └────────────────────────────┘     │
│                                      │
└──────────────────────────────────────┘
```

### Terminal Application (CLI)

```bash
$ goconnect

  🔗 GoConnect v3.0.0

  ? What would you like to do?
  ❯ 🌐 Create Network
    🔗 Join Network
    📋 My Networks
    ⚙️  Settings
    ❌ Exit
```

---

## 4. Create or Join Network

### Option A: Create New Network

**When to use?**
- You want to play games with friends
- You want to set up your own private LAN
- You need a network for file sharing

**Steps:**

1. Select "Create Network"
2. Enter network details:
   - **Network Name**: `My Minecraft Server`
   - **Description**: `Survival world with friends`
3. Click "Create"
4. Copy and share the invite link!

```
✅ Network created!

📋 Invite Link:
   gc://join.goconnect.io/abc123xyz

🔗 Share this link with your friends!
```

### Option B: Join Existing Network

**When to use?**
- Someone sent you an invite link
- You want to join someone else's network

**Steps:**

1. Select "Join Network"
2. Paste invite link: `gc://join.goconnect.io/abc123xyz`
3. Click "Connect"
4. You're connected!

```
✅ Connected successfully!

🌐 Network: My Minecraft Server
🖥️ Your IP: 10.0.1.5
👥 Online: 3 members

You're now on the same LAN!
```

---

## 5. Use

### Check Connection Status

**Desktop:**
- Look at the GoConnect icon in system tray
- 🟢 Green = Connected
- 🔴 Red = Disconnected

**Terminal:**
```bash
$ goconnect status

🌐 Connected Networks:
   • My Minecraft Server (10.0.1.0/24)
     IP: 10.0.1.5
     Online: 3 members
```

### Access Other Devices

Now you can reach other devices on the network by IP:

```bash
# Ping
ping 10.0.1.2

# SSH connection
ssh user@10.0.1.3

# File sharing
\\10.0.1.4\shared  # Windows
smb://10.0.1.4/shared  # macOS
```

### Minecraft LAN Example

1. Open Minecraft
2. Open world → "Open to LAN"
3. Note the port number (e.g., 25565)
4. Friends connect via "Direct Connect": `10.0.1.1:25565`

---

## 🎉 Congratulations!

You've successfully set up and started using GoConnect!

### Next Steps

- 📖 [Full Documentation](README.md)
- 🏗️ [Architecture Overview](docs/ARCHITECTURE.md)
- ❓ [FAQ](README.md#-faq)
- 🐛 [Report Issues](https://github.com/orhaniscoding/goconnect/issues)

### Need Help?

- 💬 [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 📧 Support: Open an issue on GitHub

---

<div align="center">

**[← Back to README](README.md)**

</div>
