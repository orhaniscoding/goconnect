# 🐧 Linux Kurulum Rehberi

Linux için GoConnect kurulumu, yapılandırması ve sorun giderme.

---

## 📑 İçindekiler

- [Desteklenen Dağıtımlar](#desteklenen-dağıtımlar)
- [Sistem Gereksinimleri](#sistem-gereksinimleri)
- [Kurulum Yöntemleri](#kurulum-yöntemleri)
- [Desktop Application](#desktop-application)
- [CLI Application](#cli-application)
- [Self-Hosted Server](#self-hosted-server)
- [Sorun Giderme](#sorun-giderme)

---

## 🐧 Desteklenen Dağıtımlar

| Dağıtım | Sürüm | Durum | Notlar |
|---------|-------|-------|--------|
| **Ubuntu** | 20.04+, 22.04+, 24.04 | ✅ Full support | .deb packages |
| **Debian** | 11+, 12 | ✅ Full support | .deb packages |
| **Fedora** | 35+, 36+, 37+ | ✅ Full support | RPM packages |
| **Arch Linux** | Rolling | ✅ Full support | AUR packages |
| **CentOS/RHEL** | 8+, 9 | ✅ Support | RPM packages |
| **openSUSE** | Tumbleweed | ⚠️ Community | AppImage |

---

## 💻 Sistem Gereksinimleri

### Minimum Gereksinimler

| Bileşen | Minimum | Önerilen |
|---------|---------|----------|
| **CPU** | x86_64 (64-bit) | 2+ core |
| **RAM** | 2 GB | 4 GB+ |
| **Disk** | 100 MB | 200 MB+ |
| **Kernel** | 5.4+ | 5.10+ |

### Gerekli Paketler

**Desktop App için:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y \
  libwebkit2gtk-4.1-dev \
  libappindicator3-dev \
  librsvg2-dev \
  libcairo2-dev \
  libpango1.0-dev \
  libgdk-pixbuf2.0-dev
```

**CLI için:**
```bash
# Sadece wget/curl gerekli
sudo apt install wget
```

---

## 📦 Kurulum Yöntemleri

| Yöntem | Zorluk | Avantaj |
|--------|--------|---------|
| **.deb Package** | ⭐ Basit | Package manager ile |
| **AppImage** | ⭐ Basit | Dağıtımdan bağımsız |
| **Binary** | ⭐⭐ Orta | Her distro |
| **Snap** | ⭐ Basit | Universal (Ubuntu) |
| **AUR** | ⭐⭐ Orta | Arch Linux |

---

## 🖥️ Desktop Application

### Yöntem 1: Debian Package (.deb)

#### Ubuntu/Debian

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_amd64.deb

# Kurun
sudo dpkg -i goconnect_amd64.deb

# Eksik bağımlılıkları çözün
sudo apt-get install -f -y

# Başlatın
goconnect
```

#### Update/Upgrade

```bash
# Repository'yi ekleyin (otomatik güncellemeler için)
sudo apt install -y software-properties-common
sudo add-apt-repository -y "deb https://apt.goconnect.io/ stable main"
wget -qO- https://apt.goconnect.io/KEY.gpg | sudo apt-key add -

# Güncelleyin
sudo apt update
sudo apt install goconnect
```

#### Kaldırma

```bash
sudo apt remove goconnect
sudo apt autoremove
```

---

### Yöntem 2: AppImage

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-amd64.AppImage

# Çalıştırılabilir yapın
chmod +x GoConnect-amd64.AppImage

# Çalıştırın
./GoConnect-amd64.AppImage

# İsteğe bağlı: Sisteme kurun
sudo mv GoConnect-amd64.AppImage /usr/local/bin/goconnect
```

**Desktop entry oluşturun:**
```bash
sudo cat > /usr/share/applications/goconnect.desktop << 'EOF'
[Desktop Entry]
Name=GoConnect
Comment=Virtual LAN made simple
Exec=/usr/local/bin/goconnect
Icon=goconnect
Type=Application
Categories=Network;VPN;
EOF
```

---

### Yöntem 3: Snap (Ubuntu)

```bash
# Kurun
sudo snap install goconnect

# Classic confinement
sudo snap install goconnect --classic

# Başlatın
goconnect

# Kaldırın
sudo snap remove goconnect
```

---

### Yöntem 4: AUR (Arch Linux)

```bash
# Yay (helper) ile
yay -S goconnect

# Veya manuel
git clone https://aur.archlinux.org/goconnect.git
cd goconnect
makepkg -si

# Kaldırma
yay -R goconnect
```

---

## 💻 CLI Application

### Yöntem 1: Binary Download

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_linux_amd64.tar.gz

# Çıkarın
tar -xzf goconnect_linux_amd64.tar.gz
cd goconnect-linux-amd64

# PATH'e ekleyin
sudo cp goconnect /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect

# Kullanın
goconnect
```

### Cross-Platform

```bash
# ARM64 (Raspberry Pi vb.)
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_linux_arm64.tar.gz

# ARMv7 (Raspberry Pi 3 ve öncesi)
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_linux_armv7.tar.gz
```

---

### Yöntem 2: Package Manager

**Homebrew (Linux):**
```bash
# Homebrew kurun
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# GoConnect kurun
brew install goconnect

# Güncelleyin
brew upgrade goconnect

# Kaldırın
brew uninstall goconnect
```

---

## 🏢 Self-Hosted Server

### Docker ile Kurulum (Önerilen)

#### Docker Kurulumu

**Ubuntu/Debian:**
```bash
# Repository'yi ekleyin
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Docker'ı kurun
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Kullanıcınızı docker grubuna ekleyin
sudo usermod -aG docker $USER
newgrp docker

# Test edin
docker run hello-world
```

**Fedora:**
```bash
sudo dnf -y install dnf-plugins-core
sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

sudo systemctl start docker
sudo systemctl enable docker
```

---

#### GoConnect Docker Stack

```bash
# docker-compose.yml indirin
wget https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# .env oluşturun
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DATABASE_URL=postgres://goconnect:$(openssl rand -base64 16)@db:5432/goconnect?sslmode=disable
WG_SERVER_ENDPOINT=$(curl -s ifconfig.me):51820
EOF

# Başlatın
docker compose up -d

# Logları görün
docker compose logs -f

# Durumu kontrol edin
docker compose ps
```

---

### Manual Systemd Service

#### Binary Installation

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_linux_amd64.tar.gz
tar -xzf goconnect-server-linux-amd64.tar.gz
cd goconnect-server-linux-amd64

# Kullanıcı ve dizin oluşturun
sudo useradd -r -s /bin/false goconnect
sudo mkdir -p /etc/goconnect /var/lib/goconnect /var/log/goconnect

# Binary'yi kopyalayın
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# Konfigürasyon
sudo cp config.example.env /etc/goconnect/.env
sudo chown -R goconnect:goconnect /etc/goconnect /var/lib/goconnect /var/log/goconnect
sudo nano /etc/goconnect/.env  # Edit config
```

#### Systemd Service

```bash
# Service dosyası oluşturun
sudo cat > /etc/systemd/system/goconnect.service << 'EOF'
[Unit]
Description=GoConnect Server
After=network.target postgresql.service
Wants=network-online.target

[Service]
Type=simple
User=goconnect
Group=goconnect
ExecStart=/usr/local/bin/goconnect-server -config /etc/goconnect/.env
Restart=on-failure
RestartSec=5s
AmbientCapabilities=CAP_NET_ADMIN

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/goconnect /var/log/goconnect

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goconnect

[Install]
WantedBy=multi-user.target
EOF

# Reload ve başlat
sudo systemctl daemon-reload
sudo systemctl enable goconnect
sudo systemctl start goconnect

# Durumu kontrol edin
sudo systemctl status goconnect
sudo journalctl -u goconnect -f
```

---

## ⚙️ Yapılandırma

### Firewall (UFW)

```bash
# UFW'yi etkinleştirin
sudo ufw enable

# GoConnect portlarına izin verin
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP (opsiyonel)
sudo ufw allow 443/tcp   # HTTPS (opsiyonel)
sudo ufw allow 51820/udp # WireGuard

# Durumu kontrol edin
sudo ufw status
```

### Firewall (firewalld - Fedora/CentOS)

```bash
# Firewall'ı başlatın
sudo systemctl start firewalld
sudo systemctl enable firewalld

# GoConnect için zone ekleyin
sudo firewall-cmd --permanent --add-port=51820/udp
sudo firewall-cmd --permanent --add-service=http

# Reload
sudo firewall-cmd --reload

# Durumu kontrol edin
sudo firewall-cmd --list-all
```

---

### WireGuard Kernel Module

**Gerekli mi?** Opsiyonel, ama önerilir.

**Ubuntu/Debian:**
```bash
sudo apt install wireguard-dkms
```

**Fedora:**
```bash
sudo dnf install wireguard-tools kernel-devel
sudo dkms autoinstall
```

**Arch:**
```bash
sudo pacman -S wireguard-tools
```

---

## 🔧 Sorun Giderme

### "Permission denied"

**Çözüm:**
```bash
chmod +x goconnect
```

---

### "libwebkit2gtk-4.1 not found"

**Çözüm:**
```bash
# Ubuntu/Debian
sudo apt install -y libwebkit2gtk-4.1-dev libappindicator3-dev librsvg2-dev

# Fedora
sudo dnf install webkit2gtk4.1-devel libappindicator-gtk3-devel librsvg2-devel
```

---

### "Cannot create tun device"

**Çözüm:**
```bash
# Modülü yükleyin
sudo modprobe wireguard

# Kalıcı yapmak için
echo "wireguard" | sudo tee /etc/modules-load.d/wireguard.conf
```

---

### "Service won't start"

**Çözüm:**
```bash
# Logları kontrol edin
sudo journalctl -u goconnect -n 50

# Konfigürasyonu doğrulayın
sudo cat /etc/goconnect/.env

# Manual test
sudo -u goconnect /usr/local/bin/goconnect-server -config /etc/goconnect/.env
```

---

### "Port 51820 already in use"

**Çözüm:**
```bash
# Port kullanan process'i bulun
sudo lsof -i :51820
sudo ss -tulnp | grep 51820

# Process'i sonlandırın (gerekirse)
sudo kill <PID>
```

---

## 🗑️ Kaldırma

### Desktop App (.deb)

```bash
sudo apt remove goconnect
sudo apt autoremove

# Konfigürasyonu da silmek için
rm -rf ~/.config/goconnect
rm -rf ~/.local/share/goconnect
```

---

### CLI

```bash
sudo rm /usr/local/bin/goconnect
rm -rf ~/.config/goconnect
```

---

### Server (Systemd)

```bash
# Service'i durdurun
sudo systemctl stop goconnect
sudo systemctl disable goconnect

# Dosyaları silin
sudo rm /etc/systemd/system/goconnect.service
sudo systemctl daemon-reload

# Kullanıcıyı silin
sudo userdel goconnect

# Verileri silin (opsiyonel)
sudo rm -rf /etc/goconnect /var/lib/goconnect /var/log/goconnect
```

---

## 📚 Ek Kaynaklar

- [Genel Kurulum Rehberi](../INSTALLATION.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Self-Hosted Setup](../SELF_HOSTED_SETUP.md)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
