# 📚 GoConnect User Guide

This guide explains all GoConnect features in detail.

---

## 📋 Contents

1. [Introduction](#1-introduction)
2. [Installation](#2-installation)
3. [Getting Started](#3-getting-started)
4. [Network Management](#4-network-management)
5. [Member Management](#5-member-management)
6. [Chat](#6-chat)
7. [Settings](#7-settings)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Introduction

### What is GoConnect?

GoConnect is a platform that connects devices on the internet as if they're on the same local network.

**Key Concepts:**

| Concept | Description | Example |
|---------|-------------|---------|
| **Network** | Virtual LAN environment | "My Minecraft Server" |
| **Host** | Person who created the network | Server owner |
| **Member** | Person who joined the network | Players |
| **Invite Link** | Network join URL | `gc://join.goconnect.io/abc123` |
| **IP Address** | Address within the network | `10.0.1.5` |

### Supported Platforms

| Platform | Desktop App | Terminal App | Status |
|----------|-------------|--------------|--------|
| Windows 10/11 | ✅ | ✅ | Ready |
| macOS 11+ | ✅ | ✅ | Ready |
| Linux | ✅ | ✅ | Ready |
| Android | 📱 | - | Coming Soon |
| iOS | 📱 | - | Coming Soon |

---

## 2. Installation

### 2.1 System Requirements

**Minimum:**
- Processor: 1 GHz
- RAM: 512 MB
- Disk: 100 MB
- Network: Internet connection

**Recommended:**
- Processor: 2+ GHz
- RAM: 2 GB
- Disk: 500 MB
- Network: 10+ Mbps

### 2.2 Download

Download from [GitHub Releases](https://github.com/orhaniscoding/goconnect/releases/latest).

### 2.3 Platform-Specific Installation

#### Windows

1. Run `GoConnect-Setup.exe`
2. Click "Next" through the wizard
3. Select installation location (default recommended)
4. Click "Install"
5. Click "Finish"

**Note:** If Windows Defender warns, click "More info" → "Run anyway"

#### macOS

1. Open the `.dmg` file
2. Drag GoConnect to Applications
3. Gatekeeper warning will appear on first launch
4. Go to System Preferences → Security → "Open Anyway"

**Note:** Download ARM version for Apple Silicon (M1/M2/M3)

#### Linux

**Debian/Ubuntu:**
```bash
sudo dpkg -i goconnect_*.deb
sudo apt-get install -f  # Fix dependencies
```

**Fedora/RHEL:**
```bash
sudo rpm -i goconnect_*.rpm
```

**AppImage (All distros):**
```bash
chmod +x GoConnect-*.AppImage
./GoConnect-*.AppImage
```

---

## 3. Getting Started

### 3.1 Launching the Application

**Desktop:**
- Windows: Start menu → "GoConnect"
- macOS: Applications → GoConnect
- Linux: Application menu or `goconnect` command

**Terminal:**
```bash
goconnect
```

### 3.2 Main Screen

```
┌────────────────────────────────────────────────────────────┐
│  🔗 GoConnect                                    ─ □ ✕    │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                                                     │  │
│  │            🌐 Welcome!                             │  │
│  │                                                     │  │
│  │   Connect with friends on the same network.        │  │
│  │                                                     │  │
│  │   ┌───────────────┐    ┌───────────────┐          │  │
│  │   │ Create Network│    │ Join Network  │          │  │
│  │   │     🌐        │    │     🔗        │          │  │
│  │   └───────────────┘    └───────────────┘          │  │
│  │                                                     │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                            │
│  ────────────────────────────────────────────────────────  │
│  📡 My Networks (0)                                       │
│  ────────────────────────────────────────────────────────  │
│                                                            │
│  You're not connected to any networks yet.                │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 4. Network Management

### 4.1 Creating a Network

**Steps:**

1. Click "Create Network"
2. Fill in network details:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| Network Name | ✅ | Name of your network | "My Minecraft Server" |
| Description | ❌ | Short description | "Survival world" |
| Subnet | ❌ | IP range | `10.0.1.0/24` (auto) |
| Password | ❌ | Join password | Empty = no password |

3. Click "Create"
4. Copy the invite link

**Terminal:**
```bash
$ goconnect create "My Minecraft Server"

✅ Network created!

📋 Details:
   Network Name: My Minecraft Server
   Subnet: 10.0.1.0/24
   Your IP: 10.0.1.1

🔗 Invite Link:
   gc://join.goconnect.io/abc123xyz

   Share this link with your friends!
```

### 4.2 Joining a Network

**Steps:**

1. Click "Join Network"
2. Paste the invite link
3. Enter password if required
4. Click "Connect"

**Terminal:**
```bash
$ goconnect join gc://join.goconnect.io/abc123xyz

🔗 Connecting to: My Minecraft Server...

✅ Connected successfully!

📋 Details:
   Network Name: My Minecraft Server
   Subnet: 10.0.1.0/24
   Your IP: 10.0.1.5
   Online: 3 members
```

### 4.3 Managing Connections

**Disconnect:**
- Click "Disconnect" on the network card
- Or `goconnect disconnect`

**Reconnect:**
- Click "Connect" on the network card
- Or `goconnect connect "Network Name"`

### 4.4 Network Settings (Host)

As host, you can change network settings:

| Setting | Description |
|---------|-------------|
| Network Name | Change name |
| Description | Update description |
| Password | Add/remove join password |
| Invite Link | Generate new link |
| Delete Network | Permanently delete |

---

## 5. Member Management

### 5.1 Viewing Members

In the network details screen, see all members under "Members" tab:

```
┌─────────────────────────────────────────┐
│ 👥 Members (5)                          │
├─────────────────────────────────────────┤
│ 🟢 Alice (Host)        10.0.1.1         │
│ 🟢 Bob                 10.0.1.2         │
│ 🟢 Charlie             10.0.1.3         │
│ 🟡 Diana (Idle)        10.0.1.4         │
│ ⚫ Eve (Offline)       10.0.1.5         │
└─────────────────────────────────────────┘
```

**Status Indicators:**
- 🟢 Online
- 🟡 Idle (5+ minutes inactive)
- ⚫ Offline

### 5.2 Member Actions (Host)

As host, you can manage members:

| Action | Description |
|--------|-------------|
| **Kick** | Remove from network (can rejoin) |
| **Ban** | Permanently ban |
| **Unban** | Remove ban |

---

## 6. Chat

### 6.1 Text Channels

Each network has default chat channels:

- **#general** - General chat
- **#announcements** - Host only (optional)

### 6.2 Sending Messages

1. Select a channel from the list
2. Type in the text box at bottom
3. Press Enter or click "Send"

**Supported Features:**
- 📎 File sharing (up to 5 MB)
- 😀 Emoji
- @mention (user tagging)
- Edit/delete (your own messages)

---

## 7. Settings

### 7.1 General Settings

| Setting | Description | Default |
|---------|-------------|---------|
| Start on boot | Launch on computer start | ✅ |
| Minimize to tray | Go to tray on close | ✅ |
| Notifications | Desktop notifications | ✅ |
| Language | Interface language | English |
| Theme | Dark/Light | Dark |

### 7.2 Network Settings

| Setting | Description | Default |
|---------|-------------|---------|
| Auto-connect | Connect on app launch | ✅ |
| Reconnect | Retry on disconnect | ✅ |
| DNS settings | Custom DNS server | System |

### 7.3 Advanced Settings

| Setting | Description |
|---------|-------------|
| WireGuard interface | Network interface name |
| Logging level | Debug/Info/Warning/Error |
| Data folder | Config files location |

---

## 8. Troubleshooting

### 8.1 Common Issues

<details>
<summary><b>❌ Cannot connect</b></summary>

**Possible Causes:**
1. No internet connection
2. Firewall blocking
3. Host is offline

**Solutions:**
1. Check your internet connection
2. Allow GoConnect in firewall
3. Ensure host is online

```bash
# Windows Firewall
netsh advfirewall firewall add rule name="GoConnect" dir=in action=allow program="C:\Program Files\GoConnect\goconnect.exe"

# Linux UFW
sudo ufw allow 51820/udp
```
</details>

<details>
<summary><b>❌ Cannot ping other devices</b></summary>

**Possible Causes:**
1. Target device offline
2. Firewall blocking ping
3. Wrong IP address

**Solutions:**
1. Check if target is online
2. Allow ICMP on both sides
3. Verify IP from "Members" list
</details>

<details>
<summary><b>❌ App won't start</b></summary>

**Solutions:**
1. Restart computer
2. Reinstall application
3. Check log files:
   - Windows: `%APPDATA%\GoConnect\logs`
   - macOS: `~/Library/Logs/GoConnect`
   - Linux: `~/.local/share/goconnect/logs`
</details>

### 8.2 Viewing Logs

**Desktop:**
Settings → Advanced → "Open Logs"

**Terminal:**
```bash
goconnect logs
goconnect logs --level debug
```

### 8.3 Getting Support

1. [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues) - Bug reports
2. [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions) - Questions
3. [FAQ](../README.md#-faq) - Frequently asked questions

---

<div align="center">

**[← Home](../README.md)** | **[Quick Start →](../QUICK_START.md)**

</div>
