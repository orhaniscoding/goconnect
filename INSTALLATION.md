# 📥 Kurulum Rehberi

GoConnect'i farklı platformlara ve kullanım senaryolarına göre kurmanın tam kılavuzu.

---

## 📑 İçindekiler

- [Kurulum Seçenekleri](#kurulum-seçenekleri)
- [1. Desktop Application](#1-desktop-application)
- [2. Terminal Application (CLI)](#2-terminal-application-cli)
- [3. Self-Hosted Server](#3-self-hosted-server)
- [4. Platform-Specific Rehberler](#4-platform-specific-rehberler)
- [Sorun Giderme](#sorun-giderme)

---

## 🎯 Kurulum Seçenekleri

GoConnect'i kurmak için üç farklı yol var:

| Seçenek | İçin Uygun | Zorluk | Esneklik |
|---------|-----------|--------|----------|
| **Desktop App** | Günlük kullanıcılar | ⭐ Basit | Orta |
| **CLI** | Geliştiriciler, sunucular | ⭐⭐ Orta | Yüksek |
| **Self-Hosted** | Organizasyonlar, gizlilik | ⭐⭐⭐ Zor | Çok Yüksek |

**Hangisi sizin için?**

- 🖥️ **Bilgisayarınızda kullanmak istiyorum** → Desktop App
- 🖥️ **Terminal seviyeyim, sunucuda kuracağım** → CLI
- 🏢 **Kendi sunucumu kurmak, tüm kontrol istiyorum** → Self-Hosted

---

## 1. Desktop Application

### 🪟 Windows

#### Sistem Gereksinimleri
- Windows 10 (64-bit) veya üzeri
- 100 MB boş disk alanı
- İnternet bağlantısı (ilk kurulum için)

#### Kurulum Adımları

**Yöntem 1: Installer (.exe) - Önerilen**

1. **İndirin**:
   - [GoConnect-Setup.exe](https://github.com/orhaniscoding/goconnect/releases/latest) dosyasını indirin

2. **Çalıştırın**:
   - İndirilen dosyaya çift tıklayın
   - "Evet" diyerek Windows Defender'ı geçin
   - Kurulum sihirbazını takip edin

3. **Başlatın**:
   - Başlat menüsünden "GoConnect"i arayın
   - veya Masaüstündeki ikona tıklayın

**Yöntem 2: MSI Package (Kurumsal)**

```powershell
# PowerShell Komut İstemi'nden (Admin)
msiexec /i GoConnect-x64.msi /quiet /norestart
```

#### Güncelleme

Desktop App otomatik güncelleme özelliğine sahiptir:
- Arka planda kontrol eder
- Yeni sürüm çıktığında bildirim gösterir
- Tek tıkla günceller

#### Kaldırma

```powershell
# Ayarlar → Uygulamalar → GoConnect → Kaldır
# veya
Control Panel → Programs and Features → GoConnect → Uninstall
```

---

### 🍎 macOS

#### Sistem Gereksinimleri
- macOS 11 (Big Sur) veya üzeri
- 100 MB boş disk alanı
- Apple Silicon (M1/M2/M3) veya Intel Mac

#### Kurulum Adımları

**Apple Silicon (M1/M2/M3)**

1. **İndirin**:
   ```bash
   curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-aarch64.dmg
   ```

2. **Açın**:
   - İndirilen `.dmg` dosyasına çift tıklayın

3. **Kurun**:
   - GoConnect simgesini "Applications" klasörüne sürükleyin
   - Dock'a eklemek için simgeyi sağ tıklayın → "Options" → "Keep in Dock"

4. **İlk Çalıştırma**:
   - Launchpad'den GoConnect'i açın
   - "Open" diyerek macOS güvenlik uyarısını geçin

**Intel Mac**

1. **İndirin**:
   ```bash
   curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-x64.dmg
   ```

2. Aynı adımları izleyin

#### Güncelleme

App menüsünden → "Check for Updates..." seçeneğini kullanın.

#### Kaldırma

```bash
# Applications klasöründen sürükleyin ve Çöp Kutusuna atın
rm -rf /Applications/GoConnect.app

# Kullanıcı verilerini temizlemek için:
rm -rf ~/Library/Application\ Support/com.goconnect.app
rm -rf ~/Library/Caches/com.goconnect.app
```

---

### 🐧 Linux

#### Sistem Gereksinimleri
- Ubuntu 20.04+, Debian 11+, Fedora 35+, Arch Linux
- 100 MB boş disk alanı
- Wayland veya X11

#### Kurulum Adımları

**Yöntem 1: Debian/Ubuntu (.deb) - Önerilen**

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_amd64.deb

# Kurun
sudo dpkg -i goconnect_amd64.deb

# Eksik bağımlılıkları çözün
sudo apt-get install -f -y
```

**Yöntem 2: AppImage (Universal)**

```bash
# İndirin ve çalıştırılabilir yapın
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-amd64.AppImage
chmod +x GoConnect-amd64.AppImage

# Çalıştırın
./GoConnect-amd64.AppImage
```

**Yöntem 3: Manuel Kurulum**

```bash
# İndirin ve çıkarın
tar -xzf goconnect-linux-amd64.tar.gz
cd goconnect-linux-amd64

# Kopyalayın
sudo cp GoConnect /usr/local/bin/
sudo chmod +x /usr/local/bin/GoConnect

# Masaüstü girişi oluştur
sudo cp goconnect.desktop /usr/share/applications/
```

#### Güncelleme

```bash
# Debian/Ubuntu
sudo apt update && sudo apt install goconnect

# AppImage
# Yeni sürümü indirin ve eskisinin üzerine yazın
```

#### Kaldırma

```bash
# Debian/Ubuntu
sudo apt remove goconnect

# AppImage
# Sadece dosyayı silin
rm GoConnect-amd64.AppImage
```

---

## 2. Terminal Application (CLI)

### 🖥️ Windows

```powershell
# İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_windows_amd64.zip" -OutFile "goconnect.zip"

# Çıkarın
Expand-Archive -Path "goconnect.zip" -DestinationPath "."

# Kullanın
.\goconnect.exe
```

**PATH'e eklemek için:**

```powershell
# Klasör oluşturun
New-Item -ItemType Directory -Path "$env:USERPROFILE\goconnect" -Force

# Binary'yi taşıyın
Move-Item -Path ".\goconnect.exe" -Destination "$env:USERPROFILE\goconnect\"

# PATH'e ekleyin (kalıcı)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\goconnect", "User")

# Yeni terminal açın ve kullanın
goconnect
```

---

### 🍎 macOS

```bash
# Apple Silicon
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_darwin_arm64.tar.gz

# Intel
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_darwin_amd64.tar.gz

# Çıkarın
tar -xzf goconnect_darmin_*.tar.gz

# PATH'e ekleyin
sudo mv goconnect /usr/local/bin/

# Kullanın
goconnect
```

---

### 🐧 Linux

```bash
# Linux (x64)
curl -LO https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect_linux_amd64.tar.gz

# Çıkarın
tar -xzf goconnect_linux_amd64.tar.gz

# PATH'e ekleyin
sudo mv goconnect /usr/local/bin/

# Kullanın
goconnect
```

**Package Manager ile (Scoop - Linux)**

```bash
# Scoop kuruluysa
scoop install goconnect
```

---

## 3. Self-Hosted Server

### 🐳 Docker ile Kurulum (Önerilen)

#### Hızlı Başlangıç

```bash
# docker-compose.yml indirin
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# .env dosyası oluşturun
cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
DATABASE_URL=postgres://goconnect:$(openssl rand -base64 16)@db:5432/goconnect?sslmode=disable
WG_SERVER_ENDPOINT=your-domain.com:51820
EOF

# Başlatın
docker compose up -d

# Logları görüntüleyin
docker compose logs -f
```

#### Docker Compose Dosyası

```yaml
version: '3.8'

services:
  goconnect:
    image: ghcr.io/orhaniscoding/goconnect-server:latest
    container_name: goconnect
    restart: unless-stopped
    ports:
      - "8080:8080"    # HTTP API
      - "51820:51820/udp"  # WireGuard
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - WG_SERVER_ENDPOINT=${WG_SERVER_ENDPOINT}
      - HTTP_PORT=8080
      - LOG_LEVEL=info
    volumes:
      - goconnect-data:/data
      - /dev/net/tun:/dev/net/tun
    cap_add:
      - NET_ADMIN
    depends_on:
      - db

  db:
    image: postgres:15-alpine
    container_name: goconnect-db
    restart: unless-stopped
    environment:
      - POSTGRES_USER=goconnect
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=goconnect
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  goconnect-data:
  postgres-data:
```

---

### 🖥️ Manual Binary Installation

#### Linux

```bash
# İndirin
wget https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_linux_amd64.tar.gz

# Çıkarın
tar -xzf goconnect-server_linux_amd64.tar.gz
cd goconnect-server-linux-amd64

# Kurun
sudo cp goconnect-server /usr/local/bin/
sudo chmod +x /usr/local/bin/goconnect-server

# Kullanıcı oluşturun
sudo useradd -r -s /bin/false goconnect

# Konfigürasyon
sudo mkdir -p /etc/goconnect
sudo cp config.example.env /etc/goconnect/.env
sudo nano /etc/goconnect/.env  # Edit config

# Systemd service
sudo cp goconnect.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable goconnect
sudo systemctl start goconnect
```

#### Systemd Service Dosyası

```ini
[Unit]
Description=GoConnect Server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=goconnect
Group=goconnect
ExecStart=/usr/local/bin/goconnect-server -config /etc/goconnect/.env
Restart=on-failure
RestartSec=5s

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/goconnect

# Capabilities
CapabilityBoundingSet=CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
```

---

### 🪟 Windows (Self-Hosted)

```powershell
# İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server_windows_amd64.zip" -OutFile "server.zip"

# Çıkarın
Expand-Archive -Path "server.zip" -DestinationPath "C:\GoConnect"

# Konfigürasyon
Copy-Item "C:\GoConnect\config.example.env" "C:\GoConnect\.env"
notepad "C:\GoConnect\.env"  # Edit

# Service olarak kurun (NSSM)
# İndirin: https://nssm.cc/download
nssm install GoConnect "C:\GoConnect\goconnect-server.exe" "-config" "C:\GoConnect\.env"
nssm start GoConnect
```

---

## 4. Platform-Specific Rehberler

Platforma özgü detaylı kurulum rehberleri için:

- 🪟 **[Windows Kurulum](docs/installations/WINDOWS.md)**
- 🍎 **[macOS Kurulum](docs/installations/MACOS.md)**
- 🐧 **[Linux Kurulum](docs/installations/LINUX.md)**
- 🐳 **[Docker Kurulum](docs/installations/DOCKER.md)**

---

## 🔧 Sorun Giderme

### Kurulum Sorunları

**Windows: "Windows Defender'da virüs bulundu"**
- Bu yanlış pozitiftir
- "Daha fazla bilgi" → "Yine de çalıştır" diyerek geçin

**macOS: "GoConnect has been damaged"**
```bash
# Quarantine'i kaldırın
sudo xattr -cr /Applications/GoConnect.app
```

**Linux: "Permission denied" hatası**
```bash
# Binary'yi çalıştırılabilir yapın
chmod +x goconnect
```

### İlk Çalıştırma Sorunları

**Bağlantı kurulamıyor:**
- İnternet bağlantınızı kontrol edin
- Firewall ayarlarını kontrol edin (port 8080, 51820)
- VPN'inizi kapatmayı deneyin

**Daemon başlamıyor:**
- Log dosyasını kontrol edin
- Port zaten kullanımda olabilir
- Konfigürasyon dosyasını kontrol edin

Daha fazla sorun giderme için: 👉 [Troubleshooting Guide](TROUBLESHOOTING.md)

---

## 📞 Yardım

Kurulum sırasında sorun yaşarsanız:

- 📖 [Detaylı rehberler](docs/installations/) inceleyin
- ❓ [Sık Sorulan Sorular](FAQ.md) okuyun
- 🐙 [GitHub Issues](https://github.com/orhaniscoding/goconnect/issues/new) sorun bildirin
- 💬 [Discussions](https://github.com/orhaniscoding/goconnect/discussions) tartışmaya katılın

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
