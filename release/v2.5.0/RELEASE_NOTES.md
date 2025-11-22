# GoConnect v2.5.0 Release Notes

**Date:** November 23, 2025
**Version:** 2.5.0

## 🚀 Summary of Changes

This release introduces significant enhancements to the Real-Time Communication (RTC) capabilities of GoConnect. We have completed the implementation of the core WebSocket features required for a rich chat and collaboration experience.

### ✨ New Features

*   **Message Threads:** Users can now reply to specific messages, creating threaded conversations.
*   **Typing Indicators:** Real-time feedback when other users are typing in a room or DM.
*   **Screen Sharing Signaling:** Added support for `call.offer` with `callType: "screen"` to enable screen sharing sessions via WebRTC.
*   **File Upload Progress:** Real-time progress updates for file uploads within chat rooms.
*   **Message Reactions:** Users can react to messages with emojis.
*   **Read Receipts:** See when your messages have been read by others.
*   **Direct Messages (DMs):** Private 1-on-1 messaging with canonical scope handling.
*   **WireGuard Metrics:** New Prometheus metrics for monitoring WireGuard interface statistics.

### 🐛 Bug Fixes

*   Fixed syntax errors in test files related to previous refactors.
*   Corrected `CallSignalData` structure to properly propagate `CallType` and `Signal` payloads.

## 📦 Assets

The following archives are available for this release. Each archive contains the Server and Daemon binaries, along with the README and LICENSE.

| Platform    | Architecture | Filename                              |
| :---------- | :----------- | :------------------------------------ |
| **Windows** | amd64        | `goconnect_2.5.0_windows_amd64.zip`   |
| **Linux**   | amd64        | `goconnect_2.5.0_linux_amd64.tar.gz`  |
| **macOS**   | amd64        | `goconnect_2.5.0_darwin_amd64.tar.gz` |

## 🛠 Installation

### Server

1.  Extract the archive for your platform.
2.  Ensure you have a `config.yaml` or environment variables set up (see `README.md`).
3.  Run the binary: `./goconnect-server` (or `goconnect-server.exe` on Windows)

### Client Daemon

1.  Extract the archive for your platform.
2.  Run with root/admin privileges to manage WireGuard interfaces.
3.  Connect to a server: `./goconnect-daemon connect --server <url> --token <token>`

## 🛠 Installation

### Server

1.  Download the appropriate binary for your server OS.
2.  Ensure you have a `config.yaml` or environment variables set up (see `README.md`).
3.  Run the binary: `./goconnect-server-linux-amd64`

### Client Daemon

1.  Download the binary for your client OS.
2.  Run with root/admin privileges to manage WireGuard interfaces.
3.  Connect to a server: `./goconnect-daemon-linux-amd64 connect --server <url> --token <token>`

## ⚠️ Known Issues

*   File upload is currently signaling-only; the actual binary data transfer implementation depends on the configured storage backend (S3/Local).

---
*Generated automatically by GoConnect Release Automation*
