# Product Mission

## Pitch

**GoConnect** is a cross-platform virtual LAN application that helps **gamers and remote teams** create secure, private networks over the internet by providing **one-click network creation, integrated communication tools, and peer-to-peer file sharing** — all powered by WireGuard encryption.

> *"Virtual LAN made simple"*

## Users

### Primary Customers
- **Gamers**: Players who want to play LAN games over the internet with friends
- **Gaming Communities**: Groups organizing multiplayer sessions and private game servers
- **Tech-savvy Users**: People who value privacy and prefer self-hosted solutions

### User Personas

**Alex the Gamer** (18-35)
- **Role:** Casual to competitive gamer
- **Context:** Plays multiplayer games with friends scattered across different locations
- **Pain Points:** 
  - Complex VPN setup for LAN gaming
  - High latency with traditional VPNs
  - No integrated voice chat in networking tools
  - Limited free tier options (Hamachi 5-user cap)
- **Goals:** 
  - Play LAN games with friends seamlessly
  - Low-latency P2P connections
  - Easy invite system for new players
  - Integrated voice chat while gaming

**Jordan the Community Manager** (25-40)
- **Role:** Gaming community organizer / Discord server admin
- **Context:** Manages a gaming community with regular multiplayer events
- **Pain Points:**
  - Difficult to onboard new members to VPN
  - No central management dashboard
  - Switching between apps for chat, voice, and networking
- **Goals:**
  - One-click invite links for events
  - Member management with role-based access
  - All-in-one solution for networking + communication

## The Problem

### LAN Gaming is Broken in 2026

Many beloved games only support local network multiplayer, yet friends are scattered across the internet. Traditional VPNs are slow and complex. Existing solutions like Hamachi have user limits, ZeroTier is too technical, and Tailscale lacks gaming-focused features.

**Our Solution:** A modern, WireGuard-powered virtual LAN with integrated chat, voice, and file sharing — designed specifically for gamers who want to play together without the friction.

## Differentiators

### All-in-One Gaming Network
Unlike **Tailscale** (no chat/voice) and **ZeroTier** (complex setup), GoConnect provides integrated communication tools. This results in a seamless gaming experience without switching apps.

### Truly Free & Open Source
Unlike **Hamachi** (5-user limit) and **Tailscale** (cloud-only), GoConnect is MIT-licensed with no artificial restrictions. This means unlimited users and full self-hosting capability.

### Lightweight & Modern
Unlike **legacy VPN solutions** running heavy Electron apps, GoConnect uses Tauri (~15MB) with native performance. This results in faster startup, lower resource usage, and better gaming performance.

### WireGuard-Powered Security
Unlike **proprietary protocols**, GoConnect uses battle-tested WireGuard encryption. This provides modern cryptography with minimal latency overhead.

## Key Features

### Core Features
- **Create Private Networks:** One-click virtual LAN creation with custom network names
- **Join via Invite Links:** `gc://` protocol or invite codes for instant network access
- **Auto-Discovery:** Automatic device detection shows all connected peers

### Collaboration Features
- **Built-in Chat:** Text channels for each network, no Discord needed
- **Voice Communication:** WebRTC-powered voice chat for in-game coordination
- **P2P File Transfers:** Direct peer-to-peer file sharing at maximum speed

### Management Features
- **Member Management:** Add, remove, block members with role-based access control
- **Network Dashboard:** Visual overview of connected peers and network health
- **Multi-Platform:** Windows, macOS, Linux today — mobile coming soon
