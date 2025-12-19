# 🎮 Minecraft LAN Party with GoConnect

This guide shows you how to play Minecraft with friends over the internet using GoConnect, exactly as if you were in the same room.

## Prerequisites

*   **GoConnect** installed on all computers.
*   **Minecraft** (Java or Bedrock Edition) installed on all computers.
*   Everyone joined to the **same GoConnect network**.

---

## 🚀 Step 1: Create a Network

**Host (One person needs to do this):**

1.  Open GoConnect.
2.  Click **"New Network"** (or run `goconnect create "Minecraft Party"`).
3.  Copy the **Invite Code** or **Invite Link**.
4.  Share it with your friends.

---

## 🔗 Step 2: Join the Network

**Friends:**

1.  Open GoConnect.
2.  Click **"Join Network"**.
3.  Paste the code/link and join.
4.  Wait until everyone shows a green light (🟢) in the peer list.

---

## ⚔️ Step 3: Host the Game

1.  Open Minecraft.
2.  **Singleplayer** -> Load your world.
3.  Press `ESC` -> **Open to LAN**.
4.  Choose settings (Cheats, Game Mode) -> **Start LAN World**.
5.  Minecraft will show a port number in chat (e.g., `51234`). **Note this number.**

---

## 🏃 Step 4: Connect (Java Edition)

**Friends:**

1.  Copy the **Host's GoConnect IP** from the GoConnect app (it starts with `10.x.x.x` or similar).
2.  Open Minecraft -> **Multiplayer** -> **Direct Connection**.
3.  Enter the Address: `HOST_IP:PORT`
    *   Example: `100.64.0.1:51234`
4.  Click **Join Server**.

> 💡 **Tip:** If the Host's IP is `100.64.0.1` and the Minecraft port is `51234`, the full address is `100.64.0.1:51234`.

---

## 🪨 Step 5: Connect (Bedrock Edition)

**Friends:**

1.  Copy the **Host's GoConnect IP**.
2.  Open Minecraft -> **Play** -> **Servers**.
3.  Scroll down to **Add Server**.
    *   **Server Name**: Friend's World
    *   **Server Address**: `HOST_IP` (e.g., `100.64.0.1`)
    *   **Port**: The port shown by the host (default is 19132 for dedicated servers).
4.  Click **Play**.

---

## 🔧 Troubleshooting

*   **"Connection Refused"**: Ensure the Host has allowed **Java** through their Windows Firewall (Public & Private networks).
*   **"Timed Out"**: Check if GoConnect shows a direct connection (low ping). If relaying, latency might be higher.
*   **Can't see skins?**: This is normal in offline-mode LAN, but official accounts should see skins.

Happy Mining! ⛏️💎
