# 🪟 Windows Kurulum Rehberi

Windows için GoConnect kurulumu, yapılandırması ve sorun giderme.

---

## 📑 İçindekiler

- [Sistem Gereksinimleri](#sistem-gereksinimleri)
- [Kurulum Yöntemleri](#kurulum-yöntemleri)
- [Desktop Application](#desktop-application)
- [CLI Application](#cli-application)
- [Self-Hosted Server](#self-hosted-server)
- [Yapılandırma](#yapılandırma)
- [Sorun Giderme](#sorun-giderme)

---

## 💻 Sistem Gereksinimleri

### Minimum Gereksinimler

| Bileşen | Minimum | Önerilen |
|---------|---------|----------|
| **İşletim Sistemi** | Windows 10 (64-bit) | Windows 11 |
| **RAM** | 2 GB | 4 GB+ |
| **Disk Alanı** | 100 MB | 200 MB+ |
| **İnternet** | 1 Mbps | 10 Mbps+ |

### Desteklenen Sürümler

- ✅ Windows 10 (1903 ve üzeri)
- ✅ Windows 11 (tüm sürümler)
- ❌ Windows 8.1 ve öncesi (desteklenmiyor)
- ❌ Windows 7 (desteklenmiyor)

### Gerekli Bileşenler

**Önceden Yüklü Olmalı:**
- WebView2 (genellikle Windows 10/11'de yüklü)
- .NET Framework 4.8+ (genellikle yüklü)

**Otomatik Yüklenen:**
- WireGuard driver (ilk çalıştırmada)

---

## 📦 Kurulum Yöntemleri

### Yöntem Karşılaştırması

| Yöntem | Zorluk | Önerilen Kullanım |
|--------|--------|------------------|
| **Installer (.exe)** | ⭐ Basit | Günlük kullanıcılar |
| **MSI Package** | ⭐⭐ Orta | Kurumsal deploy |
| **Portable (.zip)** | ⭐ Basit | USB, geçici kullanım |
| **Chocolatey** | ⭐⭐ Orta | Geliştiriciler |
| **Manual** | ⭐⭐⭐ Zor | Gelişmiş kullanıcılar |

---

## 🖥️ Desktop Application

### Yöntem 1: Installer (Önerilen)

#### Adım 1: İndirin

```
https://github.com/orhaniscoding/goconnect/releases/latest/download/GoConnect-Setup-x64.exe
```

#### Adım 2: Çalıştırın

1. İndirilen `GoConnect-Setup-x64.exe` dosyasına çift tıklayın
2. **"Windows protected your PC"** uyarısı çıkarsa:
   - **"More info"** butonuna tıklayın
   - **"Run anyway"** seçeneğini seçin

#### Adım 3: Kurulum Sihirbazı

1. **Welcome** ekranında **"Next"** butonuna tıklayın
2. **License Agreement**'ı okuyun ve **"I Agree"** seçin
3. **Installation Folder** seçin (varsayılan: `C:\Users\YourName\AppData\Local\GoConnect`)
4. **Start Menu Folder** oluşturun (varsayılan: "GoConnect")
5. **Additional Tasks**:
   - ✅ Create desktop shortcut
   - ✅ Add to PATH (CLI için)
   - ✅ Auto-start on boot
6. **"Install"** butonuna tıklayın
7. **"Finish"** ile bitirin

#### Adım 4: İlk Çalıştırma

1. **Desktop** kısayoluna tıklayın
2. Veya **Start Menu** → "GoConnect"

**İlk çalıştırmada:**
- WireGuard driver yüklenir (Admin yetkisi gerektirir)
- Windows SmartScreen uyarısı çıkabilir → **"Run anyway"**

---

### Yöntem 2: MSI Package (Kurumsal)

#### Deployment

```powershell
# Tam sessiz kurulum
msiexec /i GoConnect-x64.msi /quiet /norestart

# Özelleştirilmiş dizin
msiexec /i GoConnect-x64.msi INSTALLDIR="C:\Program Files\GoConnect" /quiet

# Log file ile
msiexec /i GoConnect-x64.msi /l*v install.log /quiet
```

#### Group Policy (GPO)

1. MSI dosyasını network share'a koyun
2. **Group Policy Management**'i açın
3. **Computer Configuration** → **Software Installation**
4. MSI dosyasını ekleyin
5. **Assigned** veya **Published** seçin

---

### Yöntem 3: Portable

```powershell
# İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-portable-windows-amd64.zip" -OutFile "goconnect.zip"

# Çıkarın
Expand-Archive -Path "goconnect.zip" -DestinationPath "C:\GoConnect"

# Çalıştırın
cd C:\GoConnect
.\GoConnect.exe
```

---

## 💻 CLI Application

### Yöntem 1: Binary Download

```powershell
# İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-windows-amd64.zip" -OutFile "goconnect.zip"

# Çıkarın
Expand-Archive -Path "goconnect.zip" -DestinationPath "$env:USERPROFILE\goconnect"

# Çalıştırın
$env:USERPROFILE\goconnect\goconnect.exe
```

### Yöntem 2: Chocolatey

```powershell
# Chocolatey kurun (yoksa)
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# GoConnect kurun
choco install goconnect

# Güncelleyin
choco upgrade goconnect

# Kaldırın
choco uninstall goconnect
```

### Yöntem 3: Scoop

```powershell
# Scoop kurun
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# GoConnect kurun
scoop bucket add extras
scoop install goconnect

# Güncelleyin
scoop update goconnect

# Kaldırın
scoop uninstall goconnect
```

---

## 🏢 Self-Hosted Server

### Docker ile Kurulum

#### Önemli: WSL2 Gerekli

Windows Docker Desktop **WSL2** backend gerektirir.

```powershell
# WSL2'yi aktifleştirin
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# Yeniden başlatın
Restart-Computer

# WSL2'yi varsayılan yapın
wsl --set-default-version 2

# Ubuntu'yu kurun
wsl --install -d Ubuntu
```

#### Docker Kurulumu

```powershell
# Docker Desktop indirin
# https://www.docker.com/products/docker-desktop

# Kurun ve WSL2 backend seçin
```

#### GoConnect Server

```powershell
# docker-compose.yml indirin
curl -LO https://raw.githubusercontent.com/orhaniscoding/goconnect/main/docker-compose.yml

# .env oluşturun
$env:JWT_SECRET = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | % {[char]$_})
$env:DATABASE_URL = "postgres://goconnect:$(-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 16 | % {[char]$_}))@db:5432/goconnect?sslmode=disable"
$env:WG_SERVER_ENDPOINT = "$(curl -s ifconfig.me):51820"

"JWT_SECRET=$env:JWT_SECRET" | Out-File -Encoding ASCII .env
"DATABASE_URL=$env:DATABASE_URL" | Out-File -Encoding ASCII -Append .env
"WG_SERVER_ENDPOINT=$env:WG_SERVER_ENDPOINT" | Out-File -Encoding ASCII -Append .env

# Başlatın
docker compose up -d

# Logları görün
docker compose logs -f
```

---

### Manual Binary Installation

```powershell
# İndirin
Invoke-WebRequest -Uri "https://github.com/orhaniscoding/goconnect/releases/latest/download/goconnect-server-windows-amd64.zip" -OutFile "server.zip"

# Çıkarın
Expand-Archive -Path "server.zip" -DestinationPath "C:\GoConnect-Server"

# Konfigürasyon
Copy-Item "C:\GoConnect-Server\config.example.env" "C:\GoConnect-Server\.env"
notepad "C:\GoConnect-Server\.env"  # Edit config

# Service kurun (NSSM)
# İndir: https://nssm.cc/download
nssm install GoConnect "C:\GoConnect-Server\goconnect-server.exe" "-config" "C:\GoConnect-Server\.env"
nssm set GoConnect AppDirectory C:\GoConnect-Server
nssm set GoConnect DisplayName GoConnect Server
nssm set GoConnect Description GoConnect P2P VPN Server
nssm set GoConnect Start SERVICE_AUTO_START

# Başlatın
nssm start GoConnect

# Durumu kontrol edin
nssm status GoConnect
```

---

## ⚙️ Yapılandırma

### Firewall Kuralları

#### PowerShell ile

```powershell
# GoConnect için inbound rule
New-NetFirewallRule -DisplayName "GoConnect (HTTP)" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "GoConnect (WireGuard)" -Direction Inbound -LocalPort 51820 -Protocol UDP -Action Allow

# Outbound (opsiyonel)
New-NetFirewallRule -DisplayName "GoConnect Outbound" -Direction Outbound -Program "C:\Users\$env:USERNAME\AppData\Local\GoConnect\GoConnect.exe" -Action Allow
```

#### GUI ile

1. **Windows Security** → **Firewall & network protection**
2. **"Allow an app through firewall"**
3. **"Change settings"** → **"Allow another app..."**
4. GoConnect executable'ı bulun:
   - `C:\Users\YourName\AppData\Local\GoConnect\GoConnect.exe`
5. **Private** ve **Public** network'leri işaretleyin
6. **OK**

---

### Windows Defender Exclusion

**Hata:** "Windows Defender virüs buldu" (yanlış pozitif)

#### PowerShell ile

```powershell
# GoConnect dizinini hariç tut
Add-MpPreference -ExclusionPath "C:\Users\$env:USERNAME\AppData\Local\GoConnect"

# Process'i hariç tut
Add-MpPreference -ExclusionProcess "GoConnect.exe"
```

#### GUI ile

1. **Windows Security** → **Virus & threat protection**
2. **"Manage settings"**
3. **"Exclusions"** → **"Add or remove exclusions"**
4. **"Add an exclusion"** → **Folder**
5. `C:\Users\YourName\AppData\Local\GoConnect`

---

### Proxy Ayarları

GoConnect Windows proxy ayarlarını otomatik kullanır.

**Manuel ayar:**

```powershell
# Environment variable
$env:HTTP_PROXY = "http://proxy.example.com:8080"
$env:HTTPS_PROXY = "http://proxy.example.com:8080"
$env:NO_PROXY = "localhost,127.0.0.1,.local"
```

---

## 🔧 Sorun Giderme

### "Windows Defender engelliyor"

**Açıklama:** Yanlış pozitif

**Çözüm:**
```
Windows Security → Virus & threat protection
→ "Current threats" → "Protection history"
→ "Actions" → "Allow on device"
```

---

### "WebView2 eksik"

**Çözüm:**
```powershell
# İndirin
winget install Microsoft.WebView2.Runtime

# Veya manuel indirin
# https://developer.microsoft.com/en-us/microsoft-edge/webview2/
```

---

### "Driver yüklenemedi"

**Çözüm:**
```powershell
# Update & Security → Windows Update
# Tüm güncellemeleri yükleyin
```

---

### "Service başlamıyor"

```powershell
# Logları kontrol edin
nssm status GoConnect
nssm edit GoConnect  # Config'i kontrol edin

# Event Viewer'da logları görün
eventvwr.msc
→ Windows Logs → Application
```

---

### "Port zaten kullanımda"

```powershell
# Port kullanan process'i bulun
netstat -ano | findstr :51820

# Process'i sonlandırın (gerekirse)
taskkill /PID <PID> /F
```

---

## 🗑️ Kaldırma

### Desktop App

```
Settings → Apps → Installed apps
→ GoConnect → Uninstall
```

**Manuel temizlik:**
```powershell
# User data sil (opsiyonel)
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\GoConnect"
Remove-Item -Recurse -Force "$env:APPDATA\GoConnect"
```

---

### CLI

```powershell
# Chocolatey ile kurulduysa
choco uninstall goconnect

# Scoop ile kurulduysa
scoop uninstall goconnect

# Manuel
Remove-Item "$env:USERPROFILE\goconnect"
```

---

### Server

```powershell
# Service sil
nssm stop GoConnect
nssm remove GoConnect confirm

# Dosyaları sil
Remove-Item -Recurse -Force "C:\GoConnect-Server"
```

---

## 📚 Ek Kaynaklar

- [Genel Kurulum Rehberi](../INSTALLATION.md)
- [Troubleshooting](../TROUBLESHOOTING.md)
- [Geliştirme Rehberi](../DEVELOPMENT.md)

---

**Son güncelleme**: 2025-01-24
**Belge sürümü**: v3.0.0
