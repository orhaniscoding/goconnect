# Remote Desktop Access Tutorial

**Access your home computer from anywhere using GoConnect.**

---

## 📋 Overview

This tutorial shows you how to:
- Access your home computer from work or while traveling
- Use Remote Desktop (RDP) on Windows
- Use Screen Sharing on macOS
- Access your computer securely through GoConnect

**Use Cases:**
- 🏠 Access files on your home computer
- 🛠️ Help family members with technical problems
- 💼 Work from home using your office computer
- 🎮 Play games stored on your home PC

---

## Prerequisites

- **Two computers:**
  - Home computer (the one you'll connect TO)
  - Remote computer (the one you'll connect FROM)
- **GoConnect installed** on both computers
- **Internet connection** on both computers
- **Administrator access** on both computers

---

## Windows Remote Desktop (RDP)

### Step 1: Enable Remote Desktop on Home Computer

**IMPORTANT:** Windows Home edition doesn't support RDP as a HOST. You need Windows Pro, Enterprise, or Ultimate.

1. Press `Win + R`, type `sysdm.cpl`, press Enter
2. **Remote** tab → Check "Allow remote connections to this computer"
3. Click "Select Users"
4. Add your user account if not listed
5. Click **OK** twice

### Step 2: Find Your Computer's IP

1. Open GoConnect on home computer
2. Note your virtual IP (e.g., `10.0.1.5`)
3. This will be your RDP address

### Step 3: Configure Windows Firewall

On home computer:

1. Press `Win + R`, type `wf.msc`, press Enter
2. **Inbound Rules** → **New Rule**
3. **Port** → Select "TCP" → Specify "3389"
4. **Action** → "Allow the connection"
5. **Profile** → Check all boxes (Domain, Private, Public)
6. **Name** → "Remote Desktop"
7. Click **Finish**

### Step 4: Create GoConnect Network

On either computer:

1. Open GoConnect
2. Click **"Create Network"**
3. Name it (e.g., "Remote Desktop Network")
4. Click **Create**
5. Copy invite link

### Step 5: Join Network on Remote Computer

1. Open GoConnect on remote computer
2. Click **"Join Network"**
3. Paste invite link
4. Click **Join**
5. Wait for connection

### Step 6: Connect via Remote Desktop

On remote computer:

1. Press `Win + R`, type `mstsc.exe`, press Enter
2. In "Computer:" field, enter home computer's GoConnect IP:
   ```
   10.0.1.5
   ```
3. Click **Connect**
4. Enter your Windows credentials
5. Check "Don't ask me again"
6. Click **Yes**

**Success!** You're now remotely controlling your home computer!

---

## macOS Screen Sharing

### Step 1: Enable Screen Sharing on Home Mac

On home Mac:

1. **System Preferences** → **Sharing**
2. Check "Screen Sharing"
3. Click "Computer Settings"
4. Allow access for:
   - ✅ "These users:"
   - Add your user account
5. Note the computer name shown at the top

### Step 2: Configure Firewall

Screen Sharing automatically configures firewall. If prompted:
1. Click "Allow"
2. Enter your password

### Step 3: Create and Join GoConnect Network

Same as Windows steps above (Steps 4-5).

### Step 4: Connect via Screen Sharing

On remote Mac:

**Option 1: Using Finder**

1. Open Finder
2. Press `Cmd + K` or **Go** → **Connect to Server**
3. Enter:
   ```
   vnc://10.0.1.5
   ```
   (replace `10.0.1.5` with home Mac's GoConnect IP)
4. Click **Connect**
5. Enter home Mac credentials
6. Click **Connect**

**Option 2: Using Terminal**

```bash
# Open Screen Sharing directly
open vnc://10.0.1.5
```

### Step 5: Control Screen

Once connected:
- Observe screen in real-time
- Move mouse, click, type
- Copy/paste between computers

---

## Linux Remote Desktop (VNC)

### Step 1: Install VNC Server

On home Linux computer:

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install -y x11vnc xvfb
```

**Fedora:**

```bash
sudo dnf install -y tigervnc-server
```

### Step 2: Configure VNC Server

```bash
# Set VNC password
vncpasswd

# Start VNC server
x11vnc -display :0 -forever -shared -rfbauth ~/.vnc/passwd
```

### Step 3: Create Systemd Service (Optional)

```bash
sudo nano /etc/systemd/system/vncserver.service
```

Content:

```ini
[Unit]
Description=VNC Server
After=network.target

[Service]
Type=simple
User=yourusername
ExecStart=/usr/bin/x11vnc -display :0 -forever -shared -rfbauth /home/yourusername/.vnc/passwd

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable vncserver
sudo systemctl start vncserver
```

### Step 4: Create and Join GoConnect Network

Same as Windows steps above.

### Step 5: Connect

On remote computer:

**Windows Client:**

1. Install VNC Viewer (e.g., RealVNC)
2. Enter home computer's GoConnect IP:
   ```
   10.0.1.5:5900
   ```
3. Enter VNC password

**macOS Client:**

1. Open Safari
2. Go to: `vnc://10.0.1.5:5900`
3. Enter VNC password

**Linux Client:**

```bash
vncviewer 10.0.1.5:5900
```

---

## Security Best Practices

### ✅ Do's

**Use Strong Passwords:**
- Windows: Use strong user password
- macOS: Use strong user password
- VNC: Use separate VNC password

**Limit Access:**
- Only allow users you trust
- Remove access when not needed
- Use separate user account for remote access

**Keep Updated:**
- Keep GoConnect updated
- Install security patches
- Update remote desktop software

**Enable Firewall:**
- Only expose necessary ports (3389, 5900)
- Use Windows Firewall / macOS Firewall
- Don't disable firewall

### ❌ Don'ts

**Don't expose RDP/VNC to public internet:**
- Always use GoConnect
- Don't port forward RDP/VNC to internet
- GoConnect provides the secure tunnel

**Don't use weak passwords:**
- Avoid "123456", "password", etc.
- Use unique, complex passwords
- Consider password manager

**Don't leave unattended sessions open:**
- Lock remote computer when done
- Sign out when finished
- Don't leave visible for others to see

---

## Troubleshooting

### "Remote Connection Blocked"

**Problem:** Connection refused

**Solutions:**

1. **Check if GoConnect is connected:**
   - Both computers should show "Connected" status
   - Check virtual IP addresses

2. **Check firewall:**
   - Windows: Allow RDP through firewall
   - macOS: Check Screen Sharing is enabled
   - Linux: Check VNC server is running

3. **Test connectivity:**
   ```bash
   # From remote computer, ping home computer
   ping 10.0.1.5
   ```

### "Authentication Failed"

**Problem:** Wrong credentials

**Solutions:**

1. **Check username:**
   - Windows: Use format `COMPUTERNAME\username`
   - macOS/Linux: Use username

2. **Reset password:**
   - Windows: Reset user password
   - macOS: Reset user password
   - VNC: Run `vncpasswd` again

3. **Check permissions:**
   - Ensure your user is allowed to connect
   - Windows: Add to "Remote Desktop Users" group
   - macOS: Add to Screen Sharing users list

### Slow Performance

**Problem:** Laggy remote desktop

**Solutions:**

1. **Reduce color depth:**
   - RDP: Adjust experience settings
   - VNC: Reduce color depth to 16-bit

2. **Close unnecessary programs:**
   - Close browser, media players on home computer
   - Free up resources

3. **Check bandwidth:**
   - Both computers need good internet
   - Consider wired connection on at least one side

4. **Adjust GoConnect MTU:**
   - Lower MTU can help with slow connections
   - GoConnect → Settings → Advanced → MTU: 1280

### "Computer Sleeping" Issue

**Problem:** Can't connect when computer sleeps

**Solutions:**

1. **Disable sleep:**
   - Windows: Settings → System → Power & sleep → Sleep → Never
   - macOS: System Preferences → Energy Saver → Prevent computer from sleeping
   - Linux: Configure power management

2. **Enable wake-on-LAN:**
   - BIOS/UEFI settings
   - Network adapter properties

3. **Use hibernation instead:**
   - Computer resumes faster from hibernation
   - Preserves state

---

## Advanced Usage

### Multiple Monitors

**Windows RDP:**

1. Open Remote Desktop Connection
2. Show Options → Display tab
3. Check "Use all my monitors"
4. Connect

### Clipboard Sharing

**Automatically enabled** in:
- Windows RDP
- macOS Screen Sharing
- Most VNC implementations

**Test it:**
1. Copy text on remote computer
2. Paste on local computer (or vice versa)

### File Transfer

**During Remote Session:**

1. **Windows RDP:**
   - Open Remote Desktop Connection
   - Show Options → Local Resources → More
   - Check "Drives"
   - Connect
   - Access home computer drives in File Explorer

2. **macOS:**
   - Screen sharing doesn't include file transfer
   - Use GoConnect file transfer feature instead:
     - Open GoConnect chat
     - Click attachment icon
     - Send file

3. **Linux VNC:**
   - Use GoConnect file transfer
   - Or use SCP/SFTP

---

## Alternative: SSH Tunneling (Linux)

### For Command Line Access

On remote computer:

```bash
# SSH into home computer
ssh username@10.0.1.5

# Tunnel specific application
ssh -L 5900:localhost:5900 username@10.0.1.5
```

---

## Next Steps

### Test Connection

1. Ensure both computers are on
2. Open GoConnect on both
3. Verify connection status
4. Test remote desktop connection

### Disconnect Securely

1. Close remote desktop window
2. Sign out of remote session
3. Leave GoConnect running or exit

### Permanent Setup

**Auto-start GoConnect:**
- Windows: Add to startup folder
- macOS: Login items
- Linux: Systemd service (see above)

**Auto-reconnect:**
- GoConnect → Settings
- Enable "Auto-reconnect on startup"

---

## Tips

### Optimize Performance

**Home Computer:**
- Use wired Ethernet instead of Wi-Fi
- Close unnecessary programs
- Disable animations and effects
- Use SSD instead of HDD

**Remote Computer:**
- Stable internet connection
- Good Wi-Fi or Ethernet
- Close other bandwidth-heavy apps

### Security

**Change Password Regularly:**
- Every 60-90 days
- Use strong, unique passwords
- Enable 2FA where available

**Monitor Access:**
- Check login logs regularly
- Review connected devices
- Remove unknown users

---

## Need Help?

- 📖 [Full Documentation](../../README.md)
- 💬 [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 🐛 [Report Issue](https://github.com/orhaniscoding/goconnect/issues/new)

---

**Last Updated:** 2025-01-24
**Version:** 1.0.0
