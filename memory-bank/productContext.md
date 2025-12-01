# Product Context

## Problem Statement
Users want to connect devices (e.g., for gaming, file sharing, or remote work) as if they were on the same LAN, but the internet makes this hard due to NATs, firewalls, and complex IP configurations. Traditional VPNs are too technical, and existing P2P solutions often lack a unified, simple user experience.

## User Journeys

### Journey 1: Alice & The Minecraft Server (Host)
Alice wants to host a game server. She creates a network in GoConnect ("Alice's World"), copies the invite link, and shares it. She manages members (kick/ban) and enjoys a lag-free P2P connection.

### Journey 2: Bob Joins the Game (Client)
Bob clicks Alice's link, installs GoConnect, and joins. No config needed. He gets a virtual IP (e.g., 10.0.1.5) and connects to the game server immediately.

### Journey 3: Sarah's Secure File Drop (Work)
Sarah connects to her office network from a coffee shop. She drags and drops files via Windows Explorer using the virtual IP, bypassing slow cloud uploads thanks to P2P transfer.

## Experience Goals
- **"It Just Works":** Automatic discovery, NAT traversal, and reconnection.
- **Zero Friction:** No account creation for basic usage (MVP), no manual port forwarding.
- **Visual Feedback:** Clear status indicators (Connected/Disconnected, Direct/Relay).

## Functional Requirements (Summary)
- **Network:** Create, Join, Delete, Disconnect.
- **Members:** List, Kick, Ban.
- **Communication:** Broadcast Text Chat.
- **System:** Tray Icon, Notifications, Auto-start.
- **Interfaces:** Desktop GUI (Tauri), CLI TUI (Bubbletea).
