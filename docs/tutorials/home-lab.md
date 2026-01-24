# Home Lab Setup Tutorial

**Build your own home server lab with GoConnect for remote access to your NAS, servers, and devices.**

---

## 📋 Overview

This tutorial shows you how to:
- Access your home NAS from anywhere
- Manage home servers remotely
- Connect to multiple devices
- Set up permanent connections
- Secure your home lab

**Use Cases:**
- 🖥️ Access home server/cluster from work
- 💾 Manage NAS storage remotely
- 🏠 Control smart home devices
- 🎮 Game servers
- 📊 Monitor systems remotely

---

## Prerequisites

### Hardware
- **GoConnect installed** on all devices
- **Home server/NAS** (always-on device recommended)
- **Multiple devices** you want to access
- **Internet connection** on both ends

### Software
- GoConnect client on:
  - Home server (Linux/Windows/macOS)
  - Your laptop (remote)
  - Other devices you want to access

---

## Scenario 1: NAS Access

### Step 1: Set Up Home NAS

**Enable Remote Access on NAS:**

**Synology NAS:**
1. Open DSM (DiskStation Manager)
2. **Control Panel** → **Remote Services**
3. Enable "QuickConnect"
4. Or enable SSH: **Terminal & SNMP** → Enable SSH service

**QNAP NAS:**
1. QTS control panel
2. **Applications** → **App Center**
3. Install "myQNAPcloud" or enable SSH

**Generic NAS:**
1. Enable SSH access
2. Note SSH port (usually 22)
3. Set up static IP for NAS

### Step 2: Install GoConnect on NAS

**Synology (DSM 7.0+):**

1. Open **Package Center**
2. Search for "GoConnect" (or install via community repo)
3. Install and start GoConnect

**QNAP (QuTShero):**

1. Open **App Center**
2. Install from "GitHub" or upload package
3. Start GoConnect

**Linux-based NAS:**

```bash
# Download GoConnect CLI for Linux
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server-linux-amd64.tar.gz

# Extract
tar -xzf goconnect-server-linux-amd64.tar.gz

# Install
sudo mv goconnect /usr/local/bin/

# Create service
sudo useradd -r -s /bin/false goconnect
```

### Step 3: Create Permanent Network

1. Open GoConnect on NAS
2. Click **"Create Network"**
3. Name: "Home Lab - NAS"
4. Set description: "Access home NAS and servers"
5. Click **Create**

**Important:** Note the invite link!

### Step 4: Connect Remote Computer

On your laptop (remote):

1. Open GoConnect
2. Click **"Join Network"**
3. Paste invite link
4. Click **Join**

### Step 5: Access NAS

**Via File Manager:**

**Windows:**
1. Open File Explorer
2. In address bar, type: `\\<NAS-IP>`
   - Replace `<NAS-IP>` with GoConnect virtual IP (e.g., `\\10.0.1.10`)
   - Or use NAS hostname: `\\<hostname>`

**macOS:**
1. Open Finder
2. **Go** → **Connect to Server** (Cmd+K)
3. Enter: `smb://10.0.1.10`
4. Enter NAS credentials
5. Click **Connect**

**Linux:**
```bash
# Mount NAS share
sudo mount -t cifs //10.0.1.1/share /mnt/nas -o user=username,password=pass
```

---

## Scenario 2: Home Server Management

### Step 1: Prepare Home Server

**Install GoConnect:**

```bash
# On Ubuntu/Debian
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server-linux-amd64.deb
sudo dpkg -i goconnect-server-linux-amd64.deb
```

**Or compile from source:**

```bash
git clone https://github.com/orhaniscoding/goconnect.git
cd goconnect/cli
go build -o goconnect ./cmd/goconnect
sudo mv goconnect /usr/local/bin/
```

### Step 2: Create Network

Same as NAS scenario above.

### Step 3: Access via SSH

**From remote computer:**

```bash
# SSH to home server using GoConnect IP
ssh username@10.0.1.5

# Or with specific port
ssh -p 22 username@10.0.1.5
```

**Web Interface:**

If home server has web UI (e.g., Portainer, Cockpit):

1. Open browser
2. Navigate to `http://10.0.1.5:port`
3. Login with credentials

---

## Scenario 3: Multi-Device Setup

### Architecture

```
                    ┌──────────────────┐
                    │   Home Network    │
                    │                    │
    ┌───────────┐   │                    │
    │  NAS      │   │  ┌──────────────┐  │
    │ 10.0.1.10│◄──┘  │  │ Home Server │  │
    └───────────┘      │  │  10.0.1.5   │  │
                      │  └──────────────┘  │
    ┌───────────┐   │                    │
    │  Desktop  │   │  ┌──────────────┐  │
    │ 10.0.1.20│◄──┘  │  │  Media       │  │
    └───────────┘      │  │  10.0.1.30   │  │
                      │  └──────────────┘  │
    ┌───────────┐   │                    │
    │  Laptop   │   │  ┌──────────────┐  │
    │ (Remote)  │   │  │  Pi Cluster  │  │
    │ 10.0.1.50│   │  │  10.0.1.40   │  │
    └───────────┘   │  └──────────────┘  │
                      └────────────────────┘
```

### Step 1: Assign Static IPs

**On each device, set static GoConnect IP:**

**Method 1: GoConnect Settings**

1. Open GoConnect on device
2. **Settings** → **Network**
3. **"Request specific IP"**
4. Enter desired IP (e.g., `10.0.1.10`)
5. Reconnect to network

**Method 2: DHCP Reservation (on router)**

1. Access router admin panel
2. Find device by MAC address
3. Reserve IP for that device
4. Reboot device

### Step 2: Document IP Assignments

Create `~/home-lab-inventory.md`:

```markdown
# Home Lab IP Inventory

## Devices

### NAS
- **Name:** Synology DS920+
- **GoConnect IP:** 10.0.1.10
- **Local IP:** 192.168.1.100
- **MAC:** 00:11:32:AB:CD:EF
- **Services:** SMB, SSH (22), Web (5001)

### Home Server
- **Name:** Ubuntu 22.04 LTS
- **GoConnect IP:** 10.0.1.5
- **Local IP:** 192.168.1.50
- **MAC:** 00:11:32:AB:CD:F0
- **Services:** SSH (22), Web (80, 443), Cockpit (9090)

### Desktop
- **Name:** Windows 11
- **GoConnect IP:** 10.0.1.20
- **Local IP:** 192.168.1.20
- **MAC:** 00:11:32:AB:CD:F1

### Media Server
- **Name:** Plex Server
- **GoConnect IP:** 10.0.1.30
- **Local IP:** 192.168.1.30
- **Services:** Plex (32400)

### Pi Cluster
- **Name:** Kubernetes Cluster
- **GoConnect IP:** 10.0.1.40
- **Services:** API (6443), Dashboard (8443)
```

### Step 3: Connect All Devices

1. Install GoConnect on each device
2. Join same network on all devices
3. Verify connectivity:
   ```bash
   # From any device, ping others
   ping 10.0.1.5
   ping 10.0.1.10
   ping 10.0.1.30
   ```

---

## Scenario 4: Automation & Monitoring

### Monitor Multiple Servers

**Using GoConnect:**

1. Create monitoring dashboard (e.g., Grafana)
2. Connect to all home lab servers
3. Collect metrics via GoConnect network

**Example Prometheus Configuration:**

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'home-server'
    static_configs:
      - targets: ['10.0.1.5:9090']

  - job_name: 'nas'
    static_configs:
      - targets: ['10.0.1.10:5001']

  - job_name: 'media-server'
    static_configs:
      - targets: ['10.0.1.30:32400']
```

### Automated Backups

**Backup NAS to Cloud:**

```bash
#!/bin/bash
# backup-nas.sh

# Sync NAS to backup server via GoConnect
rsync -avz --progress \
  username@10.0.1.10:/volume1/ \
  /backup/nas/

# Send notification
echo "NAS backup completed at $(date)" | \
  mail -s "Home Lab Backup" your@email.com
```

**Schedule with cron:**

```bash
# Every night at 2 AM
0 2 * * * /home/user/scripts/backup-nas.sh
```

---

## Security Best Practices

### ✅ Do's

**Network Security:**
- Use strong passwords for all devices
- Enable GoConnect encryption (WireGuard)
- Keep firmware updated
- Use separate VLAN for home lab (if router supports)

**Access Control:**
- Only join networks you trust
- Remove access when not needed
- Use separate user accounts for remote access
- Review connected devices regularly

**Monitoring:**
- Enable logging on all devices
- Check access logs regularly
- Set up alerts for suspicious activity

### ❌ Don'ts

- Don't expose management ports to public internet
- Don't use default passwords
- Don't skip security updates
- Don't share invite links publicly

---

## Advanced: Home Automation

### Smart Home Integration

**Home Assistant via GoConnect:**

1. Install Home Assistant on home server
2. Access via GoConnect IP:
   ```
   http://10.0.1.5:8123
   ```
3. Control devices remotely

### IoT Device Management

**Connect IoT devices:**

1. Install GoConnect on IoT hub
2. Join home lab network
3. Access IoT devices securely

---

## Troubleshooting

### "Can't Access NAS"

**Solutions:**

1. **Check GoConnect status:**
   - Both devices should show "Connected"
   - Verify virtual IPs

2. **Test network:**
   ```bash
   ping 10.0.1.10
   ```

3. **Check NAS services:**
   - Verify SSH is enabled
   - Check SMB/CIFS services

### "Slow File Transfer"

**Solutions:**

1. **Check bandwidth:**
   - Test with speed test tool
   - Check both connections

2. **Optimize GoConnect:**
   - Lower MTU if needed
   - Use wired connections when possible

3. **Check NAS performance:**
   - Check disk usage
   - Check network interface stats

### "Server Unreachable"

**Solutions:**

1. **Verify server is on:**
   - Check power
   - Check network connection
   - Try accessing from different device

2. **Check GoConnect:**
   - Restart GoConnect on both ends
   - Rejoin network if needed

3. **Check firewall:**
   - Ensure GoConnect allowed through firewall
   - Check server firewall

---

## Maintenance

### Regular Tasks

**Weekly:**
- Review access logs
- Check for security updates
- Test remote connections

**Monthly:**
- Update all devices
- Review and update documentation
- Check disk space on NAS

**Quarterly:**
- Review network performance
- Audit user access
- Test backup and restore

### Backup Strategy

**Critical Data:**
- NAS data
- Server configurations
- Application settings

**Backup Locations:**
- Offsite backup (cloud or second location)
- Local backup (external drive)
- Version control for configs

---

## Next Steps

### Expand Your Home Lab

1. **Add more devices:**
   - Smart home hub
   - Security cameras
   - Media servers
   - Development boxes

2. **Add services:**
   - Git server (Gitea/Gogs)
   - CI/CD (Jenkins/GitLab CI)
   - Monitoring (Prometheus/Grafana)
   - Backup (Borg/Restic)

3. **Document everything:**
   - Network diagram
   - IP inventory
   - Service documentation
   - Runbook

---

## Tips

### Organization

- Use naming conventions for devices
- Label cables
- Document all changes
- Use inventory management tools

### Power Management

- Use UPS for servers
- Configure auto-shutdown for NAS
- Schedule wake/sleep for devices

### Performance

- Use wired connections for servers
- Place WiFi optimally
- Use quality networking gear

---

## Need Help?

- 📖 [Full Documentation](../../README.md)
- 💬 [GitHub Discussions](https://github.com/orhaniscoding/goconnect/discussions)
- 🐛 [Report Issue](https://github.com/orhaniscoding/goconnect/issues/new)

---

**Last Updated:** 2025-01-24
**Version:** 1.0.0
